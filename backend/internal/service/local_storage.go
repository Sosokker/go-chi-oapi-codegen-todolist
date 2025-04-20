package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sosokker/todolist-backend/internal/config"
	"github.com/google/uuid"
)

type localStorageService struct {
	basePath string
	logger   *slog.Logger
}

// NewLocalStorageService creates a service for storing files on the local disk.
func NewLocalStorageService(cfg config.LocalStorageConfig, logger *slog.Logger) (FileStorageService, error) {
	if cfg.Path == "" {
		return nil, fmt.Errorf("local storage path cannot be empty")
	}

	// Ensure the base directory exists
	err := os.MkdirAll(cfg.Path, 0750) // Use appropriate permissions
	if err != nil {
		return nil, fmt.Errorf("failed to create local storage directory '%s': %w", cfg.Path, err)
	}

	logger.Info("Local file storage initialized", "path", cfg.Path)
	return &localStorageService{
		basePath: cfg.Path,
		logger:   logger.With("service", "localstorage"),
	}, nil
}

// GenerateUniqueObjectName creates a unique path/filename for storage.
// Example: user_uuid/todo_uuid/file_uuid.ext
func (s *localStorageService) GenerateUniqueObjectName(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	fileName := uuid.NewString() + ext
	return fileName
}

func (s *localStorageService) Upload(ctx context.Context, userID, todoID uuid.UUID, originalFilename string, reader io.Reader, size int64) (string, string, error) {
	// Create a unique filename
	uniqueFilename := s.GenerateUniqueObjectName(originalFilename)

	// Create user/todo specific subdirectory structure
	subDir := filepath.Join(userID.String(), todoID.String())
	fullDir := filepath.Join(s.basePath, subDir)
	if err := os.MkdirAll(fullDir, 0750); err != nil {
		s.logger.ErrorContext(ctx, "Failed to create subdirectory for upload", "error", err, "path", fullDir)
		return "", "", fmt.Errorf("could not create storage directory: %w", err)
	}

	// Define the full path for the file
	filePath := filepath.Join(fullDir, uniqueFilename)
	storageID := filepath.Join(subDir, uniqueFilename) // Relative path used as ID

	// Create the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to create destination file", "error", err, "path", filePath)
		return "", "", fmt.Errorf("could not create file: %w", err)
	}
	defer dst.Close()

	// Copy the content from the reader to the destination file
	written, err := io.Copy(dst, reader)
	if err != nil {
		// Attempt to clean up partially written file
		os.Remove(filePath)
		s.logger.ErrorContext(ctx, "Failed to copy file content", "error", err, "path", filePath)
		return "", "", fmt.Errorf("could not write file content: %w", err)
	}
	if written != size {
		// Attempt to clean up file if size mismatch (could indicate truncated upload)
		os.Remove(filePath)
		s.logger.WarnContext(ctx, "File size mismatch during upload", "expected", size, "written", written, "path", filePath)
		return "", "", fmt.Errorf("file size mismatch during upload")
	}

	// Detect content type
	contentType := s.detectContentType(filePath, originalFilename)

	s.logger.InfoContext(ctx, "File uploaded successfully", "storageId", storageID, "originalName", originalFilename, "size", size, "contentType", contentType)
	// Return the relative path as the storage identifier
	return storageID, contentType, nil
}

func (s *localStorageService) Delete(ctx context.Context, storageID string) error {
	// Prevent directory traversal attacks
	cleanStorageID := filepath.Clean(storageID)
	if strings.Contains(cleanStorageID, "..") {
		s.logger.WarnContext(ctx, "Attempted directory traversal in delete", "storageId", storageID)
		return fmt.Errorf("invalid storage ID")
	}

	fullPath := filepath.Join(s.basePath, cleanStorageID)

	err := os.Remove(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			s.logger.WarnContext(ctx, "Attempted to delete non-existent file", "storageId", storageID)
			// Consider returning nil here if deleting non-existent is okay
			return nil
		}
		s.logger.ErrorContext(ctx, "Failed to delete file", "error", err, "storageId", storageID)
		return fmt.Errorf("could not delete file: %w", err)
	}

	s.logger.InfoContext(ctx, "File deleted successfully", "storageId", storageID)

	dir := filepath.Dir(fullPath)
	if isEmpty, _ := IsDirEmpty(dir); isEmpty {
		os.Remove(dir)
	}
	dir = filepath.Dir(dir) // Go up one more level
	if isEmpty, _ := IsDirEmpty(dir); isEmpty {
		os.Remove(dir)
	}

	return nil
}

// GetURL for local storage might just return a path or require a separate file server.
// This implementation returns a placeholder indicating it's not a direct URL.
func (s *localStorageService) GetURL(ctx context.Context, storageID string) (string, error) {
	// Local storage doesn't inherently provide a web URL.
	// You would typically need a separate static file server pointing to `basePath`.
	// For now, return the storageID itself or a placeholder path.
	// Example: If you have a file server at /static/uploads mapped to basePath:
	// return "/static/uploads/" + filepath.ToSlash(storageID), nil
	return fmt.Sprintf("local://%s", storageID), nil // Placeholder indicating local storage
}

// detectContentType tries to determine the MIME type of the file.
func (s *localStorageService) detectContentType(filePath string, originalFilename string) string {
	// First, try based on file extension
	ext := filepath.Ext(originalFilename)
	mimeType := mime.TypeByExtension(ext)
	if mimeType != "" {
		return mimeType
	}

	// If extension didn't work, try reading the first 512 bytes
	file, err := os.Open(filePath)
	if err != nil {
		s.logger.Warn("Could not open file for content type detection", "error", err, "path", filePath)
		return "application/octet-stream" // Default fallback
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		s.logger.Warn("Could not read file for content type detection", "error", err, "path", filePath)
		return "application/octet-stream"
	}

	// http.DetectContentType works best with the file beginning
	mimeType = http.DetectContentType(buffer[:n])
	return mimeType
}

func IsDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// Read just one entry. If EOF, directory is empty.
	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error during read
}

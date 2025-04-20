package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/Sosokker/todolist-backend/internal/config"
	"github.com/google/uuid"
	"google.golang.org/api/option"
)

type gcsStorageService struct {
	bucket          string
	client          *storage.Client
	logger          *slog.Logger
	baseDir         string
	signedURLExpiry time.Duration
}

func NewGCStorageService(cfg config.GCSStorageConfig, logger *slog.Logger) (FileStorageService, error) {
	if cfg.BucketName == "" {
		return nil, fmt.Errorf("GCS bucket name is required")
	}

	opts := []option.ClientOption{}
	// Prefer environment variable GOOGLE_APPLICATION_CREDENTIALS
	// Only use CredentialsFile from config if it's explicitly set
	if cfg.CredentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(cfg.CredentialsFile))
		logger.Info("Using GCS credentials file specified in config", "path", cfg.CredentialsFile)
	} else {
		logger.Info("Using default GCS credentials (e.g., GOOGLE_APPLICATION_CREDENTIALS or Application Default Credentials)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	// Check bucket existence and permissions
	_, err = client.Bucket(cfg.BucketName).Attrs(ctx)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to access GCS bucket '%s': %w", cfg.BucketName, err)
	}

	logger.Info("GCS storage service initialized", "bucket", cfg.BucketName, "baseDir", cfg.BaseDir)

	return &gcsStorageService{
		bucket:          cfg.BucketName,
		client:          client,
		logger:          logger.With("service", "gcsstorage"),
		baseDir:         strings.Trim(cfg.BaseDir, "/"), // Ensure no leading/trailing slashes
		signedURLExpiry: 168 * time.Hour,                // Default signed URL validity
	}, nil
}

// GenerateUniqueObjectName creates a unique object path within the bucket's base directory.
// Example: attachments/<user_uuid>/<todo_uuid>/<file_uuid>.<ext>
func (s *gcsStorageService) GenerateUniqueObjectName(userID, todoID uuid.UUID, originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	fileName := uuid.NewString() + ext
	objectPath := filepath.Join(s.baseDir, userID.String(), todoID.String(), fileName)
	return filepath.ToSlash(objectPath)
}
func (s *gcsStorageService) Upload(ctx context.Context, userID, todoID uuid.UUID, originalFilename string, reader io.Reader, size int64) (string, string, error) {
	objectName := s.GenerateUniqueObjectName(userID, todoID, originalFilename)

	ctxUpload, cancel := context.WithTimeout(ctx, 5*time.Minute) // Timeout for upload
	defer cancel()

	wc := s.client.Bucket(s.bucket).Object(objectName).NewWriter(ctxUpload)

	// Attempt to determine Content-Type
	contentType := mime.TypeByExtension(filepath.Ext(originalFilename))
	if contentType == "" {
		contentType = "application/octet-stream" // Default fallback
		// Could potentially read first 512 bytes from reader here if it's TeeReader, but might be complex
	}
	wc.ContentType = contentType
	wc.ChunkSize = 0 // Recommended for better performance unless files are huge

	s.logger.DebugContext(ctx, "Uploading file to GCS", "bucket", s.bucket, "object", objectName, "contentType", contentType, "size", size)

	written, err := io.Copy(wc, reader)
	if err != nil {
		// Close writer explicitly on error to clean up potential partial uploads
		_ = wc.CloseWithError(fmt.Errorf("copy failed: %w", err))
		s.logger.ErrorContext(ctx, "Failed to copy data to GCS", "error", err, "object", objectName)
		return "", "", fmt.Errorf("failed to upload to GCS: %w", err)
	}

	// Close the writer to finalize the upload
	if err := wc.Close(); err != nil {
		s.logger.ErrorContext(ctx, "Failed to finalize GCS upload", "error", err, "object", objectName)
		return "", "", fmt.Errorf("failed to finalize upload: %w", err)
	}

	if written != size {
		s.logger.WarnContext(ctx, "File size mismatch during GCS upload", "expected", size, "written", written, "object", objectName)
		// Optionally delete the potentially corrupted file
		_ = s.Delete(context.Background(), objectName) // Use background context for cleanup
		return "", "", fmt.Errorf("file size mismatch during upload")
	}

	s.logger.InfoContext(ctx, "File uploaded successfully to GCS", "object", objectName, "size", written, "contentType", contentType)
	// Return the object name (path) as the storage ID
	return objectName, contentType, nil
}

func (s *gcsStorageService) Delete(ctx context.Context, storageID string) error {
	objectName := filepath.Clean(storageID) // storageID is the object path
	if strings.Contains(objectName, "..") || !strings.HasPrefix(objectName, s.baseDir+"/") && s.baseDir != "" {
		s.logger.WarnContext(ctx, "Attempted invalid delete operation", "storageId", storageID, "baseDir", s.baseDir)
		return fmt.Errorf("invalid storage ID for deletion")
	}

	ctxDelete, cancel := context.WithTimeout(ctx, 30*time.Second) // Timeout for delete
	defer cancel()

	o := s.client.Bucket(s.bucket).Object(objectName)
	err := o.Delete(ctxDelete)

	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			s.logger.WarnContext(ctx, "Attempted to delete non-existent GCS object", "storageId", storageID)
			return nil // Treat as success if already deleted
		}
		s.logger.ErrorContext(ctx, "Failed to delete GCS object", "error", err, "storageId", storageID)
		return fmt.Errorf("could not delete GCS object: %w", err)
	}

	s.logger.InfoContext(ctx, "GCS object deleted successfully", "storageId", storageID)
	return nil
}

// GetURL generates a signed URL for accessing the private GCS object.
func (s *gcsStorageService) GetURL(ctx context.Context, storageID string) (string, error) {
	objectName := storageID
	if strings.Contains(objectName, "..") || (s.baseDir != "" && !strings.HasPrefix(objectName, s.baseDir+"/")) {
		s.logger.WarnContext(ctx, "Attempted invalid GetURL operation", "storageId", storageID, "baseDir", s.baseDir)
		return "", fmt.Errorf("invalid storage ID for URL generation")
	}

	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(s.signedURLExpiry),
	}

	url, err := s.client.Bucket(s.bucket).SignedURL(objectName, opts)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to generate signed URL", "error", err, "object", objectName)
		return "", fmt.Errorf("could not get signed URL for object: %w", err)
	}

	s.logger.DebugContext(ctx, "Generated signed URL", "object", objectName, "expiry", opts.Expires)
	return url, nil
}

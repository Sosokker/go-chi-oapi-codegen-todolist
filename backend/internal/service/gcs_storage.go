package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/Sosokker/todolist-backend/internal/config"
	"github.com/google/uuid"
	"google.golang.org/api/option"
)

type gcsStorageService struct {
	bucket  string
	client  *storage.Client
	logger  *slog.Logger
	baseDir string
}

func NewGCStorageService(cfg config.GCSStorageConfig, logger *slog.Logger) (FileStorageService, error) {
	opts := []option.ClientOption{}
	if cfg.CredentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(cfg.CredentialsFile))
	}
	client, err := storage.NewClient(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}
	return &gcsStorageService{
		bucket:  cfg.BucketName,
		client:  client,
		logger:  logger.With("service", "gcsstorage"),
		baseDir: cfg.BaseDir,
	}, nil
}

func (s *gcsStorageService) GenerateUniqueObjectName(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	return uuid.NewString() + ext
}

func (s *gcsStorageService) Upload(ctx context.Context, userID, todoID uuid.UUID, originalFilename string, reader io.Reader, size int64) (string, string, error) {
	objectName := filepath.Join(s.baseDir, userID.String(), todoID.String(), s.GenerateUniqueObjectName(originalFilename))
	wc := s.client.Bucket(s.bucket).Object(objectName).NewWriter(ctx)
	wc.ContentType = mime.TypeByExtension(filepath.Ext(originalFilename))
	wc.ChunkSize = 0
	written, err := io.Copy(wc, reader)
	if err != nil {
		wc.Close()
		s.logger.ErrorContext(ctx, "Failed to upload to GCS", "error", err, "object", objectName)
		return "", "", fmt.Errorf("failed to upload to GCS: %w", err)
	}
	if written != size {
		wc.Close()
		s.logger.WarnContext(ctx, "File size mismatch during GCS upload", "expected", size, "written", written, "object", objectName)
		return "", "", fmt.Errorf("file size mismatch during upload")
	}
	if err := wc.Close(); err != nil {
		s.logger.ErrorContext(ctx, "Failed to finalize GCS upload", "error", err, "object", objectName)
		return "", "", fmt.Errorf("failed to finalize upload: %w", err)
	}
	contentType := wc.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return objectName, contentType, nil
}

func (s *gcsStorageService) Delete(ctx context.Context, storageID string) error {
	objectName := filepath.Clean(storageID)
	if strings.Contains(objectName, "..") {
		s.logger.WarnContext(ctx, "Attempted directory traversal in GCS delete", "storageId", storageID)
		return fmt.Errorf("invalid storage ID")
	}
	o := s.client.Bucket(s.bucket).Object(objectName)
	err := o.Delete(ctx)
	if err != nil && err != storage.ErrObjectNotExist {
		s.logger.ErrorContext(ctx, "Failed to delete GCS object", "error", err, "storageId", storageID)
		return fmt.Errorf("could not delete GCS object: %w", err)
	}
	s.logger.InfoContext(ctx, "GCS object deleted", "storageId", storageID)
	return nil
}

func (s *gcsStorageService) GetURL(ctx context.Context, storageID string) (string, error) {
	objectName := filepath.Clean(storageID)
	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucket, objectName)
	return url, nil
}

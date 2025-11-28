// Package storage provides file storage implementations.
package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/xiiisorate/granula_api/floorplan-service/internal/config"
	apperrors "github.com/xiiisorate/granula_api/shared/pkg/errors"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
)

// MinIOStorage provides S3-compatible storage via MinIO.
type MinIOStorage struct {
	client *minio.Client
	bucket string
	cfg    config.StorageConfig
	log    *logger.Logger
}

// NewMinIOStorage creates a new MinIO storage client.
func NewMinIOStorage(cfg config.StorageConfig, log *logger.Logger) (*MinIOStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, apperrors.Internal("failed to create MinIO client").WithCause(err)
	}

	return &MinIOStorage{
		client: client,
		bucket: cfg.Bucket,
		cfg:    cfg,
		log:    log,
	}, nil
}

// EnsureBucket ensures the bucket exists.
func (s *MinIOStorage) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return apperrors.Internal("failed to check bucket").WithCause(err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{
			Region: s.cfg.Region,
		})
		if err != nil {
			return apperrors.Internal("failed to create bucket").WithCause(err)
		}
		s.log.Info("created bucket", logger.String("bucket", s.bucket))
	}

	return nil
}

// UploadFile uploads a file to storage.
func (s *MinIOStorage) UploadFile(ctx context.Context, req UploadRequest) (*UploadResult, error) {
	// Validate file size
	if req.Size > s.cfg.MaxFileSize {
		return nil, apperrors.InvalidArgument("file", "file too large")
	}

	// Validate MIME type
	if !s.isAllowedType(req.MimeType) {
		return nil, apperrors.InvalidArgument("file", fmt.Sprintf("file type %s not allowed", req.MimeType))
	}

	// Generate storage path
	storagePath := s.generateStoragePath(req.WorkspaceID, req.FloorPlanID, req.OriginalName)

	// Upload to MinIO
	_, err := s.client.PutObject(ctx, s.bucket, storagePath, req.Reader, req.Size, minio.PutObjectOptions{
		ContentType: req.MimeType,
		UserMetadata: map[string]string{
			"workspace_id":  req.WorkspaceID.String(),
			"floor_plan_id": req.FloorPlanID.String(),
			"original_name": req.OriginalName,
		},
	})
	if err != nil {
		return nil, apperrors.Internal("failed to upload file").WithCause(err)
	}

	s.log.Info("file uploaded",
		logger.String("path", storagePath),
		logger.Int64("size", req.Size),
	)

	return &UploadResult{
		StoragePath: storagePath,
		Size:        req.Size,
		MimeType:    req.MimeType,
	}, nil
}

// DownloadFile downloads a file from storage.
func (s *MinIOStorage) DownloadFile(ctx context.Context, storagePath string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, storagePath, minio.GetObjectOptions{})
	if err != nil {
		return nil, apperrors.Internal("failed to download file").WithCause(err)
	}

	return obj, nil
}

// GetFileInfo retrieves file metadata from storage.
func (s *MinIOStorage) GetFileInfo(ctx context.Context, storagePath string) (*FileMetadata, error) {
	stat, err := s.client.StatObject(ctx, s.bucket, storagePath, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return nil, apperrors.NotFoundMsg("file not found")
		}
		return nil, apperrors.Internal("failed to stat file").WithCause(err)
	}

	return &FileMetadata{
		StoragePath:  storagePath,
		Size:         stat.Size,
		MimeType:     stat.ContentType,
		LastModified: stat.LastModified,
		ETag:         stat.ETag,
	}, nil
}

// DeleteFile deletes a file from storage.
func (s *MinIOStorage) DeleteFile(ctx context.Context, storagePath string) error {
	err := s.client.RemoveObject(ctx, s.bucket, storagePath, minio.RemoveObjectOptions{})
	if err != nil {
		return apperrors.Internal("failed to delete file").WithCause(err)
	}

	s.log.Info("file deleted", logger.String("path", storagePath))
	return nil
}

// GetPresignedURL generates a presigned URL for direct download.
func (s *MinIOStorage) GetPresignedURL(ctx context.Context, storagePath string, duration time.Duration) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucket, storagePath, duration, nil)
	if err != nil {
		return "", apperrors.Internal("failed to generate presigned URL").WithCause(err)
	}
	return url.String(), nil
}

// UploadThumbnail uploads a thumbnail image.
func (s *MinIOStorage) UploadThumbnail(ctx context.Context, floorPlanID uuid.UUID, size string, data []byte) (string, error) {
	storagePath := fmt.Sprintf("thumbnails/%s/%s.jpg", floorPlanID.String(), size)

	_, err := s.client.PutObject(ctx, s.bucket, storagePath, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return "", apperrors.Internal("failed to upload thumbnail").WithCause(err)
	}

	return storagePath, nil
}

// generateStoragePath generates a unique storage path.
func (s *MinIOStorage) generateStoragePath(workspaceID, floorPlanID uuid.UUID, originalName string) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().Format("20060102")
	return fmt.Sprintf("workspaces/%s/floorplans/%s/%s%s",
		workspaceID.String(),
		timestamp,
		floorPlanID.String(),
		ext,
	)
}

// isAllowedType checks if a MIME type is allowed.
func (s *MinIOStorage) isAllowedType(mimeType string) bool {
	if len(s.cfg.AllowedTypes) == 0 {
		return true
	}
	for _, t := range s.cfg.AllowedTypes {
		if t == mimeType {
			return true
		}
	}
	return false
}

// UploadRequest for file upload.
type UploadRequest struct {
	WorkspaceID  uuid.UUID
	FloorPlanID  uuid.UUID
	OriginalName string
	MimeType     string
	Size         int64
	Reader       io.Reader
}

// UploadResult from file upload.
type UploadResult struct {
	StoragePath string
	Size        int64
	MimeType    string
}

// FileMetadata from storage.
type FileMetadata struct {
	StoragePath  string
	Size         int64
	MimeType     string
	LastModified time.Time
	ETag         string
}

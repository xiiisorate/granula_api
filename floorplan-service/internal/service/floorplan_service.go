// Package service provides business logic for Floor Plan Service.
package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/floorplan-service/internal/domain/entity"
	"github.com/xiiisorate/granula_api/floorplan-service/internal/repository/postgres"
	"github.com/xiiisorate/granula_api/floorplan-service/internal/storage"
	pb "github.com/xiiisorate/granula_api/shared/gen/ai/v1"
	apperrors "github.com/xiiisorate/granula_api/shared/pkg/errors"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
)

// FloorPlanService handles floor plan operations.
type FloorPlanService struct {
	repo     *postgres.FloorPlanRepository
	storage  *storage.MinIOStorage
	aiClient pb.AIServiceClient
	log      *logger.Logger
}

// NewFloorPlanService creates a new FloorPlanService.
func NewFloorPlanService(
	repo *postgres.FloorPlanRepository,
	storage *storage.MinIOStorage,
	aiClient pb.AIServiceClient,
	log *logger.Logger,
) *FloorPlanService {
	return &FloorPlanService{
		repo:     repo,
		storage:  storage,
		aiClient: aiClient,
		log:      log,
	}
}

// Upload uploads a new floor plan.
func (s *FloorPlanService) Upload(ctx context.Context, req UploadRequest) (*entity.FloorPlan, error) {
	s.log.Info("uploading floor plan",
		logger.String("workspace_id", req.WorkspaceID.String()),
		logger.String("name", req.Name),
	)

	// Create floor plan entity
	fp := entity.NewFloorPlan(req.WorkspaceID, req.OwnerID, req.Name)
	fp.Description = req.Description

	// Calculate checksum
	data, err := io.ReadAll(req.FileData)
	if err != nil {
		return nil, apperrors.Internal("failed to read file").WithCause(err)
	}
	checksum := calculateMD5(data)

	// Upload to storage
	uploadResult, err := s.storage.UploadFile(ctx, storage.UploadRequest{
		WorkspaceID:  req.WorkspaceID,
		FloorPlanID:  fp.ID,
		OriginalName: req.FileName,
		MimeType:     req.MimeType,
		Size:         int64(len(data)),
		Reader:       bytes.NewReader(data),
	})
	if err != nil {
		return nil, err
	}

	// Create floor plan record
	if err := s.repo.Create(ctx, fp); err != nil {
		// Cleanup uploaded file
		_ = s.storage.DeleteFile(ctx, uploadResult.StoragePath)
		return nil, err
	}

	// Create file info record
	fileInfo := entity.NewFileInfo(fp.ID, req.FileName, uploadResult.StoragePath, req.MimeType, uploadResult.Size)
	fileInfo.Checksum = checksum

	if err := s.repo.CreateFileInfo(ctx, fileInfo); err != nil {
		return nil, err
	}

	fp.FileInfo = fileInfo

	s.log.Info("floor plan uploaded",
		logger.String("id", fp.ID.String()),
		logger.String("path", uploadResult.StoragePath),
	)

	return fp, nil
}

// Get retrieves a floor plan by ID.
func (s *FloorPlanService) Get(ctx context.Context, id uuid.UUID) (*entity.FloorPlan, error) {
	return s.repo.GetByID(ctx, id)
}

// List lists floor plans in a workspace.
func (s *FloorPlanService) List(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*entity.FloorPlan, int64, error) {
	return s.repo.ListByWorkspace(ctx, workspaceID, postgres.ListOptions{
		Limit:  limit,
		Offset: offset,
	})
}

// Update updates a floor plan.
func (s *FloorPlanService) Update(ctx context.Context, id uuid.UUID, name, description string) (*entity.FloorPlan, error) {
	fp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		fp.Name = name
	}
	if description != "" {
		fp.Description = description
	}

	if err := s.repo.Update(ctx, fp); err != nil {
		return nil, err
	}

	return fp, nil
}

// Delete deletes a floor plan.
func (s *FloorPlanService) Delete(ctx context.Context, id uuid.UUID) error {
	fp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete file from storage
	if fp.FileInfo != nil {
		_ = s.storage.DeleteFile(ctx, fp.FileInfo.StoragePath)
	}

	// Delete thumbnails
	thumbnails, _ := s.repo.GetThumbnails(ctx, id)
	for _, th := range thumbnails {
		_ = s.storage.DeleteFile(ctx, th.StoragePath)
	}

	return s.repo.Delete(ctx, id)
}

// StartRecognition starts AI recognition for a floor plan.
func (s *FloorPlanService) StartRecognition(ctx context.Context, id uuid.UUID, options RecognitionOptions) (*entity.FloorPlan, string, error) {
	fp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, "", err
	}

	if fp.FileInfo == nil {
		return nil, "", apperrors.InvalidArgument("floor_plan", "no file uploaded")
	}

	// Download file from storage
	reader, err := s.storage.DownloadFile(ctx, fp.FileInfo.StoragePath)
	if err != nil {
		return nil, "", err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", apperrors.Internal("failed to read file").WithCause(err)
	}

	// Call AI Service for recognition
	resp, err := s.aiClient.RecognizeFloorPlan(ctx, &pb.RecognizeFloorPlanRequest{
		FloorPlanId: fp.ID.String(),
		ImageData:   data,
		ImageType:   fp.FileInfo.MimeType,
		Options: &pb.RecognitionOptions{
			DetectLoadBearing: options.DetectLoadBearing,
			DetectWetZones:    options.DetectWetZones,
			DetectFurniture:   options.DetectFurniture,
			Scale:             float32(options.Scale),
			Orientation:       int32(options.Orientation),
			DetailLevel:       int32(options.DetailLevel),
		},
	})
	if err != nil {
		return nil, "", apperrors.Internal("AI recognition failed").WithCause(err)
	}

	// Update floor plan status
	jobID, _ := uuid.Parse(resp.JobId)
	fp.StartProcessing(jobID)
	if err := s.repo.Update(ctx, fp); err != nil {
		return nil, "", err
	}

	s.log.Info("recognition started",
		logger.String("floor_plan_id", fp.ID.String()),
		logger.String("job_id", resp.JobId),
	)

	return fp, resp.JobId, nil
}

// GetRecognitionStatus gets the recognition status.
func (s *FloorPlanService) GetRecognitionStatus(ctx context.Context, jobID string) (*RecognitionStatus, error) {
	resp, err := s.aiClient.GetRecognitionStatus(ctx, &pb.GetRecognitionStatusRequest{
		JobId: jobID,
	})
	if err != nil {
		return nil, apperrors.Internal("failed to get recognition status").WithCause(err)
	}

	return &RecognitionStatus{
		JobID:    resp.JobId,
		Status:   resp.Status.String(),
		Progress: int(resp.Progress),
		Error:    resp.Error,
	}, nil
}

// GetDownloadURL generates a presigned URL for downloading.
func (s *FloorPlanService) GetDownloadURL(ctx context.Context, id uuid.UUID) (string, error) {
	fp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}

	if fp.FileInfo == nil {
		return "", apperrors.InvalidArgument("floor_plan", "no file uploaded")
	}

	return s.storage.GetPresignedURL(ctx, fp.FileInfo.StoragePath, 15*60*1000000000) // 15 minutes
}

// UploadRequest for uploading a floor plan.
type UploadRequest struct {
	WorkspaceID uuid.UUID
	OwnerID     uuid.UUID
	Name        string
	Description string
	FileName    string
	MimeType    string
	FileData    io.Reader
}

// RecognitionOptions for starting recognition.
type RecognitionOptions struct {
	DetectLoadBearing bool
	DetectWetZones    bool
	DetectFurniture   bool
	Scale             float64
	Orientation       int
	DetailLevel       int
}

// RecognitionStatus represents recognition job status.
type RecognitionStatus struct {
	JobID    string
	Status   string
	Progress int
	Error    string
}

// calculateMD5 calculates MD5 checksum.
func calculateMD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

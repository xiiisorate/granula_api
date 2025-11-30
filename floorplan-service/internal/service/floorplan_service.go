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

// GetRecognitionStatus gets the recognition status with full model.
func (s *FloorPlanService) GetRecognitionStatus(ctx context.Context, jobID string) (*RecognitionStatus, error) {
	resp, err := s.aiClient.GetRecognitionStatus(ctx, &pb.GetRecognitionStatusRequest{
		JobId: jobID,
	})
	if err != nil {
		return nil, apperrors.Internal("failed to get recognition status").WithCause(err)
	}

	status := &RecognitionStatus{
		JobID:    resp.JobId,
		Status:   resp.Status.String(),
		Progress: int(resp.Progress),
		Error:    resp.Error,
	}

	// If recognition is complete, include the model
	if resp.Scene != nil {
		status.Model = convertAISceneToModel(resp.Scene)
	}

	return status, nil
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
	Model    *RecognitionModel
}

// RecognitionModel represents the recognition result model.
type RecognitionModel struct {
	Bounds           Bounds3D
	TotalArea        float64
	Confidence       float64
	Elements         SceneElements
	Recognition      RecognitionMeta
	Warnings         []string
	ProcessingTimeMs int64
}

// Bounds3D represents 3D bounds.
type Bounds3D struct {
	Width  float64
	Height float64
	Depth  float64
}

// SceneElements contains all scene elements.
type SceneElements struct {
	Walls     []Wall3D
	Rooms     []Room3D
	Furniture []Furniture
	Utilities []Utility
}

// Wall3D represents a 3D wall.
type Wall3D struct {
	ID         string
	Type       string
	Name       string
	Start      Point3D
	End        Point3D
	Height     float64
	Thickness  float64
	Properties WallProperties
	Openings   []Opening3D
	Metadata   ElementMetadata
}

// Point3D represents a 3D point.
type Point3D struct {
	X, Y, Z float64
}

// Point2D represents a 2D point.
type Point2D struct {
	X, Y float64
}

// WallProperties contains wall properties.
type WallProperties struct {
	IsLoadBearing  bool
	Material       string
	CanDemolish    bool
	StructuralType string
}

// Opening3D represents an opening in a wall.
type Opening3D struct {
	ID            string
	Type          string
	Subtype       string
	Position      float64
	Width         float64
	Height        float64
	Elevation     float64
	OpensTo       string
	HasDoor       bool
	ConnectsRooms []string
}

// Room3D represents a 3D room.
type Room3D struct {
	ID         string
	Type       string
	Name       string
	RoomType   string
	Polygon    []Point2D
	Area       float64
	Perimeter  float64
	Properties RoomProperties
	WallIDs    []string
	Metadata   RoomMetadata
}

// RoomProperties contains room properties.
type RoomProperties struct {
	HasWetZone     bool
	HasVentilation bool
	HasWindow      bool
	MinAllowedArea float64
	CeilingHeight  float64
}

// RoomMetadata contains room metadata.
type RoomMetadata struct {
	Confidence  float64
	LabelOnPlan string
	AreaOnPlan  float64
}

// Furniture represents furniture item.
type Furniture struct {
	ID            string
	Type          string
	Name          string
	FurnitureType string
	Position      Point3D
	Rotation      Rotation3D
	Dimensions    Dimensions3D
	RoomID        string
	Properties    FurnitureProperties
}

// Rotation3D represents 3D rotation.
type Rotation3D struct {
	X, Y, Z float64
}

// Dimensions3D represents 3D dimensions.
type Dimensions3D struct {
	Width, Height, Depth float64
}

// FurnitureProperties contains furniture properties.
type FurnitureProperties struct {
	CanRelocate   bool
	Category      string
	RequiresWater bool
	RequiresGas   bool
	RequiresDrain bool
}

// Utility represents a utility element.
type Utility struct {
	ID          string
	Type        string
	Name        string
	UtilityType string
	Position    Point3D
	Dimensions  UtilityDimensions
	RoomID      string
	Properties  UtilityProperties
}

// UtilityDimensions contains utility dimensions.
type UtilityDimensions struct {
	Diameter float64
	Width    float64
	Depth    float64
}

// UtilityProperties contains utility properties.
type UtilityProperties struct {
	CanRelocate         bool
	ProtectionZone      float64
	SharedWithNeighbors bool
}

// ElementMetadata contains element metadata.
type ElementMetadata struct {
	Confidence float64
	Source     string
	Locked     bool
	Visible    bool
	ModelURL   string
}

// RecognitionMeta contains recognition metadata.
type RecognitionMeta struct {
	SourceType     string
	Quality        string
	Scale          string
	Orientation    int
	HasDimensions  bool
	HasAnnotations bool
	BuildingType   string
}

// calculateMD5 calculates MD5 checksum.
func calculateMD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// convertAISceneToModel converts AI recognized scene to internal model.
func convertAISceneToModel(scene *pb.RecognizedScene) *RecognitionModel {
	if scene == nil {
		return nil
	}

	model := &RecognitionModel{
		TotalArea: float64(scene.TotalArea),
	}

	// Convert dimensions to bounds
	if scene.Dimensions != nil {
		model.Bounds = Bounds3D{
			Width:  scene.Dimensions.Width,
			Depth:  scene.Dimensions.Height,
			Height: 2.7, // Default ceiling height
		}
	}

	// Convert walls
	for _, w := range scene.Walls {
		wall := Wall3D{
			ID:        w.TempId,
			Type:      "wall",
			Thickness: float64(w.Thickness),
			Properties: WallProperties{
				IsLoadBearing: w.IsLoadBearing,
			},
			Metadata: ElementMetadata{
				Confidence: float64(w.Confidence),
			},
		}
		if w.Start != nil {
			wall.Start = Point3D{X: w.Start.X, Y: 0, Z: w.Start.Y}
		}
		if w.End != nil {
			wall.End = Point3D{X: w.End.X, Y: 0, Z: w.End.Y}
		}
		model.Elements.Walls = append(model.Elements.Walls, wall)
	}

	// Convert rooms
	for _, r := range scene.Rooms {
		room := Room3D{
			ID:       r.TempId,
			Type:     "room",
			RoomType: r.Type.String(),
			Area:     float64(r.Area),
			Properties: RoomProperties{
				HasWetZone: r.IsWetZone,
			},
			WallIDs: r.WallIds,
			Metadata: RoomMetadata{
				Confidence: float64(r.Confidence),
			},
		}
		if r.Boundary != nil {
			for _, v := range r.Boundary.Vertices {
				room.Polygon = append(room.Polygon, Point2D{X: v.X, Y: v.Y})
			}
		}
		model.Elements.Rooms = append(model.Elements.Rooms, room)
	}

	// Convert openings
	for _, o := range scene.Openings {
		// Find wall and add opening to it
		for i := range model.Elements.Walls {
			if model.Elements.Walls[i].ID == o.WallId {
				opening := Opening3D{
					ID:       o.TempId,
					Type:     o.Type.String(),
					Width:    float64(o.Width),
					Position: o.Position.X,
				}
				model.Elements.Walls[i].Openings = append(model.Elements.Walls[i].Openings, opening)
				break
			}
		}
	}

	// Set metadata
	if scene.Metadata != nil {
		model.ProcessingTimeMs = scene.Metadata.ProcessingTimeMs
	}

	// Calculate confidence
	var totalConf float64
	var count int
	for _, w := range model.Elements.Walls {
		totalConf += w.Metadata.Confidence
		count++
	}
	for _, r := range model.Elements.Rooms {
		totalConf += r.Metadata.Confidence
		count++
	}
	if count > 0 {
		model.Confidence = totalConf / float64(count)
	}

	return model
}

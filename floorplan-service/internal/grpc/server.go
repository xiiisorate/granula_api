// Package grpc provides gRPC handlers for Floor Plan Service.
package grpc

import (
	"bytes"
	"context"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/floorplan-service/internal/domain/entity"
	"github.com/xiiisorate/granula_api/floorplan-service/internal/service"
	pb "github.com/xiiisorate/granula_api/shared/gen/floorplan/v1"
	apperrors "github.com/xiiisorate/granula_api/shared/pkg/errors"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FloorPlanServer implements the gRPC Floor Plan Service.
type FloorPlanServer struct {
	pb.UnimplementedFloorPlanServiceServer
	service *service.FloorPlanService
	log     *logger.Logger
}

// NewFloorPlanServer creates a new FloorPlanServer.
func NewFloorPlanServer(svc *service.FloorPlanService, log *logger.Logger) *FloorPlanServer {
	return &FloorPlanServer{
		service: svc,
		log:     log,
	}
}

// Upload uploads a new floor plan.
func (s *FloorPlanServer) Upload(ctx context.Context, req *pb.UploadRequest) (*pb.UploadResponse, error) {
	s.log.Info("Upload called",
		logger.String("workspace_id", req.WorkspaceId),
		logger.String("name", req.Name),
	)

	workspaceID, err := uuid.Parse(req.WorkspaceId)
	if err != nil {
		return nil, apperrors.InvalidArgument("workspace_id", "invalid UUID").ToGRPCError()
	}

	ownerID, err := uuid.Parse(req.OwnerId)
	if err != nil {
		return nil, apperrors.InvalidArgument("owner_id", "invalid UUID").ToGRPCError()
	}

	fp, err := s.service.Upload(ctx, service.UploadRequest{
		WorkspaceID: workspaceID,
		OwnerID:     ownerID,
		Name:        req.Name,
		Description: req.Description,
		FileName:    req.FileName,
		MimeType:    req.MimeType,
		FileData:    bytes.NewReader(req.FileData),
	})
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.UploadResponse{
		Success:   true,
		FloorPlan: convertFloorPlanToPB(fp),
	}, nil
}

// Get retrieves a floor plan by ID.
func (s *FloorPlanServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperrors.InvalidArgument("id", "invalid UUID").ToGRPCError()
	}

	fp, err := s.service.Get(ctx, id)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.GetResponse{
		FloorPlan: convertFloorPlanToPB(fp),
	}, nil
}

// List lists floor plans in a workspace.
func (s *FloorPlanServer) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceId)
	if err != nil {
		return nil, apperrors.InvalidArgument("workspace_id", "invalid UUID").ToGRPCError()
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}

	offset := int(req.Offset)

	floorPlans, total, err := s.service.List(ctx, workspaceID, limit, offset)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	pbFloorPlans := make([]*pb.FloorPlan, 0, len(floorPlans))
	for _, fp := range floorPlans {
		pbFloorPlans = append(pbFloorPlans, convertFloorPlanToPB(fp))
	}

	return &pb.ListResponse{
		FloorPlans: pbFloorPlans,
		Total:      int32(total),
	}, nil
}

// Update updates a floor plan.
func (s *FloorPlanServer) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperrors.InvalidArgument("id", "invalid UUID").ToGRPCError()
	}

	fp, err := s.service.Update(ctx, id, req.Name, req.Description)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.UpdateResponse{
		Success:   true,
		FloorPlan: convertFloorPlanToPB(fp),
	}, nil
}

// Delete deletes a floor plan.
func (s *FloorPlanServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperrors.InvalidArgument("id", "invalid UUID").ToGRPCError()
	}

	if err := s.service.Delete(ctx, id); err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.DeleteResponse{
		Success: true,
	}, nil
}

// StartRecognition starts AI recognition for a floor plan.
func (s *FloorPlanServer) StartRecognition(ctx context.Context, req *pb.StartRecognitionRequest) (*pb.StartRecognitionResponse, error) {
	s.log.Info("StartRecognition called", logger.String("floor_plan_id", req.FloorPlanId))

	id, err := uuid.Parse(req.FloorPlanId)
	if err != nil {
		return nil, apperrors.InvalidArgument("floor_plan_id", "invalid UUID").ToGRPCError()
	}

	options := service.RecognitionOptions{
		DetectLoadBearing: req.Options.GetDetectLoadBearing(),
		DetectWetZones:    req.Options.GetDetectWetZones(),
		DetectFurniture:   req.Options.GetDetectFurniture(),
		Scale:             float64(req.Options.GetScale()),
		Orientation:       int(req.Options.GetOrientation()),
		DetailLevel:       int(req.Options.GetDetailLevel()),
	}

	_, jobID, err := s.service.StartRecognition(ctx, id, options)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.StartRecognitionResponse{
		Success: true,
		JobId:   jobID,
	}, nil
}

// GetRecognitionStatus gets the recognition status.
func (s *FloorPlanServer) GetRecognitionStatus(ctx context.Context, req *pb.GetRecognitionStatusRequest) (*pb.GetRecognitionStatusResponse, error) {
	status, err := s.service.GetRecognitionStatus(ctx, req.JobId)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.GetRecognitionStatusResponse{
		JobId:    status.JobID,
		Status:   status.Status,
		Progress: int32(status.Progress),
		Error:    status.Error,
	}, nil
}

// GetDownloadURL generates a presigned download URL.
func (s *FloorPlanServer) GetDownloadURL(ctx context.Context, req *pb.GetDownloadURLRequest) (*pb.GetDownloadURLResponse, error) {
	id, err := uuid.Parse(req.FloorPlanId)
	if err != nil {
		return nil, apperrors.InvalidArgument("floor_plan_id", "invalid UUID").ToGRPCError()
	}

	url, err := s.service.GetDownloadURL(ctx, id)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.GetDownloadURLResponse{
		Url:       url,
		ExpiresIn: 900, // 15 minutes
	}, nil
}

// convertFloorPlanToPB converts entity to protobuf.
func convertFloorPlanToPB(fp *entity.FloorPlan) *pb.FloorPlan {
	pbFP := &pb.FloorPlan{
		Id:          fp.ID.String(),
		WorkspaceId: fp.WorkspaceID.String(),
		OwnerId:     fp.OwnerID.String(),
		Name:        fp.Name,
		Description: fp.Description,
		Status:      convertStatusToPB(fp.Status),
		CreatedAt:   timestamppb.New(fp.CreatedAt),
		UpdatedAt:   timestamppb.New(fp.UpdatedAt),
	}

	if fp.RecognitionJobID != nil {
		jobID := fp.RecognitionJobID.String()
		pbFP.RecognitionJobId = &jobID
	}

	if fp.SceneID != nil {
		sceneID := fp.SceneID.String()
		pbFP.SceneId = &sceneID
	}

	if fp.FileInfo != nil {
		pbFP.FileInfo = &pb.FileInfo{
			Id:           fp.FileInfo.ID.String(),
			OriginalName: fp.FileInfo.OriginalName,
			MimeType:     fp.FileInfo.MimeType,
			Size:         fp.FileInfo.Size,
			Width:        int32(fp.FileInfo.Width),
			Height:       int32(fp.FileInfo.Height),
		}
	}

	return pbFP
}

func convertStatusToPB(status entity.FloorPlanStatus) pb.FloorPlanStatus {
	switch status {
	case entity.FloorPlanStatusUploaded:
		return pb.FloorPlanStatus_FLOOR_PLAN_STATUS_UPLOADED
	case entity.FloorPlanStatusProcessing:
		return pb.FloorPlanStatus_FLOOR_PLAN_STATUS_PROCESSING
	case entity.FloorPlanStatusRecognized:
		return pb.FloorPlanStatus_FLOOR_PLAN_STATUS_RECOGNIZED
	case entity.FloorPlanStatusConfirmed:
		return pb.FloorPlanStatus_FLOOR_PLAN_STATUS_CONFIRMED
	case entity.FloorPlanStatusFailed:
		return pb.FloorPlanStatus_FLOOR_PLAN_STATUS_FAILED
	default:
		return pb.FloorPlanStatus_FLOOR_PLAN_STATUS_UNSPECIFIED
	}
}


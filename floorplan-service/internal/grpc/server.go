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

	resp := &pb.GetRecognitionStatusResponse{
		JobId:    status.JobID,
		Status:   status.Status,
		Progress: int32(status.Progress),
		Error:    status.Error,
	}

	// Include model if available
	if status.Model != nil {
		resp.Model = convertModelToPB(status.Model)
	}

	return resp, nil
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

// convertModelToPB converts service model to proto.
func convertModelToPB(model *service.RecognitionModel) *pb.RecognitionModel {
	if model == nil {
		return nil
	}

	pbModel := &pb.RecognitionModel{
		Bounds: &pb.Bounds3D{
			Width:  model.Bounds.Width,
			Height: model.Bounds.Height,
			Depth:  model.Bounds.Depth,
		},
		TotalArea:        model.TotalArea,
		Confidence:       model.Confidence,
		Warnings:         model.Warnings,
		ProcessingTimeMs: model.ProcessingTimeMs,
		Elements:         &pb.SceneElements{},
		Recognition: &pb.RecognitionMeta{
			SourceType:     model.Recognition.SourceType,
			Quality:        model.Recognition.Quality,
			Scale:          model.Recognition.Scale,
			Orientation:    int32(model.Recognition.Orientation),
			HasDimensions:  model.Recognition.HasDimensions,
			HasAnnotations: model.Recognition.HasAnnotations,
			BuildingType:   model.Recognition.BuildingType,
		},
	}

	// Convert walls
	for _, w := range model.Elements.Walls {
		pbWall := &pb.Wall3D{
			Id:        w.ID,
			Type:      w.Type,
			Name:      w.Name,
			Start:     &pb.Point3D{X: w.Start.X, Y: w.Start.Y, Z: w.Start.Z},
			End:       &pb.Point3D{X: w.End.X, Y: w.End.Y, Z: w.End.Z},
			Height:    w.Height,
			Thickness: w.Thickness,
			Properties: &pb.WallProperties{
				IsLoadBearing:  w.Properties.IsLoadBearing,
				Material:       w.Properties.Material,
				CanDemolish:    w.Properties.CanDemolish,
				StructuralType: w.Properties.StructuralType,
			},
			Metadata: &pb.ElementMetadata{
				Confidence: w.Metadata.Confidence,
				Source:     w.Metadata.Source,
				Locked:     w.Metadata.Locked,
				Visible:    w.Metadata.Visible,
				ModelUrl:   w.Metadata.ModelURL,
			},
		}
		for _, o := range w.Openings {
			pbWall.Openings = append(pbWall.Openings, &pb.Opening3D{
				Id:            o.ID,
				Type:          o.Type,
				Subtype:       o.Subtype,
				Position:      o.Position,
				Width:         o.Width,
				Height:        o.Height,
				Elevation:     o.Elevation,
				OpensTo:       o.OpensTo,
				HasDoor:       o.HasDoor,
				ConnectsRooms: o.ConnectsRooms,
			})
		}
		pbModel.Elements.Walls = append(pbModel.Elements.Walls, pbWall)
	}

	// Convert rooms
	for _, r := range model.Elements.Rooms {
		pbRoom := &pb.Room3D{
			Id:        r.ID,
			Type:      r.Type,
			Name:      r.Name,
			RoomType:  r.RoomType,
			Area:      r.Area,
			Perimeter: r.Perimeter,
			WallIds:   r.WallIDs,
			Properties: &pb.RoomProperties{
				HasWetZone:     r.Properties.HasWetZone,
				HasVentilation: r.Properties.HasVentilation,
				HasWindow:      r.Properties.HasWindow,
				MinAllowedArea: r.Properties.MinAllowedArea,
				CeilingHeight:  r.Properties.CeilingHeight,
			},
			RoomMetadata: &pb.RoomMetadata{
				Confidence:  r.Metadata.Confidence,
				LabelOnPlan: r.Metadata.LabelOnPlan,
				AreaOnPlan:  r.Metadata.AreaOnPlan,
			},
		}
		for _, p := range r.Polygon {
			pbRoom.Polygon = append(pbRoom.Polygon, &pb.Point2D{X: p.X, Y: p.Y})
		}
		pbModel.Elements.Rooms = append(pbModel.Elements.Rooms, pbRoom)
	}

	// Convert furniture
	for _, f := range model.Elements.Furniture {
		pbModel.Elements.Furniture = append(pbModel.Elements.Furniture, &pb.Furniture{
			Id:            f.ID,
			Type:          f.Type,
			Name:          f.Name,
			FurnitureType: f.FurnitureType,
			Position:      &pb.Point3D{X: f.Position.X, Y: f.Position.Y, Z: f.Position.Z},
			Rotation:      &pb.Rotation3D{X: f.Rotation.X, Y: f.Rotation.Y, Z: f.Rotation.Z},
			Dimensions:    &pb.Dimensions3D{Width: f.Dimensions.Width, Height: f.Dimensions.Height, Depth: f.Dimensions.Depth},
			RoomId:        f.RoomID,
			Properties: &pb.FurnitureProperties{
				CanRelocate:   f.Properties.CanRelocate,
				Category:      f.Properties.Category,
				RequiresWater: f.Properties.RequiresWater,
				RequiresGas:   f.Properties.RequiresGas,
				RequiresDrain: f.Properties.RequiresDrain,
			},
		})
	}

	// Convert utilities
	for _, u := range model.Elements.Utilities {
		pbModel.Elements.Utilities = append(pbModel.Elements.Utilities, &pb.Utility{
			Id:          u.ID,
			Type:        u.Type,
			Name:        u.Name,
			UtilityType: u.UtilityType,
			Position:    &pb.Point3D{X: u.Position.X, Y: u.Position.Y, Z: u.Position.Z},
			Dimensions: &pb.UtilityDimensions{
				Diameter: u.Dimensions.Diameter,
				Width:    u.Dimensions.Width,
				Depth:    u.Dimensions.Depth,
			},
			RoomId: u.RoomID,
			Properties: &pb.UtilityProperties{
				CanRelocate:         u.Properties.CanRelocate,
				ProtectionZone:      u.Properties.ProtectionZone,
				SharedWithNeighbors: u.Properties.SharedWithNeighbors,
			},
		})
	}

	return pbModel
}


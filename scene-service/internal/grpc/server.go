// Package grpc provides gRPC handlers for Scene Service.
package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/scene-service/internal/domain/entity"
	"github.com/xiiisorate/granula_api/scene-service/internal/service"
	commonpb "github.com/xiiisorate/granula_api/shared/gen/common/v1"
	pb "github.com/xiiisorate/granula_api/shared/gen/scene/v1"
	apperrors "github.com/xiiisorate/granula_api/shared/pkg/errors"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SceneServer implements the gRPC Scene Service.
type SceneServer struct {
	pb.UnimplementedSceneServiceServer
	service *service.SceneService
	log     *logger.Logger
}

// NewSceneServer creates a new SceneServer.
func NewSceneServer(svc *service.SceneService, log *logger.Logger) *SceneServer {
	return &SceneServer{
		service: svc,
		log:     log,
	}
}

// CreateScene creates a new scene.
func (s *SceneServer) CreateScene(ctx context.Context, req *pb.CreateSceneRequest) (*pb.CreateSceneResponse, error) {
	s.log.Info("CreateScene called", logger.String("name", req.Name))

	workspaceID, err := uuid.Parse(req.WorkspaceId)
	if err != nil {
		return nil, apperrors.InvalidArgument("workspace_id", "invalid UUID").ToGRPCError()
	}

	ownerID, err := uuid.Parse(req.OwnerId)
	if err != nil {
		return nil, apperrors.InvalidArgument("owner_id", "invalid UUID").ToGRPCError()
	}

	createReq := service.CreateSceneRequest{
		WorkspaceID: workspaceID,
		OwnerID:     ownerID,
		Name:        req.Name,
		Description: req.Description,
	}

	if req.FloorPlanId != "" {
		fpID, err := uuid.Parse(req.FloorPlanId)
		if err != nil {
			return nil, apperrors.InvalidArgument("floor_plan_id", "invalid UUID").ToGRPCError()
		}
		createReq.FloorPlanID = &fpID
	}

	if req.Dimensions != nil {
		createReq.Dimensions = entity.Dimensions3D{
			Width:  req.Dimensions.Width,
			Depth:  req.Dimensions.Height, // 2D height maps to 3D depth
			Height: 2.7,                   // Default room height
		}
	}

	scene, err := s.service.CreateScene(ctx, createReq)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.CreateSceneResponse{
		Success: true,
		Scene:   convertSceneToPB(scene),
	}, nil
}

// GetScene retrieves a scene by ID.
func (s *SceneServer) GetScene(ctx context.Context, req *pb.GetSceneRequest) (*pb.GetSceneResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperrors.InvalidArgument("id", "invalid UUID").ToGRPCError()
	}

	scene, err := s.service.GetScene(ctx, id)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.GetSceneResponse{
		Scene: convertSceneToPB(scene),
	}, nil
}

// UpdateScene updates a scene.
func (s *SceneServer) UpdateScene(ctx context.Context, req *pb.UpdateSceneRequest) (*pb.UpdateSceneResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperrors.InvalidArgument("id", "invalid UUID").ToGRPCError()
	}

	scene, err := s.service.UpdateScene(ctx, id, req.Name, req.Description)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.UpdateSceneResponse{
		Success: true,
		Scene:   convertSceneToPB(scene),
	}, nil
}

// DeleteScene deletes a scene.
func (s *SceneServer) DeleteScene(ctx context.Context, req *pb.DeleteSceneRequest) (*pb.DeleteSceneResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperrors.InvalidArgument("id", "invalid UUID").ToGRPCError()
	}

	if err := s.service.DeleteScene(ctx, id); err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.DeleteSceneResponse{
		Success: true,
	}, nil
}

// ListScenes lists scenes in a workspace.
func (s *SceneServer) ListScenes(ctx context.Context, req *pb.ListScenesRequest) (*pb.ListScenesResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceId)
	if err != nil {
		return nil, apperrors.InvalidArgument("workspace_id", "invalid UUID").ToGRPCError()
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}

	scenes, total, err := s.service.ListScenes(ctx, workspaceID, limit, int(req.Offset))
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	pbScenes := make([]*pb.Scene, 0, len(scenes))
	for _, scene := range scenes {
		pbScenes = append(pbScenes, convertSceneToPB(scene))
	}

	return &pb.ListScenesResponse{
		Scenes: pbScenes,
		Total:  int32(total),
	}, nil
}

// CreateElement creates a new element.
func (s *SceneServer) CreateElement(ctx context.Context, req *pb.CreateElementRequest) (*pb.CreateElementResponse, error) {
	sceneID, err := uuid.Parse(req.SceneId)
	if err != nil {
		return nil, apperrors.InvalidArgument("scene_id", "invalid UUID").ToGRPCError()
	}

	branchID, err := uuid.Parse(req.BranchId)
	if err != nil {
		return nil, apperrors.InvalidArgument("branch_id", "invalid UUID").ToGRPCError()
	}

	createReq := service.CreateElementRequest{
		SceneID:  sceneID,
		BranchID: branchID,
		Type:     convertElementTypeFromPB(req.Type),
		Name:     req.Name,
	}

	if req.Position != nil {
		createReq.Position = entity.Point3D{X: req.Position.X, Y: req.Position.Y, Z: req.Position.Z}
	}
	if req.Rotation != nil {
		createReq.Rotation = entity.Rotation3D{X: req.Rotation.X, Y: req.Rotation.Y, Z: req.Rotation.Z}
	}
	if req.Dimensions != nil {
		createReq.Dimensions = entity.Dimensions3D{
			Width:  req.Dimensions.Width,
			Depth:  req.Dimensions.Height,
			Height: req.Dimensions.Depth,
		}
	}

	if req.ParentId != "" {
		parentID, err := uuid.Parse(req.ParentId)
		if err != nil {
			return nil, apperrors.InvalidArgument("parent_id", "invalid UUID").ToGRPCError()
		}
		createReq.ParentID = &parentID
	}

	elem, err := s.service.CreateElement(ctx, createReq)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.CreateElementResponse{
		Success: true,
		Element: convertElementToPB(elem),
	}, nil
}

// GetElement retrieves an element.
func (s *SceneServer) GetElement(ctx context.Context, req *pb.GetElementRequest) (*pb.GetElementResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperrors.InvalidArgument("id", "invalid UUID").ToGRPCError()
	}

	elem, err := s.service.GetElement(ctx, id)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.GetElementResponse{
		Element: convertElementToPB(elem),
	}, nil
}

// UpdateElement updates an element.
func (s *SceneServer) UpdateElement(ctx context.Context, req *pb.UpdateElementRequest) (*pb.UpdateElementResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperrors.InvalidArgument("id", "invalid UUID").ToGRPCError()
	}

	updateReq := service.UpdateElementRequest{
		Name: req.Name,
	}

	if req.Position != nil {
		pos := entity.Point3D{X: req.Position.X, Y: req.Position.Y, Z: req.Position.Z}
		updateReq.Position = &pos
	}

	elem, err := s.service.UpdateElement(ctx, id, updateReq)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.UpdateElementResponse{
		Success: true,
		Element: convertElementToPB(elem),
	}, nil
}

// DeleteElement deletes an element.
func (s *SceneServer) DeleteElement(ctx context.Context, req *pb.DeleteElementRequest) (*pb.DeleteElementResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperrors.InvalidArgument("id", "invalid UUID").ToGRPCError()
	}

	if err := s.service.DeleteElement(ctx, id); err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.DeleteElementResponse{
		Success: true,
	}, nil
}

// ListElements lists elements in a branch.
func (s *SceneServer) ListElements(ctx context.Context, req *pb.ListElementsRequest) (*pb.ListElementsResponse, error) {
	branchID, err := uuid.Parse(req.BranchId)
	if err != nil {
		return nil, apperrors.InvalidArgument("branch_id", "invalid UUID").ToGRPCError()
	}

	var elemType string
	if req.Type != pb.ElementType_ELEMENT_TYPE_UNSPECIFIED {
		elemType = string(convertElementTypeFromPB(req.Type))
	}

	elements, err := s.service.ListElements(ctx, branchID, elemType, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	pbElements := make([]*pb.Element, 0, len(elements))
	for _, elem := range elements {
		pbElements = append(pbElements, convertElementToPB(elem))
	}

	return &pb.ListElementsResponse{
		Elements: pbElements,
	}, nil
}

// CheckCompliance checks scene compliance.
func (s *SceneServer) CheckCompliance(ctx context.Context, req *pb.CheckComplianceRequest) (*pb.CheckComplianceResponse, error) {
	sceneID, err := uuid.Parse(req.SceneId)
	if err != nil {
		return nil, apperrors.InvalidArgument("scene_id", "invalid UUID").ToGRPCError()
	}

	branchID, err := uuid.Parse(req.BranchId)
	if err != nil {
		return nil, apperrors.InvalidArgument("branch_id", "invalid UUID").ToGRPCError()
	}

	result, err := s.service.CheckCompliance(ctx, sceneID, branchID)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	violations := make([]*pb.Violation, 0, len(result.Violations))
	for _, v := range result.Violations {
		violations = append(violations, &pb.Violation{
			Id:          v.ID,
			RuleId:      v.RuleID,
			Severity:    v.Severity,
			Title:       v.Title,
			Description: v.Description,
			ElementIds:  v.ElementIDs,
		})
	}

	return &pb.CheckComplianceResponse{
		IsCompliant:  result.IsCompliant,
		Violations:   violations,
		TotalChecks:  int32(result.TotalChecks),
		PassedChecks: int32(result.PassedChecks),
		FailedChecks: int32(result.FailedChecks),
	}, nil
}

// Conversion helpers

func convertSceneToPB(scene *entity.Scene) *pb.Scene {
	pbScene := &pb.Scene{
		Id:           scene.ID.String(),
		WorkspaceId:  scene.WorkspaceID.String(),
		OwnerId:      scene.OwnerID.String(),
		Name:         scene.Name,
		Description:  scene.Description,
		MainBranchId: scene.MainBranchID.String(),
		Dimensions: &commonpb.Dimensions2D{
			Width:  scene.Dimensions.Width,
			Height: scene.Dimensions.Depth,
		},
		CreatedAt: timestamppb.New(scene.CreatedAt),
		UpdatedAt: timestamppb.New(scene.UpdatedAt),
	}

	if scene.FloorPlanID != nil {
		fpID := scene.FloorPlanID.String()
		pbScene.FloorPlanId = fpID
	}

	return pbScene
}

func convertElementToPB(elem *entity.Element) *pb.Element {
	pbElem := &pb.Element{
		Id:       elem.ID.String(),
		SceneId:  elem.SceneID.String(),
		BranchId: elem.BranchID.String(),
		Type:     convertElementTypeToPB(elem.Type),
		Name:     elem.Name,
		Position: &commonpb.Point3D{
			X: elem.Position.X,
			Y: elem.Position.Y,
			Z: elem.Position.Z,
		},
		Rotation: &commonpb.Rotation3D{
			X: elem.Rotation.X,
			Y: elem.Rotation.Y,
			Z: elem.Rotation.Z,
		},
		Dimensions: &commonpb.Dimensions3D{
			Width:  elem.Dimensions.Width,
			Height: elem.Dimensions.Height,
			Depth:  elem.Dimensions.Depth,
		},
		Version:   elem.Version,
		CreatedAt: timestamppb.New(elem.CreatedAt),
		UpdatedAt: timestamppb.New(elem.UpdatedAt),
	}

	if elem.ParentID != nil {
		parentID := elem.ParentID.String()
		pbElem.ParentId = parentID
	}

	return pbElem
}

func convertElementTypeToPB(t entity.ElementType) pb.ElementType {
	switch t {
	case entity.ElementTypeWall:
		return pb.ElementType_ELEMENT_TYPE_WALL
	case entity.ElementTypeRoom:
		return pb.ElementType_ELEMENT_TYPE_ROOM
	case entity.ElementTypeDoor:
		return pb.ElementType_ELEMENT_TYPE_DOOR
	case entity.ElementTypeWindow:
		return pb.ElementType_ELEMENT_TYPE_WINDOW
	case entity.ElementTypeFurniture:
		return pb.ElementType_ELEMENT_TYPE_FURNITURE
	case entity.ElementTypeFixture:
		return pb.ElementType_ELEMENT_TYPE_FIXTURE
	case entity.ElementTypeDecor:
		return pb.ElementType_ELEMENT_TYPE_DECOR
	default:
		return pb.ElementType_ELEMENT_TYPE_UNSPECIFIED
	}
}

func convertElementTypeFromPB(t pb.ElementType) entity.ElementType {
	switch t {
	case pb.ElementType_ELEMENT_TYPE_WALL:
		return entity.ElementTypeWall
	case pb.ElementType_ELEMENT_TYPE_ROOM:
		return entity.ElementTypeRoom
	case pb.ElementType_ELEMENT_TYPE_DOOR:
		return entity.ElementTypeDoor
	case pb.ElementType_ELEMENT_TYPE_WINDOW:
		return entity.ElementTypeWindow
	case pb.ElementType_ELEMENT_TYPE_FURNITURE:
		return entity.ElementTypeFurniture
	case pb.ElementType_ELEMENT_TYPE_FIXTURE:
		return entity.ElementTypeFixture
	case pb.ElementType_ELEMENT_TYPE_DECOR:
		return entity.ElementTypeDecor
	default:
		return entity.ElementTypeFurniture
	}
}



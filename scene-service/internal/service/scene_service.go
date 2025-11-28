// Package service provides business logic for Scene Service.
package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/scene-service/internal/domain/entity"
	"github.com/xiiisorate/granula_api/scene-service/internal/repository/mongodb"
	compliancepb "github.com/xiiisorate/granula_api/shared/gen/compliance/v1"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
)

// SceneService handles scene operations.
type SceneService struct {
	sceneRepo        *mongodb.SceneRepository
	elementRepo      *mongodb.ElementRepository
	complianceClient compliancepb.ComplianceServiceClient
	log              *logger.Logger
}

// NewSceneService creates a new SceneService.
func NewSceneService(
	sceneRepo *mongodb.SceneRepository,
	elementRepo *mongodb.ElementRepository,
	complianceClient compliancepb.ComplianceServiceClient,
	log *logger.Logger,
) *SceneService {
	return &SceneService{
		sceneRepo:        sceneRepo,
		elementRepo:      elementRepo,
		complianceClient: complianceClient,
		log:              log,
	}
}

// CreateScene creates a new scene.
func (s *SceneService) CreateScene(ctx context.Context, req CreateSceneRequest) (*entity.Scene, error) {
	s.log.Info("creating scene",
		logger.String("workspace_id", req.WorkspaceID.String()),
		logger.String("name", req.Name),
	)

	scene := entity.NewScene(req.WorkspaceID, req.OwnerID, req.Name)
	scene.Description = req.Description
	scene.Dimensions = req.Dimensions

	if req.FloorPlanID != nil {
		scene.FloorPlanID = req.FloorPlanID
	}

	if err := s.sceneRepo.Create(ctx, scene); err != nil {
		return nil, err
	}

	s.log.Info("scene created",
		logger.String("id", scene.ID.String()),
		logger.String("main_branch_id", scene.MainBranchID.String()),
	)

	return scene, nil
}

// GetScene retrieves a scene by ID.
func (s *SceneService) GetScene(ctx context.Context, id uuid.UUID) (*entity.Scene, error) {
	return s.sceneRepo.GetByID(ctx, id)
}

// UpdateScene updates a scene.
func (s *SceneService) UpdateScene(ctx context.Context, id uuid.UUID, name, description string) (*entity.Scene, error) {
	scene, err := s.sceneRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		scene.Name = name
	}
	if description != "" {
		scene.Description = description
	}

	if err := s.sceneRepo.Update(ctx, scene); err != nil {
		return nil, err
	}

	return scene, nil
}

// DeleteScene deletes a scene and all its elements.
func (s *SceneService) DeleteScene(ctx context.Context, id uuid.UUID) error {
	// TODO: Delete all elements, branches, etc.
	return s.sceneRepo.Delete(ctx, id)
}

// ListScenes lists scenes in a workspace.
func (s *SceneService) ListScenes(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*entity.Scene, int64, error) {
	return s.sceneRepo.ListByWorkspace(ctx, workspaceID, limit, offset)
}

// CreateElement creates a new element in a scene.
func (s *SceneService) CreateElement(ctx context.Context, req CreateElementRequest) (*entity.Element, error) {
	elem := entity.NewElement(req.SceneID, req.BranchID, req.Type, req.Name)
	elem.Position = req.Position
	elem.Rotation = req.Rotation
	elem.Dimensions = req.Dimensions
	elem.Properties = req.Properties

	if req.ParentID != nil {
		elem.ParentID = req.ParentID
	}

	if err := s.elementRepo.Create(ctx, elem); err != nil {
		return nil, err
	}

	return elem, nil
}

// GetElement retrieves an element by ID.
func (s *SceneService) GetElement(ctx context.Context, id uuid.UUID) (*entity.Element, error) {
	return s.elementRepo.GetByID(ctx, id)
}

// UpdateElement updates an element.
func (s *SceneService) UpdateElement(ctx context.Context, id uuid.UUID, req UpdateElementRequest) (*entity.Element, error) {
	elem, err := s.elementRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		elem.Name = req.Name
	}
	if req.Position != nil {
		elem.Position = *req.Position
	}
	if req.Rotation != nil {
		elem.Rotation = *req.Rotation
	}
	if req.Dimensions != nil {
		elem.Dimensions = *req.Dimensions
	}
	if req.Properties != nil {
		elem.Properties = *req.Properties
	}

	if err := s.elementRepo.Update(ctx, elem); err != nil {
		return nil, err
	}

	return elem, nil
}

// DeleteElement deletes an element.
func (s *SceneService) DeleteElement(ctx context.Context, id uuid.UUID) error {
	return s.elementRepo.Delete(ctx, id)
}

// ListElements lists elements in a branch.
func (s *SceneService) ListElements(ctx context.Context, branchID uuid.UUID, elemType string, limit, offset int) ([]*entity.Element, error) {
	return s.elementRepo.ListByBranch(ctx, branchID, mongodb.ListOptions{
		Type:   elemType,
		Limit:  limit,
		Offset: offset,
	})
}

// CheckCompliance checks scene against building codes.
func (s *SceneService) CheckCompliance(ctx context.Context, sceneID, branchID uuid.UUID) (*ComplianceResult, error) {
	s.log.Info("checking compliance",
		logger.String("scene_id", sceneID.String()),
		logger.String("branch_id", branchID.String()),
	)

	// Call Compliance Service
	// Note: Compliance Service will fetch scene data internally or via another call
	resp, err := s.complianceClient.CheckCompliance(ctx, &compliancepb.CheckComplianceRequest{
		SceneId:  sceneID.String(),
		BranchId: branchID.String(),
	})
	if err != nil {
		s.log.Error("compliance check failed", logger.Err(err))
		return nil, err
	}

	// Convert response
	violations := make([]Violation, 0, len(resp.Violations))
	for _, v := range resp.Violations {
		violations = append(violations, Violation{
			ID:          v.Id,
			RuleID:      v.RuleId,
			Severity:    v.Severity.String(),
			Title:       v.Title,
			Description: v.Description,
			ElementIDs:  []string{v.ElementId},
		})
	}

	var totalChecks, passedChecks, failedChecks int
	if resp.Stats != nil {
		totalChecks = int(resp.Stats.TotalRulesChecked)
		failedChecks = int(resp.Stats.ErrorsCount + resp.Stats.WarningsCount)
		passedChecks = totalChecks - failedChecks
	}

	return &ComplianceResult{
		IsCompliant:  resp.Compliant,
		Violations:   violations,
		CheckedAt:    resp.CheckedAt.AsTime(),
		TotalChecks:  totalChecks,
		PassedChecks: passedChecks,
		FailedChecks: failedChecks,
	}, nil
}

// CreateSceneRequest for creating a scene.
type CreateSceneRequest struct {
	WorkspaceID uuid.UUID
	OwnerID     uuid.UUID
	FloorPlanID *uuid.UUID
	Name        string
	Description string
	Dimensions  entity.Dimensions3D
}

// CreateElementRequest for creating an element.
type CreateElementRequest struct {
	SceneID    uuid.UUID
	BranchID   uuid.UUID
	Type       entity.ElementType
	Name       string
	Position   entity.Point3D
	Rotation   entity.Rotation3D
	Dimensions entity.Dimensions3D
	Properties entity.ElementProperties
	ParentID   *uuid.UUID
}

// UpdateElementRequest for updating an element.
type UpdateElementRequest struct {
	Name       string
	Position   *entity.Point3D
	Rotation   *entity.Rotation3D
	Dimensions *entity.Dimensions3D
	Properties *entity.ElementProperties
}

// ComplianceResult from compliance check.
type ComplianceResult struct {
	IsCompliant  bool
	Violations   []Violation
	CheckedAt    interface{}
	TotalChecks  int
	PassedChecks int
	FailedChecks int
}

// Violation from compliance check.
type Violation struct {
	ID          string
	RuleID      string
	Severity    string
	Title       string
	Description string
	ElementIDs  []string
}



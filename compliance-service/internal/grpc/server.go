// Package grpc provides gRPC handlers for Compliance Service.
package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/compliance-service/internal/domain/entity"
	"github.com/xiiisorate/granula_api/compliance-service/internal/engine"
	"github.com/xiiisorate/granula_api/compliance-service/internal/service"
	commonpb "github.com/xiiisorate/granula_api/shared/gen/common/v1"
	pb "github.com/xiiisorate/granula_api/shared/gen/compliance/v1"
	apperrors "github.com/xiiisorate/granula_api/shared/pkg/errors"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ComplianceServer implements the gRPC Compliance Service.
type ComplianceServer struct {
	pb.UnimplementedComplianceServiceServer
	service *service.ComplianceService
	log     *logger.Logger
}

// NewComplianceServer creates a new ComplianceServer.
func NewComplianceServer(svc *service.ComplianceService, log *logger.Logger) *ComplianceServer {
	return &ComplianceServer{
		service: svc,
		log:     log,
	}
}

// CheckCompliance performs a full compliance check on a scene.
func (s *ComplianceServer) CheckCompliance(ctx context.Context, req *pb.CheckComplianceRequest) (*pb.CheckComplianceResponse, error) {
	s.log.Info("CheckCompliance called",
		logger.String("scene_id", req.SceneId),
	)

	// For now, we create a minimal scene data
	// In production, this would fetch from Scene Service
	sceneData := &engine.SceneData{
		ID: req.SceneId,
	}

	result, err := s.service.CheckCompliance(ctx, sceneData)
	if err != nil {
		s.log.Error("CheckCompliance failed", logger.Err(err))
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.CheckComplianceResponse{
		Compliant:    result.Compliant,
		Violations:   convertViolationsToPB(result.Violations),
		Stats:        convertStatsToPB(result.Stats),
		CheckedAt:    timestamppb.Now(),
		RulesVersion: result.RulesVersion,
	}, nil
}

// CheckOperation checks if a specific operation is allowed.
func (s *ComplianceServer) CheckOperation(ctx context.Context, req *pb.CheckOperationRequest) (*pb.CheckOperationResponse, error) {
	s.log.Info("CheckOperation called",
		logger.String("scene_id", req.SceneId),
		logger.String("operation_type", req.Operation.Type.String()),
	)

	sceneData := &engine.SceneData{
		ID: req.SceneId,
	}

	operation := convertOperationFromPB(req.Operation)

	result, err := s.service.CheckOperation(ctx, sceneData, operation)
	if err != nil {
		s.log.Error("CheckOperation failed", logger.Err(err))
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	// Separate violations and warnings
	violations := make([]*pb.Violation, 0)
	warnings := make([]*pb.Violation, 0)

	for _, v := range result.Violations {
		pbViolation := convertViolationToPB(v)
		if v.Severity == entity.SeverityError {
			violations = append(violations, pbViolation)
		} else {
			warnings = append(warnings, pbViolation)
		}
	}

	// Determine approval type based on violations
	approvalType := pb.ApprovalType_APPROVAL_TYPE_NONE
	if len(violations) > 0 {
		for _, v := range result.Violations {
			switch v.ApprovalRequired {
			case entity.ApprovalTypeProhibited:
				approvalType = pb.ApprovalType_APPROVAL_TYPE_PROHIBITED
			case entity.ApprovalTypeExpertise:
				if approvalType != pb.ApprovalType_APPROVAL_TYPE_PROHIBITED {
					approvalType = pb.ApprovalType_APPROVAL_TYPE_EXPERTISE
				}
			case entity.ApprovalTypeProject:
				if approvalType != pb.ApprovalType_APPROVAL_TYPE_PROHIBITED &&
					approvalType != pb.ApprovalType_APPROVAL_TYPE_EXPERTISE {
					approvalType = pb.ApprovalType_APPROVAL_TYPE_PROJECT
				}
			case entity.ApprovalTypeNotification:
				if approvalType == pb.ApprovalType_APPROVAL_TYPE_NONE {
					approvalType = pb.ApprovalType_APPROVAL_TYPE_NOTIFICATION
				}
			}
		}
	}

	return &pb.CheckOperationResponse{
		Allowed:          result.Compliant,
		Violations:       violations,
		Warnings:         warnings,
		RequiresApproval: approvalType != pb.ApprovalType_APPROVAL_TYPE_NONE,
		ApprovalType:     approvalType,
	}, nil
}

// GetRules returns rules with optional filtering.
func (s *ComplianceServer) GetRules(ctx context.Context, req *pb.GetRulesRequest) (*pb.GetRulesResponse, error) {
	opts := service.GetRulesOptions{
		ActiveOnly: req.ActiveOnly,
	}

	if req.Category != "" {
		opts.Category = entity.RuleCategory(req.Category)
	}

	if req.Severity != pb.Severity_SEVERITY_UNSPECIFIED {
		opts.Severity = convertSeverityFromPB(req.Severity)
	}

	if req.Pagination != nil {
		opts.Limit = int(req.Pagination.PageSize)
		opts.Offset = (int(req.Pagination.Page) - 1) * int(req.Pagination.PageSize)
		if opts.Limit == 0 {
			opts.Limit = 20
		}
		if opts.Offset < 0 {
			opts.Offset = 0
		}
	} else {
		opts.Limit = 20
	}

	rules, total, err := s.service.GetRules(ctx, opts)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.GetRulesResponse{
		Rules: convertRulesToPB(rules),
		Pagination: &commonpb.PaginationResponse{
			Total:      int32(total),
			Page:       req.Pagination.GetPage(),
			PageSize:   int32(opts.Limit),
			TotalPages: int32((total + opts.Limit - 1) / opts.Limit),
			HasNext:    opts.Offset+len(rules) < total,
			HasPrev:    opts.Offset > 0,
		},
	}, nil
}

// GetRule returns a single rule by ID.
func (s *ComplianceServer) GetRule(ctx context.Context, req *pb.GetRuleRequest) (*pb.Rule, error) {
	ruleID, err := uuid.Parse(req.RuleId)
	if err != nil {
		return nil, apperrors.InvalidArgument("rule_id", "invalid UUID format").ToGRPCError()
	}

	rule, err := s.service.GetRule(ctx, ruleID)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return convertRuleToPB(rule), nil
}

// GetRuleCategories returns all rule categories.
func (s *ComplianceServer) GetRuleCategories(ctx context.Context, req *pb.GetRuleCategoriesRequest) (*pb.GetRuleCategoriesResponse, error) {
	categories, err := s.service.GetCategories(ctx)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	pbCategories := make([]*pb.RuleCategory, 0, len(categories))
	for _, cat := range categories {
		pbCategories = append(pbCategories, &pb.RuleCategory{
			Id:          string(cat.ID),
			Name:        cat.Name,
			Description: cat.Description,
			Icon:        cat.Icon,
			RulesCount:  int32(cat.RulesCount),
		})
	}

	return &pb.GetRuleCategoriesResponse{
		Categories: pbCategories,
	}, nil
}

// ValidateScene performs a quick validation.
func (s *ComplianceServer) ValidateScene(ctx context.Context, req *pb.ValidateSceneRequest) (*pb.ValidateSceneResponse, error) {
	sceneData := &engine.SceneData{
		ID: req.SceneId,
	}

	result, err := s.service.ValidateScene(ctx, sceneData)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.ValidateSceneResponse{
		Valid:          result.Valid,
		CriticalErrors: convertViolationsToPB(result.CriticalErrors),
		WarningsCount:  int32(result.WarningsCount),
	}, nil
}

// GenerateReport generates a compliance report.
func (s *ComplianceServer) GenerateReport(ctx context.Context, req *pb.GenerateReportRequest) (*pb.GenerateReportResponse, error) {
	// TODO: Implement PDF/JSON report generation
	return nil, apperrors.Internal("report generation not implemented").ToGRPCError()
}

// =============================================================================
// Conversion helpers
// =============================================================================

func convertViolationsToPB(violations []*entity.Violation) []*pb.Violation {
	result := make([]*pb.Violation, 0, len(violations))
	for _, v := range violations {
		result = append(result, convertViolationToPB(v))
	}
	return result
}

func convertViolationToPB(v *entity.Violation) *pb.Violation {
	pbViolation := &pb.Violation{
		Id:          v.ID.String(),
		RuleId:      v.RuleID.String(),
		RuleCode:    v.RuleCode,
		Severity:    convertSeverityToPB(v.Severity),
		Category:    string(v.Category),
		Title:       v.Title,
		Description: v.Description,
		ElementId:   v.ElementID,
		ElementType: convertElementTypeToPB(v.ElementType),
		Suggestion:  v.Suggestion,
		References:  convertReferencesToPB(v.References),
	}

	if v.Position != nil {
		pbViolation.Position = &commonpb.Point2D{
			X: v.Position.X,
			Y: v.Position.Y,
		}
	}

	return pbViolation
}

func convertSeverityToPB(s entity.Severity) pb.Severity {
	switch s {
	case entity.SeverityInfo:
		return pb.Severity_SEVERITY_INFO
	case entity.SeverityWarning:
		return pb.Severity_SEVERITY_WARNING
	case entity.SeverityError:
		return pb.Severity_SEVERITY_ERROR
	default:
		return pb.Severity_SEVERITY_UNSPECIFIED
	}
}

func convertSeverityFromPB(s pb.Severity) entity.Severity {
	switch s {
	case pb.Severity_SEVERITY_INFO:
		return entity.SeverityInfo
	case pb.Severity_SEVERITY_WARNING:
		return entity.SeverityWarning
	case pb.Severity_SEVERITY_ERROR:
		return entity.SeverityError
	default:
		return ""
	}
}

func convertElementTypeToPB(t entity.ElementType) pb.ElementType {
	switch t {
	case entity.ElementTypeWall:
		return pb.ElementType_ELEMENT_TYPE_WALL
	case entity.ElementTypeLoadBearingWall:
		return pb.ElementType_ELEMENT_TYPE_LOAD_BEARING_WALL
	case entity.ElementTypeRoom:
		return pb.ElementType_ELEMENT_TYPE_ROOM
	case entity.ElementTypeDoor:
		return pb.ElementType_ELEMENT_TYPE_DOOR
	case entity.ElementTypeWindow:
		return pb.ElementType_ELEMENT_TYPE_WINDOW
	case entity.ElementTypeWetZone:
		return pb.ElementType_ELEMENT_TYPE_WET_ZONE
	case entity.ElementTypeKitchen:
		return pb.ElementType_ELEMENT_TYPE_KITCHEN
	case entity.ElementTypeBathroom:
		return pb.ElementType_ELEMENT_TYPE_BATHROOM
	case entity.ElementTypeToilet:
		return pb.ElementType_ELEMENT_TYPE_TOILET
	case entity.ElementTypeSink:
		return pb.ElementType_ELEMENT_TYPE_SINK
	case entity.ElementTypeVentilation:
		return pb.ElementType_ELEMENT_TYPE_VENTILATION
	default:
		return pb.ElementType_ELEMENT_TYPE_UNSPECIFIED
	}
}

func convertStatsToPB(stats entity.ComplianceStats) *pb.ComplianceStats {
	return &pb.ComplianceStats{
		TotalRulesChecked: int32(stats.TotalRulesChecked),
		ErrorsCount:       int32(stats.ErrorsCount),
		WarningsCount:     int32(stats.WarningsCount),
		InfoCount:         int32(stats.InfoCount),
		ComplianceScore:   int32(stats.ComplianceScore),
	}
}

func convertReferencesToPB(refs []entity.DocumentReference) []*pb.DocumentReference {
	result := make([]*pb.DocumentReference, 0, len(refs))
	for _, ref := range refs {
		result = append(result, &pb.DocumentReference{
			Code:    ref.Code,
			Title:   ref.Title,
			Section: ref.Section,
			Url:     ref.URL,
		})
	}
	return result
}

func convertRulesToPB(rules []*entity.Rule) []*pb.Rule {
	result := make([]*pb.Rule, 0, len(rules))
	for _, rule := range rules {
		result = append(result, convertRuleToPB(rule))
	}
	return result
}

func convertRuleToPB(rule *entity.Rule) *pb.Rule {
	appliesTo := make([]pb.ElementType, 0, len(rule.AppliesTo))
	for _, t := range rule.AppliesTo {
		appliesTo = append(appliesTo, convertElementTypeToPB(t))
	}

	appliesToOps := make([]pb.OperationType, 0, len(rule.AppliesToOperations))
	for _, op := range rule.AppliesToOperations {
		appliesToOps = append(appliesToOps, convertOperationTypeToPB(op))
	}

	return &pb.Rule{
		Id:                  rule.ID.String(),
		Code:                rule.Code,
		Category:            string(rule.Category),
		Name:                rule.Name,
		Description:         rule.Description,
		Severity:            convertSeverityToPB(rule.Severity),
		Active:              rule.Active,
		AppliesTo:           appliesTo,
		AppliesToOperations: appliesToOps,
		References:          convertReferencesToPB(rule.References),
		Version:             rule.Version,
		UpdatedAt:           timestamppb.New(rule.UpdatedAt),
	}
}

func convertOperationTypeToPB(op entity.OperationType) pb.OperationType {
	switch op {
	case entity.OperationTypeDemolishWall:
		return pb.OperationType_OPERATION_TYPE_DEMOLISH_WALL
	case entity.OperationTypeAddWall:
		return pb.OperationType_OPERATION_TYPE_ADD_WALL
	case entity.OperationTypeMoveWall:
		return pb.OperationType_OPERATION_TYPE_MOVE_WALL
	case entity.OperationTypeAddOpening:
		return pb.OperationType_OPERATION_TYPE_ADD_OPENING
	case entity.OperationTypeCloseOpening:
		return pb.OperationType_OPERATION_TYPE_CLOSE_OPENING
	case entity.OperationTypeMergeRooms:
		return pb.OperationType_OPERATION_TYPE_MERGE_ROOMS
	case entity.OperationTypeSplitRoom:
		return pb.OperationType_OPERATION_TYPE_SPLIT_ROOM
	case entity.OperationTypeChangeRoomType:
		return pb.OperationType_OPERATION_TYPE_CHANGE_ROOM_TYPE
	case entity.OperationTypeMoveWetZone:
		return pb.OperationType_OPERATION_TYPE_MOVE_WET_ZONE
	case entity.OperationTypeExpandWetZone:
		return pb.OperationType_OPERATION_TYPE_EXPAND_WET_ZONE
	case entity.OperationTypeMovePlumbing:
		return pb.OperationType_OPERATION_TYPE_MOVE_PLUMBING
	case entity.OperationTypeMoveVentilation:
		return pb.OperationType_OPERATION_TYPE_MOVE_VENTILATION
	default:
		return pb.OperationType_OPERATION_TYPE_UNSPECIFIED
	}
}

func convertOperationFromPB(op *pb.Operation) *engine.OperationData {
	return &engine.OperationData{
		Type:        convertOperationTypeFromPB(op.Type),
		ElementID:   op.ElementId,
		ElementType: convertElementTypeFromPB(op.ElementType),
		Params:      op.Params,
	}
}

func convertOperationTypeFromPB(op pb.OperationType) entity.OperationType {
	switch op {
	case pb.OperationType_OPERATION_TYPE_DEMOLISH_WALL:
		return entity.OperationTypeDemolishWall
	case pb.OperationType_OPERATION_TYPE_ADD_WALL:
		return entity.OperationTypeAddWall
	case pb.OperationType_OPERATION_TYPE_MOVE_WALL:
		return entity.OperationTypeMoveWall
	case pb.OperationType_OPERATION_TYPE_ADD_OPENING:
		return entity.OperationTypeAddOpening
	case pb.OperationType_OPERATION_TYPE_CLOSE_OPENING:
		return entity.OperationTypeCloseOpening
	case pb.OperationType_OPERATION_TYPE_MERGE_ROOMS:
		return entity.OperationTypeMergeRooms
	case pb.OperationType_OPERATION_TYPE_SPLIT_ROOM:
		return entity.OperationTypeSplitRoom
	case pb.OperationType_OPERATION_TYPE_CHANGE_ROOM_TYPE:
		return entity.OperationTypeChangeRoomType
	case pb.OperationType_OPERATION_TYPE_MOVE_WET_ZONE:
		return entity.OperationTypeMoveWetZone
	case pb.OperationType_OPERATION_TYPE_EXPAND_WET_ZONE:
		return entity.OperationTypeExpandWetZone
	case pb.OperationType_OPERATION_TYPE_MOVE_PLUMBING:
		return entity.OperationTypeMovePlumbing
	case pb.OperationType_OPERATION_TYPE_MOVE_VENTILATION:
		return entity.OperationTypeMoveVentilation
	default:
		return ""
	}
}

func convertElementTypeFromPB(t pb.ElementType) entity.ElementType {
	switch t {
	case pb.ElementType_ELEMENT_TYPE_WALL:
		return entity.ElementTypeWall
	case pb.ElementType_ELEMENT_TYPE_LOAD_BEARING_WALL:
		return entity.ElementTypeLoadBearingWall
	case pb.ElementType_ELEMENT_TYPE_ROOM:
		return entity.ElementTypeRoom
	case pb.ElementType_ELEMENT_TYPE_DOOR:
		return entity.ElementTypeDoor
	case pb.ElementType_ELEMENT_TYPE_WINDOW:
		return entity.ElementTypeWindow
	case pb.ElementType_ELEMENT_TYPE_WET_ZONE:
		return entity.ElementTypeWetZone
	case pb.ElementType_ELEMENT_TYPE_KITCHEN:
		return entity.ElementTypeKitchen
	case pb.ElementType_ELEMENT_TYPE_BATHROOM:
		return entity.ElementTypeBathroom
	case pb.ElementType_ELEMENT_TYPE_TOILET:
		return entity.ElementTypeToilet
	case pb.ElementType_ELEMENT_TYPE_SINK:
		return entity.ElementTypeSink
	case pb.ElementType_ELEMENT_TYPE_VENTILATION:
		return entity.ElementTypeVentilation
	default:
		return ""
	}
}

// PaginationResponse type for using in GetRules
type PaginationResponse = commonpb.PaginationResponse

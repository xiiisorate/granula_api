// Package service provides business logic for Compliance Service.
package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/compliance-service/internal/domain/entity"
	"github.com/xiiisorate/granula_api/compliance-service/internal/engine"
	"github.com/xiiisorate/granula_api/compliance-service/internal/repository/postgres"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
)

// ComplianceService provides compliance checking operations.
type ComplianceService struct {
	ruleRepo *postgres.RuleRepository
	log      *logger.Logger
}

// NewComplianceService creates a new ComplianceService.
func NewComplianceService(ruleRepo *postgres.RuleRepository, log *logger.Logger) *ComplianceService {
	return &ComplianceService{
		ruleRepo: ruleRepo,
		log:      log,
	}
}

// CheckCompliance performs a full compliance check on a scene.
func (s *ComplianceService) CheckCompliance(ctx context.Context, sceneData *engine.SceneData) (*entity.ComplianceResult, error) {
	s.log.Info("checking compliance for scene",
		logger.String("scene_id", sceneData.ID),
	)

	// Load all active rules
	rules, _, err := s.ruleRepo.List(ctx, postgres.ListOptions{ActiveOnly: true, Limit: 1000})
	if err != nil {
		s.log.Error("failed to load rules", logger.Err(err))
		return nil, err
	}

	// Create rule engine and check
	ruleEngine := engine.NewRuleEngine(rules)
	result := ruleEngine.CheckScene(ctx, sceneData)

	s.log.Info("compliance check completed",
		logger.String("scene_id", sceneData.ID),
		logger.Bool("compliant", result.Compliant),
		logger.Int("violations", len(result.Violations)),
	)

	return result, nil
}

// CheckOperation checks if a specific operation is allowed.
func (s *ComplianceService) CheckOperation(ctx context.Context, sceneData *engine.SceneData, operation *engine.OperationData) (*entity.ComplianceResult, error) {
	s.log.Info("checking operation",
		logger.String("scene_id", sceneData.ID),
		logger.String("operation_type", string(operation.Type)),
		logger.String("element_id", operation.ElementID),
	)

	// Load rules that apply to this operation type
	rules, err := s.ruleRepo.ListByOperation(ctx, operation.Type)
	if err != nil {
		s.log.Error("failed to load rules for operation", logger.Err(err))
		return nil, err
	}

	// If no specific rules, load all active rules
	if len(rules) == 0 {
		rules, _, err = s.ruleRepo.List(ctx, postgres.ListOptions{ActiveOnly: true, Limit: 1000})
		if err != nil {
			return nil, err
		}
	}

	// Create rule engine and check operation
	ruleEngine := engine.NewRuleEngine(rules)
	result := ruleEngine.CheckOperation(ctx, sceneData, operation)

	s.log.Info("operation check completed",
		logger.String("scene_id", sceneData.ID),
		logger.String("operation_type", string(operation.Type)),
		logger.Bool("allowed", result.Compliant),
		logger.Int("violations", len(result.Violations)),
	)

	return result, nil
}

// GetRules returns rules with optional filtering.
func (s *ComplianceService) GetRules(ctx context.Context, opts GetRulesOptions) ([]*entity.Rule, int, error) {
	listOpts := postgres.ListOptions{
		Category:   opts.Category,
		Severity:   opts.Severity,
		ActiveOnly: opts.ActiveOnly,
		Limit:      opts.Limit,
		Offset:     opts.Offset,
	}

	return s.ruleRepo.List(ctx, listOpts)
}

// GetRule returns a single rule by ID.
func (s *ComplianceService) GetRule(ctx context.Context, ruleID uuid.UUID) (*entity.Rule, error) {
	return s.ruleRepo.GetByID(ctx, ruleID)
}

// GetRuleByCode returns a single rule by code.
func (s *ComplianceService) GetRuleByCode(ctx context.Context, code string) (*entity.Rule, error) {
	return s.ruleRepo.GetByCode(ctx, code)
}

// GetCategories returns all rule categories with metadata.
func (s *ComplianceService) GetCategories(ctx context.Context) ([]*entity.RuleCategoryInfo, error) {
	return s.ruleRepo.GetCategories(ctx)
}

// GetRulesOptions for filtering rules.
type GetRulesOptions struct {
	Category   entity.RuleCategory
	Severity   entity.Severity
	ActiveOnly bool
	Limit      int
	Offset     int
}

// ValidateScene performs a quick validation (only critical errors).
func (s *ComplianceService) ValidateScene(ctx context.Context, sceneData *engine.SceneData) (*ValidationResult, error) {
	result, err := s.CheckCompliance(ctx, sceneData)
	if err != nil {
		return nil, err
	}

	// Filter only critical errors
	criticalErrors := result.FilterBySeverity(entity.SeverityError)

	return &ValidationResult{
		Valid:          len(criticalErrors) == 0,
		CriticalErrors: criticalErrors,
		WarningsCount:  result.Stats.WarningsCount,
	}, nil
}

// ValidationResult is a simplified validation result.
type ValidationResult struct {
	Valid          bool                `json:"valid"`
	CriticalErrors []*entity.Violation `json:"critical_errors"`
	WarningsCount  int                 `json:"warnings_count"`
}

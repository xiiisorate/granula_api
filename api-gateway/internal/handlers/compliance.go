// =============================================================================
// Package handlers provides HTTP handlers for API Gateway.
// =============================================================================
// ComplianceHandler handles compliance check HTTP requests including:
// - Check: Full compliance check of a scene/branch
// - CheckOperation: Check if a specific operation is allowed
// - GetRules: List all compliance rules
// - GetRule: Get details of a specific rule
// - GenerateReport: Generate compliance report (PDF/JSON)
//
// Documentation: docs/api/compliance.md
//
// The compliance service checks against Russian building codes:
// - СНиП 31-01-2003 (residential buildings)
// - ЖК РФ (Housing Code of Russian Federation)
// - СП 54.13330.2016 (updated SNiP)
// - Regional regulations
// =============================================================================
package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"

	commonpb "github.com/xiiisorate/granula_api/shared/gen/common/v1"
	compliancev1 "github.com/xiiisorate/granula_api/shared/gen/compliance/v1"
)

// =============================================================================
// ComplianceHandler
// =============================================================================

// ComplianceHandler handles compliance check HTTP requests.
// It communicates with the Compliance microservice via gRPC.
type ComplianceHandler struct {
	// client is the gRPC client for ComplianceService.
	client compliancev1.ComplianceServiceClient
}

// NewComplianceHandler creates a new ComplianceHandler.
//
// Parameters:
//   - conn: gRPC connection to the Compliance service
//
// Returns:
//   - *ComplianceHandler: New handler instance
func NewComplianceHandler(conn *grpc.ClientConn) *ComplianceHandler {
	return &ComplianceHandler{
		client: compliancev1.NewComplianceServiceClient(conn),
	}
}

// =============================================================================
// Check - POST /compliance/check
// =============================================================================

// CheckComplianceInput represents input for full compliance check.
type CheckComplianceInput struct {
	// SceneID - ID сцены для проверки (обязательно)
	SceneID string `json:"scene_id" validate:"required"`
	// BranchID - ID ветки (опционально, по умолчанию активная ветка)
	BranchID string `json:"branch_id,omitempty"`
	// Categories - категории для проверки (пустой = все)
	Categories []string `json:"categories,omitempty"`
	// IncludeSuggestions - включить предложения по исправлению
	IncludeSuggestions bool `json:"include_suggestions,omitempty"`
	// IncludeReferences - включить ссылки на нормативы
	IncludeReferences bool `json:"include_references,omitempty"`
}

// Check выполняет полную проверку сцены на соответствие нормам.
//
// @Summary Проверить соответствие нормам
// @Description Полная проверка сцены/ветки на соответствие строительным нормам РФ.
// Возвращает список всех нарушений и предупреждений.
// @Tags compliance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CheckComplianceInput true "Параметры проверки"
// @Success 200 {object} ComplianceCheckResponse "Результат проверки"
// @Failure 400 {object} ErrorResponse "Некорректный запрос"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Сцена не найдена"
// @Router /compliance/check [post]
func (h *ComplianceHandler) Check(c *fiber.Ctx) error {
	// Parse request body
	var input CheckComplianceInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// Validate required fields
	if input.SceneID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "scene_id is required")
	}

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	req := &compliancev1.CheckComplianceRequest{
		SceneId:  input.SceneID,
		BranchId: input.BranchID,
		Options: &compliancev1.CheckOptions{
			Categories:         input.Categories,
			IncludeSuggestions: input.IncludeSuggestions,
			IncludeReferences:  input.IncludeReferences,
		},
	}

	// Call gRPC service
	resp, err := h.client.CheckCompliance(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Convert response
	violations := violationsToResponse(resp.Violations)

	// Build stats
	stats := fiber.Map{}
	if resp.Stats != nil {
		stats = fiber.Map{
			"total_rules_checked": resp.Stats.TotalRulesChecked,
			"errors_count":        resp.Stats.ErrorsCount,
			"warnings_count":      resp.Stats.WarningsCount,
			"info_count":          resp.Stats.InfoCount,
			"compliance_score":    resp.Stats.ComplianceScore,
		}
	}

	// Return response
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"scene_id":      input.SceneID,
			"branch_id":     input.BranchID,
			"compliant":     resp.Compliant,
			"violations":    violations,
			"stats":         stats,
			"rules_version": resp.RulesVersion,
			"checked_at":    time.Now(),
		},
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// CheckOperation - POST /compliance/check-operation
// =============================================================================

// CheckOperationInput represents input for checking a single operation.
type CheckOperationInput struct {
	// SceneID - ID сцены (обязательно)
	SceneID string `json:"scene_id" validate:"required"`
	// BranchID - ID ветки (опционально)
	BranchID string `json:"branch_id,omitempty"`
	// Operation - операция для проверки
	Operation OperationInput `json:"operation" validate:"required"`
}

// OperationInput represents an operation to check.
type OperationInput struct {
	// Type - тип операции (demolish_wall, add_wall, move_wall, etc.)
	Type string `json:"type" validate:"required"`
	// ElementID - ID элемента (для существующих элементов)
	ElementID string `json:"element_id,omitempty"`
	// ElementType - тип элемента (wall, room, door, etc.)
	ElementType string `json:"element_type" validate:"required"`
	// Params - дополнительные параметры операции
	Params map[string]string `json:"params,omitempty"`
}

// CheckOperation проверяет возможность выполнения операции.
//
// @Summary Проверить операцию
// @Description Превентивная проверка - можно ли выполнить операцию (снос стены, перенос и т.д.).
// Используется для валидации действий пользователя в реальном времени.
// @Tags compliance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CheckOperationInput true "Операция для проверки"
// @Success 200 {object} CheckOperationResponse "Результат проверки операции"
// @Failure 400 {object} ErrorResponse "Некорректный запрос"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Сцена не найдена"
// @Router /compliance/check-operation [post]
func (h *ComplianceHandler) CheckOperation(c *fiber.Ctx) error {
	// Parse request body
	var input CheckOperationInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// Validate required fields
	if input.SceneID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "scene_id is required")
	}
	if input.Operation.Type == "" {
		return fiber.NewError(fiber.StatusBadRequest, "operation.type is required")
	}

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &compliancev1.CheckOperationRequest{
		SceneId:  input.SceneID,
		BranchId: input.BranchID,
		Operation: &compliancev1.Operation{
			Type:        stringToOperationType(input.Operation.Type),
			ElementId:   input.Operation.ElementID,
			ElementType: stringToElementType(input.Operation.ElementType),
			Params:      input.Operation.Params,
		},
	}

	// Call gRPC service
	resp, err := h.client.CheckOperation(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Convert response
	violations := violationsToResponse(resp.Violations)
	warnings := violationsToResponse(resp.Warnings)

	// Return response
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"allowed":           resp.Allowed,
			"violations":        violations,
			"warnings":          warnings,
			"requires_approval": resp.RequiresApproval,
			"approval_type":     approvalTypeToString(resp.ApprovalType),
		},
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// GetRules - GET /compliance/rules
// =============================================================================

// GetRules возвращает список правил проверки.
//
// @Summary Список правил
// @Description Получить справочник правил проверки соответствия нормам.
// @Tags compliance
// @Produce json
// @Security BearerAuth
// @Param category query string false "Фильтр по категории"
// @Param severity query string false "Фильтр по серьёзности (error, warning, info)"
// @Param active query bool false "Только активные правила" default(true)
// @Param limit query int false "Количество записей" default(50)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {object} RulesListResponse "Список правил"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Router /compliance/rules [get]
func (h *ComplianceHandler) GetRules(c *fiber.Ctx) error {
	// Get filter parameters
	category := c.Query("category", "")
	severityStr := c.Query("severity", "")
	activeOnly := c.QueryBool("active", true)

	// Get pagination
	pagination := GetPaginationParams(c)

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &compliancev1.GetRulesRequest{
		Category:   category,
		Severity:   stringToSeverity(severityStr),
		ActiveOnly: activeOnly,
		Pagination: &commonpb.PaginationRequest{
			Page:     int32(pagination.Page),
			PageSize: int32(pagination.Limit),
		},
	}

	// Call gRPC service
	resp, err := h.client.GetRules(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Convert rules to response format
	rules := make([]fiber.Map, 0, len(resp.Rules))
	for _, r := range resp.Rules {
		rules = append(rules, ruleToResponse(r))
	}

	// Get total from pagination response
	total := 0
	if resp.Pagination != nil {
		total = int(resp.Pagination.Total)
	}

	// Return response
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"rules": rules,
			"total": total,
		},
		"pagination": PaginationResponse(pagination.Page, pagination.Limit, total),
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// GetRule - GET /compliance/rules/:id
// =============================================================================

// GetRule возвращает детали правила по ID.
//
// @Summary Получить правило
// @Description Получить детальную информацию о правиле, включая ссылки на нормативы.
// @Tags compliance
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID правила"
// @Success 200 {object} RuleResponse "Детали правила"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Правило не найдено"
// @Router /compliance/rules/{id} [get]
func (h *ComplianceHandler) GetRule(c *fiber.Ctx) error {
	// Get rule ID from path
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "rule ID is required")
	}

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &compliancev1.GetRuleRequest{
		RuleId: id,
	}

	// Call gRPC service
	rule, err := h.client.GetRule(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Return response
	return SuccessResponseData(c, ruleDetailedToResponse(rule))
}

// =============================================================================
// GenerateReport - POST /compliance/report
// =============================================================================

// GenerateReportInput represents input for report generation.
type GenerateReportInput struct {
	// SceneID - ID сцены (обязательно)
	SceneID string `json:"scene_id" validate:"required"`
	// BranchID - ID ветки (опционально)
	BranchID string `json:"branch_id,omitempty"`
	// Format - формат отчёта (pdf, json, html)
	Format string `json:"format,omitempty" default:"pdf"`
	// IncludeFloorPlan - включить схему планировки
	IncludeFloorPlan bool `json:"include_floor_plan,omitempty"`
	// IncludeRecommendations - включить рекомендации
	IncludeRecommendations bool `json:"include_recommendations,omitempty"`
	// Language - язык отчёта (ru, en)
	Language string `json:"language,omitempty" default:"ru"`
}

// GenerateReport генерирует отчёт о соответствии.
//
// @Summary Сгенерировать отчёт
// @Description Создать полный отчёт о соответствии планировки нормам.
// Доступные форматы: PDF, JSON, HTML.
// @Tags compliance
// @Accept json
// @Produce json
// @Produce application/pdf
// @Security BearerAuth
// @Param body body GenerateReportInput true "Параметры отчёта"
// @Success 200 {object} GenerateReportResponse "Отчёт сгенерирован"
// @Failure 400 {object} ErrorResponse "Некорректный запрос"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Сцена не найдена"
// @Router /compliance/report [post]
func (h *ComplianceHandler) GenerateReport(c *fiber.Ctx) error {
	// Parse request body
	var input GenerateReportInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// Validate required fields
	if input.SceneID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "scene_id is required")
	}

	// Default format
	if input.Format == "" {
		input.Format = "pdf"
	}
	if input.Language == "" {
		input.Language = "ru"
	}

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 60*time.Second)
	defer cancel()

	req := &compliancev1.GenerateReportRequest{
		SceneId:  input.SceneID,
		BranchId: input.BranchID,
		Format:   stringToReportFormat(input.Format),
		Options: &compliancev1.ReportOptions{
			IncludeFloorPlan:        input.IncludeFloorPlan,
			IncludeViolationDetails: true,
			IncludeRecommendations:  input.IncludeRecommendations,
			IncludeReferences:       true,
			Language:                input.Language,
		},
	}

	// Call gRPC service
	resp, err := h.client.GenerateReport(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// If format is PDF/HTML, return as file download
	if input.Format == "pdf" || input.Format == "html" {
		c.Set("Content-Type", resp.ContentType)
		c.Set("Content-Disposition", "attachment; filename=\""+resp.Filename+"\"")
		return c.Send(resp.Content)
	}

	// For JSON format, return as regular response
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"filename":     resp.Filename,
			"content_type": resp.ContentType,
			"size":         resp.Size,
			"content":      string(resp.Content),
		},
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// Helper Functions
// =============================================================================

// violationsToResponse converts proto Violation slice to response format.
func violationsToResponse(violations []*compliancev1.Violation) []fiber.Map {
	result := make([]fiber.Map, 0, len(violations))
	for _, v := range violations {
		violation := fiber.Map{
			"id":           v.Id,
			"rule_id":      v.RuleId,
			"rule_code":    v.RuleCode,
			"severity":     severityToString(v.Severity),
			"category":     v.Category,
			"title":        v.Title,
			"description":  v.Description,
			"element_id":   v.ElementId,
			"element_type": elementTypeToString(v.ElementType),
			"suggestion":   v.Suggestion,
		}

		// Add position if present
		if v.Position != nil {
			violation["position"] = fiber.Map{
				"x": v.Position.X,
				"y": v.Position.Y,
			}
		}

		// Add references if present
		if len(v.References) > 0 {
			refs := make([]fiber.Map, 0, len(v.References))
			for _, ref := range v.References {
				refs = append(refs, fiber.Map{
					"code":    ref.Code,
					"title":   ref.Title,
					"section": ref.Section,
					"url":     ref.Url,
				})
			}
			violation["references"] = refs
		}

		result = append(result, violation)
	}
	return result
}

// ruleToResponse converts proto Rule to API response format (brief).
func ruleToResponse(r *compliancev1.Rule) fiber.Map {
	return fiber.Map{
		"id":          r.Id,
		"code":        r.Code,
		"category":    r.Category,
		"name":        r.Name,
		"description": r.Description,
		"severity":    severityToString(r.Severity),
		"active":      r.Active,
	}
}

// ruleDetailedToResponse converts proto Rule to API response format (detailed).
func ruleDetailedToResponse(r *compliancev1.Rule) fiber.Map {
	result := fiber.Map{
		"id":          r.Id,
		"code":        r.Code,
		"category":    r.Category,
		"name":        r.Name,
		"description": r.Description,
		"severity":    severityToString(r.Severity),
		"active":      r.Active,
		"version":     r.Version,
	}

	// Add applies_to element types
	if len(r.AppliesTo) > 0 {
		types := make([]string, 0, len(r.AppliesTo))
		for _, t := range r.AppliesTo {
			types = append(types, elementTypeToString(t))
		}
		result["applies_to"] = types
	}

	// Add applies_to operations
	if len(r.AppliesToOperations) > 0 {
		ops := make([]string, 0, len(r.AppliesToOperations))
		for _, o := range r.AppliesToOperations {
			ops = append(ops, operationTypeToString(o))
		}
		result["applies_to_operations"] = ops
	}

	// Add references
	if len(r.References) > 0 {
		refs := make([]fiber.Map, 0, len(r.References))
		for _, ref := range r.References {
			refs = append(refs, fiber.Map{
				"code":    ref.Code,
				"title":   ref.Title,
				"section": ref.Section,
				"url":     ref.Url,
			})
		}
		result["references"] = refs
	}

	// Add timestamp
	if r.UpdatedAt != nil {
		result["updated_at"] = r.UpdatedAt.AsTime()
	}

	return result
}

// severityToString converts proto Severity to string.
func severityToString(s compliancev1.Severity) string {
	switch s {
	case compliancev1.Severity_SEVERITY_INFO:
		return "info"
	case compliancev1.Severity_SEVERITY_WARNING:
		return "warning"
	case compliancev1.Severity_SEVERITY_ERROR:
		return "error"
	default:
		return "unknown"
	}
}

// stringToSeverity converts string to proto Severity.
func stringToSeverity(s string) compliancev1.Severity {
	switch s {
	case "info":
		return compliancev1.Severity_SEVERITY_INFO
	case "warning":
		return compliancev1.Severity_SEVERITY_WARNING
	case "error":
		return compliancev1.Severity_SEVERITY_ERROR
	default:
		return compliancev1.Severity_SEVERITY_UNSPECIFIED
	}
}

// elementTypeToString converts proto ElementType to string.
func elementTypeToString(t compliancev1.ElementType) string {
	switch t {
	case compliancev1.ElementType_ELEMENT_TYPE_WALL:
		return "wall"
	case compliancev1.ElementType_ELEMENT_TYPE_LOAD_BEARING_WALL:
		return "load_bearing_wall"
	case compliancev1.ElementType_ELEMENT_TYPE_ROOM:
		return "room"
	case compliancev1.ElementType_ELEMENT_TYPE_DOOR:
		return "door"
	case compliancev1.ElementType_ELEMENT_TYPE_WINDOW:
		return "window"
	case compliancev1.ElementType_ELEMENT_TYPE_WET_ZONE:
		return "wet_zone"
	case compliancev1.ElementType_ELEMENT_TYPE_KITCHEN:
		return "kitchen"
	case compliancev1.ElementType_ELEMENT_TYPE_BATHROOM:
		return "bathroom"
	case compliancev1.ElementType_ELEMENT_TYPE_TOILET:
		return "toilet"
	case compliancev1.ElementType_ELEMENT_TYPE_SINK:
		return "sink"
	case compliancev1.ElementType_ELEMENT_TYPE_BATHTUB:
		return "bathtub"
	case compliancev1.ElementType_ELEMENT_TYPE_SHOWER:
		return "shower"
	case compliancev1.ElementType_ELEMENT_TYPE_STOVE:
		return "stove"
	case compliancev1.ElementType_ELEMENT_TYPE_VENTILATION:
		return "ventilation"
	case compliancev1.ElementType_ELEMENT_TYPE_RADIATOR:
		return "radiator"
	case compliancev1.ElementType_ELEMENT_TYPE_FURNITURE:
		return "furniture"
	default:
		return "unknown"
	}
}

// stringToElementType converts string to proto ElementType.
func stringToElementType(s string) compliancev1.ElementType {
	switch s {
	case "wall":
		return compliancev1.ElementType_ELEMENT_TYPE_WALL
	case "load_bearing_wall":
		return compliancev1.ElementType_ELEMENT_TYPE_LOAD_BEARING_WALL
	case "room":
		return compliancev1.ElementType_ELEMENT_TYPE_ROOM
	case "door":
		return compliancev1.ElementType_ELEMENT_TYPE_DOOR
	case "window":
		return compliancev1.ElementType_ELEMENT_TYPE_WINDOW
	case "wet_zone":
		return compliancev1.ElementType_ELEMENT_TYPE_WET_ZONE
	case "kitchen":
		return compliancev1.ElementType_ELEMENT_TYPE_KITCHEN
	case "bathroom":
		return compliancev1.ElementType_ELEMENT_TYPE_BATHROOM
	case "toilet":
		return compliancev1.ElementType_ELEMENT_TYPE_TOILET
	case "sink":
		return compliancev1.ElementType_ELEMENT_TYPE_SINK
	case "bathtub":
		return compliancev1.ElementType_ELEMENT_TYPE_BATHTUB
	case "shower":
		return compliancev1.ElementType_ELEMENT_TYPE_SHOWER
	case "stove":
		return compliancev1.ElementType_ELEMENT_TYPE_STOVE
	case "ventilation":
		return compliancev1.ElementType_ELEMENT_TYPE_VENTILATION
	case "radiator":
		return compliancev1.ElementType_ELEMENT_TYPE_RADIATOR
	case "furniture":
		return compliancev1.ElementType_ELEMENT_TYPE_FURNITURE
	default:
		return compliancev1.ElementType_ELEMENT_TYPE_UNSPECIFIED
	}
}

// operationTypeToString converts proto OperationType to string.
func operationTypeToString(t compliancev1.OperationType) string {
	switch t {
	case compliancev1.OperationType_OPERATION_TYPE_DEMOLISH_WALL:
		return "demolish_wall"
	case compliancev1.OperationType_OPERATION_TYPE_ADD_WALL:
		return "add_wall"
	case compliancev1.OperationType_OPERATION_TYPE_MOVE_WALL:
		return "move_wall"
	case compliancev1.OperationType_OPERATION_TYPE_ADD_OPENING:
		return "add_opening"
	case compliancev1.OperationType_OPERATION_TYPE_CLOSE_OPENING:
		return "close_opening"
	case compliancev1.OperationType_OPERATION_TYPE_MERGE_ROOMS:
		return "merge_rooms"
	case compliancev1.OperationType_OPERATION_TYPE_SPLIT_ROOM:
		return "split_room"
	case compliancev1.OperationType_OPERATION_TYPE_CHANGE_ROOM_TYPE:
		return "change_room_type"
	case compliancev1.OperationType_OPERATION_TYPE_MOVE_WET_ZONE:
		return "move_wet_zone"
	case compliancev1.OperationType_OPERATION_TYPE_EXPAND_WET_ZONE:
		return "expand_wet_zone"
	case compliancev1.OperationType_OPERATION_TYPE_MOVE_PLUMBING:
		return "move_plumbing"
	case compliancev1.OperationType_OPERATION_TYPE_MOVE_VENTILATION:
		return "move_ventilation"
	case compliancev1.OperationType_OPERATION_TYPE_ADD_ELEMENT:
		return "add_element"
	case compliancev1.OperationType_OPERATION_TYPE_REMOVE_ELEMENT:
		return "remove_element"
	case compliancev1.OperationType_OPERATION_TYPE_MOVE_ELEMENT:
		return "move_element"
	case compliancev1.OperationType_OPERATION_TYPE_RESIZE_ELEMENT:
		return "resize_element"
	default:
		return "unknown"
	}
}

// stringToOperationType converts string to proto OperationType.
func stringToOperationType(s string) compliancev1.OperationType {
	switch s {
	case "demolish_wall":
		return compliancev1.OperationType_OPERATION_TYPE_DEMOLISH_WALL
	case "add_wall":
		return compliancev1.OperationType_OPERATION_TYPE_ADD_WALL
	case "move_wall":
		return compliancev1.OperationType_OPERATION_TYPE_MOVE_WALL
	case "add_opening":
		return compliancev1.OperationType_OPERATION_TYPE_ADD_OPENING
	case "close_opening":
		return compliancev1.OperationType_OPERATION_TYPE_CLOSE_OPENING
	case "merge_rooms":
		return compliancev1.OperationType_OPERATION_TYPE_MERGE_ROOMS
	case "split_room":
		return compliancev1.OperationType_OPERATION_TYPE_SPLIT_ROOM
	case "change_room_type":
		return compliancev1.OperationType_OPERATION_TYPE_CHANGE_ROOM_TYPE
	case "move_wet_zone":
		return compliancev1.OperationType_OPERATION_TYPE_MOVE_WET_ZONE
	case "expand_wet_zone":
		return compliancev1.OperationType_OPERATION_TYPE_EXPAND_WET_ZONE
	case "move_plumbing":
		return compliancev1.OperationType_OPERATION_TYPE_MOVE_PLUMBING
	case "move_ventilation":
		return compliancev1.OperationType_OPERATION_TYPE_MOVE_VENTILATION
	case "add_element", "add":
		return compliancev1.OperationType_OPERATION_TYPE_ADD_ELEMENT
	case "remove_element", "remove":
		return compliancev1.OperationType_OPERATION_TYPE_REMOVE_ELEMENT
	case "move_element", "move":
		return compliancev1.OperationType_OPERATION_TYPE_MOVE_ELEMENT
	case "resize_element", "resize":
		return compliancev1.OperationType_OPERATION_TYPE_RESIZE_ELEMENT
	default:
		return compliancev1.OperationType_OPERATION_TYPE_UNSPECIFIED
	}
}

// approvalTypeToString converts proto ApprovalType to string.
func approvalTypeToString(t compliancev1.ApprovalType) string {
	switch t {
	case compliancev1.ApprovalType_APPROVAL_TYPE_NONE:
		return "none"
	case compliancev1.ApprovalType_APPROVAL_TYPE_NOTIFICATION:
		return "notification"
	case compliancev1.ApprovalType_APPROVAL_TYPE_PROJECT:
		return "project"
	case compliancev1.ApprovalType_APPROVAL_TYPE_EXPERTISE:
		return "expertise"
	case compliancev1.ApprovalType_APPROVAL_TYPE_PROHIBITED:
		return "prohibited"
	default:
		return "unknown"
	}
}

// stringToReportFormat converts string to proto ReportFormat.
func stringToReportFormat(s string) compliancev1.ReportFormat {
	switch s {
	case "pdf":
		return compliancev1.ReportFormat_REPORT_FORMAT_PDF
	case "json":
		return compliancev1.ReportFormat_REPORT_FORMAT_JSON
	case "html":
		return compliancev1.ReportFormat_REPORT_FORMAT_HTML
	default:
		return compliancev1.ReportFormat_REPORT_FORMAT_PDF
	}
}

// =============================================================================
// Response Types (for Swagger documentation)
// =============================================================================

// ComplianceCheckResponse represents compliance check result.
// swagger:model
type ComplianceCheckResponse struct {
	Data      ComplianceCheckData `json:"data"`
	RequestID string              `json:"request_id"`
}

// ComplianceCheckData contains check result data.
type ComplianceCheckData struct {
	SceneID      string        `json:"scene_id"`
	BranchID     string        `json:"branch_id,omitempty"`
	Compliant    bool          `json:"compliant"`
	Violations   []interface{} `json:"violations"`
	Stats        interface{}   `json:"stats"`
	RulesVersion string        `json:"rules_version"`
	CheckedAt    time.Time     `json:"checked_at"`
}

// CheckOperationResponse represents operation check result.
// swagger:model
type CheckOperationResponse struct {
	Data      CheckOperationData `json:"data"`
	RequestID string             `json:"request_id"`
}

// CheckOperationData contains operation check result.
type CheckOperationData struct {
	Allowed          bool          `json:"allowed"`
	Violations       []interface{} `json:"violations"`
	Warnings         []interface{} `json:"warnings"`
	RequiresApproval bool          `json:"requires_approval"`
	ApprovalType     string        `json:"approval_type"`
}

// RulesListResponse represents a list of compliance rules.
// swagger:model
type RulesListResponse struct {
	Data       RulesListData `json:"data"`
	Pagination fiber.Map     `json:"pagination"`
	RequestID  string        `json:"request_id"`
}

// RulesListData contains rules list data.
type RulesListData struct {
	Rules []interface{} `json:"rules"`
	Total int           `json:"total"`
}

// RuleResponse represents a compliance rule.
// swagger:model
type RuleResponse struct {
	Data      interface{} `json:"data"`
	RequestID string      `json:"request_id"`
}

// GenerateReportResponse represents report generation result.
// swagger:model
type GenerateReportResponse struct {
	Data      GenerateReportData `json:"data"`
	RequestID string             `json:"request_id"`
}

// GenerateReportData contains report info.
type GenerateReportData struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	Content     string `json:"content,omitempty"`
}

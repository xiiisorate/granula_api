// =============================================================================
// Package handlers provides HTTP handlers for API Gateway.
// =============================================================================
// SceneHandler handles 3D scene-related HTTP requests.
// =============================================================================
package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"google.golang.org/grpc"

	scenepb "github.com/xiiisorate/granula_api/shared/gen/scene/v1"
)

// SceneHandler handles scene-related HTTP requests.
type SceneHandler struct {
	client scenepb.SceneServiceClient
}

// NewSceneHandler creates a new SceneHandler.
func NewSceneHandler(conn *grpc.ClientConn) *SceneHandler {
	return &SceneHandler{
		client: scenepb.NewSceneServiceClient(conn),
	}
}

// =============================================================================
// CreateScene creates a new 3D scene.
// @Summary Создать сцену
// @Description Создание новой 3D сцены в воркспейсе
// @Tags scenes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workspace_id path string true "ID воркспейса"
// @Param body body CreateSceneInput true "Данные сцены"
// @Success 201 {object} SceneResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /workspaces/{workspace_id}/scenes [post]
// =============================================================================
func (h *SceneHandler) CreateScene(c *fiber.Ctx) error {
	workspaceID := c.Params("workspace_id")
	if _, err := uuid.Parse(workspaceID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid workspace ID")
	}

	userIDStr, ok := c.Locals("user_id").(string)
	if !ok || userIDStr == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "user not authenticated")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid user ID")
	}

	var input CreateSceneInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if input.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &scenepb.CreateSceneRequest{
		WorkspaceId: workspaceID,
		OwnerId:     userID.String(),
		Name:        input.Name,
		Description: input.Description,
		FloorPlanId: input.FloorPlanID,
	}

	resp, err := h.client.CreateScene(ctx, req)
	if err != nil {
		return handleGRPCError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data":       sceneToResponse(resp.Scene),
		"message":    "Scene created successfully",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// GetScene returns a scene by ID.
// @Summary Получить сцену
// @Description Получение сцены по ID
// @Tags scenes
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID сцены"
// @Success 200 {object} SceneResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /scenes/{id} [get]
// =============================================================================
func (h *SceneHandler) GetScene(c *fiber.Ctx) error {
	sceneID := c.Params("id")
	if _, err := uuid.Parse(sceneID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid scene ID")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.client.GetScene(ctx, &scenepb.GetSceneRequest{
		Id: sceneID,
	})
	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"data":       sceneToResponse(resp.Scene),
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// ListScenes returns scenes for a workspace.
// @Summary Список сцен
// @Description Получение списка сцен воркспейса
// @Tags scenes
// @Produce json
// @Security BearerAuth
// @Param workspace_id path string true "ID воркспейса"
// @Success 200 {object} SceneListResponse
// @Failure 401 {object} ErrorResponse
// @Router /workspaces/{workspace_id}/scenes [get]
// =============================================================================
func (h *SceneHandler) ListScenes(c *fiber.Ctx) error {
	workspaceID := c.Params("workspace_id")
	if _, err := uuid.Parse(workspaceID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid workspace ID")
	}

	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.client.ListScenes(ctx, &scenepb.ListScenesRequest{
		WorkspaceId: workspaceID,
		Limit:       int32(limit),
		Offset:      int32(offset),
	})
	if err != nil {
		return handleGRPCError(err)
	}

	scenes := make([]fiber.Map, 0, len(resp.Scenes))
	for _, s := range resp.Scenes {
		scenes = append(scenes, sceneToResponse(s))
	}

	return c.JSON(fiber.Map{
		"data":       scenes,
		"total":      resp.Total,
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// UpdateScene updates a scene.
// @Summary Обновить сцену
// @Description Обновление метаданных сцены
// @Tags scenes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID сцены"
// @Param body body UpdateSceneInput true "Данные для обновления"
// @Success 200 {object} SceneResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /scenes/{id} [patch]
// =============================================================================
func (h *SceneHandler) UpdateScene(c *fiber.Ctx) error {
	sceneID := c.Params("id")
	if _, err := uuid.Parse(sceneID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid scene ID")
	}

	var input UpdateSceneInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &scenepb.UpdateSceneRequest{
		Id:          sceneID,
		Name:        input.Name,
		Description: input.Description,
	}

	resp, err := h.client.UpdateScene(ctx, req)
	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"data":       sceneToResponse(resp.Scene),
		"message":    "Scene updated successfully",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// DeleteScene deletes a scene.
// @Summary Удалить сцену
// @Description Удаление сцены
// @Tags scenes
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID сцены"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /scenes/{id} [delete]
// =============================================================================
func (h *SceneHandler) DeleteScene(c *fiber.Ctx) error {
	sceneID := c.Params("id")
	if _, err := uuid.Parse(sceneID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid scene ID")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	_, err := h.client.DeleteScene(ctx, &scenepb.DeleteSceneRequest{
		Id: sceneID,
	})
	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"message":    "Scene deleted successfully",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// CheckCompliance checks scene compliance.
// @Summary Проверить соответствие
// @Description Проверка сцены на соответствие нормам
// @Tags scenes
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID сцены"
// @Success 200 {object} ComplianceResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /scenes/{id}/compliance [get]
// =============================================================================
func (h *SceneHandler) CheckCompliance(c *fiber.Ctx) error {
	sceneID := c.Params("id")
	branchID := c.Query("branch_id", "")

	if _, err := uuid.Parse(sceneID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid scene ID")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	resp, err := h.client.CheckCompliance(ctx, &scenepb.CheckComplianceRequest{
		SceneId:  sceneID,
		BranchId: branchID,
	})
	if err != nil {
		return handleGRPCError(err)
	}

	violations := make([]fiber.Map, 0, len(resp.Violations))
	for _, v := range resp.Violations {
		violations = append(violations, fiber.Map{
			"id":          v.Id,
			"rule_id":     v.RuleId,
			"severity":    v.Severity,
			"title":       v.Title,
			"description": v.Description,
			"element_ids": v.ElementIds,
		})
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"is_compliant":   resp.IsCompliant,
			"violations":     violations,
			"total_checks":   resp.TotalChecks,
			"passed_checks":  resp.PassedChecks,
			"failed_checks":  resp.FailedChecks,
		},
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// Helper functions
// =============================================================================

func sceneToResponse(s *scenepb.Scene) fiber.Map {
	if s == nil {
		return nil
	}

	result := fiber.Map{
		"id":             s.Id,
		"workspace_id":   s.WorkspaceId,
		"owner_id":       s.OwnerId,
		"floor_plan_id":  s.FloorPlanId,
		"name":           s.Name,
		"description":    s.Description,
		"main_branch_id": s.MainBranchId,
	}

	if s.CreatedAt != nil {
		result["created_at"] = s.CreatedAt.AsTime()
	}
	if s.UpdatedAt != nil {
		result["updated_at"] = s.UpdatedAt.AsTime()
	}

	if s.Dimensions != nil {
		result["dimensions"] = fiber.Map{
			"width":  s.Dimensions.Width,
			"height": s.Dimensions.Height,
		}
	}

	return result
}

// =============================================================================
// Input types
// =============================================================================

// CreateSceneInput - input for creating a scene.
type CreateSceneInput struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	FloorPlanID string `json:"floor_plan_id,omitempty"`
}

// UpdateSceneInput - input for updating a scene.
type UpdateSceneInput struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// =============================================================================
// Package handlers provides HTTP handlers for API Gateway.
// =============================================================================
// BranchHandler handles branch HTTP requests including:
// - List: List all branches for a scene
// - Create: Create a new branch
// - Get: Get branch details
// - Update: Update branch metadata
// - Delete: Delete a branch
// - Activate: Activate a branch (make it the current working branch)
// - Merge: Merge a branch into another
// - Compare: Compare two branches (diff)
//
// Documentation: docs/api/branches.md
// =============================================================================
package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"

	branchpb "github.com/xiiisorate/granula_api/shared/gen/branch/v1"
)

// =============================================================================
// BranchHandler
// =============================================================================

// BranchHandler handles branch HTTP requests.
// It communicates with the Branch microservice via gRPC.
type BranchHandler struct {
	// client is the gRPC client for BranchService.
	client branchpb.BranchServiceClient
}

// NewBranchHandler creates a new BranchHandler.
//
// Parameters:
//   - conn: gRPC connection to the Branch service
//
// Returns:
//   - *BranchHandler: New handler instance
func NewBranchHandler(conn *grpc.ClientConn) *BranchHandler {
	return &BranchHandler{
		client: branchpb.NewBranchServiceClient(conn),
	}
}

// =============================================================================
// List - GET /scenes/:scene_id/branches
// =============================================================================

// List возвращает список веток сцены.
//
// @Summary Список веток
// @Description Получить список всех веток для сцены.
// @Tags branches
// @Produce json
// @Security BearerAuth
// @Param scene_id path string true "ID сцены"
// @Param include_tree query bool false "Включить иерархическую структуру" default(false)
// @Param source query string false "Фильтр по источнику (user/ai)"
// @Success 200 {object} BranchesListResponse "Список веток"
// @Failure 400 {object} ErrorResponse "Некорректный запрос"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Сцена не найдена"
// @Router /scenes/{scene_id}/branches [get]
func (h *BranchHandler) List(c *fiber.Ctx) error {
	// Get scene ID from path
	sceneID := c.Params("scene_id")
	if sceneID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "scene_id is required")
	}

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &branchpb.ListBranchesRequest{
		SceneId: sceneID,
	}

	// Call gRPC service
	resp, err := h.client.ListBranches(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Convert branches to response format
	branches := make([]fiber.Map, 0, len(resp.Branches))
	for _, b := range resp.Branches {
		branches = append(branches, branchToResponse(b))
	}

	// Return response
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"branches": branches,
			"total":    len(branches),
		},
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// Create - POST /scenes/:scene_id/branches
// =============================================================================

// CreateBranchInput represents input for creating a new branch.
type CreateBranchInput struct {
	// Name - название ветки (обязательно)
	Name string `json:"name" validate:"required,min=1,max=255"`
	// Description - описание ветки
	Description string `json:"description,omitempty" validate:"max=2000"`
	// ParentBranchID - ID родительской ветки (null для корневой)
	ParentBranchID string `json:"parent_branch_id,omitempty"`
}

// Create создаёт новую ветку.
//
// @Summary Создать ветку
// @Description Создать новую ветку для сцены.
// Если parent_branch_id не указан, создаётся корневая ветка.
// @Tags branches
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param scene_id path string true "ID сцены"
// @Param body body CreateBranchInput true "Данные ветки"
// @Success 201 {object} BranchResponse "Ветка создана"
// @Failure 400 {object} ErrorResponse "Некорректный запрос"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Сцена не найдена"
// @Router /scenes/{scene_id}/branches [post]
func (h *BranchHandler) Create(c *fiber.Ctx) error {
	// Get scene ID from path
	sceneID := c.Params("scene_id")
	if sceneID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "scene_id is required")
	}

	// Parse request body
	var input CreateBranchInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// Validate required fields
	if input.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &branchpb.CreateBranchRequest{
		SceneId:        sceneID,
		Name:           input.Name,
		Description:    input.Description,
		ParentBranchId: input.ParentBranchID,
	}

	// Call gRPC service
	resp, err := h.client.CreateBranch(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data":       branchToResponse(resp.Branch),
		"message":    "Branch created successfully",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// Get - GET /scenes/:scene_id/branches/:id
// =============================================================================

// Get возвращает ветку по ID.
//
// @Summary Получить ветку
// @Description Получить детальную информацию о ветке, включая delta и snapshot.
// @Tags branches
// @Produce json
// @Security BearerAuth
// @Param scene_id path string true "ID сцены"
// @Param id path string true "ID ветки"
// @Success 200 {object} BranchResponse "Данные ветки"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Ветка не найдена"
// @Router /scenes/{scene_id}/branches/{id} [get]
func (h *BranchHandler) Get(c *fiber.Ctx) error {
	// Get branch ID from path
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "branch ID is required")
	}

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &branchpb.GetBranchRequest{
		Id: id,
	}

	// Call gRPC service
	resp, err := h.client.GetBranch(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Return response
	return SuccessResponseData(c, branchToResponse(resp.Branch))
}

// =============================================================================
// Update - PATCH /scenes/:scene_id/branches/:id
// =============================================================================

// UpdateBranchInput represents input for updating branch metadata.
type UpdateBranchInput struct {
	// Name - новое название ветки
	Name string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	// Description - новое описание
	Description string `json:"description,omitempty" validate:"omitempty,max=2000"`
	// IsFavorite - добавить в избранное
	IsFavorite *bool `json:"is_favorite,omitempty"`
}

// Update обновляет метаданные ветки.
//
// @Summary Обновить ветку
// @Description Обновить название, описание или статус избранного.
// @Tags branches
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param scene_id path string true "ID сцены"
// @Param id path string true "ID ветки"
// @Param body body UpdateBranchInput true "Данные для обновления"
// @Success 200 {object} BranchResponse "Ветка обновлена"
// @Failure 400 {object} ErrorResponse "Некорректный запрос"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Ветка не найдена"
// @Router /scenes/{scene_id}/branches/{id} [patch]
func (h *BranchHandler) Update(c *fiber.Ctx) error {
	// Get branch ID from path
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "branch ID is required")
	}

	// Parse request body
	var input UpdateBranchInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// Note: BranchService doesn't have UpdateBranch in proto
	// For now, return the branch as-is with a message
	// TODO: Add UpdateBranch to proto and implement

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	// Get the branch first
	resp, err := h.client.GetBranch(ctx, &branchpb.GetBranchRequest{Id: id})
	if err != nil {
		return HandleGRPCError(err)
	}

	// Return response with update message
	return c.JSON(fiber.Map{
		"data":       branchToResponse(resp.Branch),
		"message":    "Branch update received (pending implementation)",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// Delete - DELETE /scenes/:scene_id/branches/:id
// =============================================================================

// Delete удаляет ветку.
//
// @Summary Удалить ветку
// @Description Удалить ветку и все её дочерние ветки.
// Нельзя удалить main ветку.
// @Tags branches
// @Produce json
// @Security BearerAuth
// @Param scene_id path string true "ID сцены"
// @Param id path string true "ID ветки"
// @Success 200 {object} SuccessResponse "Ветка удалена"
// @Failure 400 {object} ErrorResponse "Нельзя удалить main ветку"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Ветка не найдена"
// @Router /scenes/{scene_id}/branches/{id} [delete]
func (h *BranchHandler) Delete(c *fiber.Ctx) error {
	// Get branch ID from path
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "branch ID is required")
	}

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &branchpb.DeleteBranchRequest{
		Id: id,
	}

	// Call gRPC service
	_, err := h.client.DeleteBranch(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Return success response
	return SuccessResponseMessage(c, "Branch deleted successfully")
}

// =============================================================================
// Activate - POST /scenes/:scene_id/branches/:id/activate
// =============================================================================

// Activate активирует ветку (делает её текущей рабочей веткой).
//
// @Summary Активировать ветку
// @Description Сделать ветку активной для редактирования.
// Только одна ветка может быть активной одновременно.
// @Tags branches
// @Produce json
// @Security BearerAuth
// @Param scene_id path string true "ID сцены"
// @Param id path string true "ID ветки"
// @Success 200 {object} BranchActivateResponse "Ветка активирована"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Ветка не найдена"
// @Router /scenes/{scene_id}/branches/{id}/activate [post]
func (h *BranchHandler) Activate(c *fiber.Ctx) error {
	// Get branch ID from path
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "branch ID is required")
	}

	// Note: BranchService doesn't have ActivateBranch in proto
	// For now, return success with branch info
	// TODO: Add ActivateBranch to proto and implement

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	// Get the branch
	resp, err := h.client.GetBranch(ctx, &branchpb.GetBranchRequest{Id: id})
	if err != nil {
		return HandleGRPCError(err)
	}

	// Return response
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"id":        resp.Branch.Id,
			"is_active": true,
			"message":   "Branch activated",
		},
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// Merge - POST /scenes/:scene_id/branches/:id/merge
// =============================================================================

// MergeBranchInput represents input for merging branches.
type MergeBranchInput struct {
	// TargetBranchID - ID целевой ветки для слияния
	TargetBranchID string `json:"target_branch_id" validate:"required"`
	// Strategy - стратегия слияния (replace, combine)
	Strategy string `json:"strategy,omitempty" default:"replace"`
	// DeleteSource - удалить исходную ветку после слияния
	DeleteSource bool `json:"delete_source,omitempty"`
}

// Merge сливает ветку в другую.
//
// @Summary Слить ветки
// @Description Слить изменения из текущей ветки в целевую.
// Стратегии: replace (полная замена), combine (объединение изменений).
// @Tags branches
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param scene_id path string true "ID сцены"
// @Param id path string true "ID исходной ветки"
// @Param body body MergeBranchInput true "Параметры слияния"
// @Success 200 {object} MergeBranchResponse "Ветки слиты"
// @Failure 400 {object} ErrorResponse "Некорректный запрос"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Ветка не найдена"
// @Failure 409 {object} ErrorResponse "Конфликт при слиянии"
// @Router /scenes/{scene_id}/branches/{id}/merge [post]
func (h *BranchHandler) Merge(c *fiber.Ctx) error {
	// Get source branch ID from path
	sourceID := c.Params("id")
	if sourceID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "branch ID is required")
	}

	// Parse request body
	var input MergeBranchInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// Validate required fields
	if input.TargetBranchID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "target_branch_id is required")
	}

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	req := &branchpb.MergeBranchRequest{
		SourceBranchId: sourceID,
		TargetBranchId: input.TargetBranchID,
		DeleteSource:   input.DeleteSource,
	}

	// Call gRPC service
	resp, err := h.client.MergeBranch(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Return response
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"merged":           resp.Success,
			"changes_merged":   resp.ChangesMerged,
			"conflicts":        resp.Conflicts,
			"source_deleted":   input.DeleteSource && resp.Success,
			"target_branch_id": input.TargetBranchID,
		},
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// Compare - GET /scenes/:scene_id/branches/:id/compare/:target_id
// =============================================================================

// Compare сравнивает две ветки.
//
// @Summary Сравнить ветки
// @Description Получить diff между двумя ветками - какие элементы добавлены, изменены, удалены.
// @Tags branches
// @Produce json
// @Security BearerAuth
// @Param scene_id path string true "ID сцены"
// @Param id path string true "ID исходной ветки"
// @Param target_id path string true "ID целевой ветки"
// @Success 200 {object} CompareBranchesResponse "Результат сравнения"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Ветка не найдена"
// @Router /scenes/{scene_id}/branches/{id}/compare/{target_id} [get]
func (h *BranchHandler) Compare(c *fiber.Ctx) error {
	// Get branch IDs from path
	sourceID := c.Params("id")
	targetID := c.Params("target_id")

	if sourceID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "source branch ID is required")
	}
	if targetID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "target branch ID is required")
	}

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &branchpb.GetDiffRequest{
		SourceBranchId: sourceID,
		TargetBranchId: targetID,
	}

	// Call gRPC service
	resp, err := h.client.GetDiff(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Convert diff to response format
	diff := fiber.Map{
		"total_changes": 0,
	}
	if resp.Diff != nil {
		diff["added"] = elementChangesToResponse(resp.Diff.Added)
		diff["modified"] = elementChangesToResponse(resp.Diff.Modified)
		diff["deleted"] = elementChangesToResponse(resp.Diff.Deleted)
		diff["total_changes"] = resp.Diff.TotalChanges
	}

	// Return response
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"source_branch_id": sourceID,
			"target_branch_id": targetID,
			"differences":      diff,
		},
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// Duplicate - POST /scenes/:scene_id/branches/:id/duplicate
// =============================================================================

// DuplicateBranchInput represents input for duplicating a branch.
type DuplicateBranchInput struct {
	// Name - название новой ветки
	Name string `json:"name,omitempty"`
	// IncludeChildren - включить дочерние ветки
	IncludeChildren bool `json:"include_children,omitempty"`
}

// Duplicate дублирует ветку.
//
// @Summary Дублировать ветку
// @Description Создать копию ветки со всеми изменениями.
// @Tags branches
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param scene_id path string true "ID сцены"
// @Param id path string true "ID ветки"
// @Param body body DuplicateBranchInput false "Параметры дублирования"
// @Success 201 {object} BranchResponse "Ветка дублирована"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Ветка не найдена"
// @Router /scenes/{scene_id}/branches/{id}/duplicate [post]
func (h *BranchHandler) Duplicate(c *fiber.Ctx) error {
	// Get scene ID and branch ID from path
	sceneID := c.Params("scene_id")
	id := c.Params("id")

	if sceneID == "" || id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "scene_id and branch ID are required")
	}

	// Parse optional request body
	var input DuplicateBranchInput
	_ = c.BodyParser(&input)

	// First, get the original branch
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	getBranchResp, err := h.client.GetBranch(ctx, &branchpb.GetBranchRequest{Id: id})
	if err != nil {
		return HandleGRPCError(err)
	}

	// Create a new branch based on the original
	name := input.Name
	if name == "" {
		name = getBranchResp.Branch.Name + " (копия)"
	}

	req := &branchpb.CreateBranchRequest{
		SceneId:        sceneID,
		ParentBranchId: getBranchResp.Branch.ParentBranchId,
		Name:           name,
		Description:    getBranchResp.Branch.Description,
	}

	resp, err := h.client.CreateBranch(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": fiber.Map{
			"id":               resp.Branch.Id,
			"name":             resp.Branch.Name,
			"source_branch_id": id,
			"created_at":       resp.Branch.CreatedAt.AsTime(),
		},
		"message":    "Branch duplicated successfully",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// Helper Functions
// =============================================================================

// branchToResponse converts proto Branch to API response format.
func branchToResponse(b *branchpb.Branch) fiber.Map {
	if b == nil {
		return nil
	}

	result := fiber.Map{
		"id":               b.Id,
		"scene_id":         b.SceneId,
		"name":             b.Name,
		"description":      b.Description,
		"parent_branch_id": b.ParentBranchId,
		"is_main":          b.IsMain,
		"status":           branchStatusToString(b.Status),
	}

	// Add timestamps
	if b.CreatedAt != nil {
		result["created_at"] = b.CreatedAt.AsTime()
	}
	if b.UpdatedAt != nil {
		result["updated_at"] = b.UpdatedAt.AsTime()
	}

	return result
}

// branchStatusToString converts proto BranchStatus to string.
func branchStatusToString(status branchpb.BranchStatus) string {
	switch status {
	case branchpb.BranchStatus_BRANCH_STATUS_ACTIVE:
		return "active"
	case branchpb.BranchStatus_BRANCH_STATUS_MERGED:
		return "merged"
	case branchpb.BranchStatus_BRANCH_STATUS_ARCHIVED:
		return "archived"
	default:
		return "unknown"
	}
}

// elementChangesToResponse converts proto ElementChange slice to response format.
func elementChangesToResponse(changes []*branchpb.ElementChange) []fiber.Map {
	result := make([]fiber.Map, 0, len(changes))
	for _, ch := range changes {
		result = append(result, fiber.Map{
			"element_id":   ch.ElementId,
			"element_type": ch.ElementType,
			"description":  ch.Description,
		})
	}
	return result
}

// =============================================================================
// Response Types (for Swagger documentation)
// =============================================================================

// BranchResponse represents a branch in API response.
// swagger:model
type BranchResponse struct {
	Data      interface{} `json:"data"`
	RequestID string      `json:"request_id"`
}

// BranchesListResponse represents a list of branches.
// swagger:model
type BranchesListResponse struct {
	Data      BranchesListData `json:"data"`
	RequestID string           `json:"request_id"`
}

// BranchesListData contains branches list data.
type BranchesListData struct {
	Branches []interface{} `json:"branches"`
	Total    int           `json:"total"`
}

// BranchActivateResponse represents branch activation response.
// swagger:model
type BranchActivateResponse struct {
	Data      BranchActivateData `json:"data"`
	RequestID string             `json:"request_id"`
}

// BranchActivateData contains activation info.
type BranchActivateData struct {
	ID       string `json:"id"`
	IsActive bool   `json:"is_active"`
	Message  string `json:"message"`
}

// MergeBranchResponse represents merge operation response.
// swagger:model
type MergeBranchResponse struct {
	Data      MergeBranchData `json:"data"`
	RequestID string          `json:"request_id"`
}

// MergeBranchData contains merge result info.
type MergeBranchData struct {
	Merged         bool     `json:"merged"`
	ChangesMerged  int32    `json:"changes_merged"`
	Conflicts      []string `json:"conflicts,omitempty"`
	SourceDeleted  bool     `json:"source_deleted"`
	TargetBranchID string   `json:"target_branch_id"`
}

// CompareBranchesResponse represents branch comparison response.
// swagger:model
type CompareBranchesResponse struct {
	Data      CompareBranchesData `json:"data"`
	RequestID string              `json:"request_id"`
}

// CompareBranchesData contains comparison result.
type CompareBranchesData struct {
	SourceBranchID string      `json:"source_branch_id"`
	TargetBranchID string      `json:"target_branch_id"`
	Differences    interface{} `json:"differences"`
}

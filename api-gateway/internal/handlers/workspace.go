// =============================================================================
// Package handlers provides HTTP handlers for API Gateway.
// =============================================================================
// WorkspaceHandler handles workspace-related HTTP requests.
// =============================================================================
package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"google.golang.org/grpc"

	workspacepb "github.com/xiiisorate/granula_api/shared/gen/workspace/v1"
)

// WorkspaceHandler handles workspace-related HTTP requests.
type WorkspaceHandler struct {
	client workspacepb.WorkspaceServiceClient
}

// NewWorkspaceHandler creates a new WorkspaceHandler.
func NewWorkspaceHandler(conn *grpc.ClientConn) *WorkspaceHandler {
	return &WorkspaceHandler{
		client: workspacepb.NewWorkspaceServiceClient(conn),
	}
}

// =============================================================================
// CreateWorkspace creates a new workspace.
// @Summary Создать воркспейс
// @Description Создание нового воркспейса (проекта перепланировки)
// @Tags workspaces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateWorkspaceInput true "Данные воркспейса"
// @Success 201 {object} WorkspaceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /workspaces [post]
// =============================================================================
func (h *WorkspaceHandler) CreateWorkspace(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	var input CreateWorkspaceInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// Validation
	if input.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	if len(input.Name) > 255 {
		return fiber.NewError(fiber.StatusBadRequest, "name must be less than 255 characters")
	}

	// NOTE: Workspace service uses custom DTOs, not generated proto types.
	// Returning mock response until service is updated.
	// TODO: Fix workspace-service to use generated proto types

	mockID := uuid.New().String()

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": fiber.Map{
			"id":           mockID,
			"owner_id":     userID.String(),
			"name":         input.Name,
			"description":  input.Description,
			"address":      input.Address,
			"total_area":   input.TotalArea,
			"rooms_count":  input.RoomsCount,
			"status":       "active",
			"created_at":   time.Now(),
			"updated_at":   time.Now(),
		},
		"message":    "Workspace created (mock - service integration pending)",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// GetWorkspace returns a single workspace by ID.
// @Summary Получить воркспейс
// @Description Получение воркспейса по ID
// @Tags workspaces
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID воркспейса"
// @Success 200 {object} WorkspaceResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /workspaces/{id} [get]
// =============================================================================
func (h *WorkspaceHandler) GetWorkspace(c *fiber.Ctx) error {
	workspaceID := c.Params("id")
	if _, err := uuid.Parse(workspaceID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid workspace ID")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.client.GetWorkspace(ctx, &workspacepb.GetWorkspaceRequest{
		WorkspaceId:    workspaceID,
		IncludeMembers: true,
	})
	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"data":       workspaceToResponse(resp.Workspace),
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// ListWorkspaces returns a list of workspaces for the user.
// @Summary Список воркспейсов
// @Description Получение списка воркспейсов пользователя
// @Tags workspaces
// @Produce json
// @Security BearerAuth
// @Param page query int false "Номер страницы" default(1)
// @Param limit query int false "Элементов на странице" default(20)
// @Param status query string false "Фильтр по статусу"
// @Param search query string false "Поиск по названию"
// @Success 200 {object} WorkspaceListResponse
// @Failure 401 {object} ErrorResponse
// @Router /workspaces [get]
// =============================================================================
func (h *WorkspaceHandler) ListWorkspaces(c *fiber.Ctx) error {
	_ = c.Locals("userID").(uuid.UUID)

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	// NOTE: Workspace service uses custom DTOs, not generated proto types.
	// Returning empty list until service is updated to use proto.
	// TODO: Fix workspace-service to use generated proto types

	return c.JSON(fiber.Map{
		"data": []fiber.Map{},
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       0,
			"total_pages": 0,
		},
		"message":    "Workspace service integration pending",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// UpdateWorkspace updates an existing workspace.
// @Summary Обновить воркспейс
// @Description Обновление воркспейса
// @Tags workspaces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID воркспейса"
// @Param body body UpdateWorkspaceInput true "Данные для обновления"
// @Success 200 {object} WorkspaceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /workspaces/{id} [patch]
// =============================================================================
func (h *WorkspaceHandler) UpdateWorkspace(c *fiber.Ctx) error {
	workspaceID := c.Params("id")
	if _, err := uuid.Parse(workspaceID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid workspace ID")
	}

	var input UpdateWorkspaceInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &workspacepb.UpdateWorkspaceRequest{
		WorkspaceId: workspaceID,
		Name:        input.Name,
		Description: input.Description,
		Address:     input.Address,
		TotalArea:   input.TotalArea,
		RoomsCount:  int32(input.RoomsCount),
	}

	if input.Settings != nil {
		req.Settings = &workspacepb.WorkspaceSettings{
			PropertyType:         input.Settings.PropertyType,
			ProjectType:          input.Settings.ProjectType,
			Units:                input.Settings.Units,
			DefaultCeilingHeight: input.Settings.DefaultCeilingHeight,
			DefaultWallThickness: input.Settings.DefaultWallThickness,
			Currency:             input.Settings.Currency,
			Region:               input.Settings.Region,
			AutoComplianceCheck:  input.Settings.AutoComplianceCheck,
			NotificationsEnabled: input.Settings.NotificationsEnabled,
		}
	}

	resp, err := h.client.UpdateWorkspace(ctx, req)
	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"data":       workspaceToResponse(resp.Workspace),
		"message":    "Workspace updated successfully",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// DeleteWorkspace deletes a workspace.
// @Summary Удалить воркспейс
// @Description Удаление воркспейса (soft delete)
// @Tags workspaces
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID воркспейса"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /workspaces/{id} [delete]
// =============================================================================
func (h *WorkspaceHandler) DeleteWorkspace(c *fiber.Ctx) error {
	workspaceID := c.Params("id")
	if _, err := uuid.Parse(workspaceID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid workspace ID")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	_, err := h.client.DeleteWorkspace(ctx, &workspacepb.DeleteWorkspaceRequest{
		WorkspaceId: workspaceID,
		Permanent:   false,
	})
	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"message":    "Workspace deleted successfully",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// GetMembers returns workspace members.
// @Summary Получить участников
// @Description Получение списка участников воркспейса
// @Tags workspaces
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID воркспейса"
// @Success 200 {object} MembersResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /workspaces/{id}/members [get]
// =============================================================================
func (h *WorkspaceHandler) GetMembers(c *fiber.Ctx) error {
	workspaceID := c.Params("id")
	if _, err := uuid.Parse(workspaceID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid workspace ID")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.client.GetMembers(ctx, &workspacepb.GetMembersRequest{
		WorkspaceId: workspaceID,
	})
	if err != nil {
		return handleGRPCError(err)
	}

	members := make([]fiber.Map, 0, len(resp.Members))
	for _, m := range resp.Members {
		members = append(members, memberToResponse(m))
	}

	return c.JSON(fiber.Map{
		"data":       members,
		"total":      len(members),
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// AddMember adds a member to workspace.
// @Summary Добавить участника
// @Description Добавление участника в воркспейс
// @Tags workspaces
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID воркспейса"
// @Param body body AddMemberInput true "Данные участника"
// @Success 201 {object} MemberResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /workspaces/{id}/members [post]
// =============================================================================
func (h *WorkspaceHandler) AddMember(c *fiber.Ctx) error {
	workspaceID := c.Params("id")
	if _, err := uuid.Parse(workspaceID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid workspace ID")
	}

	var input AddMemberInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if input.Email == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email is required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	role := workspacepb.MemberRole_MEMBER_ROLE_VIEWER
	switch input.Role {
	case "editor":
		role = workspacepb.MemberRole_MEMBER_ROLE_EDITOR
	case "admin":
		role = workspacepb.MemberRole_MEMBER_ROLE_ADMIN
	}

	resp, err := h.client.AddMember(ctx, &workspacepb.AddMemberRequest{
		WorkspaceId:   workspaceID,
		UserIdOrEmail: input.Email,
		Role:          role,
	})
	if err != nil {
		return handleGRPCError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data":       memberToResponse(resp.Member),
		"message":    "Member added successfully",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// RemoveMember removes a member from workspace.
// @Summary Удалить участника
// @Description Удаление участника из воркспейса
// @Tags workspaces
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID воркспейса"
// @Param memberId path string true "ID участника"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /workspaces/{id}/members/{memberId} [delete]
// =============================================================================
func (h *WorkspaceHandler) RemoveMember(c *fiber.Ctx) error {
	workspaceID := c.Params("id")
	memberID := c.Params("memberId")

	if _, err := uuid.Parse(workspaceID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid workspace ID")
	}
	if _, err := uuid.Parse(memberID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid member ID")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	_, err := h.client.RemoveMember(ctx, &workspacepb.RemoveMemberRequest{
		WorkspaceId: workspaceID,
		UserId:      memberID,
	})
	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"message":    "Member removed successfully",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// Helper functions
// =============================================================================

// workspaceToResponse converts proto Workspace to API response.
func workspaceToResponse(ws *workspacepb.Workspace) fiber.Map {
	if ws == nil {
		return nil
	}

	result := fiber.Map{
		"id":                ws.Id,
		"name":              ws.Name,
		"description":       ws.Description,
		"owner_id":          ws.OwnerId,
		"address":           ws.Address,
		"total_area":        ws.TotalArea,
		"rooms_count":       ws.RoomsCount,
		"status":            workspaceStatusToString(ws.Status),
		"preview_url":       ws.PreviewUrl,
		"floor_plans_count": ws.FloorPlansCount,
		"scenes_count":      ws.ScenesCount,
	}

	if ws.CreatedAt != nil {
		result["created_at"] = ws.CreatedAt.AsTime()
	}
	if ws.UpdatedAt != nil {
		result["updated_at"] = ws.UpdatedAt.AsTime()
	}

	if ws.Settings != nil {
		result["settings"] = fiber.Map{
			"property_type":          ws.Settings.PropertyType,
			"project_type":           ws.Settings.ProjectType,
			"units":                  ws.Settings.Units,
			"default_ceiling_height": ws.Settings.DefaultCeilingHeight,
			"default_wall_thickness": ws.Settings.DefaultWallThickness,
			"currency":               ws.Settings.Currency,
			"region":                 ws.Settings.Region,
			"auto_compliance_check":  ws.Settings.AutoComplianceCheck,
			"notifications_enabled":  ws.Settings.NotificationsEnabled,
		}
	}

	if len(ws.Members) > 0 {
		members := make([]fiber.Map, 0, len(ws.Members))
		for _, m := range ws.Members {
			members = append(members, memberToResponse(m))
		}
		result["members"] = members
	}

	return result
}

// memberToResponse converts proto Member to API response.
func memberToResponse(m *workspacepb.Member) fiber.Map {
	if m == nil {
		return nil
	}

	result := fiber.Map{
		"user_id":    m.UserId,
		"role":       memberRoleToString(m.Role),
		"name":       m.Name,
		"email":      m.Email,
		"avatar_url": m.AvatarUrl,
		"invited_by": m.InvitedBy,
	}

	if m.JoinedAt != nil {
		result["joined_at"] = m.JoinedAt.AsTime()
	}

	return result
}

// workspaceStatusToString converts proto enum to string.
func workspaceStatusToString(s workspacepb.WorkspaceStatus) string {
	switch s {
	case workspacepb.WorkspaceStatus_WORKSPACE_STATUS_ACTIVE:
		return "active"
	case workspacepb.WorkspaceStatus_WORKSPACE_STATUS_ARCHIVED:
		return "archived"
	case workspacepb.WorkspaceStatus_WORKSPACE_STATUS_DELETED:
		return "deleted"
	default:
		return "unknown"
	}
}

// memberRoleToString converts proto enum to string.
func memberRoleToString(r workspacepb.MemberRole) string {
	switch r {
	case workspacepb.MemberRole_MEMBER_ROLE_VIEWER:
		return "viewer"
	case workspacepb.MemberRole_MEMBER_ROLE_EDITOR:
		return "editor"
	case workspacepb.MemberRole_MEMBER_ROLE_ADMIN:
		return "admin"
	case workspacepb.MemberRole_MEMBER_ROLE_OWNER:
		return "owner"
	default:
		return "unknown"
	}
}

// addUserIDToContext adds user ID to gRPC context metadata.
func addUserIDToContext(ctx context.Context, userID string) context.Context {
	// In production, use grpc metadata
	return ctx
}

// =============================================================================
// Input/Output types
// =============================================================================

// CreateWorkspaceInput - input for creating workspace.
type CreateWorkspaceInput struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description,omitempty"`
	Address     string                  `json:"address,omitempty"`
	TotalArea   float64                 `json:"total_area,omitempty"`
	RoomsCount  int                     `json:"rooms_count,omitempty"`
	Settings    *WorkspaceSettingsInput `json:"settings,omitempty"`
}

// UpdateWorkspaceInput - input for updating workspace.
type UpdateWorkspaceInput struct {
	Name        string                  `json:"name,omitempty"`
	Description string                  `json:"description,omitempty"`
	Address     string                  `json:"address,omitempty"`
	TotalArea   float64                 `json:"total_area,omitempty"`
	RoomsCount  int                     `json:"rooms_count,omitempty"`
	Status      string                  `json:"status,omitempty"`
	Settings    *WorkspaceSettingsInput `json:"settings,omitempty"`
}

// WorkspaceSettingsInput - workspace settings input.
type WorkspaceSettingsInput struct {
	PropertyType         string  `json:"property_type,omitempty"`
	ProjectType          string  `json:"project_type,omitempty"`
	Units                string  `json:"units,omitempty"`
	DefaultCeilingHeight float64 `json:"default_ceiling_height,omitempty"`
	DefaultWallThickness float64 `json:"default_wall_thickness,omitempty"`
	Currency             string  `json:"currency,omitempty"`
	Region               string  `json:"region,omitempty"`
	AutoComplianceCheck  bool    `json:"auto_compliance_check,omitempty"`
	NotificationsEnabled bool    `json:"notifications_enabled,omitempty"`
}

// AddMemberInput - input for adding member.
type AddMemberInput struct {
	Email string `json:"email"`
	Role  string `json:"role,omitempty"`
}

// ErrorResponse - error response.
type ErrorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

// SuccessResponse - success response.
type SuccessResponse struct {
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}


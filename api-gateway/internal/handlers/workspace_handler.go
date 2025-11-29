// =============================================================================
// Package handlers provides HTTP request handlers for API Gateway.
// =============================================================================
// This package contains all HTTP endpoint handlers that coordinate
// between clients and backend gRPC services.
//
// =============================================================================
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/xiiisorate/granula_api/api-gateway/internal/dto"
)

// =============================================================================
// WorkspaceHandler handles workspace-related HTTP requests.
// =============================================================================

// WorkspaceHandler provides HTTP handlers for workspace operations.
// @Summary Workspace Management
// @Description Endpoints for managing workspaces and members
type WorkspaceHandler struct {
	// workspaceClient pb.WorkspaceServiceClient
}

// NewWorkspaceHandler creates a new WorkspaceHandler.
func NewWorkspaceHandler() *WorkspaceHandler {
	return &WorkspaceHandler{}
}

// CreateWorkspace godoc
// @Summary Создать воркспейс
// @Description Создает новый воркспейс для текущего пользователя
// @Tags workspaces
// @Accept json
// @Produce json
// @Param request body dto.CreateWorkspaceRequest true "Данные воркспейса"
// @Success 201 {object} dto.WorkspaceResponse "Созданный воркспейс"
// @Failure 400 {object} dto.ErrorResponse "Ошибка валидации"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка"
// @Security BearerAuth
// @Router /workspaces [post]
func (h *WorkspaceHandler) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	// TODO: Get user ID from context (set by auth middleware)
	// userID := r.Context().Value("user_id").(string)

	// TODO: Call workspace gRPC service
	// resp, err := h.workspaceClient.CreateWorkspace(r.Context(), &pb.CreateWorkspaceRequest{...})

	// Placeholder response
	response := dto.WorkspaceResponse{
		ID:          "placeholder-id",
		Name:        req.Name,
		Description: req.Description,
		MemberCount: 1,
	}

	respondJSON(w, http.StatusCreated, response)
}

// GetWorkspace godoc
// @Summary Получить воркспейс
// @Description Возвращает информацию о воркспейсе по ID
// @Tags workspaces
// @Accept json
// @Produce json
// @Param id path string true "ID воркспейса" format(uuid)
// @Success 200 {object} dto.WorkspaceResponse "Данные воркспейса"
// @Failure 400 {object} dto.ErrorResponse "Неверный ID"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Воркспейс не найден"
// @Security BearerAuth
// @Router /workspaces/{id} [get]
func (h *WorkspaceHandler) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	// workspaceID := chi.URLParam(r, "id") // Using chi router

	// TODO: Call workspace gRPC service

	// Placeholder response
	response := dto.WorkspaceResponse{
		ID:          "placeholder-id",
		Name:        "Sample Workspace",
		MemberCount: 3,
	}

	respondJSON(w, http.StatusOK, response)
}

// ListWorkspaces godoc
// @Summary Список воркспейсов
// @Description Возвращает список воркспейсов текущего пользователя
// @Tags workspaces
// @Accept json
// @Produce json
// @Param page query int false "Номер страницы" default(1) minimum(1)
// @Param page_size query int false "Размер страницы" default(20) minimum(1) maximum(100)
// @Param name query string false "Фильтр по названию"
// @Success 200 {object} dto.WorkspaceListResponse "Список воркспейсов"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка"
// @Security BearerAuth
// @Router /workspaces [get]
func (h *WorkspaceHandler) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// TODO: Call workspace gRPC service

	// Placeholder response
	response := dto.WorkspaceListResponse{
		Workspaces: []dto.WorkspaceResponse{},
		Pagination: dto.PaginationResponse{
			Page:       page,
			PageSize:   pageSize,
			Total:      0,
			TotalPages: 0,
		},
	}

	respondJSON(w, http.StatusOK, response)
}

// UpdateWorkspace godoc
// @Summary Обновить воркспейс
// @Description Обновляет название и/или описание воркспейса
// @Tags workspaces
// @Accept json
// @Produce json
// @Param id path string true "ID воркспейса" format(uuid)
// @Param request body dto.UpdateWorkspaceRequest true "Данные для обновления"
// @Success 200 {object} dto.WorkspaceResponse "Обновленный воркспейс"
// @Failure 400 {object} dto.ErrorResponse "Ошибка валидации"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Воркспейс не найден"
// @Security BearerAuth
// @Router /workspaces/{id} [patch]
func (h *WorkspaceHandler) UpdateWorkspace(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	// TODO: Call workspace gRPC service

	response := dto.WorkspaceResponse{
		ID:   "placeholder-id",
		Name: req.Name,
	}

	respondJSON(w, http.StatusOK, response)
}

// DeleteWorkspace godoc
// @Summary Удалить воркспейс
// @Description Удаляет воркспейс и все связанные данные (только для владельца)
// @Tags workspaces
// @Accept json
// @Produce json
// @Param id path string true "ID воркспейса" format(uuid)
// @Success 200 {object} dto.SuccessResponse "Успешно удален"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен (только владелец)"
// @Failure 404 {object} dto.ErrorResponse "Воркспейс не найден"
// @Security BearerAuth
// @Router /workspaces/{id} [delete]
func (h *WorkspaceHandler) DeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	// TODO: Call workspace gRPC service

	respondJSON(w, http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "Workspace deleted successfully",
	})
}

// =============================================================================
// Member Management Handlers
// =============================================================================

// GetMembers godoc
// @Summary Список участников воркспейса
// @Description Возвращает список всех участников воркспейса
// @Tags workspaces
// @Accept json
// @Produce json
// @Param id path string true "ID воркспейса" format(uuid)
// @Success 200 {array} dto.MemberResponse "Список участников"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Воркспейс не найден"
// @Security BearerAuth
// @Router /workspaces/{id}/members [get]
func (h *WorkspaceHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	// TODO: Call workspace gRPC service

	members := []dto.MemberResponse{}
	respondJSON(w, http.StatusOK, members)
}

// AddMember godoc
// @Summary Добавить участника
// @Description Добавляет пользователя в воркспейс с указанной ролью
// @Tags workspaces
// @Accept json
// @Produce json
// @Param id path string true "ID воркспейса" format(uuid)
// @Param request body dto.AddMemberRequest true "Данные участника"
// @Success 201 {object} dto.MemberResponse "Добавленный участник"
// @Failure 400 {object} dto.ErrorResponse "Ошибка валидации"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен (требуется admin/owner)"
// @Failure 404 {object} dto.ErrorResponse "Воркспейс или пользователь не найден"
// @Failure 409 {object} dto.ErrorResponse "Пользователь уже участник"
// @Security BearerAuth
// @Router /workspaces/{id}/members [post]
func (h *WorkspaceHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	var req dto.AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	// TODO: Call workspace gRPC service

	member := dto.MemberResponse{
		UserID: req.UserID,
		Role:   req.Role,
	}
	respondJSON(w, http.StatusCreated, member)
}

// UpdateMemberRole godoc
// @Summary Изменить роль участника
// @Description Изменяет роль участника в воркспейсе
// @Tags workspaces
// @Accept json
// @Produce json
// @Param id path string true "ID воркспейса" format(uuid)
// @Param user_id path string true "ID пользователя" format(uuid)
// @Param request body dto.UpdateMemberRoleRequest true "Новая роль"
// @Success 200 {object} dto.MemberResponse "Обновленный участник"
// @Failure 400 {object} dto.ErrorResponse "Ошибка валидации"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен (только owner)"
// @Failure 404 {object} dto.ErrorResponse "Участник не найден"
// @Security BearerAuth
// @Router /workspaces/{id}/members/{user_id} [patch]
func (h *WorkspaceHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	// TODO: Call workspace gRPC service

	member := dto.MemberResponse{
		Role: req.Role,
	}
	respondJSON(w, http.StatusOK, member)
}

// RemoveMember godoc
// @Summary Удалить участника
// @Description Удаляет участника из воркспейса
// @Tags workspaces
// @Accept json
// @Produce json
// @Param id path string true "ID воркспейса" format(uuid)
// @Param user_id path string true "ID пользователя" format(uuid)
// @Success 200 {object} dto.SuccessResponse "Успешно удален"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Участник не найден"
// @Security BearerAuth
// @Router /workspaces/{id}/members/{user_id} [delete]
func (h *WorkspaceHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	// TODO: Call workspace gRPC service

	respondJSON(w, http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "Member removed successfully",
	})
}

// =============================================================================
// Helper Functions
// =============================================================================

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError writes an error response.
func respondError(w http.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, dto.ErrorResponse{
		Code:    code,
		Message: message,
	})
}


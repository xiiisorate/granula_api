// =============================================================================
// Floor Plan Handler - HTTP handlers for floor plan operations.
// =============================================================================
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/xiiisorate/granula_api/api-gateway/internal/dto"
)

// =============================================================================
// FloorPlanHandler handles floor plan-related HTTP requests.
// =============================================================================

// FloorPlanHandler provides HTTP handlers for floor plan operations.
type FloorPlanHandler struct {
	// floorPlanClient pb.FloorPlanServiceClient
	// aiClient        pb.AIServiceClient
}

// NewFloorPlanHandler creates a new FloorPlanHandler.
func NewFloorPlanHandler() *FloorPlanHandler {
	return &FloorPlanHandler{}
}

// UploadFloorPlan godoc
// @Summary Загрузить планировку
// @Description Загружает изображение планировки и создает запись в системе
// @Tags floor-plans
// @Accept multipart/form-data
// @Produce json
// @Param workspace_id formData string true "ID воркспейса" format(uuid)
// @Param name formData string true "Название планировки"
// @Param description formData string false "Описание"
// @Param address formData string false "Адрес объекта"
// @Param file formData file true "Изображение планировки (PNG, JPG, PDF)"
// @Success 201 {object} dto.FloorPlanResponse "Загруженная планировка"
// @Failure 400 {object} dto.ErrorResponse "Ошибка валидации или формата файла"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 413 {object} dto.ErrorResponse "Файл слишком большой (макс. 10MB)"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка"
// @Security BearerAuth
// @Router /floor-plans [post]
func (h *FloorPlanHandler) UploadFloorPlan(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_FORM", "Failed to parse form data")
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "NO_FILE", "No file uploaded")
		return
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		respondError(w, http.StatusBadRequest, "INVALID_FILE_TYPE", "Only PNG, JPG, and PDF files are allowed")
		return
	}

	// Get form fields
	workspaceID := r.FormValue("workspace_id")
	name := r.FormValue("name")
	description := r.FormValue("description")
	address := r.FormValue("address")

	// TODO: Upload file to MinIO/S3
	// TODO: Call floor plan gRPC service

	response := dto.FloorPlanResponse{
		ID:          "placeholder-id",
		WorkspaceID: workspaceID,
		Name:        name,
		Description: description,
		Address:     address,
		Status:      "uploaded",
	}

	respondJSON(w, http.StatusCreated, response)
}

// GetFloorPlan godoc
// @Summary Получить планировку
// @Description Возвращает информацию о планировке по ID
// @Tags floor-plans
// @Accept json
// @Produce json
// @Param id path string true "ID планировки" format(uuid)
// @Success 200 {object} dto.FloorPlanResponse "Данные планировки"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Планировка не найдена"
// @Security BearerAuth
// @Router /floor-plans/{id} [get]
func (h *FloorPlanHandler) GetFloorPlan(w http.ResponseWriter, r *http.Request) {
	// TODO: Call floor plan gRPC service

	response := dto.FloorPlanResponse{
		ID:     "placeholder-id",
		Status: "recognized",
	}

	respondJSON(w, http.StatusOK, response)
}

// ListFloorPlans godoc
// @Summary Список планировок
// @Description Возвращает список планировок в воркспейсе
// @Tags floor-plans
// @Accept json
// @Produce json
// @Param workspace_id query string true "ID воркспейса" format(uuid)
// @Param page query int false "Номер страницы" default(1)
// @Param page_size query int false "Размер страницы" default(20)
// @Param status query string false "Фильтр по статусу" Enums(uploaded, processing, recognized, failed)
// @Success 200 {object} dto.FloorPlanListResponse "Список планировок"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Security BearerAuth
// @Router /floor-plans [get]
func (h *FloorPlanHandler) ListFloorPlans(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// TODO: Call floor plan gRPC service

	response := dto.FloorPlanListResponse{
		FloorPlans: []dto.FloorPlanResponse{},
		Pagination: dto.PaginationResponse{
			Page:     page,
			PageSize: pageSize,
		},
	}

	respondJSON(w, http.StatusOK, response)
}

// DeleteFloorPlan godoc
// @Summary Удалить планировку
// @Description Удаляет планировку и связанные данные
// @Tags floor-plans
// @Accept json
// @Produce json
// @Param id path string true "ID планировки" format(uuid)
// @Success 200 {object} dto.SuccessResponse "Успешно удалена"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещен"
// @Failure 404 {object} dto.ErrorResponse "Планировка не найдена"
// @Security BearerAuth
// @Router /floor-plans/{id} [delete]
func (h *FloorPlanHandler) DeleteFloorPlan(w http.ResponseWriter, r *http.Request) {
	// TODO: Call floor plan gRPC service

	respondJSON(w, http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "Floor plan deleted successfully",
	})
}

// ProcessFloorPlan godoc
// @Summary Запустить распознавание
// @Description Запускает AI распознавание загруженной планировки
// @Tags floor-plans
// @Accept json
// @Produce json
// @Param id path string true "ID планировки" format(uuid)
// @Param request body dto.ProcessFloorPlanRequest true "Параметры обработки"
// @Success 202 {object} dto.RecognitionStatusResponse "Задача запущена"
// @Failure 400 {object} dto.ErrorResponse "Ошибка валидации"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 404 {object} dto.ErrorResponse "Планировка не найдена"
// @Failure 409 {object} dto.ErrorResponse "Планировка уже обрабатывается"
// @Security BearerAuth
// @Router /floor-plans/{id}/process [post]
func (h *FloorPlanHandler) ProcessFloorPlan(w http.ResponseWriter, r *http.Request) {
	var req dto.ProcessFloorPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Allow empty body - use defaults
		req = dto.ProcessFloorPlanRequest{
			CreateScene:   true,
			RunCompliance: true,
		}
	}

	// TODO: Call AI gRPC service for recognition

	response := dto.RecognitionStatusResponse{
		TaskID:      "task-placeholder-id",
		FloorPlanID: "floor-plan-id",
		Status:      "queued",
		Progress:    0,
	}

	respondJSON(w, http.StatusAccepted, response)
}

// GetProcessingStatus godoc
// @Summary Статус обработки
// @Description Возвращает статус задачи распознавания планировки
// @Tags floor-plans
// @Accept json
// @Produce json
// @Param id path string true "ID планировки" format(uuid)
// @Param task_id path string true "ID задачи" format(uuid)
// @Success 200 {object} dto.RecognitionStatusResponse "Статус задачи"
// @Failure 401 {object} dto.ErrorResponse "Не авторизован"
// @Failure 404 {object} dto.ErrorResponse "Задача не найдена"
// @Security BearerAuth
// @Router /floor-plans/{id}/process/{task_id} [get]
func (h *FloorPlanHandler) GetProcessingStatus(w http.ResponseWriter, r *http.Request) {
	// TODO: Call AI gRPC service

	response := dto.RecognitionStatusResponse{
		TaskID:      "task-id",
		FloorPlanID: "floor-plan-id",
		Status:      "processing",
		Progress:    65,
		CurrentStep: "Распознавание комнат",
	}

	respondJSON(w, http.StatusOK, response)
}

// =============================================================================
// Helper Functions
// =============================================================================

// isValidImageType checks if the content type is a valid image format.
func isValidImageType(contentType string) bool {
	validTypes := map[string]bool{
		"image/png":       true,
		"image/jpeg":      true,
		"image/jpg":       true,
		"application/pdf": true,
	}
	return validTypes[contentType]
}


// =============================================================================
// Package handlers provides HTTP handlers for API Gateway.
// =============================================================================
// FloorPlanHandler handles floor plan HTTP requests including:
// - Upload: Upload new floor plan images (BTI, scans, photos, sketches)
// - List: List floor plans in a workspace
// - Get: Get floor plan details
// - Update: Update floor plan metadata
// - Delete: Delete a floor plan
// - StartRecognition: Start AI recognition process
// - GetRecognitionStatus: Get recognition job status
// - GetDownloadURL: Get presigned download URL
//
// Documentation: docs/api/floor-plans.md
// =============================================================================
package handlers

import (
	"context"
	"io"
	"time"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"

	floorplanpb "github.com/xiiisorate/granula_api/shared/gen/floorplan/v1"
)

// =============================================================================
// FloorPlanHandler
// =============================================================================

// FloorPlanHandler handles floor plan HTTP requests.
// It communicates with the FloorPlan microservice via gRPC.
type FloorPlanHandler struct {
	// client is the gRPC client for FloorPlanService.
	client floorplanpb.FloorPlanServiceClient
}

// NewFloorPlanHandler creates a new FloorPlanHandler.
//
// Parameters:
//   - conn: gRPC connection to the FloorPlan service
//
// Returns:
//   - *FloorPlanHandler: New handler instance
func NewFloorPlanHandler(conn *grpc.ClientConn) *FloorPlanHandler {
	return &FloorPlanHandler{
		client: floorplanpb.NewFloorPlanServiceClient(conn),
	}
}

// =============================================================================
// Upload - POST /floor-plans
// =============================================================================

// Upload загружает новую планировку.
//
// @Summary Загрузить планировку
// @Description Загрузка изображения планировки (BTI, скан, фото, эскиз).
// Поддерживаемые форматы: PDF, JPEG, PNG, TIFF, WEBP.
// Максимальный размер: 50MB для PDF/TIFF, 20MB для остальных.
// @Tags floor-plans
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Файл планировки"
// @Param workspace_id formData string true "ID воркспейса"
// @Param name formData string false "Название (по умолчанию из имени файла)"
// @Param description formData string false "Описание"
// @Success 201 {object} FloorPlanResponse "Планировка успешно загружена"
// @Failure 400 {object} ErrorResponse "Некорректный запрос"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 413 {object} ErrorResponse "Файл слишком большой"
// @Router /floor-plans [post]
func (h *FloorPlanHandler) Upload(c *fiber.Ctx) error {
	// Extract user ID from context (set by auth middleware)
	userID, err := GetUserIDFromContextRequired(c)
	if err != nil {
		return err
	}

	// Get required workspace_id parameter
	workspaceID := c.FormValue("workspace_id")
	if workspaceID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "workspace_id is required")
	}

	// Get file from form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "file is required")
	}

	// Open and read file
	file, err := fileHeader.Open()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "failed to open uploaded file")
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "failed to read uploaded file")
	}

	// Get optional parameters with defaults
	name := c.FormValue("name")
	if name == "" {
		name = fileHeader.Filename
	}
	description := c.FormValue("description")
	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 60*time.Second)
	defer cancel()

	req := &floorplanpb.UploadRequest{
		WorkspaceId: workspaceID,
		OwnerId:     userID,
		Name:        name,
		Description: description,
		FileName:    fileHeader.Filename,
		MimeType:    mimeType,
		FileData:    fileData,
	}

	// Call gRPC service
	resp, err := h.client.Upload(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data":       floorPlanToResponse(resp.FloorPlan),
		"message":    "Floor plan uploaded successfully",
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// List - GET /floor-plans
// =============================================================================

// List возвращает список планировок воркспейса.
//
// @Summary Список планировок
// @Description Получить список планировок воркспейса с пагинацией.
// @Tags floor-plans
// @Produce json
// @Security BearerAuth
// @Param workspace_id query string true "ID воркспейса"
// @Param limit query int false "Количество записей" default(20)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {object} FloorPlansListResponse "Список планировок"
// @Failure 400 {object} ErrorResponse "workspace_id обязателен"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Router /floor-plans [get]
func (h *FloorPlanHandler) List(c *fiber.Ctx) error {
	// Get required workspace_id
	workspaceID := c.Query("workspace_id")
	if workspaceID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "workspace_id is required")
	}

	// Get pagination parameters
	pagination := GetPaginationParams(c)

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &floorplanpb.ListRequest{
		WorkspaceId: workspaceID,
		Limit:       int32(pagination.Limit),
		Offset:      int32(pagination.Offset),
	}

	// Call gRPC service
	resp, err := h.client.List(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Convert floor plans to response format
	items := make([]fiber.Map, 0, len(resp.FloorPlans))
	for _, fp := range resp.FloorPlans {
		items = append(items, floorPlanToResponse(fp))
	}

	// Return paginated response
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"items": items,
			"total": resp.Total,
		},
		"pagination": PaginationResponse(pagination.Page, pagination.Limit, int(resp.Total)),
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// Get - GET /floor-plans/:id
// =============================================================================

// Get возвращает планировку по ID.
//
// @Summary Получить планировку
// @Description Получить детальную информацию о планировке, включая данные распознавания.
// @Tags floor-plans
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID планировки"
// @Success 200 {object} FloorPlanResponse "Данные планировки"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Планировка не найдена"
// @Router /floor-plans/{id} [get]
func (h *FloorPlanHandler) Get(c *fiber.Ctx) error {
	// Get floor plan ID from path
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "floor plan ID is required")
	}

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &floorplanpb.GetRequest{
		Id: id,
	}

	// Call gRPC service
	resp, err := h.client.Get(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Return response
	return SuccessResponseData(c, floorPlanToResponse(resp.FloorPlan))
}

// =============================================================================
// Update - PATCH /floor-plans/:id
// =============================================================================

// UpdateFloorPlanInput represents input for updating floor plan metadata.
type UpdateFloorPlanInput struct {
	// Name - новое название планировки (опционально)
	Name string `json:"name,omitempty"`
	// Description - новое описание планировки (опционально)
	Description string `json:"description,omitempty"`
}

// Update обновляет метаданные планировки.
//
// @Summary Обновить планировку
// @Description Обновить название и описание планировки.
// @Tags floor-plans
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID планировки"
// @Param body body UpdateFloorPlanInput true "Данные для обновления"
// @Success 200 {object} FloorPlanResponse "Планировка обновлена"
// @Failure 400 {object} ErrorResponse "Некорректный запрос"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Планировка не найдена"
// @Router /floor-plans/{id} [patch]
func (h *FloorPlanHandler) Update(c *fiber.Ctx) error {
	// Get floor plan ID from path
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "floor plan ID is required")
	}

	// Parse request body
	var input UpdateFloorPlanInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &floorplanpb.UpdateRequest{
		Id:          id,
		Name:        input.Name,
		Description: input.Description,
	}

	// Call gRPC service
	resp, err := h.client.Update(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Return response
	return SuccessResponseDataMessage(c, floorPlanToResponse(resp.FloorPlan), "Floor plan updated successfully")
}

// =============================================================================
// Delete - DELETE /floor-plans/:id
// =============================================================================

// Delete удаляет планировку.
//
// @Summary Удалить планировку
// @Description Удалить планировку. Созданные на её основе сцены не удаляются.
// @Tags floor-plans
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID планировки"
// @Success 200 {object} SuccessResponse "Планировка удалена"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Планировка не найдена"
// @Router /floor-plans/{id} [delete]
func (h *FloorPlanHandler) Delete(c *fiber.Ctx) error {
	// Get floor plan ID from path
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "floor plan ID is required")
	}

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &floorplanpb.DeleteRequest{
		Id: id,
	}

	// Call gRPC service
	_, err := h.client.Delete(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Return success response
	return SuccessResponseMessage(c, "Floor plan deleted successfully")
}

// =============================================================================
// StartRecognition - POST /floor-plans/:id/recognize
// =============================================================================

// RecognitionOptionsInput represents options for floor plan recognition.
type RecognitionOptionsInput struct {
	// DetectLoadBearing - определять несущие стены
	DetectLoadBearing bool `json:"detect_load_bearing,omitempty"`
	// DetectWetZones - определять мокрые зоны
	DetectWetZones bool `json:"detect_wet_zones,omitempty"`
	// DetectFurniture - определять мебель
	DetectFurniture bool `json:"detect_furniture,omitempty"`
	// Scale - масштаб чертежа (если известен)
	Scale float32 `json:"scale,omitempty"`
	// Orientation - поворот в градусах
	Orientation int32 `json:"orientation,omitempty"`
	// DetailLevel - уровень детализации (1-3)
	DetailLevel int32 `json:"detail_level,omitempty"`
}

// StartRecognition запускает AI распознавание планировки.
//
// @Summary Запустить распознавание
// @Description Запустить AI распознавание элементов планировки (стены, комнаты, проёмы).
// Процесс асинхронный - статус можно отслеживать через GetRecognitionStatus.
// @Tags floor-plans
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID планировки"
// @Param body body RecognitionOptionsInput false "Опции распознавания"
// @Success 200 {object} RecognitionJobResponse "Задача распознавания создана"
// @Failure 400 {object} ErrorResponse "Некорректный запрос"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Планировка не найдена"
// @Router /floor-plans/{id}/recognize [post]
func (h *FloorPlanHandler) StartRecognition(c *fiber.Ctx) error {
	// Get floor plan ID from path
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "floor plan ID is required")
	}

	// Parse optional request body
	var input RecognitionOptionsInput
	// Ignore error - body is optional
	_ = c.BodyParser(&input)

	// Set default values if not specified
	if input.DetailLevel == 0 {
		input.DetailLevel = 2
	}

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
	defer cancel()

	req := &floorplanpb.StartRecognitionRequest{
		FloorPlanId: id,
		Options: &floorplanpb.RecognitionOptions{
			DetectLoadBearing: input.DetectLoadBearing,
			DetectWetZones:    input.DetectWetZones,
			DetectFurniture:   input.DetectFurniture,
			Scale:             input.Scale,
			Orientation:       input.Orientation,
			DetailLevel:       input.DetailLevel,
		},
	}

	// Call gRPC service
	resp, err := h.client.StartRecognition(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Return response with job ID
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"job_id":  resp.JobId,
			"status":  "pending",
			"message": "Recognition started",
		},
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// GetRecognitionStatus - GET /floor-plans/:id/recognition-status
// =============================================================================

// GetRecognitionStatus возвращает статус задачи распознавания.
//
// @Summary Статус распознавания
// @Description Получить текущий статус задачи AI распознавания.
// @Tags floor-plans
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID планировки"
// @Success 200 {object} RecognitionStatusResponse "Статус распознавания"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Задача не найдена"
// @Router /floor-plans/{id}/recognition-status [get]
func (h *FloorPlanHandler) GetRecognitionStatus(c *fiber.Ctx) error {
	// Get floor plan ID from path
	// Note: The job_id is derived from floor_plan_id in the service
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "floor plan ID is required")
	}

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &floorplanpb.GetRecognitionStatusRequest{
		JobId: id, // Using floor plan ID as job ID
	}

	// Call gRPC service
	resp, err := h.client.GetRecognitionStatus(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Return response
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"job_id":   resp.JobId,
			"status":   resp.Status,
			"progress": resp.Progress,
			"error":    resp.Error,
		},
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// GetDownloadURL - GET /floor-plans/:id/download-url
// =============================================================================

// GetDownloadURL возвращает presigned URL для скачивания файла.
//
// @Summary URL для скачивания
// @Description Получить временный URL для скачивания файла планировки.
// @Tags floor-plans
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID планировки"
// @Success 200 {object} DownloadURLResponse "URL для скачивания"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Планировка не найдена"
// @Router /floor-plans/{id}/download-url [get]
func (h *FloorPlanHandler) GetDownloadURL(c *fiber.Ctx) error {
	// Get floor plan ID from path
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "floor plan ID is required")
	}

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	req := &floorplanpb.GetDownloadURLRequest{
		FloorPlanId: id,
	}

	// Call gRPC service
	resp, err := h.client.GetDownloadURL(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Return response
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"url":        resp.Url,
			"expires_in": resp.ExpiresIn,
		},
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// Helper Functions
// =============================================================================

// floorPlanToResponse converts proto FloorPlan to API response format.
// This function handles null safety and time conversion.
func floorPlanToResponse(fp *floorplanpb.FloorPlan) fiber.Map {
	if fp == nil {
		return nil
	}

	result := fiber.Map{
		"id":           fp.Id,
		"workspace_id": fp.WorkspaceId,
		"owner_id":     fp.OwnerId,
		"name":         fp.Name,
		"description":  fp.Description,
		"status":       floorPlanStatusToString(fp.Status),
	}

	// Add optional fields
	if fp.RecognitionJobId != nil {
		result["recognition_job_id"] = *fp.RecognitionJobId
	}
	if fp.SceneId != nil {
		result["scene_id"] = *fp.SceneId
	}

	// Add file info if present
	if fp.FileInfo != nil {
		result["file_info"] = fiber.Map{
			"id":            fp.FileInfo.Id,
			"original_name": fp.FileInfo.OriginalName,
			"mime_type":     fp.FileInfo.MimeType,
			"size":          fp.FileInfo.Size,
			"width":         fp.FileInfo.Width,
			"height":        fp.FileInfo.Height,
		}
	}

	// Add timestamps
	if fp.CreatedAt != nil {
		result["created_at"] = fp.CreatedAt.AsTime()
	}
	if fp.UpdatedAt != nil {
		result["updated_at"] = fp.UpdatedAt.AsTime()
	}

	return result
}

// floorPlanStatusToString converts proto FloorPlanStatus to string.
func floorPlanStatusToString(status floorplanpb.FloorPlanStatus) string {
	switch status {
	case floorplanpb.FloorPlanStatus_FLOOR_PLAN_STATUS_UPLOADED:
		return "uploaded"
	case floorplanpb.FloorPlanStatus_FLOOR_PLAN_STATUS_PROCESSING:
		return "processing"
	case floorplanpb.FloorPlanStatus_FLOOR_PLAN_STATUS_RECOGNIZED:
		return "recognized"
	case floorplanpb.FloorPlanStatus_FLOOR_PLAN_STATUS_CONFIRMED:
		return "confirmed"
	case floorplanpb.FloorPlanStatus_FLOOR_PLAN_STATUS_FAILED:
		return "failed"
	default:
		return "unknown"
	}
}

// =============================================================================
// Response Types (for Swagger documentation)
// =============================================================================

// FloorPlanResponse represents a floor plan in API response.
// swagger:model
type FloorPlanResponse struct {
	Data      interface{} `json:"data"`
	RequestID string      `json:"request_id"`
}

// FloorPlansListResponse represents a list of floor plans.
// swagger:model
type FloorPlansListResponse struct {
	Data       FloorPlansListData `json:"data"`
	Pagination fiber.Map          `json:"pagination"`
	RequestID  string             `json:"request_id"`
}

// FloorPlansListData contains floor plans list data.
type FloorPlansListData struct {
	Items []interface{} `json:"items"`
	Total int           `json:"total"`
}

// RecognitionJobResponse represents recognition job creation response.
// swagger:model
type RecognitionJobResponse struct {
	Data      RecognitionJobData `json:"data"`
	RequestID string             `json:"request_id"`
}

// RecognitionJobData contains job info.
type RecognitionJobData struct {
	JobID   string `json:"job_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// RecognitionStatusResponse represents recognition status.
// swagger:model
type RecognitionStatusResponse struct {
	Data      RecognitionStatusData `json:"data"`
	RequestID string                `json:"request_id"`
}

// RecognitionStatusData contains status info.
type RecognitionStatusData struct {
	JobID    string `json:"job_id"`
	Status   string `json:"status"`
	Progress int32  `json:"progress"`
	Error    string `json:"error,omitempty"`
}

// DownloadURLResponse represents download URL response.
// swagger:model
type DownloadURLResponse struct {
	Data      DownloadURLData `json:"data"`
	RequestID string          `json:"request_id"`
}

// DownloadURLData contains download URL info.
type DownloadURLData struct {
	URL       string `json:"url"`
	ExpiresIn int32  `json:"expires_in"`
}

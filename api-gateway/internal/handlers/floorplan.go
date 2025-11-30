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

	aipb "github.com/xiiisorate/granula_api/shared/gen/ai/v1"
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
	// aiClient is the gRPC client for AIService (for recognition results).
	aiClient aipb.AIServiceClient
}

// NewFloorPlanHandler creates a new FloorPlanHandler.
//
// Parameters:
//   - conn: gRPC connection to the FloorPlan service
//   - aiConn: gRPC connection to the AI service (for recognition results)
//
// Returns:
//   - *FloorPlanHandler: New handler instance
func NewFloorPlanHandler(conn *grpc.ClientConn, aiConn *grpc.ClientConn) *FloorPlanHandler {
	return &FloorPlanHandler{
		client:   floorplanpb.NewFloorPlanServiceClient(conn),
		aiClient: aipb.NewAIServiceClient(aiConn),
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

	// Convert to response
	result := floorPlanToResponse(resp.FloorPlan)

	// If floor plan has recognition job, fetch the result directly from AI Service
	if resp.FloorPlan != nil && resp.FloorPlan.RecognitionJobId != nil && *resp.FloorPlan.RecognitionJobId != "" {
		aiResp, err := h.aiClient.GetRecognitionStatus(ctx, &aipb.GetRecognitionStatusRequest{
			JobId: *resp.FloorPlan.RecognitionJobId,
		})
		if err == nil && aiResp.Status == aipb.JobStatus_JOB_STATUS_COMPLETED && aiResp.Scene != nil {
			// Use the same format as /ai/recognize/{job_id}/status
			result["model"] = convertAISceneToResponse(aiResp.Scene)
		}
	}

	// Return response
	return SuccessResponseData(c, result)
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
// @Description Получить текущий статус задачи AI распознавания и модель планировки.
// @Tags floor-plans
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID планировки"
// @Success 200 {object} RecognitionStatusResponse "Статус распознавания с моделью"
// @Failure 401 {object} ErrorResponse "Не авторизован"
// @Failure 404 {object} ErrorResponse "Задача не найдена"
// @Router /floor-plans/{id}/recognition-status [get]
func (h *FloorPlanHandler) GetRecognitionStatus(c *fiber.Ctx) error {
	// Get floor plan ID from path
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "floor plan ID is required")
	}

	// Create gRPC request
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	// First, get floor plan to retrieve recognition_job_id
	getReq := &floorplanpb.GetRequest{Id: id}
	fpResp, err := h.client.Get(ctx, getReq)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Check if floor plan has recognition job
	if fpResp.FloorPlan.RecognitionJobId == nil || *fpResp.FloorPlan.RecognitionJobId == "" {
		return fiber.NewError(fiber.StatusNotFound, "no recognition job found for this floor plan")
	}

	req := &floorplanpb.GetRecognitionStatusRequest{
		JobId: *fpResp.FloorPlan.RecognitionJobId,
	}

	// Call gRPC service
	resp, err := h.client.GetRecognitionStatus(ctx, req)
	if err != nil {
		return HandleGRPCError(err)
	}

	// Build response
	data := fiber.Map{
		"job_id":   resp.JobId,
		"status":   resp.Status,
		"progress": resp.Progress,
		"error":    resp.Error,
	}

	// Include model if available
	if resp.Model != nil {
		data["model"] = recognitionModelToResponse(resp.Model)
	}

	// Return response
	return c.JSON(fiber.Map{
		"data":       data,
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

// convertAISceneToResponse converts AI recognized scene to API response.
// Uses the same format as /ai/recognize/{job_id}/status for consistency.
func convertAISceneToResponse(scene *aipb.RecognizedScene) fiber.Map {
	if scene == nil {
		return nil
	}

	// Convert walls
	walls := make([]fiber.Map, 0, len(scene.Walls))
	for _, w := range scene.Walls {
		wall := fiber.Map{
			"temp_id":                 w.TempId,
			"thickness":               w.Thickness,
			"is_load_bearing":         w.IsLoadBearing,
			"confidence":              w.Confidence,
			"load_bearing_confidence": w.LoadBearingConfidence,
		}
		if w.Start != nil {
			wall["start"] = fiber.Map{"x": w.Start.X, "y": w.Start.Y}
		}
		if w.End != nil {
			wall["end"] = fiber.Map{"x": w.End.X, "y": w.End.Y}
		}
		walls = append(walls, wall)
	}

	// Convert rooms
	rooms := make([]fiber.Map, 0, len(scene.Rooms))
	for _, r := range scene.Rooms {
		room := fiber.Map{
			"temp_id":     r.TempId,
			"type":        r.Type.String(),
			"area":        r.Area,
			"is_wet_zone": r.IsWetZone,
			"confidence":  r.Confidence,
			"wall_ids":    r.WallIds,
		}
		if r.Boundary != nil && len(r.Boundary.Vertices) > 0 {
			vertices := make([]fiber.Map, 0, len(r.Boundary.Vertices))
			for _, v := range r.Boundary.Vertices {
				vertices = append(vertices, fiber.Map{"x": v.X, "y": v.Y})
			}
			room["boundary"] = vertices
		}
		rooms = append(rooms, room)
	}

	// Convert openings
	openings := make([]fiber.Map, 0, len(scene.Openings))
	for _, o := range scene.Openings {
		opening := fiber.Map{
			"temp_id":    o.TempId,
			"type":       o.Type.String(),
			"width":      o.Width,
			"wall_id":    o.WallId,
			"confidence": o.Confidence,
		}
		if o.Position != nil {
			opening["position"] = fiber.Map{"x": o.Position.X, "y": o.Position.Y}
		}
		openings = append(openings, opening)
	}

	// Convert elements (furniture, etc)
	elements := make([]fiber.Map, 0, len(scene.Elements))
	for _, e := range scene.Elements {
		elem := fiber.Map{
			"temp_id":      e.TempId,
			"element_type": e.ElementType,
			"room_id":      e.RoomId,
			"confidence":   e.Confidence,
			"rotation":     e.Rotation,
		}
		if e.Position != nil {
			elem["position"] = fiber.Map{"x": e.Position.X, "y": e.Position.Y}
		}
		if e.Dimensions != nil {
			elem["dimensions"] = fiber.Map{
				"width":  e.Dimensions.Width,
				"height": e.Dimensions.Height,
			}
		}
		elements = append(elements, elem)
	}

	result := fiber.Map{
		"total_area": scene.TotalArea,
		"walls":      walls,
		"rooms":      rooms,
		"openings":   openings,
		"elements":   elements,
	}

	// Add dimensions if present
	if scene.Dimensions != nil {
		result["dimensions"] = fiber.Map{
			"width":  scene.Dimensions.Width,
			"height": scene.Dimensions.Height,
		}
	}

	// Add metadata if present
	if scene.Metadata != nil {
		result["metadata"] = fiber.Map{
			"model_version":      scene.Metadata.ModelVersion,
			"processing_time_ms": scene.Metadata.ProcessingTimeMs,
		}
	}

	return result
}

// recognitionModelToResponse converts proto recognition model to API response.
func recognitionModelToResponse(model *floorplanpb.RecognitionModel) fiber.Map {
	if model == nil {
		return nil
	}

	result := fiber.Map{
		"confidence":         model.Confidence,
		"total_area":         model.TotalArea,
		"warnings":           model.Warnings,
		"processing_time_ms": model.ProcessingTimeMs,
	}

	// Add bounds
	if model.Bounds != nil {
		result["bounds"] = fiber.Map{
			"width":  model.Bounds.Width,
			"height": model.Bounds.Height,
			"depth":  model.Bounds.Depth,
		}
	}

	// Add recognition metadata
	if model.Recognition != nil {
		result["recognition"] = fiber.Map{
			"source_type":     model.Recognition.SourceType,
			"quality":         model.Recognition.Quality,
			"scale":           model.Recognition.Scale,
			"orientation":     model.Recognition.Orientation,
			"has_dimensions":  model.Recognition.HasDimensions,
			"has_annotations": model.Recognition.HasAnnotations,
			"building_type":   model.Recognition.BuildingType,
		}
	}

	// Add elements
	if model.Elements != nil {
		elements := fiber.Map{}

		// Convert walls
		walls := make([]fiber.Map, 0, len(model.Elements.Walls))
		for _, w := range model.Elements.Walls {
			wall := fiber.Map{
				"id":        w.Id,
				"type":      w.Type,
				"name":      w.Name,
				"height":    w.Height,
				"thickness": w.Thickness,
			}
			if w.Start != nil {
				wall["start"] = fiber.Map{"x": w.Start.X, "y": w.Start.Y, "z": w.Start.Z}
			}
			if w.End != nil {
				wall["end"] = fiber.Map{"x": w.End.X, "y": w.End.Y, "z": w.End.Z}
			}
			if w.Properties != nil {
				wall["properties"] = fiber.Map{
					"is_load_bearing": w.Properties.IsLoadBearing,
					"material":        w.Properties.Material,
					"can_demolish":    w.Properties.CanDemolish,
					"structural_type": w.Properties.StructuralType,
				}
			}
			if w.Metadata != nil {
				wall["metadata"] = fiber.Map{
					"confidence": w.Metadata.Confidence,
					"source":     w.Metadata.Source,
					"locked":     w.Metadata.Locked,
					"visible":    w.Metadata.Visible,
				}
			}
			// Convert openings
			if len(w.Openings) > 0 {
				openings := make([]fiber.Map, 0, len(w.Openings))
				for _, o := range w.Openings {
					openings = append(openings, fiber.Map{
						"id":             o.Id,
						"type":           o.Type,
						"subtype":        o.Subtype,
						"position":       o.Position,
						"width":          o.Width,
						"height":         o.Height,
						"elevation":      o.Elevation,
						"opens_to":       o.OpensTo,
						"has_door":       o.HasDoor,
						"connects_rooms": o.ConnectsRooms,
					})
				}
				wall["openings"] = openings
			}
			walls = append(walls, wall)
		}
		elements["walls"] = walls

		// Convert rooms
		rooms := make([]fiber.Map, 0, len(model.Elements.Rooms))
		for _, r := range model.Elements.Rooms {
			room := fiber.Map{
				"id":        r.Id,
				"type":      r.Type,
				"name":      r.Name,
				"room_type": r.RoomType,
				"area":      r.Area,
				"perimeter": r.Perimeter,
				"wall_ids":  r.WallIds,
			}
			// Convert polygon
			if len(r.Polygon) > 0 {
				polygon := make([]fiber.Map, 0, len(r.Polygon))
				for _, p := range r.Polygon {
					polygon = append(polygon, fiber.Map{"x": p.X, "y": p.Y})
				}
				room["polygon"] = polygon
			}
			if r.Properties != nil {
				room["properties"] = fiber.Map{
					"has_wet_zone":     r.Properties.HasWetZone,
					"has_ventilation":  r.Properties.HasVentilation,
					"has_window":       r.Properties.HasWindow,
					"min_allowed_area": r.Properties.MinAllowedArea,
					"ceiling_height":   r.Properties.CeilingHeight,
				}
			}
			if r.RoomMetadata != nil {
				room["metadata"] = fiber.Map{
					"confidence":    r.RoomMetadata.Confidence,
					"label_on_plan": r.RoomMetadata.LabelOnPlan,
					"area_on_plan":  r.RoomMetadata.AreaOnPlan,
				}
			}
			rooms = append(rooms, room)
		}
		elements["rooms"] = rooms

		// Convert furniture
		furniture := make([]fiber.Map, 0, len(model.Elements.Furniture))
		for _, f := range model.Elements.Furniture {
			item := fiber.Map{
				"id":             f.Id,
				"type":           f.Type,
				"name":           f.Name,
				"furniture_type": f.FurnitureType,
				"room_id":        f.RoomId,
			}
			if f.Position != nil {
				item["position"] = fiber.Map{"x": f.Position.X, "y": f.Position.Y, "z": f.Position.Z}
			}
			if f.Rotation != nil {
				item["rotation"] = fiber.Map{"x": f.Rotation.X, "y": f.Rotation.Y, "z": f.Rotation.Z}
			}
			if f.Dimensions != nil {
				item["dimensions"] = fiber.Map{"width": f.Dimensions.Width, "height": f.Dimensions.Height, "depth": f.Dimensions.Depth}
			}
			if f.Properties != nil {
				item["properties"] = fiber.Map{
					"can_relocate":   f.Properties.CanRelocate,
					"category":       f.Properties.Category,
					"requires_water": f.Properties.RequiresWater,
					"requires_gas":   f.Properties.RequiresGas,
					"requires_drain": f.Properties.RequiresDrain,
				}
			}
			furniture = append(furniture, item)
		}
		elements["furniture"] = furniture

		// Convert utilities
		utilities := make([]fiber.Map, 0, len(model.Elements.Utilities))
		for _, u := range model.Elements.Utilities {
			item := fiber.Map{
				"id":           u.Id,
				"type":         u.Type,
				"name":         u.Name,
				"utility_type": u.UtilityType,
				"room_id":      u.RoomId,
			}
			if u.Position != nil {
				item["position"] = fiber.Map{"x": u.Position.X, "y": u.Position.Y, "z": u.Position.Z}
			}
			if u.Dimensions != nil {
				item["dimensions"] = fiber.Map{"diameter": u.Dimensions.Diameter, "width": u.Dimensions.Width, "depth": u.Dimensions.Depth}
			}
			if u.Properties != nil {
				item["properties"] = fiber.Map{
					"can_relocate":          u.Properties.CanRelocate,
					"protection_zone":       u.Properties.ProtectionZone,
					"shared_with_neighbors": u.Properties.SharedWithNeighbors,
				}
			}
			utilities = append(utilities, item)
		}
		elements["utilities"] = utilities

		result["elements"] = elements
	}

	return result
}

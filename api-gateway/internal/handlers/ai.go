// =============================================================================
// Package handlers provides HTTP handlers for API Gateway.
// =============================================================================
// AIHandler handles AI-related HTTP requests including:
// - Floor plan recognition
// - Design variants generation
// - Interactive chat with AI assistant
// =============================================================================
package handlers

import (
	"context"
	"encoding/base64"
	"io"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"google.golang.org/grpc"

	aipb "github.com/xiiisorate/granula_api/shared/gen/ai/v1"
	commonpb "github.com/xiiisorate/granula_api/shared/gen/common/v1"
)

// AIHandler handles AI-related HTTP requests.
type AIHandler struct {
	client aipb.AIServiceClient
}

// NewAIHandler creates a new AIHandler.
func NewAIHandler(conn *grpc.ClientConn) *AIHandler {
	return &AIHandler{
		client: aipb.NewAIServiceClient(conn),
	}
}

// =============================================================================
// RecognizeFloorPlan recognizes a floor plan from an image.
// @Summary Распознать планировку
// @Description AI распознавание планировки из изображения
// @Tags ai
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param file formData file false "Изображение планировки"
// @Param image_url body string false "URL изображения"
// @Param floor_plan_id body string true "ID планировки"
// @Param options body RecognitionOptionsInput false "Опции распознавания"
// @Success 200 {object} RecognitionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /ai/recognize [post]
// =============================================================================
func (h *AIHandler) RecognizeFloorPlan(c *fiber.Ctx) error {
	var imageData []byte
	var imageType string

	// Check for file upload
	fileHeader, err := c.FormFile("file")
	if err == nil && fileHeader != nil {
		file, err := fileHeader.Open()
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "failed to open uploaded file")
		}
		defer file.Close()

		imageData, err = io.ReadAll(file)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "failed to read uploaded file")
		}
		imageType = fileHeader.Header.Get("Content-Type")
	} else {
		// Try JSON body with image_url or base64
		var input RecognizeInput
		if err := c.BodyParser(&input); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		if input.ImageBase64 != "" {
			imageData, err = base64.StdEncoding.DecodeString(input.ImageBase64)
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, "invalid base64 image data")
			}
			imageType = input.ImageType
		} else if input.ImageURL != "" {
			// In production: fetch image from URL
			// For now, just return a placeholder response
			return c.JSON(fiber.Map{
				"data": fiber.Map{
					"job_id":  uuid.New().String(),
					"status":  "pending",
					"message": "Recognition started from URL",
				},
				"request_id": c.GetRespHeader("X-Request-ID"),
			})
		}
	}

	floorPlanID := c.FormValue("floor_plan_id")
	if floorPlanID == "" {
		// Try from body
		var input struct {
			FloorPlanID string `json:"floor_plan_id"`
		}
		c.BodyParser(&input)
		floorPlanID = input.FloorPlanID
	}

	if floorPlanID == "" {
		floorPlanID = uuid.New().String() // Auto-generate if not provided
	}

	ctx, cancel := context.WithTimeout(c.Context(), 60*time.Second)
	defer cancel()

	// Build gRPC request
	req := &aipb.RecognizeFloorPlanRequest{
		FloorPlanId: floorPlanID,
		ImageData:   imageData,
		ImageType:   imageType,
		Options: &aipb.RecognitionOptions{
			DetectLoadBearing: true,
			DetectWetZones:    true,
			DetectFurniture:   false,
			DetailLevel:       2,
		},
	}

	resp, err := h.client.RecognizeFloorPlan(ctx, req)
	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"data":       recognitionResponseToMap(resp),
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// GetRecognitionStatus gets the status of a recognition job.
// @Summary Статус распознавания
// @Description Получить статус задачи распознавания
// @Tags ai
// @Produce json
// @Security BearerAuth
// @Param job_id path string true "ID задачи"
// @Success 200 {object} RecognitionStatusResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /ai/recognize/{job_id}/status [get]
// =============================================================================
func (h *AIHandler) GetRecognitionStatus(c *fiber.Ctx) error {
	jobID := c.Params("job_id")
	if jobID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "job_id is required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.client.GetRecognitionStatus(ctx, &aipb.GetRecognitionStatusRequest{
		JobId: jobID,
	})
	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"job_id":   resp.JobId,
			"status":   jobStatusToString(resp.Status),
			"progress": resp.Progress,
			"error":    resp.Error,
		},
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// GenerateVariants generates design variants using AI.
// @Summary Сгенерировать варианты
// @Description AI генерация вариантов планировки
// @Tags ai
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body GenerateVariantsInput true "Параметры генерации"
// @Success 200 {object} GenerationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /ai/generate [post]
// =============================================================================
func (h *AIHandler) GenerateVariants(c *fiber.Ctx) error {
	var input GenerateVariantsInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// Validation
	if input.SceneID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "scene_id is required")
	}
	if input.Prompt == "" {
		return fiber.NewError(fiber.StatusBadRequest, "prompt is required")
	}

	// Default variants count
	variantsCount := input.VariantsCount
	if variantsCount < 1 || variantsCount > 5 {
		variantsCount = 3
	}

	ctx, cancel := context.WithTimeout(c.Context(), 120*time.Second)
	defer cancel()

	// Parse generation style
	style := aipb.GenerationStyle_GENERATION_STYLE_MODERATE
	switch input.Style {
	case "minimal":
		style = aipb.GenerationStyle_GENERATION_STYLE_MINIMAL
	case "creative":
		style = aipb.GenerationStyle_GENERATION_STYLE_CREATIVE
	}

	req := &aipb.GenerateVariantsRequest{
		SceneId:       input.SceneID,
		BranchId:      input.BranchID,
		Prompt:        input.Prompt,
		VariantsCount: int32(variantsCount),
		Options: &aipb.GenerationOptions{
			PreserveLoadBearing: input.PreserveLoadBearing,
			CheckCompliance:     input.CheckCompliance,
			PreserveWetZones:    input.PreserveWetZones,
			Style:               style,
			Budget:              float32(input.Budget),
		},
	}

	resp, err := h.client.GenerateVariants(ctx, req)
	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"data":       generationResponseToMap(resp),
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// GetGenerationStatus gets the status of a generation job.
// @Summary Статус генерации
// @Description Получить статус задачи генерации
// @Tags ai
// @Produce json
// @Security BearerAuth
// @Param job_id path string true "ID задачи"
// @Success 200 {object} GenerationStatusResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /ai/generate/{job_id}/status [get]
// =============================================================================
func (h *AIHandler) GetGenerationStatus(c *fiber.Ctx) error {
	jobID := c.Params("job_id")
	if jobID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "job_id is required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.client.GetGenerationStatus(ctx, &aipb.GetGenerationStatusRequest{
		JobId: jobID,
	})
	if err != nil {
		return handleGRPCError(err)
	}

	variants := make([]fiber.Map, 0, len(resp.Variants))
	for _, v := range resp.Variants {
		variants = append(variants, variantToMap(v))
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"job_id":   resp.JobId,
			"status":   jobStatusToString(resp.Status),
			"progress": resp.Progress,
			"variants": variants,
			"error":    resp.Error,
		},
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// SendChatMessage sends a message to the AI chat.
// @Summary Отправить сообщение в чат
// @Description Отправка сообщения AI-ассистенту
// @Tags ai
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ChatInput true "Сообщение"
// @Success 200 {object} ChatResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /ai/chat [post]
// =============================================================================
func (h *AIHandler) SendChatMessage(c *fiber.Ctx) error {
	var input ChatInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if input.Message == "" {
		return fiber.NewError(fiber.StatusBadRequest, "message is required")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 60*time.Second)
	defer cancel()

	req := &aipb.ChatMessageRequest{
		SceneId:   input.SceneID,
		BranchId:  input.BranchID,
		Message:   input.Message,
		ContextId: input.ContextID,
	}

	resp, err := h.client.SendChatMessage(ctx, req)
	if err != nil {
		return handleGRPCError(err)
	}

	actions := make([]fiber.Map, 0, len(resp.Actions))
	for _, a := range resp.Actions {
		actions = append(actions, actionToMap(a))
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message_id":         resp.MessageId,
			"response":           resp.Response,
			"context_id":         resp.ContextId,
			"actions":            actions,
			"generation_time_ms": resp.GenerationTimeMs,
			"token_usage": fiber.Map{
				"prompt_tokens":     resp.TokenUsage.GetPromptTokens(),
				"completion_tokens": resp.TokenUsage.GetCompletionTokens(),
				"total_tokens":      resp.TokenUsage.GetTotalTokens(),
			},
		},
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// GetChatHistory gets the chat history for a scene/branch.
// @Summary История чата
// @Description Получить историю чата с AI
// @Tags ai
// @Produce json
// @Security BearerAuth
// @Param scene_id query string false "ID сцены"
// @Param branch_id query string false "ID ветки"
// @Param limit query int false "Лимит сообщений" default(50)
// @Success 200 {object} ChatHistoryResponse
// @Failure 401 {object} ErrorResponse
// @Router /ai/chat/history [get]
// =============================================================================
func (h *AIHandler) GetChatHistory(c *fiber.Ctx) error {
	sceneID := c.Query("scene_id")
	branchID := c.Query("branch_id")
	limit := c.QueryInt("limit", 50)
	cursor := c.Query("cursor", "")

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.client.GetChatHistory(ctx, &aipb.GetChatHistoryRequest{
		SceneId:  sceneID,
		BranchId: branchID,
		Limit:    int32(limit),
		Cursor:   cursor,
	})
	if err != nil {
		return handleGRPCError(err)
	}

	messages := make([]fiber.Map, 0, len(resp.Messages))
	for _, m := range resp.Messages {
		msg := fiber.Map{
			"id":      m.Id,
			"role":    m.Role,
			"content": m.Content,
		}
		if m.CreatedAt != nil {
			msg["created_at"] = m.CreatedAt.AsTime()
		}
		if len(m.Actions) > 0 {
			actions := make([]fiber.Map, 0, len(m.Actions))
			for _, a := range m.Actions {
				actions = append(actions, actionToMap(a))
			}
			msg["actions"] = actions
		}
		messages = append(messages, msg)
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"messages":    messages,
			"has_more":    resp.HasMore,
			"next_cursor": resp.NextCursor,
		},
		"request_id": c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// ClearChatHistory clears the chat history.
// @Summary Очистить историю чата
// @Description Удалить историю чата с AI
// @Tags ai
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ClearChatInput true "Параметры очистки"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Router /ai/chat/history [delete]
// =============================================================================
func (h *AIHandler) ClearChatHistory(c *fiber.Ctx) error {
	var input ClearChatInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	resp, err := h.client.ClearChatHistory(ctx, &aipb.ClearChatHistoryRequest{
		SceneId:   input.SceneID,
		BranchId:  input.BranchID,
		ContextId: input.ContextID,
	})
	if err != nil {
		return handleGRPCError(err)
	}

	return c.JSON(fiber.Map{
		"message":       "Chat history cleared",
		"deleted_count": resp.DeletedCount,
		"request_id":    c.GetRespHeader("X-Request-ID"),
	})
}

// =============================================================================
// Helper functions
// =============================================================================

func recognitionResponseToMap(resp *aipb.RecognizeFloorPlanResponse) fiber.Map {
	result := fiber.Map{
		"success":    resp.Success,
		"job_id":     resp.JobId,
		"status":     jobStatusToString(resp.Status),
		"confidence": resp.Confidence,
		"warnings":   resp.Warnings,
		"error":      resp.Error,
	}

	if resp.Scene != nil {
		walls := make([]fiber.Map, 0, len(resp.Scene.Walls))
		for _, w := range resp.Scene.Walls {
			walls = append(walls, fiber.Map{
				"temp_id":                 w.TempId,
				"start":                   point2DToMap(w.Start),
				"end":                     point2DToMap(w.End),
				"thickness":               w.Thickness,
				"is_load_bearing":         w.IsLoadBearing,
				"confidence":              w.Confidence,
				"load_bearing_confidence": w.LoadBearingConfidence,
			})
		}

		rooms := make([]fiber.Map, 0, len(resp.Scene.Rooms))
		for _, r := range resp.Scene.Rooms {
			rooms = append(rooms, fiber.Map{
				"temp_id":     r.TempId,
				"type":        roomTypeToString(r.Type),
				"area":        r.Area,
				"is_wet_zone": r.IsWetZone,
				"confidence":  r.Confidence,
				"wall_ids":    r.WallIds,
			})
		}

		openings := make([]fiber.Map, 0, len(resp.Scene.Openings))
		for _, o := range resp.Scene.Openings {
			openings = append(openings, fiber.Map{
				"temp_id":    o.TempId,
				"type":       openingTypeToString(o.Type),
				"position":   point2DToMap(o.Position),
				"width":      o.Width,
				"wall_id":    o.WallId,
				"confidence": o.Confidence,
			})
		}

		result["scene"] = fiber.Map{
			"total_area": resp.Scene.TotalArea,
			"walls":      walls,
			"rooms":      rooms,
			"openings":   openings,
		}

		if resp.Scene.Metadata != nil {
			result["metadata"] = fiber.Map{
				"model_version":        resp.Scene.Metadata.ModelVersion,
				"processing_time_ms":   resp.Scene.Metadata.ProcessingTimeMs,
				"detected_scale":       resp.Scene.Metadata.DetectedScale,
				"detected_orientation": resp.Scene.Metadata.DetectedOrientation,
			}
		}
	}

	return result
}

func generationResponseToMap(resp *aipb.GenerateVariantsResponse) fiber.Map {
	variants := make([]fiber.Map, 0, len(resp.Variants))
	for _, v := range resp.Variants {
		variants = append(variants, variantToMap(v))
	}

	return fiber.Map{
		"success":  resp.Success,
		"job_id":   resp.JobId,
		"status":   jobStatusToString(resp.Status),
		"variants": variants,
		"error":    resp.Error,
	}
}

func variantToMap(v *aipb.GeneratedVariant) fiber.Map {
	changes := make([]fiber.Map, 0, len(v.Changes))
	for _, ch := range v.Changes {
		changes = append(changes, fiber.Map{
			"type":        ch.Type,
			"description": ch.Description,
			"element_ids": ch.ElementIds,
		})
	}

	return fiber.Map{
		"id":             v.Id,
		"branch_id":      v.BranchId,
		"name":           v.Name,
		"description":    v.Description,
		"score":          v.Score,
		"changes":        changes,
		"is_compliant":   v.IsCompliant,
		"estimated_cost": v.EstimatedCost,
	}
}

func actionToMap(a *aipb.SuggestedAction) fiber.Map {
	return fiber.Map{
		"id":                    a.Id,
		"type":                  a.Type,
		"description":           a.Description,
		"params":                a.Params,
		"confidence":            a.Confidence,
		"requires_confirmation": a.RequiresConfirmation,
	}
}

func point2DToMap(p *commonpb.Point2D) fiber.Map {
	if p == nil {
		return nil
	}
	return fiber.Map{
		"x": p.X,
		"y": p.Y,
	}
}

func jobStatusToString(s aipb.JobStatus) string {
	switch s {
	case aipb.JobStatus_JOB_STATUS_PENDING:
		return "pending"
	case aipb.JobStatus_JOB_STATUS_PROCESSING:
		return "processing"
	case aipb.JobStatus_JOB_STATUS_COMPLETED:
		return "completed"
	case aipb.JobStatus_JOB_STATUS_FAILED:
		return "failed"
	case aipb.JobStatus_JOB_STATUS_CANCELLED:
		return "cancelled"
	default:
		return "unknown"
	}
}

func roomTypeToString(t aipb.RoomType) string {
	switch t {
	case aipb.RoomType_ROOM_TYPE_LIVING:
		return "living"
	case aipb.RoomType_ROOM_TYPE_BEDROOM:
		return "bedroom"
	case aipb.RoomType_ROOM_TYPE_KITCHEN:
		return "kitchen"
	case aipb.RoomType_ROOM_TYPE_BATHROOM:
		return "bathroom"
	case aipb.RoomType_ROOM_TYPE_TOILET:
		return "toilet"
	case aipb.RoomType_ROOM_TYPE_COMBINED_BATHROOM:
		return "combined_bathroom"
	case aipb.RoomType_ROOM_TYPE_HALLWAY:
		return "hallway"
	case aipb.RoomType_ROOM_TYPE_BALCONY:
		return "balcony"
	case aipb.RoomType_ROOM_TYPE_LOGGIA:
		return "loggia"
	case aipb.RoomType_ROOM_TYPE_STORAGE:
		return "storage"
	case aipb.RoomType_ROOM_TYPE_LAUNDRY:
		return "laundry"
	case aipb.RoomType_ROOM_TYPE_OFFICE:
		return "office"
	case aipb.RoomType_ROOM_TYPE_CHILDREN:
		return "children"
	case aipb.RoomType_ROOM_TYPE_DINING:
		return "dining"
	case aipb.RoomType_ROOM_TYPE_KITCHEN_LIVING:
		return "kitchen_living"
	default:
		return "unknown"
	}
}

func openingTypeToString(t aipb.OpeningType) string {
	switch t {
	case aipb.OpeningType_OPENING_TYPE_DOOR:
		return "door"
	case aipb.OpeningType_OPENING_TYPE_WINDOW:
		return "window"
	case aipb.OpeningType_OPENING_TYPE_ARCH:
		return "arch"
	case aipb.OpeningType_OPENING_TYPE_PASSAGE:
		return "passage"
	default:
		return "unknown"
	}
}

// =============================================================================
// Input types
// =============================================================================

// RecognizeInput - input for floor plan recognition.
type RecognizeInput struct {
	FloorPlanID string `json:"floor_plan_id,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
	ImageBase64 string `json:"image_base64,omitempty"`
	ImageType   string `json:"image_type,omitempty"`
}

// GenerateVariantsInput - input for variant generation.
type GenerateVariantsInput struct {
	SceneID             string  `json:"scene_id"`
	BranchID            string  `json:"branch_id,omitempty"`
	Prompt              string  `json:"prompt"`
	VariantsCount       int     `json:"variants_count,omitempty"`
	Style               string  `json:"style,omitempty"` // minimal, moderate, creative
	PreserveLoadBearing bool    `json:"preserve_load_bearing,omitempty"`
	CheckCompliance     bool    `json:"check_compliance,omitempty"`
	PreserveWetZones    bool    `json:"preserve_wet_zones,omitempty"`
	Budget              float64 `json:"budget,omitempty"`
}

// ChatInput - input for chat message.
type ChatInput struct {
	SceneID   string `json:"scene_id,omitempty"`
	BranchID  string `json:"branch_id,omitempty"`
	Message   string `json:"message"`
	ContextID string `json:"context_id,omitempty"`
}

// ClearChatInput - input for clearing chat history.
type ClearChatInput struct {
	SceneID   string `json:"scene_id,omitempty"`
	BranchID  string `json:"branch_id,omitempty"`
	ContextID string `json:"context_id,omitempty"`
}


// Package service provides business logic for AI Service.
package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/ai-service/internal/domain/entity"
	"github.com/xiiisorate/granula_api/ai-service/internal/openrouter"
	"github.com/xiiisorate/granula_api/ai-service/internal/prompts"
	"github.com/xiiisorate/granula_api/ai-service/internal/repository/mongodb"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
)

// RecognitionService handles floor plan recognition.
type RecognitionService struct {
	jobRepo *mongodb.JobRepository
	client  *openrouter.Client
	log     *logger.Logger
}

// NewRecognitionService creates a new RecognitionService.
func NewRecognitionService(jobRepo *mongodb.JobRepository, client *openrouter.Client, log *logger.Logger) *RecognitionService {
	return &RecognitionService{
		jobRepo: jobRepo,
		client:  client,
		log:     log,
	}
}

// StartRecognition starts a floor plan recognition job.
func (s *RecognitionService) StartRecognition(ctx context.Context, floorPlanID string, imageData []byte, imageType string, options entity.RecognitionOptions) (*entity.RecognitionJob, error) {
	s.log.Info("starting recognition",
		logger.String("floor_plan_id", floorPlanID),
		logger.Int("image_size", len(imageData)),
	)

	// Create job
	job := entity.NewRecognitionJob(floorPlanID, options)
	if err := s.jobRepo.SaveRecognitionJob(ctx, job); err != nil {
		return nil, err
	}

	// Process in background
	go s.processRecognition(context.Background(), job, imageData, imageType)

	return job, nil
}

// GetRecognitionStatus retrieves a recognition job status.
func (s *RecognitionService) GetRecognitionStatus(ctx context.Context, jobID uuid.UUID) (*entity.RecognitionJob, error) {
	return s.jobRepo.GetRecognitionJob(ctx, jobID)
}

// processRecognition performs the actual recognition using Vision API.
// This method sends the REAL image to a vision-capable AI model for analysis.
func (s *RecognitionService) processRecognition(ctx context.Context, job *entity.RecognitionJob, imageData []byte, imageType string) {
	startTime := time.Now()

	// Mark as processing
	job.Start()
	_ = s.jobRepo.UpdateRecognitionJob(ctx, job)

	// Encode image to base64 data URL for Vision API
	// This is the FULL image data, not truncated!
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	dataURL := fmt.Sprintf("data:%s;base64,%s", imageType, base64Image)

	s.log.Info("processing floor plan image",
		logger.String("job_id", job.ID.String()),
		logger.Int("image_size_bytes", len(imageData)),
		logger.String("image_type", imageType),
	)

	// Update progress: image prepared
	job.UpdateProgress(10)
	_ = s.jobRepo.UpdateRecognitionJob(ctx, job)

	// Build user prompt based on recognition options
	prompt := "Проанализируй эту планировку квартиры и извлеки структурированные данные. "
	if job.Options.DetectLoadBearing {
		prompt += "Определи несущие стены (по толщине линий и расположению). "
	}
	if job.Options.DetectWetZones {
		prompt += "Определи мокрые зоны (ванная, туалет, кухня). "
	}
	if job.Options.DetectFurniture {
		prompt += "Определи мебель и оборудование. "
	}
	prompt += "Верни результат ТОЛЬКО в формате JSON без markdown-обёртки."

	// Build multimodal message with REAL image data
	// This sends the actual image to the Vision API, not just a placeholder!
	messages := []openrouter.MultimodalMessage{
		{
			Role: "user",
			Content: []openrouter.ContentPart{
				{
					Type: "text",
					Text: prompt,
				},
				{
					Type: "image_url",
					ImageURL: &openrouter.ImageURL{
						URL:    dataURL, // Full base64 image data!
						Detail: "high",  // High quality for accurate recognition
					},
				},
			},
		},
	}

	// Update progress: sending to AI
	job.UpdateProgress(30)
	_ = s.jobRepo.UpdateRecognitionJob(ctx, job)

	// Call OpenRouter Vision API with detailed recognition prompt
	// Using Claude 3.5 Sonnet which has excellent vision capabilities
	resp, err := s.client.ChatCompletionWithImages(ctx, messages, openrouter.ChatOptions{
		SystemPrompt: prompts.GetRecognitionPrompt(),
		MaxTokens:    8192,                        // Large response for detailed JSON
		Temperature:  0.2,                         // Low temperature for consistent output
		Model:        "anthropic/claude-sonnet-4", // Claude Sonnet 4 with Vision
	})
	if err != nil {
		s.log.Error("recognition failed - Vision API error",
			logger.Err(err),
			logger.String("job_id", job.ID.String()),
		)
		job.Fail(err.Error())
		_ = s.jobRepo.UpdateRecognitionJob(ctx, job)
		return
	}

	// Update progress: processing AI response
	job.UpdateProgress(70)
	_ = s.jobRepo.UpdateRecognitionJob(ctx, job)

	if len(resp.Choices) == 0 {
		s.log.Error("recognition failed - no response from AI",
			logger.String("job_id", job.ID.String()),
		)
		job.Fail("no response from AI")
		_ = s.jobRepo.UpdateRecognitionJob(ctx, job)
		return
	}

	// Parse recognition result from AI response
	content := resp.Choices[0].Message.Content
	result, err := s.parseRecognitionResult(content)
	if err != nil {
		s.log.Warn("failed to parse recognition result",
			logger.Err(err),
			logger.String("job_id", job.ID.String()),
			logger.String("content_preview", truncateString(content, 500)),
		)
		// Create a minimal result with warning
		result = &entity.RecognitionResult{
			Confidence:   0.5,
			Warnings:     []string{"Не удалось полностью распознать планировку. Проверьте качество изображения."},
			ModelVersion: "1.0.0",
		}
	}

	result.ProcessingTimeMs = time.Since(startTime).Milliseconds()

	// Update progress: finalizing
	job.UpdateProgress(90)
	_ = s.jobRepo.UpdateRecognitionJob(ctx, job)

	// Complete job with result
	job.Complete(result)
	_ = s.jobRepo.UpdateRecognitionJob(ctx, job)

	s.log.Info("recognition completed successfully",
		logger.String("job_id", job.ID.String()),
		logger.Int64("processing_time_ms", result.ProcessingTimeMs),
		logger.F("confidence", result.Confidence),
		logger.Int("walls_count", len(result.Walls)),
		logger.Int("rooms_count", len(result.Rooms)),
	)
}

// truncateString truncates a string to maxLen characters with ellipsis.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// parseRecognitionResult parses the AI response into structured result.
// Supports both new format (elements.walls, bounds) and legacy format (walls, dimensions).
func (s *RecognitionService) parseRecognitionResult(content string) (*entity.RecognitionResult, error) {
	// Find JSON in response
	start := -1
	end := -1
	braceCount := 0

	for i, c := range content {
		if c == '{' {
			if braceCount == 0 {
				start = i
			}
			braceCount++
		} else if c == '}' {
			braceCount--
			if braceCount == 0 && start != -1 {
				end = i + 1
				break
			}
		}
	}

	if start == -1 || end == -1 {
		return nil, fmt.Errorf("no JSON found in response")
	}

	jsonStr := content[start:end]

	// Log the raw JSON for debugging
	s.log.Debug("parsing recognition JSON",
		logger.String("json_preview", truncateString(jsonStr, 1000)),
		logger.Int("json_length", len(jsonStr)),
	)

	// First, try to parse into a generic map to detect format
	var rawData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON as map: %w", err)
	}

	// Check if this is the new format (with elements.walls) or legacy format
	result := &entity.RecognitionResult{
		ModelVersion: "1.0.0",
	}

	// Parse bounds (new format) or dimensions (legacy)
	if bounds, ok := rawData["bounds"].(map[string]interface{}); ok {
		result.Bounds = entity.Bounds3D{
			Width:  getFloat(bounds, "width"),
			Height: getFloat(bounds, "height"),
			Depth:  getFloat(bounds, "depth"),
		}
		result.Dimensions = entity.Dimensions2D{
			Width:  result.Bounds.Width,
			Height: result.Bounds.Depth, // For 2D, height = depth
		}
		result.TotalArea = result.Bounds.Width * result.Bounds.Depth
	} else if dims, ok := rawData["dimensions"].(map[string]interface{}); ok {
		result.Dimensions = entity.Dimensions2D{
			Width:  getFloat(dims, "width"),
			Height: getFloat(dims, "height"),
		}
		result.Bounds = entity.Bounds3D{
			Width:  result.Dimensions.Width,
			Height: 2.7, // Default ceiling height
			Depth:  result.Dimensions.Height,
		}
	}

	// Parse stats
	if stats, ok := rawData["stats"].(map[string]interface{}); ok {
		result.Stats = entity.RecognitionStats{
			TotalArea:      getFloat(stats, "totalArea"),
			RoomsCount:     getInt(stats, "roomsCount"),
			WallsCount:     getInt(stats, "wallsCount"),
			FurnitureCount: getInt(stats, "furnitureCount"),
		}
		if result.TotalArea == 0 {
			result.TotalArea = result.Stats.TotalArea
		}
	}

	// Parse elements (new format)
	if elements, ok := rawData["elements"].(map[string]interface{}); ok {
		result.Elements = parseSceneElements(elements)

		// Also populate legacy fields from elements
		result.Walls = convertWalls3DToLegacy(result.Elements.Walls)
		result.Rooms = convertRooms3DToLegacy(result.Elements.Rooms)
		result.Openings = extractOpeningsFromWalls(result.Elements.Walls)
	} else {
		// Legacy format: walls, rooms, openings at top level
		if wallsRaw, ok := rawData["walls"].([]interface{}); ok {
			result.Walls = parseLegacyWalls(wallsRaw)
		}
		if roomsRaw, ok := rawData["rooms"].([]interface{}); ok {
			result.Rooms = parseLegacyRooms(roomsRaw)
		}
		if openingsRaw, ok := rawData["openings"].([]interface{}); ok {
			result.Openings = parseLegacyOpenings(openingsRaw)
		}
	}

	// Parse recognition metadata
	if recognition, ok := rawData["recognition"].(map[string]interface{}); ok {
		result.Recognition = entity.RecognitionMeta{
			SourceType:     getString(recognition, "sourceType"),
			Quality:        getString(recognition, "quality"),
			Scale:          getString(recognition, "scale"),
			Orientation:    getInt(recognition, "orientation"),
			HasDimensions:  getBool(recognition, "hasDimensions"),
			HasAnnotations: getBool(recognition, "hasAnnotations"),
			BuildingType:   getString(recognition, "buildingType"),
		}
		if warnings, ok := recognition["warnings"].([]interface{}); ok {
			for _, w := range warnings {
				if ws, ok := w.(string); ok {
					result.Warnings = append(result.Warnings, ws)
				}
			}
		}
	}

	// Parse warnings at top level
	if warnings, ok := rawData["warnings"].([]interface{}); ok {
		for _, w := range warnings {
			if ws, ok := w.(string); ok {
				result.Warnings = append(result.Warnings, ws)
			}
		}
	}

	result.Confidence = calculateOverallConfidence(result)

	s.log.Info("recognition result parsed",
		logger.Int("walls", len(result.Walls)),
		logger.Int("rooms", len(result.Rooms)),
		logger.Int("openings", len(result.Openings)),
		logger.Int("elements_walls", len(result.Elements.Walls)),
		logger.Int("elements_rooms", len(result.Elements.Rooms)),
		logger.F("confidence", result.Confidence),
	)

	return result, nil
}

// =============================================================================
// JSON PARSING HELPERS
// =============================================================================

func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

// parseSceneElements parses the elements object from new format.
func parseSceneElements(elements map[string]interface{}) entity.SceneElements {
	result := entity.SceneElements{}

	if wallsRaw, ok := elements["walls"].([]interface{}); ok {
		for _, w := range wallsRaw {
			if wm, ok := w.(map[string]interface{}); ok {
				wall := parseWall3D(wm)
				result.Walls = append(result.Walls, wall)
			}
		}
	}

	if roomsRaw, ok := elements["rooms"].([]interface{}); ok {
		for _, r := range roomsRaw {
			if rm, ok := r.(map[string]interface{}); ok {
				room := parseRoom3D(rm)
				result.Rooms = append(result.Rooms, room)
			}
		}
	}

	if furnitureRaw, ok := elements["furniture"].([]interface{}); ok {
		for _, f := range furnitureRaw {
			if fm, ok := f.(map[string]interface{}); ok {
				furniture := parseFurniture(fm)
				result.Furniture = append(result.Furniture, furniture)
			}
		}
	}

	if utilitiesRaw, ok := elements["utilities"].([]interface{}); ok {
		for _, u := range utilitiesRaw {
			if um, ok := u.(map[string]interface{}); ok {
				utility := parseUtility(um)
				result.Utilities = append(result.Utilities, utility)
			}
		}
	}

	return result
}

func parseWall3D(m map[string]interface{}) entity.Wall3D {
	wall := entity.Wall3D{
		ID:        getString(m, "id"),
		Type:      getString(m, "type"),
		Name:      getString(m, "name"),
		Height:    getFloat(m, "height"),
		Thickness: getFloat(m, "thickness"),
	}

	if start, ok := m["start"].(map[string]interface{}); ok {
		wall.Start = entity.Point3D{
			X: getFloat(start, "x"),
			Y: getFloat(start, "y"),
			Z: getFloat(start, "z"),
		}
	}

	if end, ok := m["end"].(map[string]interface{}); ok {
		wall.End = entity.Point3D{
			X: getFloat(end, "x"),
			Y: getFloat(end, "y"),
			Z: getFloat(end, "z"),
		}
	}

	if props, ok := m["properties"].(map[string]interface{}); ok {
		wall.Properties = entity.WallProperties{
			IsLoadBearing:  getBool(props, "isLoadBearing"),
			Material:       getString(props, "material"),
			CanDemolish:    getBool(props, "canDemolish"),
			StructuralType: getString(props, "structuralType"),
		}
	}

	if openingsRaw, ok := m["openings"].([]interface{}); ok {
		for _, o := range openingsRaw {
			if om, ok := o.(map[string]interface{}); ok {
				opening := parseOpening3D(om)
				wall.Openings = append(wall.Openings, opening)
			}
		}
	}

	if meta, ok := m["metadata"].(map[string]interface{}); ok {
		wall.Metadata = entity.ElementMetadata{
			Confidence: getFloat(meta, "confidence"),
			Source:     getString(meta, "source"),
			Locked:     getBool(meta, "locked"),
			Visible:    getBool(meta, "visible"),
		}
	}

	return wall
}

func parseOpening3D(m map[string]interface{}) entity.Opening3D {
	opening := entity.Opening3D{
		ID:        getString(m, "id"),
		Type:      getString(m, "type"),
		Subtype:   getString(m, "subtype"),
		Position:  getFloat(m, "position"),
		Width:     getFloat(m, "width"),
		Height:    getFloat(m, "height"),
		Elevation: getFloat(m, "elevation"),
		OpensTo:   getString(m, "opens_to"),
		HasDoor:   getBool(m, "has_door"),
	}

	if connects, ok := m["connects_rooms"].([]interface{}); ok {
		for _, c := range connects {
			if cs, ok := c.(string); ok {
				opening.ConnectsRooms = append(opening.ConnectsRooms, cs)
			}
		}
	}

	return opening
}

func parseRoom3D(m map[string]interface{}) entity.Room3D {
	room := entity.Room3D{
		ID:        getString(m, "id"),
		Type:      getString(m, "type"),
		Name:      getString(m, "name"),
		RoomType:  getString(m, "roomType"),
		Area:      getFloat(m, "area"),
		Perimeter: getFloat(m, "perimeter"),
	}

	if polygon, ok := m["polygon"].([]interface{}); ok {
		for _, p := range polygon {
			if pm, ok := p.(map[string]interface{}); ok {
				point := entity.Point2D{
					X: getFloat(pm, "x"),
					Y: getFloat(pm, "z"), // Note: z becomes y in 2D
				}
				room.Polygon = append(room.Polygon, point)
			}
		}
	}

	if props, ok := m["properties"].(map[string]interface{}); ok {
		room.Properties = entity.RoomProperties{
			HasWetZone:     getBool(props, "hasWetZone"),
			HasVentilation: getBool(props, "hasVentilation"),
			HasWindow:      getBool(props, "hasWindow"),
			MinAllowedArea: getFloat(props, "minAllowedArea"),
			CeilingHeight:  getFloat(props, "ceilingHeight"),
		}
	}

	if wallIds, ok := m["wallIds"].([]interface{}); ok {
		for _, wid := range wallIds {
			if ws, ok := wid.(string); ok {
				room.WallIDs = append(room.WallIDs, ws)
			}
		}
	}

	if meta, ok := m["metadata"].(map[string]interface{}); ok {
		room.Metadata = entity.RoomMetadata{
			Confidence:  getFloat(meta, "confidence"),
			LabelOnPlan: getString(meta, "labelOnPlan"),
			AreaOnPlan:  getFloat(meta, "areaOnPlan"),
		}
	}

	return room
}

func parseFurniture(m map[string]interface{}) entity.Furniture {
	f := entity.Furniture{
		ID:            getString(m, "id"),
		Type:          getString(m, "type"),
		Name:          getString(m, "name"),
		FurnitureType: getString(m, "furnitureType"),
		RoomID:        getString(m, "roomId"),
	}

	if pos, ok := m["position"].(map[string]interface{}); ok {
		f.Position = entity.Point3D{
			X: getFloat(pos, "x"),
			Y: getFloat(pos, "y"),
			Z: getFloat(pos, "z"),
		}
	}

	if rot, ok := m["rotation"].(map[string]interface{}); ok {
		f.Rotation = entity.Rotation3D{
			X: getFloat(rot, "x"),
			Y: getFloat(rot, "y"),
			Z: getFloat(rot, "z"),
		}
	}

	if dims, ok := m["dimensions"].(map[string]interface{}); ok {
		f.Dimensions = entity.Dimensions3D{
			Width:  getFloat(dims, "width"),
			Height: getFloat(dims, "height"),
			Depth:  getFloat(dims, "depth"),
		}
	}

	if props, ok := m["properties"].(map[string]interface{}); ok {
		f.Properties = entity.FurnitureProps{
			CanRelocate:   getBool(props, "canRelocate"),
			Category:      getString(props, "category"),
			RequiresWater: getBool(props, "requiresWater"),
			RequiresGas:   getBool(props, "requiresGas"),
			RequiresDrain: getBool(props, "requiresDrain"),
		}
	}

	return f
}

func parseUtility(m map[string]interface{}) entity.Utility {
	u := entity.Utility{
		ID:          getString(m, "id"),
		Type:        getString(m, "type"),
		Name:        getString(m, "name"),
		UtilityType: getString(m, "utilityType"),
		RoomID:      getString(m, "roomId"),
	}

	if pos, ok := m["position"].(map[string]interface{}); ok {
		u.Position = entity.Point3D{
			X: getFloat(pos, "x"),
			Y: getFloat(pos, "y"),
			Z: getFloat(pos, "z"),
		}
	}

	if dims, ok := m["dimensions"].(map[string]interface{}); ok {
		u.Dimensions = entity.UtilityDims{
			Diameter: getFloat(dims, "diameter"),
			Width:    getFloat(dims, "width"),
			Depth:    getFloat(dims, "depth"),
		}
	}

	if props, ok := m["properties"].(map[string]interface{}); ok {
		u.Properties = entity.UtilityProps{
			CanRelocate:         getBool(props, "canRelocate"),
			ProtectionZone:      getFloat(props, "protectionZone"),
			SharedWithNeighbors: getBool(props, "sharedWithNeighbors"),
		}
	}

	return u
}

// =============================================================================
// LEGACY FORMAT PARSERS
// =============================================================================

func parseLegacyWalls(wallsRaw []interface{}) []entity.RecognizedWall {
	var walls []entity.RecognizedWall
	for _, w := range wallsRaw {
		wm, ok := w.(map[string]interface{})
		if !ok {
			continue
		}
		wall := entity.RecognizedWall{
			TempID:        getString(wm, "temp_id"),
			Thickness:     getFloat(wm, "thickness"),
			IsLoadBearing: getBool(wm, "is_load_bearing"),
			Material:      getString(wm, "material"),
			CanDemolish:   getBool(wm, "can_demolish"),
			Confidence:    getFloat(wm, "confidence"),
		}
		if start, ok := wm["start"].(map[string]interface{}); ok {
			wall.Start = entity.Point2D{X: getFloat(start, "x"), Y: getFloat(start, "y")}
		}
		if end, ok := wm["end"].(map[string]interface{}); ok {
			wall.End = entity.Point2D{X: getFloat(end, "x"), Y: getFloat(end, "y")}
		}
		walls = append(walls, wall)
	}
	return walls
}

func parseLegacyRooms(roomsRaw []interface{}) []entity.RecognizedRoom {
	var rooms []entity.RecognizedRoom
	for _, r := range roomsRaw {
		rm, ok := r.(map[string]interface{})
		if !ok {
			continue
		}
		room := entity.RecognizedRoom{
			TempID:     getString(rm, "temp_id"),
			Type:       getString(rm, "type"),
			Name:       getString(rm, "name"),
			Area:       getFloat(rm, "area"),
			Perimeter:  getFloat(rm, "perimeter"),
			IsWetZone:  getBool(rm, "is_wet_zone"),
			HasWindow:  getBool(rm, "has_window"),
			Confidence: getFloat(rm, "confidence"),
		}
		if boundary, ok := rm["boundary"].([]interface{}); ok {
			for _, p := range boundary {
				if pm, ok := p.(map[string]interface{}); ok {
					room.Boundary = append(room.Boundary, entity.Point2D{
						X: getFloat(pm, "x"),
						Y: getFloat(pm, "y"),
					})
				}
			}
		}
		if polygon, ok := rm["polygon"].([]interface{}); ok {
			for _, p := range polygon {
				if pm, ok := p.(map[string]interface{}); ok {
					room.Polygon = append(room.Polygon, entity.Point2D{
						X: getFloat(pm, "x"),
						Y: getFloat(pm, "y"),
					})
				}
			}
		}
		rooms = append(rooms, room)
	}
	return rooms
}

func parseLegacyOpenings(openingsRaw []interface{}) []entity.RecognizedOpening {
	var openings []entity.RecognizedOpening
	for _, o := range openingsRaw {
		om, ok := o.(map[string]interface{})
		if !ok {
			continue
		}
		opening := entity.RecognizedOpening{
			TempID:     getString(om, "temp_id"),
			Type:       getString(om, "type"),
			Subtype:    getString(om, "subtype"),
			Width:      getFloat(om, "width"),
			Height:     getFloat(om, "height"),
			Elevation:  getFloat(om, "elevation"),
			WallID:     getString(om, "wall_id"),
			OpensTo:    getString(om, "opens_to"),
			Confidence: getFloat(om, "confidence"),
		}
		if pos, ok := om["position"].(map[string]interface{}); ok {
			opening.Position = entity.Point2D{X: getFloat(pos, "x"), Y: getFloat(pos, "y")}
		}
		openings = append(openings, opening)
	}
	return openings
}

// =============================================================================
// CONVERTERS: New format -> Legacy format
// =============================================================================

func convertWalls3DToLegacy(walls []entity.Wall3D) []entity.RecognizedWall {
	var result []entity.RecognizedWall
	for _, w := range walls {
		result = append(result, entity.RecognizedWall{
			TempID:                w.ID,
			Start:                 entity.Point2D{X: w.Start.X, Y: w.Start.Z},
			End:                   entity.Point2D{X: w.End.X, Y: w.End.Z},
			Thickness:             w.Thickness,
			IsLoadBearing:         w.Properties.IsLoadBearing,
			Material:              w.Properties.Material,
			CanDemolish:           w.Properties.CanDemolish,
			Confidence:            w.Metadata.Confidence,
			LoadBearingConfidence: w.Metadata.Confidence,
		})
	}
	return result
}

func convertRooms3DToLegacy(rooms []entity.Room3D) []entity.RecognizedRoom {
	var result []entity.RecognizedRoom
	for _, r := range rooms {
		room := entity.RecognizedRoom{
			TempID:     r.ID,
			Type:       r.RoomType,
			Name:       r.Name,
			Boundary:   r.Polygon,
			Polygon:    r.Polygon,
			Area:       r.Area,
			Perimeter:  r.Perimeter,
			IsWetZone:  r.Properties.HasWetZone,
			HasWindow:  r.Properties.HasWindow,
			Confidence: r.Metadata.Confidence,
			WallIDs:    r.WallIDs,
		}
		result = append(result, room)
	}
	return result
}

func extractOpeningsFromWalls(walls []entity.Wall3D) []entity.RecognizedOpening {
	var result []entity.RecognizedOpening
	for _, w := range walls {
		for _, o := range w.Openings {
			result = append(result, entity.RecognizedOpening{
				TempID:        o.ID,
				Type:          o.Type,
				Subtype:       o.Subtype,
				Position:      entity.Point2D{X: o.Position, Y: 0},
				Width:         o.Width,
				Height:        o.Height,
				Elevation:     o.Elevation,
				WallID:        w.ID,
				OpensTo:       o.OpensTo,
				ConnectsRooms: o.ConnectsRooms,
				Confidence:    w.Metadata.Confidence,
			})
		}
	}
	return result
}

// calculateOverallConfidence calculates average confidence.
func calculateOverallConfidence(result *entity.RecognitionResult) float64 {
	var total float64
	var count int

	for _, w := range result.Walls {
		total += w.Confidence
		count++
	}
	for _, r := range result.Rooms {
		total += r.Confidence
		count++
	}
	for _, o := range result.Openings {
		total += o.Confidence
		count++
	}

	if count == 0 {
		return 0.5
	}

	return total / float64(count)
}

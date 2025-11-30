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
		MaxTokens:    8192,                          // Large response for detailed JSON
		Temperature:  0.2,                           // Low temperature for consistent output
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

	var result entity.RecognitionResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	result.ModelVersion = "1.0.0"
	result.Confidence = calculateOverallConfidence(&result)

	return &result, nil
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

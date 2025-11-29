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

// processRecognition performs the actual recognition.
func (s *RecognitionService) processRecognition(ctx context.Context, job *entity.RecognitionJob, imageData []byte, imageType string) {
	startTime := time.Now()

	// Mark as processing
	job.Start()
	_ = s.jobRepo.UpdateRecognitionJob(ctx, job)

	// Encode image to base64
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	dataURL := fmt.Sprintf("data:%s;base64,%s", imageType, base64Image)

	// Update progress
	job.UpdateProgress(10)
	_ = s.jobRepo.UpdateRecognitionJob(ctx, job)

	// Build prompt with image
	prompt := "Проанализируй эту планировку квартиры и извлеки структурированные данные. "
	if job.Options.DetectLoadBearing {
		prompt += "Определи несущие стены. "
	}
	if job.Options.DetectWetZones {
		prompt += "Определи мокрые зоны. "
	}
	if job.Options.DetectFurniture {
		prompt += "Определи мебель и оборудование. "
	}
	prompt += "Верни результат в JSON."

	// For now, we'll use a text description since Claude doesn't support images via this API directly
	// In production, you would use a vision model or separate image analysis service
	messages := []openrouter.Message{
		{
			Role:    "user",
			Content: prompt + "\n\n[Изображение планировки загружено: " + dataURL[:100] + "...]",
		},
	}

	job.UpdateProgress(30)
	_ = s.jobRepo.UpdateRecognitionJob(ctx, job)

	// Call OpenRouter with detailed recognition prompt
	resp, err := s.client.ChatCompletionWithOptions(ctx, messages, openrouter.ChatOptions{
		SystemPrompt: prompts.GetRecognitionPrompt(),
		MaxTokens:    8192, // Increased for detailed JSON response
		Temperature:  0.2,  // Lower temperature for more consistent output
	})
	if err != nil {
		s.log.Error("recognition failed", logger.Err(err))
		job.Fail(err.Error())
		_ = s.jobRepo.UpdateRecognitionJob(ctx, job)
		return
	}

	job.UpdateProgress(70)
	_ = s.jobRepo.UpdateRecognitionJob(ctx, job)

	if len(resp.Choices) == 0 {
		job.Fail("no response from AI")
		_ = s.jobRepo.UpdateRecognitionJob(ctx, job)
		return
	}

	// Parse result
	content := resp.Choices[0].Message.Content
	result, err := s.parseRecognitionResult(content)
	if err != nil {
		s.log.Warn("failed to parse recognition result", logger.Err(err), logger.String("content", content))
		// Create a minimal result
		result = &entity.RecognitionResult{
			Confidence:   0.5,
			Warnings:     []string{"Не удалось полностью распознать планировку"},
			ModelVersion: "1.0.0",
		}
	}

	result.ProcessingTimeMs = time.Since(startTime).Milliseconds()

	job.UpdateProgress(90)
	_ = s.jobRepo.UpdateRecognitionJob(ctx, job)

	// Complete job
	job.Complete(result)
	_ = s.jobRepo.UpdateRecognitionJob(ctx, job)

	s.log.Info("recognition completed",
		logger.String("job_id", job.ID.String()),
		logger.Int64("processing_time_ms", result.ProcessingTimeMs),
	)
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


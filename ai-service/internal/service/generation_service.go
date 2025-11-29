// Package service provides business logic for AI Service.
package service

import (
	"context"
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

// GenerationService handles layout generation.
type GenerationService struct {
	jobRepo *mongodb.JobRepository
	client  *openrouter.Client
	log     *logger.Logger
}

// NewGenerationService creates a new GenerationService.
func NewGenerationService(jobRepo *mongodb.JobRepository, client *openrouter.Client, log *logger.Logger) *GenerationService {
	return &GenerationService{
		jobRepo: jobRepo,
		client:  client,
		log:     log,
	}
}

// StartGeneration starts a layout generation job.
func (s *GenerationService) StartGeneration(ctx context.Context, req GenerateRequest) (*entity.GenerationJob, error) {
	s.log.Info("starting generation",
		logger.String("scene_id", req.SceneID),
		logger.String("branch_id", req.BranchID),
		logger.Int("variants_count", req.VariantsCount),
	)

	// Create job
	job := entity.NewGenerationJob(req.SceneID, req.BranchID, req.Prompt, req.VariantsCount, req.Options)
	if err := s.jobRepo.SaveGenerationJob(ctx, job); err != nil {
		return nil, err
	}

	// Process in background
	go s.processGeneration(context.Background(), job, req.SceneData)

	return job, nil
}

// GetGenerationStatus retrieves a generation job status.
func (s *GenerationService) GetGenerationStatus(ctx context.Context, jobID uuid.UUID) (*entity.GenerationJob, error) {
	return s.jobRepo.GetGenerationJob(ctx, jobID)
}

// GenerateRequest for starting generation.
type GenerateRequest struct {
	SceneID       string
	BranchID      string
	Prompt        string
	VariantsCount int
	Options       entity.GenerationOptions
	SceneData     string // JSON representation of the scene
}

// processGeneration performs the actual generation.
func (s *GenerationService) processGeneration(ctx context.Context, job *entity.GenerationJob, sceneData string) {
	startTime := time.Now()

	// Mark as processing
	job.Start()
	_ = s.jobRepo.UpdateGenerationJob(ctx, job)

	// Build prompt
	prompt := fmt.Sprintf(`Запрос пользователя: %s

Количество вариантов: %d
Стиль: %s

Ограничения:
`, job.Prompt, job.VariantsCount, job.Options.Style)

	if job.Options.PreserveLoadBearing {
		prompt += "- Не изменять несущие стены\n"
	}
	if job.Options.PreserveWetZones {
		prompt += "- Не переносить мокрые зоны\n"
	}
	if job.Options.CheckCompliance {
		prompt += "- Проверять соответствие нормам\n"
	}
	if len(job.Options.RequiredRooms) > 0 {
		prompt += fmt.Sprintf("- Обязательные комнаты: %v\n", job.Options.RequiredRooms)
	}
	if job.Options.Budget > 0 {
		prompt += fmt.Sprintf("- Бюджет: %.0f руб.\n", job.Options.Budget)
	}

	prompt += "\nСгенерируй варианты перепланировки."

	job.UpdateProgress(20)
	_ = s.jobRepo.UpdateGenerationJob(ctx, job)

	// Call OpenRouter with detailed generation prompt
	messages := []openrouter.Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	systemPrompt := prompts.GetGenerationPrompt(sceneData)
	resp, err := s.client.ChatCompletionWithOptions(ctx, messages, openrouter.ChatOptions{
		SystemPrompt: systemPrompt,
		MaxTokens:    8192, // Increased for detailed JSON response with multiple variants
		Temperature:  0.7,  // Balanced temperature for creativity with consistency
	})
	if err != nil {
		s.log.Error("generation failed", logger.Err(err))
		job.Fail(err.Error())
		_ = s.jobRepo.UpdateGenerationJob(ctx, job)
		return
	}

	job.UpdateProgress(70)
	_ = s.jobRepo.UpdateGenerationJob(ctx, job)

	if len(resp.Choices) == 0 {
		job.Fail("no response from AI")
		_ = s.jobRepo.UpdateGenerationJob(ctx, job)
		return
	}

	// Parse result
	content := resp.Choices[0].Message.Content
	variants, err := s.parseGenerationResult(content, job.SceneID)
	if err != nil {
		s.log.Warn("failed to parse generation result", logger.Err(err))
		job.Fail("failed to parse AI response")
		_ = s.jobRepo.UpdateGenerationJob(ctx, job)
		return
	}

	job.UpdateProgress(90)
	_ = s.jobRepo.UpdateGenerationJob(ctx, job)

	// Complete job
	job.Complete(variants)
	_ = s.jobRepo.UpdateGenerationJob(ctx, job)

	s.log.Info("generation completed",
		logger.String("job_id", job.ID.String()),
		logger.Int("variants_count", len(variants)),
		logger.Int64("processing_time_ms", time.Since(startTime).Milliseconds()),
	)
}

// parseGenerationResult parses the AI response into variants.
func (s *GenerationService) parseGenerationResult(content string, sceneID string) ([]entity.GeneratedVariant, error) {
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

	var result struct {
		Variants []entity.GeneratedVariant `json:"variants"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Generate branch IDs for each variant
	for i := range result.Variants {
		if result.Variants[i].ID == "" {
			result.Variants[i].ID = uuid.New().String()
		}
		// Branch would be created by Branch Service
		result.Variants[i].BranchID = fmt.Sprintf("%s-variant-%d", sceneID, i+1)
	}

	return result.Variants, nil
}

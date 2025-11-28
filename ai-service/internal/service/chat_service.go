// Package service provides business logic for AI Service.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/ai-service/internal/domain/entity"
	"github.com/xiiisorate/granula_api/ai-service/internal/openrouter"
	"github.com/xiiisorate/granula_api/ai-service/internal/repository/mongodb"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
)

const (
	// ChatSystemPrompt is the system prompt for the chat assistant.
	ChatSystemPrompt = `Ты - AI-ассистент сервиса Granula для планирования ремонта и перепланировки квартир в России.

Твои возможности:
1. Отвечать на вопросы о перепланировке и ремонте
2. Объяснять требования СНиП и Жилищного кодекса РФ
3. Предлагать изменения планировки
4. Помогать оптимизировать пространство

Правила:
- Всегда учитывай российские строительные нормы (СНиП, СП, ЖК РФ)
- Предупреждай о запрещённых перепланировках (снос несущих стен, перенос мокрых зон над жилыми комнатами)
- Отвечай на русском языке
- Будь конкретным и практичным
- Если предлагаешь действие, укажи его в формате JSON в конце ответа:

{"action": {"type": "DEMOLISH_WALL", "element_id": "...", "description": "..."}}

Типы действий:
- DEMOLISH_WALL: снос стены
- ADD_WALL: добавление стены
- MOVE_WET_ZONE: перенос мокрой зоны
- ADD_FURNITURE: добавление мебели
- CHANGE_ROOM_TYPE: изменение типа комнаты

Текущая сцена: %s`
)

// ChatService handles chat operations.
type ChatService struct {
	chatRepo *mongodb.ChatRepository
	client   *openrouter.Client
	log      *logger.Logger
}

// NewChatService creates a new ChatService.
func NewChatService(chatRepo *mongodb.ChatRepository, client *openrouter.Client, log *logger.Logger) *ChatService {
	return &ChatService{
		chatRepo: chatRepo,
		client:   client,
		log:      log,
	}
}

// SendMessage sends a message and gets a complete response.
func (s *ChatService) SendMessage(ctx context.Context, req SendMessageRequest) (*ChatResponse, error) {
	s.log.Info("sending chat message",
		logger.String("scene_id", req.SceneID),
		logger.String("branch_id", req.BranchID),
	)

	// Get or create context ID
	contextID := req.ContextID
	if contextID == "" {
		contextID = uuid.New().String()
	}

	// Save user message
	userMsg := entity.NewChatMessage(req.SceneID, req.BranchID, contextID, "user", req.Message)
	if err := s.chatRepo.Save(ctx, userMsg); err != nil {
		return nil, err
	}

	// Build message history for LLM
	messages, err := s.buildMessageHistory(ctx, req.SceneID, req.BranchID, contextID, req.Message)
	if err != nil {
		return nil, err
	}

	// Call OpenRouter
	systemPrompt := fmt.Sprintf(ChatSystemPrompt, s.getSceneSummary(req.SceneID))
	resp, err := s.client.ChatCompletionWithOptions(ctx, messages, openrouter.ChatOptions{
		SystemPrompt: systemPrompt,
	})
	if err != nil {
		s.log.Error("OpenRouter request failed", logger.Err(err))
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	// Parse response
	content := resp.Choices[0].Message.Content
	actions := s.parseActions(content)

	// Save assistant message
	assistantMsg := entity.NewChatMessage(req.SceneID, req.BranchID, contextID, "assistant", content)
	assistantMsg.WithActions(actions)
	assistantMsg.WithTokenUsage(&entity.TokenUsage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	})

	if err := s.chatRepo.Save(ctx, assistantMsg); err != nil {
		s.log.Warn("failed to save assistant message", logger.Err(err))
	}

	return &ChatResponse{
		MessageID:        assistantMsg.ID.String(),
		Response:         content,
		ContextID:        contextID,
		Actions:          actions,
		GenerationTimeMs: 0, // TODO: track time
		TokenUsage:       assistantMsg.TokenUsage,
	}, nil
}

// StreamMessage sends a message and streams the response.
func (s *ChatService) StreamMessage(ctx context.Context, req SendMessageRequest) (<-chan StreamChunk, error) {
	s.log.Info("streaming chat message",
		logger.String("scene_id", req.SceneID),
		logger.String("branch_id", req.BranchID),
	)

	// Get or create context ID
	contextID := req.ContextID
	if contextID == "" {
		contextID = uuid.New().String()
	}

	// Save user message
	userMsg := entity.NewChatMessage(req.SceneID, req.BranchID, contextID, "user", req.Message)
	if err := s.chatRepo.Save(ctx, userMsg); err != nil {
		return nil, err
	}

	// Build message history
	messages, err := s.buildMessageHistory(ctx, req.SceneID, req.BranchID, contextID, req.Message)
	if err != nil {
		return nil, err
	}

	// Start streaming
	systemPrompt := fmt.Sprintf(ChatSystemPrompt, s.getSceneSummary(req.SceneID))
	stream, err := s.client.ChatCompletionStream(ctx, messages, openrouter.ChatOptions{
		SystemPrompt: systemPrompt,
	})
	if err != nil {
		return nil, err
	}

	// Create output channel
	output := make(chan StreamChunk, 100)
	messageID := uuid.New().String()

	go func() {
		defer close(output)

		var fullContent strings.Builder
		chunkIndex := 0

		// Send initial chunk with IDs
		output <- StreamChunk{
			MessageID: messageID,
			ContextID: contextID,
			Index:     0,
		}
		chunkIndex++

		for event := range stream {
			if event.Error != nil {
				output <- StreamChunk{
					Error: event.Error,
					Done:  true,
				}
				return
			}

			if event.Content != "" {
				fullContent.WriteString(event.Content)
				output <- StreamChunk{
					Content: event.Content,
					Index:   chunkIndex,
				}
				chunkIndex++
			}

			if event.Done {
				// Parse actions from full content
				content := fullContent.String()
				actions := s.parseActions(content)

				// Save assistant message
				assistantMsg := entity.NewChatMessage(req.SceneID, req.BranchID, contextID, "assistant", content)
				assistantMsg.WithActions(actions)
				_ = s.chatRepo.Save(ctx, assistantMsg)

				// Send final chunk
				output <- StreamChunk{
					Done:    true,
					Actions: actions,
					Index:   chunkIndex,
				}
				return
			}
		}
	}()

	return output, nil
}

// GetHistory retrieves chat history.
func (s *ChatService) GetHistory(ctx context.Context, req GetHistoryRequest) (*GetHistoryResponse, error) {
	messages, hasMore, err := s.chatRepo.GetHistory(ctx, req.SceneID, req.BranchID, mongodb.GetHistoryOptions{
		ContextID: req.ContextID,
		Limit:     req.Limit,
		Cursor:    req.Cursor,
	})
	if err != nil {
		return nil, err
	}

	var nextCursor string
	if hasMore && len(messages) > 0 {
		nextCursor = messages[len(messages)-1].ID.String()
	}

	return &GetHistoryResponse{
		Messages:   messages,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}, nil
}

// ClearHistory clears chat history.
func (s *ChatService) ClearHistory(ctx context.Context, sceneID, branchID, contextID string) (int64, error) {
	return s.chatRepo.DeleteHistory(ctx, sceneID, branchID, contextID)
}

// buildMessageHistory builds the message history for LLM context.
func (s *ChatService) buildMessageHistory(ctx context.Context, sceneID, branchID, contextID, newMessage string) ([]openrouter.Message, error) {
	// Get recent messages
	recentMessages, err := s.chatRepo.GetRecentMessages(ctx, sceneID, branchID, contextID, 10)
	if err != nil {
		s.log.Warn("failed to get recent messages", logger.Err(err))
		recentMessages = []*entity.ChatMessage{}
	}

	// Build messages array
	messages := make([]openrouter.Message, 0, len(recentMessages)+1)

	for _, msg := range recentMessages {
		messages = append(messages, openrouter.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Add new user message
	messages = append(messages, openrouter.Message{
		Role:    "user",
		Content: newMessage,
	})

	return messages, nil
}

// parseActions extracts suggested actions from AI response.
func (s *ChatService) parseActions(content string) []entity.SuggestedAction {
	actions := make([]entity.SuggestedAction, 0)

	// Look for JSON action block at the end
	idx := strings.LastIndex(content, `{"action":`)
	if idx == -1 {
		return actions
	}

	jsonStr := content[idx:]
	// Find the closing brace
	braceCount := 0
	endIdx := -1
	for i, c := range jsonStr {
		if c == '{' {
			braceCount++
		} else if c == '}' {
			braceCount--
			if braceCount == 0 {
				endIdx = i + 1
				break
			}
		}
	}

	if endIdx == -1 {
		return actions
	}

	jsonStr = jsonStr[:endIdx]

	// Parse JSON
	var actionWrapper struct {
		Action struct {
			Type        string            `json:"type"`
			ElementID   string            `json:"element_id"`
			Description string            `json:"description"`
			Params      map[string]string `json:"params"`
		} `json:"action"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &actionWrapper); err != nil {
		s.log.Debug("failed to parse action JSON", logger.String("json", jsonStr), logger.Err(err))
		return actions
	}

	if actionWrapper.Action.Type != "" {
		actions = append(actions, entity.SuggestedAction{
			ID:                   uuid.New().String(),
			Type:                 actionWrapper.Action.Type,
			Description:          actionWrapper.Action.Description,
			Params:               actionWrapper.Action.Params,
			Confidence:           0.8,
			RequiresConfirmation: true,
		})
	}

	return actions
}

// getSceneSummary returns a summary of the scene for context.
// TODO: This should fetch actual scene data from Scene Service.
func (s *ChatService) getSceneSummary(sceneID string) string {
	return fmt.Sprintf("Scene ID: %s (данные сцены будут загружены из Scene Service)", sceneID)
}

// Request/Response types

// SendMessageRequest for sending a chat message.
type SendMessageRequest struct {
	SceneID   string
	BranchID  string
	Message   string
	ContextID string
}

// ChatResponse for chat completion.
type ChatResponse struct {
	MessageID        string
	Response         string
	ContextID        string
	Actions          []entity.SuggestedAction
	GenerationTimeMs int64
	TokenUsage       *entity.TokenUsage
}

// StreamChunk for streaming response.
type StreamChunk struct {
	Content   string
	MessageID string
	ContextID string
	Done      bool
	Actions   []entity.SuggestedAction
	Error     error
	Index     int
}

// GetHistoryRequest for fetching chat history.
type GetHistoryRequest struct {
	SceneID   string
	BranchID  string
	ContextID string
	Limit     int
	Cursor    string
}

// GetHistoryResponse for chat history.
type GetHistoryResponse struct {
	Messages   []*entity.ChatMessage
	HasMore    bool
	NextCursor string
}

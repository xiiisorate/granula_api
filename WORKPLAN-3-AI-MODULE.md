# 🤖 WORKPLAN-3: AI Модуль (Критический)

> **Приоритет:** 🔴 КРИТИЧЕСКИЙ — ключевая функциональность сервиса  
> **Время:** 4-5 часов  
> **Зависимости:** WORKPLAN-1-PROTO.md  
> **Результат:** AI распознаёт планы, чат знает контекст, генерация работает

---

## 🎯 ЦЕЛЬ

Исправить AI модуль чтобы:
1. **Распознавание изображений работало** — отправлять реальные картинки в AI
2. **Чат знал контекст сцены** — загружать данные из Scene Service
3. **Генерация вариантов работала** — получать данные планировки
4. **SelectSuggestion был реализован** — выбор варианта из AI

---

## 📋 ПРОБЛЕМЫ (подробный анализ)

### 🚨 Проблема 1: Изображения НЕ отправляются в AI

**Файл:** `ai-service/internal/service/recognition_service.go`  
**Строки:** 88-95

```go
// СЕЙЧАС — КРИТИЧЕСКАЯ ОШИБКА!
messages := []openrouter.Message{
    {
        Role:    "user",
        Content: prompt + "\n\n[Изображение планировки загружено: " + dataURL[:100] + "...]",
        //                                                          ^^^^^^^^^^^^
        //                                                          Только 100 символов!
    },
}
```

**Что происходит:**
1. Изображение читается и конвертируется в base64 (строка 68)
2. Создаётся data URL (строка 69): `data:image/png;base64,iVBORw0...` (очень длинный)
3. **НО:** В AI отправляются только первые 100 символов как текст
4. AI получает: `[Изображение планировки загружено: data:image/png;base64,iVBORw0KGgoAAAA...]`
5. **Результат:** AI не видит изображение, распознавание невозможно!

**Интересно:** В `openrouter/client.go` (строки 63-74) уже есть структуры для Vision:
```go
type ImageContent struct {
    Type     string    `json:"type"` // "text" or "image_url"
    Text     string    `json:"text,omitempty"`
    ImageURL *ImageURL `json:"image_url,omitempty"`
}

type ImageURL struct {
    URL    string `json:"url"`
    Detail string `json:"detail,omitempty"`
}
```

**НО:** Метода для отправки изображений нет!

---

### 🚨 Проблема 2: Чат не знает планировку

**Файл:** `ai-service/internal/service/chat_service.go`  
**Строки:** 314-318

```go
// TODO: This should fetch actual scene data from Scene Service via gRPC.
func (s *ChatService) getSceneSummary(sceneID string) string {
    return "Scene ID: " + sceneID + " (данные сцены будут загружены из Scene Service)"
}
```

**Последствия:**
- Промпт `ChatSystemPrompt` содержит `%s` для контекста
- Но вместо реальных данных приходит заглушка
- AI не знает какие стены, комнаты, проёмы есть
- Не может давать конкретные рекомендации

---

### 🚨 Проблема 3: Генерация без данных сцены

**Файл:** `ai-service/internal/grpc/server.go`  
**Строка:** 131

```go
generateReq := service.GenerateRequest{
    SceneID:       req.SceneId,
    BranchID:      req.BranchId,
    Prompt:        req.Prompt,
    VariantsCount: int(req.VariantsCount),
    Options:       options,
    SceneData:     "", // TODO: fetch from Scene Service  <-- ПУСТАЯ СТРОКА!
}
```

**Последствия:**
- `GenerationSystemPrompt` тоже использует `%s` для данных
- AI генерирует абстрактные варианты
- Не может указать конкретные `element_ids`

---

### 🚨 Проблема 4: SelectSuggestion не реализован

**Документация:** `docs/api/chat.md` (строки 187-219)

Описан endpoint:
```
POST /api/v1/scenes/:sceneId/chat/messages/:messageId/select
```

**НЕ реализован** ни в AI Service, ни в API Gateway!

---

### 🚨 Проблема 5: GetContext/UpdateContext не реализованы

**Файл:** `ai-service/internal/grpc/server.go`  
**Строки:** 296-303

```go
func (s *AIServer) GetContext(...) {
    return nil, apperrors.Internal("not implemented").ToGRPCError()
}
func (s *AIServer) UpdateContext(...) {
    return nil, apperrors.Internal("not implemented").ToGRPCError()
}
```

---

## 🔧 ПОШАГОВАЯ ИНСТРУКЦИЯ

### ШАГ 1: Добавить Vision API в OpenRouter клиент

**Файл:** `ai-service/internal/openrouter/client.go`

#### 1.1. Добавить типы для multimodal сообщений

```go
// После существующих типов (после строки 74)

// MultimodalMessage represents a message with text and/or images.
type MultimodalMessage struct {
    Role    string        `json:"role"`
    Content []ContentPart `json:"content"`
}

// ContentPart is a part of multimodal message content.
type ContentPart struct {
    Type     string    `json:"type"` // "text" or "image_url"
    Text     string    `json:"text,omitempty"`
    ImageURL *ImageURL `json:"image_url,omitempty"`
}

// MultimodalChatRequest is the request body for multimodal chat completions.
type MultimodalChatRequest struct {
    Model       string              `json:"model"`
    Messages    []MultimodalMessage `json:"messages"`
    MaxTokens   int                 `json:"max_tokens,omitempty"`
    Temperature float64             `json:"temperature,omitempty"`
}
```

#### 1.2. Добавить метод ChatCompletionWithImages

```go
// ChatCompletionWithImages performs a chat completion with image inputs.
// Use this for vision models like claude-sonnet-4 or gpt-4o.
func (c *Client) ChatCompletionWithImages(ctx context.Context, messages []MultimodalMessage, opts ChatOptions) (*ChatResponse, error) {
    // Wait for rate limit
    if err := c.waitForRateLimit(ctx); err != nil {
        return nil, err
    }

    // Use vision model
    model := "anthropic/claude-sonnet-4-20250514" // Vision-capable model
    if opts.Model != "" {
        model = opts.Model
    }

    // Prepend system message if provided
    if opts.SystemPrompt != "" {
        systemMsg := MultimodalMessage{
            Role: "system",
            Content: []ContentPart{
                {Type: "text", Text: opts.SystemPrompt},
            },
        }
        messages = append([]MultimodalMessage{systemMsg}, messages...)
    }

    maxTokens := c.cfg.MaxTokens
    if opts.MaxTokens > 0 {
        maxTokens = opts.MaxTokens
    }

    temperature := c.cfg.Temperature
    if opts.Temperature > 0 {
        temperature = opts.Temperature
    }

    req := MultimodalChatRequest{
        Model:       model,
        Messages:    messages,
        MaxTokens:   maxTokens,
        Temperature: temperature,
    }

    // Execute with retries
    var lastErr error
    for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
        if attempt > 0 {
            backoff := time.Duration(1<<uint(attempt-1)) * 100 * time.Millisecond
            select {
            case <-ctx.Done():
                return nil, ctx.Err()
            case <-time.After(backoff):
            }
        }

        resp, err := c.doMultimodalRequest(ctx, req)
        if err == nil {
            return resp, nil
        }

        lastErr = err
        c.log.Warn("OpenRouter multimodal request failed, retrying",
            logger.Int("attempt", attempt+1),
            logger.Err(err),
        )
    }

    return nil, apperrors.Wrap(lastErr, "all retries exhausted")
}

// doMultimodalRequest performs the actual HTTP request for multimodal.
func (c *Client) doMultimodalRequest(ctx context.Context, req MultimodalChatRequest) (*ChatResponse, error) {
    body, err := json.Marshal(req)
    if err != nil {
        return nil, apperrors.Internal("failed to marshal request").WithCause(err)
    }

    httpReq, err := http.NewRequestWithContext(ctx, "POST", c.cfg.BaseURL+"/chat/completions", bytes.NewReader(body))
    if err != nil {
        return nil, apperrors.Internal("failed to create request").WithCause(err)
    }

    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
    httpReq.Header.Set("HTTP-Referer", "https://granula.ru")
    httpReq.Header.Set("X-Title", "Granula")

    c.log.Debug("sending OpenRouter multimodal request",
        logger.String("model", req.Model),
        logger.Int("messages", len(req.Messages)),
    )

    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, apperrors.Unavailable("openrouter").WithCause(err)
    }
    defer resp.Body.Close()

    c.recordRequest()

    if resp.StatusCode != http.StatusOK {
        bodyBytes, _ := io.ReadAll(resp.Body)
        c.log.Error("OpenRouter error response",
            logger.Int("status", resp.StatusCode),
            logger.String("body", string(bodyBytes)),
        )

        if resp.StatusCode == 429 {
            return nil, apperrors.RateLimited("OpenRouter rate limit exceeded")
        }
        return nil, apperrors.Internalf("OpenRouter error: %d - %s", resp.StatusCode, string(bodyBytes))
    }

    var chatResp ChatResponse
    if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
        return nil, apperrors.Internal("failed to decode response").WithCause(err)
    }

    c.log.Debug("OpenRouter multimodal response received",
        logger.Int("prompt_tokens", chatResp.Usage.PromptTokens),
        logger.Int("completion_tokens", chatResp.Usage.CompletionTokens),
    )

    return &chatResp, nil
}
```

---

### ШАГ 2: Исправить RecognitionService

**Файл:** `ai-service/internal/service/recognition_service.go`

#### 2.1. Заменить метод processRecognition (строки 60-148)

```go
// processRecognition performs the actual recognition.
func (s *RecognitionService) processRecognition(ctx context.Context, job *entity.RecognitionJob, imageData []byte, imageType string) {
    startTime := time.Now()

    // Mark as processing
    job.Start()
    _ = s.jobRepo.UpdateRecognitionJob(ctx, job)

    // Encode image to base64 data URL
    base64Image := base64.StdEncoding.EncodeToString(imageData)
    dataURL := fmt.Sprintf("data:%s;base64,%s", imageType, base64Image)

    // Update progress
    job.UpdateProgress(10)
    _ = s.jobRepo.UpdateRecognitionJob(ctx, job)

    // Build prompt
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
    prompt += "Верни результат ТОЛЬКО в формате JSON без markdown."

    // Build multimodal message with REAL image
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
                        URL:    dataURL, // Полное изображение в base64!
                        Detail: "high",  // Высокое качество для точного распознавания
                    },
                },
            },
        },
    }

    job.UpdateProgress(30)
    _ = s.jobRepo.UpdateRecognitionJob(ctx, job)

    // Call OpenRouter with Vision API
    resp, err := s.client.ChatCompletionWithImages(ctx, messages, openrouter.ChatOptions{
        SystemPrompt: prompts.GetRecognitionPrompt(),
        MaxTokens:    8192,
        Temperature:  0.2,
        Model:        "anthropic/claude-sonnet-4-20250514", // Vision model
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
```

---

### ШАГ 3: Интегрировать AI Service с Scene Service

#### 3.1. Добавить Scene gRPC клиент в AI Service

**Создать файл:** `ai-service/internal/grpc/scene_client.go`

```go
package grpc

import (
    "context"
    "encoding/json"
    "fmt"

    scenepb "github.com/xiiisorate/granula_api/shared/gen/scene/v1"
    "google.golang.org/grpc"
)

// SceneClient wraps scene service gRPC client.
type SceneClient struct {
    client scenepb.SceneServiceClient
}

// NewSceneClient creates a new scene client.
func NewSceneClient(addr string) (*SceneClient, error) {
    conn, err := grpc.Dial(addr, grpc.WithInsecure())
    if err != nil {
        return nil, err
    }
    return &SceneClient{
        client: scenepb.NewSceneServiceClient(conn),
    }, nil
}

// GetSceneContext returns scene data formatted for AI context.
func (c *SceneClient) GetSceneContext(ctx context.Context, sceneID string) (string, error) {
    // Get scene
    scene, err := c.client.GetScene(ctx, &scenepb.GetSceneRequest{Id: sceneID})
    if err != nil {
        return "", err
    }

    // Get elements
    elements, err := c.client.ListElements(ctx, &scenepb.ListElementsRequest{
        SceneId: sceneID,
        Limit:   1000,
    })
    if err != nil {
        return "", err
    }

    // Format for AI
    context := struct {
        SceneID    string `json:"scene_id"`
        Name       string `json:"name"`
        TotalArea  float64 `json:"total_area"`
        Walls      []interface{} `json:"walls"`
        Rooms      []interface{} `json:"rooms"`
        Openings   []interface{} `json:"openings"`
        Furniture  []interface{} `json:"furniture"`
    }{
        SceneID:   sceneID,
        Name:      scene.Scene.Name,
        TotalArea: float64(scene.Scene.TotalArea),
        Walls:     make([]interface{}, 0),
        Rooms:     make([]interface{}, 0),
        Openings:  make([]interface{}, 0),
        Furniture: make([]interface{}, 0),
    }

    for _, el := range elements.Elements {
        switch el.Type {
        case "wall":
            context.Walls = append(context.Walls, map[string]interface{}{
                "id":              el.Id,
                "is_load_bearing": el.Properties["is_load_bearing"],
                "thickness":       el.Properties["thickness"],
            })
        case "room":
            context.Rooms = append(context.Rooms, map[string]interface{}{
                "id":          el.Id,
                "room_type":   el.Properties["room_type"],
                "area":        el.Properties["area"],
                "is_wet_zone": el.Properties["is_wet_zone"],
            })
        case "door", "window":
            context.Openings = append(context.Openings, map[string]interface{}{
                "id":    el.Id,
                "type":  el.Type,
                "width": el.Properties["width"],
            })
        case "furniture":
            context.Furniture = append(context.Furniture, map[string]interface{}{
                "id":   el.Id,
                "name": el.Name,
            })
        }
    }

    jsonBytes, err := json.MarshalIndent(context, "", "  ")
    if err != nil {
        return "", err
    }

    return fmt.Sprintf("Текущая планировка:\n```json\n%s\n```", string(jsonBytes)), nil
}
```

#### 3.2. Обновить ChatService

**Файл:** `ai-service/internal/service/chat_service.go`

**Добавить поле sceneClient в структуру:**
```go
type ChatService struct {
    chatRepo    *mongodb.ChatRepository
    client      *openrouter.Client
    sceneClient *grpc.SceneClient  // NEW
    log         *logger.Logger
}

func NewChatService(chatRepo *mongodb.ChatRepository, client *openrouter.Client, sceneClient *grpc.SceneClient, log *logger.Logger) *ChatService {
    return &ChatService{
        chatRepo:    chatRepo,
        client:      client,
        sceneClient: sceneClient,  // NEW
        log:         log,
    }
}
```

**Обновить метод getSceneSummary (строки 314-318):**
```go
// getSceneSummary returns a summary of the scene for context.
func (s *ChatService) getSceneSummary(sceneID string) string {
    if sceneID == "" {
        return "Контекст сцены не загружен. Спроси пользователя о деталях планировки."
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    summary, err := s.sceneClient.GetSceneContext(ctx, sceneID)
    if err != nil {
        s.log.Warn("failed to get scene context", logger.Err(err))
        return "Scene ID: " + sceneID + " (не удалось загрузить данные)"
    }
    
    return summary
}
```

---

### ШАГ 4: Обновить GenerateVariants для получения данных сцены

**Файл:** `ai-service/internal/grpc/server.go`

**Обновить метод GenerateVariants (строки 104-144):**

```go
// GenerateVariants generates layout variants.
func (s *AIServer) GenerateVariants(ctx context.Context, req *pb.GenerateVariantsRequest) (*pb.GenerateVariantsResponse, error) {
    s.log.Info("GenerateVariants called",
        logger.String("scene_id", req.SceneId),
        logger.Int("variants_count", int(req.VariantsCount)),
    )

    // Fetch scene data from Scene Service
    sceneData := ""
    if s.sceneClient != nil && req.SceneId != "" {
        data, err := s.sceneClient.GetSceneContext(ctx, req.SceneId)
        if err != nil {
            s.log.Warn("failed to get scene data for generation", logger.Err(err))
        } else {
            sceneData = data
        }
    }

    options := entity.GenerationOptions{
        PreserveLoadBearing: req.Options.GetPreserveLoadBearing(),
        CheckCompliance:     req.Options.GetCheckCompliance(),
        PreserveWetZones:    req.Options.GetPreserveWetZones(),
        Style:               convertGenerationStyleFromPB(req.Options.GetStyle()),
        Budget:              float64(req.Options.GetBudget()),
    }

    generateReq := service.GenerateRequest{
        SceneID:       req.SceneId,
        BranchID:      req.BranchId,
        Prompt:        req.Prompt,
        VariantsCount: int(req.VariantsCount),
        Options:       options,
        SceneData:     sceneData, // NOW WITH REAL DATA!
    }

    job, err := s.generationService.StartGeneration(ctx, generateReq)
    if err != nil {
        return nil, apperrors.FromGRPCError(err).ToGRPCError()
    }

    return &pb.GenerateVariantsResponse{
        Success: true,
        JobId:   job.ID.String(),
        Status:  convertJobStatusToPB(job.Status),
    }, nil
}
```

---

### ШАГ 5: Реализовать SelectSuggestion

#### 5.1. Добавить метод в AI Service proto

**Файл:** `shared/proto/ai/v1/ai.proto`

**Добавить в service AIService:**
```protobuf
// SelectSuggestion selects a suggestion from AI response.
rpc SelectSuggestion(SelectSuggestionRequest) returns (SelectSuggestionResponse);
```

**Добавить messages:**
```protobuf
message SelectSuggestionRequest {
    string scene_id = 1;
    string message_id = 2;
    int32 suggestion_index = 3;
}

message SelectSuggestionResponse {
    string selected_branch_id = 1;
    bool branch_activated = 2;
    string confirmation_message = 3;
}
```

#### 5.2. Реализовать в AI Server

**Файл:** `ai-service/internal/grpc/server.go`

```go
// SelectSuggestion selects a variant from AI suggestions.
func (s *AIServer) SelectSuggestion(ctx context.Context, req *pb.SelectSuggestionRequest) (*pb.SelectSuggestionResponse, error) {
    s.log.Info("SelectSuggestion called",
        logger.String("scene_id", req.SceneId),
        logger.String("message_id", req.MessageId),
        logger.Int("suggestion_index", int(req.SuggestionIndex)),
    )

    // Get the message with suggestions
    messageID, err := uuid.Parse(req.MessageId)
    if err != nil {
        return nil, apperrors.InvalidArgument("message_id", "invalid UUID").ToGRPCError()
    }

    message, err := s.chatService.GetMessage(ctx, messageID)
    if err != nil {
        return nil, apperrors.NotFound("message", req.MessageId).ToGRPCError()
    }

    // Validate suggestion index
    if int(req.SuggestionIndex) >= len(message.Actions) {
        return nil, apperrors.InvalidArgument("suggestion_index", "out of range").ToGRPCError()
    }

    selectedAction := message.Actions[req.SuggestionIndex]
    branchID := selectedAction.Params["branch_id"]

    // TODO: Activate branch via Branch Service
    // branchClient.Activate(ctx, branchID)

    // Create confirmation message
    confirmationMsg := entity.NewChatMessage(req.SceneId, branchID, message.ContextID, "assistant", 
        fmt.Sprintf("Отлично! Я активировал вариант \"%s\". Теперь вы можете редактировать планировку в 3D редакторе.", selectedAction.Description))
    _ = s.chatService.SaveMessage(ctx, confirmationMsg)

    return &pb.SelectSuggestionResponse{
        SelectedBranchId:    branchID,
        BranchActivated:     true,
        ConfirmationMessage: confirmationMsg.Content,
    }, nil
}
```

#### 5.3. Добавить endpoint в API Gateway

**Файл:** `api-gateway/internal/handlers/ai.go`

```go
// SelectSuggestion selects a variant from AI suggestions.
// @Summary Выбрать вариант
// @Description Выбор варианта из предложенных AI
// @Tags ai
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param scene_id path string true "ID сцены"
// @Param message_id path string true "ID сообщения"
// @Param body body SelectSuggestionInput true "Индекс варианта"
// @Success 200 {object} SelectSuggestionResponse
// @Router /scenes/{scene_id}/chat/messages/{message_id}/select [post]
func (h *AIHandler) SelectSuggestion(c *fiber.Ctx) error {
    sceneID := c.Params("scene_id")
    messageID := c.Params("message_id")

    var input SelectSuggestionInput
    if err := c.BodyParser(&input); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
    }

    ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
    defer cancel()

    resp, err := h.client.SelectSuggestion(ctx, &aipb.SelectSuggestionRequest{
        SceneId:         sceneID,
        MessageId:       messageID,
        SuggestionIndex: int32(input.SuggestionIndex),
    })
    if err != nil {
        return handleGRPCError(err)
    }

    return c.JSON(fiber.Map{
        "data": fiber.Map{
            "selected_branch_id":    resp.SelectedBranchId,
            "branch_activated":      resp.BranchActivated,
            "confirmation_message":  resp.ConfirmationMessage,
        },
        "request_id": c.GetRespHeader("X-Request-ID"),
    })
}

// SelectSuggestionInput - input for selecting suggestion.
type SelectSuggestionInput struct {
    SuggestionIndex int `json:"suggestion_index" validate:"gte=0,lte=4"`
}
```

**Добавить route в main.go:**
```go
// В группе scenes
scenes.Post("/:scene_id/chat/messages/:message_id/select", aiHandler.SelectSuggestion)
```

---

### ШАГ 6: Реализовать GetContext/UpdateContext

**Файл:** `ai-service/internal/grpc/server.go`

```go
// GetContext retrieves AI context for a scene.
func (s *AIServer) GetContext(ctx context.Context, req *pb.GetContextRequest) (*pb.GetContextResponse, error) {
    // Get scene summary
    summary := ""
    if s.sceneClient != nil && req.SceneId != "" {
        data, err := s.sceneClient.GetSceneContext(ctx, req.SceneId)
        if err == nil {
            summary = data
        }
    }

    // Get recent messages for context size estimation
    messages, err := s.chatService.GetRecentMessages(ctx, req.SceneId, req.BranchId, "", 10)
    contextSize := 0
    for _, msg := range messages {
        contextSize += openrouter.EstimateTokens(msg.Content)
    }

    return &pb.GetContextResponse{
        SceneId:      req.SceneId,
        BranchId:     req.BranchId,
        SceneSummary: summary,
        ContextSize:  int32(contextSize),
    }, nil
}

// UpdateContext updates AI context (reloads scene data).
func (s *AIServer) UpdateContext(ctx context.Context, req *pb.UpdateContextRequest) (*pb.UpdateContextResponse, error) {
    // Force reload scene data
    summary := ""
    if s.sceneClient != nil && req.SceneId != "" {
        data, err := s.sceneClient.GetSceneContext(ctx, req.SceneId)
        if err != nil {
            return nil, apperrors.Internal("failed to load scene").WithCause(err).ToGRPCError()
        }
        summary = data
    }

    return &pb.UpdateContextResponse{
        Success:      true,
        SceneSummary: summary,
    }, nil
}
```

---

### ШАГ 7: Добавить tracking времени генерации

**Файл:** `ai-service/internal/service/chat_service.go`

**В методе SendMessage (строка 35):**
```go
func (s *ChatService) SendMessage(ctx context.Context, req SendMessageRequest) (*ChatResponse, error) {
    startTime := time.Now() // ADD THIS
    
    // ... existing code ...
    
    return &ChatResponse{
        MessageID:        assistantMsg.ID.String(),
        Response:         content,
        ContextID:        contextID,
        Actions:          actions,
        GenerationTimeMs: time.Since(startTime).Milliseconds(), // CHANGE THIS
        TokenUsage:       assistantMsg.TokenUsage,
    }, nil
}
```

---

## ✅ КРИТЕРИИ УСПЕХА

### Vision API
- [ ] Метод `ChatCompletionWithImages` добавлен в OpenRouter клиент
- [ ] `MultimodalMessage` и `ContentPart` типы определены
- [ ] RecognitionService отправляет реальные изображения
- [ ] При загрузке планировки AI видит картинку

### Scene Integration
- [ ] SceneClient создан в AI Service
- [ ] ChatService получает данные сцены
- [ ] GenerationService получает данные сцены
- [ ] Промпты содержат реальный контекст планировки

### SelectSuggestion
- [ ] Proto обновлён с новым RPC
- [ ] Метод реализован в AI Server
- [ ] Endpoint добавлен в API Gateway
- [ ] Пользователь может выбрать вариант

### Context Management
- [ ] GetContext реализован
- [ ] UpdateContext реализован

---

## 📚 СВЯЗАННАЯ ДОКУМЕНТАЦИЯ

| Документ | Путь | Для чего |
|----------|------|----------|
| AI Chat API | `docs/api/chat.md` | Спецификация endpoints |
| Промпты | `ai-service/internal/prompts/prompts.go` | Системные промпты |
| OpenRouter | `ai-service/internal/openrouter/client.go` | HTTP клиент |
| Recognition | `ai-service/internal/service/recognition_service.go` | Распознавание |
| Chat | `ai-service/internal/service/chat_service.go` | Чат |
| Generation | `ai-service/internal/service/generation_service.go` | Генерация |

---

## 🧪 ТЕСТИРОВАНИЕ

### Тест распознавания
```bash
# 1. Загрузить планировку
curl -X POST http://localhost:8080/api/v1/floor-plans \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@plan.png" \
  -F "workspace_id=ws_123"

# 2. Запустить распознавание
curl -X POST http://localhost:8080/api/v1/floor-plans/fp_123/recognize \
  -H "Authorization: Bearer $TOKEN"

# 3. Проверить статус
curl http://localhost:8080/api/v1/floor-plans/fp_123/recognition-status \
  -H "Authorization: Bearer $TOKEN"
```

### Тест чата с контекстом
```bash
curl -X POST http://localhost:8080/api/v1/ai/chat \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "scene_id": "sc_123",
    "message": "Можно ли снести стену между кухней и гостиной?"
  }'
```

### Тест генерации
```bash
curl -X POST http://localhost:8080/api/v1/ai/generate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "scene_id": "sc_123",
    "prompt": "Объедини кухню с гостиной",
    "variants_count": 3
  }'
```

---

## ➡️ СЛЕДУЮЩИЙ ШАГ

После исправления AI модуля, переходите к:
- [WORKPLAN-4-INTEGRATIONS.md](./WORKPLAN-4-INTEGRATIONS.md) — интеграции между сервисами


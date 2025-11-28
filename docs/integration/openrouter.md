# Интеграция с OpenRouter

## Обзор

OpenRouter используется для AI функционала:
- Распознавание планировок (Vision API)
- Генерация вариантов планировки
- Чат-ассистент
- Проверка соответствия нормам

## Конфигурация

```env
# OpenRouter API
OPENROUTER_API_KEY=sk-or-v1-xxx
OPENROUTER_BASE_URL=https://openrouter.ai/api/v1
OPENROUTER_DEFAULT_MODEL=anthropic/claude-sonnet-4
OPENROUTER_VISION_MODEL=anthropic/claude-sonnet-4
OPENROUTER_TIMEOUT=120s
OPENROUTER_MAX_RETRIES=3

# Rate Limiting
OPENROUTER_RPM=60
OPENROUTER_TPM=100000
```

## Архитектура

```
┌─────────────────────────────────────────────────────────────┐
│                      Granula API                             │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                    AI Service                            ││
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐     ││
│  │  │ Recognition │  │ Generation  │  │    Chat     │     ││
│  │  │   Service   │  │   Service   │  │   Service   │     ││
│  │  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘     ││
│  │         │                │                │             ││
│  │         └────────────────┼────────────────┘             ││
│  │                          │                              ││
│  │                   ┌──────┴──────┐                       ││
│  │                   │ OpenRouter  │                       ││
│  │                   │   Client    │                       ││
│  │                   └──────┬──────┘                       ││
│  └──────────────────────────┼──────────────────────────────┘│
│                             │                               │
│                      ┌──────┴──────┐                        │
│                      │ Worker Pool │                        │
│                      │  (5 workers)│                        │
│                      └──────┬──────┘                        │
└─────────────────────────────┼───────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │   OpenRouter    │
                    │      API        │
                    └─────────────────┘
```

## Клиент OpenRouter

```go
// internal/integration/openrouter/client.go

// Client клиент OpenRouter API.
type Client struct {
    httpClient  *http.Client
    baseURL     string
    apiKey      string
    defaultModel string
    visionModel  string
    logger      *zap.Logger
    metrics     *Metrics
}

// Config конфигурация клиента.
type Config struct {
    APIKey       string        `env:"OPENROUTER_API_KEY,required"`
    BaseURL      string        `env:"OPENROUTER_BASE_URL" envDefault:"https://openrouter.ai/api/v1"`
    DefaultModel string        `env:"OPENROUTER_DEFAULT_MODEL" envDefault:"anthropic/claude-sonnet-4"`
    VisionModel  string        `env:"OPENROUTER_VISION_MODEL" envDefault:"anthropic/claude-sonnet-4"`
    Timeout      time.Duration `env:"OPENROUTER_TIMEOUT" envDefault:"120s"`
    MaxRetries   int           `env:"OPENROUTER_MAX_RETRIES" envDefault:"3"`
}

// NewClient создаёт новый клиент.
func NewClient(cfg *Config, logger *zap.Logger) *Client {
    return &Client{
        httpClient: &http.Client{
            Timeout: cfg.Timeout,
            Transport: &http.Transport{
                MaxIdleConns:        100,
                MaxIdleConnsPerHost: 100,
                IdleConnTimeout:     90 * time.Second,
            },
        },
        baseURL:      cfg.BaseURL,
        apiKey:       cfg.APIKey,
        defaultModel: cfg.DefaultModel,
        visionModel:  cfg.VisionModel,
        logger:       logger,
        metrics:      NewMetrics(),
    }
}

// ChatCompletion отправляет запрос на генерацию.
func (c *Client) ChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
    timer := prometheus.NewTimer(c.metrics.RequestDuration.WithLabelValues("chat"))
    defer timer.ObserveDuration()
    
    if req.Model == "" {
        req.Model = c.defaultModel
    }
    
    // Добавляем headers для OpenRouter
    headers := map[string]string{
        "Authorization": "Bearer " + c.apiKey,
        "Content-Type":  "application/json",
        "HTTP-Referer":  "https://granula.ru",
        "X-Title":       "Granula",
    }
    
    body, err := json.Marshal(req)
    if err != nil {
        return nil, fmt.Errorf("marshal request: %w", err)
    }
    
    httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }
    
    for k, v := range headers {
        httpReq.Header.Set(k, v)
    }
    
    resp, err := c.doWithRetry(httpReq)
    if err != nil {
        c.metrics.RequestErrors.WithLabelValues("chat").Inc()
        return nil, err
    }
    defer resp.Body.Close()
    
    var result ChatCompletionResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("decode response: %w", err)
    }
    
    // Логируем использование токенов
    c.metrics.TokensUsed.WithLabelValues("prompt").Add(float64(result.Usage.PromptTokens))
    c.metrics.TokensUsed.WithLabelValues("completion").Add(float64(result.Usage.CompletionTokens))
    
    return &result, nil
}

// doWithRetry выполняет запрос с ретраями.
func (c *Client) doWithRetry(req *http.Request) (*http.Response, error) {
    var lastErr error
    
    for attempt := 0; attempt <= c.maxRetries; attempt++ {
        if attempt > 0 {
            // Exponential backoff
            backoff := time.Duration(math.Pow(2, float64(attempt))) * 100 * time.Millisecond
            time.Sleep(backoff)
        }
        
        resp, err := c.httpClient.Do(req)
        if err != nil {
            lastErr = err
            continue
        }
        
        // Retry on 429 (rate limit) and 5xx
        if resp.StatusCode == 429 || resp.StatusCode >= 500 {
            resp.Body.Close()
            lastErr = fmt.Errorf("status %d", resp.StatusCode)
            continue
        }
        
        if resp.StatusCode != 200 {
            body, _ := io.ReadAll(resp.Body)
            resp.Body.Close()
            return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
        }
        
        return resp, nil
    }
    
    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

## Типы запросов/ответов

```go
// internal/integration/openrouter/types.go

// ChatCompletionRequest запрос генерации.
type ChatCompletionRequest struct {
    Model       string    `json:"model"`
    Messages    []Message `json:"messages"`
    MaxTokens   int       `json:"max_tokens,omitempty"`
    Temperature float64   `json:"temperature,omitempty"`
    TopP        float64   `json:"top_p,omitempty"`
    Stream      bool      `json:"stream,omitempty"`
    Stop        []string  `json:"stop,omitempty"`
}

// Message сообщение в диалоге.
type Message struct {
    Role    string        `json:"role"` // system, user, assistant
    Content MessageContent `json:"content"`
}

// MessageContent контент сообщения (текст или multimodal).
type MessageContent interface{}

// TextContent текстовый контент.
type TextContent string

// MultimodalContent мультимодальный контент.
type MultimodalContent []ContentPart

// ContentPart часть контента.
type ContentPart struct {
    Type     string    `json:"type"` // text, image_url
    Text     string    `json:"text,omitempty"`
    ImageURL *ImageURL `json:"image_url,omitempty"`
}

// ImageURL URL изображения.
type ImageURL struct {
    URL    string `json:"url"` // URL или base64 data URI
    Detail string `json:"detail,omitempty"` // low, high, auto
}

// ChatCompletionResponse ответ генерации.
type ChatCompletionResponse struct {
    ID      string   `json:"id"`
    Object  string   `json:"object"`
    Created int64    `json:"created"`
    Model   string   `json:"model"`
    Choices []Choice `json:"choices"`
    Usage   Usage    `json:"usage"`
}

// Choice вариант ответа.
type Choice struct {
    Index        int     `json:"index"`
    Message      Message `json:"message"`
    FinishReason string  `json:"finish_reason"`
}

// Usage использование токенов.
type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}
```

## Сервис распознавания планировок

```go
// internal/service/ai/recognition.go

// RecognitionService сервис распознавания планировок.
type RecognitionService struct {
    client *openrouter.Client
    logger *zap.Logger
}

// RecognizeFloorPlan распознаёт планировку из изображения.
func (s *RecognitionService) RecognizeFloorPlan(
    ctx context.Context,
    imageData []byte,
    hints *RecognitionHints,
) (*RecognitionResult, error) {
    // Кодируем изображение в base64
    base64Image := base64.StdEncoding.EncodeToString(imageData)
    imageURL := fmt.Sprintf("data:image/png;base64,%s", base64Image)
    
    // Формируем системный промпт
    systemPrompt := `Ты - эксперт по анализу архитектурных планов и чертежей квартир.
Твоя задача - распознать элементы планировки и вернуть структурированные данные в формате JSON.

Анализируй:
1. Стены (определи несущие по толщине - обычно > 15см)
2. Комнаты и их типы (кухня, спальня, ванная, коридор и т.д.)
3. Двери и окна с размерами
4. Инженерные элементы (стояки, вентиляция)

Верни JSON в следующем формате:
{
  "bounds": { "width": float, "height": float, "depth": float },
  "walls": [{ "id": string, "start": {x,y,z}, "end": {x,y,z}, "thickness": float, "is_load_bearing": bool }],
  "rooms": [{ "id": string, "type": string, "name": string, "polygon": [{x,z}], "area": float }],
  "openings": [{ "id": string, "type": "door"|"window", "wall_id": string, "position": float, "width": float, "height": float }],
  "utilities": [{ "id": string, "type": string, "position": {x,y,z} }]
}

Все размеры в метрах. Координатная система: X - ширина, Y - высота, Z - глубина.`

    // Добавляем подсказки если есть
    userPrompt := "Проанализируй эту планировку квартиры и верни структурированные данные."
    if hints != nil {
        if hints.Scale > 0 {
            userPrompt += fmt.Sprintf("\nМасштаб чертежа: 1:%d", hints.Scale)
        }
        if len(hints.LoadBearingWalls) > 0 {
            userPrompt += fmt.Sprintf("\nИзвестные несущие стены: %v", hints.LoadBearingWalls)
        }
    }
    
    req := &openrouter.ChatCompletionRequest{
        Model: s.client.VisionModel(),
        Messages: []openrouter.Message{
            {Role: "system", Content: systemPrompt},
            {
                Role: "user",
                Content: openrouter.MultimodalContent{
                    {Type: "text", Text: userPrompt},
                    {Type: "image_url", ImageURL: &openrouter.ImageURL{URL: imageURL, Detail: "high"}},
                },
            },
        },
        MaxTokens:   4096,
        Temperature: 0.1, // Низкая temperature для точности
    }
    
    resp, err := s.client.ChatCompletion(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("openrouter request: %w", err)
    }
    
    if len(resp.Choices) == 0 {
        return nil, fmt.Errorf("no response from AI")
    }
    
    // Парсим JSON из ответа
    content := resp.Choices[0].Message.Content.(string)
    
    // Извлекаем JSON из markdown если нужно
    jsonStr := extractJSON(content)
    
    var result RecognitionResult
    if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
        return nil, fmt.Errorf("parse AI response: %w", err)
    }
    
    result.Metadata = &RecognitionMetadata{
        Model:           resp.Model,
        ProcessingTimeMs: 0, // Заполняется вызывающим кодом
        RecognizedAt:    time.Now(),
    }
    
    return &result, nil
}

// extractJSON извлекает JSON из markdown code block.
func extractJSON(content string) string {
    // Ищем ```json ... ``` блок
    re := regexp.MustCompile("```json\\s*([\\s\\S]*?)\\s*```")
    matches := re.FindStringSubmatch(content)
    if len(matches) > 1 {
        return matches[1]
    }
    return content
}
```

## Сервис генерации вариантов

```go
// internal/service/ai/generation.go

// GenerationService сервис генерации вариантов планировки.
type GenerationService struct {
    client       *openrouter.Client
    sceneRepo    repository.SceneRepository
    branchRepo   repository.BranchRepository
    compliance   *ComplianceService
    logger       *zap.Logger
}

// GenerateVariants генерирует варианты планировки.
func (s *GenerationService) GenerateVariants(
    ctx context.Context,
    sceneID string,
    branchID *string,
    prompt string,
    count int,
    constraints *GenerationConstraints,
) ([]*GeneratedVariant, error) {
    // Получаем текущее состояние сцены
    scene, err := s.sceneRepo.GetByID(ctx, sceneID)
    if err != nil {
        return nil, fmt.Errorf("get scene: %w", err)
    }
    
    // Если указана ветка, применяем её изменения
    var currentState *SceneSnapshot
    if branchID != nil {
        branch, err := s.branchRepo.GetByID(ctx, *branchID)
        if err != nil {
            return nil, fmt.Errorf("get branch: %w", err)
        }
        currentState = branch.Snapshot
    } else {
        currentState = &SceneSnapshot{
            Elements: scene.Elements,
            Bounds:   scene.Bounds,
        }
    }
    
    // Сериализуем текущее состояние
    stateJSON, _ := json.Marshal(currentState)
    
    // Формируем промпт
    systemPrompt := `Ты - профессиональный архитектор-дизайнер интерьеров.
Твоя задача - предложить варианты перепланировки квартиры на основе запроса пользователя.

Важные правила:
1. НИКОГДА не трогай несущие стены (is_load_bearing: true)
2. Мокрые зоны (кухня, ванная) нельзя переносить над жилыми комнатами соседей
3. Минимальная площадь кухни - 5 м², жилой комнаты - 9 м²
4. Сохраняй доступ к вентиляционным каналам и стоякам

Для каждого варианта верни JSON:
{
  "variants": [
    {
      "title": "Название варианта",
      "description": "Описание изменений",
      "reasoning": "Почему этот вариант хорош",
      "delta": {
        "added": { "walls": [], "rooms": [], "furniture": [] },
        "modified": { "element_id": { изменения } },
        "removed": ["element_id"]
      },
      "estimated_cost": "low|medium|high",
      "compliance_notes": "Заметки о соответствии нормам"
    }
  ]
}`

    constraintsStr := ""
    if constraints != nil {
        if constraints.PreserveLoadBearing {
            constraintsStr += "\n- Обязательно сохранить все несущие стены"
        }
        if constraints.PreserveUtilities {
            constraintsStr += "\n- Не перемещать инженерные коммуникации"
        }
        if len(constraints.PreserveRooms) > 0 {
            constraintsStr += fmt.Sprintf("\n- Не изменять комнаты: %v", constraints.PreserveRooms)
        }
        if constraints.Style != "" {
            constraintsStr += fmt.Sprintf("\n- Стиль: %s", constraints.Style)
        }
    }
    
    userPrompt := fmt.Sprintf(`Текущая планировка:
%s

Запрос пользователя: %s

Дополнительные ограничения:%s

Предложи %d варианта перепланировки.`, stateJSON, prompt, constraintsStr, count)

    req := &openrouter.ChatCompletionRequest{
        Model: s.client.DefaultModel(),
        Messages: []openrouter.Message{
            {Role: "system", Content: systemPrompt},
            {Role: "user", Content: userPrompt},
        },
        MaxTokens:   8192,
        Temperature: 0.7, // Выше для креативности
    }
    
    resp, err := s.client.ChatCompletion(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("openrouter request: %w", err)
    }
    
    // Парсим ответ
    content := resp.Choices[0].Message.Content.(string)
    jsonStr := extractJSON(content)
    
    var aiResponse struct {
        Variants []struct {
            Title           string          `json:"title"`
            Description     string          `json:"description"`
            Reasoning       string          `json:"reasoning"`
            Delta           json.RawMessage `json:"delta"`
            EstimatedCost   string          `json:"estimated_cost"`
            ComplianceNotes string          `json:"compliance_notes"`
        } `json:"variants"`
    }
    
    if err := json.Unmarshal([]byte(jsonStr), &aiResponse); err != nil {
        return nil, fmt.Errorf("parse AI response: %w", err)
    }
    
    // Создаём ветки для каждого варианта
    variants := make([]*GeneratedVariant, 0, len(aiResponse.Variants))
    
    for i, v := range aiResponse.Variants {
        // Создаём ветку
        branch := &entity.Branch{
            SceneID:        sceneID,
            ParentBranchID: branchID,
            Name:           v.Title,
            Description:    v.Description,
            Source:         "ai",
            Order:          i,
            AIContext: &entity.AIContext{
                Prompt:      prompt,
                Model:       resp.Model,
                GeneratedAt: time.Now(),
                Reasoning:   v.Reasoning,
            },
        }
        
        // Парсим delta
        if err := json.Unmarshal(v.Delta, &branch.Delta); err != nil {
            s.logger.Warn("Failed to parse delta", zap.Error(err))
            continue
        }
        
        // Вычисляем snapshot
        branch.Snapshot = s.applyDelta(currentState, branch.Delta)
        
        // Проверяем compliance
        complianceResult, _ := s.compliance.Check(ctx, branch.Snapshot)
        branch.ComplianceResult = complianceResult
        
        // Сохраняем ветку
        if err := s.branchRepo.Create(ctx, branch); err != nil {
            s.logger.Error("Failed to save branch", zap.Error(err))
            continue
        }
        
        variants = append(variants, &GeneratedVariant{
            BranchID:    branch.ID,
            Title:       v.Title,
            Description: v.Description,
            IsCompliant: complianceResult.IsCompliant,
        })
    }
    
    return variants, nil
}
```

## Сервис чата

```go
// internal/service/ai/chat.go

// ChatService сервис чата с AI.
type ChatService struct {
    client      *openrouter.Client
    contextRepo repository.AIContextRepository
    messageRepo repository.ChatMessageRepository
    sceneRepo   repository.SceneRepository
    logger      *zap.Logger
}

// SendMessage обрабатывает сообщение пользователя.
func (s *ChatService) SendMessage(
    ctx context.Context,
    sceneID string,
    branchID *string,
    userMessage string,
) (*ChatResponse, error) {
    // Получаем или создаём контекст
    aiContext, err := s.getOrCreateContext(ctx, sceneID)
    if err != nil {
        return nil, err
    }
    
    // Сохраняем сообщение пользователя
    userMsg := &entity.ChatMessage{
        SceneID:     sceneID,
        BranchID:    branchID,
        Role:        "user",
        Content:     userMessage,
        MessageType: "text",
        UserID:      getUserIDFromContext(ctx),
    }
    if err := s.messageRepo.Create(ctx, userMsg); err != nil {
        return nil, err
    }
    
    // Формируем запрос с историей
    messages := []openrouter.Message{
        {Role: "system", Content: aiContext.SystemPrompt},
    }
    
    // Добавляем историю (последние N сообщений)
    for _, msg := range aiContext.MessageHistory {
        messages = append(messages, openrouter.Message{
            Role:    msg.Role,
            Content: msg.Content,
        })
    }
    
    // Добавляем новое сообщение
    messages = append(messages, openrouter.Message{
        Role:    "user",
        Content: userMessage,
    })
    
    req := &openrouter.ChatCompletionRequest{
        Model:       s.client.DefaultModel(),
        Messages:    messages,
        MaxTokens:   2048,
        Temperature: 0.7,
    }
    
    resp, err := s.client.ChatCompletion(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("openrouter request: %w", err)
    }
    
    assistantContent := resp.Choices[0].Message.Content.(string)
    
    // Сохраняем ответ ассистента
    assistantMsg := &entity.ChatMessage{
        SceneID:     sceneID,
        BranchID:    branchID,
        Role:        "assistant",
        Content:     assistantContent,
        MessageType: "text",
        AIMetadata: &entity.AIMetadata{
            Model:            resp.Model,
            PromptTokens:     resp.Usage.PromptTokens,
            CompletionTokens: resp.Usage.CompletionTokens,
            TotalTokens:      resp.Usage.TotalTokens,
        },
    }
    if err := s.messageRepo.Create(ctx, assistantMsg); err != nil {
        return nil, err
    }
    
    // Обновляем контекст
    aiContext.MessageHistory = append(aiContext.MessageHistory,
        HistoryMessage{Role: "user", Content: userMessage},
        HistoryMessage{Role: "assistant", Content: assistantContent},
    )
    
    // Обрезаем историю если слишком длинная
    if len(aiContext.MessageHistory) > 20 {
        aiContext.MessageHistory = aiContext.MessageHistory[len(aiContext.MessageHistory)-20:]
    }
    
    s.contextRepo.Update(ctx, aiContext)
    
    return &ChatResponse{
        UserMessage:      userMsg,
        AssistantMessage: assistantMsg,
    }, nil
}
```

## Метрики

```go
// internal/integration/openrouter/metrics.go

// Metrics метрики OpenRouter.
type Metrics struct {
    RequestDuration *prometheus.HistogramVec
    RequestErrors   *prometheus.CounterVec
    TokensUsed      *prometheus.CounterVec
}

func NewMetrics() *Metrics {
    return &Metrics{
        RequestDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "granula_openrouter_request_duration_seconds",
                Help:    "OpenRouter request duration",
                Buckets: []float64{0.5, 1, 2, 5, 10, 30, 60, 120},
            },
            []string{"type"},
        ),
        RequestErrors: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "granula_openrouter_errors_total",
                Help: "OpenRouter request errors",
            },
            []string{"type"},
        ),
        TokensUsed: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "granula_openrouter_tokens_total",
                Help: "OpenRouter tokens used",
            },
            []string{"type"},
        ),
    }
}
```


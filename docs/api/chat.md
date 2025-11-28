# API чата с AI

## Обзор

Чат с AI позволяет пользователю взаимодействовать с нейросетью для:
- Получения рекомендаций по планировке
- Генерации вариантов дизайна
- Проверки соответствия нормам
- Ответов на вопросы о ремонте

## Endpoints

### POST /api/v1/scenes/:sceneId/chat/messages

Отправка сообщения в чат.

**Request:**

```http
POST /api/v1/scenes/sc_770e8400/chat/messages
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "content": "Хочу объединить кухню с гостиной. Это законно?",
  "branch_id": "br_880e8400"
}
```

**Request Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `content` | string | Yes | Текст сообщения |
| `branch_id` | string | No | ID ветки для контекста |

**Response 200:**

```json
{
  "data": {
    "user_message": {
      "id": "msg_001",
      "role": "user",
      "content": "Хочу объединить кухню с гостиной. Это законно?",
      "message_type": "text",
      "branch_id": "br_880e8400",
      "created_at": "2024-01-21T12:00:00Z"
    },
    "assistant_message": {
      "id": "msg_002",
      "role": "assistant",
      "content": "Объединение кухни с гостиной возможно, но есть важные ограничения:\n\n1. **Газовая плита**: Если на кухне установлена газовая плита, объединение с жилой комнатой запрещено по нормам безопасности. Потребуется либо установить раздвижную перегородку, либо заменить плиту на электрическую.\n\n2. **Несущие стены**: Проверьте, не является ли перегородка между помещениями несущей. На вашем плане стена между кухней и гостиной — ненесущая перегородка, её можно демонтировать.\n\n3. **Вентиляция**: Кухня должна сохранить доступ к вентиляционному каналу.\n\nХотите, чтобы я сгенерировал варианты объединения для вашей планировки?",
      "message_type": "text",
      "suggestions": null,
      "ai_metadata": {
        "model": "anthropic/claude-sonnet-4",
        "prompt_tokens": 1520,
        "completion_tokens": 280,
        "total_tokens": 1800,
        "latency_ms": 2340
      },
      "created_at": "2024-01-21T12:00:02Z"
    }
  },
  "request_id": "req_abc123"
}
```

---

### POST /api/v1/scenes/:sceneId/chat/generate

Запрос генерации вариантов планировки.

**Request:**

```http
POST /api/v1/scenes/sc_770e8400/chat/generate
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "prompt": "Предложи 3 варианта объединения кухни и гостиной",
  "branch_id": "br_880e8400",
  "variants_count": 3,
  "constraints": {
    "preserve_load_bearing": true,
    "preserve_utilities": true,
    "max_budget": "medium"
  }
}
```

**Request Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `prompt` | string | Yes | Описание желаемого результата |
| `branch_id` | string | No | Базовая ветка |
| `variants_count` | int | No | Количество вариантов (1-5, default: 3) |
| `constraints` | object | No | Ограничения генерации |

**Constraints:**

| Field | Type | Description |
|-------|------|-------------|
| `preserve_load_bearing` | bool | Не трогать несущие стены |
| `preserve_utilities` | bool | Не перемещать коммуникации |
| `preserve_rooms` | []string | Комнаты, которые нельзя менять |
| `max_budget` | string | Бюджет: low, medium, high |
| `style` | string | Стиль: modern, classic, minimalist |

**Response 200:**

```json
{
  "data": {
    "generation_id": "gen_990e8400",
    "status": "completed",
    "user_message": {
      "id": "msg_003",
      "role": "user",
      "content": "Предложи 3 варианта объединения кухни и гостиной",
      "message_type": "text",
      "created_at": "2024-01-21T12:05:00Z"
    },
    "assistant_message": {
      "id": "msg_004",
      "role": "assistant",
      "content": "Я подготовил 3 варианта объединения кухни и гостиной для вашей квартиры. Все варианты соответствуют строительным нормам:\n\n**Вариант 1: Полное объединение**\nСнос ненесущей перегородки, создание единого пространства. Зонирование за счёт разного напольного покрытия.\n\n**Вариант 2: С барной стойкой**\nЧастичный снос перегородки с установкой барной стойки на границе зон.\n\n**Вариант 3: Широкий проём с раздвижными дверями**\nПроём 2м с раздвижными стеклянными дверями для возможности изоляции зон.\n\nВыберите вариант для детального просмотра:",
      "message_type": "suggestion",
      "suggestions": [
        {
          "index": 0,
          "branch_id": "br_gen_001",
          "title": "Вариант 1: Полное объединение",
          "description": "Снос ненесущей перегородки, создание единого пространства 28м²",
          "preview_url": "https://storage.granula.ru/previews/br_gen_001.png",
          "is_compliant": true,
          "stats": {
            "rooms_count": 3,
            "area_change": 0
          }
        },
        {
          "index": 1,
          "branch_id": "br_gen_002",
          "title": "Вариант 2: С барной стойкой",
          "description": "Частичное объединение с барной стойкой",
          "preview_url": "https://storage.granula.ru/previews/br_gen_002.png",
          "is_compliant": true,
          "stats": {
            "rooms_count": 3,
            "area_change": 0
          }
        },
        {
          "index": 2,
          "branch_id": "br_gen_003",
          "title": "Вариант 3: Проём с раздвижными дверями",
          "description": "Широкий проём 2м с возможностью изоляции",
          "preview_url": "https://storage.granula.ru/previews/br_gen_003.png",
          "is_compliant": true,
          "stats": {
            "rooms_count": 4,
            "area_change": 0
          }
        }
      ],
      "ai_metadata": {
        "model": "anthropic/claude-sonnet-4",
        "prompt_tokens": 3200,
        "completion_tokens": 1450,
        "total_tokens": 4650,
        "latency_ms": 8540
      },
      "created_at": "2024-01-21T12:05:08Z"
    }
  },
  "request_id": "req_abc123"
}
```

---

### POST /api/v1/scenes/:sceneId/chat/messages/:messageId/select

Выбор варианта из предложенных.

**Request:**

```http
POST /api/v1/scenes/sc_770e8400/chat/messages/msg_004/select
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "suggestion_index": 1
}
```

**Response 200:**

```json
{
  "data": {
    "selected_branch_id": "br_gen_002",
    "branch_activated": true,
    "message": {
      "id": "msg_005",
      "role": "assistant",
      "content": "Отлично! Я активировал \"Вариант 2: С барной стойкой\". Теперь вы можете:\n\n- Редактировать планировку в 3D редакторе\n- Попросить меня внести дополнительные изменения\n- Создать новые варианты на основе этого\n\nЧто хотите сделать дальше?",
      "message_type": "text",
      "created_at": "2024-01-21T12:06:00Z"
    }
  },
  "request_id": "req_abc123"
}
```

---

### GET /api/v1/scenes/:sceneId/chat/messages

История сообщений чата.

**Request:**

```http
GET /api/v1/scenes/sc_770e8400/chat/messages?limit=50&before=msg_010
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | int | 50 | Количество сообщений (max 100) |
| `before` | string | - | Cursor для пагинации (id сообщения) |
| `after` | string | - | Cursor для новых сообщений |
| `branch_id` | string | - | Фильтр по ветке |

**Response 200:**

```json
{
  "data": {
    "messages": [
      {
        "id": "msg_001",
        "role": "user",
        "content": "Хочу объединить кухню с гостиной. Это законно?",
        "message_type": "text",
        "branch_id": "br_880e8400",
        "created_at": "2024-01-21T12:00:00Z",
        "user_id": "550e8400"
      },
      {
        "id": "msg_002",
        "role": "assistant",
        "content": "Объединение кухни с гостиной возможно...",
        "message_type": "text",
        "suggestions": null,
        "created_at": "2024-01-21T12:00:02Z"
      }
    ],
    "has_more": true,
    "next_cursor": "msg_001"
  },
  "request_id": "req_abc123"
}
```

---

### DELETE /api/v1/scenes/:sceneId/chat/messages

Очистка истории чата.

**Request:**

```http
DELETE /api/v1/scenes/sc_770e8400/chat/messages
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "deleted_count": 15,
    "message": "Chat history cleared"
  },
  "request_id": "req_abc123"
}
```

---

### POST /api/v1/scenes/:sceneId/chat/context/reset

Сброс контекста AI (начать заново).

**Request:**

```http
POST /api/v1/scenes/sc_770e8400/chat/context/reset
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "context_reset": true,
    "message": "AI context has been reset"
  },
  "request_id": "req_abc123"
}
```

---

## WebSocket: Streaming ответов

```javascript
const ws = new WebSocket('wss://api.granula.ru/ws');

// Аутентификация
ws.send(JSON.stringify({
  type: 'auth',
  token: 'access_token'
}));

// Подписка на чат сцены
ws.send(JSON.stringify({
  type: 'subscribe',
  channel: 'chat:sc_770e8400'
}));

// Отправка сообщения через WebSocket (streaming)
ws.send(JSON.stringify({
  type: 'chat:message',
  scene_id: 'sc_770e8400',
  content: 'Как оптимизировать пространство в спальне?',
  branch_id: 'br_880e8400'
}));

// Получение streaming ответа
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  
  switch (data.type) {
    case 'chat:stream:start':
      // Начало генерации ответа
      // { message_id: 'msg_006' }
      break;
      
    case 'chat:stream:token':
      // Очередной токен ответа
      // { message_id: 'msg_006', token: 'Для' }
      appendToResponse(data.token);
      break;
      
    case 'chat:stream:end':
      // Конец генерации
      // { message_id: 'msg_006', full_content: '...' }
      break;
      
    case 'chat:generation:progress':
      // Прогресс генерации вариантов
      // { generation_id: 'gen_001', progress: 65, current_variant: 2 }
      break;
      
    case 'chat:generation:complete':
      // Генерация завершена
      // { generation_id: 'gen_001', suggestions: [...] }
      break;
      
    case 'chat:error':
      // Ошибка
      // { code: 'RATE_LIMIT', message: '...' }
      break;
  }
};
```

---

## DTO Types

```go
// internal/dto/chat.go

// SendMessageInput данные для отправки сообщения.
type SendMessageInput struct {
    // Текст сообщения
    Content string `json:"content" validate:"required,min=1,max=4000"`
    
    // ID ветки для контекста (опционально)
    BranchID *string `json:"branch_id,omitempty"`
}

// GenerateVariantsInput данные для генерации вариантов.
type GenerateVariantsInput struct {
    // Описание желаемого результата
    Prompt string `json:"prompt" validate:"required,min=10,max=2000"`
    
    // Базовая ветка
    BranchID *string `json:"branch_id,omitempty"`
    
    // Количество вариантов
    VariantsCount int `json:"variants_count,omitempty" validate:"omitempty,gte=1,lte=5"`
    
    // Ограничения
    Constraints *GenerationConstraints `json:"constraints,omitempty"`
}

// GenerationConstraints ограничения генерации.
type GenerationConstraints struct {
    PreserveLoadBearing bool     `json:"preserve_load_bearing"`
    PreserveUtilities   bool     `json:"preserve_utilities"`
    PreserveRooms       []string `json:"preserve_rooms,omitempty"`
    MaxBudget           string   `json:"max_budget,omitempty"` // low, medium, high
    Style               string   `json:"style,omitempty"`      // modern, classic, minimalist
}

// SelectSuggestionInput выбор варианта.
type SelectSuggestionInput struct {
    SuggestionIndex int `json:"suggestion_index" validate:"gte=0,lte=4"`
}

// ChatMessageResponse сообщение чата.
type ChatMessageResponse struct {
    ID          string                `json:"id"`
    Role        string                `json:"role"` // user, assistant, system
    Content     string                `json:"content"`
    MessageType string                `json:"message_type"` // text, suggestion, action, error
    BranchID    *string               `json:"branch_id,omitempty"`
    Suggestions []SuggestionResponse  `json:"suggestions,omitempty"`
    AIMetadata  *AIMetadataResponse   `json:"ai_metadata,omitempty"`
    SceneContext *SceneContextResponse `json:"scene_context,omitempty"`
    CreatedAt   time.Time             `json:"created_at"`
    UserID      *string               `json:"user_id,omitempty"`
}

// SuggestionResponse вариант планировки.
type SuggestionResponse struct {
    Index       int               `json:"index"`
    BranchID    string            `json:"branch_id"`
    Title       string            `json:"title"`
    Description string            `json:"description"`
    PreviewURL  string            `json:"preview_url"`
    IsCompliant bool              `json:"is_compliant"`
    Stats       *SuggestionStats  `json:"stats,omitempty"`
}

// SuggestionStats статистика варианта.
type SuggestionStats struct {
    RoomsCount  int     `json:"rooms_count"`
    AreaChange  float64 `json:"area_change"`
}

// AIMetadataResponse метаданные AI запроса.
type AIMetadataResponse struct {
    Model            string `json:"model"`
    PromptTokens     int    `json:"prompt_tokens"`
    CompletionTokens int    `json:"completion_tokens"`
    TotalTokens      int    `json:"total_tokens"`
    LatencyMs        int64  `json:"latency_ms"`
}

// SceneContextResponse контекст сцены.
type SceneContextResponse struct {
    SnapshotID     string `json:"snapshot_id"`
    ActiveBranchID string `json:"active_branch_id,omitempty"`
}

// ChatMessagesListResponse список сообщений.
type ChatMessagesListResponse struct {
    Messages   []ChatMessageResponse `json:"messages"`
    HasMore    bool                  `json:"has_more"`
    NextCursor *string               `json:"next_cursor,omitempty"`
}
```

## Rate Limits

| Action | Limit | Window |
|--------|-------|--------|
| Send message | 30 | 1 min |
| Generate variants | 10 | 1 min |
| Total AI tokens | 100k | 1 hour |

## Типы сообщений

| Type | Description |
|------|-------------|
| `text` | Обычное текстовое сообщение |
| `suggestion` | Сообщение с вариантами для выбора |
| `action` | Подтверждение выполненного действия |
| `error` | Сообщение об ошибке |


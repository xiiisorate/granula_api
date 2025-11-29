# 📊 Полный анализ API Granula

> **Дата анализа:** 29 ноября 2024  
> **Статус:** Требуется доработка перед запуском

---

## 🚨 КРИТИЧЕСКИЕ ПРОБЛЕМЫ (Блокирующие запуск)

### 1. Proto файлы не сгенерированы

**Проблема:** Папка `shared/gen/` **ПУСТАЯ** - Go код из proto файлов не сгенерирован.

**Расположение proto файлов:**
```
shared/proto/
├── ai/v1/ai.proto
├── auth/v1/auth.proto
├── branch/v1/branch.proto
├── common/v1/common.proto
├── compliance/v1/compliance.proto
├── floorplan/v1/floorplan.proto
├── notification/v1/notification.proto
├── request/v1/request.proto
├── scene/v1/scene.proto
├── user/v1/user.proto
└── workspace/v1/workspace.proto
```

**Влияние:** Все сервисы импортируют код из `shared/gen/...` и **НЕ БУДУТ КОМПИЛИРОВАТЬСЯ**.

**Решение:**
```powershell
cd shared
# Создать папки для генерации
mkdir -p gen/auth/v1 gen/user/v1 gen/workspace/v1 gen/scene/v1 gen/branch/v1 gen/ai/v1 gen/compliance/v1 gen/floorplan/v1 gen/request/v1 gen/notification/v1 gen/common/v1

# Генерация proto
protoc --proto_path=proto \
  --go_out=gen --go_opt=paths=source_relative \
  --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
  proto/common/v1/common.proto \
  proto/auth/v1/auth.proto \
  proto/user/v1/user.proto \
  proto/workspace/v1/workspace.proto \
  proto/scene/v1/scene.proto \
  proto/branch/v1/branch.proto \
  proto/ai/v1/ai.proto \
  proto/compliance/v1/compliance.proto \
  proto/floorplan/v1/floorplan.proto \
  proto/request/v1/request.proto \
  proto/notification/v1/notification.proto
```

---

### 2. Несовпадение путей пакетов в Proto файлах

**Проблема:** Proto файлы содержат неправильный `go_package`:
```protobuf
// В proto файлах:
option go_package = "github.com/granula/shared/gen/auth/v1;authv1";

// В коде используется:
import "github.com/xiiisorate/granula_api/shared/gen/auth/v1"
```

**Решение:** Обновить `go_package` во всех proto файлах:
```protobuf
option go_package = "github.com/xiiisorate/granula_api/shared/gen/auth/v1;authv1";
```

Файлы для исправления:
- [ ] `shared/proto/auth/v1/auth.proto`
- [ ] `shared/proto/user/v1/user.proto`
- [ ] `shared/proto/workspace/v1/workspace.proto`
- [ ] `shared/proto/scene/v1/scene.proto`
- [ ] `shared/proto/branch/v1/branch.proto`
- [ ] `shared/proto/ai/v1/ai.proto`
- [ ] `shared/proto/compliance/v1/compliance.proto`
- [ ] `shared/proto/floorplan/v1/floorplan.proto`
- [ ] `shared/proto/request/v1/request.proto`
- [ ] `shared/proto/notification/v1/notification.proto`
- [ ] `shared/proto/common/v1/common.proto`

---

### 3. 🤖 AI Service: Распознавание изображений НЕ РАБОТАЕТ

**Проблема:** В `ai-service/internal/service/recognition_service.go` изображения НЕ отправляются в AI модель!

```go
// Строки 88-95 — КРИТИЧЕСКАЯ ОШИБКА
// Отправляется только обрезанный текст вместо изображения:
messages := []openrouter.Message{
    {
        Role:    "user", 
        Content: prompt + "\n\n[Изображение планировки загружено: " + dataURL[:100] + "...]",
        // ^^^ Только первые 100 символов base64!
    },
}
```

**Влияние:** Ключевая функция сервиса — распознавание планировок из фото/сканов — **ПОЛНОСТЬЮ НЕ РАБОТАЕТ**.

**Решение:**
1. Добавить поддержку мультимодальных сообщений в OpenRouter клиент
2. Отправлять полное изображение в формате Vision API
3. Использовать модель с поддержкой Vision (claude-sonnet-4, gpt-4o)

**Детальный анализ AI модуля:** см. раздел "🤖 ПОЛНЫЙ АНАЛИЗ AI МОДУЛЯ" ниже.

---

## 📋 СТАТУС МИКРОСЕРВИСОВ

### ✅ Полностью реализованные сервисы

| Сервис | gRPC | Service Layer | Repository | Миграции | Dockerfile |
|--------|------|---------------|------------|----------|------------|
| auth-service | ✅ | ✅ | ✅ | ❌ | ✅ |
| workspace-service | ✅ | ✅ | ✅ | ✅ | ✅ |
| request-service | ✅ | ✅ | ✅ | ✅ | ✅ |
| compliance-service | ✅ | ✅ | ✅ | ✅ | ✅ |
| floorplan-service | ✅ | ✅ | ✅ | ✅ | ✅ |
| scene-service | ✅ | ✅ | ✅ | ❌ | ✅ |
| ai-service | ✅ | ✅ | ✅ | ❌ | ✅ |

### ⚠️ Частично реализованные сервисы

| Сервис | gRPC | Service Layer | Repository | Миграции | Dockerfile |
|--------|------|---------------|------------|----------|------------|
| user-service | ✅ | ✅ (базовый) | ✅ | ❌ | ✅ |
| notification-service | ✅ | ✅ (базовый) | ✅ | ❌ | ✅ |
| branch-service | ✅ | ⚠️ (TODO) | ✅ | ❌ | ✅ |

---

## 🔧 НЕЗАВЕРШЁННАЯ ФУНКЦИОНАЛЬНОСТЬ ПО СЕРВИСАМ

### API Gateway (`api-gateway/`)

**Реализованные handlers:**
- ✅ `auth.go` - Аутентификация
- ✅ `user_handler.go` - Профиль пользователя
- ✅ `notification_handler.go` - Уведомления  
- ✅ `workspace.go` - Воркспейсы
- ✅ `scene.go` - Сцены
- ✅ `ai.go` - AI функции

**Placeholder handlers (НЕ РЕАЛИЗОВАНЫ):**
```go
// В api-gateway/cmd/main.go строки 244-291
// Используются placeholderHandler() вместо реальных обработчиков:

❌ GET/POST/PATCH/DELETE /floor-plans/* - все endpoints
❌ GET/POST/PATCH/DELETE /scenes/:scene_id/branches/* - все endpoints  
❌ POST /compliance/check
❌ POST /compliance/check-operation
❌ GET /compliance/rules
❌ GET /compliance/rules/:id
❌ GET/POST/PATCH/DELETE /requests/* - все endpoints
```

**Необходимо создать:**
- [ ] `handlers/floorplan.go` - FloorPlan HTTP handlers
- [ ] `handlers/branch.go` - Branch HTTP handlers
- [ ] `handlers/compliance.go` - Compliance HTTP handlers
- [ ] `handlers/request.go` - Expert Request HTTP handlers

---

### Auth Service (`auth-service/`)

**Реализовано:**
- ✅ Register (регистрация)
- ✅ Login (вход)
- ✅ ValidateToken (валидация JWT)
- ✅ RefreshToken (обновление токена)
- ✅ Logout (выход)
- ✅ LogoutAll (выход из всех устройств)
- ✅ ChangePassword (смена пароля)

**Не реализовано:**
- [ ] OAuth 2.0 (Google, Yandex) - описан в документации
- [ ] Password Reset (сброс пароля по email)
- [ ] Email Verification (подтверждение email)
- [ ] Миграции БД

---

### User Service (`user-service/`)

**Реализовано:**
- ✅ CreateProfile
- ✅ GetProfile
- ✅ UpdateProfile
- ✅ ChangePassword (валидация)
- ✅ DeleteAccount (soft delete)
- ✅ UpdateAvatar / DeleteAvatar

**Не реализовано:**
- [ ] GetSessions - получение активных сессий
- [ ] RevokeSession - отзыв сессии
- [ ] Admin endpoints (управление пользователями)
- [ ] Миграции БД

---

### Notification Service (`notification-service/`)

**Реализовано:**
- ✅ Create - создание уведомления
- ✅ GetList - список уведомлений
- ✅ GetUnreadCount - количество непрочитанных
- ✅ MarkAsRead - пометить как прочитанное
- ✅ MarkAllAsRead - пометить все
- ✅ Delete - удаление
- ✅ DeleteAllRead - удалить прочитанные

**Не реализовано:**
- [ ] WebSocket для real-time уведомлений
- [ ] Push notifications (FCM/APNs)
- [ ] Email notifications
- [ ] UpdateSettings - настройки уведомлений
- [ ] Миграции БД

---

### Branch Service (`branch-service/`)

**Реализовано:**
- ✅ CreateBranch
- ✅ GetBranch
- ✅ ListBranches
- ✅ DeleteBranch

**Частично реализовано (TODO в коде):**
```go
// branch-service/internal/service/branch_service.go

// Строка 37: TODO: Copy elements from parent branch if parentID is set
func (s *BranchService) CreateBranch(...) {
    // TODO: Copy elements from parent branch if parentID is set
}

// Строка 75: TODO: Implement actual merge logic with conflict detection
func (s *BranchService) MergeBranch(...) {
    // TODO: Implement actual merge logic with conflict detection
}

// Строка 87: TODO: Implement diff logic
func (s *BranchService) GetDiff(...) {
    // TODO: Implement diff logic
}

// Строка 98: TODO: Serialize current elements
func (s *BranchService) CreateSnapshot(...) {
    // TODO: Serialize current elements
}

// Строка 115: TODO: Restore elements from snapshot data
func (s *BranchService) RestoreSnapshot(...) {
    // TODO: Restore elements from snapshot data
}
```

**Не реализовано:**
- [ ] Merge logic с детекцией конфликтов
- [ ] Diff между ветками
- [ ] Serialization/Restore snapshots
- [ ] Копирование элементов из parent branch
- [ ] Миграции БД

---

### Scene Service (`scene-service/`)

**Реализовано:**
- ✅ CreateScene
- ✅ GetScene
- ✅ UpdateScene  
- ✅ ListScenes
- ✅ CreateElement
- ✅ GetElement
- ✅ UpdateElement
- ✅ DeleteElement
- ✅ ListElements
- ✅ CheckCompliance (через Compliance Service)

**Не реализовано:**
```go
// scene-service/internal/service/scene_service.go строка 92
func (s *SceneService) DeleteScene(...) error {
    // TODO: Delete all elements, branches, etc.
    return s.sceneRepo.Delete(ctx, id)
}
```

- [ ] Каскадное удаление элементов и веток при удалении сцены
- [ ] DuplicateScene - дублирование сцены
- [ ] RequestRender - запрос рендера
- [ ] WebSocket для real-time updates
- [ ] Миграции БД

---

### AI Service (`ai-service/`) — ДЕТАЛЬНЫЙ АНАЛИЗ

---

## 🤖 ПОЛНЫЙ АНАЛИЗ AI МОДУЛЯ

### Структура AI Service

```
ai-service/internal/
├── config/config.go
├── domain/entity/
│   ├── chat.go          # ChatMessage, SuggestedAction, TokenUsage
│   ├── generation.go    # GenerationJob, GeneratedVariant, VariantChange
│   └── recognition.go   # RecognitionJob, RecognitionResult, RecognizedWall/Room/Opening
├── grpc/server.go       # gRPC handlers
├── openrouter/client.go # OpenRouter API клиент
├── prompts/prompts.go   # System prompts (810+ строк!)
├── repository/mongodb/
│   ├── chat_repository.go
│   └── job_repository.go
└── service/
    ├── chat_service.go        # Чат с AI
    ├── generation_service.go  # Генерация вариантов
    └── recognition_service.go # Распознавание планов
```

---

### ✅ ПРОМПТЫ — КАЧЕСТВЕННО РЕАЛИЗОВАНЫ

**Файл:** `ai-service/internal/prompts/prompts.go` (810 строк)

#### 1. RecognitionSystemPrompt — Распознавание планировок

**Содержит:**
- Стандарты ГОСТ 21.501-2018 (Архитектурные чертежи)
- Стандарты ГОСТ 21.205-93 (Инженерные коммуникации)  
- Таблицы условных обозначений:
  - Линии стен и перегородок (несущие/ненесущие)
  - Проёмы и двери (распашные, раздвижные, двупольные)
  - Окна (одинарные, двойные, с форточкой)
  - Лестницы
  - Сантехника (умывальник, унитаз, ванна, душ, раковина)
  - Трубопроводы и стояки (канализация, водоснабжение, отопление, вентиляция)
  - Электрика (выключатели, розетки, светильники)
  - Кухонное оборудование
- Типы помещений с минимальными площадями
- Признаки несущих/ненесущих конструкций
- Правила определения масштаба

**JSON схема вывода (полная и корректная):**
```json
{
  "dimensions": { "width": <м>, "height": <м> },
  "total_area": <м²>,
  "detected_scale": "1:100",
  "walls": [{ "temp_id", "start", "end", "thickness", "is_load_bearing", "material", "confidence" }],
  "rooms": [{ "temp_id", "type", "boundary", "area", "is_wet_zone", "has_window", "wall_ids", "confidence" }],
  "openings": [{ "temp_id", "type", "subtype", "position", "width", "height", "wall_id", "opens_to", "confidence" }],
  "utilities": [{ "temp_id", "type", "position", "can_relocate", "protection_zone", "room_id", "confidence" }],
  "equipment": [{ "temp_id", "type", "position", "dimensions", "room_id", "confidence" }],
  "metadata": { "source_type", "quality", "orientation", "has_dimensions", "has_annotations" },
  "warnings": [],
  "notes": []
}
```

#### 2. ChatSystemPrompt — AI-консультант по перепланировке

**Содержит:**
- Знание нормативной базы:
  - СНиП 31-01-2003
  - СП 54.13330.2016
  - Жилищный кодекс РФ (ст. 25-29)
  - ПП Москвы №508-ПП
  - СанПиН, ФЗ-123 (пожарная безопасность)
- **Абсолютные запреты перепланировки** (8 пунктов)
- **Разрешённые перепланировки** (без согласования, уведомительный порядок, с проектом)
- **Минимальные площади/размеры помещений**
- **Типы действий** для JSON вывода:
  - `DEMOLISH_WALL`, `ADD_WALL`, `ADD_OPENING`, `CLOSE_OPENING`
  - `MERGE_ROOMS`, `SPLIT_ROOM`, `MOVE_WET_ZONE`, `CHANGE_ROOM_TYPE`
  - `ADD_FURNITURE`, `RELOCATE_KITCHEN`

**JSON для рекомендованных действий:**
```json
{
  "action": {
    "type": "<тип действия>",
    "element_id": "<id элемента>",
    "description": "<описание>",
    "requires_approval": true/false,
    "approval_type": "none|notification|project|expertise",
    "estimated_cost": "<стоимость>",
    "risks": ["<риски>"]
  }
}
```

#### 3. GenerationSystemPrompt — Генерация вариантов

**Содержит:**
- Стили генерации: MINIMAL, MODERATE, CREATIVE
- Ориентировочные цены на работы (Москва 2024)
- Коэффициенты по регионам
- Минимальные требования к помещениям

**JSON схема вывода (полная):**
```json
{
  "analysis": { "current_layout_summary", "user_request_interpretation", "constraints_identified", "opportunities" },
  "variants": [{
    "id", "name", "description", "style", "score",
    "scores_breakdown": { "functionality", "aesthetics", "compliance", "cost_efficiency" },
    "changes": [{ "type", "description", "element_ids", "impact", "requires_reinforcement" }],
    "new_layout": { "rooms": [...], "removed_walls": [...], "added_walls": [...] },
    "compliance": { "is_compliant", "violations", "warnings", "approval_type", "approval_difficulty" },
    "cost_estimate": { "works", "materials", "approval", "total", "currency", "confidence" },
    "timeline": { "works_days", "approval_months", "total_weeks" },
    "pros", "cons", "recommendations"
  }],
  "comparison": { "best_for_budget", "best_for_space", "best_for_quick_approval", "recommended", "recommendation_reason" }
}
```

#### 4. ComplianceCheckPrompt — Проверка соответствия нормам

**Категории проверок:** structural, plumbing, ventilation, gas, fire_safety, general

---

### 🚨 КРИТИЧЕСКИЕ ПРОБЛЕМЫ AI МОДУЛЯ

#### 1. РАСПОЗНАВАНИЕ ИЗОБРАЖЕНИЙ НЕ РАБОТАЕТ! ❌

**Проблема в файле:** `ai-service/internal/service/recognition_service.go`

```go
// Строки 88-95 — КРИТИЧЕСКАЯ ОШИБКА!
// For now, we'll use a text description since Claude doesn't support images via this API directly
// In production, you would use a vision model or separate image analysis service
messages := []openrouter.Message{
    {
        Role:    "user",
        Content: prompt + "\n\n[Изображение планировки загружено: " + dataURL[:100] + "...]",
    },
}
```

**Что происходит:** 
- Изображение конвертируется в base64 (строка 68-69)
- НО отправляется только первые 100 символов base64 как текст!
- AI получает: `[Изображение планировки загружено: data:image/png;base64,iVBORw0KGg...]` — обрезанный текст
- **Распознавание планировок ПОЛНОСТЬЮ НЕ РАБОТАЕТ!**

**Интересно:** В `openrouter/client.go` (строки 63-74) уже определены структуры для Vision API:
```go
type ImageContent struct {
    Type     string    `json:"type"` // "text" or "image_url"
    Text     string    `json:"text,omitempty"`
    ImageURL *ImageURL `json:"image_url,omitempty"`
}

type ImageURL struct {
    URL    string `json:"url"`
    Detail string `json:"detail,omitempty"` // "low", "high", "auto"
}
```

**НО:** Метода `ChatCompletionWithImages` нет! Структуры не используются.

**Что нужно исправить:**
1. Добавить метод `ChatCompletionWithImages` в OpenRouter клиент
2. Использовать multimodal messages с `content: [{type: "text"}, {type: "image_url"}]`
3. Изменить `recognition_service.go` для отправки реальных изображений

---

#### 2. CHAT НЕ ПОЛУЧАЕТ ДАННЫЕ СЦЕНЫ! ⚠️

**Проблема в файле:** `ai-service/internal/service/chat_service.go`

```go
// Строки 314-318
// TODO: This should fetch actual scene data from Scene Service via gRPC.
func (s *ChatService) getSceneSummary(sceneID string) string {
    return "Scene ID: " + sceneID + " (данные сцены будут загружены из Scene Service)"
}
```

**Последствия:**
- AI-чат не знает структуру текущей планировки
- Не может давать конкретные рекомендации по существующим стенам/комнатам
- Промпт ChatSystemPrompt содержит `%s` для контекста, но получает заглушку

**Что нужно:**
- Добавить gRPC клиент для Scene Service
- Загружать текущие элементы сцены (стены, комнаты, проёмы)
- Форматировать их в JSON для промпта

---

#### 3. ГЕНЕРАЦИЯ ВАРИАНТОВ НЕ ПОЛУЧАЕТ ДАННЫЕ СЦЕНЫ! ⚠️

**Проблема в файле:** `ai-service/internal/grpc/server.go`

```go
// Строка 131
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
- AI не знает текущую планировку
- Генерирует абстрактные варианты без привязки к реальным данным
- Не может указать конкретные `element_ids` для изменений

---

#### 4. GetContext и UpdateContext НЕ РЕАЛИЗОВАНЫ ❌

```go
// ai-service/internal/grpc/server.go строки 296-303
func (s *AIServer) GetContext(...) {
    return nil, apperrors.Internal("not implemented").ToGRPCError()
}
func (s *AIServer) UpdateContext(...) {
    return nil, apperrors.Internal("not implemented").ToGRPCError()
}
```

---

#### 5. SelectSuggestion ИЗ ДОКУМЕНТАЦИИ НЕ РЕАЛИЗОВАН ❌

В `docs/api/chat.md` описан endpoint:
```
POST /api/v1/scenes/:sceneId/chat/messages/:messageId/select
```

**НЕ реализован** ни в AI Service, ни в API Gateway!

Этот endpoint нужен для:
- Выбора варианта из предложенных AI
- Активации ветки с выбранным вариантом
- Интеграции AI с Branch Service

---

#### 6. ВРЕМЯ ГЕНЕРАЦИИ НЕ ТРЕКАЕТСЯ ⚠️

```go
// ai-service/internal/service/chat_service.go строка 96
return &ChatResponse{
    // ...
    GenerationTimeMs: 0, // TODO: track time
}
```

---

### ✅ ЧТО РАБОТАЕТ В AI МОДУЛЕ

| Функция | gRPC | Service | Работает? | Проблема |
|---------|------|---------|-----------|----------|
| SendChatMessage | ✅ | ✅ | ⚠️ | Нет данных сцены |
| StreamChatResponse | ✅ | ✅ | ⚠️ | Нет данных сцены |
| GetChatHistory | ✅ | ✅ | ✅ | - |
| ClearChatHistory | ✅ | ✅ | ✅ | - |
| RecognizeFloorPlan | ✅ | ✅ | ❌ | **Изображения не отправляются!** |
| GetRecognitionStatus | ✅ | ✅ | ⚠️ | Зависит от Recognition |
| GenerateVariants | ✅ | ✅ | ⚠️ | Нет данных сцены |
| GetGenerationStatus | ✅ | ✅ | ⚠️ | Зависит от Generation |
| GetContext | ✅ | ❌ | ❌ | Не реализован |
| UpdateContext | ✅ | ❌ | ❌ | Не реализован |
| SelectSuggestion | ❌ | ❌ | ❌ | Не реализован (из документации) |

---

### ✅ ПОЗИТИВНЫЕ МОМЕНТЫ AI

1. **Промпты качественные и детальные** — 810 строк профессионального контента
2. **Entities правильно определены** — Chat, Generation, Recognition
3. **Job-based async processing** — фоновая обработка с прогрессом
4. **OpenRouter клиент имеет:**
   - Rate limiting
   - Retries с exponential backoff
   - Streaming support (SSE parsing)
5. **parseActions** — парсит JSON из ответа AI
6. **История чата сохраняется** в MongoDB

---

### 🔧 ПЛАН ИСПРАВЛЕНИЯ AI МОДУЛЯ

#### Приоритет 1 (Критично — без этого AI не работает):

1. **Добавить поддержку Vision API в OpenRouter клиент**
   ```go
   // openrouter/client.go
   type MultimodalMessage struct {
       Role    string           `json:"role"`
       Content []ContentPart    `json:"content"`
   }
   
   type ContentPart struct {
       Type     string    `json:"type"` // "text" или "image_url"
       Text     string    `json:"text,omitempty"`
       ImageURL *ImageURL `json:"image_url,omitempty"`
   }
   
   func (c *Client) ChatCompletionWithImages(ctx context.Context, messages []MultimodalMessage, opts ChatOptions) (*ChatResponse, error)
   ```

2. **Исправить RecognitionService для реальной отправки изображений**
   - Использовать multimodal messages
   - Использовать модель с Vision (claude-sonnet-4 или gpt-4o)

3. **Интегрировать AI Service с Scene Service**
   - Добавить gRPC клиент для Scene Service
   - Загружать данные сцены перед запросом к AI
   - Форматировать в JSON для промптов

#### Приоритет 2 (Важно):

4. **Реализовать GetContext/UpdateContext**
5. **Добавить SelectSuggestion endpoint** в AI Service и API Gateway
6. **Добавить tracking времени генерации**

#### Приоритет 3 (Улучшения):

7. **WebSocket для стриминга в API Gateway**
8. **Кэширование контекста в Redis**
9. **Лимиты токенов и rate limiting по пользователям**

---

### Compliance Service (`compliance-service/`)

**Реализовано:**
- ✅ CheckCompliance - полная проверка сцены
- ✅ CheckOperation - проверка операции
- ✅ GetRules - получение правил
- ✅ GetRule - одно правило
- ✅ GetRuleByCode - правило по коду
- ✅ GetCategories - категории правил
- ✅ ValidateScene - быстрая валидация

**Rule Engine реализован для категорий:**
- ✅ LoadBearing (несущие стены)
- ✅ WetZones (мокрые зоны)
- ✅ MinArea (минимальные площади)
- ✅ Ventilation (вентиляция)
- ✅ FireSafety (пожарная безопасность)
- ✅ Daylight (естественное освещение)
- ✅ General (общие правила)

**Не реализовано:**
- [ ] GenerateReport - генерация отчёта
- [ ] Больше правил СНиП и ЖК РФ
- [ ] AI проверка сложных случаев

---

### FloorPlan Service (`floorplan-service/`)

**Реализовано:**
- ✅ Upload - загрузка плана
- ✅ Get - получение плана
- ✅ List - список планов
- ✅ Update - обновление
- ✅ Delete - удаление
- ✅ StartRecognition - запуск распознавания
- ✅ GetRecognitionStatus - статус
- ✅ GetDownloadURL - presigned URL
- ✅ MinIO storage integration

**Не реализовано:**
- [ ] Reprocess - повторная обработка
- [ ] CreateSceneFromFloorPlan - создание сцены из плана
- [ ] WebSocket для статуса обработки
- [ ] Thumbnails generation

---

### Request Service (`request-service/`)

**Реализовано:**
- ✅ CreateRequest
- ✅ GetRequest
- ✅ ListRequests
- ✅ UpdateRequest
- ✅ SubmitRequest
- ✅ CancelRequest
- ✅ UpdateStatus
- ✅ AssignExpert
- ✅ RejectRequest
- ✅ CompleteRequest
- ✅ AddDocument
- ✅ GetDocuments
- ✅ GetStatusHistory

**Не реализовано (TODO в коде):**
```go
// request-service/internal/service/request_service.go

// Строка 219: TODO: Send notification to staff
// Строка 309: TODO: Send notification to user  
// Строка 353: TODO: Send notifications to user and expert
// Строка 396: TODO: Send notification to user
// Строка 442: TODO: Send notification to user
```

- [ ] Интеграция с Notification Service
- [ ] GetRequestCost - расчёт стоимости
- [ ] ScheduleVisit - планирование визита

---

### Workspace Service (`workspace-service/`)

**Реализовано:**
- ✅ CreateWorkspace
- ✅ GetWorkspace / GetWorkspaceBasic
- ✅ ListWorkspaces
- ✅ UpdateWorkspace
- ✅ DeleteWorkspace
- ✅ AddMember
- ✅ RemoveMember
- ✅ UpdateMemberRole
- ✅ GetMembers
- ✅ InviteMember

**Не реализовано (TODO в коде):**
```go
// workspace-service/internal/service/workspace_service.go

// Строка 155: TODO: Publish workspace.created event
// Строка 345: TODO: Publish workspace.deleted event
// Строка 415: TODO: Send notification to new member
// Строка 601: TODO: Send notification to invitee
```

- [ ] Event publishing (Redis Pub/Sub)
- [ ] Интеграция с Notification Service
- [ ] AcceptInvite / DeclineInvite
- [ ] TransferOwnership

---

## 📁 ОТСУТСТВУЮЩИЕ МИГРАЦИИ БД

| Сервис | Статус миграций |
|--------|-----------------|
| auth-service | ❌ Отсутствуют |
| user-service | ❌ Отсутствуют |
| workspace-service | ✅ Присутствуют |
| request-service | ✅ Присутствуют |
| floorplan-service | ✅ Присутствуют |
| compliance-service | ✅ Присутствуют |
| scene-service | ❌ Отсутствуют (MongoDB - возможно не нужны) |
| branch-service | ❌ Отсутствуют (MongoDB - возможно не нужны) |
| ai-service | ❌ Отсутствуют (MongoDB - возможно не нужны) |
| notification-service | ❌ Отсутствуют |

---

## 🔌 ИНТЕГРАЦИИ МЕЖДУ СЕРВИСАМИ

### Реализованные интеграции
- ✅ API Gateway → All Services (gRPC)
- ✅ Scene Service → Compliance Service (проверка соответствия)
- ✅ FloorPlan Service → AI Service (распознавание)
- ✅ FloorPlan Service → MinIO (хранение файлов)

### Отсутствующие интеграции
- ❌ Request Service → Notification Service
- ❌ Workspace Service → Notification Service
- ❌ AI Service → Scene Service (получение данных сцены)
- ❌ Branch Service → Scene Service (элементы веток)
- ❌ Redis Pub/Sub для событий

---

## 📝 ДОКУМЕНТАЦИЯ API

**Swagger/OpenAPI:**
- ✅ Annotations в handlers (`@Summary`, `@Description`, `@Tags`, etc.)
- ❌ Swagger UI генерация (`swag init` не выполнена)
- ❌ `docs/swagger.yaml` не сгенерирован

---

## 🐳 ИНФРАСТРУКТУРА

### Docker
- ✅ `docker-compose.yml` - полная конфигурация
- ✅ Dockerfile для всех 11 сервисов
- ✅ Volumes для персистентности
- ✅ Health checks
- ✅ Networks

### Базы данных
- ✅ PostgreSQL containers (6 штук)
- ✅ MongoDB container
- ✅ Redis container
- ✅ MinIO container

---

## 📋 ЧЕКЛИСТ ДЛЯ ЗАПУСКА

### Критические задачи (ОБЯЗАТЕЛЬНО)

- [ ] 1. Исправить `go_package` во всех proto файлах
- [ ] 2. Сгенерировать Go код из proto (`protoc`)
- [ ] 3. Выполнить `go mod tidy` во всех сервисах
- [ ] 4. Создать миграции для auth-service, user-service, notification-service
- [ ] 5. Создать handlers в API Gateway для: floor-plans, branches, compliance, requests
- [ ] 6. Сгенерировать Swagger документацию

### 🤖 Критические AI задачи (ОБЯЗАТЕЛЬНО для AI функций)

- [ ] 7. **Добавить поддержку Vision API в OpenRouter клиент** — без этого распознавание планов НЕ РАБОТАЕТ
- [ ] 8. **Исправить recognition_service.go** — отправлять реальные изображения, а не обрезанный base64 текст
- [ ] 9. **Интегрировать AI Service → Scene Service** — получать данные сцены для чата и генерации
- [ ] 10. **Реализовать SelectSuggestion endpoint** — выбор варианта из предложенных AI (описан в документации, не реализован)

### Важные задачи (для полной функциональности)

- [ ] Реализовать merge/diff/snapshots в Branch Service
- [ ] Реализовать GetContext/UpdateContext в AI Service
- [ ] Добавить WebSocket endpoints для AI streaming
- [ ] Интегрировать Notification Service
- [ ] Реализовать OAuth 2.0
- [ ] Добавить tracking времени генерации AI

### Дополнительные улучшения

- [ ] Unit тесты
- [ ] Integration тесты
- [ ] CI/CD pipeline
- [ ] Мониторинг (Prometheus, Grafana)
- [ ] Distributed tracing (Jaeger)
- [ ] Rate limiting по токенам AI на пользователя
- [ ] Кэширование AI контекста в Redis

---

## 🎯 РЕКОМЕНДУЕМЫЙ ПОРЯДОК РАБОТЫ

### Этап 1: Базовый запуск (2-3 часа)
1. Исправить proto и сгенерировать код (30 мин)
2. Создать недостающие API Gateway handlers (1-2 часа)
3. Создать миграции (30 мин)
4. `go mod tidy` и тест компиляции (30 мин)

### Этап 2: AI функциональность (3-4 часа)
5. Добавить Vision API в OpenRouter клиент (1 час)
6. Исправить RecognitionService для работы с изображениями (1 час)
7. Интегрировать AI с Scene Service (1-2 часа)
8. Реализовать SelectSuggestion (30 мин)

### Этап 3: Интеграционное тестирование (1-2 часа)
9. Docker Compose up
10. Тест всех endpoints
11. Тест распознавания планов
12. Тест чата и генерации вариантов

---

## 📊 СВОДКА КРИТИЧЕСКИХ ПРОБЛЕМ

| # | Проблема | Влияние | Приоритет |
|---|----------|---------|-----------|
| 1 | Proto не сгенерированы | Сервисы не компилируются | 🔴 Критично |
| 2 | Пути go_package неверные | Proto не сгенерируется | 🔴 Критично |
| 3 | **AI: изображения не отправляются** | Распознавание планов не работает | 🔴 Критично |
| 4 | AI: нет данных сцены | Чат/генерация без контекста | 🟠 Высокий |
| 5 | API Gateway: placeholders | Половина API не работает | 🟠 Высокий |
| 6 | Миграции отсутствуют | auth/user/notification не стартуют | 🟠 Высокий |
| 7 | SelectSuggestion не реализован | Нельзя выбрать вариант AI | 🟡 Средний |
| 8 | Branch merge/diff не реализован | Ветки не сливаются | 🟡 Средний |

---

*Анализ выполнен автоматически на основе полной проверки кодовой базы*
*Обновлено: Детальный анализ AI модуля добавлен 29.11.2024*


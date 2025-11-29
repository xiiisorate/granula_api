// =============================================================================
// AI Service DTOs
// =============================================================================
package dto

import "time"

// =============================================================================
// AI Chat DTOs
// =============================================================================

// ChatMessageRequest represents a chat message input.
// @Description Сообщение для AI чата
type ChatMessageRequest struct {
	// ID сессии (опционально, создается автоматически)
	SessionID string `json:"session_id,omitempty" validate:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`

	// Текст сообщения
	Message string `json:"message" validate:"required,min=1,max=4000" example:"Как лучше организовать пространство в квартире 50 кв.м?"`

	// Контекст (ID сцены для анализа)
	SceneID string `json:"scene_id,omitempty" validate:"omitempty,uuid" example:"660e8400-e29b-41d4-a716-446655440001"`
}

// ChatMessageResponse represents AI response.
// @Description Ответ AI на сообщение
type ChatMessageResponse struct {
	// ID сессии
	SessionID string `json:"session_id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// ID сообщения
	MessageID string `json:"message_id" example:"660e8400-e29b-41d4-a716-446655440001"`

	// Текст ответа AI
	Response string `json:"response" example:"Для квартиры 50 кв.м рекомендую рассмотреть следующие варианты..."`

	// Рекомендации (если применимо)
	Suggestions []AISuggestionResponse `json:"suggestions,omitempty"`

	// Токенов использовано
	TokensUsed int `json:"tokens_used" example:"450"`

	// Время создания
	CreatedAt time.Time `json:"created_at" example:"2024-11-29T14:20:00Z"`
}

// AISuggestionResponse represents an AI suggestion.
// @Description Рекомендация от AI
type AISuggestionResponse struct {
	// Тип рекомендации
	// @enum(layout_change, furniture_placement, space_optimization, compliance_warning)
	Type string `json:"type" example:"layout_change"`

	// Название
	Title string `json:"title" example:"Объединение кухни и гостиной"`

	// Описание
	Description string `json:"description" example:"Объединение позволит увеличить полезную площадь на 15%"`

	// Уверенность (0-1)
	Confidence float64 `json:"confidence" example:"0.87"`

	// Действие (если применимо)
	Action *SuggestionActionResponse `json:"action,omitempty"`
}

// SuggestionActionResponse represents an actionable suggestion.
// @Description Действие по рекомендации
type SuggestionActionResponse struct {
	// Тип действия
	ActionType string `json:"action_type" example:"apply_change"`

	// Параметры действия
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// ChatHistoryResponse represents chat session history.
// @Description История чата
type ChatHistoryResponse struct {
	// ID сессии
	SessionID string `json:"session_id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// Сообщения
	Messages []ChatHistoryMessageResponse `json:"messages"`

	// Всего сообщений
	TotalMessages int `json:"total_messages" example:"12"`

	// Дата создания сессии
	CreatedAt time.Time `json:"created_at" example:"2024-11-29T10:00:00Z"`
}

// ChatHistoryMessageResponse represents a message in history.
// @Description Сообщение в истории
type ChatHistoryMessageResponse struct {
	// ID сообщения
	ID string `json:"id" example:"660e8400-e29b-41d4-a716-446655440001"`

	// Роль: user или assistant
	Role string `json:"role" example:"user"`

	// Текст сообщения
	Content string `json:"content" example:"Как оптимизировать планировку?"`

	// Время создания
	CreatedAt time.Time `json:"created_at" example:"2024-11-29T14:20:00Z"`
}

// =============================================================================
// AI Generation DTOs
// =============================================================================

// GenerateVariantsRequest represents AI generation input.
// @Description Запрос на генерацию вариантов планировки
type GenerateVariantsRequest struct {
	// ID исходной сцены
	SceneID string `json:"scene_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`

	// Количество вариантов (1-5)
	Count int `json:"count" validate:"required,min=1,max=5" example:"3"`

	// Стиль генерации
	// @enum(modern, classic, minimalist, scandinavian, loft)
	Style string `json:"style,omitempty" example:"modern"`

	// Приоритеты
	Priorities []string `json:"priorities,omitempty" example:"[\"maximize_light\",\"open_space\"]"`

	// Ограничения (не трогать элементы)
	KeepElements []string `json:"keep_elements,omitempty" example:"[\"wall-123\",\"kitchen-sink\"]"`

	// Бюджет (опционально)
	BudgetRange *BudgetRangeRequest `json:"budget_range,omitempty"`
}

// BudgetRangeRequest represents budget constraints.
// @Description Бюджетные ограничения
type BudgetRangeRequest struct {
	// Минимум (рубли)
	Min int `json:"min" example:"100000"`

	// Максимум (рубли)
	Max int `json:"max" example:"500000"`
}

// GenerateVariantsResponse represents generation results.
// @Description Результат генерации вариантов
type GenerateVariantsResponse struct {
	// ID задачи генерации
	TaskID string `json:"task_id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// Статус: processing, completed, failed
	Status string `json:"status" example:"completed"`

	// Сгенерированные варианты
	Variants []GeneratedVariantResponse `json:"variants,omitempty"`

	// Время генерации (мс)
	ProcessingTime int `json:"processing_time_ms,omitempty" example:"5200"`

	// Ошибка (если failed)
	Error string `json:"error,omitempty"`
}

// GeneratedVariantResponse represents a generated variant.
// @Description Сгенерированный вариант планировки
type GeneratedVariantResponse struct {
	// ID варианта
	ID string `json:"id" example:"660e8400-e29b-41d4-a716-446655440001"`

	// Название
	Name string `json:"name" example:"Вариант 1 - Открытое пространство"`

	// Описание изменений
	Description string `json:"description" example:"Объединение кухни и гостиной, оптимизация прихожей"`

	// ID созданной ветки
	BranchID string `json:"branch_id" example:"770e8400-e29b-41d4-a716-446655440002"`

	// Оценка качества (0-100)
	QualityScore int `json:"quality_score" example:"85"`

	// Оценочная стоимость реализации
	EstimatedCost int `json:"estimated_cost,omitempty" example:"350000"`

	// Превью (URL изображения)
	PreviewURL string `json:"preview_url,omitempty" example:"https://storage.granula.ru/previews/var-123.jpg"`

	// Изменения
	Changes []ChangeResponse `json:"changes"`
}

// ChangeResponse represents a change in variant.
// @Description Изменение в варианте
type ChangeResponse struct {
	// Тип изменения
	Type string `json:"type" example:"wall_removed"`

	// Описание
	Description string `json:"description" example:"Удалена стена между кухней и гостиной"`

	// Затронутые элементы
	AffectedElements []string `json:"affected_elements,omitempty" example:"[\"wall-123\"]"`
}

// =============================================================================
// AI Recognition DTOs
// =============================================================================

// RecognitionStatusResponse represents recognition task status.
// @Description Статус задачи распознавания
type RecognitionStatusResponse struct {
	// ID задачи
	TaskID string `json:"task_id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// ID планировки
	FloorPlanID string `json:"floor_plan_id" example:"660e8400-e29b-41d4-a716-446655440001"`

	// Статус: queued, processing, completed, failed
	Status string `json:"status" example:"processing"`

	// Прогресс (0-100)
	Progress int `json:"progress" example:"65"`

	// Текущий этап
	CurrentStep string `json:"current_step,omitempty" example:"Распознавание комнат"`

	// Результат (если completed)
	Result *RecognitionResultResponse `json:"result,omitempty"`

	// Ошибка (если failed)
	Error string `json:"error,omitempty"`

	// Время начала
	StartedAt time.Time `json:"started_at" example:"2024-11-29T14:20:00Z"`

	// Время завершения
	CompletedAt *time.Time `json:"completed_at,omitempty" example:"2024-11-29T14:21:15Z"`
}


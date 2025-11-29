// =============================================================================
// Request Service DTOs
// =============================================================================
package dto

import "time"

// =============================================================================
// Expert Request DTOs
// =============================================================================

// CreateExpertRequestRequest represents expert request creation input.
// @Description Данные для создания заявки на услуги эксперта
type CreateExpertRequestRequest struct {
	// ID воркспейса
	WorkspaceID string `json:"workspace_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`

	// Заголовок заявки (5-200 символов)
	Title string `json:"title" validate:"required,min=5,max=200" example:"Консультация по перепланировке 3-комнатной квартиры"`

	// Описание заявки
	Description string `json:"description" validate:"required,min=10,max=5000" example:"Необходима консультация по возможности объединения кухни и гостиной..."`

	// Категория услуги
	// @enum(consultation, documentation, expert_visit, full_package)
	Category string `json:"category" validate:"required,oneof=consultation documentation expert_visit full_package" example:"consultation"`

	// Приоритет (опционально)
	// @enum(low, normal, high, urgent)
	Priority string `json:"priority,omitempty" example:"normal"`

	// Контактный телефон
	ContactPhone string `json:"contact_phone,omitempty" validate:"omitempty,e164" example:"+79991234567"`

	// Контактный email
	ContactEmail string `json:"contact_email,omitempty" validate:"omitempty,email" example:"user@example.com"`
}

// ExpertRequestResponse represents expert request data.
// @Description Данные заявки на услуги эксперта
type ExpertRequestResponse struct {
	// ID заявки
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// ID воркспейса
	WorkspaceID string `json:"workspace_id" example:"660e8400-e29b-41d4-a716-446655440001"`

	// ID пользователя (автор)
	UserID string `json:"user_id" example:"770e8400-e29b-41d4-a716-446655440002"`

	// Заголовок
	Title string `json:"title" example:"Консультация по перепланировке"`

	// Описание
	Description string `json:"description" example:"Необходима консультация..."`

	// Категория
	Category string `json:"category" example:"consultation"`

	// Приоритет
	Priority string `json:"priority" example:"normal"`

	// Статус заявки
	// @enum(draft, pending, in_review, approved, rejected, assigned, in_progress, completed, cancelled)
	Status string `json:"status" example:"pending"`

	// ID назначенного эксперта
	ExpertID string `json:"expert_id,omitempty" example:"880e8400-e29b-41d4-a716-446655440003"`

	// Дата назначения эксперта
	AssignedAt *time.Time `json:"assigned_at,omitempty" example:"2024-11-30T10:00:00Z"`

	// Оценочная стоимость (рубли)
	EstimatedCost int `json:"estimated_cost" example:"2000"`

	// Финальная стоимость (рубли)
	FinalCost *int `json:"final_cost,omitempty" example:"2500"`

	// Причина отказа (если rejected)
	RejectionReason string `json:"rejection_reason,omitempty" example:""`

	// Контактный телефон
	ContactPhone string `json:"contact_phone,omitempty" example:"+79991234567"`

	// Контактный email
	ContactEmail string `json:"contact_email,omitempty" example:"user@example.com"`

	// Дата создания
	CreatedAt time.Time `json:"created_at" example:"2024-11-29T14:20:00Z"`

	// Дата обновления
	UpdatedAt time.Time `json:"updated_at" example:"2024-11-29T14:20:00Z"`

	// Дата завершения
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// UpdateExpertRequestRequest represents expert request update input.
// @Description Данные для обновления заявки (только для статуса draft)
type UpdateExpertRequestRequest struct {
	// Новый заголовок
	Title string `json:"title,omitempty" validate:"omitempty,min=5,max=200" example:"Обновленный заголовок"`

	// Новое описание
	Description string `json:"description,omitempty" validate:"omitempty,min=10,max=5000" example:"Обновленное описание..."`

	// Контактный телефон
	ContactPhone string `json:"contact_phone,omitempty" validate:"omitempty,e164" example:"+79991234567"`

	// Контактный email
	ContactEmail string `json:"contact_email,omitempty" validate:"omitempty,email" example:"user@example.com"`
}

// ExpertRequestListResponse represents a list of expert requests.
// @Description Список заявок с пагинацией
type ExpertRequestListResponse struct {
	// Заявки
	Requests []ExpertRequestResponse `json:"requests"`

	// Пагинация
	Pagination PaginationResponse `json:"pagination"`
}

// =============================================================================
// Request Status History DTOs
// =============================================================================

// StatusHistoryResponse represents request status history.
// @Description История изменения статуса заявки
type StatusHistoryResponse struct {
	// История изменений
	History []StatusChangeResponse `json:"history"`
}

// StatusChangeResponse represents a status change record.
// @Description Запись изменения статуса
type StatusChangeResponse struct {
	// ID записи
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// Предыдущий статус
	FromStatus string `json:"from_status" example:"pending"`

	// Новый статус
	ToStatus string `json:"to_status" example:"in_review"`

	// Комментарий
	Comment string `json:"comment" example:"Заявка взята на рассмотрение"`

	// ID пользователя, изменившего статус
	ChangedBy string `json:"changed_by,omitempty" example:"660e8400-e29b-41d4-a716-446655440001"`

	// Дата изменения
	ChangedAt time.Time `json:"changed_at" example:"2024-11-29T15:00:00Z"`
}

// =============================================================================
// Request Documents DTOs
// =============================================================================

// DocumentResponse represents an attached document.
// @Description Приложенный документ
type DocumentResponse struct {
	// ID документа
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// Тип документа
	// @enum(floor_plan, bti_certificate, ownership, other)
	Type string `json:"type" example:"floor_plan"`

	// Имя файла
	Name string `json:"name" example:"plan_floor_2.pdf"`

	// URL для скачивания
	URL string `json:"url" example:"https://storage.granula.ru/docs/doc-123.pdf"`

	// MIME тип
	MimeType string `json:"mime_type" example:"application/pdf"`

	// Размер (байты)
	Size int64 `json:"size" example:"1048576"`

	// ID загрузившего
	UploadedBy string `json:"uploaded_by" example:"660e8400-e29b-41d4-a716-446655440001"`

	// Дата загрузки
	UploadedAt time.Time `json:"uploaded_at" example:"2024-11-29T14:30:00Z"`
}

// UploadDocumentRequest represents document upload metadata.
// @Description Метаданные загружаемого документа
type UploadDocumentRequest struct {
	// Тип документа
	Type string `json:"type" validate:"required,oneof=floor_plan bti_certificate ownership other" example:"floor_plan"`

	// Имя файла (опционально, берется из файла)
	Name string `json:"name,omitempty" validate:"omitempty,max=255" example:"custom_name.pdf"`
}

// =============================================================================
// Request Pricing DTOs
// =============================================================================

// PricingResponse represents service pricing information.
// @Description Информация о ценах на услуги
type PricingResponse struct {
	// Категории услуг с ценами
	Categories []ServiceCategoryPriceResponse `json:"categories"`

	// Скидки (если есть)
	Discounts []DiscountResponse `json:"discounts,omitempty"`
}

// ServiceCategoryPriceResponse represents category pricing.
// @Description Цена категории услуг
type ServiceCategoryPriceResponse struct {
	// Код категории
	Code string `json:"code" example:"consultation"`

	// Название
	Name string `json:"name" example:"Консультация"`

	// Описание
	Description string `json:"description" example:"Консультация специалиста БТИ по вопросам перепланировки"`

	// Базовая цена (рубли)
	BasePrice int `json:"base_price" example:"2000"`

	// Срок выполнения
	EstimatedDays string `json:"estimated_days" example:"3-5 рабочих дней"`

	// Что включено
	Includes []string `json:"includes" example:"[\"Анализ документации\",\"Устная консультация\",\"Письменное заключение\"]"`
}

// DiscountResponse represents a discount.
// @Description Скидка
type DiscountResponse struct {
	// Код скидки
	Code string `json:"code" example:"FIRST_ORDER"`

	// Название
	Name string `json:"name" example:"Скидка на первый заказ"`

	// Процент скидки
	Percent int `json:"percent" example:"10"`

	// Действует до
	ValidUntil *time.Time `json:"valid_until,omitempty" example:"2024-12-31T23:59:59Z"`
}


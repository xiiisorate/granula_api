// =============================================================================
// Floor Plan DTOs
// =============================================================================
package dto

import "time"

// =============================================================================
// Floor Plan DTOs
// =============================================================================

// UploadFloorPlanRequest represents floor plan upload metadata.
// @Description Метаданные для загрузки планировки (файл передается отдельно)
type UploadFloorPlanRequest struct {
	// UUID воркспейса
	WorkspaceID string `json:"workspace_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`

	// Название планировки
	Name string `json:"name" validate:"required,min=2,max=255" example:"План квартиры - этаж 2"`

	// Описание (опционально)
	Description string `json:"description,omitempty" validate:"max=1000" example:"План второго этажа жилого комплекса"`

	// Адрес объекта (опционально)
	Address string `json:"address,omitempty" validate:"max=500" example:"г. Москва, ул. Тверская, д. 15, кв. 42"`
}

// FloorPlanResponse represents floor plan data in response.
// @Description Данные планировки
type FloorPlanResponse struct {
	// UUID планировки
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// UUID воркспейса
	WorkspaceID string `json:"workspace_id" example:"660e8400-e29b-41d4-a716-446655440001"`

	// Название
	Name string `json:"name" example:"План квартиры - этаж 2"`

	// Описание
	Description string `json:"description" example:"План второго этажа"`

	// Адрес
	Address string `json:"address,omitempty" example:"г. Москва, ул. Тверская, д. 15"`

	// Статус обработки
	// @enum(uploaded, processing, recognized, failed)
	Status string `json:"status" example:"recognized"`

	// URL оригинального изображения
	ImageURL string `json:"image_url" example:"https://storage.granula.ru/floorplans/fp-123.jpg"`

	// URL миниатюры
	ThumbnailURL string `json:"thumbnail_url,omitempty" example:"https://storage.granula.ru/thumbnails/fp-123-thumb.jpg"`

	// Результат распознавания (если status=recognized)
	RecognitionResult *RecognitionResultResponse `json:"recognition_result,omitempty"`

	// UUID связанной 3D сцены
	SceneID string `json:"scene_id,omitempty" example:"770e8400-e29b-41d4-a716-446655440002"`

	// Дата создания
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`

	// Дата обновления
	UpdatedAt time.Time `json:"updated_at" example:"2024-11-29T14:20:00Z"`
}

// RecognitionResultResponse represents AI recognition results.
// @Description Результат AI распознавания планировки
type RecognitionResultResponse struct {
	// Количество распознанных комнат
	RoomsCount int `json:"rooms_count" example:"4"`

	// Общая площадь (кв.м)
	TotalArea float64 `json:"total_area" example:"78.5"`

	// Жилая площадь (кв.м)
	LivingArea float64 `json:"living_area" example:"52.3"`

	// Распознанные комнаты
	Rooms []RecognizedRoomResponse `json:"rooms"`

	// Уверенность распознавания (0-1)
	Confidence float64 `json:"confidence" example:"0.94"`

	// Время обработки (мс)
	ProcessingTime int `json:"processing_time_ms" example:"1250"`
}

// RecognizedRoomResponse represents a recognized room.
// @Description Распознанная комната
type RecognizedRoomResponse struct {
	// Тип комнаты
	// @enum(living_room, bedroom, kitchen, bathroom, toilet, hallway, balcony, storage, other)
	Type string `json:"type" example:"bedroom"`

	// Название комнаты
	Name string `json:"name" example:"Спальня 1"`

	// Площадь (кв.м)
	Area float64 `json:"area" example:"18.5"`

	// Координаты углов (для отображения на изображении)
	Polygon []PointResponse `json:"polygon"`
}

// PointResponse represents a 2D point.
// @Description Точка на плоскости
type PointResponse struct {
	X float64 `json:"x" example:"120.5"`
	Y float64 `json:"y" example:"80.3"`
}

// FloorPlanListResponse represents a list of floor plans.
// @Description Список планировок с пагинацией
type FloorPlanListResponse struct {
	// Список планировок
	FloorPlans []FloorPlanResponse `json:"floor_plans"`

	// Пагинация
	Pagination PaginationResponse `json:"pagination"`
}

// ProcessFloorPlanRequest represents a request to process floor plan.
// @Description Запрос на обработку планировки AI
type ProcessFloorPlanRequest struct {
	// Автоматически создать 3D сцену после распознавания
	CreateScene bool `json:"create_scene" example:"true"`

	// Выполнить проверку на соответствие нормам
	RunCompliance bool `json:"run_compliance" example:"true"`
}


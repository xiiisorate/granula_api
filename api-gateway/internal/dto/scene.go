// =============================================================================
// Scene and Branch DTOs
// =============================================================================
package dto

import "time"

// =============================================================================
// Scene DTOs
// =============================================================================

// SceneResponse represents 3D scene data.
// @Description 3D сцена с элементами
type SceneResponse struct {
	// UUID сцены
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// UUID воркспейса
	WorkspaceID string `json:"workspace_id" example:"660e8400-e29b-41d4-a716-446655440001"`

	// UUID исходной планировки
	FloorPlanID string `json:"floor_plan_id" example:"770e8400-e29b-41d4-a716-446655440002"`

	// Название сцены
	Name string `json:"name" example:"Сцена квартиры - вариант 1"`

	// Описание
	Description string `json:"description,omitempty" example:"Основной вариант перепланировки"`

	// Статус сцены
	// @enum(draft, active, archived)
	Status string `json:"status" example:"active"`

	// Количество элементов
	ElementCount int `json:"element_count" example:"42"`

	// Количество веток (вариантов)
	BranchCount int `json:"branch_count" example:"3"`

	// Статистика по типам элементов
	Statistics SceneStatisticsResponse `json:"statistics"`

	// Дата создания
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`

	// Дата обновления
	UpdatedAt time.Time `json:"updated_at" example:"2024-11-29T14:20:00Z"`
}

// SceneStatisticsResponse represents scene statistics.
// @Description Статистика сцены
type SceneStatisticsResponse struct {
	// Количество стен
	WallsCount int `json:"walls_count" example:"15"`

	// Количество комнат
	RoomsCount int `json:"rooms_count" example:"4"`

	// Количество окон
	WindowsCount int `json:"windows_count" example:"6"`

	// Количество дверей
	DoorsCount int `json:"doors_count" example:"5"`

	// Количество мебели
	FurnitureCount int `json:"furniture_count" example:"12"`

	// Общая площадь (кв.м)
	TotalArea float64 `json:"total_area" example:"78.5"`

	// Жилая площадь (кв.м)
	LivingArea float64 `json:"living_area" example:"52.3"`
}

// ElementResponse represents a scene element.
// @Description Элемент 3D сцены
type ElementResponse struct {
	// UUID элемента
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// UUID сцены
	SceneID string `json:"scene_id" example:"660e8400-e29b-41d4-a716-446655440001"`

	// Тип элемента
	// @enum(wall, room, door, window, furniture, fixture)
	Type string `json:"type" example:"wall"`

	// Название
	Name string `json:"name" example:"Несущая стена кухня-гостиная"`

	// Позиция (центр элемента)
	Position Point3DResponse `json:"position"`

	// Размеры
	Dimensions Dimensions3DResponse `json:"dimensions"`

	// Поворот
	Rotation RotationResponse `json:"rotation"`

	// Свойства элемента (зависят от типа)
	Properties map[string]interface{} `json:"properties,omitempty"`

	// Является ли несущим элементом
	IsLoadBearing bool `json:"is_load_bearing" example:"true"`

	// Дата создания
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
}

// Point3DResponse represents a 3D point.
// @Description Точка в 3D пространстве
type Point3DResponse struct {
	X float64 `json:"x" example:"5.0"`
	Y float64 `json:"y" example:"2.5"`
	Z float64 `json:"z" example:"0.0"`
}

// Dimensions3DResponse represents 3D dimensions.
// @Description Размеры в 3D
type Dimensions3DResponse struct {
	Width  float64 `json:"width" example:"3.5"`
	Height float64 `json:"height" example:"2.8"`
	Depth  float64 `json:"depth" example:"0.2"`
}

// RotationResponse represents rotation angles.
// @Description Углы поворота (градусы)
type RotationResponse struct {
	X float64 `json:"x" example:"0.0"`
	Y float64 `json:"y" example:"45.0"`
	Z float64 `json:"z" example:"0.0"`
}

// CreateElementRequest represents element creation input.
// @Description Данные для создания элемента
type CreateElementRequest struct {
	// Тип элемента
	Type string `json:"type" validate:"required,oneof=wall room door window furniture fixture" example:"wall"`

	// Название
	Name string `json:"name" validate:"required,min=1,max=255" example:"Новая стена"`

	// Позиция
	Position Point3DResponse `json:"position" validate:"required"`

	// Размеры
	Dimensions Dimensions3DResponse `json:"dimensions" validate:"required"`

	// Поворот (опционально)
	Rotation *RotationResponse `json:"rotation,omitempty"`

	// Свойства
	Properties map[string]interface{} `json:"properties,omitempty"`

	// Является несущим
	IsLoadBearing bool `json:"is_load_bearing" example:"false"`
}

// UpdateElementRequest represents element update input.
// @Description Данные для обновления элемента
type UpdateElementRequest struct {
	// Новое название
	Name string `json:"name,omitempty" validate:"omitempty,min=1,max=255" example:"Обновленная стена"`

	// Новая позиция
	Position *Point3DResponse `json:"position,omitempty"`

	// Новые размеры
	Dimensions *Dimensions3DResponse `json:"dimensions,omitempty"`

	// Новый поворот
	Rotation *RotationResponse `json:"rotation,omitempty"`

	// Обновленные свойства
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// =============================================================================
// Branch DTOs
// =============================================================================

// BranchResponse represents a design branch/variant.
// @Description Ветка дизайна (вариант планировки)
type BranchResponse struct {
	// UUID ветки
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// UUID сцены
	SceneID string `json:"scene_id" example:"660e8400-e29b-41d4-a716-446655440001"`

	// UUID родительской ветки (для форков)
	ParentBranchID string `json:"parent_branch_id,omitempty" example:"770e8400-e29b-41d4-a716-446655440002"`

	// Название ветки
	Name string `json:"name" example:"Вариант с объединенной кухней"`

	// Описание
	Description string `json:"description,omitempty" example:"Вариант с объединением кухни и гостиной"`

	// Статус ветки
	// @enum(active, archived, merged)
	Status string `json:"status" example:"active"`

	// Является ли главной веткой
	IsMain bool `json:"is_main" example:"false"`

	// UUID автора
	CreatedBy string `json:"created_by" example:"880e8400-e29b-41d4-a716-446655440003"`

	// Дата создания
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`

	// Дата обновления
	UpdatedAt time.Time `json:"updated_at" example:"2024-11-29T14:20:00Z"`
}

// CreateBranchRequest represents branch creation input.
// @Description Данные для создания ветки
type CreateBranchRequest struct {
	// Название ветки
	Name string `json:"name" validate:"required,min=2,max=255" example:"Вариант 2 - студия"`

	// Описание
	Description string `json:"description,omitempty" validate:"max=1000" example:"Вариант с объединением всех комнат в студию"`

	// UUID родительской ветки (опционально, по умолчанию main)
	ParentBranchID string `json:"parent_branch_id,omitempty" validate:"omitempty,uuid" example:"770e8400-e29b-41d4-a716-446655440002"`
}

// CompareBranchesResponse represents branch comparison result.
// @Description Результат сравнения веток
type CompareBranchesResponse struct {
	// Ветка 1
	Branch1 BranchSummaryResponse `json:"branch_1"`

	// Ветка 2
	Branch2 BranchSummaryResponse `json:"branch_2"`

	// Различия
	Differences []DifferenceResponse `json:"differences"`

	// Количество различий
	DifferenceCount int `json:"difference_count" example:"5"`
}

// BranchSummaryResponse represents branch summary for comparison.
// @Description Сводка ветки для сравнения
type BranchSummaryResponse struct {
	ID           string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name         string  `json:"name" example:"Вариант 1"`
	ElementCount int     `json:"element_count" example:"42"`
	TotalArea    float64 `json:"total_area" example:"78.5"`
}

// DifferenceResponse represents a difference between branches.
// @Description Различие между ветками
type DifferenceResponse struct {
	// Тип различия: added, removed, modified
	Type string `json:"type" example:"modified"`

	// ID элемента
	ElementID string `json:"element_id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// Тип элемента
	ElementType string `json:"element_type" example:"wall"`

	// Описание изменения
	Description string `json:"description" example:"Изменена позиция стены"`
}


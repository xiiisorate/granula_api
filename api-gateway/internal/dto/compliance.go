// =============================================================================
// Compliance Service DTOs
// =============================================================================
package dto

import "time"

// =============================================================================
// Compliance Check DTOs
// =============================================================================

// CheckComplianceRequest represents compliance check input.
// @Description Запрос на проверку соответствия нормам
type CheckComplianceRequest struct {
	// ID сцены для проверки
	SceneID string `json:"scene_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`

	// ID ветки (опционально, по умолчанию main)
	BranchID string `json:"branch_id,omitempty" validate:"omitempty,uuid" example:"660e8400-e29b-41d4-a716-446655440001"`

	// Категории правил для проверки (пусто = все)
	Categories []string `json:"categories,omitempty" example:"[\"minimum_area\",\"ventilation\",\"load_bearing\"]"`

	// Тип операции (для контекста)
	// @enum(construction, renovation, redevelopment)
	OperationType string `json:"operation_type,omitempty" example:"redevelopment"`
}

// ComplianceCheckResponse represents compliance check results.
// @Description Результат проверки на соответствие нормам
type ComplianceCheckResponse struct {
	// ID проверки
	CheckID string `json:"check_id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// ID сцены
	SceneID string `json:"scene_id" example:"660e8400-e29b-41d4-a716-446655440001"`

	// Общий статус: passed, warning, failed
	Status string `json:"status" example:"warning"`

	// Общий балл соответствия (0-100)
	Score int `json:"score" example:"78"`

	// Количество нарушений
	ViolationCount int `json:"violation_count" example:"2"`

	// Количество предупреждений
	WarningCount int `json:"warning_count" example:"3"`

	// Нарушения
	Violations []ViolationResponse `json:"violations"`

	// Статистика по категориям
	CategoryStats []CategoryStatResponse `json:"category_stats"`

	// Время проверки
	CheckedAt time.Time `json:"checked_at" example:"2024-11-29T14:20:00Z"`

	// Время выполнения (мс)
	ProcessingTime int `json:"processing_time_ms" example:"320"`
}

// ViolationResponse represents a compliance violation.
// @Description Нарушение нормы
type ViolationResponse struct {
	// ID нарушения
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// ID правила
	RuleID string `json:"rule_id" example:"SNIP-MIN-AREA-01"`

	// Название правила
	RuleName string `json:"rule_name" example:"Минимальная площадь жилой комнаты"`

	// Категория
	Category string `json:"category" example:"minimum_area"`

	// Серьезность: error (блокирующая), warning (предупреждение), info
	Severity string `json:"severity" example:"error"`

	// Описание нарушения
	Description string `json:"description" example:"Площадь спальни (7.5 кв.м) меньше минимально допустимой (8 кв.м) согласно СНиП 31-01-2003"`

	// Затронутые элементы
	AffectedElements []AffectedElementResponse `json:"affected_elements"`

	// Рекомендации по исправлению
	Recommendations []string `json:"recommendations" example:"[\"Увеличить площадь комнаты минимум на 0.5 кв.м\",\"Рассмотреть объединение с соседней комнатой\"]"`

	// Ссылка на нормативный документ
	LegalReference string `json:"legal_reference,omitempty" example:"СНиП 31-01-2003, п. 5.3"`
}

// AffectedElementResponse represents an element affected by violation.
// @Description Затронутый элемент
type AffectedElementResponse struct {
	// ID элемента
	ElementID string `json:"element_id" example:"660e8400-e29b-41d4-a716-446655440001"`

	// Тип элемента
	ElementType string `json:"element_type" example:"room"`

	// Название элемента
	ElementName string `json:"element_name" example:"Спальня 2"`
}

// CategoryStatResponse represents statistics by category.
// @Description Статистика по категории правил
type CategoryStatResponse struct {
	// Категория
	Category string `json:"category" example:"minimum_area"`

	// Название категории
	CategoryName string `json:"category_name" example:"Минимальные площади"`

	// Общее количество правил
	TotalRules int `json:"total_rules" example:"8"`

	// Пройдено
	Passed int `json:"passed" example:"6"`

	// Нарушения
	Violations int `json:"violations" example:"1"`

	// Предупреждения
	Warnings int `json:"warnings" example:"1"`
}

// =============================================================================
// Compliance Rules DTOs
// =============================================================================

// ComplianceRuleResponse represents a compliance rule.
// @Description Правило проверки соответствия
type ComplianceRuleResponse struct {
	// ID правила
	ID string `json:"id" example:"SNIP-MIN-AREA-01"`

	// Название
	Name string `json:"name" example:"Минимальная площадь жилой комнаты"`

	// Категория
	Category string `json:"category" example:"minimum_area"`

	// Описание
	Description string `json:"description" example:"Площадь жилой комнаты должна быть не менее 8 кв.м"`

	// Серьезность по умолчанию
	DefaultSeverity string `json:"default_severity" example:"error"`

	// Нормативный документ
	LegalReference string `json:"legal_reference" example:"СНиП 31-01-2003"`

	// Применимые типы операций
	ApplicableOperations []string `json:"applicable_operations" example:"[\"construction\",\"redevelopment\"]"`

	// Активно
	IsActive bool `json:"is_active" example:"true"`
}

// RulesListResponse represents a list of compliance rules.
// @Description Список правил проверки
type RulesListResponse struct {
	// Правила
	Rules []ComplianceRuleResponse `json:"rules"`

	// Категории
	Categories []RuleCategoryResponse `json:"categories"`

	// Всего правил
	TotalRules int `json:"total_rules" example:"45"`
}

// RuleCategoryResponse represents a rule category.
// @Description Категория правил
type RuleCategoryResponse struct {
	// Код категории
	Code string `json:"code" example:"minimum_area"`

	// Название
	Name string `json:"name" example:"Минимальные площади"`

	// Описание
	Description string `json:"description" example:"Правила минимальных площадей помещений"`

	// Количество правил
	RuleCount int `json:"rule_count" example:"8"`
}

// =============================================================================
// Compliance History DTOs
// =============================================================================

// ComplianceHistoryResponse represents compliance check history.
// @Description История проверок соответствия
type ComplianceHistoryResponse struct {
	// Проверки
	Checks []ComplianceCheckSummaryResponse `json:"checks"`

	// Пагинация
	Pagination PaginationResponse `json:"pagination"`
}

// ComplianceCheckSummaryResponse represents a summary of compliance check.
// @Description Сводка проверки
type ComplianceCheckSummaryResponse struct {
	// ID проверки
	CheckID string `json:"check_id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// Статус
	Status string `json:"status" example:"warning"`

	// Балл
	Score int `json:"score" example:"78"`

	// Нарушений
	ViolationCount int `json:"violation_count" example:"2"`

	// Предупреждений
	WarningCount int `json:"warning_count" example:"3"`

	// Дата проверки
	CheckedAt time.Time `json:"checked_at" example:"2024-11-29T14:20:00Z"`
}


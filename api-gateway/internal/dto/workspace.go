// =============================================================================
// Workspace DTOs
// =============================================================================
package dto

import "time"

// =============================================================================
// Workspace DTOs
// =============================================================================

// CreateWorkspaceRequest represents workspace creation input.
// @Description Данные для создания воркспейса
type CreateWorkspaceRequest struct {
	// Название воркспейса (2-100 символов)
	Name string `json:"name" validate:"required,min=2,max=100" example:"Перепланировка квартиры на Тверской"`

	// Описание (опционально, до 1000 символов)
	Description string `json:"description,omitempty" validate:"max=1000" example:"Проект перепланировки 3-комнатной квартиры в центре Москвы"`
}

// WorkspaceResponse represents workspace data in response.
// @Description Данные воркспейса
type WorkspaceResponse struct {
	// UUID воркспейса
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// UUID владельца
	OwnerID string `json:"owner_id" example:"660e8400-e29b-41d4-a716-446655440001"`

	// Название
	Name string `json:"name" example:"Перепланировка квартиры на Тверской"`

	// Описание
	Description string `json:"description" example:"Проект перепланировки 3-комнатной квартиры"`

	// Количество участников
	MemberCount int `json:"member_count" example:"3"`

	// Количество проектов (планировок)
	ProjectCount int `json:"project_count" example:"2"`

	// Дата создания
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`

	// Дата обновления
	UpdatedAt time.Time `json:"updated_at" example:"2024-11-29T14:20:00Z"`

	// Участники (если запрошены)
	Members []MemberResponse `json:"members,omitempty"`
}

// UpdateWorkspaceRequest represents workspace update input.
// @Description Данные для обновления воркспейса
type UpdateWorkspaceRequest struct {
	// Новое название (опционально)
	Name string `json:"name,omitempty" validate:"omitempty,min=2,max=100" example:"Новое название проекта"`

	// Новое описание (опционально)
	Description string `json:"description,omitempty" validate:"max=1000" example:"Обновленное описание проекта"`
}

// MemberResponse represents workspace member data.
// @Description Участник воркспейса
type MemberResponse struct {
	// UUID участника
	ID string `json:"id" example:"770e8400-e29b-41d4-a716-446655440002"`

	// UUID пользователя
	UserID string `json:"user_id" example:"880e8400-e29b-41d4-a716-446655440003"`

	// Роль: owner, admin, editor, viewer
	Role string `json:"role" example:"editor"`

	// Дата присоединения
	JoinedAt time.Time `json:"joined_at" example:"2024-02-20T12:00:00Z"`
}

// AddMemberRequest represents member addition input.
// @Description Данные для добавления участника
type AddMemberRequest struct {
	// UUID пользователя для добавления
	UserID string `json:"user_id" validate:"required,uuid" example:"880e8400-e29b-41d4-a716-446655440003"`

	// Роль: admin, editor, viewer (owner назначается автоматически)
	Role string `json:"role" validate:"required,oneof=admin editor viewer" example:"editor"`
}

// UpdateMemberRoleRequest represents member role update input.
// @Description Данные для изменения роли участника
type UpdateMemberRoleRequest struct {
	// Новая роль
	Role string `json:"role" validate:"required,oneof=admin editor viewer" example:"admin"`
}

// WorkspaceListResponse represents a list of workspaces.
// @Description Список воркспейсов с пагинацией
type WorkspaceListResponse struct {
	// Список воркспейсов
	Workspaces []WorkspaceResponse `json:"workspaces"`

	// Пагинация
	Pagination PaginationResponse `json:"pagination"`
}


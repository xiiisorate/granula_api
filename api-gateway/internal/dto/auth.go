// =============================================================================
// Package dto contains Data Transfer Objects for API request/response.
// =============================================================================
// DTOs define the structure of data exchanged between the API Gateway
// and clients. They include validation tags and Swagger documentation.
//
// =============================================================================
package dto

import "time"

// =============================================================================
// Auth DTOs
// =============================================================================

// RegisterRequest represents user registration input.
// @Description Данные для регистрации нового пользователя
type RegisterRequest struct {
	// Email пользователя (уникальный)
	// @example user@example.com
	Email string `json:"email" validate:"required,email,max=255" example:"user@example.com"`

	// Пароль (минимум 8 символов)
	// @example SecurePass123!
	Password string `json:"password" validate:"required,min=8,max=72" example:"SecurePass123!"`

	// Имя пользователя
	// @example Иван Петров
	Name string `json:"name" validate:"required,min=2,max=255" example:"Иван Петров"`
}

// RegisterResponse represents successful registration response.
// @Description Ответ успешной регистрации с токенами
type RegisterResponse struct {
	// UUID созданного пользователя
	UserID string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// JWT Access Token (время жизни 15 минут)
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`

	// Refresh Token для обновления Access Token
	RefreshToken string `json:"refresh_token" example:"dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4..."`

	// Время истечения Access Token (Unix timestamp)
	ExpiresAt int64 `json:"expires_at" example:"1701234567"`

	// Тип токена (всегда "Bearer")
	TokenType string `json:"token_type" example:"Bearer"`
}

// LoginRequest represents user login input.
// @Description Данные для входа в систему
type LoginRequest struct {
	// Email пользователя
	Email string `json:"email" validate:"required,email" example:"user@example.com"`

	// Пароль
	Password string `json:"password" validate:"required" example:"SecurePass123!"`

	// Идентификатор устройства (опционально, для управления сессиями)
	DeviceID string `json:"device_id,omitempty" example:"device-123-abc"`
}

// LoginResponse represents successful login response.
// @Description Ответ успешного входа с токенами
type LoginResponse struct {
	// UUID пользователя
	UserID string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// JWT Access Token
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`

	// Refresh Token
	RefreshToken string `json:"refresh_token" example:"dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4..."`

	// Время истечения (Unix timestamp)
	ExpiresAt int64 `json:"expires_at" example:"1701234567"`

	// Тип токена
	TokenType string `json:"token_type" example:"Bearer"`
}

// RefreshRequest represents token refresh input.
// @Description Запрос на обновление токенов
type RefreshRequest struct {
	// Refresh Token, полученный при логине
	RefreshToken string `json:"refresh_token" validate:"required" example:"dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4..."`
}

// RefreshResponse represents successful token refresh response.
// @Description Ответ с новыми токенами
type RefreshResponse struct {
	// Новый Access Token
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`

	// Новый Refresh Token
	RefreshToken string `json:"refresh_token" example:"bmV3IHJlZnJlc2ggdG9rZW4..."`

	// Время истечения нового Access Token
	ExpiresAt int64 `json:"expires_at" example:"1701234567"`
}

// LogoutRequest represents logout input.
// @Description Запрос на выход из системы
type LogoutRequest struct {
	// Refresh Token для инвалидации
	RefreshToken string `json:"refresh_token" validate:"required" example:"dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4..."`
}

// LogoutAllResponse represents logout from all devices response.
// @Description Ответ выхода со всех устройств
type LogoutAllResponse struct {
	// Успешность операции
	Success bool `json:"success" example:"true"`

	// Количество отозванных сессий
	SessionsRevoked int `json:"sessions_revoked" example:"3"`
}

// =============================================================================
// User DTOs
// =============================================================================

// UserProfile represents user profile data.
// @Description Профиль пользователя
type UserProfile struct {
	// UUID пользователя
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`

	// Email
	Email string `json:"email" example:"user@example.com"`

	// Имя
	Name string `json:"name" example:"Иван Петров"`

	// URL аватара
	AvatarURL string `json:"avatar_url,omitempty" example:"https://storage.granula.ru/avatars/user-123.jpg"`

	// Телефон
	Phone string `json:"phone,omitempty" example:"+7 (999) 123-45-67"`

	// Дата регистрации
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`

	// Дата последнего обновления
	UpdatedAt time.Time `json:"updated_at" example:"2024-11-29T14:20:00Z"`
}

// UpdateProfileRequest represents profile update input.
// @Description Данные для обновления профиля
type UpdateProfileRequest struct {
	// Новое имя (опционально)
	Name string `json:"name,omitempty" validate:"omitempty,min=2,max=255" example:"Иван Сидоров"`

	// Новый телефон (опционально)
	Phone string `json:"phone,omitempty" validate:"omitempty,e164" example:"+79991234567"`
}

// =============================================================================
// Common Response DTOs
// =============================================================================

// SuccessResponse represents a generic success response.
// @Description Успешный ответ без данных
type SuccessResponse struct {
	// Успешность операции
	Success bool `json:"success" example:"true"`

	// Сообщение (опционально)
	Message string `json:"message,omitempty" example:"Operation completed successfully"`
}

// ErrorResponse represents an error response.
// @Description Ответ с ошибкой
type ErrorResponse struct {
	// Код ошибки
	Code string `json:"code" example:"VALIDATION_ERROR"`

	// Человекочитаемое сообщение
	Message string `json:"message" example:"Invalid input data"`

	// Детали ошибки (для валидации)
	Details map[string]string `json:"details,omitempty"`

	// ID запроса для отладки
	RequestID string `json:"request_id,omitempty" example:"req-abc123"`
}

// PaginationRequest represents pagination parameters.
// @Description Параметры пагинации
type PaginationRequest struct {
	// Номер страницы (начиная с 1)
	Page int `json:"page" query:"page" validate:"min=1" example:"1"`

	// Количество элементов на странице (1-100)
	PageSize int `json:"page_size" query:"page_size" validate:"min=1,max=100" example:"20"`
}

// PaginationResponse represents pagination metadata in response.
// @Description Метаданные пагинации
type PaginationResponse struct {
	// Общее количество элементов
	Total int `json:"total" example:"150"`

	// Текущая страница
	Page int `json:"page" example:"1"`

	// Размер страницы
	PageSize int `json:"page_size" example:"20"`

	// Всего страниц
	TotalPages int `json:"total_pages" example:"8"`

	// Есть следующая страница
	HasNext bool `json:"has_next" example:"true"`

	// Есть предыдущая страница
	HasPrev bool `json:"has_prev" example:"false"`
}


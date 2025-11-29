// =============================================================================
// Package swagger provides Swagger/OpenAPI documentation for Granula API.
// =============================================================================
// This file contains the main Swagger documentation annotations using swaggo.
// All API endpoints are documented with request/response schemas, examples,
// and error codes.
//
// To regenerate Swagger docs:
//
//	swag init -g cmd/main.go -o internal/swagger/docs
//
// =============================================================================
package swagger

// =============================================================================
// @title Granula API
// @version 1.0.0
// @description API для интеллектуального сервиса планирования ремонта и перепланировки.
// @description
// @description ## Возможности
// @description - Загрузка и распознавание планировок квартир
// @description - 3D редактор для изменения планировки
// @description - Проверка соответствия нормам СНиП и ЖК РФ
// @description - AI-генерация вариантов планировки
// @description - Заявки на консультации экспертов БТИ
// @description
// @description ## Аутентификация
// @description API использует JWT токены для аутентификации.
// @description Получите токены через `/api/v1/auth/login` или `/api/v1/auth/register`.
// @description Передавайте Access Token в заголовке `Authorization: Bearer <token>`.
// @description
// @description ## Rate Limiting
// @description - Анонимные запросы: 60 req/min
// @description - Аутентифицированные: 300 req/min
// @description - AI endpoints: 30 req/min
// @description
// @description ## Версионирование
// @description Все endpoints имеют префикс `/api/v1/`.
// @description При выпуске новых версий старые версии поддерживаются минимум 6 месяцев.

// @termsOfService https://granula.ru/terms

// @contact.name Granula API Support
// @contact.url https://granula.ru/support
// @contact.email api@granula.ru

// @license.name Proprietary
// @license.url https://granula.ru/license

// @host api.granula.ru
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите JWT токен в формате: Bearer {token}

// @tag.name auth
// @tag.description Аутентификация и управление сессиями

// @tag.name users
// @tag.description Профили пользователей и настройки

// @tag.name workspaces
// @tag.description Рабочие пространства для проектов

// @tag.name floor-plans
// @tag.description Загрузка и обработка планировок

// @tag.name scenes
// @tag.description 3D сцены и элементы

// @tag.name branches
// @tag.description Версионирование и варианты дизайна

// @tag.name ai
// @tag.description AI распознавание и генерация

// @tag.name compliance
// @tag.description Проверка соответствия нормам

// @tag.name requests
// @tag.description Заявки на услуги экспертов

// @tag.name notifications
// @tag.description Уведомления пользователей

// =============================================================================

// SwaggerInfo holds the Swagger documentation metadata.
// It is populated by swag init during build.
var SwaggerInfo = struct {
	Title       string
	Description string
	Version     string
	Host        string
	BasePath    string
}{
	Title:       "Granula API",
	Description: "API для интеллектуального сервиса планирования ремонта",
	Version:     "1.0.0",
	Host:        "api.granula.ru",
	BasePath:    "/api/v1",
}

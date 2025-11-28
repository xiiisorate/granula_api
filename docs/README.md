# Granula API - Техническая документация

## Обзор проекта

**Granula** — интеллектуальный сервис для планирования ремонта и перепланировки квартир. API обеспечивает серверную логику для:

- Распознавания и оцифровки планов помещений
- Управления 3D-сценами и версионирования изменений
- AI-генерации вариантов планировки через систему веток (branches)
- Проверки соответствия строительным нормам (СНиП, ЖК РФ)
- Workflow заявок на экспертизу

## Технологический стек

| Компонент | Технология | Назначение |
|-----------|------------|------------|
| **Runtime** | Go 1.22+ | Основной язык |
| **Web Framework** | Fiber v2 | HTTP сервер |
| **PostgreSQL** | 16+ | Реляционные данные |
| **MongoDB** | 7+ | 3D сцены, документы |
| **Redis** | 7+ | Кэш, сессии, очереди |
| **MinIO/S3** | Latest | Файловое хранилище |
| **OpenRouter** | API | AI/LLM интеграция |

## Архитектура (Микросервисы)

Система построена на микросервисной архитектуре для независимого масштабирования и параллельной разработки.

```
                              ┌─────────────────┐
                              │  Load Balancer  │
                              └────────┬────────┘
                                       │
                                       ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                              API GATEWAY (:8080)                              │
│                    REST routing, JWT validation, Rate limiting               │
└──────────────────────────────────────┬───────────────────────────────────────┘
                                       │ gRPC
       ┌───────────────┬───────────────┼───────────────┬───────────────┐
       │               │               │               │               │
       ▼               ▼               ▼               ▼               ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│    Auth     │ │    User     │ │  Workspace  │ │ Floor Plan  │ │    Scene    │
│   Service   │ │   Service   │ │   Service   │ │   Service   │ │   Service   │
│   :50051    │ │   :50052    │ │   :50053    │ │   :50054    │ │   :50055    │
└──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └──────┬──────┘
       │               │               │               │               │
       ▼               ▼               ▼               │               ▼
┌─────────────────────────────────────────────┐       │        ┌─────────────┐
│                 PostgreSQL                   │       │        │   MongoDB   │
│  auth_db │ users_db │ workspaces_db │ ...   │       │        │  scenes_db  │
└─────────────────────────────────────────────┘       │        └─────────────┘
                                                      │
┌─────────────┐ ┌─────────────┐ ┌─────────────┐       │        ┌─────────────┐
│   Branch    │ │     AI      │ │ Compliance  │       │        │   Request   │
│   Service   │ │   Service   │ │   Service   │◄──────┘        │   Service   │
│   :50056    │ │   :50057    │ │   :50058    │                │   :50059    │
└──────┬──────┘ └──────┬──────┘ └─────────────┘                └──────┬──────┘
       │               │                                               │
       ▼               ▼                                               ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐                ┌─────────────┐
│   MongoDB   │ │ OpenRouter  │ │    Redis    │                │Notification │
│ branches_db │ │    API      │ │   Pub/Sub   │◄───────────────│   Service   │
└─────────────┘ └─────────────┘ └─────────────┘                │   :50060    │
                                                               └─────────────┘
```

### Микросервисы

| Сервис | Порт | Описание |
|--------|------|----------|
| **API Gateway** | 8080 | REST API, JWT validation, WebSocket |
| **Auth Service** | 50051 | Регистрация, логин, OAuth, токены |
| **User Service** | 50052 | Профили, настройки, сессии |
| **Workspace Service** | 50053 | Проекты, участники |
| **Floor Plan Service** | 50054 | Загрузка, распознавание планировок |
| **Scene Service** | 50055 | 3D сцены, элементы |
| **Branch Service** | 50056 | Ветки дизайна, версионирование |
| **AI Service** | 50057 | Распознавание, генерация, чат |
| **Compliance Service** | 50058 | Проверка норм СНиП/ЖК |
| **Request Service** | 50059 | Заявки на экспертов |
| **Notification Service** | 50060 | Уведомления, email, push |

Подробнее: [docs/architecture/microservices.md](architecture/microservices.md)

## Структура документации

```
docs/
├── README.md                      # Этот файл
├── SERVICE.md                     # Описание сервиса (для трекеров)
├── STACK.md                       # Технологический стек
├── DEVELOPMENT.md                 # План разработки и TODO
├── architecture/
│   ├── overview.md                # Детальная архитектура
│   ├── microservices.md           # Микросервисная архитектура
│   ├── database.md                # Схемы баз данных
│   ├── caching.md                 # Стратегии кэширования
│   └── security.md                # Безопасность
├── api/
│   ├── authentication.md          # Аутентификация
│   ├── users.md                   # API пользователей
│   ├── workspaces.md              # API воркспейсов
│   ├── floor-plans.md             # API планировок
│   ├── scenes.md                  # API 3D сцен
│   ├── branches.md                # API веток дизайна
│   ├── chat.md                    # API чата с AI
│   ├── compliance.md              # API проверки норм
│   ├── requests.md                # API заявок
│   └── notifications.md           # API уведомлений
├── models/
│   └── entities.md                # Доменные сущности
├── integration/
│   ├── openrouter.md              # Интеграция OpenRouter
│   └── storage.md                 # Файловое хранилище
└── deployment/
    ├── configuration.md           # Конфигурация
    ├── docker.md                  # Docker
    └── monitoring.md              # Мониторинг
```

## Быстрый старт

### Требования

- Go 1.22+
- Docker & Docker Compose
- Make

### Запуск для разработки

```bash
# Клонирование репозитория
git clone https://github.com/granula/api.git
cd api

# Копирование конфигурации
cp .env.example .env

# Запуск инфраструктуры
make docker-up

# Запуск API
make run

# API доступен на http://localhost:8080
# Swagger UI: http://localhost:8080/swagger
```

### Переменные окружения

```env
# Server
APP_ENV=development
APP_PORT=8080
APP_HOST=0.0.0.0

# PostgreSQL
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=granula
POSTGRES_PASSWORD=secret
POSTGRES_DB=granula

# MongoDB
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=granula

# Redis
REDIS_URL=redis://localhost:6379

# MinIO/S3
S3_ENDPOINT=localhost:9000
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_BUCKET=granula
S3_USE_SSL=false

# OpenRouter
OPENROUTER_API_KEY=sk-or-xxx
OPENROUTER_MODEL=anthropic/claude-sonnet-4

# JWT
JWT_SECRET=your-256-bit-secret
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=7d
```

## API Endpoints Overview

| Группа | Базовый путь | Описание |
|--------|--------------|----------|
| Auth | `/api/v1/auth` | Аутентификация |
| Users | `/api/v1/users` | Управление пользователями |
| Workspaces | `/api/v1/workspaces` | Воркспейсы проектов |
| Floor Plans | `/api/v1/workspaces/:id/floor-plans` | Планировки |
| Scenes | `/api/v1/workspaces/:id/scenes` | 3D сцены |
| Branches | `/api/v1/scenes/:id/branches` | Ветки дизайна |
| Chat | `/api/v1/scenes/:id/chat` | AI чат |
| Compliance | `/api/v1/compliance` | Проверка норм |
| Requests | `/api/v1/requests` | Заявки на экспертов |
| Notifications | `/api/v1/notifications` | Уведомления |

## Версионирование API

API использует семантическое версионирование в URL:

- **v1** — текущая стабильная версия
- Deprecated endpoints помечаются заголовком `X-Deprecated: true`
- Breaking changes только в major версиях

## Rate Limiting

| Тип запроса | Лимит | Окно |
|-------------|-------|------|
| Аутентификация | 10 req | 1 min |
| AI генерация | 20 req | 1 min |
| Загрузка файлов | 50 req | 1 hour |
| Общие запросы | 1000 req | 1 min |

## Контакты

- **Техническая поддержка**: api-support@granula.ru
- **Документация**: https://docs.granula.ru
- **Статус API**: https://status.granula.ru


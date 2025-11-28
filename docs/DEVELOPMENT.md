# План разработки Granula API

## Команда

| Роль | Зона ответственности |
|------|---------------------|
| **Developer 1 (Core)** | Инфраструктура, Auth, User, Workspace, Request, Notification |
| **Developer 2 (AI/3D)** | Floor Plan, Scene, Branch, AI, Compliance |

---

## Разделение микросервисов

### Developer 1 — Core Services

```
┌─────────────────────────────────────────────────────────────┐
│                    DEVELOPER 1 (Core)                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────────┐     ┌─────────────────┐                │
│  │   API Gateway   │     │  Auth Service   │                │
│  │   (Port 8080)   │     │  (Port 50051)   │                │
│  └─────────────────┘     └─────────────────┘                │
│                                                              │
│  ┌─────────────────┐     ┌─────────────────┐                │
│  │  User Service   │     │Workspace Service│                │
│  │  (Port 50052)   │     │  (Port 50053)   │                │
│  └─────────────────┘     └─────────────────┘                │
│                                                              │
│  ┌─────────────────┐     ┌─────────────────┐                │
│  │Request Service  │     │Notification Svc │                │
│  │  (Port 50059)   │     │  (Port 50060)   │                │
│  └─────────────────┘     └─────────────────┘                │
│                                                              │
│  Инфраструктура:                                            │
│  • Docker Compose                                            │
│  • CI/CD pipeline                                            │
│  • Shared proto/pkg                                          │
│  • PostgreSQL setup                                          │
│  • Redis setup                                               │
│  • MinIO setup                                               │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Итого сервисов:** 6
**Баз данных:** PostgreSQL (auth_db, users_db, workspaces_db, requests_db, notifications_db), Redis

---

### Developer 2 — AI/3D Services

```
┌─────────────────────────────────────────────────────────────┐
│                    DEVELOPER 2 (AI/3D)                       │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────────┐     ┌─────────────────┐                │
│  │Floor Plan Svc   │     │  Scene Service  │                │
│  │  (Port 50054)   │     │  (Port 50055)   │                │
│  └─────────────────┘     └─────────────────┘                │
│                                                              │
│  ┌─────────────────┐     ┌─────────────────┐                │
│  │ Branch Service  │     │   AI Service    │                │
│  │  (Port 50056)   │     │  (Port 50057)   │                │
│  └─────────────────┘     └─────────────────┘                │
│                                                              │
│  ┌─────────────────┐                                        │
│  │Compliance Svc   │                                        │
│  │  (Port 50058)   │                                        │
│  └─────────────────┘                                        │
│                                                              │
│  Интеграции:                                                │
│  • OpenRouter API                                            │
│  • MongoDB setup                                             │
│  • AI Worker Pool                                            │
│  • Compliance rules DB                                       │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Итого сервисов:** 5
**Баз данных:** PostgreSQL (floor_plans_db, compliance_db), MongoDB (scenes_db, branches_db, ai_db)

---

## Зависимости между сервисами

```
                    API Gateway
                         │
         ┌───────────────┼───────────────┐
         │               │               │
         ▼               ▼               ▼
    Auth Service    User Service    Workspace Service
         │               │               │
         │               │               ├──────────────┐
         │               │               │              │
         │               │               ▼              ▼
         │               │         Floor Plan ◄──► Scene Service
         │               │           Service         │
         │               │               │           │
         │               │               │           ▼
         │               │               │      Branch Service
         │               │               │           │
         │               │               ▼           │
         │               │          AI Service ◄────┘
         │               │               │
         │               │               ▼
         │               │      Compliance Service
         │               │               │
         └───────────────┼───────────────┘
                         │
                         ▼
              Request Service ◄──► Notification Service
```

### Порядок разработки (учёт зависимостей)

**Фаза 1 (параллельно):**
- Dev 1: Shared libs → API Gateway (заглушки) → Auth Service
- Dev 2: Shared libs → Compliance Service (rules DB)

**Фаза 2 (параллельно):**
- Dev 1: User Service → Workspace Service
- Dev 2: AI Service → Floor Plan Service

**Фаза 3 (параллельно):**
- Dev 1: Request Service → Notification Service
- Dev 2: Scene Service → Branch Service

**Фаза 4 (совместно):**
- Интеграция всех сервисов
- API Gateway полная настройка
- E2E тестирование

---

## Полный TODO проекта

### Фаза 0: Подготовка (совместно, 2 часа)

```
[ ] 0.1 Создание репозиториев
    [ ] 0.1.1 Создать monorepo структуру
    [ ] 0.1.2 Настроить .gitignore
    [ ] 0.1.3 Создать README.md
    
[ ] 0.2 Shared модуль
    [ ] 0.2.1 Создать shared/go.mod
    [ ] 0.2.2 Создать shared/proto/ структуру
    [ ] 0.2.3 Создать shared/pkg/logger
    [ ] 0.2.4 Создать shared/pkg/errors
    [ ] 0.2.5 Создать shared/pkg/validator
    [ ] 0.2.6 Создать shared/pkg/config
    [ ] 0.2.7 Создать shared/pkg/grpc (клиент хелперы)
    
[ ] 0.3 Инфраструктура
    [ ] 0.3.1 Создать docker-compose.dev.yml
    [ ] 0.3.2 Настроить PostgreSQL с несколькими DB
    [ ] 0.3.3 Настроить MongoDB
    [ ] 0.3.4 Настроить Redis
    [ ] 0.3.5 Настроить MinIO
    [ ] 0.3.6 Создать скрипт инициализации БД
```

---

### Developer 1: TODO

#### Auth Service (4 часа)

```
[ ] 1.1 Инициализация сервиса
    [ ] 1.1.1 Создать auth-service/go.mod
    [ ] 1.1.2 Создать структуру директорий
    [ ] 1.1.3 Создать Dockerfile
    [ ] 1.1.4 Создать конфигурацию (.env, config.go)

[ ] 1.2 Proto файлы
    [ ] 1.2.1 Определить auth.proto (сообщения)
    [ ] 1.2.2 Определить auth.proto (сервисы)
    [ ] 1.2.3 Сгенерировать Go код

[ ] 1.3 База данных
    [ ] 1.3.1 Создать миграцию: users table
    [ ] 1.3.2 Создать миграцию: refresh_tokens table
    [ ] 1.3.3 Создать миграцию: email_verifications table
    [ ] 1.3.4 Создать миграцию: password_resets table
    [ ] 1.3.5 Создать репозиторий UserRepository
    [ ] 1.3.6 Создать репозиторий TokenRepository

[ ] 1.4 Domain
    [ ] 1.4.1 Создать entity: User
    [ ] 1.4.2 Создать entity: RefreshToken
    [ ] 1.4.3 Создать domain errors

[ ] 1.5 Services
    [ ] 1.5.1 Создать AuthService interface
    [ ] 1.5.2 Реализовать Register
    [ ] 1.5.3 Реализовать Login
    [ ] 1.5.4 Реализовать ValidateToken
    [ ] 1.5.5 Реализовать RefreshToken
    [ ] 1.5.6 Реализовать Logout
    [ ] 1.5.7 Реализовать ResetPassword
    [ ] 1.5.8 Реализовать VerifyEmail

[ ] 1.6 JWT
    [ ] 1.6.1 Создать JWTService
    [ ] 1.6.2 Реализовать GenerateAccessToken
    [ ] 1.6.3 Реализовать GenerateRefreshToken
    [ ] 1.6.4 Реализовать ValidateToken

[ ] 1.7 OAuth
    [ ] 1.7.1 Создать OAuthProvider interface
    [ ] 1.7.2 Реализовать GoogleProvider
    [ ] 1.7.3 Реализовать YandexProvider
    [ ] 1.7.4 Реализовать OAuthLogin

[ ] 1.8 gRPC Server
    [ ] 1.8.1 Создать gRPC сервер
    [ ] 1.8.2 Реализовать AuthServiceServer
    [ ] 1.8.3 Добавить interceptors (logging, recovery)

[ ] 1.9 Тесты
    [ ] 1.9.1 Unit тесты AuthService
    [ ] 1.9.2 Unit тесты JWTService
    [ ] 1.9.3 Integration тесты
```

#### User Service (3 часа)

```
[ ] 2.1 Инициализация сервиса
    [ ] 2.1.1 Создать user-service/go.mod
    [ ] 2.1.2 Создать структуру директорий
    [ ] 2.1.3 Создать Dockerfile
    [ ] 2.1.4 Создать конфигурацию

[ ] 2.2 Proto файлы
    [ ] 2.2.1 Определить user.proto
    [ ] 2.2.2 Сгенерировать Go код

[ ] 2.3 База данных
    [ ] 2.3.1 Создать миграцию: user_profiles table
    [ ] 2.3.2 Создать миграцию: user_settings table
    [ ] 2.3.3 Создать миграцию: user_sessions table
    [ ] 2.3.4 Создать репозиторий ProfileRepository
    [ ] 2.3.5 Создать репозиторий SettingsRepository
    [ ] 2.3.6 Создать репозиторий SessionRepository

[ ] 2.4 Domain & Services
    [ ] 2.4.1 Создать entity: UserProfile
    [ ] 2.4.2 Создать entity: UserSettings
    [ ] 2.4.3 Создать UserService interface
    [ ] 2.4.4 Реализовать GetProfile
    [ ] 2.4.5 Реализовать UpdateProfile
    [ ] 2.4.6 Реализовать UploadAvatar
    [ ] 2.4.7 Реализовать GetSettings
    [ ] 2.4.8 Реализовать UpdateSettings
    [ ] 2.4.9 Реализовать GetSessions
    [ ] 2.4.10 Реализовать RevokeSession
    [ ] 2.4.11 Реализовать DeleteAccount

[ ] 2.5 Storage
    [ ] 2.5.1 Создать AvatarStorage
    [ ] 2.5.2 Реализовать загрузку в MinIO
    [ ] 2.5.3 Реализовать ресайз изображений

[ ] 2.6 gRPC Server
    [ ] 2.6.1 Реализовать UserServiceServer
```

#### Workspace Service (3 часа)

```
[ ] 3.1 Инициализация сервиса
    [ ] 3.1.1 Создать workspace-service/go.mod
    [ ] 3.1.2 Создать структуру директорий
    [ ] 3.1.3 Создать Dockerfile

[ ] 3.2 Proto файлы
    [ ] 3.2.1 Определить workspace.proto
    [ ] 3.2.2 Сгенерировать Go код

[ ] 3.3 База данных
    [ ] 3.3.1 Создать миграцию: workspaces table
    [ ] 3.3.2 Создать миграцию: workspace_members table
    [ ] 3.3.3 Создать миграцию: workspace_invites table
    [ ] 3.3.4 Создать WorkspaceRepository
    [ ] 3.3.5 Создать MemberRepository

[ ] 3.4 Domain & Services
    [ ] 3.4.1 Создать entity: Workspace
    [ ] 3.4.2 Создать entity: WorkspaceMember
    [ ] 3.4.3 Создать WorkspaceService interface
    [ ] 3.4.4 Реализовать CreateWorkspace
    [ ] 3.4.5 Реализовать GetWorkspace
    [ ] 3.4.6 Реализовать ListWorkspaces
    [ ] 3.4.7 Реализовать UpdateWorkspace
    [ ] 3.4.8 Реализовать DeleteWorkspace
    [ ] 3.4.9 Реализовать InviteMember
    [ ] 3.4.10 Реализовать RemoveMember
    [ ] 3.4.11 Реализовать UpdateMemberRole

[ ] 3.5 gRPC Server
    [ ] 3.5.1 Реализовать WorkspaceServiceServer
```

#### Request Service (3 часа)

```
[ ] 4.1 Инициализация сервиса
    [ ] 4.1.1 Создать request-service/go.mod
    [ ] 4.1.2 Создать структуру директорий
    [ ] 4.1.3 Создать Dockerfile

[ ] 4.2 Proto файлы
    [ ] 4.2.1 Определить request.proto
    [ ] 4.2.2 Сгенерировать Go код

[ ] 4.3 База данных
    [ ] 4.3.1 Создать миграцию: expert_requests table
    [ ] 4.3.2 Создать миграцию: request_documents table
    [ ] 4.3.3 Создать миграцию: status_history table
    [ ] 4.3.4 Создать RequestRepository

[ ] 4.4 Domain & Services
    [ ] 4.4.1 Создать entity: ExpertRequest
    [ ] 4.4.2 Создать entity: StatusHistory
    [ ] 4.4.3 Создать RequestService interface
    [ ] 4.4.4 Реализовать CreateRequest
    [ ] 4.4.5 Реализовать GetRequest
    [ ] 4.4.6 Реализовать ListRequests
    [ ] 4.4.7 Реализовать UpdateRequest
    [ ] 4.4.8 Реализовать CancelRequest
    [ ] 4.4.9 Реализовать status transitions
    [ ] 4.4.10 Реализовать UploadDocument

[ ] 4.5 Events
    [ ] 4.5.1 Публиковать request.created
    [ ] 4.5.2 Публиковать request.status_changed

[ ] 4.6 gRPC Server
    [ ] 4.6.1 Реализовать RequestServiceServer
```

#### Notification Service (3 часа)

```
[ ] 5.1 Инициализация сервиса
    [ ] 5.1.1 Создать notification-service/go.mod
    [ ] 5.1.2 Создать структуру директорий
    [ ] 5.1.3 Создать Dockerfile

[ ] 5.2 Proto файлы
    [ ] 5.2.1 Определить notification.proto
    [ ] 5.2.2 Сгенерировать Go код

[ ] 5.3 База данных
    [ ] 5.3.1 Создать миграцию: notifications table
    [ ] 5.3.2 Создать миграцию: notification_settings table
    [ ] 5.3.3 Создать миграцию: push_subscriptions table
    [ ] 5.3.4 Создать NotificationRepository

[ ] 5.4 Domain & Services
    [ ] 5.4.1 Создать entity: Notification
    [ ] 5.4.2 Создать NotificationService
    [ ] 5.4.3 Реализовать SendNotification
    [ ] 5.4.4 Реализовать GetNotifications
    [ ] 5.4.5 Реализовать MarkAsRead
    [ ] 5.4.6 Реализовать GetUnreadCount
    [ ] 5.4.7 Реализовать UpdateSettings

[ ] 5.5 Email
    [ ] 5.5.1 Создать EmailService
    [ ] 5.5.2 Создать email templates
    [ ] 5.5.3 Реализовать SendEmail

[ ] 5.6 Event Subscribers
    [ ] 5.6.1 Подписка на user.created
    [ ] 5.6.2 Подписка на request.status_changed
    [ ] 5.6.3 Подписка на workspace.invite

[ ] 5.7 gRPC Server
    [ ] 5.7.1 Реализовать NotificationServiceServer
```

#### API Gateway (5 часов)

```
[ ] 6.1 Инициализация
    [ ] 6.1.1 Создать api-gateway/go.mod
    [ ] 6.1.2 Создать структуру директорий
    [ ] 6.1.3 Создать Dockerfile
    [ ] 6.1.4 Настроить Fiber app

[ ] 6.2 gRPC Clients
    [ ] 6.2.1 Создать AuthClient
    [ ] 6.2.2 Создать UserClient
    [ ] 6.2.3 Создать WorkspaceClient
    [ ] 6.2.4 Создать FloorPlanClient
    [ ] 6.2.5 Создать SceneClient
    [ ] 6.2.6 Создать BranchClient
    [ ] 6.2.7 Создать AIClient
    [ ] 6.2.8 Создать ComplianceClient
    [ ] 6.2.9 Создать RequestClient
    [ ] 6.2.10 Создать NotificationClient

[ ] 6.3 Middleware
    [ ] 6.3.1 Создать RequestID middleware
    [ ] 6.3.2 Создать Logger middleware
    [ ] 6.3.3 Создать Recover middleware
    [ ] 6.3.4 Создать CORS middleware
    [ ] 6.3.5 Создать Auth middleware (JWT validation)
    [ ] 6.3.6 Создать RateLimit middleware

[ ] 6.4 REST Routes
    [ ] 6.4.1 Auth routes (/api/v1/auth/*)
    [ ] 6.4.2 User routes (/api/v1/users/*)
    [ ] 6.4.3 Workspace routes (/api/v1/workspaces/*)
    [ ] 6.4.4 Floor Plan routes
    [ ] 6.4.5 Scene routes
    [ ] 6.4.6 Branch routes
    [ ] 6.4.7 Chat routes
    [ ] 6.4.8 Compliance routes
    [ ] 6.4.9 Request routes
    [ ] 6.4.10 Notification routes

[ ] 6.5 WebSocket
    [ ] 6.5.1 Создать WebSocket hub
    [ ] 6.5.2 Реализовать connection handling
    [ ] 6.5.3 Реализовать authentication
    [ ] 6.5.4 Реализовать subscriptions
    [ ] 6.5.5 Интеграция с Notification Service

[ ] 6.6 Response Transformation
    [ ] 6.6.1 Создать response helpers
    [ ] 6.6.2 Создать error handlers
    [ ] 6.6.3 Создать pagination helpers

[ ] 6.7 Health & Metrics
    [ ] 6.7.1 Реализовать /health endpoint
    [ ] 6.7.2 Реализовать /metrics endpoint
```

---

### Developer 2: TODO

#### Compliance Service (3 часа)

```
[ ] 7.1 Инициализация сервиса
    [ ] 7.1.1 Создать compliance-service/go.mod
    [ ] 7.1.2 Создать структуру директорий
    [ ] 7.1.3 Создать Dockerfile

[ ] 7.2 Proto файлы
    [ ] 7.2.1 Определить compliance.proto
    [ ] 7.2.2 Сгенерировать Go код

[ ] 7.3 База данных
    [ ] 7.3.1 Создать миграцию: compliance_rules table
    [ ] 7.3.2 Создать миграцию: rule_categories table
    [ ] 7.3.3 Создать seed: базовые правила СНиП
    [ ] 7.3.4 Создать seed: правила ЖК РФ
    [ ] 7.3.5 Создать RuleRepository

[ ] 7.4 Domain & Services
    [ ] 7.4.1 Создать entity: ComplianceRule
    [ ] 7.4.2 Создать entity: ComplianceResult
    [ ] 7.4.3 Создать entity: Violation
    [ ] 7.4.4 Создать ComplianceService interface
    [ ] 7.4.5 Реализовать CheckCompliance
    [ ] 7.4.6 Реализовать CheckOperation
    [ ] 7.4.7 Реализовать GetRules
    [ ] 7.4.8 Реализовать GetRule
    [ ] 7.4.9 Реализовать GenerateReport

[ ] 7.5 Rule Engine
    [ ] 7.5.1 Создать RuleEngine
    [ ] 7.5.2 Реализовать load_bearing_check
    [ ] 7.5.3 Реализовать wet_zone_check
    [ ] 7.5.4 Реализовать min_area_check
    [ ] 7.5.5 Реализовать ventilation_check
    [ ] 7.5.6 Реализовать fire_safety_check

[ ] 7.6 gRPC Server
    [ ] 7.6.1 Реализовать ComplianceServiceServer
```

#### AI Service (5 часов)

```
[ ] 8.1 Инициализация сервиса
    [ ] 8.1.1 Создать ai-service/go.mod
    [ ] 8.1.2 Создать структуру директорий
    [ ] 8.1.3 Создать Dockerfile

[ ] 8.2 Proto файлы
    [ ] 8.2.1 Определить ai.proto (с streaming)
    [ ] 8.2.2 Сгенерировать Go код

[ ] 8.3 OpenRouter Client
    [ ] 8.3.1 Создать OpenRouterClient
    [ ] 8.3.2 Реализовать ChatCompletion
    [ ] 8.3.3 Реализовать ChatCompletionStream
    [ ] 8.3.4 Реализовать retry logic
    [ ] 8.3.5 Реализовать rate limiting

[ ] 8.4 MongoDB
    [ ] 8.4.1 Создать коллекцию chat_messages
    [ ] 8.4.2 Создать коллекцию ai_contexts
    [ ] 8.4.3 Создать ChatRepository
    [ ] 8.4.4 Создать ContextRepository

[ ] 8.5 Recognition
    [ ] 8.5.1 Создать RecognitionService
    [ ] 8.5.2 Создать system prompt для распознавания
    [ ] 8.5.3 Реализовать RecognizeFloorPlan
    [ ] 8.5.4 Реализовать parseRecognitionResult

[ ] 8.6 Generation
    [ ] 8.6.1 Создать GenerationService
    [ ] 8.6.2 Создать system prompt для генерации
    [ ] 8.6.3 Реализовать GenerateVariants
    [ ] 8.6.4 Интеграция с Branch Service (создание веток)

[ ] 8.7 Chat
    [ ] 8.7.1 Создать ChatService
    [ ] 8.7.2 Создать system prompt для чата
    [ ] 8.7.3 Реализовать SendMessage
    [ ] 8.7.4 Реализовать StreamResponse (gRPC streaming)
    [ ] 8.7.5 Реализовать context management
    [ ] 8.7.6 Реализовать GetHistory
    [ ] 8.7.7 Реализовать ClearHistory

[ ] 8.8 Worker Pool
    [ ] 8.8.1 Создать WorkerPool
    [ ] 8.8.2 Реализовать job queue
    [ ] 8.8.3 Реализовать graceful shutdown

[ ] 8.9 gRPC Server
    [ ] 8.9.1 Реализовать AIServiceServer
    [ ] 8.9.2 Реализовать streaming handlers
```

#### Floor Plan Service (3 часа)

```
[ ] 9.1 Инициализация сервиса
    [ ] 9.1.1 Создать floor-plan-service/go.mod
    [ ] 9.1.2 Создать структуру директорий
    [ ] 9.1.3 Создать Dockerfile

[ ] 9.2 Proto файлы
    [ ] 9.2.1 Определить floor_plan.proto
    [ ] 9.2.2 Сгенерировать Go код

[ ] 9.3 База данных
    [ ] 9.3.1 Создать миграцию: floor_plans table
    [ ] 9.3.2 Создать миграцию: processing_jobs table
    [ ] 9.3.3 Создать FloorPlanRepository

[ ] 9.4 Storage
    [ ] 9.4.1 Создать FloorPlanStorage (MinIO)
    [ ] 9.4.2 Реализовать Upload
    [ ] 9.4.3 Реализовать Download
    [ ] 9.4.4 Реализовать GenerateThumbnail

[ ] 9.5 Domain & Services
    [ ] 9.5.1 Создать entity: FloorPlan
    [ ] 9.5.2 Создать FloorPlanService
    [ ] 9.5.3 Реализовать Upload
    [ ] 9.5.4 Реализовать Get
    [ ] 9.5.5 Реализовать List
    [ ] 9.5.6 Реализовать Update
    [ ] 9.5.7 Реализовать Delete
    [ ] 9.5.8 Реализовать Process (вызов AI Service)
    [ ] 9.5.9 Реализовать GetStatus
    [ ] 9.5.10 Реализовать CreateScene (вызов Scene Service)

[ ] 9.6 Events
    [ ] 9.6.1 Публиковать floor_plan.uploaded
    [ ] 9.6.2 Публиковать floor_plan.processed

[ ] 9.7 gRPC Server
    [ ] 9.7.1 Реализовать FloorPlanServiceServer
```

#### Scene Service (4 часа)

```
[ ] 10.1 Инициализация сервиса
    [ ] 10.1.1 Создать scene-service/go.mod
    [ ] 10.1.2 Создать структуру директорий
    [ ] 10.1.3 Создать Dockerfile

[ ] 10.2 Proto файлы
    [ ] 10.2.1 Определить scene.proto
    [ ] 10.2.2 Сгенерировать Go код

[ ] 10.3 MongoDB
    [ ] 10.3.1 Создать коллекцию scenes
    [ ] 10.3.2 Создать индексы
    [ ] 10.3.3 Создать SceneRepository

[ ] 10.4 Domain
    [ ] 10.4.1 Создать entity: Scene
    [ ] 10.4.2 Создать entity: SceneElements
    [ ] 10.4.3 Создать entity: WallElement
    [ ] 10.4.4 Создать entity: RoomElement
    [ ] 10.4.5 Создать entity: FurnitureElement
    [ ] 10.4.6 Создать entity: UtilityElement
    [ ] 10.4.7 Создать entity: DisplaySettings

[ ] 10.5 Services
    [ ] 10.5.1 Создать SceneService interface
    [ ] 10.5.2 Реализовать Create
    [ ] 10.5.3 Реализовать Get
    [ ] 10.5.4 Реализовать List
    [ ] 10.5.5 Реализовать Update
    [ ] 10.5.6 Реализовать Delete
    [ ] 10.5.7 Реализовать UpdateElements
    [ ] 10.5.8 Реализовать ApplyOperation
    [ ] 10.5.9 Реализовать Duplicate
    [ ] 10.5.10 Реализовать CalculateStats

[ ] 10.6 Compliance Integration
    [ ] 10.6.1 Вызов CheckCompliance при изменениях
    [ ] 10.6.2 Сохранение ComplianceResult

[ ] 10.7 Events
    [ ] 10.7.1 Публиковать scene.created
    [ ] 10.7.2 Публиковать scene.updated

[ ] 10.8 gRPC Server
    [ ] 10.8.1 Реализовать SceneServiceServer
```

#### Branch Service (4 часа)

```
[ ] 11.1 Инициализация сервиса
    [ ] 11.1.1 Создать branch-service/go.mod
    [ ] 11.1.2 Создать структуру директорий
    [ ] 11.1.3 Создать Dockerfile

[ ] 11.2 Proto файлы
    [ ] 11.2.1 Определить branch.proto
    [ ] 11.2.2 Сгенерировать Go код

[ ] 11.3 MongoDB
    [ ] 11.3.1 Создать коллекцию branches
    [ ] 11.3.2 Создать индексы
    [ ] 11.3.3 Создать BranchRepository

[ ] 11.4 Domain
    [ ] 11.4.1 Создать entity: Branch
    [ ] 11.4.2 Создать entity: BranchDelta
    [ ] 11.4.3 Создать entity: BranchSnapshot
    [ ] 11.4.4 Создать entity: AIContext

[ ] 11.5 Services
    [ ] 11.5.1 Создать BranchService interface
    [ ] 11.5.2 Реализовать Create
    [ ] 11.5.3 Реализовать Get
    [ ] 11.5.4 Реализовать List
    [ ] 11.5.5 Реализовать GetTree
    [ ] 11.5.6 Реализовать Update
    [ ] 11.5.7 Реализовать Delete
    [ ] 11.5.8 Реализовать UpdateDelta
    [ ] 11.5.9 Реализовать Activate
    [ ] 11.5.10 Реализовать Compare
    [ ] 11.5.11 Реализовать Merge
    [ ] 11.5.12 Реализовать Duplicate

[ ] 11.6 Delta Engine
    [ ] 11.6.1 Создать DeltaEngine
    [ ] 11.6.2 Реализовать applyDelta
    [ ] 11.6.3 Реализовать calculateSnapshot
    [ ] 11.6.4 Реализовать diffBranches

[ ] 11.7 gRPC Server
    [ ] 11.7.1 Реализовать BranchServiceServer
```

---

### Фаза интеграции (совместно, 3 часа)

```
[ ] 12.1 API Gateway Integration
    [ ] 12.1.1 Подключить все gRPC клиенты
    [ ] 12.1.2 Настроить роутинг
    [ ] 12.1.3 Тестирование всех endpoints

[ ] 12.2 Event Bus
    [ ] 12.2.1 Проверить все pub/sub
    [ ] 12.2.2 Тестирование событий

[ ] 12.3 E2E тестирование
    [ ] 12.3.1 Тест: регистрация → создание воркспейса
    [ ] 12.3.2 Тест: загрузка планировки → распознавание
    [ ] 12.3.3 Тест: редактирование сцены → compliance
    [ ] 12.3.4 Тест: AI генерация → ветки
    [ ] 12.3.5 Тест: создание заявки → уведомления

[ ] 12.4 Docker Compose
    [ ] 12.4.1 Финальная настройка
    [ ] 12.4.2 Health checks
    [ ] 12.4.3 Logging

[ ] 12.5 Документация
    [ ] 12.5.1 Обновить API docs
    [ ] 12.5.2 Создать Postman collection
```

---

## Временная оценка

### Developer 1 (Core)

| Сервис | Часы |
|--------|------|
| Shared (совместно) | 1 |
| Auth Service | 4 |
| User Service | 3 |
| Workspace Service | 3 |
| Request Service | 3 |
| Notification Service | 3 |
| API Gateway | 5 |
| Интеграция (совместно) | 1.5 |
| **Итого** | **23.5** |

### Developer 2 (AI/3D)

| Сервис | Часы |
|--------|------|
| Shared (совместно) | 1 |
| Compliance Service | 3 |
| AI Service | 5 |
| Floor Plan Service | 3 |
| Scene Service | 4 |
| Branch Service | 4 |
| Интеграция (совместно) | 1.5 |
| **Итого** | **21.5** |

---

## Рекомендуемый порядок работы (48 часов хакатона)

### День 1 (0-12 часов)

**Developer 1:**
```
Час 0-1:   Shared libs (совместно)
Час 1-2:   Docker Compose, инфраструктура
Час 2-6:   Auth Service
Час 6-9:   User Service
Час 9-12:  Workspace Service
```

**Developer 2:**
```
Час 0-1:   Shared libs (совместно)
Час 1-4:   Compliance Service + rules seed
Час 4-9:   AI Service (OpenRouter, prompts)
Час 9-12:  Floor Plan Service
```

### День 1 (12-24 часа)

**Developer 1:**
```
Час 12-15: Request Service
Час 15-18: Notification Service
Час 18-24: API Gateway (routes, middleware)
```

**Developer 2:**
```
Час 12-16: Scene Service
Час 16-20: Branch Service
Час 20-24: AI Chat + Streaming
```

### День 2 (24-36 часов)

**Developer 1:**
```
Час 24-29: API Gateway (WebSocket, auth)
Час 29-32: Интеграция Auth/User/Workspace
Час 32-36: Интеграция Request/Notification
```

**Developer 2:**
```
Час 24-28: AI Generation + Branch integration
Час 28-32: Scene ↔ Compliance integration
Час 32-36: Floor Plan → Scene flow
```

### День 2 (36-48 часов)

**Совместно:**
```
Час 36-40: Полная интеграция Gateway
Час 40-44: E2E тестирование
Час 44-46: Bugfixes
Час 46-48: Docker Compose финализация, демо
```

---

## Синхронизация

### Daily Sync Points

| Время | Что синхронизируем |
|-------|-------------------|
| Час 1 | Shared libs готовы |
| Час 6 | Auth Service готов для Gateway |
| Час 12 | Базовые сервисы готовы |
| Час 24 | Все сервисы имеют базовый функционал |
| Час 36 | Интеграции готовы |
| Час 44 | Feature freeze, только bugfixes |

### Shared Contracts

Перед началом согласовать:
1. Proto файлы для всех сервисов
2. Event names и payloads
3. Error codes
4. Auth token structure


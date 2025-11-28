# Granula API — TODO Checklist

> **Команда:** 2 backend-разработчика  
> **Срок:** 48 часов (хакатон)  
> **Архитектура:** 11 микросервисов + gRPC

---

## 📋 Условные обозначения

- ⬜ — не начато
- 🔄 — в работе
- ✅ — готово
- 🧑‍💻 **D1** — Developer 1 (Core)
- 🧑‍💻 **D2** — Developer 2 (AI/3D)

---

## Фаза 0: Подготовка [совместно] — 2ч

| # | Задача | Ответственный | Статус |
|---|--------|---------------|--------|
| 0.1 | Создать monorepo структуру | D1 + D2 | ⬜ |
| 0.2 | Создать `shared/go.mod` | D1 | ⬜ |
| 0.3 | Создать `shared/proto/` (все proto файлы) | D1 + D2 | ⬜ |
| 0.4 | Создать `shared/pkg/logger` (Zap wrapper) | D1 | ⬜ |
| 0.5 | Создать `shared/pkg/errors` (domain errors) | D1 | ⬜ |
| 0.6 | Создать `shared/pkg/config` (Viper wrapper) | D1 | ⬜ |
| 0.7 | Создать `shared/pkg/grpc` (server/client helpers) | D2 | ⬜ |
| 0.8 | Настроить `docker-compose.dev.yml` | D1 | ⬜ |
| 0.9 | Создать `scripts/init-databases.sql` | D1 | ⬜ |
| 0.10 | Протестировать инфраструктуру | D1 + D2 | ⬜ |

---

## 🧑‍💻 Developer 1: Core Services

### Auth Service — 4ч

| # | Задача | Статус |
|---|--------|--------|
| 1.1 | Инициализация: go.mod, структура, Dockerfile | ⬜ |
| 1.2 | Proto: `auth.proto` → сгенерировать Go | ⬜ |
| 1.3 | Миграции: users, refresh_tokens, email_verif, password_resets | ⬜ |
| 1.4 | Repository: UserRepository, TokenRepository | ⬜ |
| 1.5 | Entity: User, RefreshToken | ⬜ |
| 1.6 | Service: Register, Login | ⬜ |
| 1.7 | Service: ValidateToken, RefreshToken | ⬜ |
| 1.8 | Service: Logout, ResetPassword, VerifyEmail | ⬜ |
| 1.9 | JWT: GenerateAccessToken, GenerateRefreshToken, Validate | ⬜ |
| 1.10 | OAuth: GoogleProvider, YandexProvider | ⬜ |
| 1.11 | gRPC Server: AuthServiceServer + interceptors | ⬜ |
| 1.12 | Unit tests | ⬜ |

### User Service — 3ч

| # | Задача | Статус |
|---|--------|--------|
| 2.1 | Инициализация: go.mod, структура, Dockerfile | ⬜ |
| 2.2 | Proto: `user.proto` → сгенерировать Go | ⬜ |
| 2.3 | Миграции: user_profiles, user_settings, user_sessions | ⬜ |
| 2.4 | Repository: ProfileRepo, SettingsRepo, SessionRepo | ⬜ |
| 2.5 | Service: GetProfile, UpdateProfile, UploadAvatar | ⬜ |
| 2.6 | Service: GetSettings, UpdateSettings | ⬜ |
| 2.7 | Service: GetSessions, RevokeSession, DeleteAccount | ⬜ |
| 2.8 | MinIO: AvatarStorage (upload, resize) | ⬜ |
| 2.9 | gRPC Server: UserServiceServer | ⬜ |

### Workspace Service — 3ч

| # | Задача | Статус |
|---|--------|--------|
| 3.1 | Инициализация: go.mod, структура, Dockerfile | ⬜ |
| 3.2 | Proto: `workspace.proto` → сгенерировать Go | ⬜ |
| 3.3 | Миграции: workspaces, workspace_members, workspace_invites | ⬜ |
| 3.4 | Repository: WorkspaceRepo, MemberRepo | ⬜ |
| 3.5 | Service: Create, Get, List, Update, Delete | ⬜ |
| 3.6 | Service: InviteMember, RemoveMember, UpdateRole | ⬜ |
| 3.7 | gRPC Server: WorkspaceServiceServer | ⬜ |

### Request Service — 3ч

| # | Задача | Статус |
|---|--------|--------|
| 4.1 | Инициализация: go.mod, структура, Dockerfile | ⬜ |
| 4.2 | Proto: `request.proto` → сгенерировать Go | ⬜ |
| 4.3 | Миграции: expert_requests, request_documents, status_history | ⬜ |
| 4.4 | Repository: RequestRepo | ⬜ |
| 4.5 | Service: Create, Get, List, Update, Cancel | ⬜ |
| 4.6 | Service: UpdateStatus, UploadDocument | ⬜ |
| 4.7 | Events: request.created, request.status_changed | ⬜ |
| 4.8 | gRPC Server: RequestServiceServer | ⬜ |

### Notification Service — 3ч

| # | Задача | Статус |
|---|--------|--------|
| 5.1 | Инициализация: go.mod, структура, Dockerfile | ⬜ |
| 5.2 | Proto: `notification.proto` → сгенерировать Go | ⬜ |
| 5.3 | Миграции: notifications, notification_settings, push_subs | ⬜ |
| 5.4 | Repository: NotificationRepo | ⬜ |
| 5.5 | Service: Send, GetList, MarkAsRead, GetUnreadCount | ⬜ |
| 5.6 | EmailService: templates, SendEmail | ⬜ |
| 5.7 | Redis Pub/Sub subscribers | ⬜ |
| 5.8 | gRPC Server: NotificationServiceServer | ⬜ |

### API Gateway — 5ч

| # | Задача | Статус |
|---|--------|--------|
| 6.1 | Инициализация: go.mod, структура, Dockerfile | ⬜ |
| 6.2 | gRPC Clients: все 10 сервисов | ⬜ |
| 6.3 | Middleware: RequestID, Logger, Recover, CORS | ⬜ |
| 6.4 | Middleware: Auth (JWT validation via Auth Service) | ⬜ |
| 6.5 | Middleware: RateLimit (Redis) | ⬜ |
| 6.6 | Routes: /api/v1/auth/* | ⬜ |
| 6.7 | Routes: /api/v1/users/* | ⬜ |
| 6.8 | Routes: /api/v1/workspaces/* | ⬜ |
| 6.9 | Routes: /api/v1/floor-plans/* | ⬜ |
| 6.10 | Routes: /api/v1/scenes/*, /api/v1/branches/* | ⬜ |
| 6.11 | Routes: /api/v1/chat/* (с streaming) | ⬜ |
| 6.12 | Routes: /api/v1/compliance/*, /api/v1/requests/* | ⬜ |
| 6.13 | Routes: /api/v1/notifications/* | ⬜ |
| 6.14 | WebSocket Hub: notifications, chat streaming | ⬜ |
| 6.15 | Health: /health, /metrics | ⬜ |

---

## 🧑‍💻 Developer 2: AI/3D Services

### Compliance Service — 3ч

| # | Задача | Статус |
|---|--------|--------|
| 7.1 | Инициализация: go.mod, структура, Dockerfile | ⬜ |
| 7.2 | Proto: `compliance.proto` → сгенерировать Go | ⬜ |
| 7.3 | Миграции: compliance_rules, rule_categories | ⬜ |
| 7.4 | Seeds: базовые правила СНиП, правила ЖК РФ | ⬜ |
| 7.5 | Repository: RuleRepo | ⬜ |
| 7.6 | Entity: ComplianceRule, Violation, ComplianceResult | ⬜ |
| 7.7 | Service: CheckCompliance, CheckOperation | ⬜ |
| 7.8 | Service: GetRules, GetRule, GenerateReport | ⬜ |
| 7.9 | RuleEngine: load_bearing, wet_zone, min_area, fire_safety | ⬜ |
| 7.10 | gRPC Server: ComplianceServiceServer | ⬜ |

### AI Service — 5ч

| # | Задача | Статус |
|---|--------|--------|
| 8.1 | Инициализация: go.mod, структура, Dockerfile | ⬜ |
| 8.2 | Proto: `ai.proto` (с streaming) → сгенерировать Go | ⬜ |
| 8.3 | OpenRouter Client: ChatCompletion, Stream, retry | ⬜ |
| 8.4 | MongoDB: chat_messages, ai_contexts collections | ⬜ |
| 8.5 | Repository: ChatRepo, ContextRepo | ⬜ |
| 8.6 | RecognitionService: system prompt, RecognizeFloorPlan | ⬜ |
| 8.7 | GenerationService: system prompt, GenerateVariants | ⬜ |
| 8.8 | ChatService: SendMessage, StreamResponse | ⬜ |
| 8.9 | ChatService: GetHistory, ClearHistory, ResetContext | ⬜ |
| 8.10 | Worker Pool: job queue, graceful shutdown | ⬜ |
| 8.11 | gRPC Server: AIServiceServer (with streaming) | ⬜ |

### Floor Plan Service — 3ч

| # | Задача | Статус |
|---|--------|--------|
| 9.1 | Инициализация: go.mod, структура, Dockerfile | ⬜ |
| 9.2 | Proto: `floor_plan.proto` → сгенерировать Go | ⬜ |
| 9.3 | Миграции: floor_plans, processing_jobs | ⬜ |
| 9.4 | Repository: FloorPlanRepo | ⬜ |
| 9.5 | MinIO: FloorPlanStorage (upload, download, thumbnail) | ⬜ |
| 9.6 | Service: Upload, Get, List, Update, Delete | ⬜ |
| 9.7 | Service: Process (→ AI Service) | ⬜ |
| 9.8 | Service: GetStatus, CreateScene (→ Scene Service) | ⬜ |
| 9.9 | Events: floor_plan.uploaded, floor_plan.processed | ⬜ |
| 9.10 | gRPC Server: FloorPlanServiceServer | ⬜ |

### Scene Service — 4ч

| # | Задача | Статус |
|---|--------|--------|
| 10.1 | Инициализация: go.mod, структура, Dockerfile | ⬜ |
| 10.2 | Proto: `scene.proto` → сгенерировать Go | ⬜ |
| 10.3 | MongoDB: scenes collection + indexes | ⬜ |
| 10.4 | Repository: SceneRepo | ⬜ |
| 10.5 | Entity: Scene, SceneElements, Wall, Room, Furniture, Utility | ⬜ |
| 10.6 | Service: Create, Get, List, Update, Delete | ⬜ |
| 10.7 | Service: UpdateElements, ApplyOperation | ⬜ |
| 10.8 | Service: Duplicate, CalculateStats | ⬜ |
| 10.9 | Compliance Integration: CheckCompliance при изменениях | ⬜ |
| 10.10 | Events: scene.created, scene.updated | ⬜ |
| 10.11 | gRPC Server: SceneServiceServer | ⬜ |

### Branch Service — 4ч

| # | Задача | Статус |
|---|--------|--------|
| 11.1 | Инициализация: go.mod, структура, Dockerfile | ⬜ |
| 11.2 | Proto: `branch.proto` → сгенерировать Go | ⬜ |
| 11.3 | MongoDB: branches collection + indexes | ⬜ |
| 11.4 | Repository: BranchRepo | ⬜ |
| 11.5 | Entity: Branch, BranchDelta, BranchSnapshot, AIContext | ⬜ |
| 11.6 | Service: Create, Get, List, GetTree, Update, Delete | ⬜ |
| 11.7 | Service: UpdateDelta, Activate | ⬜ |
| 11.8 | Service: Compare, Merge, Duplicate | ⬜ |
| 11.9 | DeltaEngine: applyDelta, calculateSnapshot, diffBranches | ⬜ |
| 11.10 | gRPC Server: BranchServiceServer | ⬜ |

---

## Фаза интеграции [совместно] — 3ч

| # | Задача | Ответственный | Статус |
|---|--------|---------------|--------|
| 12.1 | API Gateway: подключить все gRPC клиенты | D1 | ⬜ |
| 12.2 | Проверить все pub/sub события | D1 + D2 | ⬜ |
| 12.3 | E2E: регистрация → воркспейс | D1 | ⬜ |
| 12.4 | E2E: планировка → распознавание → сцена | D2 | ⬜ |
| 12.5 | E2E: редактирование → compliance | D2 | ⬜ |
| 12.6 | E2E: AI генерация → ветки | D2 | ⬜ |
| 12.7 | E2E: заявка → уведомления | D1 | ⬜ |
| 12.8 | Docker Compose: финальная настройка | D1 | ⬜ |
| 12.9 | Health checks всех сервисов | D1 + D2 | ⬜ |
| 12.10 | Демо-прогон | D1 + D2 | ⬜ |

---

## 📊 Прогресс

### Developer 1
```
Auth Service:       [░░░░░░░░░░] 0/12
User Service:       [░░░░░░░░░░] 0/9
Workspace Service:  [░░░░░░░░░░] 0/7
Request Service:    [░░░░░░░░░░] 0/8
Notification Svc:   [░░░░░░░░░░] 0/8
API Gateway:        [░░░░░░░░░░] 0/15
─────────────────────────────────────
TOTAL:              [░░░░░░░░░░] 0/59
```

### Developer 2
```
Compliance Service: [░░░░░░░░░░] 0/10
AI Service:         [░░░░░░░░░░] 0/11
Floor Plan Service: [░░░░░░░░░░] 0/10
Scene Service:      [░░░░░░░░░░] 0/11
Branch Service:     [░░░░░░░░░░] 0/10
─────────────────────────────────────
TOTAL:              [░░░░░░░░░░] 0/52
```

---

## ⏰ Timeline (48 часов)

```
Час 0-1:    Подготовка (совместно)
Час 1-6:    D1: Auth Service | D2: Compliance Service
Час 6-9:    D1: User Service | D2: AI Service (начало)
Час 9-12:   D1: Workspace Svc | D2: AI Service (продолжение)
Час 12-15:  D1: Request Svc | D2: Floor Plan Service
Час 15-18:  D1: Notification | D2: Scene Service (начало)
Час 18-24:  D1: Gateway (начало) | D2: Scene Service (конец)
Час 24-29:  D1: Gateway (конец) | D2: Branch Service
Час 29-36:  D1: Integration D1 | D2: AI + Branch integration
Час 36-44:  Полная интеграция (совместно)
Час 44-48:  Bugfixes, демо (совместно)
```

---

## 🔗 Sync Points

| Час | Checkpoint |
|-----|------------|
| 1 | ✓ Shared libs готовы, docker-compose работает |
| 6 | ✓ Auth Service работает (Login, Register, ValidateToken) |
| 12 | ✓ Базовые сервисы D1 готовы, AI Service распознаёт |
| 18 | ✓ Все CRUD сервисы работают |
| 24 | ✓ Все сервисы имеют базовый функционал |
| 36 | ✓ Все интеграции работают |
| 44 | ⚠️ FEATURE FREEZE — только bugfixes |
| 48 | 🎯 Демо готово |


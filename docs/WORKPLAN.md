# 📋 GRANULA API - ПОЛНЫЙ ПЛАН РАБОТ

> **Дата обновления:** 29.11.2024  
> **Статус проекта:** В разработке (80% готово)  
> **Следующий этап:** Финализация и локальный деплой

---

## 📊 ТЕКУЩИЙ СТАТУС СИСТЕМЫ

### Результаты анализа (29.11.2024):

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            СТАТУС КОМПИЛЯЦИИ                                     │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  SHARED MODULE                               D1 SERVICES (Core)                  │
│  ─────────────────                           ─────────────────────               │
│  [✅] shared/pkg/*                           [✅] auth-service                   │
│  [⚠️] shared/gen/* (ПУСТО!)                  [❌] user-service (go.sum)          │
│  [✅] shared/proto/*                         [❌] api-gateway (go.sum)           │
│                                              [❌] notification-service (go.sum)  │
│                                              [❌] workspace-service (НЕТ)        │
│                                              [❌] request-service (НЕТ)          │
│                                                                                  │
│  D2 SERVICES (AI/3D)                         ТЕСТЫ                               │
│  ────────────────────                        ───────────                         │
│  [✅] compliance-service                     [✅] compliance-service/engine_test │
│  [✅] ai-service                             [✅] ai-service/client_test         │
│  [✅] floorplan-service                      [✅] ai-service/chat_test           │
│  [✅] scene-service                          [✅] floorplan-service/entity_test  │
│  [✅] branch-service                         [✅] scene-service/entity_test      │
│                                              [✅] branch-service/entity_test     │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘

Легенда: [✅] Готово/Компилируется  [⚠️] Частично  [❌] Требует исправления
```

---

## 🎯 КРИТИЧЕСКИЕ ПРОБЛЕМЫ (ИСПРАВИТЬ НЕМЕДЛЕННО)

### 1. Proto файлы не сгенерированы
```
Проблема: shared/gen/ пустая директория
Решение: Запустить protoc для всех .proto файлов
Команда: make proto (или вручную через protoc)
```

### 2. D1 сервисы не компилируются (missing go.sum)
```
Проблема: user-service, api-gateway, notification-service отсутствуют go.sum
Решение: go mod tidy в каждом сервисе
```

### 3. Отсутствующие сервисы
```
Проблема: workspace-service и request-service не существуют
Решение: Создать с нуля по шаблону других сервисов
```

---

## 📝 ДЕТАЛЬНЫЙ ПЛАН РАБОТ

### ЭТАП 1: Исправление критических проблем (2-3 часа)

#### 1.1 Генерация Proto файлов
```powershell
# Windows
cd R:\granula\api
mkdir -p shared/gen

# Генерация для каждого proto файла
$protos = @(
    "common/v1/common.proto",
    "auth/v1/auth.proto",
    "user/v1/user.proto",
    "workspace/v1/workspace.proto",
    "floor_plan/v1/floor_plan.proto",
    "scene/v1/scene.proto",
    "branch/v1/branch.proto",
    "ai/v1/ai.proto",
    "compliance/v1/compliance.proto",
    "request/v1/request.proto",
    "notification/v1/notification.proto"
)

foreach ($p in $protos) {
    protoc --go_out=shared/gen --go_opt=paths=source_relative `
           --go-grpc_out=shared/gen --go-grpc_opt=paths=source_relative `
           -I shared/proto shared/proto/$p
}
```

#### 1.2 Исправление go.sum для D1 сервисов
```powershell
# Для каждого D1 сервиса
cd user-service && go mod tidy && cd ..
cd api-gateway && go mod tidy && cd ..
cd notification-service && go mod tidy && cd ..
```

#### 1.3 Исправление импортов в auth-service
```
Файл: auth-service/internal/service/auth.go
Проблема: Неправильные импорты (github.com/xiiisorate/github.com/xiiisorate/...)
Исправить на: github.com/xiiisorate/granula_api/shared/pkg/errors
```

---

### ЭТАП 2: Создание недостающих сервисов (4-5 часов)

#### 2.1 Workspace Service
```
workspace-service/
├── cmd/server/main.go
├── go.mod
├── Dockerfile
├── internal/
│   ├── config/config.go
│   ├── domain/entity/workspace.go
│   ├── repository/postgres/workspace_repository.go
│   ├── service/workspace_service.go
│   └── grpc/server.go
└── migrations/
    ├── 000001_create_workspaces.up.sql
    └── 000001_create_workspaces.down.sql
```

**Методы WorkspaceService:**
- CreateWorkspace(ctx, userID, name, description) → Workspace
- GetWorkspace(ctx, id) → Workspace
- ListWorkspaces(ctx, userID, pagination) → []Workspace
- UpdateWorkspace(ctx, id, name, description) → Workspace
- DeleteWorkspace(ctx, id) → error
- AddMember(ctx, workspaceID, userID, role) → Member
- RemoveMember(ctx, workspaceID, userID) → error
- UpdateMemberRole(ctx, workspaceID, userID, role) → Member
- ListMembers(ctx, workspaceID) → []Member

#### 2.2 Request Service
```
request-service/
├── cmd/server/main.go
├── go.mod
├── Dockerfile
├── internal/
│   ├── config/config.go
│   ├── domain/entity/request.go
│   ├── repository/postgres/request_repository.go
│   ├── service/request_service.go
│   └── grpc/server.go
└── migrations/
    ├── 000001_create_requests.up.sql
    └── 000001_create_requests.down.sql
```

**Методы RequestService:**
- CreateRequest(ctx, workspaceID, title, description, category) → Request
- GetRequest(ctx, id) → Request
- ListRequests(ctx, workspaceID, status, pagination) → []Request
- UpdateRequest(ctx, id, title, description) → Request
- CancelRequest(ctx, id) → error
- AssignExpert(ctx, requestID, expertID) → Request
- UpdateStatus(ctx, requestID, status, comment) → Request

---

### ЭТАП 3: Dockerfiles для D2 сервисов (1-2 часа)

#### Шаблон Dockerfile для D2 сервисов:
```dockerfile
# =============================================================================
# Build stage
# =============================================================================
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy shared module
COPY shared/ ./shared/

# Copy service
COPY ${SERVICE_NAME}/ ./${SERVICE_NAME}/

WORKDIR /app/${SERVICE_NAME}

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/service ./cmd/server

# =============================================================================
# Production stage
# =============================================================================
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1000 granula && \
    adduser -u 1000 -G granula -s /bin/sh -D granula

WORKDIR /app
COPY --from=builder /app/service .
COPY --from=builder /app/${SERVICE_NAME}/migrations ./migrations 2>/dev/null || true

RUN chown -R granula:granula /app
USER granula

EXPOSE 50054

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT}/health || exit 1

ENTRYPOINT ["./service"]
```

**Создать для:**
- [ ] compliance-service/Dockerfile (порт 50058)
- [ ] ai-service/Dockerfile (порт 50057)
- [ ] floorplan-service/Dockerfile (порт 50054)
- [ ] scene-service/Dockerfile (порт 50055)
- [ ] branch-service/Dockerfile (порт 50056)

---

### ЭТАП 4: API Gateway интеграция (3-4 часа)

#### 4.1 Добавить gRPC клиенты
```go
// api-gateway/internal/grpc/clients.go
package grpc

type Clients struct {
    Auth         authv1.AuthServiceClient
    User         userv1.UserServiceClient
    Workspace    workspacev1.WorkspaceServiceClient
    FloorPlan    floorplanv1.FloorPlanServiceClient
    Scene        scenev1.SceneServiceClient
    Branch       branchv1.BranchServiceClient
    AI           aiv1.AIServiceClient
    Compliance   compliancev1.ComplianceServiceClient
    Request      requestv1.RequestServiceClient
    Notification notificationv1.NotificationServiceClient
}
```

#### 4.2 Реализовать HTTP handlers
```
api-gateway/internal/handlers/
├── auth.go          ← Существует (TODO: gRPC integration)
├── user.go          ← Существует (TODO: gRPC integration)
├── notification.go  ← Существует (TODO: gRPC integration)
├── workspace.go     ← СОЗДАТЬ
├── floorplan.go     ← СОЗДАТЬ
├── scene.go         ← СОЗДАТЬ
├── branch.go        ← СОЗДАТЬ
├── ai.go            ← СОЗДАТЬ (включая streaming)
├── compliance.go    ← СОЗДАТЬ
└── request.go       ← СОЗДАТЬ
```

#### 4.3 Настроить роутинг
```go
// Routes
api := app.Group("/api/v1")

// Auth (public)
api.Post("/auth/register", handlers.Register)
api.Post("/auth/login", handlers.Login)
api.Post("/auth/refresh", handlers.RefreshToken)
api.Post("/auth/logout", handlers.Logout)

// Protected routes
protected := api.Use(middleware.Auth(cfg))

// Users
protected.Get("/users/me", handlers.GetProfile)
protected.Put("/users/me", handlers.UpdateProfile)
protected.Post("/users/me/avatar", handlers.UploadAvatar)

// Workspaces
protected.Post("/workspaces", handlers.CreateWorkspace)
protected.Get("/workspaces", handlers.ListWorkspaces)
protected.Get("/workspaces/:id", handlers.GetWorkspace)
protected.Put("/workspaces/:id", handlers.UpdateWorkspace)
protected.Delete("/workspaces/:id", handlers.DeleteWorkspace)

// Floor Plans
protected.Post("/workspaces/:id/floor-plans", handlers.UploadFloorPlan)
protected.Get("/workspaces/:id/floor-plans", handlers.ListFloorPlans)
protected.Get("/floor-plans/:id", handlers.GetFloorPlan)
protected.Delete("/floor-plans/:id", handlers.DeleteFloorPlan)
protected.Post("/floor-plans/:id/process", handlers.ProcessFloorPlan)

// Scenes
protected.Get("/workspaces/:id/scenes", handlers.ListScenes)
protected.Post("/scenes", handlers.CreateScene)
protected.Get("/scenes/:id", handlers.GetScene)
protected.Put("/scenes/:id", handlers.UpdateScene)
protected.Delete("/scenes/:id", handlers.DeleteScene)

// Elements
protected.Get("/scenes/:id/elements", handlers.ListElements)
protected.Post("/scenes/:id/elements", handlers.CreateElement)
protected.Put("/elements/:id", handlers.UpdateElement)
protected.Delete("/elements/:id", handlers.DeleteElement)

// Branches
protected.Get("/scenes/:id/branches", handlers.ListBranches)
protected.Post("/branches", handlers.CreateBranch)
protected.Get("/branches/:id", handlers.GetBranch)
protected.Post("/branches/:id/merge", handlers.MergeBranch)

// AI
protected.Post("/ai/chat", handlers.SendChatMessage)
protected.Get("/ai/chat/:scene_id/history", handlers.GetChatHistory)
protected.Post("/ai/recognize", handlers.RecognizeFloorPlan)
protected.Post("/ai/generate-variants", handlers.GenerateVariants)

// Compliance
protected.Post("/compliance/check", handlers.CheckCompliance)
protected.Get("/compliance/rules", handlers.GetRules)

// Requests
protected.Post("/requests", handlers.CreateRequest)
protected.Get("/requests", handlers.ListRequests)
protected.Get("/requests/:id", handlers.GetRequest)
protected.Put("/requests/:id/status", handlers.UpdateRequestStatus)

// Notifications
protected.Get("/notifications", handlers.GetNotifications)
protected.Post("/notifications/:id/read", handlers.MarkAsRead)
protected.Get("/notifications/unread-count", handlers.GetUnreadCount)
```

---

### ЭТАП 5: Тестирование (2-3 часа)

#### 5.1 Unit тесты для D1 сервисов
```
auth-service/internal/service/auth_test.go
user-service/internal/service/user_test.go
notification-service/internal/service/notification_test.go
workspace-service/internal/service/workspace_test.go
request-service/internal/service/request_test.go
```

#### 5.2 Запуск всех тестов
```powershell
# Запуск тестов по всем сервисам
go test ./... -v -cover

# С coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

---

### ЭТАП 6: Локальный деплой (1-2 часа)

#### 6.1 Запуск инфраструктуры
```powershell
# Запуск только БД и кэша
docker-compose up -d postgres mongodb redis minio

# Проверка
docker-compose ps
```

#### 6.2 Запуск миграций
```powershell
# Auth DB
migrate -path auth-service/migrations -database "postgres://granula:granula_secret@localhost:5432/auth_db?sslmode=disable" up

# Users DB
migrate -path user-service/migrations -database "postgres://granula:granula_secret@localhost:5432/users_db?sslmode=disable" up

# И так далее для всех сервисов...
```

#### 6.3 Запуск всех сервисов
```powershell
# Сборка и запуск
docker-compose up -d --build

# Проверка логов
docker-compose logs -f api-gateway

# Проверка здоровья
curl http://localhost:8080/health
```

---

## 📋 ЧЕКЛИСТ ЗАВЕРШЕНИЯ

### Критический путь (обязательно):
- [ ] Proto генерация работает
- [ ] Все сервисы компилируются
- [ ] Docker Compose запускается
- [ ] API Gateway роутит запросы
- [ ] Регистрация/логин работает
- [ ] CRUD воркспейсов работает

### Полный функционал:
- [ ] Загрузка и распознавание планировок
- [ ] Редактирование сцены (элементы)
- [ ] Ветвление и слияние
- [ ] AI чат работает
- [ ] Compliance проверка работает
- [ ] Заявки на экспертов работают
- [ ] Уведомления доставляются

### Качество кода:
- [ ] Unit тесты для всех сервисов (>70% coverage)
- [ ] Все файлы с комментариями и docstrings
- [ ] Линтер проходит без ошибок
- [ ] Логирование настроено
- [ ] Error handling везде

---

## 🔧 ПОЛЕЗНЫЕ КОМАНДЫ

```powershell
# Сборка всех сервисов
make build

# Запуск тестов
make test

# Линтинг
make lint

# Docker Compose
make docker-up
make docker-down
make docker-logs

# Миграции
make migrate-all-up

# Proto генерация
make proto
```

---

## 📞 КОНТАКТЫ И РЕСУРСЫ

- **Репозиторий:** https://github.com/xiiisorate/granula_api
- **Ветка разработки:** dev/shared
- **Документация:** ./docs/
- **Proto файлы:** ./shared/proto/

# Архитектура системы Granula API

## Обзор

Granula API построен на **микросервисной архитектуре** с 11 независимыми сервисами. Каждый сервис следует принципам **Clean Architecture** с чётким разделением слоёв. Система спроектирована для независимого масштабирования, параллельной разработки и высокой доступности.

## Микросервисная архитектура

```
                              ┌─────────────────┐
                              │  Load Balancer  │
                              └────────┬────────┘
                                       │ HTTP/WebSocket
                                       ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                              API GATEWAY (:8080)                              │
│                    REST → gRPC • JWT validation • Rate limiting               │
└──────────────────────────────────────┬───────────────────────────────────────┘
                                       │ gRPC (protobuf)
       ┌───────────────┬───────────────┼───────────────┬───────────────┐
       │               │               │               │               │
       ▼               ▼               ▼               ▼               ▼
┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│    Auth     │ │    User     │ │  Workspace  │ │ Floor Plan  │ │    Scene    │
│   Service   │ │   Service   │ │   Service   │ │   Service   │ │   Service   │
│   :50051    │ │   :50052    │ │   :50053    │ │   :50054    │ │   :50055    │
└──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └──────┬──────┘
       │               │               │               │               │
       └───────────────┴───────────────┼───────────────┴───────────────┘
                                       ▼
                    ┌─────────────────────────────────────┐
                    │  PostgreSQL │ MongoDB │ Redis │ S3  │
                    └─────────────────────────────────────┘
```

### 11 Микросервисов

| Сервис | Порт | База данных | Описание |
|--------|------|-------------|----------|
| API Gateway | 8080 | Redis | REST API, JWT validation |
| Auth Service | 50051 | PostgreSQL | Регистрация, логин, OAuth |
| User Service | 50052 | PostgreSQL + S3 | Профили, аватары |
| Workspace Service | 50053 | PostgreSQL | Проекты, участники |
| Floor Plan Service | 50054 | PostgreSQL + S3 | Планировки |
| Scene Service | 50055 | MongoDB | 3D сцены |
| Branch Service | 50056 | MongoDB | Ветки дизайна |
| AI Service | 50057 | MongoDB | Распознавание, генерация |
| Compliance Service | 50058 | PostgreSQL | Проверка норм |
| Request Service | 50059 | PostgreSQL | Заявки |
| Notification Service | 50060 | PostgreSQL | Уведомления |

## Принципы проектирования

### 1. Clean Architecture (внутри каждого сервиса)

```
┌─────────────────────────────────────────────────────────────┐
│                      gRPC Handlers                           │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                    Use Cases                             ││
│  │  ┌─────────────────────────────────────────────────────┐││
│  │  │               Domain Entities                        │││
│  │  │                                                      │││
│  │  │   User, Workspace, Scene, Branch, Request           │││
│  │  │                                                      │││
│  │  └─────────────────────────────────────────────────────┘││
│  │                                                          ││
│  │   UserService, SceneService, ChatService, etc.           ││
│  │                                                          ││
│  └─────────────────────────────────────────────────────────┘│
│                                                              │
│   PostgreSQL, MongoDB, Redis, S3, gRPC Clients               │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### 2. Dependency Injection

Все зависимости инжектируются через конструкторы, что обеспечивает:
- Тестируемость через моки
- Гибкость замены реализаций
- Явные зависимости компонентов

### 3. Repository Pattern

Абстракция доступа к данным через интерфейсы:

```go
// internal/domain/repository/user.go

// UserRepository определяет контракт для работы с пользователями.
// Имплементации могут использовать различные хранилища (PostgreSQL, MongoDB).
type UserRepository interface {
    // Create создаёт нового пользователя в системе.
    // Возвращает ErrDuplicateEmail если email уже существует.
    Create(ctx context.Context, user *entity.User) error
    
    // GetByID возвращает пользователя по идентификатору.
    // Возвращает ErrNotFound если пользователь не найден.
    GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
    
    // GetByEmail возвращает пользователя по email.
    // Возвращает ErrNotFound если пользователь не найден.
    GetByEmail(ctx context.Context, email string) (*entity.User, error)
    
    // Update обновляет данные пользователя.
    // Обновляет только непустые поля в entity.
    Update(ctx context.Context, user *entity.User) error
    
    // Delete удаляет пользователя (soft delete).
    // Устанавливает deleted_at, данные сохраняются.
    Delete(ctx context.Context, id uuid.UUID) error
}
```

## Структура проекта (Monorepo)

```
granula/
├── api-gateway/                    # API Gateway (HTTP → gRPC)
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── handler/http/           # REST handlers
│   │   ├── middleware/             # Auth, CORS, RateLimit
│   │   ├── grpc/                   # gRPC clients
│   │   └── websocket/              # WebSocket hub
│   ├── Dockerfile
│   └── go.mod
│
├── auth-service/                   # Authentication Service
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── domain/entity/
│   │   ├── repository/postgres/
│   │   ├── service/
│   │   └── grpc/                   # gRPC server
│   ├── migrations/
│   ├── Dockerfile
│   └── go.mod
│
├── user-service/                   # User Service
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── domain/entity/
│   │   ├── repository/postgres/
│   │   ├── service/
│   │   ├── storage/                # MinIO/S3 avatars
│   │   └── grpc/
│   ├── migrations/
│   ├── Dockerfile
│   └── go.mod
│
├── workspace-service/              # Workspace Service
│   ├── cmd/server/main.go
│   ├── internal/...
│   ├── migrations/
│   ├── Dockerfile
│   └── go.mod
│
├── floor-plan-service/             # Floor Plan Service
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── grpc/                   # gRPC server + AI client
│   │   └── storage/                # MinIO/S3 floor plans
│   ├── migrations/
│   ├── Dockerfile
│   └── go.mod
│
├── scene-service/                  # Scene Service (MongoDB)
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── repository/mongodb/
│   │   └── grpc/                   # + Compliance client
│   ├── Dockerfile
│   └── go.mod
│
├── branch-service/                 # Branch Service (MongoDB)
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── repository/mongodb/
│   │   └── engine/                 # Delta engine
│   ├── Dockerfile
│   └── go.mod
│
├── ai-service/                     # AI Service (OpenRouter)
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── openrouter/             # OpenRouter client
│   │   ├── recognition/            # Floor plan recognition
│   │   ├── generation/             # Variant generation
│   │   ├── chat/                   # Chat with streaming
│   │   └── worker/                 # Worker pool
│   ├── Dockerfile
│   └── go.mod
│
├── compliance-service/             # Compliance Service
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── engine/                 # Rule engine
│   │   └── repository/postgres/
│   ├── migrations/
│   │   └── seeds/                  # SNiP rules seeds
│   ├── Dockerfile
│   └── go.mod
│
├── request-service/                # Request Service
│   ├── cmd/server/main.go
│   ├── internal/...
│   ├── migrations/
│   ├── Dockerfile
│   └── go.mod
│
├── notification-service/           # Notification Service
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── email/                  # SMTP client
│   │   └── pubsub/                 # Redis subscribers
│   ├── migrations/
│   ├── Dockerfile
│   └── go.mod
│
├── shared/                         # Shared code
│   ├── proto/                      # Protobuf definitions
│   │   ├── auth/v1/auth.proto
│   │   ├── user/v1/user.proto
│   │   ├── workspace/v1/workspace.proto
│   │   ├── floor_plan/v1/floor_plan.proto
│   │   ├── scene/v1/scene.proto
│   │   ├── branch/v1/branch.proto
│   │   ├── ai/v1/ai.proto
│   │   ├── compliance/v1/compliance.proto
│   │   ├── request/v1/request.proto
│   │   ├── notification/v1/notification.proto
│   │   └── common/v1/common.proto
│   ├── gen/                        # Generated Go code
│   ├── pkg/
│   │   ├── logger/                 # Zap wrapper
│   │   ├── errors/                 # Domain errors
│   │   ├── config/                 # Viper helpers
│   │   ├── grpc/                   # gRPC helpers
│   │   └── validator/              # Validation helpers
│   └── go.mod
│
├── scripts/
│   └── init-databases.sql          # PostgreSQL init
│
├── docs/                           # Documentation
├── docker-compose.yml              # Development compose
├── Makefile
└── env.example
│   │   │       ├── response.go
│   │   │       └── errors.go
│   │   └── ws/
│   │       ├── hub.go
│   │       └── client.go
│   ├── dto/                        # Data Transfer Objects
│   │   ├── auth.go
│   │   ├── user.go
│   │   ├── workspace.go
│   │   ├── scene.go
│   │   ├── branch.go
│   │   ├── chat.go
│   │   └── request.go
│   ├── mapper/                     # Entity <-> DTO маппинг
│   │   ├── user.go
│   │   ├── workspace.go
│   │   └── ...
│   └── pkg/                        # Внутренние утилиты
│       ├── validator/
│       ├── pagination/
│       ├── hasher/
│       └── jwt/
├── pkg/                            # Публичные пакеты
│   ├── logger/
│   │   └── logger.go
│   ├── httputil/
│   │   └── httputil.go
│   └── pointer/
│       └── pointer.go
├── migrations/
│   ├── postgres/
│   │   ├── 000001_init.up.sql
│   │   └── 000001_init.down.sql
│   └── mongodb/
│       └── indexes.js
├── scripts/
│   ├── generate-mocks.sh
│   └── migrate.sh
├── deployments/
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   └── k8s/
│       ├── deployment.yaml
│       ├── service.yaml
│       └── configmap.yaml
├── docs/                           # Эта документация
├── .env.example
├── .golangci.yml
├── Makefile
├── go.mod
└── go.sum
```

## Поток данных

### Типичный HTTP запрос

```
┌─────────┐     ┌────────────┐     ┌───────────┐     ┌────────────┐
│  Client │────►│  Middleware│────►│  Handler  │────►│  Service   │
└─────────┘     └────────────┘     └───────────┘     └────────────┘
                     │                   │                  │
                     │                   │                  ▼
                     │                   │           ┌────────────┐
                     │                   │           │ Repository │
                     │                   │           └────────────┘
                     │                   │                  │
                     │                   │                  ▼
                     │                   │           ┌────────────┐
                     │                   │           │  Database  │
                     │                   │           └────────────┘
                     │                   │                  │
                     │                   │◄─────────────────┘
                     │◄──────────────────┘
                     │
              ┌──────┴──────┐
              │   Response  │
              └─────────────┘
```

### Middleware Chain

```go
// Порядок выполнения middleware (внешний → внутренний):
//
// 1. RequestID  - Генерация уникального ID запроса
// 2. Logger     - Логирование входящего запроса
// 3. Recover    - Перехват паник
// 4. CORS       - Cross-Origin Resource Sharing
// 5. RateLimit  - Ограничение частоты запросов
// 6. Auth       - Аутентификация (опционально)
// 7. Handler    - Обработка запроса

app.Use(middleware.RequestID())
app.Use(middleware.Logger(logger))
app.Use(middleware.Recover())
app.Use(middleware.CORS(corsConfig))
app.Use(middleware.RateLimit(rateLimiter))

// Защищённые маршруты
protected.Use(middleware.Auth(jwtService))
```

## Обработка ошибок

### Иерархия ошибок

```go
// internal/domain/errors/errors.go

// Базовые доменные ошибки
var (
    // ErrNotFound возвращается когда запрашиваемый ресурс не найден.
    ErrNotFound = errors.New("resource not found")
    
    // ErrAlreadyExists возвращается при попытке создать существующий ресурс.
    ErrAlreadyExists = errors.New("resource already exists")
    
    // ErrUnauthorized возвращается при отсутствии аутентификации.
    ErrUnauthorized = errors.New("unauthorized")
    
    // ErrForbidden возвращается при недостаточных правах доступа.
    ErrForbidden = errors.New("forbidden")
    
    // ErrValidation возвращается при ошибке валидации данных.
    ErrValidation = errors.New("validation error")
    
    // ErrInternal возвращается при внутренних ошибках сервера.
    ErrInternal = errors.New("internal error")
)

// DomainError представляет доменную ошибку с контекстом.
type DomainError struct {
    // Err базовая ошибка
    Err error
    // Code код ошибки для клиента
    Code string
    // Message сообщение для пользователя
    Message string
    // Details дополнительные детали
    Details map[string]interface{}
}

// Error реализует интерфейс error.
func (e *DomainError) Error() string {
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap позволяет использовать errors.Is и errors.As.
func (e *DomainError) Unwrap() error {
    return e.Err
}
```

### HTTP ответы об ошибках

```go
// internal/handler/http/response/errors.go

// ErrorResponse стандартный формат ошибки API.
type ErrorResponse struct {
    // Error объект ошибки
    Error ErrorDetail `json:"error"`
    // RequestID идентификатор запроса для отладки
    RequestID string `json:"request_id"`
}

// ErrorDetail детали ошибки.
type ErrorDetail struct {
    // Code машиночитаемый код ошибки
    Code string `json:"code"`
    // Message человекочитаемое сообщение
    Message string `json:"message"`
    // Details дополнительная информация (опционально)
    Details map[string]interface{} `json:"details,omitempty"`
}

// Пример ответа:
// {
//   "error": {
//     "code": "VALIDATION_ERROR",
//     "message": "Ошибка валидации данных",
//     "details": {
//       "email": "Некорректный формат email"
//     }
//   },
//   "request_id": "550e8400-e29b-41d4-a716-446655440000"
// }
```

## Конкурентность

### Worker Pool для AI запросов

```go
// internal/service/ai/worker_pool.go

// WorkerPool управляет пулом воркеров для обработки AI запросов.
// Ограничивает количество параллельных запросов к OpenRouter.
type WorkerPool struct {
    workers    int
    jobQueue   chan Job
    results    chan Result
    ctx        context.Context
    cancel     context.CancelFunc
    wg         sync.WaitGroup
}

// Job представляет задачу для обработки.
type Job struct {
    ID      string
    Type    JobType
    Payload interface{}
}

// JobType тип задачи.
type JobType string

const (
    // JobTypeRecognition распознавание планировки
    JobTypeRecognition JobType = "recognition"
    // JobTypeGeneration генерация вариантов
    JobTypeGeneration JobType = "generation"
    // JobTypeChat обработка чата
    JobTypeChat JobType = "chat"
)

// NewWorkerPool создаёт новый пул воркеров.
// workers - количество параллельных воркеров
// queueSize - размер очереди задач
func NewWorkerPool(workers, queueSize int) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())
    
    pool := &WorkerPool{
        workers:  workers,
        jobQueue: make(chan Job, queueSize),
        results:  make(chan Result, queueSize),
        ctx:      ctx,
        cancel:   cancel,
    }
    
    pool.start()
    return pool
}
```

### Graceful Shutdown

```go
// cmd/server/main.go

func main() {
    // ... инициализация ...
    
    // Канал для сигналов завершения
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    
    // Запуск сервера в горутине
    go func() {
        if err := app.Listen(cfg.Server.Address()); err != nil {
            log.Fatal("Server error", zap.Error(err))
        }
    }()
    
    // Ожидание сигнала завершения
    <-quit
    log.Info("Shutting down server...")
    
    // Контекст с таймаутом для graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Закрытие соединений
    if err := app.ShutdownWithContext(ctx); err != nil {
        log.Error("Server shutdown error", zap.Error(err))
    }
    
    // Закрытие пула воркеров
    workerPool.Shutdown()
    
    // Закрытие соединений с БД
    pgPool.Close()
    mongoClient.Disconnect(ctx)
    redisClient.Close()
    
    log.Info("Server stopped")
}
```

## Мониторинг и трейсинг

### Метрики (Prometheus)

```go
// internal/pkg/metrics/metrics.go

var (
    // HTTPRequestsTotal общее количество HTTP запросов
    HTTPRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "granula_http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    // HTTPRequestDuration длительность HTTP запросов
    HTTPRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "granula_http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"method", "path"},
    )
    
    // AIRequestsTotal количество AI запросов
    AIRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "granula_ai_requests_total",
            Help: "Total number of AI requests",
        },
        []string{"type", "model", "status"},
    )
    
    // ActiveWebSocketConnections активные WebSocket соединения
    ActiveWebSocketConnections = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "granula_websocket_connections_active",
            Help: "Number of active WebSocket connections",
        },
    )
)
```

### Structured Logging

```go
// pkg/logger/logger.go

// Logger обёртка над zap.Logger с контекстом.
type Logger struct {
    *zap.Logger
}

// WithContext возвращает логгер с полями из контекста.
// Автоматически добавляет request_id, user_id если доступны.
func (l *Logger) WithContext(ctx context.Context) *Logger {
    fields := make([]zap.Field, 0, 3)
    
    if requestID := ctx.Value(ctxKeyRequestID); requestID != nil {
        fields = append(fields, zap.String("request_id", requestID.(string)))
    }
    
    if userID := ctx.Value(ctxKeyUserID); userID != nil {
        fields = append(fields, zap.String("user_id", userID.(string)))
    }
    
    if traceID := ctx.Value(ctxKeyTraceID); traceID != nil {
        fields = append(fields, zap.String("trace_id", traceID.(string)))
    }
    
    return &Logger{l.With(fields...)}
}
```

## Тестирование

### Уровни тестирования

1. **Unit тесты** - тестирование отдельных функций/методов
2. **Integration тесты** - тестирование с реальными БД (testcontainers)
3. **E2E тесты** - полный flow через HTTP API

### Пример unit теста

```go
// internal/service/user/service_test.go

func TestUserService_Create(t *testing.T) {
    t.Parallel()
    
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    
    // Arrange
    mockRepo := mock.NewMockUserRepository(ctrl)
    mockHasher := mock.NewMockHasher(ctrl)
    
    svc := NewService(mockRepo, mockHasher)
    
    input := &dto.CreateUserInput{
        Email:    "test@example.com",
        Password: "password123",
        Name:     "Test User",
    }
    
    mockHasher.EXPECT().
        Hash("password123").
        Return("hashed_password", nil)
    
    mockRepo.EXPECT().
        Create(gomock.Any(), gomock.Any()).
        DoAndReturn(func(ctx context.Context, user *entity.User) error {
            assert.Equal(t, input.Email, user.Email)
            assert.Equal(t, "hashed_password", user.PasswordHash)
            return nil
        })
    
    // Act
    user, err := svc.Create(context.Background(), input)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, input.Email, user.Email)
}
```

### Integration тест с testcontainers

```go
// internal/repository/postgres/user_test.go

func TestUserRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    ctx := context.Background()
    
    // Запуск PostgreSQL контейнера
    pgContainer, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:16-alpine"),
        postgres.WithDatabase("test_db"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
    )
    require.NoError(t, err)
    defer pgContainer.Terminate(ctx)
    
    connStr, err := pgContainer.ConnectionString(ctx)
    require.NoError(t, err)
    
    // Подключение и миграции
    pool, err := pgxpool.New(ctx, connStr)
    require.NoError(t, err)
    defer pool.Close()
    
    runMigrations(t, pool)
    
    repo := NewUserRepository(pool)
    
    t.Run("Create and GetByID", func(t *testing.T) {
        user := &entity.User{
            ID:           uuid.New(),
            Email:        "test@example.com",
            PasswordHash: "hashed",
            Name:         "Test User",
        }
        
        err := repo.Create(ctx, user)
        require.NoError(t, err)
        
        found, err := repo.GetByID(ctx, user.ID)
        require.NoError(t, err)
        assert.Equal(t, user.Email, found.Email)
    })
}
```


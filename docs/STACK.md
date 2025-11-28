# Технологический стек Granula API

## Обзор

Granula API — высоконагруженный backend построенный на **микросервисной архитектуре** с 11 независимыми сервисами, gRPC коммуникацией и гибридным подходом к хранению данных.

### Архитектура высокого уровня

```
┌─────────────────────────────────────────────────────────────────────────┐
│                            CLIENTS                                       │
└─────────────────────────────────┬───────────────────────────────────────┘
                                  │ HTTP/WS
                                  ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    API GATEWAY (Go + Fiber)                              │
│              REST → gRPC • JWT validation • Rate Limiting               │
└─────────────────────────────────┬───────────────────────────────────────┘
                                  │ gRPC (protobuf)
    ┌─────────────┬───────────────┼───────────────┬─────────────┐
    ▼             ▼               ▼               ▼             ▼
┌───────┐   ┌─────────┐    ┌───────────┐   ┌─────────┐   ┌──────────┐
│ Auth  │   │  User   │    │ Workspace │   │Floor Pln│   │   ...    │
│:50051 │   │ :50052  │    │  :50053   │   │ :50054  │   │ +6 more  │
└───┬───┘   └────┬────┘    └─────┬─────┘   └────┬────┘   └────┬─────┘
    │            │               │              │             │
    └────────────┴───────────────┼──────────────┴─────────────┘
                                 ▼
┌─────────────────────────────────────────────────────────────────────────┐
│   PostgreSQL   │    MongoDB     │     Redis      │    MinIO/S3         │
│   (6 DBs)      │   (3 DBs)      │   Pub/Sub      │    Files            │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Языки программирования

### Go 1.22+

**Основной язык разработки**

```
Версия: 1.22+
Стандарт: Go Modules
```

**Почему Go:**
- Высокая производительность и низкое потребление памяти
- Отличная поддержка конкурентности (goroutines, channels)
- Быстрая компиляция и простой деплой (единственный бинарник)
- Строгая типизация с выводом типов
- Богатая стандартная библиотека
- Отличная экосистема для backend-разработки

**Используемые возможности Go 1.22:**
- Generics для типобезопасных утилит
- `log/slog` для структурированного логирования
- Улучшенный `net/http` routing
- `slices` и `maps` пакеты из стандартной библиотеки

---

## Web Framework

### Fiber v2

**Высокопроизводительный HTTP фреймворк**

```
Пакет: github.com/gofiber/fiber/v2
Версия: 2.52+
```

**Характеристики:**
- Построен на fasthttp (в 10x быстрее net/http)
- Express.js-подобный API
- Zero memory allocation в hot paths
- Встроенная поддержка WebSocket
- Middleware ecosystem

**Используемые middleware:**
```go
// Стандартные middleware Fiber
fiber.Use(recover.New())           // Перехват паник
fiber.Use(cors.New())              // CORS
fiber.Use(limiter.New())           // Rate limiting
fiber.Use(requestid.New())         // Request ID
fiber.Use(logger.New())            // HTTP logging
fiber.Use(compress.New())          // Gzip/Brotli compression
fiber.Use(cache.New())             // Response caching
fiber.Use(helmet.New())            // Security headers
```

**Альтернативы (не выбраны):**
- `gin` — менее производителен
- `echo` — меньше middleware
- `chi` — требует больше boilerplate

---

## gRPC (межсервисная коммуникация)

### Protocol Buffers + gRPC

**Коммуникация между микросервисами**

```
Пакет: google.golang.org/grpc
Версия: 1.60+
Protobuf: google.golang.org/protobuf
```

**Почему gRPC:**
- Высокая производительность (бинарный протокол)
- Строгая типизация (protobuf schemas)
- Bi-directional streaming для AI чата
- Автогенерация клиентов
- Встроенная поддержка deadline/timeout
- gRPC interceptors для logging/auth

**Структура proto файлов:**

```
shared/proto/
├── auth/v1/
│   └── auth.proto          # AuthService
├── user/v1/
│   └── user.proto          # UserService
├── workspace/v1/
│   └── workspace.proto     # WorkspaceService
├── floor_plan/v1/
│   └── floor_plan.proto    # FloorPlanService
├── scene/v1/
│   └── scene.proto         # SceneService
├── branch/v1/
│   └── branch.proto        # BranchService
├── ai/v1/
│   └── ai.proto            # AIService (with streaming)
├── compliance/v1/
│   └── compliance.proto    # ComplianceService
├── request/v1/
│   └── request.proto       # RequestService
├── notification/v1/
│   └── notification.proto  # NotificationService
└── common/v1/
    └── common.proto        # Shared messages
```

**Пример proto файла:**

```protobuf
syntax = "proto3";

package auth.v1;

option go_package = "github.com/granula/shared/gen/auth/v1;authv1";

service AuthService {
  // Регистрация нового пользователя
  rpc Register(RegisterRequest) returns (RegisterResponse);
  
  // Логин по email/password
  rpc Login(LoginRequest) returns (LoginResponse);
  
  // Валидация JWT токена (для Gateway)
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  
  // Обновление токена
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
}

message RegisterRequest {
  string email = 1;
  string password = 2;
  string name = 3;
}

message RegisterResponse {
  string user_id = 1;
  string access_token = 2;
  string refresh_token = 3;
}
```

**gRPC Streaming для AI чата:**

```protobuf
service AIService {
  // Streaming ответ от AI
  rpc StreamChatResponse(ChatRequest) returns (stream ChatChunk);
}

message ChatChunk {
  string content = 1;
  bool is_final = 2;
}
```

**Interceptors:**

```go
// Server interceptors
grpc.ChainUnaryInterceptor(
    grpc_recovery.UnaryServerInterceptor(),
    grpc_zap.UnaryServerInterceptor(logger),
    grpc_validator.UnaryServerInterceptor(),
)

// Client interceptors (в Gateway)
grpc.WithChainUnaryInterceptor(
    grpc_retry.UnaryClientInterceptor(),
    otelgrpc.UnaryClientInterceptor(),
)
```

---

## Базы данных

### PostgreSQL 16+

**Основная реляционная БД**

```
Версия: 16+
Драйвер: github.com/jackc/pgx/v5
Пул соединений: pgxpool
```

**Хранит:**
- Пользователи и аутентификация
- Воркспейсы и участники
- Планировки (метаданные)
- Заявки на экспертов
- Правила compliance
- Уведомления
- Сессии

**Особенности использования:**
```go
// Connection pooling
pool, _ := pgxpool.New(ctx, connString)
pool.Config().MaxConns = 25
pool.Config().MinConns = 5
pool.Config().MaxConnLifetime = time.Hour
pool.Config().MaxConnIdleTime = 30 * time.Minute
```

**Расширения PostgreSQL:**
- `uuid-ossp` — генерация UUID
- `pg_trgm` — полнотекстовый поиск
- `btree_gin` — составные индексы

**Миграции:**
```
Инструмент: golang-migrate/migrate
Формат: SQL файлы с версионированием
```

---

### MongoDB 7+

**Документная БД для неструктурированных данных**

```
Версия: 7+
Драйвер: go.mongodb.org/mongo-driver
```

**Хранит:**
- 3D сцены (SceneElements, стены, комнаты, мебель)
- Ветки дизайна (branches) с delta-изменениями
- История чата с AI
- AI контексты
- Снапшоты состояний

**Почему MongoDB для сцен:**
- Гибкая схема для 3D элементов
- Эффективное хранение вложенных структур
- Быстрые обновления частей документа
- Нативная поддержка массивов и объектов

**Индексы:**
```javascript
db.scenes.createIndex({ "workspaceId": 1 })
db.branches.createIndex({ "sceneId": 1, "isActive": 1 })
db.chat_messages.createIndex({ "sceneId": 1, "createdAt": 1 })
```

---

### Redis 7+

**In-memory хранилище для кэша и очередей**

```
Версия: 7+
Клиент: github.com/redis/go-redis/v9
```

**Используется для:**

| Функция | Структура данных | TTL |
|---------|------------------|-----|
| Кэш сущностей | Hash | 15 min |
| Сессии | Hash | 7 days |
| Rate limiting | Sorted Set | Dynamic |
| Очереди задач | List (FIFO) | - |
| Pub/Sub | Channels | - |
| Распределённые блокировки | String | 30 sec |
| Онлайн статус | Set | - |

**Lua скрипты:**
```lua
-- Sliding window rate limiting
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
redis.call('ZREMRANGEBYSCORE', key, 0, now - window)
local count = redis.call('ZCARD', key)
-- ...
```

---

## Хранилище файлов

### MinIO / AWS S3

**S3-совместимое объектное хранилище**

```
Development: MinIO (self-hosted)
Production: AWS S3 / Yandex Object Storage
SDK: github.com/aws/aws-sdk-go-v2/service/s3
```

**Структура bucket:**
```
granula/
├── floor-plans/{workspace_id}/{floor_plan_id}/
│   ├── original.pdf
│   ├── processed.png
│   └── thumbnail.png
├── renders/{scene_id}/{branch_id}/
│   ├── preview.png
│   └── render-3d.png
├── avatars/{user_id}.jpg
├── models/furniture/
│   └── *.glb
└── exports/{workspace_id}/
    └── *.pdf, *.dwg
```

**Функции:**
- Presigned URLs для безопасной загрузки/скачивания
- Lifecycle policies для автоочистки
- Public bucket policies для статики

---

## AI / ML интеграция

### OpenRouter API

**Унифицированный API для LLM**

```
URL: https://openrouter.ai/api/v1
Модели: anthropic/claude-sonnet-4
SDK: HTTP клиент (custom)
```

**Используемые модели:**

| Задача | Модель | Особенности |
|--------|--------|-------------|
| Распознавание планировок | Claude Sonnet 4 (Vision) | Multimodal, JSON output |
| Генерация вариантов | Claude Sonnet 4 | Structured output |
| Чат-ассистент | Claude Sonnet 4 | Streaming, context |
| Compliance проверка | Claude Sonnet 4 | Rule-based reasoning |

**Worker Pool:**
```go
// Ограничение параллельных запросов
type WorkerPool struct {
    workers  int           // 5 workers
    jobQueue chan Job
    results  chan Result
}
```

**Rate Limits:**
- 60 RPM (requests per minute)
- 100k TPM (tokens per minute)

---

## Аутентификация и безопасность

### JWT (JSON Web Tokens)

```
Библиотека: github.com/golang-jwt/jwt/v5
Алгоритм: HS256 (HMAC-SHA256)
```

**Токены:**

| Тип | TTL | Хранение |
|-----|-----|----------|
| Access Token | 15 min | Client (memory) |
| Refresh Token | 7 days | PostgreSQL (hash) |

**Claims структура:**
```go
type AccessTokenClaims struct {
    jwt.RegisteredClaims
    UserID   string `json:"uid"`
    Email    string `json:"email"`
    Role     string `json:"role"`
    DeviceID string `json:"did,omitempty"`
}
```

---

### bcrypt

**Хеширование паролей**

```
Библиотека: golang.org/x/crypto/bcrypt
Cost factor: 12
```

---

### OAuth 2.0

**Социальная авторизация**

```
Библиотека: golang.org/x/oauth2
Провайдеры: Google, Yandex
Flow: Authorization Code
```

---

## Валидация

### go-playground/validator

**Валидация структур**

```
Пакет: github.com/go-playground/validator/v10
```

**Примеры правил:**
```go
type CreateUserInput struct {
    Email    string `validate:"required,email,max=255"`
    Password string `validate:"required,min=8,max=72"`
    Name     string `validate:"required,min=2,max=255"`
}
```

**Кастомные валидаторы:**
- `phone_ru` — российский номер телефона
- `safe_string` — защита от XSS
- `password_strength` — сложность пароля

---

## Логирование

### Uber Zap

**Высокопроизводительное структурированное логирование**

```
Пакет: go.uber.org/zap
Формат: JSON (production) / Console (development)
```

**Уровни логирования:**
- `debug` — детальная отладка
- `info` — информационные сообщения
- `warn` — предупреждения
- `error` — ошибки

**Пример использования:**
```go
logger.Info("User created",
    zap.String("user_id", user.ID.String()),
    zap.String("email", user.Email),
    zap.Duration("duration", time.Since(start)),
)
```

**Контекстное логирование:**
```go
logger.WithContext(ctx) // Добавляет request_id, user_id, trace_id
```

---

## Мониторинг и метрики

### Prometheus

**Сбор метрик**

```
Пакет: github.com/prometheus/client_golang
Endpoint: /metrics
```

**Типы метрик:**

| Метрика | Тип | Labels |
|---------|-----|--------|
| `http_requests_total` | Counter | method, path, status |
| `http_request_duration_seconds` | Histogram | method, path |
| `db_connections_open` | Gauge | database |
| `ai_tokens_used_total` | Counter | type |
| `cache_hits_total` | Counter | cache |

---

### Grafana

**Визуализация метрик**

```
Версия: 10+
Datasources: Prometheus, Loki
```

**Dashboards:**
- API Overview (RPS, latency, errors)
- Database Performance
- AI Service Metrics
- Business Metrics

---

### Jaeger / OpenTelemetry

**Распределённый трейсинг**

```
SDK: go.opentelemetry.io/otel
Exporter: Jaeger
```

**Трейсы покрывают:**
- HTTP запросы
- Database queries
- Redis operations
- External API calls (OpenRouter)

---

### Sentry

**Error tracking**

```
SDK: github.com/getsentry/sentry-go
```

**Функции:**
- Автоматический capture паник
- Контекст ошибок (user, request)
- Performance monitoring
- Release tracking

---

## Тестирование

### Стандартный testing

```go
func TestUserService_Create(t *testing.T) {
    t.Parallel()
    // ...
}
```

### testify

**Assertions и mocking**

```
Пакет: github.com/stretchr/testify
```

```go
assert.NoError(t, err)
assert.Equal(t, expected, actual)
require.NotNil(t, result)
```

### gomock

**Генерация моков**

```
Пакет: github.com/golang/mock/gomock
Генератор: mockgen
```

```go
ctrl := gomock.NewController(t)
mockRepo := mock.NewMockUserRepository(ctrl)
mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
```

### testcontainers

**Интеграционные тесты с реальными БД**

```
Пакет: github.com/testcontainers/testcontainers-go
```

```go
pgContainer, _ := postgres.RunContainer(ctx,
    testcontainers.WithImage("postgres:16-alpine"),
    postgres.WithDatabase("test_db"),
)
```

---

## CI/CD

### GitHub Actions

**Continuous Integration**

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - run: go test -v -race ./...
  
  build:
    needs: test
    steps:
      - uses: docker/build-push-action@v5
```

### Docker

**Контейнеризация**

```
Base image: golang:1.22-alpine (build)
Runtime image: alpine:3.19
Multi-stage build: Yes
```

**Оптимизации:**
- Multi-stage builds
- Layer caching
- Non-root user
- Health checks

### Docker Compose

**Локальная разработка**

```yaml
services:
  api:
    build: .
    depends_on:
      - postgres
      - mongodb
      - redis
      - minio
```

---

## Линтинг и форматирование

### golangci-lint

**Агрегатор линтеров**

```
Версия: latest
Конфиг: .golangci.yml
```

**Включённые линтеры:**
- `errcheck` — проверка обработки ошибок
- `gosimple` — упрощение кода
- `govet` — подозрительные конструкции
- `ineffassign` — неиспользуемые присваивания
- `staticcheck` — статический анализ
- `unused` — неиспользуемый код
- `gofmt` — форматирование
- `goimports` — сортировка импортов
- `misspell` — орфография
- `gosec` — безопасность

### goimports

**Автоформатирование импортов**

```
Пакет: golang.org/x/tools/cmd/goimports
```

---

## Утилиты разработки

### Air

**Hot reload для Go**

```
Пакет: github.com/cosmtrek/air
```

```bash
air  # Автоперезапуск при изменениях
```

### golang-migrate

**Миграции базы данных**

```
Пакет: github.com/golang-migrate/migrate/v4
```

```bash
migrate -path migrations/postgres -database $DSN up
migrate create -ext sql -dir migrations/postgres -seq add_users
```

### swag

**Генерация OpenAPI/Swagger**

```
Пакет: github.com/swaggo/swag
```

```bash
swag init -g cmd/server/main.go -o docs/swagger
```

---

## Основные Go пакеты

### Стандартная библиотека

| Пакет | Назначение |
|-------|------------|
| `context` | Передача контекста, отмена операций |
| `crypto/rand` | Криптографически безопасные числа |
| `encoding/json` | JSON сериализация |
| `errors` | Обработка ошибок |
| `fmt` | Форматирование |
| `io` | I/O операции |
| `net/http` | HTTP клиент |
| `os` | Операционная система |
| `sync` | Примитивы синхронизации |
| `time` | Работа со временем |

### Сторонние пакеты

| Пакет | Версия | Назначение |
|-------|--------|------------|
| `github.com/gofiber/fiber/v2` | 2.52+ | Web framework |
| `github.com/jackc/pgx/v5` | 5.5+ | PostgreSQL driver |
| `go.mongodb.org/mongo-driver` | 1.14+ | MongoDB driver |
| `github.com/redis/go-redis/v9` | 9.4+ | Redis client |
| `github.com/aws/aws-sdk-go-v2` | 2.x | AWS SDK (S3) |
| `github.com/golang-jwt/jwt/v5` | 5.2+ | JWT tokens |
| `golang.org/x/crypto` | latest | Crypto utilities |
| `golang.org/x/oauth2` | latest | OAuth 2.0 |
| `github.com/go-playground/validator/v10` | 10.18+ | Validation |
| `go.uber.org/zap` | 1.27+ | Logging |
| `github.com/prometheus/client_golang` | 1.18+ | Prometheus metrics |
| `github.com/google/uuid` | 1.6+ | UUID generation |
| `github.com/caarlos0/env/v10` | 10.0+ | Env parsing |
| `github.com/joho/godotenv` | 1.5+ | .env loading |

---

## Архитектурные паттерны

### Clean Architecture

```
┌─────────────────────────────────────┐
│           Handlers (HTTP)           │  ← Внешний слой
├─────────────────────────────────────┤
│           Services (Use Cases)      │  ← Бизнес-логика
├─────────────────────────────────────┤
│           Domain (Entities)         │  ← Ядро
├─────────────────────────────────────┤
│      Repository Interfaces          │  ← Абстракции
├─────────────────────────────────────┤
│   Repository Implementations        │  ← Инфраструктура
└─────────────────────────────────────┘
```

### Repository Pattern

Абстракция доступа к данным через интерфейсы.

### Dependency Injection

Все зависимости инжектируются через конструкторы.

### CQRS (частично)

Разделение команд и запросов для сложных операций.

---

## Требования к окружению

### Development

| Компонент | Версия |
|-----------|--------|
| Go | 1.22+ |
| Docker | 24+ |
| Docker Compose | 2.24+ |
| Make | 4+ |
| Git | 2.40+ |

### Production

| Компонент | Рекомендация |
|-----------|--------------|
| CPU | 2+ cores |
| RAM | 2+ GB |
| Disk | SSD, 20+ GB |
| OS | Linux (Alpine) |

---

## Версионирование

| Компонент | Версия |
|-----------|--------|
| API Version | v1 |
| Schema Version | 1 |
| Supported Go | 1.22+ |

---

## Ссылки

- [Go Documentation](https://go.dev/doc/)
- [Fiber Documentation](https://docs.gofiber.io/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/16/)
- [MongoDB Documentation](https://www.mongodb.com/docs/)
- [Redis Documentation](https://redis.io/docs/)
- [OpenRouter API](https://openrouter.ai/docs)
- [Prometheus](https://prometheus.io/docs/)
- [Docker](https://docs.docker.com/)


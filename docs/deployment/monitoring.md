# Мониторинг и наблюдаемость

## Обзор

Granula API реализует полную наблюдаемость через:
- **Метрики** (Prometheus)
- **Логирование** (Structured logging с Zap)
- **Трейсинг** (OpenTelemetry/Jaeger)
- **Алертинг** (Alertmanager)

## Архитектура мониторинга

```
┌─────────────────────────────────────────────────────────────────┐
│                       Granula API                                │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │   Metrics   │  │   Logging   │  │   Tracing   │             │
│  │ (Prometheus)│  │   (Zap)     │  │ (OTel SDK)  │             │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘             │
└─────────┼────────────────┼────────────────┼─────────────────────┘
          │                │                │
          ▼                ▼                ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ Prometheus  │    │  Loki       │    │   Jaeger    │
│             │    │             │    │             │
└──────┬──────┘    └──────┬──────┘    └─────────────┘
       │                  │
       ▼                  ▼
┌─────────────────────────────────────┐
│             Grafana                  │
│  ┌─────────┐  ┌─────────┐          │
│  │Dashboards│  │ Alerts │          │
│  └─────────┘  └─────────┘          │
└─────────────────────────────────────┘
       │
       ▼
┌─────────────┐
│Alertmanager │
│   → Slack   │
│   → Email   │
│   → PagerDuty│
└─────────────┘
```

## Метрики Prometheus

### HTTP метрики

```go
// internal/pkg/metrics/http.go

var (
    // HTTPRequestsTotal общее количество HTTP запросов
    HTTPRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "granula",
            Subsystem: "http",
            Name:      "requests_total",
            Help:      "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    // HTTPRequestDuration длительность запросов
    HTTPRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "granula",
            Subsystem: "http",
            Name:      "request_duration_seconds",
            Help:      "HTTP request duration in seconds",
            Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"method", "path"},
    )
    
    // HTTPRequestSize размер запросов
    HTTPRequestSize = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "granula",
            Subsystem: "http",
            Name:      "request_size_bytes",
            Help:      "HTTP request size in bytes",
            Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
        },
        []string{"method", "path"},
    )
    
    // HTTPResponseSize размер ответов
    HTTPResponseSize = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "granula",
            Subsystem: "http",
            Name:      "response_size_bytes",
            Help:      "HTTP response size in bytes",
            Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
        },
        []string{"method", "path"},
    )
    
    // ActiveConnections активные соединения
    ActiveConnections = promauto.NewGauge(
        prometheus.GaugeOpts{
            Namespace: "granula",
            Subsystem: "http",
            Name:      "active_connections",
            Help:      "Number of active HTTP connections",
        },
    )
)
```

### Бизнес-метрики

```go
// internal/pkg/metrics/business.go

var (
    // WorkspacesCreated создано воркспейсов
    WorkspacesCreated = promauto.NewCounter(
        prometheus.CounterOpts{
            Namespace: "granula",
            Subsystem: "business",
            Name:      "workspaces_created_total",
            Help:      "Total number of workspaces created",
        },
    )
    
    // FloorPlansProcessed обработано планировок
    FloorPlansProcessed = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "granula",
            Subsystem: "business",
            Name:      "floor_plans_processed_total",
            Help:      "Total number of floor plans processed",
        },
        []string{"status"}, // success, failed
    )
    
    // AIGenerationsTotal генераций AI
    AIGenerationsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "granula",
            Subsystem: "ai",
            Name:      "generations_total",
            Help:      "Total number of AI generations",
        },
        []string{"type"}, // recognition, variants, chat
    )
    
    // AITokensUsed использовано токенов
    AITokensUsed = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "granula",
            Subsystem: "ai",
            Name:      "tokens_used_total",
            Help:      "Total AI tokens used",
        },
        []string{"type"}, // prompt, completion
    )
    
    // ExpertRequestsTotal заявок на экспертов
    ExpertRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "granula",
            Subsystem: "business",
            Name:      "expert_requests_total",
            Help:      "Total expert requests",
        },
        []string{"service_type", "status"},
    )
    
    // ActiveUsers активные пользователи
    ActiveUsers = promauto.NewGauge(
        prometheus.GaugeOpts{
            Namespace: "granula",
            Subsystem: "business",
            Name:      "active_users",
            Help:      "Number of active users (24h)",
        },
    )
)
```

### Database метрики

```go
// internal/pkg/metrics/database.go

var (
    // DBConnectionsOpen открытые соединения
    DBConnectionsOpen = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Namespace: "granula",
            Subsystem: "db",
            Name:      "connections_open",
            Help:      "Number of open database connections",
        },
        []string{"database"}, // postgres, mongodb, redis
    )
    
    // DBQueryDuration длительность запросов
    DBQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Namespace: "granula",
            Subsystem: "db",
            Name:      "query_duration_seconds",
            Help:      "Database query duration",
            Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5},
        },
        []string{"database", "operation"},
    )
    
    // DBErrors ошибки БД
    DBErrors = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "granula",
            Subsystem: "db",
            Name:      "errors_total",
            Help:      "Total database errors",
        },
        []string{"database", "operation"},
    )
    
    // CacheHitRate hit rate кэша
    CacheHits = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "granula",
            Subsystem: "cache",
            Name:      "hits_total",
            Help:      "Total cache hits",
        },
        []string{"cache"},
    )
    
    CacheMisses = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "granula",
            Subsystem: "cache",
            Name:      "misses_total",
            Help:      "Total cache misses",
        },
        []string{"cache"},
    )
)
```

## Health Checks

```go
// internal/handler/http/health.go

// HealthHandler обработчик health check.
type HealthHandler struct {
    postgres *pgxpool.Pool
    mongo    *mongo.Client
    redis    *redis.Client
}

// Health проверяет здоровье сервиса.
func (h *HealthHandler) Health(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
    defer cancel()
    
    checks := map[string]string{
        "status": "ok",
    }
    
    // PostgreSQL
    if err := h.postgres.Ping(ctx); err != nil {
        checks["postgres"] = "unhealthy"
        checks["status"] = "degraded"
    } else {
        checks["postgres"] = "healthy"
    }
    
    // MongoDB
    if err := h.mongo.Ping(ctx, nil); err != nil {
        checks["mongodb"] = "unhealthy"
        checks["status"] = "degraded"
    } else {
        checks["mongodb"] = "healthy"
    }
    
    // Redis
    if err := h.redis.Ping(ctx).Err(); err != nil {
        checks["redis"] = "unhealthy"
        checks["status"] = "degraded"
    } else {
        checks["redis"] = "healthy"
    }
    
    status := fiber.StatusOK
    if checks["status"] != "ok" {
        status = fiber.StatusServiceUnavailable
    }
    
    return c.Status(status).JSON(checks)
}

// Ready проверяет готовность к приёму трафика.
func (h *HealthHandler) Ready(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"status": "ready"})
}

// Live проверяет что процесс жив.
func (h *HealthHandler) Live(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"status": "alive"})
}
```

## Grafana Dashboards

### API Overview Dashboard

```json
{
  "title": "Granula API Overview",
  "panels": [
    {
      "title": "Request Rate",
      "type": "graph",
      "targets": [
        {
          "expr": "sum(rate(granula_http_requests_total[5m])) by (status)",
          "legendFormat": "{{status}}"
        }
      ]
    },
    {
      "title": "Request Latency (p99)",
      "type": "graph",
      "targets": [
        {
          "expr": "histogram_quantile(0.99, sum(rate(granula_http_request_duration_seconds_bucket[5m])) by (le, path))",
          "legendFormat": "{{path}}"
        }
      ]
    },
    {
      "title": "Error Rate",
      "type": "singlestat",
      "targets": [
        {
          "expr": "sum(rate(granula_http_requests_total{status=~\"5..\"}[5m])) / sum(rate(granula_http_requests_total[5m])) * 100"
        }
      ]
    },
    {
      "title": "Active Connections",
      "type": "gauge",
      "targets": [
        {
          "expr": "granula_http_active_connections"
        }
      ]
    }
  ]
}
```

## Алерты

```yaml
# alerting/rules.yml
groups:
  - name: granula-api
    rules:
      # High Error Rate
      - alert: HighErrorRate
        expr: |
          sum(rate(granula_http_requests_total{status=~"5.."}[5m])) 
          / sum(rate(granula_http_requests_total[5m])) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }} (threshold: 5%)"
      
      # High Latency
      - alert: HighLatency
        expr: |
          histogram_quantile(0.99, sum(rate(granula_http_request_duration_seconds_bucket[5m])) by (le)) > 2
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High latency detected"
          description: "P99 latency is {{ $value }}s (threshold: 2s)"
      
      # Database Connection Issues
      - alert: DatabaseConnectionPoolExhausted
        expr: granula_db_connections_open{database="postgres"} > 20
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Database connection pool near exhaustion"
      
      # AI Service Issues
      - alert: AIServiceHighLatency
        expr: |
          histogram_quantile(0.95, sum(rate(granula_openrouter_request_duration_seconds_bucket[5m])) by (le)) > 30
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "AI service high latency"
      
      # Low Cache Hit Rate
      - alert: LowCacheHitRate
        expr: |
          sum(rate(granula_cache_hits_total[5m])) 
          / (sum(rate(granula_cache_hits_total[5m])) + sum(rate(granula_cache_misses_total[5m]))) < 0.8
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Cache hit rate below 80%"
```

## Логирование

```go
// pkg/logger/logger.go

// NewLogger создаёт настроенный логгер.
func NewLogger(cfg *LoggingConfig) (*zap.Logger, error) {
    var config zap.Config
    
    if cfg.Format == "json" {
        config = zap.NewProductionConfig()
    } else {
        config = zap.NewDevelopmentConfig()
    }
    
    config.Level = zap.NewAtomicLevelAt(parseLevel(cfg.Level))
    config.OutputPaths = []string{cfg.Output}
    
    // Добавляем поля по умолчанию
    config.InitialFields = map[string]interface{}{
        "service": "granula-api",
        "version": Version,
    }
    
    return config.Build()
}

// Пример использования
logger.Info("User created",
    zap.String("user_id", user.ID.String()),
    zap.String("email", user.Email),
    zap.Duration("duration", time.Since(start)),
)

logger.Error("Failed to process floor plan",
    zap.String("floor_plan_id", id),
    zap.Error(err),
)
```

## Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /health` | Полная проверка здоровья |
| `GET /health/live` | Liveness probe (процесс жив) |
| `GET /health/ready` | Readiness probe (готов к трафику) |
| `GET /metrics` | Prometheus метрики |


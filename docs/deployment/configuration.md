# Конфигурация

## Переменные окружения

### Обязательные

```env
# =============================================================================
# SERVER
# =============================================================================
APP_ENV=production                    # development | staging | production
APP_PORT=8080                         # Порт API сервера
APP_HOST=0.0.0.0                      # Хост для прослушивания

# =============================================================================
# DATABASES
# =============================================================================
# PostgreSQL
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=granula
POSTGRES_PASSWORD=<strong-password>
POSTGRES_DB=granula
POSTGRES_SSL_MODE=require             # disable | require | verify-full
POSTGRES_MAX_CONNS=25                 # Максимум соединений в пуле
POSTGRES_MIN_CONNS=5                  # Минимум соединений

# MongoDB
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=granula
MONGODB_MAX_POOL_SIZE=100

# Redis
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=<redis-password>
REDIS_DB=0
REDIS_MAX_RETRIES=3

# =============================================================================
# STORAGE
# =============================================================================
S3_ENDPOINT=s3.amazonaws.com          # или localhost:9000 для MinIO
S3_ACCESS_KEY=<access-key>
S3_SECRET_KEY=<secret-key>
S3_BUCKET=granula
S3_REGION=eu-central-1
S3_USE_SSL=true

# =============================================================================
# AI / OpenRouter
# =============================================================================
OPENROUTER_API_KEY=sk-or-v1-xxx
OPENROUTER_DEFAULT_MODEL=anthropic/claude-sonnet-4
OPENROUTER_VISION_MODEL=anthropic/claude-sonnet-4

# =============================================================================
# AUTH
# =============================================================================
JWT_SECRET=<256-bit-secret>           # Минимум 32 символа
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h                  # 7 дней

# OAuth (опционально)
OAUTH_GOOGLE_CLIENT_ID=xxx
OAUTH_GOOGLE_CLIENT_SECRET=xxx
OAUTH_YANDEX_CLIENT_ID=xxx
OAUTH_YANDEX_CLIENT_SECRET=xxx
```

### Опциональные

```env
# =============================================================================
# LOGGING
# =============================================================================
LOG_LEVEL=info                        # debug | info | warn | error
LOG_FORMAT=json                       # json | console
LOG_OUTPUT=stdout                     # stdout | file
LOG_FILE_PATH=/var/log/granula/api.log

# =============================================================================
# RATE LIMITING
# =============================================================================
RATE_LIMIT_GENERAL=1000               # Запросов в минуту
RATE_LIMIT_AUTH=10                    # Попыток входа в минуту
RATE_LIMIT_AI=20                      # AI запросов в минуту
RATE_LIMIT_UPLOAD=50                  # Загрузок в час

# =============================================================================
# CORS
# =============================================================================
CORS_ALLOWED_ORIGINS=https://app.granula.ru,https://granula.ru
CORS_ALLOWED_METHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS
CORS_MAX_AGE=86400

# =============================================================================
# EMAIL
# =============================================================================
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=noreply@granula.ru
SMTP_PASSWORD=<app-password>
SMTP_FROM=Granula <noreply@granula.ru>

# =============================================================================
# MONITORING
# =============================================================================
METRICS_ENABLED=true
METRICS_PATH=/metrics
TRACING_ENABLED=true
TRACING_ENDPOINT=http://jaeger:14268/api/traces
SENTRY_DSN=https://xxx@sentry.io/xxx

# =============================================================================
# WORKERS
# =============================================================================
AI_WORKERS_COUNT=5                    # Количество AI воркеров
RENDER_WORKERS_COUNT=3                # Количество воркеров рендеринга
```

## Структура конфигурации (Go)

```go
// internal/config/config.go

// Config конфигурация приложения.
type Config struct {
    App       AppConfig       `envPrefix:"APP_"`
    Postgres  PostgresConfig  `envPrefix:"POSTGRES_"`
    MongoDB   MongoDBConfig   `envPrefix:"MONGODB_"`
    Redis     RedisConfig     `envPrefix:"REDIS_"`
    S3        S3Config        `envPrefix:"S3_"`
    OpenRouter OpenRouterConfig `envPrefix:"OPENROUTER_"`
    JWT       JWTConfig       `envPrefix:"JWT_"`
    OAuth     OAuthConfig
    Logging   LoggingConfig   `envPrefix:"LOG_"`
    RateLimit RateLimitConfig `envPrefix:"RATE_LIMIT_"`
    CORS      CORSConfig      `envPrefix:"CORS_"`
    SMTP      SMTPConfig      `envPrefix:"SMTP_"`
    Metrics   MetricsConfig   `envPrefix:"METRICS_"`
    Workers   WorkersConfig   `envPrefix:"WORKERS_"`
}

// AppConfig основные настройки приложения.
type AppConfig struct {
    Env  string `env:"ENV" envDefault:"development"`
    Port int    `env:"PORT" envDefault:"8080"`
    Host string `env:"HOST" envDefault:"0.0.0.0"`
}

// PostgresConfig настройки PostgreSQL.
type PostgresConfig struct {
    Host     string `env:"HOST" envDefault:"localhost"`
    Port     int    `env:"PORT" envDefault:"5432"`
    User     string `env:"USER,required"`
    Password string `env:"PASSWORD,required"`
    Database string `env:"DB,required"`
    SSLMode  string `env:"SSL_MODE" envDefault:"disable"`
    MaxConns int32  `env:"MAX_CONNS" envDefault:"25"`
    MinConns int32  `env:"MIN_CONNS" envDefault:"5"`
}

// DSN возвращает строку подключения.
func (c *PostgresConfig) DSN() string {
    return fmt.Sprintf(
        "postgres://%s:%s@%s:%d/%s?sslmode=%s",
        c.User, c.Password, c.Host, c.Port, c.Database, c.SSLMode,
    )
}

// MongoDBConfig настройки MongoDB.
type MongoDBConfig struct {
    URI         string `env:"URI,required"`
    Database    string `env:"DATABASE,required"`
    MaxPoolSize uint64 `env:"MAX_POOL_SIZE" envDefault:"100"`
}

// RedisConfig настройки Redis.
type RedisConfig struct {
    URL        string `env:"URL,required"`
    Password   string `env:"PASSWORD"`
    DB         int    `env:"DB" envDefault:"0"`
    MaxRetries int    `env:"MAX_RETRIES" envDefault:"3"`
}

// S3Config настройки S3.
type S3Config struct {
    Endpoint   string        `env:"ENDPOINT,required"`
    AccessKey  string        `env:"ACCESS_KEY,required"`
    SecretKey  string        `env:"SECRET_KEY,required"`
    Bucket     string        `env:"BUCKET,required"`
    Region     string        `env:"REGION" envDefault:"us-east-1"`
    UseSSL     bool          `env:"USE_SSL" envDefault:"true"`
    PresignTTL time.Duration `env:"PRESIGN_TTL" envDefault:"1h"`
}

// OpenRouterConfig настройки OpenRouter.
type OpenRouterConfig struct {
    APIKey       string        `env:"API_KEY,required"`
    BaseURL      string        `env:"BASE_URL" envDefault:"https://openrouter.ai/api/v1"`
    DefaultModel string        `env:"DEFAULT_MODEL" envDefault:"anthropic/claude-sonnet-4"`
    VisionModel  string        `env:"VISION_MODEL" envDefault:"anthropic/claude-sonnet-4"`
    Timeout      time.Duration `env:"TIMEOUT" envDefault:"120s"`
    MaxRetries   int           `env:"MAX_RETRIES" envDefault:"3"`
}

// JWTConfig настройки JWT.
type JWTConfig struct {
    Secret     string        `env:"SECRET,required"`
    AccessTTL  time.Duration `env:"ACCESS_TTL" envDefault:"15m"`
    RefreshTTL time.Duration `env:"REFRESH_TTL" envDefault:"168h"`
}

// Load загружает конфигурацию из окружения.
func Load() (*Config, error) {
    // Загружаем .env файл если есть
    _ = godotenv.Load()
    
    var cfg Config
    if err := env.Parse(&cfg); err != nil {
        return nil, fmt.Errorf("parse env: %w", err)
    }
    
    // Валидация
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("validate config: %w", err)
    }
    
    return &cfg, nil
}

// Validate проверяет конфигурацию.
func (c *Config) Validate() error {
    if len(c.JWT.Secret) < 32 {
        return errors.New("JWT_SECRET must be at least 32 characters")
    }
    
    if c.App.Env == "production" {
        if !c.S3.UseSSL {
            return errors.New("S3_USE_SSL must be true in production")
        }
        if c.Postgres.SSLMode == "disable" {
            return errors.New("POSTGRES_SSL_MODE must not be 'disable' in production")
        }
    }
    
    return nil
}

// IsDevelopment проверяет режим разработки.
func (c *Config) IsDevelopment() bool {
    return c.App.Env == "development"
}

// IsProduction проверяет production режим.
func (c *Config) IsProduction() bool {
    return c.App.Env == "production"
}
```

## Profiles

### Development

```env
APP_ENV=development
LOG_LEVEL=debug
LOG_FORMAT=console
POSTGRES_SSL_MODE=disable
S3_USE_SSL=false
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
```

### Staging

```env
APP_ENV=staging
LOG_LEVEL=info
LOG_FORMAT=json
POSTGRES_SSL_MODE=require
S3_USE_SSL=true
CORS_ALLOWED_ORIGINS=https://staging.granula.ru
SENTRY_DSN=https://xxx@sentry.io/staging
```

### Production

```env
APP_ENV=production
LOG_LEVEL=info
LOG_FORMAT=json
POSTGRES_SSL_MODE=verify-full
S3_USE_SSL=true
CORS_ALLOWED_ORIGINS=https://app.granula.ru,https://granula.ru
SENTRY_DSN=https://xxx@sentry.io/production
```

## Secrets Management

### HashiCorp Vault

```go
// internal/config/vault.go

// VaultClient клиент Vault.
type VaultClient struct {
    client *vault.Client
}

// LoadSecrets загружает секреты из Vault.
func (v *VaultClient) LoadSecrets(path string) (map[string]string, error) {
    secret, err := v.client.Logical().Read(path)
    if err != nil {
        return nil, err
    }
    
    if secret == nil {
        return nil, errors.New("secret not found")
    }
    
    data := secret.Data["data"].(map[string]interface{})
    result := make(map[string]string)
    
    for k, v := range data {
        result[k] = v.(string)
    }
    
    return result, nil
}

// Использование
func LoadConfigWithVault() (*Config, error) {
    vaultClient, err := NewVaultClient(os.Getenv("VAULT_ADDR"), os.Getenv("VAULT_TOKEN"))
    if err != nil {
        return nil, err
    }
    
    secrets, err := vaultClient.LoadSecrets("secret/data/granula/api")
    if err != nil {
        return nil, err
    }
    
    // Устанавливаем секреты в окружение
    for k, v := range secrets {
        os.Setenv(k, v)
    }
    
    return Load()
}
```

### AWS Secrets Manager

```go
// internal/config/aws_secrets.go

// LoadAWSSecrets загружает секреты из AWS Secrets Manager.
func LoadAWSSecrets(secretName string) (map[string]string, error) {
    cfg, err := config.LoadDefaultConfig(context.Background())
    if err != nil {
        return nil, err
    }
    
    client := secretsmanager.NewFromConfig(cfg)
    
    input := &secretsmanager.GetSecretValueInput{
        SecretId: aws.String(secretName),
    }
    
    result, err := client.GetSecretValue(context.Background(), input)
    if err != nil {
        return nil, err
    }
    
    var secrets map[string]string
    if err := json.Unmarshal([]byte(*result.SecretString), &secrets); err != nil {
        return nil, err
    }
    
    return secrets, nil
}
```

## Feature Flags

```go
// internal/config/features.go

// FeatureFlags флаги функций.
type FeatureFlags struct {
    // AIGeneration включена ли AI генерация
    AIGeneration bool `env:"FEATURE_AI_GENERATION" envDefault:"true"`
    
    // WebSocketChat включён ли WebSocket чат
    WebSocketChat bool `env:"FEATURE_WS_CHAT" envDefault:"true"`
    
    // ExpertRequests включены ли заявки на экспертов
    ExpertRequests bool `env:"FEATURE_EXPERT_REQUESTS" envDefault:"true"`
    
    // OAuth включена ли OAuth авторизация
    OAuth bool `env:"FEATURE_OAUTH" envDefault:"true"`
    
    // PushNotifications включены ли push уведомления
    PushNotifications bool `env:"FEATURE_PUSH" envDefault:"false"`
}

// IsEnabled проверяет флаг по имени.
func (f *FeatureFlags) IsEnabled(name string) bool {
    v := reflect.ValueOf(f).Elem()
    field := v.FieldByName(name)
    if !field.IsValid() {
        return false
    }
    return field.Bool()
}
```


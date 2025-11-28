# Безопасность

## Обзор

Granula API реализует многоуровневую систему безопасности, охватывающую аутентификацию, авторизацию, защиту данных и мониторинг угроз.

## Аутентификация

### JWT Tokens

```go
// internal/pkg/jwt/jwt.go

// TokenPair пара access и refresh токенов.
type TokenPair struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"`  // Секунды до истечения access token
}

// AccessTokenClaims claims для access токена.
type AccessTokenClaims struct {
    jwt.RegisteredClaims
    UserID   string `json:"uid"`
    Email    string `json:"email"`
    Role     string `json:"role"`
    DeviceID string `json:"did,omitempty"`
}

// Конфигурация токенов
// Access Token:
//   - Алгоритм: HS256 (HMAC-SHA256)
//   - TTL: 15 минут
//   - Содержит: user_id, email, role
//
// Refresh Token:
//   - Формат: 32 байта crypto/rand, base64
//   - TTL: 7 дней
//   - Хранится: хэш (SHA-256) в PostgreSQL
//   - Ротация: новый токен при каждом refresh
```

### Процесс аутентификации

```
┌─────────┐                    ┌─────────┐                    ┌─────────┐
│  Client │                    │   API   │                    │   DB    │
└────┬────┘                    └────┬────┘                    └────┬────┘
     │                              │                              │
     │  POST /auth/login            │                              │
     │  {email, password}           │                              │
     │─────────────────────────────►│                              │
     │                              │                              │
     │                              │  Verify password hash        │
     │                              │─────────────────────────────►│
     │                              │◄─────────────────────────────│
     │                              │                              │
     │                              │  Create session              │
     │                              │─────────────────────────────►│
     │                              │◄─────────────────────────────│
     │                              │                              │
     │  200 OK                      │                              │
     │  {access_token, refresh}     │                              │
     │◄─────────────────────────────│                              │
     │                              │                              │
     │  GET /api/v1/workspaces      │                              │
     │  Authorization: Bearer xxx   │                              │
     │─────────────────────────────►│                              │
     │                              │                              │
     │                              │  Verify JWT signature        │
     │                              │  (no DB call)                │
     │                              │                              │
     │  200 OK                      │                              │
     │  {workspaces: [...]}         │                              │
     │◄─────────────────────────────│                              │
```

### OAuth 2.0

Поддерживаемые провайдеры:
- **Google** - OAuth 2.0 + OpenID Connect
- **Yandex** - OAuth 2.0

```go
// internal/service/auth/oauth.go

// OAuthProvider интерфейс OAuth провайдера.
type OAuthProvider interface {
    // GetAuthURL возвращает URL для редиректа пользователя.
    GetAuthURL(state string) string
    
    // ExchangeCode обменивает code на токены.
    ExchangeCode(ctx context.Context, code string) (*OAuthTokens, error)
    
    // GetUserInfo получает информацию о пользователе.
    GetUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error)
}

// OAuthUserInfo информация о пользователе от провайдера.
type OAuthUserInfo struct {
    ID       string  // ID в системе провайдера
    Email    string  // Email (verified)
    Name     string  // Отображаемое имя
    Avatar   *string // URL аватара
}
```

## Авторизация

### Role-Based Access Control (RBAC)

```go
// internal/domain/entity/role.go

// Role роль пользователя в системе.
type Role string

const (
    // RoleUser обычный пользователь.
    RoleUser Role = "user"
    
    // RoleAdmin администратор системы.
    RoleAdmin Role = "admin"
    
    // RoleExpert эксперт БТИ.
    RoleExpert Role = "expert"
)

// Permissions матрица разрешений.
var Permissions = map[Role][]Permission{
    RoleUser: {
        PermissionWorkspaceCreate,
        PermissionWorkspaceRead,
        PermissionWorkspaceUpdate,
        PermissionWorkspaceDelete,
        PermissionSceneCreate,
        PermissionSceneRead,
        PermissionSceneUpdate,
        PermissionRequestCreate,
        PermissionRequestRead,
    },
    RoleExpert: {
        // Все права user плюс:
        PermissionRequestProcess,
        PermissionComplianceManage,
    },
    RoleAdmin: {
        // Все права
        PermissionAll,
    },
}
```

### Resource-Based Access Control

```go
// internal/domain/entity/workspace_role.go

// WorkspaceRole роль в рамках воркспейса.
type WorkspaceRole string

const (
    // WorkspaceRoleOwner владелец (полные права).
    WorkspaceRoleOwner WorkspaceRole = "owner"
    
    // WorkspaceRoleEditor редактор (чтение/запись).
    WorkspaceRoleEditor WorkspaceRole = "editor"
    
    // WorkspaceRoleViewer наблюдатель (только чтение).
    WorkspaceRoleViewer WorkspaceRole = "viewer"
)

// CanEdit проверяет право на редактирование.
func (r WorkspaceRole) CanEdit() bool {
    return r == WorkspaceRoleOwner || r == WorkspaceRoleEditor
}

// CanDelete проверяет право на удаление.
func (r WorkspaceRole) CanDelete() bool {
    return r == WorkspaceRoleOwner
}

// CanInvite проверяет право на приглашение участников.
func (r WorkspaceRole) CanInvite() bool {
    return r == WorkspaceRoleOwner || r == WorkspaceRoleEditor
}
```

### Middleware авторизации

```go
// internal/handler/http/middleware/auth.go

// AuthMiddleware проверяет JWT токен и устанавливает контекст.
func AuthMiddleware(jwtService jwt.Service) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Извлечение токена из заголовка
        authHeader := c.Get("Authorization")
        if authHeader == "" {
            return response.Unauthorized(c, "Missing authorization header")
        }
        
        // Формат: "Bearer <token>"
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            return response.Unauthorized(c, "Invalid authorization format")
        }
        
        // Валидация токена
        claims, err := jwtService.ValidateAccessToken(parts[1])
        if err != nil {
            if errors.Is(err, jwt.ErrTokenExpired) {
                return response.Unauthorized(c, "Token expired")
            }
            return response.Unauthorized(c, "Invalid token")
        }
        
        // Установка в контекст
        c.Locals("user_id", claims.UserID)
        c.Locals("user_email", claims.Email)
        c.Locals("user_role", claims.Role)
        
        return c.Next()
    }
}

// RequireRole проверяет наличие требуемой роли.
func RequireRole(roles ...entity.Role) fiber.Handler {
    return func(c *fiber.Ctx) error {
        userRole := c.Locals("user_role").(string)
        
        for _, role := range roles {
            if string(role) == userRole {
                return c.Next()
            }
        }
        
        return response.Forbidden(c, "Insufficient permissions")
    }
}

// RequireWorkspaceAccess проверяет доступ к воркспейсу.
func RequireWorkspaceAccess(
    memberRepo repository.WorkspaceMemberRepository,
    requiredRoles ...entity.WorkspaceRole,
) fiber.Handler {
    return func(c *fiber.Ctx) error {
        userID := c.Locals("user_id").(string)
        workspaceID := c.Params("workspaceId")
        
        member, err := memberRepo.GetByUserAndWorkspace(
            c.Context(),
            uuid.MustParse(userID),
            uuid.MustParse(workspaceID),
        )
        if err != nil {
            if errors.Is(err, domain.ErrNotFound) {
                return response.Forbidden(c, "No access to workspace")
            }
            return response.InternalError(c, err)
        }
        
        // Проверка роли
        if len(requiredRoles) > 0 {
            hasRole := false
            for _, role := range requiredRoles {
                if member.Role == role {
                    hasRole = true
                    break
                }
            }
            if !hasRole {
                return response.Forbidden(c, "Insufficient workspace permissions")
            }
        }
        
        c.Locals("workspace_role", member.Role)
        return c.Next()
    }
}
```

## Защита данных

### Шифрование

```go
// internal/pkg/crypto/crypto.go

// Конфигурация шифрования:
//
// Пароли:
//   - Алгоритм: bcrypt
//   - Cost: 12
//   - Salt: встроен в хэш
//
// Чувствительные данные в БД:
//   - Алгоритм: AES-256-GCM
//   - Ключ: из переменной окружения
//   - IV: случайный для каждой записи
//
// Токены:
//   - Refresh: 32 байта crypto/rand
//   - Хранение: SHA-256 хэш

// Hasher интерфейс для хэширования паролей.
type Hasher interface {
    // Hash создаёт хэш пароля.
    Hash(password string) (string, error)
    
    // Compare сравнивает пароль с хэшем.
    Compare(password, hash string) error
}

// BcryptHasher реализация на bcrypt.
type BcryptHasher struct {
    cost int
}

func NewBcryptHasher(cost int) *BcryptHasher {
    if cost < bcrypt.MinCost {
        cost = bcrypt.DefaultCost
    }
    return &BcryptHasher{cost: cost}
}

func (h *BcryptHasher) Hash(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
    return string(bytes), err
}

func (h *BcryptHasher) Compare(password, hash string) error {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
```

### Валидация входных данных

```go
// internal/pkg/validator/validator.go

// Validator валидатор входных данных.
type Validator struct {
    validate *validator.Validate
}

func NewValidator() *Validator {
    v := validator.New()
    
    // Регистрация кастомных валидаторов
    v.RegisterValidation("phone_ru", validateRussianPhone)
    v.RegisterValidation("safe_string", validateSafeString)
    
    return &Validator{validate: v}
}

// validateSafeString проверяет строку на опасные символы.
func validateSafeString(fl validator.FieldLevel) bool {
    value := fl.Field().String()
    
    // Запрещённые паттерны
    dangerous := []string{
        "<script", "javascript:", "onerror=", "onload=",
        "eval(", "expression(", "url(", "import(",
    }
    
    lower := strings.ToLower(value)
    for _, pattern := range dangerous {
        if strings.Contains(lower, pattern) {
            return false
        }
    }
    
    return true
}

// Пример использования в DTO
type CreateWorkspaceInput struct {
    Name        string `json:"name" validate:"required,min=1,max=255,safe_string"`
    Description string `json:"description" validate:"max=2000,safe_string"`
    Address     string `json:"address" validate:"max=500,safe_string"`
}
```

### SQL Injection Prevention

```go
// Все запросы используют параметризацию через pgx

// ✅ Правильно - параметризованный запрос
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
    query := `
        SELECT id, email, password_hash, name, role, created_at
        FROM users
        WHERE email = $1 AND deleted_at IS NULL
    `
    
    var user entity.User
    err := r.pool.QueryRow(ctx, query, email).Scan(
        &user.ID, &user.Email, &user.PasswordHash,
        &user.Name, &user.Role, &user.CreatedAt,
    )
    
    return &user, err
}

// ❌ НИКОГДА - конкатенация строк
// query := "SELECT * FROM users WHERE email = '" + email + "'"
```

## Rate Limiting

```go
// internal/handler/http/middleware/ratelimit.go

// RateLimitConfig конфигурация rate limiting.
type RateLimitConfig struct {
    // Общий лимит запросов
    GeneralLimit   int           // 1000 req
    GeneralWindow  time.Duration // 1 min
    
    // Лимит для аутентификации
    AuthLimit      int           // 10 req
    AuthWindow     time.Duration // 1 min
    
    // Лимит для AI запросов
    AILimit        int           // 20 req
    AIWindow       time.Duration // 1 min
    
    // Лимит для загрузки файлов
    UploadLimit    int           // 50 req
    UploadWindow   time.Duration // 1 hour
}

// RateLimiter middleware для rate limiting.
type RateLimiter struct {
    redis  *redis.Client
    config RateLimitConfig
}

// Limit применяет rate limiting.
func (rl *RateLimiter) Limit(limitType string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        var limit int
        var window time.Duration
        
        switch limitType {
        case "auth":
            limit = rl.config.AuthLimit
            window = rl.config.AuthWindow
        case "ai":
            limit = rl.config.AILimit
            window = rl.config.AIWindow
        case "upload":
            limit = rl.config.UploadLimit
            window = rl.config.UploadWindow
        default:
            limit = rl.config.GeneralLimit
            window = rl.config.GeneralWindow
        }
        
        // Идентификатор: user_id или IP
        identifier := rl.getIdentifier(c)
        key := fmt.Sprintf("ratelimit:%s:%s", limitType, identifier)
        
        allowed, remaining, retryAfter, err := rl.check(c.Context(), key, limit, window)
        if err != nil {
            // При ошибке Redis пропускаем запрос
            return c.Next()
        }
        
        // Заголовки rate limit
        c.Set("X-RateLimit-Limit", strconv.Itoa(limit))
        c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
        c.Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(window).Unix(), 10))
        
        if !allowed {
            c.Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
            return response.TooManyRequests(c, "Rate limit exceeded")
        }
        
        return c.Next()
    }
}
```

## Мониторинг безопасности

### Логирование событий безопасности

```go
// internal/service/audit/audit.go

// AuditEventType тип события аудита.
type AuditEventType string

const (
    // Аутентификация
    AuditLoginSuccess     AuditEventType = "auth.login.success"
    AuditLoginFailed      AuditEventType = "auth.login.failed"
    AuditLogout           AuditEventType = "auth.logout"
    AuditTokenRefresh     AuditEventType = "auth.token.refresh"
    AuditPasswordChange   AuditEventType = "auth.password.change"
    
    // Авторизация
    AuditAccessDenied     AuditEventType = "authz.access.denied"
    AuditRoleChange       AuditEventType = "authz.role.change"
    
    // Данные
    AuditDataExport       AuditEventType = "data.export"
    AuditDataDelete       AuditEventType = "data.delete"
    
    // Подозрительная активность
    AuditSuspiciousLogin  AuditEventType = "security.suspicious.login"
    AuditBruteForce       AuditEventType = "security.brute.force"
)

// AuditEvent событие аудита.
type AuditEvent struct {
    ID        string                 `json:"id"`
    Type      AuditEventType         `json:"type"`
    UserID    *string                `json:"user_id,omitempty"`
    IP        string                 `json:"ip"`
    UserAgent string                 `json:"user_agent"`
    Details   map[string]interface{} `json:"details,omitempty"`
    Timestamp time.Time              `json:"timestamp"`
}

// AuditLogger логгер событий безопасности.
type AuditLogger interface {
    Log(ctx context.Context, event AuditEvent) error
}
```

### Детекция аномалий

```go
// internal/service/security/anomaly.go

// AnomalyDetector детектор аномальной активности.
type AnomalyDetector struct {
    redis  *redis.Client
    logger *zap.Logger
}

// CheckLoginAnomaly проверяет аномалии при входе.
func (d *AnomalyDetector) CheckLoginAnomaly(
    ctx context.Context,
    userID string,
    ip string,
    userAgent string,
) (*AnomalyResult, error) {
    result := &AnomalyResult{}
    
    // Проверка: новый IP
    knownIPs, err := d.redis.SMembers(ctx, fmt.Sprintf("user:%s:ips", userID)).Result()
    if err == nil {
        if !contains(knownIPs, ip) && len(knownIPs) > 0 {
            result.Flags = append(result.Flags, "new_ip")
            result.RiskScore += 20
        }
    }
    
    // Проверка: новый User-Agent
    knownUA, err := d.redis.SMembers(ctx, fmt.Sprintf("user:%s:ua", userID)).Result()
    if err == nil {
        if !contains(knownUA, userAgent) && len(knownUA) > 0 {
            result.Flags = append(result.Flags, "new_device")
            result.RiskScore += 15
        }
    }
    
    // Проверка: географическая аномалия (impossible travel)
    // TODO: GeoIP lookup и проверка времени между разными локациями
    
    // Проверка: время суток
    hour := time.Now().Hour()
    if hour >= 2 && hour <= 5 {
        result.Flags = append(result.Flags, "unusual_time")
        result.RiskScore += 10
    }
    
    // Определение уровня риска
    switch {
    case result.RiskScore >= 50:
        result.Level = RiskLevelHigh
    case result.RiskScore >= 25:
        result.Level = RiskLevelMedium
    default:
        result.Level = RiskLevelLow
    }
    
    return result, nil
}
```

## Заголовки безопасности

```go
// internal/handler/http/middleware/security.go

// SecurityHeaders добавляет заголовки безопасности.
func SecurityHeaders() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Защита от XSS
        c.Set("X-XSS-Protection", "1; mode=block")
        
        // Запрет iframe embedding
        c.Set("X-Frame-Options", "DENY")
        
        // Запрет MIME sniffing
        c.Set("X-Content-Type-Options", "nosniff")
        
        // Referrer policy
        c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
        
        // Content Security Policy
        c.Set("Content-Security-Policy", strings.Join([]string{
            "default-src 'self'",
            "script-src 'self'",
            "style-src 'self' 'unsafe-inline'",
            "img-src 'self' data: https:",
            "font-src 'self'",
            "connect-src 'self'",
            "frame-ancestors 'none'",
        }, "; "))
        
        // Strict Transport Security (для HTTPS)
        if c.Protocol() == "https" {
            c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        }
        
        return c.Next()
    }
}
```

## CORS Configuration

```go
// internal/handler/http/middleware/cors.go

// CORSConfig конфигурация CORS.
func CORSConfig(allowedOrigins []string) cors.Config {
    return cors.Config{
        // Разрешённые origins
        AllowOrigins: strings.Join(allowedOrigins, ","),
        
        // Разрешённые методы
        AllowMethods: strings.Join([]string{
            fiber.MethodGet,
            fiber.MethodPost,
            fiber.MethodPut,
            fiber.MethodPatch,
            fiber.MethodDelete,
            fiber.MethodOptions,
        }, ","),
        
        // Разрешённые заголовки
        AllowHeaders: strings.Join([]string{
            "Origin",
            "Content-Type",
            "Accept",
            "Authorization",
            "X-Request-ID",
        }, ","),
        
        // Заголовки доступные клиенту
        ExposeHeaders: strings.Join([]string{
            "X-Request-ID",
            "X-RateLimit-Limit",
            "X-RateLimit-Remaining",
            "X-RateLimit-Reset",
        }, ","),
        
        // Разрешить credentials
        AllowCredentials: true,
        
        // Кэширование preflight
        MaxAge: 86400, // 24 часа
    }
}
```

## Чеклист безопасности

### Перед деплоем

- [ ] Все секреты вынесены в переменные окружения
- [ ] JWT секрет минимум 256 бит
- [ ] HTTPS обязателен в production
- [ ] Rate limiting настроен
- [ ] Логирование событий безопасности включено
- [ ] Soft delete для критичных данных
- [ ] Бэкапы базы данных настроены
- [ ] Мониторинг ошибок (Sentry) подключен

### Регулярно

- [ ] Обновление зависимостей (dependabot)
- [ ] Аудит npm/go packages
- [ ] Ревью логов безопасности
- [ ] Ротация секретов
- [ ] Тестирование DR плана


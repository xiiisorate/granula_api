# API аутентификации

## Обзор

Granula API использует JWT токены для аутентификации. Система поддерживает:
- Email/password аутентификацию
- OAuth 2.0 (Google, Yandex)
- Refresh токены для продления сессии

## Endpoints

### POST /api/v1/auth/register

Регистрация нового пользователя.

**Request:**

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "name": "Иван Петров"
}
```

**Validation:**

| Поле | Правила |
|------|---------|
| `email` | Required, valid email, max 255 chars |
| `password` | Required, min 8 chars, max 72 chars, must contain: uppercase, lowercase, digit |
| `name` | Required, min 2 chars, max 255 chars |

**Response 201:**

```json
{
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@example.com",
      "name": "Иван Петров",
      "role": "user",
      "email_verified": false,
      "created_at": "2024-01-15T10:30:00Z"
    },
    "tokens": {
      "access_token": "eyJhbGciOiJIUzI1NiIs...",
      "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2...",
      "expires_in": 900
    }
  },
  "request_id": "req_abc123"
}
```

**Errors:**

| Code | HTTP | Description |
|------|------|-------------|
| `EMAIL_ALREADY_EXISTS` | 409 | Email уже зарегистрирован |
| `VALIDATION_ERROR` | 400 | Ошибка валидации |
| `INTERNAL_ERROR` | 500 | Внутренняя ошибка |

---

### POST /api/v1/auth/login

Вход в систему.

**Request:**

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "device_id": "optional-device-fingerprint"
}
```

**Response 200:**

```json
{
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@example.com",
      "name": "Иван Петров",
      "role": "user",
      "email_verified": true,
      "avatar_url": "https://storage.granula.ru/avatars/550e8400.jpg",
      "created_at": "2024-01-15T10:30:00Z"
    },
    "tokens": {
      "access_token": "eyJhbGciOiJIUzI1NiIs...",
      "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2...",
      "expires_in": 900
    }
  },
  "request_id": "req_abc123"
}
```

**Errors:**

| Code | HTTP | Description |
|------|------|-------------|
| `INVALID_CREDENTIALS` | 401 | Неверный email или пароль |
| `ACCOUNT_DISABLED` | 403 | Аккаунт заблокирован |
| `TOO_MANY_ATTEMPTS` | 429 | Превышен лимит попыток |

---

### POST /api/v1/auth/refresh

Обновление access токена.

**Request:**

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2..."
}
```

**Response 200:**

```json
{
  "data": {
    "tokens": {
      "access_token": "eyJhbGciOiJIUzI1NiIs...",
      "refresh_token": "bmV3IHJlZnJlc2ggdG9r...",
      "expires_in": 900
    }
  },
  "request_id": "req_abc123"
}
```

**Note:** Refresh токен ротируется при каждом использовании (одноразовый).

**Errors:**

| Code | HTTP | Description |
|------|------|-------------|
| `INVALID_REFRESH_TOKEN` | 401 | Недействительный refresh токен |
| `TOKEN_EXPIRED` | 401 | Refresh токен истёк |
| `SESSION_REVOKED` | 401 | Сессия отозвана |

---

### POST /api/v1/auth/logout

Выход из системы (инвалидация токенов).

**Request:**

```http
POST /api/v1/auth/logout
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2..."
}
```

**Response 200:**

```json
{
  "data": {
    "message": "Successfully logged out"
  },
  "request_id": "req_abc123"
}
```

---

### POST /api/v1/auth/logout-all

Выход из всех устройств.

**Request:**

```http
POST /api/v1/auth/logout-all
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "message": "Successfully logged out from all devices",
    "sessions_revoked": 3
  },
  "request_id": "req_abc123"
}
```

---

### GET /api/v1/auth/oauth/:provider

Инициация OAuth flow.

**Parameters:**

| Name | Type | Description |
|------|------|-------------|
| `provider` | path | `google` или `yandex` |
| `redirect_uri` | query | URL для редиректа после авторизации |

**Request:**

```http
GET /api/v1/auth/oauth/google?redirect_uri=https://app.granula.ru/auth/callback
```

**Response 302:**

Редирект на страницу авторизации провайдера.

```http
Location: https://accounts.google.com/o/oauth2/v2/auth?client_id=...&redirect_uri=...&state=...
```

---

### POST /api/v1/auth/oauth/:provider/callback

Обработка OAuth callback.

**Request:**

```http
POST /api/v1/auth/oauth/google/callback
Content-Type: application/json

{
  "code": "authorization_code_from_provider",
  "state": "state_from_initial_request"
}
```

**Response 200:**

```json
{
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "user@gmail.com",
      "name": "Иван Петров",
      "role": "user",
      "email_verified": true,
      "avatar_url": "https://lh3.googleusercontent.com/...",
      "oauth_provider": "google",
      "created_at": "2024-01-15T10:30:00Z"
    },
    "tokens": {
      "access_token": "eyJhbGciOiJIUzI1NiIs...",
      "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2...",
      "expires_in": 900
    },
    "is_new_user": false
  },
  "request_id": "req_abc123"
}
```

---

### POST /api/v1/auth/password/forgot

Запрос на сброс пароля.

**Request:**

```http
POST /api/v1/auth/password/forgot
Content-Type: application/json

{
  "email": "user@example.com"
}
```

**Response 200:**

```json
{
  "data": {
    "message": "Password reset email sent if account exists"
  },
  "request_id": "req_abc123"
}
```

**Note:** Всегда возвращает 200, даже если email не найден (защита от enumeration).

---

### POST /api/v1/auth/password/reset

Сброс пароля по токену.

**Request:**

```http
POST /api/v1/auth/password/reset
Content-Type: application/json

{
  "token": "reset_token_from_email",
  "password": "NewSecurePassword123!"
}
```

**Response 200:**

```json
{
  "data": {
    "message": "Password successfully reset"
  },
  "request_id": "req_abc123"
}
```

**Errors:**

| Code | HTTP | Description |
|------|------|-------------|
| `INVALID_RESET_TOKEN` | 400 | Недействительный токен |
| `TOKEN_EXPIRED` | 400 | Токен истёк |

---

### POST /api/v1/auth/email/verify

Подтверждение email.

**Request:**

```http
POST /api/v1/auth/email/verify
Content-Type: application/json

{
  "token": "verification_token_from_email"
}
```

**Response 200:**

```json
{
  "data": {
    "message": "Email successfully verified"
  },
  "request_id": "req_abc123"
}
```

---

### POST /api/v1/auth/email/resend

Повторная отправка письма подтверждения.

**Request:**

```http
POST /api/v1/auth/email/resend
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "message": "Verification email sent"
  },
  "request_id": "req_abc123"
}
```

**Rate Limit:** 1 запрос в 60 секунд.

---

## DTO Types

```go
// internal/dto/auth.go

// RegisterInput данные для регистрации.
type RegisterInput struct {
    // Email пользователя
    // Required: true
    // Format: email
    // Example: user@example.com
    Email string `json:"email" validate:"required,email,max=255"`
    
    // Пароль
    // Required: true
    // MinLength: 8
    // MaxLength: 72
    // Pattern: must contain uppercase, lowercase, digit
    Password string `json:"password" validate:"required,min=8,max=72,password_strength"`
    
    // Отображаемое имя
    // Required: true
    // MinLength: 2
    // MaxLength: 255
    Name string `json:"name" validate:"required,min=2,max=255"`
}

// LoginInput данные для входа.
type LoginInput struct {
    // Email пользователя
    Email string `json:"email" validate:"required,email"`
    
    // Пароль
    Password string `json:"password" validate:"required"`
    
    // Идентификатор устройства (опционально)
    // Используется для привязки сессии к устройству
    DeviceID string `json:"device_id,omitempty" validate:"max=255"`
}

// RefreshInput данные для обновления токена.
type RefreshInput struct {
    // Refresh токен
    RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutInput данные для выхода.
type LogoutInput struct {
    // Refresh токен для инвалидации
    RefreshToken string `json:"refresh_token" validate:"required"`
}

// OAuthCallbackInput данные OAuth callback.
type OAuthCallbackInput struct {
    // Authorization code от провайдера
    Code string `json:"code" validate:"required"`
    
    // State для защиты от CSRF
    State string `json:"state" validate:"required"`
}

// ForgotPasswordInput данные для запроса сброса пароля.
type ForgotPasswordInput struct {
    // Email пользователя
    Email string `json:"email" validate:"required,email"`
}

// ResetPasswordInput данные для сброса пароля.
type ResetPasswordInput struct {
    // Токен сброса из email
    Token string `json:"token" validate:"required"`
    
    // Новый пароль
    Password string `json:"password" validate:"required,min=8,max=72,password_strength"`
}

// VerifyEmailInput данные для подтверждения email.
type VerifyEmailInput struct {
    // Токен подтверждения из email
    Token string `json:"token" validate:"required"`
}

// AuthResponse ответ аутентификации.
type AuthResponse struct {
    // Данные пользователя
    User *UserResponse `json:"user"`
    
    // Токены доступа
    Tokens *TokensResponse `json:"tokens"`
    
    // Признак нового пользователя (для OAuth)
    IsNewUser bool `json:"is_new_user,omitempty"`
}

// TokensResponse токены доступа.
type TokensResponse struct {
    // JWT access токен
    AccessToken string `json:"access_token"`
    
    // Refresh токен для обновления
    RefreshToken string `json:"refresh_token"`
    
    // Время жизни access токена в секундах
    ExpiresIn int64 `json:"expires_in"`
}
```

## Примеры использования

### JavaScript/TypeScript

```typescript
// Регистрация
const register = async (email: string, password: string, name: string) => {
  const response = await fetch('/api/v1/auth/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password, name }),
  });
  
  const { data } = await response.json();
  
  // Сохраняем токены
  localStorage.setItem('access_token', data.tokens.access_token);
  localStorage.setItem('refresh_token', data.tokens.refresh_token);
  
  return data.user;
};

// Запрос с авторизацией
const fetchWithAuth = async (url: string, options: RequestInit = {}) => {
  const token = localStorage.getItem('access_token');
  
  const response = await fetch(url, {
    ...options,
    headers: {
      ...options.headers,
      'Authorization': `Bearer ${token}`,
    },
  });
  
  // Автоматическое обновление при 401
  if (response.status === 401) {
    const refreshed = await refreshTokens();
    if (refreshed) {
      return fetchWithAuth(url, options);
    }
    // Редирект на логин
    window.location.href = '/login';
  }
  
  return response;
};

// Обновление токенов
const refreshTokens = async (): Promise<boolean> => {
  const refreshToken = localStorage.getItem('refresh_token');
  if (!refreshToken) return false;
  
  try {
    const response = await fetch('/api/v1/auth/refresh', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    
    if (!response.ok) return false;
    
    const { data } = await response.json();
    localStorage.setItem('access_token', data.tokens.access_token);
    localStorage.setItem('refresh_token', data.tokens.refresh_token);
    
    return true;
  } catch {
    return false;
  }
};
```

### cURL Examples

```bash
# Регистрация
curl -X POST https://api.granula.ru/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"SecurePass123!","name":"Иван"}'

# Вход
curl -X POST https://api.granula.ru/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"SecurePass123!"}'

# Запрос с токеном
curl -X GET https://api.granula.ru/api/v1/users/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."

# Обновление токена
curl -X POST https://api.granula.ru/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"dGhpcyBpcyBhIHJlZnJlc2..."}'
```


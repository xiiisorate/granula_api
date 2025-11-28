# API пользователей

## Обзор

API управления профилем пользователя, настройками и сессиями.

## Endpoints

### GET /api/v1/users/me

Получение профиля текущего пользователя.

**Request:**

```http
GET /api/v1/users/me
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "Иван Петров",
    "role": "user",
    "email_verified": true,
    "avatar_url": "https://storage.granula.ru/avatars/550e8400.jpg",
    "oauth_provider": null,
    "settings": {
      "language": "ru",
      "theme": "light",
      "notifications": {
        "email": true,
        "push": true
      }
    },
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-20T15:45:00Z"
  },
  "request_id": "req_abc123"
}
```

---

### PATCH /api/v1/users/me

Обновление профиля пользователя.

**Request:**

```http
PATCH /api/v1/users/me
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "name": "Иван Сидоров",
  "settings": {
    "language": "en",
    "theme": "dark"
  }
}
```

**Validation:**

| Поле | Правила |
|------|---------|
| `name` | Optional, min 2 chars, max 255 chars |
| `settings` | Optional, valid JSON object |

**Response 200:**

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "Иван Сидоров",
    "role": "user",
    "email_verified": true,
    "avatar_url": "https://storage.granula.ru/avatars/550e8400.jpg",
    "settings": {
      "language": "en",
      "theme": "dark",
      "notifications": {
        "email": true,
        "push": true
      }
    },
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-21T12:00:00Z"
  },
  "request_id": "req_abc123"
}
```

---

### PUT /api/v1/users/me/avatar

Загрузка/обновление аватара.

**Request:**

```http
PUT /api/v1/users/me/avatar
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: multipart/form-data

--boundary
Content-Disposition: form-data; name="avatar"; filename="photo.jpg"
Content-Type: image/jpeg

<binary image data>
--boundary--
```

**Constraints:**

| Parameter | Value |
|-----------|-------|
| Max file size | 5 MB |
| Allowed types | image/jpeg, image/png, image/webp |
| Output size | 256x256 px (cropped/resized) |

**Response 200:**

```json
{
  "data": {
    "avatar_url": "https://storage.granula.ru/avatars/550e8400.jpg?v=1705841200"
  },
  "request_id": "req_abc123"
}
```

---

### DELETE /api/v1/users/me/avatar

Удаление аватара.

**Request:**

```http
DELETE /api/v1/users/me/avatar
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "message": "Avatar deleted"
  },
  "request_id": "req_abc123"
}
```

---

### PUT /api/v1/users/me/password

Изменение пароля.

**Request:**

```http
PUT /api/v1/users/me/password
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "current_password": "OldPassword123!",
  "new_password": "NewSecurePass456!"
}
```

**Validation:**

| Поле | Правила |
|------|---------|
| `current_password` | Required |
| `new_password` | Required, min 8 chars, max 72 chars, must differ from current |

**Response 200:**

```json
{
  "data": {
    "message": "Password changed successfully"
  },
  "request_id": "req_abc123"
}
```

**Errors:**

| Code | HTTP | Description |
|------|------|-------------|
| `INVALID_PASSWORD` | 400 | Неверный текущий пароль |
| `SAME_PASSWORD` | 400 | Новый пароль совпадает с текущим |
| `WEAK_PASSWORD` | 400 | Пароль не соответствует требованиям |

---

### GET /api/v1/users/me/sessions

Список активных сессий пользователя.

**Request:**

```http
GET /api/v1/users/me/sessions
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "sessions": [
      {
        "id": "sess_abc123",
        "device": "Chrome on Windows",
        "ip": "192.168.1.1",
        "location": "Москва, Россия",
        "is_current": true,
        "last_active": "2024-01-21T12:00:00Z",
        "created_at": "2024-01-15T10:30:00Z"
      },
      {
        "id": "sess_def456",
        "device": "Safari on iPhone",
        "ip": "10.0.0.5",
        "location": "Санкт-Петербург, Россия",
        "is_current": false,
        "last_active": "2024-01-20T18:45:00Z",
        "created_at": "2024-01-18T09:15:00Z"
      }
    ],
    "total": 2
  },
  "request_id": "req_abc123"
}
```

---

### DELETE /api/v1/users/me/sessions/:sessionId

Завершение конкретной сессии.

**Request:**

```http
DELETE /api/v1/users/me/sessions/sess_def456
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "message": "Session revoked"
  },
  "request_id": "req_abc123"
}
```

**Errors:**

| Code | HTTP | Description |
|------|------|-------------|
| `SESSION_NOT_FOUND` | 404 | Сессия не найдена |
| `CANNOT_REVOKE_CURRENT` | 400 | Нельзя отозвать текущую сессию |

---

### DELETE /api/v1/users/me

Удаление аккаунта (soft delete).

**Request:**

```http
DELETE /api/v1/users/me
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "password": "CurrentPassword123!",
  "reason": "Больше не пользуюсь сервисом"
}
```

**Response 200:**

```json
{
  "data": {
    "message": "Account scheduled for deletion",
    "deletion_date": "2024-02-21T12:00:00Z"
  },
  "request_id": "req_abc123"
}
```

**Note:** 
- Аккаунт деактивируется немедленно
- Полное удаление данных через 30 дней
- В течение 30 дней можно восстановить аккаунт

---

## DTO Types

```go
// internal/dto/user.go

// UserResponse полные данные пользователя.
type UserResponse struct {
    // Уникальный идентификатор
    ID string `json:"id"`
    
    // Email пользователя
    Email string `json:"email"`
    
    // Отображаемое имя
    Name string `json:"name"`
    
    // Роль в системе
    Role string `json:"role"`
    
    // Подтверждён ли email
    EmailVerified bool `json:"email_verified"`
    
    // URL аватара
    AvatarURL *string `json:"avatar_url"`
    
    // OAuth провайдер (если применимо)
    OAuthProvider *string `json:"oauth_provider"`
    
    // Настройки пользователя
    Settings *UserSettings `json:"settings"`
    
    // Временные метки
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// UserSettings настройки пользователя.
type UserSettings struct {
    // Язык интерфейса (ru, en)
    Language string `json:"language"`
    
    // Тема оформления (light, dark, system)
    Theme string `json:"theme"`
    
    // Настройки уведомлений
    Notifications *NotificationSettings `json:"notifications"`
    
    // Единицы измерения (metric, imperial)
    Units string `json:"units"`
}

// NotificationSettings настройки уведомлений.
type NotificationSettings struct {
    // Email уведомления
    Email bool `json:"email"`
    
    // Push уведомления
    Push bool `json:"push"`
    
    // Уведомления о маркетинге
    Marketing bool `json:"marketing"`
}

// UpdateUserInput данные для обновления профиля.
type UpdateUserInput struct {
    // Новое имя (опционально)
    Name *string `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
    
    // Новые настройки (опционально)
    Settings *UserSettings `json:"settings,omitempty"`
}

// ChangePasswordInput данные для смены пароля.
type ChangePasswordInput struct {
    // Текущий пароль
    CurrentPassword string `json:"current_password" validate:"required"`
    
    // Новый пароль
    NewPassword string `json:"new_password" validate:"required,min=8,max=72,password_strength"`
}

// DeleteAccountInput данные для удаления аккаунта.
type DeleteAccountInput struct {
    // Пароль для подтверждения
    Password string `json:"password" validate:"required"`
    
    // Причина удаления (опционально)
    Reason string `json:"reason,omitempty" validate:"max=1000"`
}

// SessionResponse данные сессии.
type SessionResponse struct {
    // ID сессии
    ID string `json:"id"`
    
    // Описание устройства
    Device string `json:"device"`
    
    // IP адрес
    IP string `json:"ip"`
    
    // Геолокация (город, страна)
    Location string `json:"location"`
    
    // Текущая ли это сессия
    IsCurrent bool `json:"is_current"`
    
    // Последняя активность
    LastActive time.Time `json:"last_active"`
    
    // Время создания сессии
    CreatedAt time.Time `json:"created_at"`
}

// SessionsListResponse список сессий.
type SessionsListResponse struct {
    Sessions []SessionResponse `json:"sessions"`
    Total    int               `json:"total"`
}
```

## Права доступа

| Endpoint | Роль | Описание |
|----------|------|----------|
| `GET /users/me` | user, admin, expert | Получение своего профиля |
| `PATCH /users/me` | user, admin, expert | Обновление своего профиля |
| `PUT /users/me/avatar` | user, admin, expert | Загрузка аватара |
| `DELETE /users/me/avatar` | user, admin, expert | Удаление аватара |
| `PUT /users/me/password` | user, admin, expert | Смена пароля |
| `GET /users/me/sessions` | user, admin, expert | Просмотр сессий |
| `DELETE /users/me/sessions/:id` | user, admin, expert | Отзыв сессии |
| `DELETE /users/me` | user, admin | Удаление аккаунта |

## Admin Endpoints

### GET /api/v1/admin/users

Список пользователей (только admin).

**Request:**

```http
GET /api/v1/admin/users?page=1&per_page=20&role=user&search=example
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | int | 1 | Номер страницы |
| `per_page` | int | 20 | Записей на странице (max 100) |
| `role` | string | - | Фильтр по роли |
| `search` | string | - | Поиск по email/имени |
| `sort` | string | created_at | Поле сортировки |
| `order` | string | desc | Направление (asc/desc) |

**Response 200:**

```json
{
  "data": {
    "users": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "email": "user@example.com",
        "name": "Иван Петров",
        "role": "user",
        "email_verified": true,
        "created_at": "2024-01-15T10:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "per_page": 20,
      "total": 156,
      "total_pages": 8
    }
  },
  "request_id": "req_abc123"
}
```

---

### PATCH /api/v1/admin/users/:userId

Обновление пользователя админом.

**Request:**

```http
PATCH /api/v1/admin/users/550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "role": "expert",
  "email_verified": true
}
```

**Response 200:**

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "Иван Петров",
    "role": "expert",
    "email_verified": true,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-21T12:00:00Z"
  },
  "request_id": "req_abc123"
}
```

---

### DELETE /api/v1/admin/users/:userId

Блокировка/удаление пользователя админом.

**Request:**

```http
DELETE /api/v1/admin/users/550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "reason": "Нарушение правил использования"
}
```

**Response 200:**

```json
{
  "data": {
    "message": "User deactivated"
  },
  "request_id": "req_abc123"
}
```


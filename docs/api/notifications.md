# API уведомлений

## Обзор

Система уведомлений поддерживает:
- In-app уведомления (REST + WebSocket)
- Email уведомления
- Push уведомления (PWA)

## Типы уведомлений

| Type | Description | Channels |
|------|-------------|----------|
| `request_status` | Изменение статуса заявки | in-app, email, push |
| `compliance_warning` | Предупреждение о нарушении | in-app |
| `workspace_invite` | Приглашение в воркспейс | in-app, email |
| `ai_generation_complete` | Завершение генерации AI | in-app |
| `system` | Системные уведомления | in-app |

## Endpoints

### GET /api/v1/notifications

Список уведомлений пользователя.

**Request:**

```http
GET /api/v1/notifications?unread_only=true&limit=50
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | int | 50 | Количество (max 100) |
| `offset` | int | 0 | Смещение |
| `unread_only` | bool | false | Только непрочитанные |
| `type` | string | - | Фильтр по типу |

**Response 200:**

```json
{
  "data": {
    "notifications": [
      {
        "id": "notif_001",
        "type": "request_status",
        "title": "Заявка одобрена",
        "message": "Ваша заявка на оформление документации одобрена. Предварительная стоимость: 18 500 ₽",
        "data": {
          "request_id": "req_990e8400",
          "status": "approved",
          "estimated_price": 18500
        },
        "read": false,
        "read_at": null,
        "created_at": "2024-01-22T14:00:00Z"
      },
      {
        "id": "notif_002",
        "type": "workspace_invite",
        "title": "Приглашение в проект",
        "message": "Иван Петров приглашает вас в проект \"Квартира на Тверской\"",
        "data": {
          "workspace_id": "ws_550e8400",
          "workspace_name": "Квартира на Тверской",
          "invited_by": "Иван Петров",
          "role": "editor"
        },
        "read": true,
        "read_at": "2024-01-22T10:15:00Z",
        "created_at": "2024-01-22T10:00:00Z"
      }
    ],
    "unread_count": 5,
    "total": 23
  },
  "request_id": "req_abc123"
}
```

---

### GET /api/v1/notifications/count

Количество непрочитанных уведомлений.

**Request:**

```http
GET /api/v1/notifications/count
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "unread_count": 5,
    "by_type": {
      "request_status": 2,
      "compliance_warning": 1,
      "workspace_invite": 1,
      "system": 1
    }
  },
  "request_id": "req_abc123"
}
```

---

### POST /api/v1/notifications/:notificationId/read

Отметить уведомление как прочитанное.

**Request:**

```http
POST /api/v1/notifications/notif_001/read
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "id": "notif_001",
    "read": true,
    "read_at": "2024-01-22T15:00:00Z"
  },
  "request_id": "req_abc123"
}
```

---

### POST /api/v1/notifications/read-all

Отметить все уведомления как прочитанные.

**Request:**

```http
POST /api/v1/notifications/read-all
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "type": "workspace_invite"
}
```

**Request Body (optional):**

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | Отметить только определённый тип |
| `before` | datetime | Отметить уведомления до даты |

**Response 200:**

```json
{
  "data": {
    "marked_count": 3,
    "message": "Notifications marked as read"
  },
  "request_id": "req_abc123"
}
```

---

### DELETE /api/v1/notifications/:notificationId

Удаление уведомления.

**Request:**

```http
DELETE /api/v1/notifications/notif_001
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "message": "Notification deleted"
  },
  "request_id": "req_abc123"
}
```

---

### DELETE /api/v1/notifications

Удаление всех прочитанных уведомлений.

**Request:**

```http
DELETE /api/v1/notifications?read_only=true
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "deleted_count": 18,
    "message": "Notifications deleted"
  },
  "request_id": "req_abc123"
}
```

---

### GET /api/v1/notifications/settings

Настройки уведомлений пользователя.

**Request:**

```http
GET /api/v1/notifications/settings
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "email": {
      "enabled": true,
      "request_status": true,
      "workspace_invite": true,
      "marketing": false
    },
    "push": {
      "enabled": true,
      "request_status": true,
      "workspace_invite": false,
      "compliance_warning": true
    },
    "in_app": {
      "enabled": true,
      "sound": true
    }
  },
  "request_id": "req_abc123"
}
```

---

### PATCH /api/v1/notifications/settings

Обновление настроек уведомлений.

**Request:**

```http
PATCH /api/v1/notifications/settings
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "email": {
    "marketing": true
  },
  "push": {
    "workspace_invite": true
  }
}
```

**Response 200:**

```json
{
  "data": {
    "email": {
      "enabled": true,
      "request_status": true,
      "workspace_invite": true,
      "marketing": true
    },
    "push": {
      "enabled": true,
      "request_status": true,
      "workspace_invite": true,
      "compliance_warning": true
    },
    "in_app": {
      "enabled": true,
      "sound": true
    }
  },
  "request_id": "req_abc123"
}
```

---

### POST /api/v1/notifications/push/subscribe

Подписка на push-уведомления.

**Request:**

```http
POST /api/v1/notifications/push/subscribe
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "subscription": {
    "endpoint": "https://fcm.googleapis.com/fcm/send/...",
    "keys": {
      "p256dh": "BNcRdreALRFXTkOOUHK1EtK...",
      "auth": "tBHItJI5svbpez7KI4CCXg=="
    }
  },
  "device_info": {
    "type": "web",
    "browser": "Chrome",
    "os": "Windows"
  }
}
```

**Response 200:**

```json
{
  "data": {
    "subscription_id": "push_sub_001",
    "message": "Push subscription created"
  },
  "request_id": "req_abc123"
}
```

---

### DELETE /api/v1/notifications/push/unsubscribe

Отписка от push-уведомлений.

**Request:**

```http
DELETE /api/v1/notifications/push/unsubscribe
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "endpoint": "https://fcm.googleapis.com/fcm/send/..."
}
```

**Response 200:**

```json
{
  "data": {
    "message": "Push subscription removed"
  },
  "request_id": "req_abc123"
}
```

---

## WebSocket: Real-time уведомления

```javascript
const ws = new WebSocket('wss://api.granula.ru/ws');

// Аутентификация
ws.send(JSON.stringify({
  type: 'auth',
  token: 'access_token'
}));

// Автоматическая подписка на уведомления после auth

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  
  switch (data.type) {
    case 'notification':
      // Новое уведомление
      // {
      //   type: 'notification',
      //   data: {
      //     id: 'notif_003',
      //     type: 'request_status',
      //     title: 'Статус заявки изменён',
      //     message: 'Ваша заявка принята в работу',
      //     data: { request_id: '...', status: 'in_progress' },
      //     created_at: '...'
      //   }
      // }
      showNotification(data.data);
      break;
      
    case 'notification:count':
      // Обновление счётчика
      // { type: 'notification:count', data: { unread_count: 6 } }
      updateBadge(data.data.unread_count);
      break;
  }
};
```

---

## DTO Types

```go
// internal/dto/notification.go

// NotificationResponse уведомление.
type NotificationResponse struct {
    ID        string                 `json:"id"`
    Type      string                 `json:"type"`
    Title     string                 `json:"title"`
    Message   string                 `json:"message"`
    Data      map[string]interface{} `json:"data,omitempty"`
    Read      bool                   `json:"read"`
    ReadAt    *time.Time             `json:"read_at"`
    CreatedAt time.Time              `json:"created_at"`
}

// NotificationListResponse список уведомлений.
type NotificationListResponse struct {
    Notifications []NotificationResponse `json:"notifications"`
    UnreadCount   int                    `json:"unread_count"`
    Total         int                    `json:"total"`
}

// NotificationCountResponse счётчик уведомлений.
type NotificationCountResponse struct {
    UnreadCount int            `json:"unread_count"`
    ByType      map[string]int `json:"by_type"`
}

// NotificationSettingsResponse настройки.
type NotificationSettingsResponse struct {
    Email *EmailNotificationSettings `json:"email"`
    Push  *PushNotificationSettings  `json:"push"`
    InApp *InAppNotificationSettings `json:"in_app"`
}

// EmailNotificationSettings настройки email.
type EmailNotificationSettings struct {
    Enabled         bool `json:"enabled"`
    RequestStatus   bool `json:"request_status"`
    WorkspaceInvite bool `json:"workspace_invite"`
    Marketing       bool `json:"marketing"`
}

// PushNotificationSettings настройки push.
type PushNotificationSettings struct {
    Enabled           bool `json:"enabled"`
    RequestStatus     bool `json:"request_status"`
    WorkspaceInvite   bool `json:"workspace_invite"`
    ComplianceWarning bool `json:"compliance_warning"`
}

// InAppNotificationSettings настройки in-app.
type InAppNotificationSettings struct {
    Enabled bool `json:"enabled"`
    Sound   bool `json:"sound"`
}

// UpdateNotificationSettingsInput обновление настроек.
type UpdateNotificationSettingsInput struct {
    Email *EmailNotificationSettings `json:"email,omitempty"`
    Push  *PushNotificationSettings  `json:"push,omitempty"`
    InApp *InAppNotificationSettings `json:"in_app,omitempty"`
}

// PushSubscriptionInput подписка на push.
type PushSubscriptionInput struct {
    Subscription *WebPushSubscription `json:"subscription" validate:"required"`
    DeviceInfo   *DeviceInfo          `json:"device_info,omitempty"`
}

// WebPushSubscription данные подписки.
type WebPushSubscription struct {
    Endpoint string            `json:"endpoint" validate:"required,url"`
    Keys     *WebPushKeys      `json:"keys" validate:"required"`
}

// WebPushKeys ключи подписки.
type WebPushKeys struct {
    P256dh string `json:"p256dh" validate:"required"`
    Auth   string `json:"auth" validate:"required"`
}

// DeviceInfo информация об устройстве.
type DeviceInfo struct {
    Type    string `json:"type"`    // web, android, ios
    Browser string `json:"browser,omitempty"`
    OS      string `json:"os,omitempty"`
}

// MarkReadInput отметка о прочтении.
type MarkReadInput struct {
    Type   *string    `json:"type,omitempty"`
    Before *time.Time `json:"before,omitempty"`
}
```

## Email Templates

| Template | Event | Subject |
|----------|-------|---------|
| `request_approved` | Заявка одобрена | Ваша заявка одобрена |
| `request_rejected` | Заявка отклонена | Ваша заявка отклонена |
| `request_completed` | Заявка выполнена | Работа по заявке завершена |
| `workspace_invite` | Приглашение | Вас пригласили в проект |
| `password_reset` | Сброс пароля | Сброс пароля |
| `email_verify` | Подтверждение | Подтвердите email |

## Push Notification Payload

```json
{
  "title": "Заявка одобрена",
  "body": "Ваша заявка на оформление документации одобрена",
  "icon": "/icons/notification-192.png",
  "badge": "/icons/badge-72.png",
  "data": {
    "type": "request_status",
    "request_id": "req_990e8400",
    "url": "/requests/req_990e8400"
  },
  "actions": [
    {
      "action": "view",
      "title": "Посмотреть"
    },
    {
      "action": "dismiss",
      "title": "Закрыть"
    }
  ]
}
```


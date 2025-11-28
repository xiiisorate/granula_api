# API воркспейсов

## Обзор

Воркспейс — основная единица организации проектов ремонта. Каждый воркспейс содержит планировки, сцены и заявки, связанные с одной квартирой/помещением.

## Endpoints

### POST /api/v1/workspaces

Создание нового воркспейса.

**Request:**

```http
POST /api/v1/workspaces
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "name": "Квартира на Тверской",
  "description": "Ремонт двухкомнатной квартиры",
  "address": "г. Москва, ул. Тверская, д. 15, кв. 42",
  "total_area": 65.5,
  "rooms_count": 2
}
```

**Validation:**

| Поле | Правила |
|------|---------|
| `name` | Required, min 1, max 255 |
| `description` | Optional, max 2000 |
| `address` | Optional, max 500 |
| `total_area` | Optional, > 0, max 10000 |
| `rooms_count` | Optional, 1-100 |

**Response 201:**

```json
{
  "data": {
    "id": "ws_550e8400-e29b-41d4-a716-446655440000",
    "name": "Квартира на Тверской",
    "description": "Ремонт двухкомнатной квартиры",
    "address": "г. Москва, ул. Тверская, д. 15, кв. 42",
    "total_area": 65.5,
    "rooms_count": 2,
    "status": "draft",
    "owner": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Иван Петров",
      "email": "user@example.com"
    },
    "settings": {
      "units": "metric",
      "grid_size": 0.1,
      "wall_height": 2.7,
      "snap_to_grid": true,
      "show_dimensions": true
    },
    "preview_url": null,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  },
  "request_id": "req_abc123"
}
```

---

### GET /api/v1/workspaces

Список воркспейсов пользователя.

**Request:**

```http
GET /api/v1/workspaces?page=1&per_page=20&status=active
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | int | 1 | Номер страницы |
| `per_page` | int | 20 | Записей на странице (max 100) |
| `status` | string | - | Фильтр по статусу |
| `search` | string | - | Поиск по названию/адресу |
| `sort` | string | updated_at | Поле сортировки |
| `order` | string | desc | Направление (asc/desc) |

**Response 200:**

```json
{
  "data": {
    "workspaces": [
      {
        "id": "ws_550e8400-e29b-41d4-a716-446655440000",
        "name": "Квартира на Тверской",
        "description": "Ремонт двухкомнатной квартиры",
        "address": "г. Москва, ул. Тверская, д. 15, кв. 42",
        "total_area": 65.5,
        "rooms_count": 2,
        "status": "active",
        "preview_url": "https://storage.granula.ru/previews/ws_550e8400.jpg",
        "role": "owner",
        "floor_plans_count": 1,
        "scenes_count": 3,
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-20T15:45:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "per_page": 20,
      "total": 5,
      "total_pages": 1
    }
  },
  "request_id": "req_abc123"
}
```

---

### GET /api/v1/workspaces/:workspaceId

Получение воркспейса по ID.

**Request:**

```http
GET /api/v1/workspaces/ws_550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "id": "ws_550e8400-e29b-41d4-a716-446655440000",
    "name": "Квартира на Тверской",
    "description": "Ремонт двухкомнатной квартиры",
    "address": "г. Москва, ул. Тверская, д. 15, кв. 42",
    "total_area": 65.5,
    "rooms_count": 2,
    "status": "active",
    "owner": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Иван Петров",
      "email": "user@example.com"
    },
    "settings": {
      "units": "metric",
      "grid_size": 0.1,
      "wall_height": 2.7,
      "snap_to_grid": true,
      "show_dimensions": true
    },
    "preview_url": "https://storage.granula.ru/previews/ws_550e8400.jpg",
    "members": [
      {
        "user_id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "Иван Петров",
        "email": "user@example.com",
        "role": "owner",
        "joined_at": "2024-01-15T10:30:00Z"
      }
    ],
    "stats": {
      "floor_plans_count": 1,
      "scenes_count": 3,
      "branches_count": 8,
      "requests_count": 0
    },
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-20T15:45:00Z"
  },
  "request_id": "req_abc123"
}
```

---

### PATCH /api/v1/workspaces/:workspaceId

Обновление воркспейса.

**Request:**

```http
PATCH /api/v1/workspaces/ws_550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "name": "Квартира на Тверской (обновлено)",
  "status": "active",
  "settings": {
    "wall_height": 3.0
  }
}
```

**Required Role:** `owner` или `editor`

**Response 200:**

```json
{
  "data": {
    "id": "ws_550e8400-e29b-41d4-a716-446655440000",
    "name": "Квартира на Тверской (обновлено)",
    "status": "active",
    "settings": {
      "units": "metric",
      "grid_size": 0.1,
      "wall_height": 3.0,
      "snap_to_grid": true,
      "show_dimensions": true
    },
    "updated_at": "2024-01-21T12:00:00Z"
  },
  "request_id": "req_abc123"
}
```

---

### DELETE /api/v1/workspaces/:workspaceId

Удаление воркспейса (soft delete).

**Request:**

```http
DELETE /api/v1/workspaces/ws_550e8400-e29b-41d4-a716-446655440000
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Required Role:** `owner`

**Response 200:**

```json
{
  "data": {
    "message": "Workspace deleted"
  },
  "request_id": "req_abc123"
}
```

---

## Участники воркспейса

### GET /api/v1/workspaces/:workspaceId/members

Список участников.

**Request:**

```http
GET /api/v1/workspaces/ws_550e8400-e29b-41d4-a716-446655440000/members
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "members": [
      {
        "id": "mem_abc123",
        "user": {
          "id": "550e8400-e29b-41d4-a716-446655440000",
          "name": "Иван Петров",
          "email": "user@example.com",
          "avatar_url": "https://storage.granula.ru/avatars/550e8400.jpg"
        },
        "role": "owner",
        "invited_by": null,
        "joined_at": "2024-01-15T10:30:00Z"
      },
      {
        "id": "mem_def456",
        "user": {
          "id": "660e8400-e29b-41d4-a716-446655440001",
          "name": "Мария Иванова",
          "email": "maria@example.com",
          "avatar_url": null
        },
        "role": "editor",
        "invited_by": {
          "id": "550e8400-e29b-41d4-a716-446655440000",
          "name": "Иван Петров"
        },
        "joined_at": "2024-01-18T14:20:00Z"
      }
    ],
    "total": 2
  },
  "request_id": "req_abc123"
}
```

---

### POST /api/v1/workspaces/:workspaceId/members

Приглашение участника.

**Request:**

```http
POST /api/v1/workspaces/ws_550e8400-e29b-41d4-a716-446655440000/members
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "email": "colleague@example.com",
  "role": "editor"
}
```

**Required Role:** `owner` или `editor`

**Validation:**

| Поле | Правила |
|------|---------|
| `email` | Required, valid email |
| `role` | Required, one of: editor, viewer |

**Response 201:**

```json
{
  "data": {
    "id": "mem_ghi789",
    "user": {
      "id": "770e8400-e29b-41d4-a716-446655440002",
      "name": "Алексей Сидоров",
      "email": "colleague@example.com"
    },
    "role": "editor",
    "invited_by": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Иван Петров"
    },
    "joined_at": "2024-01-21T12:00:00Z"
  },
  "request_id": "req_abc123"
}
```

**Note:** Если пользователь не зарегистрирован, отправляется email-приглашение.

---

### PATCH /api/v1/workspaces/:workspaceId/members/:memberId

Изменение роли участника.

**Request:**

```http
PATCH /api/v1/workspaces/ws_550e8400-e29b-41d4-a716-446655440000/members/mem_def456
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "role": "viewer"
}
```

**Required Role:** `owner`

**Response 200:**

```json
{
  "data": {
    "id": "mem_def456",
    "role": "viewer",
    "updated_at": "2024-01-21T12:00:00Z"
  },
  "request_id": "req_abc123"
}
```

---

### DELETE /api/v1/workspaces/:workspaceId/members/:memberId

Удаление участника.

**Request:**

```http
DELETE /api/v1/workspaces/ws_550e8400-e29b-41d4-a716-446655440000/members/mem_def456
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Required Role:** `owner` (или сам участник может покинуть)

**Response 200:**

```json
{
  "data": {
    "message": "Member removed"
  },
  "request_id": "req_abc123"
}
```

---

### POST /api/v1/workspaces/:workspaceId/leave

Выход из воркспейса (для участника).

**Request:**

```http
POST /api/v1/workspaces/ws_550e8400-e29b-41d4-a716-446655440000/leave
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Note:** Владелец не может покинуть воркспейс.

**Response 200:**

```json
{
  "data": {
    "message": "Left workspace"
  },
  "request_id": "req_abc123"
}
```

---

## DTO Types

```go
// internal/dto/workspace.go

// CreateWorkspaceInput данные для создания воркспейса.
type CreateWorkspaceInput struct {
    // Название проекта
    // Required: true
    // MinLength: 1
    // MaxLength: 255
    Name string `json:"name" validate:"required,min=1,max=255,safe_string"`
    
    // Описание (опционально)
    // MaxLength: 2000
    Description string `json:"description,omitempty" validate:"max=2000,safe_string"`
    
    // Адрес квартиры (опционально)
    // MaxLength: 500
    Address string `json:"address,omitempty" validate:"max=500,safe_string"`
    
    // Общая площадь в м² (опционально)
    // Min: 0.1
    // Max: 10000
    TotalArea *float64 `json:"total_area,omitempty" validate:"omitempty,gt=0,lte=10000"`
    
    // Количество комнат (опционально)
    // Min: 1
    // Max: 100
    RoomsCount *int `json:"rooms_count,omitempty" validate:"omitempty,gte=1,lte=100"`
}

// UpdateWorkspaceInput данные для обновления воркспейса.
type UpdateWorkspaceInput struct {
    // Название (опционально)
    Name *string `json:"name,omitempty" validate:"omitempty,min=1,max=255,safe_string"`
    
    // Описание (опционально)
    Description *string `json:"description,omitempty" validate:"omitempty,max=2000,safe_string"`
    
    // Адрес (опционально)
    Address *string `json:"address,omitempty" validate:"omitempty,max=500,safe_string"`
    
    // Площадь (опционально)
    TotalArea *float64 `json:"total_area,omitempty" validate:"omitempty,gt=0,lte=10000"`
    
    // Комнаты (опционально)
    RoomsCount *int `json:"rooms_count,omitempty" validate:"omitempty,gte=1,lte=100"`
    
    // Статус (опционально)
    // Values: draft, active, completed, archived
    Status *string `json:"status,omitempty" validate:"omitempty,oneof=draft active completed archived"`
    
    // Настройки (частичное обновление)
    Settings *WorkspaceSettingsInput `json:"settings,omitempty"`
}

// WorkspaceSettingsInput настройки воркспейса.
type WorkspaceSettingsInput struct {
    // Единицы измерения
    Units *string `json:"units,omitempty" validate:"omitempty,oneof=metric imperial"`
    
    // Размер сетки в метрах
    GridSize *float64 `json:"grid_size,omitempty" validate:"omitempty,gt=0,lte=1"`
    
    // Высота стен по умолчанию
    WallHeight *float64 `json:"wall_height,omitempty" validate:"omitempty,gte=2,lte=10"`
    
    // Привязка к сетке
    SnapToGrid *bool `json:"snap_to_grid,omitempty"`
    
    // Показывать размеры
    ShowDimensions *bool `json:"show_dimensions,omitempty"`
}

// WorkspaceResponse ответ с данными воркспейса.
type WorkspaceResponse struct {
    ID           string                    `json:"id"`
    Name         string                    `json:"name"`
    Description  string                    `json:"description,omitempty"`
    Address      string                    `json:"address,omitempty"`
    TotalArea    *float64                  `json:"total_area,omitempty"`
    RoomsCount   *int                      `json:"rooms_count,omitempty"`
    Status       string                    `json:"status"`
    Owner        *UserBriefResponse        `json:"owner,omitempty"`
    Settings     *WorkspaceSettingsResponse `json:"settings"`
    PreviewURL   *string                   `json:"preview_url"`
    Members      []WorkspaceMemberResponse `json:"members,omitempty"`
    Stats        *WorkspaceStatsResponse   `json:"stats,omitempty"`
    Role         string                    `json:"role,omitempty"` // Роль текущего пользователя
    CreatedAt    time.Time                 `json:"created_at"`
    UpdatedAt    time.Time                 `json:"updated_at"`
}

// WorkspaceSettingsResponse настройки воркспейса.
type WorkspaceSettingsResponse struct {
    Units          string  `json:"units"`
    GridSize       float64 `json:"grid_size"`
    WallHeight     float64 `json:"wall_height"`
    SnapToGrid     bool    `json:"snap_to_grid"`
    ShowDimensions bool    `json:"show_dimensions"`
}

// WorkspaceStatsResponse статистика воркспейса.
type WorkspaceStatsResponse struct {
    FloorPlansCount int `json:"floor_plans_count"`
    ScenesCount     int `json:"scenes_count"`
    BranchesCount   int `json:"branches_count"`
    RequestsCount   int `json:"requests_count"`
}

// WorkspaceMemberResponse участник воркспейса.
type WorkspaceMemberResponse struct {
    ID        string              `json:"id"`
    User      *UserBriefResponse  `json:"user"`
    Role      string              `json:"role"`
    InvitedBy *UserBriefResponse  `json:"invited_by,omitempty"`
    JoinedAt  time.Time           `json:"joined_at"`
}

// UserBriefResponse краткая информация о пользователе.
type UserBriefResponse struct {
    ID        string  `json:"id"`
    Name      string  `json:"name"`
    Email     string  `json:"email,omitempty"`
    AvatarURL *string `json:"avatar_url,omitempty"`
}

// InviteMemberInput данные для приглашения.
type InviteMemberInput struct {
    // Email приглашаемого
    Email string `json:"email" validate:"required,email"`
    
    // Роль в воркспейсе
    // Values: editor, viewer
    Role string `json:"role" validate:"required,oneof=editor viewer"`
}

// UpdateMemberInput данные для обновления участника.
type UpdateMemberInput struct {
    // Новая роль
    // Values: editor, viewer
    Role string `json:"role" validate:"required,oneof=editor viewer"`
}

// WorkspaceListResponse список воркспейсов с пагинацией.
type WorkspaceListResponse struct {
    Workspaces []WorkspaceResponse  `json:"workspaces"`
    Pagination *PaginationResponse  `json:"pagination"`
}
```

## Статусы воркспейса

| Status | Description |
|--------|-------------|
| `draft` | Черновик, начальный статус |
| `active` | Активный проект в работе |
| `completed` | Проект завершён |
| `archived` | В архиве |

## Роли в воркспейсе

| Role | Permissions |
|------|-------------|
| `owner` | Полные права: редактирование, удаление, управление участниками |
| `editor` | Редактирование сцен, веток; приглашение участников |
| `viewer` | Только просмотр |


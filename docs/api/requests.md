# API заявок на экспертов

## Обзор

API для создания и отслеживания заявок на услуги БТИ:
- Консультации
- Оформление документации
- Выезд эксперта
- Полный комплекс услуг

## Статусы заявок

```
pending → reviewing → approved → in_progress → completed
                  ↓
              rejected
                  
(any state) → cancelled
```

| Status | Description |
|--------|-------------|
| `pending` | Ожидает рассмотрения |
| `reviewing` | На рассмотрении у эксперта |
| `approved` | Одобрена, ожидает оплаты/старта |
| `rejected` | Отклонена |
| `in_progress` | В работе |
| `completed` | Выполнена |
| `cancelled` | Отменена пользователем |

## Типы услуг

| Type | Description | Estimated Price |
|------|-------------|-----------------|
| `consultation` | Онлайн консультация | от 2000 ₽ |
| `documentation` | Оформление документов | от 15000 ₽ |
| `expert_visit` | Выезд эксперта | от 5000 ₽ |
| `full_service` | Полный комплекс | от 30000 ₽ |

## Endpoints

### POST /api/v1/requests

Создание заявки.

**Request:**

```http
POST /api/v1/requests
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "workspace_id": "ws_550e8400",
  "scene_id": "sc_770e8400",
  "branch_id": "br_880e8400",
  "service_type": "documentation",
  "contact_name": "Иван Петров",
  "contact_phone": "+7 (999) 123-45-67",
  "contact_email": "ivan@example.com",
  "preferred_contact_time": "Будни 10:00-18:00",
  "comment": "Хочу узаконить объединение кухни и гостиной. Стена ненесущая, система проверила — нарушений нет."
}
```

**Validation:**

| Field | Rules |
|-------|-------|
| `workspace_id` | Required, valid UUID |
| `scene_id` | Required, valid MongoDB ObjectId |
| `branch_id` | Optional |
| `service_type` | Required, one of: consultation, documentation, expert_visit, full_service |
| `contact_name` | Required, 2-255 chars |
| `contact_phone` | Required, valid Russian phone |
| `contact_email` | Required, valid email |
| `comment` | Optional, max 2000 chars |

**Response 201:**

```json
{
  "data": {
    "id": "req_990e8400-e29b-41d4-a716-446655440003",
    "workspace_id": "ws_550e8400",
    "scene_id": "sc_770e8400",
    "branch_id": "br_880e8400",
    "user_id": "550e8400",
    "service_type": "documentation",
    "status": "pending",
    "contact": {
      "name": "Иван Петров",
      "phone": "+7 (999) 123-45-67",
      "email": "ivan@example.com",
      "preferred_time": "Будни 10:00-18:00"
    },
    "comment": "Хочу узаконить объединение кухни и гостиной...",
    "assigned_expert": null,
    "estimated_date": null,
    "estimated_price": null,
    "status_history": [
      {
        "status": "pending",
        "changed_at": "2024-01-21T12:00:00Z",
        "changed_by": "550e8400",
        "comment": null
      }
    ],
    "created_at": "2024-01-21T12:00:00Z",
    "updated_at": "2024-01-21T12:00:00Z"
  },
  "request_id": "req_abc123"
}
```

---

### GET /api/v1/requests

Список заявок пользователя.

**Request:**

```http
GET /api/v1/requests?status=pending&page=1&per_page=20
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | int | 1 | Номер страницы |
| `per_page` | int | 20 | Записей на странице |
| `status` | string | - | Фильтр по статусу |
| `workspace_id` | string | - | Фильтр по воркспейсу |
| `service_type` | string | - | Фильтр по типу услуги |

**Response 200:**

```json
{
  "data": {
    "requests": [
      {
        "id": "req_990e8400",
        "workspace": {
          "id": "ws_550e8400",
          "name": "Квартира на Тверской"
        },
        "service_type": "documentation",
        "status": "reviewing",
        "contact_name": "Иван Петров",
        "assigned_expert": {
          "id": "exp_001",
          "name": "Анна Смирнова"
        },
        "estimated_date": "2024-02-01",
        "estimated_price": 18500.00,
        "created_at": "2024-01-21T12:00:00Z",
        "updated_at": "2024-01-22T10:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "per_page": 20,
      "total": 3,
      "total_pages": 1
    }
  },
  "request_id": "req_abc123"
}
```

---

### GET /api/v1/requests/:requestId

Детали заявки.

**Request:**

```http
GET /api/v1/requests/req_990e8400
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "id": "req_990e8400",
    "workspace": {
      "id": "ws_550e8400",
      "name": "Квартира на Тверской",
      "address": "г. Москва, ул. Тверская, д. 15, кв. 42"
    },
    "scene": {
      "id": "sc_770e8400",
      "name": "Основная планировка",
      "preview_url": "https://storage.granula.ru/previews/sc_770e8400.png"
    },
    "branch": {
      "id": "br_880e8400",
      "name": "Объединение кухни и гостиной",
      "preview_url": "https://storage.granula.ru/previews/br_880e8400.png"
    },
    "user": {
      "id": "550e8400",
      "name": "Иван Петров",
      "email": "ivan@example.com"
    },
    "service_type": "documentation",
    "status": "reviewing",
    "contact": {
      "name": "Иван Петров",
      "phone": "+7 (999) 123-45-67",
      "email": "ivan@example.com",
      "preferred_time": "Будни 10:00-18:00"
    },
    "comment": "Хочу узаконить объединение кухни и гостиной...",
    "assigned_expert": {
      "id": "exp_001",
      "name": "Анна Смирнова",
      "phone": "+7 (495) 123-45-67",
      "email": "anna@bti.ru"
    },
    "estimated_date": "2024-02-01",
    "estimated_price": 18500.00,
    "rejection_reason": null,
    "compliance_snapshot": {
      "checked_at": "2024-01-21T11:55:00Z",
      "is_compliant": true,
      "violations_count": 0,
      "warnings_count": 1
    },
    "documents": [
      {
        "id": "doc_001",
        "name": "Проект перепланировки.pdf",
        "type": "project",
        "url": "https://storage.granula.ru/requests/req_990e8400/project.pdf",
        "uploaded_at": "2024-01-23T14:00:00Z"
      }
    ],
    "status_history": [
      {
        "status": "pending",
        "changed_at": "2024-01-21T12:00:00Z",
        "changed_by": "Иван Петров",
        "comment": null
      },
      {
        "status": "reviewing",
        "changed_at": "2024-01-22T10:30:00Z",
        "changed_by": "Анна Смирнова",
        "comment": "Заявка принята в работу. Предварительная оценка: 18500 ₽"
      }
    ],
    "created_at": "2024-01-21T12:00:00Z",
    "updated_at": "2024-01-22T10:30:00Z"
  },
  "request_id": "req_abc123"
}
```

---

### PATCH /api/v1/requests/:requestId

Обновление заявки (пользователем).

**Request:**

```http
PATCH /api/v1/requests/req_990e8400
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "contact_phone": "+7 (999) 987-65-43",
  "comment": "Дополнительно: нужна консультация по вентиляции"
}
```

**Note:** Обновление доступно только в статусах `pending` и `reviewing`.

**Response 200:**

```json
{
  "data": {
    "id": "req_990e8400",
    "contact": {
      "phone": "+7 (999) 987-65-43"
    },
    "comment": "Дополнительно: нужна консультация по вентиляции",
    "updated_at": "2024-01-22T11:00:00Z"
  },
  "request_id": "req_abc123"
}
```

---

### POST /api/v1/requests/:requestId/cancel

Отмена заявки.

**Request:**

```http
POST /api/v1/requests/req_990e8400/cancel
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "reason": "Решил отложить ремонт"
}
```

**Response 200:**

```json
{
  "data": {
    "id": "req_990e8400",
    "status": "cancelled",
    "message": "Request cancelled"
  },
  "request_id": "req_abc123"
}
```

---

## Expert Endpoints (role: expert)

### GET /api/v1/expert/requests

Список заявок для эксперта.

**Request:**

```http
GET /api/v1/expert/requests?status=pending&assigned_to_me=false
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `status` | string | Фильтр по статусу |
| `assigned_to_me` | bool | Только назначенные мне |
| `service_type` | string | Фильтр по типу услуги |

**Response 200:**

```json
{
  "data": {
    "requests": [
      {
        "id": "req_990e8400",
        "workspace_name": "Квартира на Тверской",
        "address": "г. Москва, ул. Тверская, д. 15, кв. 42",
        "service_type": "documentation",
        "status": "pending",
        "contact_name": "Иван Петров",
        "compliance_status": "compliant",
        "created_at": "2024-01-21T12:00:00Z"
      }
    ],
    "total": 15
  },
  "request_id": "req_abc123"
}
```

---

### POST /api/v1/expert/requests/:requestId/assign

Назначение заявки себе.

**Request:**

```http
POST /api/v1/expert/requests/req_990e8400/assign
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "id": "req_990e8400",
    "status": "reviewing",
    "assigned_expert_id": "exp_001",
    "message": "Request assigned to you"
  },
  "request_id": "req_abc123"
}
```

---

### PATCH /api/v1/expert/requests/:requestId

Обновление заявки экспертом.

**Request:**

```http
PATCH /api/v1/expert/requests/req_990e8400
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "status": "approved",
  "estimated_date": "2024-02-01",
  "estimated_price": 18500.00,
  "comment": "Заявка одобрена. Предварительная стоимость: 18500 ₽. Срок выполнения: 2 недели."
}
```

**Allowed Status Transitions (expert):**

| From | To |
|------|-----|
| `pending` | `reviewing`, `rejected` |
| `reviewing` | `approved`, `rejected` |
| `approved` | `in_progress`, `cancelled` |
| `in_progress` | `completed` |

**Response 200:**

```json
{
  "data": {
    "id": "req_990e8400",
    "status": "approved",
    "estimated_date": "2024-02-01",
    "estimated_price": 18500.00,
    "updated_at": "2024-01-22T14:00:00Z"
  },
  "request_id": "req_abc123"
}
```

---

### POST /api/v1/expert/requests/:requestId/reject

Отклонение заявки.

**Request:**

```http
POST /api/v1/expert/requests/req_990e8400/reject
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "reason": "Планировка содержит нарушения несущих конструкций, которые невозможно узаконить"
}
```

**Response 200:**

```json
{
  "data": {
    "id": "req_990e8400",
    "status": "rejected",
    "rejection_reason": "Планировка содержит нарушения...",
    "message": "Request rejected"
  },
  "request_id": "req_abc123"
}
```

---

### POST /api/v1/expert/requests/:requestId/documents

Загрузка документа к заявке.

**Request:**

```http
POST /api/v1/expert/requests/req_990e8400/documents
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: multipart/form-data

--boundary
Content-Disposition: form-data; name="file"; filename="project.pdf"
Content-Type: application/pdf

<binary data>
--boundary
Content-Disposition: form-data; name="type"

project
--boundary--
```

**Document Types:**

| Type | Description |
|------|-------------|
| `project` | Проект перепланировки |
| `conclusion` | Техническое заключение |
| `act` | Акт выполненных работ |
| `other` | Прочее |

**Response 201:**

```json
{
  "data": {
    "id": "doc_002",
    "name": "project.pdf",
    "type": "project",
    "url": "https://storage.granula.ru/requests/req_990e8400/project.pdf",
    "uploaded_at": "2024-01-23T14:00:00Z"
  },
  "request_id": "req_abc123"
}
```

---

## DTO Types

```go
// internal/dto/request.go

// CreateRequestInput данные для создания заявки.
type CreateRequestInput struct {
    WorkspaceID          string  `json:"workspace_id" validate:"required,uuid"`
    SceneID              string  `json:"scene_id" validate:"required"`
    BranchID             *string `json:"branch_id,omitempty"`
    ServiceType          string  `json:"service_type" validate:"required,oneof=consultation documentation expert_visit full_service"`
    ContactName          string  `json:"contact_name" validate:"required,min=2,max=255"`
    ContactPhone         string  `json:"contact_phone" validate:"required,phone_ru"`
    ContactEmail         string  `json:"contact_email" validate:"required,email"`
    PreferredContactTime string  `json:"preferred_contact_time,omitempty" validate:"max=255"`
    Comment              string  `json:"comment,omitempty" validate:"max=2000"`
}

// UpdateRequestInput данные для обновления заявки.
type UpdateRequestInput struct {
    ContactName          *string `json:"contact_name,omitempty" validate:"omitempty,min=2,max=255"`
    ContactPhone         *string `json:"contact_phone,omitempty" validate:"omitempty,phone_ru"`
    ContactEmail         *string `json:"contact_email,omitempty" validate:"omitempty,email"`
    PreferredContactTime *string `json:"preferred_contact_time,omitempty" validate:"omitempty,max=255"`
    Comment              *string `json:"comment,omitempty" validate:"omitempty,max=2000"`
}

// ExpertUpdateRequestInput обновление экспертом.
type ExpertUpdateRequestInput struct {
    Status         *string  `json:"status,omitempty" validate:"omitempty,oneof=reviewing approved rejected in_progress completed"`
    EstimatedDate  *string  `json:"estimated_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
    EstimatedPrice *float64 `json:"estimated_price,omitempty" validate:"omitempty,gt=0"`
    Comment        string   `json:"comment,omitempty" validate:"max=2000"`
}

// CancelRequestInput отмена заявки.
type CancelRequestInput struct {
    Reason string `json:"reason,omitempty" validate:"max=1000"`
}

// RejectRequestInput отклонение заявки.
type RejectRequestInput struct {
    Reason string `json:"reason" validate:"required,min=10,max=2000"`
}

// RequestResponse ответ с данными заявки.
type RequestResponse struct {
    ID                 string                   `json:"id"`
    Workspace          *WorkspaceBriefResponse  `json:"workspace,omitempty"`
    Scene              *SceneBriefResponse      `json:"scene,omitempty"`
    Branch             *BranchBriefResponse     `json:"branch,omitempty"`
    User               *UserBriefResponse       `json:"user,omitempty"`
    ServiceType        string                   `json:"service_type"`
    Status             string                   `json:"status"`
    Contact            *ContactResponse         `json:"contact"`
    Comment            string                   `json:"comment,omitempty"`
    AssignedExpert     *ExpertResponse          `json:"assigned_expert,omitempty"`
    EstimatedDate      *string                  `json:"estimated_date,omitempty"`
    EstimatedPrice     *float64                 `json:"estimated_price,omitempty"`
    RejectionReason    *string                  `json:"rejection_reason,omitempty"`
    ComplianceSnapshot *ComplianceSnapshotResponse `json:"compliance_snapshot,omitempty"`
    Documents          []DocumentResponse       `json:"documents,omitempty"`
    StatusHistory      []StatusHistoryEntry     `json:"status_history,omitempty"`
    CreatedAt          time.Time                `json:"created_at"`
    UpdatedAt          time.Time                `json:"updated_at"`
}

// ContactResponse контактные данные.
type ContactResponse struct {
    Name          string `json:"name"`
    Phone         string `json:"phone"`
    Email         string `json:"email"`
    PreferredTime string `json:"preferred_time,omitempty"`
}

// ExpertResponse данные эксперта.
type ExpertResponse struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Phone string `json:"phone,omitempty"`
    Email string `json:"email,omitempty"`
}

// StatusHistoryEntry запись истории статусов.
type StatusHistoryEntry struct {
    Status    string    `json:"status"`
    ChangedAt time.Time `json:"changed_at"`
    ChangedBy string    `json:"changed_by"`
    Comment   *string   `json:"comment,omitempty"`
}

// DocumentResponse документ.
type DocumentResponse struct {
    ID         string    `json:"id"`
    Name       string    `json:"name"`
    Type       string    `json:"type"`
    URL        string    `json:"url"`
    UploadedAt time.Time `json:"uploaded_at"`
}
```

## Уведомления

При изменении статуса заявки автоматически отправляются уведомления:

| Event | User Notification | Expert Notification |
|-------|-------------------|---------------------|
| Created | ✓ | ✓ (all experts) |
| Assigned | ✓ | - |
| Approved | ✓ (email + push) | - |
| Rejected | ✓ (email + push) | - |
| In Progress | ✓ | - |
| Completed | ✓ (email + push) | - |
| Cancelled | - | ✓ |


# API веток дизайна

## Обзор

Ветка (Branch) — вариант планировки, созданный на основе сцены или другой ветки. Система веток позволяет:
- Создавать множество вариантов без потери оригинала
- AI-генерацию альтернативных планировок
- Иерархическое наследование изменений
- Сравнение вариантов

## Концепция

```
Scene (исходная планировка)
├── Branch 1: "Объединение кухни" (user)
│   ├── Branch 1.1: "Вариант с барной стойкой" (ai)
│   └── Branch 1.2: "Вариант с островом" (ai)
├── Branch 2: "Расширение спальни" (user)
│   └── Branch 2.1: "С гардеробной" (ai)
└── Branch 3: "AI предложение 1" (ai)
```

## Endpoints

### POST /api/v1/scenes/:sceneId/branches

Создание новой ветки.

**Request:**

```http
POST /api/v1/scenes/sc_770e8400/branches
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "name": "Объединение кухни и гостиной",
  "description": "Снос перегородки между кухней и гостиной",
  "parent_branch_id": null
}
```

**Response 201:**

```json
{
  "data": {
    "id": "br_880e8400",
    "scene_id": "sc_770e8400",
    "parent_branch_id": null,
    "name": "Объединение кухни и гостиной",
    "description": "Снос перегородки между кухней и гостиной",
    "source": "user",
    "order": 0,
    "is_active": true,
    "is_favorite": false,
    "delta": {
      "added": {},
      "modified": {},
      "removed": []
    },
    "snapshot": null,
    "compliance_result": null,
    "preview_url": null,
    "created_at": "2024-01-21T12:00:00Z",
    "created_by": "550e8400"
  },
  "request_id": "req_abc123"
}
```

---

### GET /api/v1/scenes/:sceneId/branches

Список веток сцены.

**Request:**

```http
GET /api/v1/scenes/sc_770e8400/branches?include_tree=true
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `include_tree` | bool | false | Включить иерархическую структуру |
| `source` | string | - | Фильтр по источнику (user/ai) |
| `is_favorite` | bool | - | Только избранные |

**Response 200 (flat):**

```json
{
  "data": {
    "branches": [
      {
        "id": "br_880e8400",
        "name": "Объединение кухни и гостиной",
        "parent_branch_id": null,
        "source": "user",
        "is_active": true,
        "is_favorite": false,
        "compliance_status": "compliant",
        "preview_url": "https://storage.granula.ru/previews/br_880e8400.png",
        "children_count": 2,
        "created_at": "2024-01-21T12:00:00Z"
      },
      {
        "id": "br_890e8400",
        "name": "Вариант с барной стойкой",
        "parent_branch_id": "br_880e8400",
        "source": "ai",
        "is_active": false,
        "is_favorite": true,
        "compliance_status": "compliant",
        "preview_url": "https://storage.granula.ru/previews/br_890e8400.png",
        "children_count": 0,
        "created_at": "2024-01-21T12:05:00Z"
      }
    ],
    "total": 2
  },
  "request_id": "req_abc123"
}
```

**Response 200 (tree):**

```json
{
  "data": {
    "tree": [
      {
        "id": "br_880e8400",
        "name": "Объединение кухни и гостиной",
        "source": "user",
        "is_active": true,
        "compliance_status": "compliant",
        "children": [
          {
            "id": "br_890e8400",
            "name": "Вариант с барной стойкой",
            "source": "ai",
            "is_active": false,
            "compliance_status": "compliant",
            "children": []
          },
          {
            "id": "br_891e8400",
            "name": "Вариант с островом",
            "source": "ai",
            "is_active": false,
            "compliance_status": "warning",
            "children": []
          }
        ]
      }
    ]
  },
  "request_id": "req_abc123"
}
```

---

### GET /api/v1/scenes/:sceneId/branches/:branchId

Получение ветки с полными данными.

**Request:**

```http
GET /api/v1/scenes/sc_770e8400/branches/br_880e8400
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "id": "br_880e8400",
    "scene_id": "sc_770e8400",
    "parent_branch_id": null,
    "name": "Объединение кухни и гостиной",
    "description": "Снос перегородки между кухней и гостиной",
    "source": "user",
    "order": 0,
    "is_active": true,
    "is_favorite": false,
    "delta": {
      "added": {
        "furniture": [
          {
            "id": "furn_new_001",
            "type": "furniture",
            "name": "Барная стойка",
            "furniture_type": "bar",
            "position": { "x": 4.0, "y": 0, "z": 2.0 }
          }
        ]
      },
      "modified": {
        "wall_002": {
          "end": { "x": 2.0, "y": 0, "z": 0 }
        },
        "room_001": {
          "polygon": [...],
          "area": 28.0
        }
      },
      "removed": ["wall_003"]
    },
    "snapshot": {
      "elements": {
        "walls": [...],
        "rooms": [...],
        "furniture": [...],
        "utilities": [...]
      },
      "bounds": {...},
      "stats": {
        "total_area": 65.5,
        "rooms_count": 3,
        "walls_count": 11,
        "furniture_count": 1
      }
    },
    "compliance_result": {
      "last_checked_at": "2024-01-21T12:10:00Z",
      "is_compliant": true,
      "violations": [],
      "warnings": []
    },
    "ai_context": null,
    "preview_url": "https://storage.granula.ru/previews/br_880e8400.png",
    "created_at": "2024-01-21T12:00:00Z",
    "updated_at": "2024-01-21T12:10:00Z",
    "created_by": "550e8400"
  },
  "request_id": "req_abc123"
}
```

---

### PATCH /api/v1/scenes/:sceneId/branches/:branchId

Обновление ветки.

**Request:**

```http
PATCH /api/v1/scenes/sc_770e8400/branches/br_880e8400
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "name": "Объединение кухни (финальный)",
  "is_favorite": true
}
```

**Response 200:**

```json
{
  "data": {
    "id": "br_880e8400",
    "name": "Объединение кухни (финальный)",
    "is_favorite": true,
    "updated_at": "2024-01-21T13:00:00Z"
  },
  "request_id": "req_abc123"
}
```

---

### PUT /api/v1/scenes/:sceneId/branches/:branchId/delta

Обновление изменений ветки.

**Request:**

```http
PUT /api/v1/scenes/sc_770e8400/branches/br_880e8400/delta
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "delta": {
    "added": {
      "furniture": [...]
    },
    "modified": {
      "wall_002": {
        "end": { "x": 1.5, "y": 0, "z": 0 }
      }
    },
    "removed": ["wall_003", "wall_004"]
  }
}
```

**Response 200:**

```json
{
  "data": {
    "updated": true,
    "snapshot_updated": true,
    "compliance_result": {
      "is_compliant": false,
      "violations": [
        {
          "rule_code": "SNIP_LOAD_BEARING",
          "severity": "error",
          "message": "Нельзя удалять несущую стену wall_004"
        }
      ]
    }
  },
  "request_id": "req_abc123"
}
```

---

### POST /api/v1/scenes/:sceneId/branches/:branchId/activate

Активация ветки (применение к сцене).

**Request:**

```http
POST /api/v1/scenes/sc_770e8400/branches/br_880e8400/activate
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "id": "br_880e8400",
    "is_active": true,
    "previous_active_branch": "br_870e8400",
    "message": "Branch activated"
  },
  "request_id": "req_abc123"
}
```

---

### DELETE /api/v1/scenes/:sceneId/branches/:branchId

Удаление ветки.

**Request:**

```http
DELETE /api/v1/scenes/sc_770e8400/branches/br_880e8400
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "message": "Branch deleted",
    "children_deleted": 2
  },
  "request_id": "req_abc123"
}
```

**Note:** Удаление ветки удаляет все дочерние ветки.

---

### POST /api/v1/scenes/:sceneId/branches/:branchId/duplicate

Дублирование ветки.

**Request:**

```http
POST /api/v1/scenes/sc_770e8400/branches/br_880e8400/duplicate
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "name": "Копия варианта",
  "include_children": false
}
```

**Response 201:**

```json
{
  "data": {
    "id": "br_900e8400",
    "name": "Копия варианта",
    "source_branch_id": "br_880e8400",
    "created_at": "2024-01-21T13:00:00Z"
  },
  "request_id": "req_abc123"
}
```

---

### GET /api/v1/scenes/:sceneId/branches/:branchId/compare/:targetBranchId

Сравнение двух веток.

**Request:**

```http
GET /api/v1/scenes/sc_770e8400/branches/br_880e8400/compare/br_890e8400
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "source_branch": {
      "id": "br_880e8400",
      "name": "Объединение кухни"
    },
    "target_branch": {
      "id": "br_890e8400",
      "name": "Вариант с барной стойкой"
    },
    "differences": {
      "elements_added": [
        {
          "type": "furniture",
          "id": "furn_bar_001",
          "name": "Барная стойка"
        }
      ],
      "elements_removed": [],
      "elements_modified": [
        {
          "type": "room",
          "id": "room_001",
          "changes": {
            "area": {
              "from": 28.0,
              "to": 26.5
            }
          }
        }
      ]
    },
    "stats_comparison": {
      "total_area": { "source": 65.5, "target": 65.5 },
      "rooms_count": { "source": 3, "target": 3 },
      "furniture_count": { "source": 0, "target": 1 }
    },
    "compliance_comparison": {
      "source": { "is_compliant": true, "violations": 0, "warnings": 0 },
      "target": { "is_compliant": true, "violations": 0, "warnings": 1 }
    }
  },
  "request_id": "req_abc123"
}
```

---

### POST /api/v1/scenes/:sceneId/branches/:branchId/merge

Слияние ветки в родительскую.

**Request:**

```http
POST /api/v1/scenes/sc_770e8400/branches/br_890e8400/merge
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "strategy": "replace",
  "delete_source": true
}
```

**Merge Strategies:**

| Strategy | Description |
|----------|-------------|
| `replace` | Полная замена родительской ветки |
| `combine` | Объединение изменений |

**Response 200:**

```json
{
  "data": {
    "merged": true,
    "target_branch_id": "br_880e8400",
    "source_deleted": true,
    "conflicts": []
  },
  "request_id": "req_abc123"
}
```

---

## DTO Types

```go
// internal/dto/branch.go

// CreateBranchInput данные для создания ветки.
type CreateBranchInput struct {
    // Название ветки
    Name string `json:"name" validate:"required,min=1,max=255"`
    
    // Описание
    Description string `json:"description,omitempty" validate:"max=2000"`
    
    // ID родительской ветки (null для корневой)
    ParentBranchID *string `json:"parent_branch_id,omitempty"`
}

// UpdateBranchInput данные для обновления ветки.
type UpdateBranchInput struct {
    Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
    Description *string `json:"description,omitempty" validate:"omitempty,max=2000"`
    IsFavorite  *bool   `json:"is_favorite,omitempty"`
    Order       *int    `json:"order,omitempty"`
}

// BranchDelta изменения ветки относительно родителя.
type BranchDelta struct {
    // Добавленные элементы
    Added map[string][]interface{} `json:"added"`
    
    // Изменённые элементы (id -> changes)
    Modified map[string]interface{} `json:"modified"`
    
    // Удалённые элементы (список id)
    Removed []string `json:"removed"`
}

// BranchResponse ответ с данными ветки.
type BranchResponse struct {
    ID             string                  `json:"id"`
    SceneID        string                  `json:"scene_id"`
    ParentBranchID *string                 `json:"parent_branch_id"`
    Name           string                  `json:"name"`
    Description    string                  `json:"description,omitempty"`
    Source         string                  `json:"source"` // user, ai
    Order          int                     `json:"order"`
    IsActive       bool                    `json:"is_active"`
    IsFavorite     bool                    `json:"is_favorite"`
    Delta          *BranchDelta            `json:"delta,omitempty"`
    Snapshot       *BranchSnapshot         `json:"snapshot,omitempty"`
    ComplianceResult *ComplianceResult     `json:"compliance_result,omitempty"`
    AIContext      *AIContext              `json:"ai_context,omitempty"`
    PreviewURL     *string                 `json:"preview_url"`
    ChildrenCount  int                     `json:"children_count,omitempty"`
    CreatedAt      time.Time               `json:"created_at"`
    UpdatedAt      time.Time               `json:"updated_at"`
    CreatedBy      string                  `json:"created_by"`
}

// BranchSnapshot полный снимок состояния ветки.
type BranchSnapshot struct {
    Elements SceneElements `json:"elements"`
    Bounds   BoundsData    `json:"bounds"`
    Stats    SceneStats    `json:"stats"`
}

// AIContext контекст AI-генерации.
type AIContext struct {
    Prompt      string    `json:"prompt"`
    Model       string    `json:"model"`
    GeneratedAt time.Time `json:"generated_at"`
    Reasoning   string    `json:"reasoning,omitempty"`
}

// BranchTreeNode узел дерева веток.
type BranchTreeNode struct {
    ID               string            `json:"id"`
    Name             string            `json:"name"`
    Source           string            `json:"source"`
    IsActive         bool              `json:"is_active"`
    IsFavorite       bool              `json:"is_favorite"`
    ComplianceStatus string            `json:"compliance_status"`
    PreviewURL       *string           `json:"preview_url"`
    Children         []BranchTreeNode  `json:"children"`
}

// CompareBranchesResponse результат сравнения веток.
type CompareBranchesResponse struct {
    SourceBranch       BranchBriefResponse      `json:"source_branch"`
    TargetBranch       BranchBriefResponse      `json:"target_branch"`
    Differences        BranchDifferences        `json:"differences"`
    StatsComparison    StatsComparison          `json:"stats_comparison"`
    ComplianceComparison ComplianceComparison   `json:"compliance_comparison"`
}

// BranchDifferences различия между ветками.
type BranchDifferences struct {
    ElementsAdded    []ElementDiff `json:"elements_added"`
    ElementsRemoved  []ElementDiff `json:"elements_removed"`
    ElementsModified []ElementDiff `json:"elements_modified"`
}
```

## Источники веток

| Source | Description |
|--------|-------------|
| `user` | Создана пользователем вручную |
| `ai` | Сгенерирована AI алгоритмом |

## Compliance статусы

| Status | Description |
|--------|-------------|
| `compliant` | Соответствует всем нормам |
| `warning` | Есть предупреждения |
| `violation` | Есть нарушения |
| `unchecked` | Не проверено |


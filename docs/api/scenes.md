# API 3D сцен

## Обзор

Сцена (Scene) — 3D модель квартиры, созданная на основе распознанной планировки. Содержит все элементы: стены, комнаты, мебель, инженерные системы.

## Endpoints

### POST /api/v1/workspaces/:workspaceId/scenes

Создание новой сцены.

**Request:**

```http
POST /api/v1/workspaces/ws_550e8400/scenes
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "name": "Новая планировка",
  "description": "Планировка с нуля",
  "bounds": {
    "width": 10.0,
    "height": 2.7,
    "depth": 8.0
  }
}
```

**Response 201:**

```json
{
  "data": {
    "id": "sc_770e8400-e29b-41d4-a716-446655440002",
    "workspace_id": "ws_550e8400",
    "floor_plan_id": null,
    "name": "Новая планировка",
    "description": "Планировка с нуля",
    "bounds": {
      "width": 10.0,
      "height": 2.7,
      "depth": 8.0
    },
    "elements": {
      "walls": [],
      "rooms": [],
      "furniture": [],
      "utilities": []
    },
    "display_settings": {
      "floor_texture": "wood_oak",
      "wall_color": "#FFFFFF",
      "ceiling_color": "#F5F5F5",
      "ambient_light": 0.6,
      "show_grid": true,
      "grid_size": 0.5
    },
    "compliance_result": null,
    "stats": {
      "total_area": 80.0,
      "rooms_count": 0,
      "walls_count": 0,
      "furniture_count": 0
    },
    "created_at": "2024-01-21T12:00:00Z",
    "updated_at": "2024-01-21T12:00:00Z"
  },
  "request_id": "req_abc123"
}
```

---

### GET /api/v1/workspaces/:workspaceId/scenes

Список сцен воркспейса.

**Request:**

```http
GET /api/v1/workspaces/ws_550e8400/scenes
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "scenes": [
      {
        "id": "sc_770e8400",
        "name": "Основная планировка",
        "description": "Из техпаспорта БТИ",
        "floor_plan_id": "fp_660e8400",
        "preview_url": "https://storage.granula.ru/renders/sc_770e8400/preview.png",
        "compliance_status": "compliant",
        "branches_count": 5,
        "stats": {
          "total_area": 65.5,
          "rooms_count": 4,
          "walls_count": 12,
          "furniture_count": 0
        },
        "created_at": "2024-01-21T12:00:00Z",
        "updated_at": "2024-01-21T15:30:00Z"
      }
    ],
    "total": 1
  },
  "request_id": "req_abc123"
}
```

---

### GET /api/v1/workspaces/:workspaceId/scenes/:sceneId

Получение полных данных сцены.

**Request:**

```http
GET /api/v1/workspaces/ws_550e8400/scenes/sc_770e8400
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "id": "sc_770e8400",
    "workspace_id": "ws_550e8400",
    "floor_plan_id": "fp_660e8400",
    "name": "Основная планировка",
    "description": "Из техпаспорта БТИ",
    "schema_version": 1,
    "bounds": {
      "width": 12.5,
      "height": 2.7,
      "depth": 8.3
    },
    "elements": {
      "walls": [
        {
          "id": "wall_001",
          "type": "wall",
          "name": "Несущая стена 1",
          "start": { "x": 0, "y": 0, "z": 0 },
          "end": { "x": 5.0, "y": 0, "z": 0 },
          "height": 2.7,
          "thickness": 0.2,
          "properties": {
            "is_load_bearing": true,
            "material": "brick",
            "can_demolish": false
          },
          "openings": [
            {
              "id": "opening_001",
              "type": "door",
              "position": 2.0,
              "width": 0.9,
              "height": 2.1,
              "elevation": 0
            }
          ],
          "metadata": {
            "locked": false,
            "visible": true,
            "selected": false
          }
        }
      ],
      "rooms": [
        {
          "id": "room_001",
          "type": "room",
          "name": "Кухня",
          "room_type": "kitchen",
          "polygon": [
            { "x": 0, "z": 0 },
            { "x": 4.0, "z": 0 },
            { "x": 4.0, "z": 3.5 },
            { "x": 0, "z": 3.5 }
          ],
          "area": 14.0,
          "perimeter": 15.0,
          "properties": {
            "has_wet_zone": true,
            "has_ventilation": true,
            "min_area": 5.0
          }
        }
      ],
      "furniture": [],
      "utilities": [
        {
          "id": "utility_001",
          "type": "utility",
          "name": "Стояк отопления",
          "utility_type": "heating_riser",
          "position": { "x": 0.5, "y": 0, "z": 2.0 },
          "properties": {
            "can_relocate": false,
            "protection_zone": 0.3
          }
        }
      ]
    },
    "display_settings": {
      "floor_texture": "wood_oak",
      "wall_color": "#FFFFFF",
      "ceiling_color": "#F5F5F5",
      "ambient_light": 0.6,
      "show_grid": true,
      "grid_size": 0.5
    },
    "compliance_result": {
      "last_checked_at": "2024-01-21T15:30:00Z",
      "is_compliant": true,
      "violations": [],
      "warnings": []
    },
    "stats": {
      "total_area": 65.5,
      "rooms_count": 4,
      "walls_count": 12,
      "furniture_count": 0
    },
    "created_at": "2024-01-21T12:00:00Z",
    "updated_at": "2024-01-21T15:30:00Z",
    "created_by": "550e8400",
    "updated_by": "550e8400"
  },
  "request_id": "req_abc123"
}
```

---

### PATCH /api/v1/workspaces/:workspaceId/scenes/:sceneId

Обновление метаданных сцены.

**Request:**

```http
PATCH /api/v1/workspaces/ws_550e8400/scenes/sc_770e8400
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "name": "Основная планировка (исправлено)",
  "display_settings": {
    "floor_texture": "wood_walnut",
    "wall_color": "#F0F0F0"
  }
}
```

**Response 200:**

```json
{
  "data": {
    "id": "sc_770e8400",
    "name": "Основная планировка (исправлено)",
    "display_settings": {
      "floor_texture": "wood_walnut",
      "wall_color": "#F0F0F0",
      "ceiling_color": "#F5F5F5",
      "ambient_light": 0.6,
      "show_grid": true,
      "grid_size": 0.5
    },
    "updated_at": "2024-01-21T16:00:00Z"
  },
  "request_id": "req_abc123"
}
```

---

### PUT /api/v1/workspaces/:workspaceId/scenes/:sceneId/elements

Обновление элементов сцены (полная замена).

**Request:**

```http
PUT /api/v1/workspaces/ws_550e8400/scenes/sc_770e8400/elements
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "elements": {
    "walls": [...],
    "rooms": [...],
    "furniture": [...],
    "utilities": [...]
  }
}
```

**Response 200:**

```json
{
  "data": {
    "updated": true,
    "compliance_result": {
      "is_compliant": true,
      "violations": [],
      "warnings": []
    },
    "stats": {
      "total_area": 65.5,
      "rooms_count": 4,
      "walls_count": 11,
      "furniture_count": 5
    }
  },
  "request_id": "req_abc123"
}
```

---

### PATCH /api/v1/workspaces/:workspaceId/scenes/:sceneId/elements

Частичное обновление элементов (delta update).

**Request:**

```http
PATCH /api/v1/workspaces/ws_550e8400/scenes/sc_770e8400/elements
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "operations": [
    {
      "op": "add",
      "type": "furniture",
      "element": {
        "id": "furn_001",
        "type": "furniture",
        "name": "Диван",
        "furniture_type": "sofa",
        "position": { "x": 2.0, "y": 0, "z": 1.5 },
        "rotation": { "x": 0, "y": 90, "z": 0 },
        "dimensions": {
          "width": 2.0,
          "height": 0.85,
          "depth": 0.9
        }
      }
    },
    {
      "op": "update",
      "type": "wall",
      "id": "wall_001",
      "changes": {
        "end": { "x": 4.0, "y": 0, "z": 0 }
      }
    },
    {
      "op": "remove",
      "type": "wall",
      "id": "wall_002"
    }
  ]
}
```

**Operations:**

| Op | Description |
|----|-------------|
| `add` | Добавить элемент |
| `update` | Обновить элемент |
| `remove` | Удалить элемент |

**Response 200:**

```json
{
  "data": {
    "applied_operations": 3,
    "compliance_result": {
      "is_compliant": false,
      "violations": [
        {
          "rule_code": "SNIP_LOAD_BEARING",
          "severity": "error",
          "message": "Нельзя удалять несущую стену",
          "affected_elements": ["wall_002"]
        }
      ],
      "warnings": []
    },
    "rollback_applied": true
  },
  "request_id": "req_abc123"
}
```

**Note:** Если операция нарушает критические правила, все изменения откатываются.

---

### DELETE /api/v1/workspaces/:workspaceId/scenes/:sceneId

Удаление сцены.

**Request:**

```http
DELETE /api/v1/workspaces/ws_550e8400/scenes/sc_770e8400
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "message": "Scene deleted"
  },
  "request_id": "req_abc123"
}
```

---

### POST /api/v1/workspaces/:workspaceId/scenes/:sceneId/duplicate

Дублирование сцены.

**Request:**

```http
POST /api/v1/workspaces/ws_550e8400/scenes/sc_770e8400/duplicate
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "name": "Копия планировки"
}
```

**Response 201:**

```json
{
  "data": {
    "id": "sc_880e8400",
    "name": "Копия планировки",
    "source_scene_id": "sc_770e8400",
    "created_at": "2024-01-21T16:00:00Z"
  },
  "request_id": "req_abc123"
}
```

---

### POST /api/v1/workspaces/:workspaceId/scenes/:sceneId/render

Запрос рендера сцены.

**Request:**

```http
POST /api/v1/workspaces/ws_550e8400/scenes/sc_770e8400/render
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "view_type": "3d",
  "resolution": {
    "width": 1920,
    "height": 1080
  },
  "camera": {
    "position": { "x": 10, "y": 8, "z": 10 },
    "target": { "x": 5, "y": 0, "z": 4 }
  },
  "quality": "high"
}
```

**View Types:**

| Type | Description |
|------|-------------|
| `2d` | Вид сверху (план) |
| `3d` | 3D изометрия |
| `first_person` | От первого лица |

**Response 202:**

```json
{
  "data": {
    "render_id": "render_990e8400",
    "status": "processing",
    "estimated_time_seconds": 30
  },
  "request_id": "req_abc123"
}
```

---

### GET /api/v1/workspaces/:workspaceId/scenes/:sceneId/renders/:renderId

Получение результата рендера.

**Response 200:**

```json
{
  "data": {
    "render_id": "render_990e8400",
    "status": "completed",
    "url": "https://storage.granula.ru/renders/sc_770e8400/render_990e8400.png",
    "thumbnail_url": "https://storage.granula.ru/renders/sc_770e8400/render_990e8400_thumb.png",
    "created_at": "2024-01-21T16:01:00Z",
    "expires_at": "2024-01-28T16:01:00Z"
  },
  "request_id": "req_abc123"
}
```

---

## Scene Elements Schema

```go
// internal/dto/scene.go

// SceneElements все элементы сцены.
type SceneElements struct {
    Walls     []WallElement      `json:"walls"`
    Rooms     []RoomElement      `json:"rooms"`
    Furniture []FurnitureElement `json:"furniture"`
    Utilities []UtilityElement   `json:"utilities"`
}

// WallElement элемент стены.
type WallElement struct {
    ID         string              `json:"id"`
    Type       string              `json:"type"` // always "wall"
    Name       string              `json:"name"`
    Start      Point3D             `json:"start"`
    End        Point3D             `json:"end"`
    Height     float64             `json:"height"`
    Thickness  float64             `json:"thickness"`
    Properties WallProperties      `json:"properties"`
    Openings   []OpeningElement    `json:"openings,omitempty"`
    Metadata   ElementMetadata     `json:"metadata"`
}

// WallProperties свойства стены.
type WallProperties struct {
    IsLoadBearing bool   `json:"is_load_bearing"`
    Material      string `json:"material"`     // brick, concrete, drywall, glass
    CanDemolish   bool   `json:"can_demolish"`
}

// OpeningElement проём в стене.
type OpeningElement struct {
    ID        string  `json:"id"`
    Type      string  `json:"type"` // door, window
    Position  float64 `json:"position"`
    Width     float64 `json:"width"`
    Height    float64 `json:"height"`
    Elevation float64 `json:"elevation,omitempty"`
}

// RoomElement элемент комнаты.
type RoomElement struct {
    ID         string         `json:"id"`
    Type       string         `json:"type"` // always "room"
    Name       string         `json:"name"`
    RoomType   string         `json:"room_type"` // kitchen, bedroom, bathroom, living, corridor, storage
    Polygon    []Point2D      `json:"polygon"`
    Area       float64        `json:"area"`
    Perimeter  float64        `json:"perimeter"`
    Properties RoomProperties `json:"properties"`
}

// RoomProperties свойства комнаты.
type RoomProperties struct {
    HasWetZone     bool    `json:"has_wet_zone"`
    HasVentilation bool    `json:"has_ventilation"`
    MinArea        float64 `json:"min_area"`
}

// FurnitureElement элемент мебели.
type FurnitureElement struct {
    ID            string           `json:"id"`
    Type          string           `json:"type"` // always "furniture"
    Name          string           `json:"name"`
    FurnitureType string           `json:"furniture_type"` // sofa, bed, table, chair, wardrobe, kitchen_set
    Position      Point3D          `json:"position"`
    Rotation      Point3D          `json:"rotation"`
    Dimensions    Dimensions       `json:"dimensions"`
    ModelURL      string           `json:"model_url,omitempty"`
    Metadata      FurnitureMetadata `json:"metadata"`
}

// Dimensions размеры объекта.
type Dimensions struct {
    Width  float64 `json:"width"`
    Height float64 `json:"height"`
    Depth  float64 `json:"depth"`
}

// FurnitureMetadata метаданные мебели.
type FurnitureMetadata struct {
    Category string `json:"category"`
    Color    string `json:"color,omitempty"`
}

// UtilityElement инженерный элемент.
type UtilityElement struct {
    ID          string            `json:"id"`
    Type        string            `json:"type"` // always "utility"
    Name        string            `json:"name"`
    UtilityType string            `json:"utility_type"` // water_riser, heating_riser, gas_riser, ventilation, electrical
    Position    Point3D           `json:"position"`
    Properties  UtilityProperties `json:"properties"`
}

// UtilityProperties свойства инженерного элемента.
type UtilityProperties struct {
    CanRelocate    bool    `json:"can_relocate"`
    ProtectionZone float64 `json:"protection_zone"`
}

// ElementMetadata общие метаданные элемента.
type ElementMetadata struct {
    Locked   bool `json:"locked"`
    Visible  bool `json:"visible"`
    Selected bool `json:"selected"`
}

// DisplaySettings настройки отображения.
type DisplaySettings struct {
    FloorTexture  string  `json:"floor_texture"`
    WallColor     string  `json:"wall_color"`
    CeilingColor  string  `json:"ceiling_color"`
    AmbientLight  float64 `json:"ambient_light"`
    ShowGrid      bool    `json:"show_grid"`
    GridSize      float64 `json:"grid_size"`
}
```

## WebSocket: Real-time обновления

```javascript
// Подписка на изменения сцены
ws.send(JSON.stringify({
  type: 'subscribe',
  channel: 'scene:sc_770e8400',
  token: 'access_token'
}));

// Получение обновлений
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  
  switch (data.type) {
    case 'scene:element:added':
      // Добавлен новый элемент
      break;
    case 'scene:element:updated':
      // Элемент обновлён
      break;
    case 'scene:element:removed':
      // Элемент удалён
      break;
    case 'scene:compliance':
      // Обновлён результат проверки
      break;
  }
};

// Отправка изменений (для collaborative editing)
ws.send(JSON.stringify({
  type: 'scene:operation',
  scene_id: 'sc_770e8400',
  operation: {
    op: 'update',
    type: 'furniture',
    id: 'furn_001',
    changes: {
      position: { x: 3.0, y: 0, z: 2.0 }
    }
  }
}));
```


# API планировок

## Обзор

Планировка (Floor Plan) — исходный документ с чертежом квартиры. API поддерживает загрузку различных форматов и распознавание через AI.

## Endpoints

### POST /api/v1/workspaces/:workspaceId/floor-plans

Загрузка планировки.

**Request:**

```http
POST /api/v1/workspaces/ws_550e8400/floor-plans
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: multipart/form-data

--boundary
Content-Disposition: form-data; name="file"; filename="plan.pdf"
Content-Type: application/pdf

<binary file data>
--boundary
Content-Disposition: form-data; name="source_type"

bti
--boundary--
```

**Form Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `file` | file | Yes | Файл планировки |
| `source_type` | string | Yes | Тип источника |
| `name` | string | No | Название (по умолчанию из имени файла) |

**Supported Formats:**

| Format | MIME Type | Max Size |
|--------|-----------|----------|
| PDF | application/pdf | 50 MB |
| JPEG | image/jpeg | 20 MB |
| PNG | image/png | 20 MB |
| TIFF | image/tiff | 50 MB |
| WEBP | image/webp | 20 MB |

**Source Types:**

| Type | Description |
|------|-------------|
| `bti` | Техпаспорт БТИ |
| `technical_plan` | Технический план |
| `sketch` | Рукописный эскиз |
| `other` | Другое |

**Response 201:**

```json
{
  "data": {
    "id": "fp_660e8400-e29b-41d4-a716-446655440001",
    "workspace_id": "ws_550e8400",
    "name": "plan.pdf",
    "file_path": "floor-plans/ws_550e8400/fp_660e8400/original.pdf",
    "file_type": "application/pdf",
    "file_size": 2457600,
    "source_type": "bti",
    "status": "pending",
    "recognition_data": null,
    "error_message": null,
    "preview_url": null,
    "created_at": "2024-01-21T12:00:00Z",
    "updated_at": "2024-01-21T12:00:00Z"
  },
  "request_id": "req_abc123"
}
```

**Note:** После загрузки запускается асинхронное распознавание. Статус можно отслеживать через GET или WebSocket.

---

### GET /api/v1/workspaces/:workspaceId/floor-plans

Список планировок воркспейса.

**Request:**

```http
GET /api/v1/workspaces/ws_550e8400/floor-plans
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "floor_plans": [
      {
        "id": "fp_660e8400-e29b-41d4-a716-446655440001",
        "name": "Техпаспорт БТИ",
        "file_type": "application/pdf",
        "file_size": 2457600,
        "source_type": "bti",
        "status": "completed",
        "preview_url": "https://storage.granula.ru/floor-plans/ws_550e8400/fp_660e8400/preview.png",
        "scenes_count": 2,
        "created_at": "2024-01-21T12:00:00Z"
      }
    ],
    "total": 1
  },
  "request_id": "req_abc123"
}
```

---

### GET /api/v1/workspaces/:workspaceId/floor-plans/:floorPlanId

Получение планировки с данными распознавания.

**Request:**

```http
GET /api/v1/workspaces/ws_550e8400/floor-plans/fp_660e8400
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "id": "fp_660e8400-e29b-41d4-a716-446655440001",
    "workspace_id": "ws_550e8400",
    "name": "Техпаспорт БТИ",
    "file_path": "floor-plans/ws_550e8400/fp_660e8400/original.pdf",
    "file_type": "application/pdf",
    "file_size": 2457600,
    "source_type": "bti",
    "status": "completed",
    "preview_url": "https://storage.granula.ru/floor-plans/ws_550e8400/fp_660e8400/preview.png",
    "download_url": "https://storage.granula.ru/floor-plans/ws_550e8400/fp_660e8400/original.pdf?token=xxx",
    "recognition_data": {
      "bounds": {
        "width": 12.5,
        "height": 2.7,
        "depth": 8.3
      },
      "walls": [
        {
          "id": "wall_001",
          "start": { "x": 0, "y": 0, "z": 0 },
          "end": { "x": 5.0, "y": 0, "z": 0 },
          "thickness": 0.2,
          "is_load_bearing": true,
          "confidence": 0.95
        }
      ],
      "rooms": [
        {
          "id": "room_001",
          "type": "kitchen",
          "name": "Кухня",
          "polygon": [
            { "x": 0, "z": 0 },
            { "x": 4.0, "z": 0 },
            { "x": 4.0, "z": 3.5 },
            { "x": 0, "z": 3.5 }
          ],
          "area": 14.0,
          "confidence": 0.92
        }
      ],
      "openings": [
        {
          "id": "opening_001",
          "type": "door",
          "wall_id": "wall_001",
          "position": 2.0,
          "width": 0.9,
          "height": 2.1,
          "confidence": 0.88
        }
      ],
      "utilities": [
        {
          "id": "utility_001",
          "type": "water_riser",
          "position": { "x": 0.5, "z": 2.0 },
          "confidence": 0.85
        }
      ],
      "metadata": {
        "scale": 100,
        "rotation": 0,
        "recognized_at": "2024-01-21T12:05:00Z",
        "model": "gpt-4-vision",
        "processing_time_ms": 15420
      }
    },
    "created_at": "2024-01-21T12:00:00Z",
    "updated_at": "2024-01-21T12:05:00Z"
  },
  "request_id": "req_abc123"
}
```

---

### PATCH /api/v1/workspaces/:workspaceId/floor-plans/:floorPlanId

Обновление планировки (корректировка распознавания).

**Request:**

```http
PATCH /api/v1/workspaces/ws_550e8400/floor-plans/fp_660e8400
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "name": "Техпаспорт БТИ (исправлено)",
  "recognition_data": {
    "walls": [
      {
        "id": "wall_001",
        "is_load_bearing": false
      }
    ]
  }
}
```

**Response 200:**

```json
{
  "data": {
    "id": "fp_660e8400-e29b-41d4-a716-446655440001",
    "name": "Техпаспорт БТИ (исправлено)",
    "status": "completed",
    "updated_at": "2024-01-21T12:30:00Z"
  },
  "request_id": "req_abc123"
}
```

---

### POST /api/v1/workspaces/:workspaceId/floor-plans/:floorPlanId/reprocess

Повторное распознавание планировки.

**Request:**

```http
POST /api/v1/workspaces/ws_550e8400/floor-plans/fp_660e8400/reprocess
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "hints": {
    "scale": 50,
    "load_bearing_walls": ["wall_001", "wall_003"]
  }
}
```

**Response 202:**

```json
{
  "data": {
    "id": "fp_660e8400-e29b-41d4-a716-446655440001",
    "status": "processing",
    "message": "Reprocessing started"
  },
  "request_id": "req_abc123"
}
```

---

### DELETE /api/v1/workspaces/:workspaceId/floor-plans/:floorPlanId

Удаление планировки.

**Request:**

```http
DELETE /api/v1/workspaces/ws_550e8400/floor-plans/fp_660e8400
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "message": "Floor plan deleted"
  },
  "request_id": "req_abc123"
}
```

**Note:** Удаление планировки не удаляет созданные на её основе сцены.

---

### POST /api/v1/workspaces/:workspaceId/floor-plans/:floorPlanId/create-scene

Создание 3D сцены из планировки.

**Request:**

```http
POST /api/v1/workspaces/ws_550e8400/floor-plans/fp_660e8400/create-scene
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "name": "Основная планировка",
  "description": "Исходная планировка из БТИ"
}
```

**Response 201:**

```json
{
  "data": {
    "scene_id": "sc_770e8400-e29b-41d4-a716-446655440002",
    "name": "Основная планировка",
    "status": "created",
    "message": "Scene created from floor plan"
  },
  "request_id": "req_abc123"
}
```

---

## Recognition Data Schema

```go
// internal/dto/floor_plan.go

// RecognitionData результат распознавания планировки.
type RecognitionData struct {
    // Габариты помещения в метрах
    Bounds *BoundsData `json:"bounds"`
    
    // Распознанные стены
    Walls []WallData `json:"walls"`
    
    // Распознанные комнаты
    Rooms []RoomData `json:"rooms"`
    
    // Проёмы (двери, окна)
    Openings []OpeningData `json:"openings"`
    
    // Инженерные элементы
    Utilities []UtilityData `json:"utilities"`
    
    // Метаданные распознавания
    Metadata *RecognitionMetadata `json:"metadata"`
}

// BoundsData габариты.
type BoundsData struct {
    Width  float64 `json:"width"`   // Ширина по X
    Height float64 `json:"height"`  // Высота потолков
    Depth  float64 `json:"depth"`   // Глубина по Z
}

// WallData данные стены.
type WallData struct {
    ID            string    `json:"id"`
    Start         Point3D   `json:"start"`
    End           Point3D   `json:"end"`
    Thickness     float64   `json:"thickness"`
    IsLoadBearing bool      `json:"is_load_bearing"`
    Material      string    `json:"material,omitempty"` // brick, concrete, drywall
    Confidence    float64   `json:"confidence"`         // 0-1
}

// RoomData данные комнаты.
type RoomData struct {
    ID         string    `json:"id"`
    Type       string    `json:"type"`       // kitchen, bedroom, bathroom, etc.
    Name       string    `json:"name"`
    Polygon    []Point2D `json:"polygon"`    // Контур комнаты
    Area       float64   `json:"area"`       // Площадь в м²
    Confidence float64   `json:"confidence"`
}

// OpeningData данные проёма.
type OpeningData struct {
    ID         string  `json:"id"`
    Type       string  `json:"type"`     // door, window
    WallID     string  `json:"wall_id"`
    Position   float64 `json:"position"` // От начала стены
    Width      float64 `json:"width"`
    Height     float64 `json:"height"`
    Elevation  float64 `json:"elevation,omitempty"` // Высота от пола (для окон)
    Confidence float64 `json:"confidence"`
}

// UtilityData данные инженерного элемента.
type UtilityData struct {
    ID         string  `json:"id"`
    Type       string  `json:"type"` // water_riser, heating_riser, gas_riser, ventilation
    Position   Point3D `json:"position"`
    CanRelocate bool   `json:"can_relocate"`
    Confidence float64 `json:"confidence"`
}

// Point2D точка на плоскости.
type Point2D struct {
    X float64 `json:"x"`
    Z float64 `json:"z"`
}

// Point3D точка в пространстве.
type Point3D struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
    Z float64 `json:"z"`
}

// RecognitionMetadata метаданные распознавания.
type RecognitionMetadata struct {
    Scale           int       `json:"scale"`            // Масштаб чертежа
    Rotation        int       `json:"rotation"`         // Поворот в градусах
    RecognizedAt    time.Time `json:"recognized_at"`
    Model           string    `json:"model"`            // Модель AI
    ProcessingTimeMs int64    `json:"processing_time_ms"`
}
```

## Статусы распознавания

| Status | Description |
|--------|-------------|
| `pending` | Ожидает обработки |
| `processing` | В процессе распознавания |
| `completed` | Успешно распознано |
| `failed` | Ошибка распознавания |

## WebSocket: Статус распознавания

Подписка на обновления статуса:

```javascript
const ws = new WebSocket('wss://api.granula.ru/ws');

ws.send(JSON.stringify({
  type: 'subscribe',
  channel: 'floor_plan:fp_660e8400',
  token: 'access_token'
}));

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  // {
  //   "type": "floor_plan:status",
  //   "data": {
  //     "id": "fp_660e8400",
  //     "status": "completed",
  //     "progress": 100,
  //     "recognition_data": { ... }
  //   }
  // }
};
```


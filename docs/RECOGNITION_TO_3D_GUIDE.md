# 🎯 Recognition → 3D Model: Полный Flow

> Руководство по получению JSON для 3D модели из AI распознавания планировки

---

## 📋 Содержание

1. [Общая схема](#общая-схема)
2. [Шаг 1: Загрузка плана](#шаг-1-загрузка-плана)
3. [Шаг 2: Запуск распознавания](#шаг-2-запуск-распознавания)
4. [Шаг 3: Polling статуса](#шаг-3-polling-статуса)
5. [Шаг 4: Создание 3D сцены](#шаг-4-создание-3d-сцены)
6. [Формат JSON Recognition](#формат-json-recognition)
7. [Формат JSON Scene Elements](#формат-json-scene-elements)
8. [Различия форматов](#различия-форматов)
9. [Полный пример кода](#полный-пример-кода)

---

## 🔄 Общая схема

```
┌─────────────────────────────────────────────────────────────────────┐
│ 1. ЗАГРУЗКА ПЛАНА                                                    │
│    POST /floor-plans (multipart/form-data)                          │
│    → Получаем floor_plan_id                                          │
└──────────────────────────────┬──────────────────────────────────────┘
                               ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 2. AI РАСПОЗНАВАНИЕ                                                  │
│    POST /ai/recognize { floor_plan_id, image_base64 }               │
│    → Получаем job_id                                                 │
└──────────────────────────────┬──────────────────────────────────────┘
                               ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 3. POLLING СТАТУСА (каждые 2-3 сек)                                  │
│    GET /ai/recognize/{job_id}/status                                │
│    → Когда status="completed", получаем result с JSON               │
│    ┌──────────────────────────────────────────────────────────┐     │
│    │  result: {                                                │     │
│    │    walls: [...],  rooms: [...],  openings: [...],        │     │
│    │    utilities: [...],  equipment: [...]                    │     │
│    │  }                                                        │     │
│    └──────────────────────────────────────────────────────────┘     │
└──────────────────────────────┬──────────────────────────────────────┘
                               ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 4. СОЗДАНИЕ 3D СЦЕНЫ ИЗ РЕЗУЛЬТАТА                                   │
│    POST /workspaces/{id}/scenes { floor_plan_id: "..." }            │
│    → Сцена создаётся с элементами из recognition                     │
│                                                                      │
│    ИЛИ вручную:                                                     │
│    PUT /workspaces/{id}/scenes/{scene_id}/elements                  │
│    → Передаём elements из recognition result                         │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 📤 Шаг 1: Загрузка плана

```http
POST /api/v1/floor-plans
Authorization: Bearer <token>
Content-Type: multipart/form-data

file: <binary image>
workspace_id: "workspace-uuid"
name: "План из БТИ"
```

**Поддерживаемые форматы:** JPEG, PNG, PDF

**Ответ:**
```json
{
  "data": {
    "id": "floor-plan-uuid",
    "workspace_id": "workspace-uuid",
    "name": "План из БТИ",
    "status": "uploaded",
    "file_url": "https://storage.granula.ru/...",
    "created_at": "2024-01-15T10:35:00Z"
  }
}
```

---

## 🤖 Шаг 2: Запуск распознавания

```http
POST /api/v1/ai/recognize
Authorization: Bearer <token>
Content-Type: application/json

{
  "floor_plan_id": "floor-plan-uuid",
  "image_base64": "data:image/jpeg;base64,/9j/4AAQSkZJRg...",
  "image_type": "image/jpeg",
  "options": {
    "detect_load_bearing": true,
    "detect_wet_zones": true,
    "detect_furniture": false
  }
}
```

**Ответ:**
```json
{
  "data": {
    "job_id": "recognition-job-uuid",
    "status": "processing"
  }
}
```

---

## 🔄 Шаг 3: Polling статуса

```http
GET /api/v1/ai/recognize/{job_id}/status
Authorization: Bearer <token>
```

**Ответ (в процессе):**
```json
{
  "data": {
    "job_id": "recognition-job-uuid",
    "status": "processing",
    "progress": 45
  }
}
```

**Ответ (завершено):**
```json
{
  "data": {
    "job_id": "recognition-job-uuid",
    "status": "completed",
    "progress": 100,
    "result": {
      // ← ВОТ ЗДЕСЬ JSON ДЛЯ 3D МОДЕЛИ!
      "dimensions": {...},
      "walls": [...],
      "rooms": [...],
      "openings": [...],
      "utilities": [...],
      "equipment": [...]
    }
  }
}
```

**Ответ (ошибка):**
```json
{
  "data": {
    "job_id": "recognition-job-uuid",
    "status": "failed",
    "error": "Could not recognize floor plan"
  }
}
```

---

## 🎮 Шаг 4: Создание 3D сцены

### Вариант A: Автоматически из floor_plan

```http
POST /api/v1/workspaces/{workspace_id}/scenes
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Планировка из БТИ",
  "description": "Распознанная планировка",
  "floor_plan_id": "floor-plan-uuid"
}
```

Сервер автоматически подтянет результат recognition и заполнит elements.

### Вариант B: Вручную передать elements

```http
PUT /api/v1/workspaces/{workspace_id}/scenes/{scene_id}/elements
Authorization: Bearer <token>
Content-Type: application/json

{
  "elements": {
    "walls": [...],
    "rooms": [...],
    "furniture": [],
    "utilities": [...]
  }
}
```

---

## 📦 Формат JSON Recognition

Этот JSON возвращается в поле `result` при статусе `completed`:

```json
{
  "dimensions": {
    "width": 12.5,
    "height": 8.3
  },
  "total_area": 65.5,
  "detected_scale": "1:100",
  
  "walls": [
    {
      "temp_id": "wall_001",
      "start": { "x": 0.0, "y": 0.0 },
      "end": { "x": 5.0, "y": 0.0 },
      "thickness": 0.2,
      "is_load_bearing": true,
      "material": "brick",
      "confidence": 0.95
    },
    {
      "temp_id": "wall_002",
      "start": { "x": 5.0, "y": 0.0 },
      "end": { "x": 5.0, "y": 3.5 },
      "thickness": 0.12,
      "is_load_bearing": false,
      "material": "drywall",
      "confidence": 0.88
    }
  ],
  
  "rooms": [
    {
      "temp_id": "room_001",
      "type": "KITCHEN",
      "name": "Кухня",
      "boundary": [
        { "x": 0, "y": 0 },
        { "x": 4.0, "y": 0 },
        { "x": 4.0, "y": 3.5 },
        { "x": 0, "y": 3.5 }
      ],
      "area": 14.0,
      "is_wet_zone": true,
      "has_window": true,
      "wall_ids": ["wall_001", "wall_002", "wall_003", "wall_004"],
      "confidence": 0.92
    },
    {
      "temp_id": "room_002",
      "type": "LIVING",
      "name": "Гостиная",
      "boundary": [
        { "x": 4.0, "y": 0 },
        { "x": 10.0, "y": 0 },
        { "x": 10.0, "y": 5.0 },
        { "x": 4.0, "y": 5.0 }
      ],
      "area": 30.0,
      "is_wet_zone": false,
      "has_window": true,
      "wall_ids": ["wall_005", "wall_006", "wall_007", "wall_008"],
      "confidence": 0.95
    }
  ],
  
  "openings": [
    {
      "temp_id": "opening_001",
      "type": "door",
      "subtype": "межкомнатная",
      "position": { "x": 2.0, "y": 0.0 },
      "width": 0.9,
      "height": 2.1,
      "wall_id": "wall_001",
      "opens_to": "left",
      "confidence": 0.85
    },
    {
      "temp_id": "opening_002",
      "type": "window",
      "subtype": null,
      "position": { "x": 1.5, "y": 3.5 },
      "width": 1.5,
      "height": 1.4,
      "wall_id": "wall_004",
      "opens_to": null,
      "confidence": 0.90
    }
  ],
  
  "utilities": [
    {
      "temp_id": "utility_001",
      "type": "water_riser",
      "position": { "x": 0.5, "y": 2.0 },
      "can_relocate": false,
      "protection_zone": 0.3,
      "room_id": "room_001",
      "confidence": 0.80
    },
    {
      "temp_id": "utility_002",
      "type": "ventilation",
      "position": { "x": 3.5, "y": 3.2 },
      "can_relocate": false,
      "protection_zone": 0.1,
      "room_id": "room_001",
      "confidence": 0.75
    }
  ],
  
  "equipment": [
    {
      "temp_id": "equip_001",
      "type": "кухонная_плита",
      "position": { "x": 1.0, "y": 3.0 },
      "dimensions": { "width": 0.6, "depth": 0.6 },
      "room_id": "room_001",
      "confidence": 0.75
    },
    {
      "temp_id": "equip_002",
      "type": "раковина",
      "position": { "x": 2.5, "y": 3.3 },
      "dimensions": { "width": 0.8, "depth": 0.5 },
      "room_id": "room_001",
      "confidence": 0.82
    }
  ],
  
  "metadata": {
    "source_type": "BTI",
    "quality": "high",
    "orientation": 0,
    "has_dimensions": true,
    "has_annotations": true
  },
  
  "warnings": [
    "Масштаб определён по размеру двери (0.9м)"
  ],
  "notes": [
    "Обнаружено 4 комнаты, 12 стен, 5 проёмов"
  ]
}
```

### Типы комнат (room.type)

| Код | Название | Мокрая зона |
|-----|----------|-------------|
| `LIVING` | Гостиная | ❌ |
| `BEDROOM` | Спальня | ❌ |
| `CHILDREN` | Детская | ❌ |
| `OFFICE` | Кабинет | ❌ |
| `KITCHEN` | Кухня | ✅ |
| `KITCHEN_LIVING` | Кухня-гостиная | ✅ |
| `BATHROOM` | Ванная | ✅ |
| `TOILET` | Туалет | ✅ |
| `COMBINED_BATHROOM` | Совмещённый санузел | ✅ |
| `HALLWAY` | Коридор/прихожая | ❌ |
| `STORAGE` | Кладовая | ❌ |
| `LAUNDRY` | Постирочная | ✅ |
| `BALCONY` | Балкон | ❌ |
| `LOGGIA` | Лоджия | ❌ |

### Типы материалов стен (wall.material)

| Код | Описание | Несущая? |
|-----|----------|----------|
| `brick` | Кирпичная кладка | Обычно да |
| `concrete` | Бетон монолитный | Да |
| `drywall` | Гипсокартон | Нет |
| `glass` | Стекло | Нет |
| `unknown` | Не определено | Проверить |

### Типы инженерных элементов (utility.type)

| Код | Описание | Можно перенести? |
|-----|----------|------------------|
| `water_riser` | Стояк водоснабжения | ❌ |
| `sewer_riser` | Стояк канализации | ❌ |
| `heating_riser` | Стояк отопления | ❌ |
| `ventilation` | Вентиляционный канал | ❌ |
| `electrical_panel` | Электрощит | С согласованием |

---

## 🎮 Формат JSON Scene Elements

После создания сцены, данные представлены в формате Scene Elements:

```json
{
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
    
    "furniture": [
      {
        "id": "furn_001",
        "type": "furniture",
        "name": "Диван",
        "furniture_type": "sofa",
        "position": { "x": 6.0, "y": 0, "z": 2.0 },
        "rotation": { "x": 0, "y": 90, "z": 0 },
        "dimensions": {
          "width": 2.0,
          "height": 0.85,
          "depth": 0.9
        },
        "metadata": {
          "category": "living",
          "color": "#8B4513"
        }
      }
    ],
    
    "utilities": [
      {
        "id": "utility_001",
        "type": "utility",
        "name": "Стояк водоснабжения",
        "utility_type": "water_riser",
        "position": { "x": 0.5, "y": 0, "z": 2.0 },
        "properties": {
          "can_relocate": false,
          "protection_zone": 0.3
        }
      }
    ]
  }
}
```

---

## 🔄 Различия форматов

| Recognition Result | Scene Elements |
|--------------------|----------------|
| `temp_id` | `id` |
| `boundary` (2D массив {x, y}) | `polygon` (2D массив {x, z}) |
| `start/end` — 2D (x, y) | `start/end` — 3D (x, y, z) |
| Есть `confidence` поля | Нет confidence |
| `wall_ids` в rooms | Нет связи rooms→walls |
| Плоский список `openings` | `openings` вложены в `walls` |
| `is_load_bearing` в wall | `properties.is_load_bearing` |
| Нет `height` у стен | Есть `height` у стен |
| Нет `metadata` у элементов | Есть `metadata` (locked, visible, selected) |

### Конвертация координат

**Recognition (2D план):**
```
Y ↑
  │
  └──→ X
```

**Scene (3D пространство):**
```
    Y (высота)
    ↑
    │
    └──→ X
   ╱
  Z (глубина)
```

**Правило конвертации:**
- Recognition `x` → Scene `x`
- Recognition `y` → Scene `z`
- Scene `y` = 0 (уровень пола) или высота элемента

---

## 💻 Полный пример кода

```javascript
// config
const API_BASE = 'https://api.granula.raitokyokai.tech/api/v1';

// Вспомогательная функция
async function apiRequest(endpoint, options = {}) {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers: {
      'Authorization': `Bearer ${localStorage.getItem('token')}`,
      ...options.headers,
    },
  });
  
  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message || 'API Error');
  }
  
  return response.json();
}

// =============================================================================
// MAIN FLOW
// =============================================================================

async function recognizeFloorPlan(workspaceId, imageFile) {
  console.log('📤 Step 1: Uploading floor plan...');
  
  // 1. Upload floor plan
  const formData = new FormData();
  formData.append('file', imageFile);
  formData.append('workspace_id', workspaceId);
  formData.append('name', imageFile.name);
  
  const uploadResponse = await fetch(`${API_BASE}/floor-plans`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${localStorage.getItem('token')}`,
    },
    body: formData,
  });
  
  const { data: floorPlan } = await uploadResponse.json();
  console.log('✅ Floor plan uploaded:', floorPlan.id);
  
  // 2. Convert image to base64
  console.log('🔄 Step 2: Converting image to base64...');
  const base64Image = await fileToBase64(imageFile);
  
  // 3. Start recognition
  console.log('🤖 Step 3: Starting AI recognition...');
  const { data: recognitionJob } = await apiRequest('/ai/recognize', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      floor_plan_id: floorPlan.id,
      image_base64: base64Image,
      image_type: imageFile.type,
      options: {
        detect_load_bearing: true,
        detect_wet_zones: true,
        detect_furniture: true,
      },
    }),
  });
  
  console.log('⏳ Recognition job started:', recognitionJob.job_id);
  
  // 4. Poll for completion
  console.log('🔄 Step 4: Polling for status...');
  const recognitionResult = await pollRecognitionStatus(recognitionJob.job_id);
  console.log('✅ Recognition completed!');
  console.log('📊 Found:', {
    walls: recognitionResult.walls?.length || 0,
    rooms: recognitionResult.rooms?.length || 0,
    openings: recognitionResult.openings?.length || 0,
  });
  
  // 5. Create 3D scene
  console.log('🎮 Step 5: Creating 3D scene...');
  const { data: scene } = await apiRequest(`/workspaces/${workspaceId}/scenes`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      name: `Планировка: ${imageFile.name}`,
      description: 'Создано из распознанного плана',
      floor_plan_id: floorPlan.id,
    }),
  });
  
  console.log('✅ Scene created:', scene.id);
  
  return {
    floorPlan,
    recognitionResult,
    scene,
  };
}

// =============================================================================
// HELPERS
// =============================================================================

async function fileToBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result);
    reader.onerror = reject;
    reader.readAsDataURL(file);
  });
}

async function pollRecognitionStatus(jobId, maxAttempts = 60, intervalMs = 2000) {
  for (let attempt = 0; attempt < maxAttempts; attempt++) {
    const { data } = await apiRequest(`/ai/recognize/${jobId}/status`);
    
    console.log(`  Progress: ${data.progress || 0}% (attempt ${attempt + 1})`);
    
    if (data.status === 'completed') {
      return data.result;
    }
    
    if (data.status === 'failed') {
      throw new Error(`Recognition failed: ${data.error}`);
    }
    
    // Wait before next poll
    await new Promise(resolve => setTimeout(resolve, intervalMs));
  }
  
  throw new Error('Recognition timeout');
}

// =============================================================================
// USAGE EXAMPLE
// =============================================================================

// HTML: <input type="file" id="floorPlanInput" accept="image/*">

document.getElementById('floorPlanInput').addEventListener('change', async (e) => {
  const file = e.target.files[0];
  if (!file) return;
  
  try {
    const workspaceId = 'your-workspace-id'; // Get from context
    
    const result = await recognizeFloorPlan(workspaceId, file);
    
    // Use result.scene.elements for Three.js rendering
    renderScene(result.scene.elements);
    
  } catch (error) {
    console.error('Error:', error);
    alert(`Ошибка: ${error.message}`);
  }
});

// =============================================================================
// THREE.JS INTEGRATION EXAMPLE
// =============================================================================

function renderScene(elements) {
  // elements.walls — массив стен
  elements.walls.forEach(wall => {
    // Создаём геометрию стены
    const length = Math.sqrt(
      Math.pow(wall.end.x - wall.start.x, 2) +
      Math.pow(wall.end.z - wall.start.z, 2)
    );
    
    const geometry = new THREE.BoxGeometry(
      length,           // длина
      wall.height,      // высота
      wall.thickness    // толщина
    );
    
    const material = new THREE.MeshStandardMaterial({
      color: wall.properties.is_load_bearing ? 0x8B4513 : 0xCCCCCC,
    });
    
    const mesh = new THREE.Mesh(geometry, material);
    
    // Позиционируем
    mesh.position.set(
      (wall.start.x + wall.end.x) / 2,
      wall.height / 2,
      (wall.start.z + wall.end.z) / 2
    );
    
    // Поворачиваем
    const angle = Math.atan2(
      wall.end.z - wall.start.z,
      wall.end.x - wall.start.x
    );
    mesh.rotation.y = -angle;
    
    scene.add(mesh);
  });
  
  // elements.rooms — для отрисовки полов
  elements.rooms.forEach(room => {
    const shape = new THREE.Shape();
    room.polygon.forEach((point, i) => {
      if (i === 0) {
        shape.moveTo(point.x, point.z);
      } else {
        shape.lineTo(point.x, point.z);
      }
    });
    shape.closePath();
    
    const geometry = new THREE.ShapeGeometry(shape);
    geometry.rotateX(-Math.PI / 2); // Поворачиваем в горизонтальную плоскость
    
    const material = new THREE.MeshStandardMaterial({
      color: room.properties.has_wet_zone ? 0x4169E1 : 0xDEB887,
      side: THREE.DoubleSide,
    });
    
    const floor = new THREE.Mesh(geometry, material);
    floor.position.y = 0.01; // Чуть выше нуля чтобы не z-fighting
    
    scene.add(floor);
  });
}
```

---

## 📊 Где какие данные брать

| Данные | Endpoint | Поле | Когда использовать |
|--------|----------|------|-------------------|
| Сырой JSON распознавания | `GET /ai/recognize/{job_id}/status` | `result` | Для отладки, кастомной обработки |
| Готовые 3D элементы | `GET /scenes/{scene_id}` | `elements` | Для рендеринга в Three.js |
| Обновлённые элементы | `PATCH /scenes/{id}/elements` | response | После редактирования пользователем |
| Сгенерированные варианты | `GET /ai/generate/{job_id}/status` | `variants` | При AI-генерации перепланировок |

---

## ⚠️ Важные замечания

1. **Confidence** — уверенность AI (0.0-1.0). Элементы с `confidence < 0.7` стоит подсветить для ручной проверки.

2. **is_load_bearing** — критически важно! Несущие стены нельзя сносить. Отображай их другим цветом.

3. **can_relocate** в utilities — стояки и вентканалы переносить запрещено. Блокируй их перемещение в редакторе.

4. **protection_zone** — радиус вокруг инженерных элементов, где нельзя строить.

5. **Координаты в метрах** — все размеры в реальных метрах с точностью до 0.01.

---

*Документация актуальна на: 30 ноября 2024*


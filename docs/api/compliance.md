# API проверки норм (Compliance)

## Обзор

Система проверки соответствия строительным нормам (СНиП, Жилищный кодекс РФ). Работает в реальном времени при изменениях сцены.

## Категории правил

| Category | Description |
|----------|-------------|
| `structural` | Несущие конструкции |
| `plumbing` | Сантехника и мокрые зоны |
| `electrical` | Электрика |
| `ventilation` | Вентиляция |
| `fire_safety` | Пожарная безопасность |
| `accessibility` | Доступность |
| `general` | Общие требования |

## Уровни критичности

| Severity | Description | Action |
|----------|-------------|--------|
| `error` | Критическое нарушение | Блокирует сохранение |
| `warning` | Предупреждение | Рекомендация исправить |
| `info` | Информация | Справочно |

## Endpoints

### POST /api/v1/compliance/check

Проверка элементов на соответствие нормам.

**Request:**

```http
POST /api/v1/compliance/check
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "scene_id": "sc_770e8400",
  "branch_id": "br_880e8400",
  "elements": {
    "walls": [...],
    "rooms": [...],
    "utilities": [...]
  },
  "check_types": ["structural", "plumbing", "fire_safety"]
}
```

**Request Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `scene_id` | string | Yes | ID сцены |
| `branch_id` | string | No | ID ветки |
| `elements` | object | No | Элементы для проверки (если не указано — вся сцена) |
| `check_types` | []string | No | Категории для проверки (если не указано — все) |

**Response 200:**

```json
{
  "data": {
    "scene_id": "sc_770e8400",
    "branch_id": "br_880e8400",
    "checked_at": "2024-01-21T12:00:00Z",
    "is_compliant": false,
    "summary": {
      "total_checks": 45,
      "passed": 42,
      "violations": 2,
      "warnings": 1,
      "info": 0
    },
    "violations": [
      {
        "id": "viol_001",
        "rule_code": "SNIP_2.08.01-89_4.1",
        "rule_name": "Минимальная площадь кухни",
        "category": "general",
        "severity": "error",
        "message": "Площадь кухни (4.8 м²) меньше минимально допустимой (5 м²)",
        "affected_elements": [
          {
            "type": "room",
            "id": "room_kitchen",
            "name": "Кухня"
          }
        ],
        "suggestion": "Увеличьте площадь кухни минимум до 5 м²",
        "source": "СНиП 2.08.01-89 п.4.1",
        "source_url": "https://docs.cntd.ru/document/5200094"
      },
      {
        "id": "viol_002",
        "rule_code": "SNIP_LOAD_BEARING_001",
        "rule_name": "Запрет сноса несущих стен",
        "category": "structural",
        "severity": "error",
        "message": "Несущая стена не может быть удалена или существенно изменена",
        "affected_elements": [
          {
            "type": "wall",
            "id": "wall_003",
            "name": "Несущая стена 2"
          }
        ],
        "suggestion": "Восстановите несущую стену или обратитесь к инженеру для проектирования усиления",
        "source": "ЖК РФ ст.26",
        "source_url": "http://www.consultant.ru/document/cons_doc_LAW_51057/..."
      }
    ],
    "warnings": [
      {
        "id": "warn_001",
        "rule_code": "SNIP_2.08.01-89_4.3",
        "rule_name": "Рекомендуемая ширина коридора",
        "category": "accessibility",
        "severity": "warning",
        "message": "Ширина коридора (0.95 м) меньше рекомендуемой (1.1 м)",
        "affected_elements": [
          {
            "type": "room",
            "id": "room_corridor",
            "name": "Коридор"
          }
        ],
        "suggestion": "Рассмотрите возможность расширения коридора для удобства перемещения",
        "source": "СНиП 2.08.01-89 п.4.3"
      }
    ],
    "info": []
  },
  "request_id": "req_abc123"
}
```

---

### POST /api/v1/compliance/check-operation

Проверка одной операции до её применения.

**Request:**

```http
POST /api/v1/compliance/check-operation
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
Content-Type: application/json

{
  "scene_id": "sc_770e8400",
  "branch_id": "br_880e8400",
  "operation": {
    "op": "remove",
    "type": "wall",
    "id": "wall_003"
  }
}
```

**Response 200:**

```json
{
  "data": {
    "operation_allowed": false,
    "violations": [
      {
        "rule_code": "SNIP_LOAD_BEARING_001",
        "severity": "error",
        "message": "Нельзя удалить несущую стену",
        "affected_elements": ["wall_003"]
      }
    ],
    "warnings": [],
    "suggestion": "Эта стена является несущей. Для её модификации требуется проект усиления от лицензированной организации."
  },
  "request_id": "req_abc123"
}
```

---

### GET /api/v1/compliance/rules

Справочник правил.

**Request:**

```http
GET /api/v1/compliance/rules?category=structural&active=true
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `category` | string | Фильтр по категории |
| `severity` | string | Фильтр по критичности |
| `active` | bool | Только активные правила |
| `search` | string | Поиск по названию/описанию |

**Response 200:**

```json
{
  "data": {
    "rules": [
      {
        "id": "rule_001",
        "code": "SNIP_LOAD_BEARING_001",
        "name": "Запрет сноса несущих стен",
        "category": "structural",
        "severity": "error",
        "description": "Несущие стены и колонны нельзя демонтировать или существенно изменять без проекта усиления",
        "source": "ЖК РФ ст.26, СП 54.13330.2016",
        "source_url": "https://docs.cntd.ru/document/456054198",
        "active": true
      },
      {
        "id": "rule_002",
        "code": "SNIP_2.08.01-89_4.1",
        "name": "Минимальная площадь кухни",
        "category": "general",
        "severity": "error",
        "description": "Площадь кухни в квартире должна быть не менее 5 м² (для однокомнатных квартир допускается 5 м² кухни-ниши)",
        "source": "СНиП 2.08.01-89 п.4.1",
        "source_url": "https://docs.cntd.ru/document/5200094",
        "active": true
      }
    ],
    "total": 2
  },
  "request_id": "req_abc123"
}
```

---

### GET /api/v1/compliance/rules/:ruleId

Детали правила.

**Request:**

```http
GET /api/v1/compliance/rules/rule_001
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response 200:**

```json
{
  "data": {
    "id": "rule_001",
    "code": "SNIP_LOAD_BEARING_001",
    "name": "Запрет сноса несущих стен",
    "category": "structural",
    "severity": "error",
    "description": "Несущие стены и колонны нельзя демонтировать или существенно изменять без проекта усиления, разработанного лицензированной проектной организацией.",
    "detailed_description": "Несущие конструкции воспринимают нагрузку от вышележащих этажей и кровли. Их повреждение может привести к:\n- Трещинам в конструкциях\n- Провисанию перекрытий\n- Обрушению части здания\n\nДля изменения несущих конструкций требуется:\n1. Техническое заключение о состоянии конструкций\n2. Проект перепланировки с усилением\n3. Согласование в Мосжилинспекции\n4. Авторский надзор при выполнении работ",
    "rule_config": {
      "type": "element_property",
      "params": {
        "element_type": "wall",
        "property": "is_load_bearing",
        "forbidden_operations": ["remove", "resize_significant"]
      }
    },
    "source": "ЖК РФ ст.26, СП 54.13330.2016",
    "source_url": "https://docs.cntd.ru/document/456054198",
    "related_rules": ["SNIP_STRUCTURAL_002", "SNIP_STRUCTURAL_003"],
    "examples": [
      {
        "title": "Запрещено",
        "description": "Полный демонтаж несущей стены",
        "image_url": "https://storage.granula.ru/rules/examples/load_bearing_forbidden.png"
      },
      {
        "title": "Допустимо",
        "description": "Устройство проёма с усилением металлоконструкциями",
        "image_url": "https://storage.granula.ru/rules/examples/load_bearing_allowed.png"
      }
    ],
    "active": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-15T10:00:00Z"
  },
  "request_id": "req_abc123"
}
```

---

### GET /api/v1/scenes/:sceneId/compliance/report

Полный отчёт о соответствии для сцены.

**Request:**

```http
GET /api/v1/scenes/sc_770e8400/compliance/report?branch_id=br_880e8400&format=json
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `branch_id` | string | - | ID ветки |
| `format` | string | json | Формат: json, pdf |

**Response 200 (JSON):**

```json
{
  "data": {
    "report_id": "rep_990e8400",
    "scene_id": "sc_770e8400",
    "branch_id": "br_880e8400",
    "generated_at": "2024-01-21T12:00:00Z",
    "summary": {
      "is_compliant": true,
      "total_checks": 52,
      "by_category": {
        "structural": { "passed": 8, "violations": 0, "warnings": 0 },
        "plumbing": { "passed": 6, "violations": 0, "warnings": 1 },
        "electrical": { "passed": 4, "violations": 0, "warnings": 0 },
        "ventilation": { "passed": 5, "violations": 0, "warnings": 0 },
        "fire_safety": { "passed": 10, "violations": 0, "warnings": 0 },
        "accessibility": { "passed": 7, "violations": 0, "warnings": 2 },
        "general": { "passed": 12, "violations": 0, "warnings": 0 }
      },
      "by_severity": {
        "error": 0,
        "warning": 3,
        "info": 2
      }
    },
    "violations": [],
    "warnings": [...],
    "info": [...],
    "checked_elements": {
      "walls": 12,
      "rooms": 4,
      "utilities": 5
    },
    "recommendations": [
      "Все несущие конструкции сохранены",
      "Мокрые зоны не перенесены над жилыми помещениями",
      "Вентиляционные каналы не затронуты"
    ],
    "legal_status": {
      "can_legalize": true,
      "required_documents": [
        "Проект перепланировки",
        "Техническое заключение",
        "Акт скрытых работ"
      ],
      "estimated_approval_time": "2-3 месяца"
    }
  },
  "request_id": "req_abc123"
}
```

**Response 200 (PDF):**

```http
Content-Type: application/pdf
Content-Disposition: attachment; filename="compliance_report_sc_770e8400.pdf"

<binary PDF data>
```

---

## DTO Types

```go
// internal/dto/compliance.go

// CheckComplianceInput запрос проверки.
type CheckComplianceInput struct {
    SceneID    string         `json:"scene_id" validate:"required"`
    BranchID   *string        `json:"branch_id,omitempty"`
    Elements   *SceneElements `json:"elements,omitempty"`
    CheckTypes []string       `json:"check_types,omitempty"`
}

// CheckOperationInput запрос проверки операции.
type CheckOperationInput struct {
    SceneID   string                `json:"scene_id" validate:"required"`
    BranchID  *string               `json:"branch_id,omitempty"`
    Operation *ElementOperation     `json:"operation" validate:"required"`
}

// ElementOperation операция над элементом.
type ElementOperation struct {
    Op      string                 `json:"op" validate:"required,oneof=add update remove"`
    Type    string                 `json:"type" validate:"required"`
    ID      string                 `json:"id,omitempty"`
    Element map[string]interface{} `json:"element,omitempty"`
    Changes map[string]interface{} `json:"changes,omitempty"`
}

// ComplianceResult результат проверки.
type ComplianceResult struct {
    SceneID     string                   `json:"scene_id"`
    BranchID    *string                  `json:"branch_id,omitempty"`
    CheckedAt   time.Time                `json:"checked_at"`
    IsCompliant bool                     `json:"is_compliant"`
    Summary     *ComplianceSummary       `json:"summary"`
    Violations  []ComplianceViolation    `json:"violations"`
    Warnings    []ComplianceViolation    `json:"warnings"`
    Info        []ComplianceViolation    `json:"info"`
}

// ComplianceSummary сводка проверки.
type ComplianceSummary struct {
    TotalChecks int `json:"total_checks"`
    Passed      int `json:"passed"`
    Violations  int `json:"violations"`
    Warnings    int `json:"warnings"`
    Info        int `json:"info"`
}

// ComplianceViolation нарушение/предупреждение.
type ComplianceViolation struct {
    ID               string            `json:"id"`
    RuleCode         string            `json:"rule_code"`
    RuleName         string            `json:"rule_name"`
    Category         string            `json:"category"`
    Severity         string            `json:"severity"`
    Message          string            `json:"message"`
    AffectedElements []AffectedElement `json:"affected_elements"`
    Suggestion       string            `json:"suggestion,omitempty"`
    Source           string            `json:"source"`
    SourceURL        string            `json:"source_url,omitempty"`
}

// AffectedElement затронутый элемент.
type AffectedElement struct {
    Type string `json:"type"`
    ID   string `json:"id"`
    Name string `json:"name"`
}

// ComplianceRuleResponse правило.
type ComplianceRuleResponse struct {
    ID                  string                 `json:"id"`
    Code                string                 `json:"code"`
    Name                string                 `json:"name"`
    Category            string                 `json:"category"`
    Severity            string                 `json:"severity"`
    Description         string                 `json:"description"`
    DetailedDescription string                 `json:"detailed_description,omitempty"`
    RuleConfig          map[string]interface{} `json:"rule_config,omitempty"`
    Source              string                 `json:"source"`
    SourceURL           string                 `json:"source_url,omitempty"`
    RelatedRules        []string               `json:"related_rules,omitempty"`
    Examples            []RuleExample          `json:"examples,omitempty"`
    Active              bool                   `json:"active"`
    CreatedAt           time.Time              `json:"created_at"`
    UpdatedAt           time.Time              `json:"updated_at"`
}

// RuleExample пример для правила.
type RuleExample struct {
    Title       string `json:"title"`
    Description string `json:"description"`
    ImageURL    string `json:"image_url,omitempty"`
}

// ComplianceReportResponse полный отчёт.
type ComplianceReportResponse struct {
    ReportID          string                      `json:"report_id"`
    SceneID           string                      `json:"scene_id"`
    BranchID          *string                     `json:"branch_id,omitempty"`
    GeneratedAt       time.Time                   `json:"generated_at"`
    Summary           *ComplianceReportSummary    `json:"summary"`
    Violations        []ComplianceViolation       `json:"violations"`
    Warnings          []ComplianceViolation       `json:"warnings"`
    Info              []ComplianceViolation       `json:"info"`
    CheckedElements   map[string]int              `json:"checked_elements"`
    Recommendations   []string                    `json:"recommendations"`
    LegalStatus       *LegalStatus                `json:"legal_status"`
}

// LegalStatus правовой статус.
type LegalStatus struct {
    CanLegalize           bool     `json:"can_legalize"`
    RequiredDocuments     []string `json:"required_documents"`
    EstimatedApprovalTime string   `json:"estimated_approval_time"`
}
```

## Основные правила

### Структурные (structural)

- `SNIP_LOAD_BEARING_001` — Запрет сноса несущих стен
- `SNIP_LOAD_BEARING_002` — Ограничения проёмов в несущих стенах
- `SNIP_STRUCTURAL_003` — Защита колонн и ригелей

### Сантехника (plumbing)

- `SNIP_PLUMBING_001` — Мокрые зоны над жилыми помещениями
- `SNIP_PLUMBING_002` — Расположение санузлов
- `SNIP_PLUMBING_003` — Доступ к стоякам

### Пожарная безопасность (fire_safety)

- `SNIP_FIRE_001` — Эвакуационные пути
- `SNIP_FIRE_002` — Ширина проходов
- `SNIP_FIRE_003` — Газовая плита в жилой комнате

### Общие (general)

- `SNIP_GENERAL_001` — Минимальная площадь жилой комнаты (9 м²)
- `SNIP_GENERAL_002` — Минимальная площадь кухни (5 м²)
- `SNIP_GENERAL_003` — Естественное освещение жилых комнат


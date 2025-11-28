# Схемы баз данных

## Обзор

Granula API использует гибридный подход к хранению данных:

| База данных | Назначение | Тип данных |
|-------------|------------|------------|
| **PostgreSQL** | Реляционные данные | Пользователи, воркспейсы, заявки |
| **MongoDB** | Документы | 3D сцены, ветки, чаты |
| **Redis** | Кэш/очереди | Сессии, rate limiting |
| **MinIO/S3** | Файлы | Планировки, рендеры |

## PostgreSQL Schema

### ER диаграмма

```
┌──────────────────┐       ┌──────────────────┐       ┌──────────────────┐
│      users       │       │   workspaces     │       │   floor_plans    │
├──────────────────┤       ├──────────────────┤       ├──────────────────┤
│ id (PK)          │       │ id (PK)          │       │ id (PK)          │
│ email            │◄──────│ owner_id (FK)    │◄──────│ workspace_id (FK)│
│ password_hash    │       │ name             │       │ file_path        │
│ name             │       │ description      │       │ file_type        │
│ avatar_url       │       │ settings         │       │ original_name    │
│ role             │       │ created_at       │       │ recognition_data │
│ email_verified   │       │ updated_at       │       │ status           │
│ created_at       │       │ deleted_at       │       │ created_at       │
│ updated_at       │       └────────┬─────────┘       │ updated_at       │
│ deleted_at       │                │                 └──────────────────┘
└──────────────────┘                │
         │                          │
         │                          │
         ▼                          ▼
┌──────────────────┐       ┌──────────────────┐
│   user_sessions  │       │workspace_members │
├──────────────────┤       ├──────────────────┤
│ id (PK)          │       │ id (PK)          │
│ user_id (FK)     │       │ workspace_id (FK)│
│ refresh_token    │       │ user_id (FK)     │
│ user_agent       │       │ role             │
│ ip_address       │       │ invited_by (FK)  │
│ expires_at       │       │ joined_at        │
│ created_at       │       └──────────────────┘
└──────────────────┘

┌──────────────────┐       ┌──────────────────┐       ┌──────────────────┐
│expert_requests   │       │compliance_rules  │       │ notifications    │
├──────────────────┤       ├──────────────────┤       ├──────────────────┤
│ id (PK)          │       │ id (PK)          │       │ id (PK)          │
│ workspace_id (FK)│       │ code             │       │ user_id (FK)     │
│ user_id (FK)     │       │ name             │       │ type             │
│ scene_id         │       │ category         │       │ title            │
│ branch_id        │       │ description      │       │ message          │
│ status           │       │ rule_config      │       │ data             │
│ contact_name     │       │ severity         │       │ read             │
│ contact_phone    │       │ source           │       │ created_at       │
│ contact_email    │       │ active           │       └──────────────────┘
│ comment          │       │ created_at       │
│ assigned_expert  │       │ updated_at       │
│ estimated_date   │       └──────────────────┘
│ created_at       │
│ updated_at       │
└──────────────────┘
```

### DDL: users

```sql
-- Таблица пользователей системы.
-- Содержит основную информацию для аутентификации и профиля.
CREATE TABLE users (
    -- Уникальный идентификатор пользователя (UUID v4)
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Email пользователя, используется для входа
    -- Уникальный, индексируется для быстрого поиска
    email VARCHAR(255) NOT NULL,
    
    -- Хэш пароля (bcrypt, cost=12)
    -- NULL для OAuth пользователей
    password_hash VARCHAR(255),
    
    -- Отображаемое имя пользователя
    name VARCHAR(255) NOT NULL,
    
    -- URL аватара в S3 (опционально)
    avatar_url VARCHAR(512),
    
    -- Роль пользователя в системе
    -- user - обычный пользователь
    -- admin - администратор системы
    -- expert - эксперт БТИ
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    
    -- Флаг подтверждения email
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- OAuth провайдер (google, yandex) если применимо
    oauth_provider VARCHAR(50),
    
    -- ID пользователя в OAuth провайдере
    oauth_id VARCHAR(255),
    
    -- Настройки пользователя (JSON)
    settings JSONB NOT NULL DEFAULT '{}',
    
    -- Временные метки
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- Ограничения
    CONSTRAINT users_email_unique UNIQUE (email),
    CONSTRAINT users_oauth_unique UNIQUE (oauth_provider, oauth_id),
    CONSTRAINT users_role_check CHECK (role IN ('user', 'admin', 'expert'))
);

-- Индексы для быстрого поиска
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_oauth ON users(oauth_provider, oauth_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role ON users(role) WHERE deleted_at IS NULL;

-- Триггер автоматического обновления updated_at
CREATE TRIGGER users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Комментарии
COMMENT ON TABLE users IS 'Пользователи системы Granula';
COMMENT ON COLUMN users.id IS 'Уникальный идентификатор (UUID v4)';
COMMENT ON COLUMN users.email IS 'Email для аутентификации (уникальный)';
COMMENT ON COLUMN users.password_hash IS 'Bcrypt хэш пароля (cost=12)';
COMMENT ON COLUMN users.role IS 'Роль: user, admin, expert';
```

### DDL: workspaces

```sql
-- Воркспейсы (проекты) пользователей.
-- Каждый воркспейс представляет один проект ремонта/перепланировки.
CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Владелец воркспейса
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Название проекта
    name VARCHAR(255) NOT NULL,
    
    -- Описание проекта (опционально)
    description TEXT,
    
    -- Адрес квартиры (опционально)
    address TEXT,
    
    -- Общая площадь в м² (опционально)
    total_area DECIMAL(10, 2),
    
    -- Количество комнат (опционально)
    rooms_count SMALLINT,
    
    -- Настройки воркспейса
    -- {
    --   "units": "metric",           // metric | imperial
    --   "gridSize": 0.1,             // размер сетки в метрах
    --   "wallHeight": 2.7,           // высота стен по умолчанию
    --   "snapToGrid": true,          // привязка к сетке
    --   "showDimensions": true       // показывать размеры
    -- }
    settings JSONB NOT NULL DEFAULT '{
        "units": "metric",
        "gridSize": 0.1,
        "wallHeight": 2.7,
        "snapToGrid": true,
        "showDimensions": true
    }',
    
    -- Текущий статус проекта
    -- draft - черновик
    -- active - активный
    -- completed - завершён
    -- archived - в архиве
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    
    -- Превью изображение в S3
    preview_url VARCHAR(512),
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    CONSTRAINT workspaces_status_check CHECK (
        status IN ('draft', 'active', 'completed', 'archived')
    )
);

CREATE INDEX idx_workspaces_owner ON workspaces(owner_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_workspaces_status ON workspaces(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_workspaces_created ON workspaces(created_at DESC) WHERE deleted_at IS NULL;

CREATE TRIGGER workspaces_updated_at
    BEFORE UPDATE ON workspaces
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE workspaces IS 'Проекты ремонта/перепланировки пользователей';
```

### DDL: workspace_members

```sql
-- Участники воркспейса для совместной работы.
CREATE TABLE workspace_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Роль в воркспейсе
    -- owner - владелец (полные права)
    -- editor - редактор (может изменять)
    -- viewer - просмотр (только чтение)
    role VARCHAR(50) NOT NULL DEFAULT 'viewer',
    
    -- Кто пригласил
    invited_by UUID REFERENCES users(id) ON DELETE SET NULL,
    
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT workspace_members_unique UNIQUE (workspace_id, user_id),
    CONSTRAINT workspace_members_role_check CHECK (
        role IN ('owner', 'editor', 'viewer')
    )
);

CREATE INDEX idx_workspace_members_workspace ON workspace_members(workspace_id);
CREATE INDEX idx_workspace_members_user ON workspace_members(user_id);

COMMENT ON TABLE workspace_members IS 'Участники воркспейса с ролями доступа';
```

### DDL: floor_plans

```sql
-- Загруженные планировки квартир.
-- Хранит метаданные и результаты распознавания.
CREATE TABLE floor_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    
    -- Путь к файлу в S3
    file_path VARCHAR(512) NOT NULL,
    
    -- Тип файла (mime type)
    file_type VARCHAR(100) NOT NULL,
    
    -- Оригинальное имя файла
    original_name VARCHAR(255) NOT NULL,
    
    -- Размер файла в байтах
    file_size BIGINT NOT NULL,
    
    -- Статус обработки
    -- pending - ожидает обработки
    -- processing - в процессе распознавания
    -- completed - успешно распознано
    -- failed - ошибка распознавания
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    
    -- Результат распознавания (JSON структура)
    -- См. документацию models/entities.md для структуры
    recognition_data JSONB,
    
    -- Ошибка обработки (если status = failed)
    error_message TEXT,
    
    -- Источник планировки
    -- bti - техпаспорт БТИ
    -- technical_plan - технический план
    -- sketch - рукописный эскиз
    -- other - другое
    source_type VARCHAR(50) NOT NULL DEFAULT 'other',
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT floor_plans_status_check CHECK (
        status IN ('pending', 'processing', 'completed', 'failed')
    ),
    CONSTRAINT floor_plans_source_check CHECK (
        source_type IN ('bti', 'technical_plan', 'sketch', 'other')
    )
);

CREATE INDEX idx_floor_plans_workspace ON floor_plans(workspace_id);
CREATE INDEX idx_floor_plans_status ON floor_plans(status);

CREATE TRIGGER floor_plans_updated_at
    BEFORE UPDATE ON floor_plans
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE floor_plans IS 'Загруженные планировки для распознавания';
COMMENT ON COLUMN floor_plans.recognition_data IS 'JSON с распознанными элементами планировки';
```

### DDL: expert_requests

```sql
-- Заявки на экспертизу и оформление документов.
CREATE TABLE expert_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Связи
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- ID сцены в MongoDB
    scene_id VARCHAR(24) NOT NULL,
    
    -- ID ветки в MongoDB (опционально)
    branch_id VARCHAR(24),
    
    -- Статус заявки
    -- pending - ожидает рассмотрения
    -- reviewing - на рассмотрении
    -- approved - одобрена
    -- rejected - отклонена
    -- in_progress - в работе
    -- completed - выполнена
    -- cancelled - отменена
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    
    -- Тип услуги
    -- consultation - консультация
    -- documentation - оформление документов
    -- expert_visit - выезд эксперта
    -- full_service - полный комплекс
    service_type VARCHAR(50) NOT NULL,
    
    -- Контактные данные
    contact_name VARCHAR(255) NOT NULL,
    contact_phone VARCHAR(50) NOT NULL,
    contact_email VARCHAR(255) NOT NULL,
    
    -- Предпочтительное время связи
    preferred_contact_time VARCHAR(100),
    
    -- Комментарий пользователя
    comment TEXT,
    
    -- Данные обработки
    assigned_expert_id UUID REFERENCES users(id) ON DELETE SET NULL,
    estimated_date DATE,
    estimated_price DECIMAL(12, 2),
    
    -- Причина отклонения (если status = rejected)
    rejection_reason TEXT,
    
    -- История изменений статуса (JSON array)
    status_history JSONB NOT NULL DEFAULT '[]',
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT expert_requests_status_check CHECK (
        status IN ('pending', 'reviewing', 'approved', 'rejected', 
                   'in_progress', 'completed', 'cancelled')
    ),
    CONSTRAINT expert_requests_service_check CHECK (
        service_type IN ('consultation', 'documentation', 
                        'expert_visit', 'full_service')
    )
);

CREATE INDEX idx_expert_requests_workspace ON expert_requests(workspace_id);
CREATE INDEX idx_expert_requests_user ON expert_requests(user_id);
CREATE INDEX idx_expert_requests_status ON expert_requests(status);
CREATE INDEX idx_expert_requests_expert ON expert_requests(assigned_expert_id);
CREATE INDEX idx_expert_requests_created ON expert_requests(created_at DESC);

CREATE TRIGGER expert_requests_updated_at
    BEFORE UPDATE ON expert_requests
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE expert_requests IS 'Заявки на экспертизу и оформление документации';
```

### DDL: compliance_rules

```sql
-- Справочник строительных норм и правил.
CREATE TABLE compliance_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Код правила (уникальный)
    code VARCHAR(50) NOT NULL,
    
    -- Название правила
    name VARCHAR(255) NOT NULL,
    
    -- Категория
    -- structural - несущие конструкции
    -- plumbing - сантехника
    -- electrical - электрика
    -- ventilation - вентиляция
    -- fire_safety - пожарная безопасность
    -- accessibility - доступность
    -- general - общие требования
    category VARCHAR(50) NOT NULL,
    
    -- Подробное описание
    description TEXT NOT NULL,
    
    -- Конфигурация правила для валидации (JSON)
    -- {
    --   "type": "min_distance",
    --   "params": {
    --     "from": "toilet",
    --     "to": "kitchen",
    --     "minDistance": 3.0
    --   }
    -- }
    rule_config JSONB NOT NULL,
    
    -- Критичность нарушения
    -- error - критическая ошибка (запрещено)
    -- warning - предупреждение (нежелательно)
    -- info - информация (рекомендация)
    severity VARCHAR(20) NOT NULL DEFAULT 'error',
    
    -- Источник нормы
    source VARCHAR(255) NOT NULL,
    
    -- Ссылка на документ
    source_url VARCHAR(512),
    
    -- Флаг активности
    active BOOLEAN NOT NULL DEFAULT TRUE,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT compliance_rules_code_unique UNIQUE (code),
    CONSTRAINT compliance_rules_category_check CHECK (
        category IN ('structural', 'plumbing', 'electrical', 
                    'ventilation', 'fire_safety', 'accessibility', 'general')
    ),
    CONSTRAINT compliance_rules_severity_check CHECK (
        severity IN ('error', 'warning', 'info')
    )
);

CREATE INDEX idx_compliance_rules_category ON compliance_rules(category) WHERE active = TRUE;
CREATE INDEX idx_compliance_rules_code ON compliance_rules(code);

CREATE TRIGGER compliance_rules_updated_at
    BEFORE UPDATE ON compliance_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE compliance_rules IS 'Справочник строительных норм (СНиП, ЖК РФ)';
```

### DDL: notifications

```sql
-- Уведомления пользователей.
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Тип уведомления
    -- request_status - изменение статуса заявки
    -- compliance_warning - предупреждение о нарушении
    -- workspace_invite - приглашение в воркспейс
    -- system - системное уведомление
    type VARCHAR(50) NOT NULL,
    
    -- Заголовок уведомления
    title VARCHAR(255) NOT NULL,
    
    -- Текст уведомления
    message TEXT NOT NULL,
    
    -- Дополнительные данные (JSON)
    data JSONB,
    
    -- Флаг прочтения
    read BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Дата прочтения
    read_at TIMESTAMPTZ,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT notifications_type_check CHECK (
        type IN ('request_status', 'compliance_warning', 
                'workspace_invite', 'system')
    )
);

CREATE INDEX idx_notifications_user ON notifications(user_id);
CREATE INDEX idx_notifications_unread ON notifications(user_id, read) WHERE read = FALSE;
CREATE INDEX idx_notifications_created ON notifications(created_at DESC);

COMMENT ON TABLE notifications IS 'Уведомления пользователей';
```

### DDL: user_sessions

```sql
-- Сессии пользователей для refresh токенов.
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Хэш refresh токена (SHA-256)
    refresh_token_hash VARCHAR(64) NOT NULL,
    
    -- Информация об устройстве
    user_agent TEXT,
    ip_address INET,
    
    -- Device fingerprint для дополнительной безопасности
    device_id VARCHAR(255),
    
    -- Время истечения
    expires_at TIMESTAMPTZ NOT NULL,
    
    -- Время последнего использования
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT user_sessions_token_unique UNIQUE (refresh_token_hash)
);

CREATE INDEX idx_user_sessions_user ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_token ON user_sessions(refresh_token_hash);
CREATE INDEX idx_user_sessions_expires ON user_sessions(expires_at);

COMMENT ON TABLE user_sessions IS 'Активные сессии пользователей';
```

### Общие функции и триггеры

```sql
-- Функция обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Функция для очистки устаревших сессий
CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM user_sessions WHERE expires_at < NOW();
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Scheduled job для очистки сессий (pg_cron)
-- SELECT cron.schedule('cleanup-sessions', '0 * * * *', 'SELECT cleanup_expired_sessions()');
```

---

## MongoDB Schema

### Collection: scenes

```javascript
// 3D сцены квартир
// Database: granula
// Collection: scenes

{
  // Уникальный идентификатор сцены
  "_id": ObjectId("..."),
  
  // ID воркспейса в PostgreSQL
  "workspaceId": "uuid-string",
  
  // ID исходной планировки в PostgreSQL
  "floorPlanId": "uuid-string",
  
  // Название сцены
  "name": "Основная планировка",
  
  // Описание (опционально)
  "description": "Исходная планировка из техпаспорта БТИ",
  
  // Версия схемы данных для миграций
  "schemaVersion": 1,
  
  // Габариты помещения (в метрах)
  "bounds": {
    "width": 12.5,    // Ширина по X
    "height": 2.7,    // Высота потолков
    "depth": 8.3      // Глубина по Z
  },
  
  // Элементы сцены
  "elements": {
    // Стены
    "walls": [
      {
        "id": "wall-uuid",
        "type": "wall",
        "name": "Стена 1",
        
        // Геометрия стены
        "start": { "x": 0, "y": 0, "z": 0 },
        "end": { "x": 5.0, "y": 0, "z": 0 },
        "height": 2.7,
        "thickness": 0.2,
        
        // Свойства
        "properties": {
          "isLoadBearing": true,     // Несущая стена
          "material": "brick",        // Материал
          "canDemolish": false        // Можно ли сносить
        },
        
        // Проёмы в стене
        "openings": [
          {
            "id": "opening-uuid",
            "type": "door",           // door | window
            "position": 2.0,          // Позиция от начала стены
            "width": 0.9,
            "height": 2.1,
            "elevation": 0            // Высота от пола
          }
        ],
        
        // Метаданные для редактора
        "metadata": {
          "locked": false,
          "visible": true,
          "selected": false
        }
      }
    ],
    
    // Комнаты/зоны
    "rooms": [
      {
        "id": "room-uuid",
        "type": "room",
        "name": "Кухня",
        
        // Тип помещения
        "roomType": "kitchen",        // kitchen | bedroom | bathroom | living | corridor | storage
        
        // Полигон комнаты (замкнутый контур)
        "polygon": [
          { "x": 0, "z": 0 },
          { "x": 4.0, "z": 0 },
          { "x": 4.0, "z": 3.5 },
          { "x": 0, "z": 3.5 }
        ],
        
        // Вычисляемые свойства
        "area": 14.0,                 // Площадь в м²
        "perimeter": 15.0,            // Периметр в м
        
        // Свойства
        "properties": {
          "hasWetZone": true,         // Мокрая зона
          "hasVentilation": true,     // Вентиляция
          "minArea": 5.0              // Минимальная допустимая площадь
        }
      }
    ],
    
    // Мебель и объекты
    "furniture": [
      {
        "id": "furniture-uuid",
        "type": "furniture",
        "name": "Диван",
        
        // Тип мебели
        "furnitureType": "sofa",      // sofa | bed | table | chair | wardrobe | kitchen_set | ...
        
        // Позиция и ориентация
        "position": { "x": 2.0, "y": 0, "z": 1.5 },
        "rotation": { "x": 0, "y": 90, "z": 0 },
        
        // Размеры
        "dimensions": {
          "width": 2.0,
          "height": 0.85,
          "depth": 0.9
        },
        
        // 3D модель (опционально)
        "modelUrl": "s3://granula/models/sofa-01.glb",
        
        // Метаданные
        "metadata": {
          "category": "living_room",
          "color": "#8B4513"
        }
      }
    ],
    
    // Инженерные элементы
    "utilities": [
      {
        "id": "utility-uuid",
        "type": "utility",
        "name": "Стояк отопления",
        
        // Тип
        "utilityType": "heating_riser", // water_riser | heating_riser | gas_riser | ventilation | electrical
        
        "position": { "x": 0.5, "y": 0, "z": 2.0 },
        
        "properties": {
          "canRelocate": false,       // Можно ли переносить
          "protectionZone": 0.3       // Зона защиты в метрах
        }
      }
    ]
  },
  
  // Настройки отображения
  "displaySettings": {
    "floorTexture": "wood_oak",
    "wallColor": "#FFFFFF",
    "ceilingColor": "#F5F5F5",
    "ambientLight": 0.6,
    "showGrid": true,
    "gridSize": 0.5
  },
  
  // Результаты проверки compliance
  "complianceResult": {
    "lastCheckedAt": ISODate("2024-01-15T10:30:00Z"),
    "isCompliant": false,
    "violations": [
      {
        "ruleCode": "SNIP_2.08.01-89_4.1",
        "severity": "error",
        "message": "Площадь кухни менее 5 м²",
        "affectedElements": ["room-uuid-kitchen"],
        "suggestion": "Увеличьте площадь кухни минимум до 5 м²"
      }
    ],
    "warnings": [
      {
        "ruleCode": "SNIP_2.08.01-89_4.2",
        "severity": "warning",
        "message": "Рекомендуемая ширина коридора - 1.1 м",
        "affectedElements": ["room-uuid-corridor"]
      }
    ]
  },
  
  // Статистика
  "stats": {
    "totalArea": 65.5,
    "roomsCount": 4,
    "wallsCount": 12,
    "furnitureCount": 15
  },
  
  // Временные метки
  "createdAt": ISODate("2024-01-10T08:00:00Z"),
  "updatedAt": ISODate("2024-01-15T10:35:00Z"),
  
  // Кто создал/изменил
  "createdBy": "user-uuid",
  "updatedBy": "user-uuid"
}

// Индексы
db.scenes.createIndex({ "workspaceId": 1 });
db.scenes.createIndex({ "floorPlanId": 1 });
db.scenes.createIndex({ "createdAt": -1 });
db.scenes.createIndex({ "complianceResult.isCompliant": 1 });
```

### Collection: branches

```javascript
// Ветки дизайна (варианты планировки)
// Database: granula
// Collection: branches

{
  "_id": ObjectId("..."),
  
  // ID родительской сцены
  "sceneId": ObjectId("..."),
  
  // ID родительской ветки (null для корневых)
  "parentBranchId": ObjectId("...") | null,
  
  // Название ветки
  "name": "Вариант с объединённой кухней",
  
  // Описание
  "description": "Объединение кухни и гостиной, перенос двери",
  
  // Источник создания
  // user - создано пользователем
  // ai - сгенерировано AI
  "source": "ai",
  
  // Порядок сортировки среди siblings
  "order": 0,
  
  // Является ли активной/выбранной
  "isActive": true,
  
  // Является ли избранной
  "isFavorite": false,
  
  // Дельта изменений относительно родителя
  // Хранит только изменённые элементы
  "delta": {
    // Добавленные элементы
    "added": {
      "walls": [...],
      "rooms": [...],
      "furniture": [...]
    },
    
    // Изменённые элементы (id -> изменения)
    "modified": {
      "wall-uuid-1": {
        "end": { "x": 3.0, "y": 0, "z": 0 }  // Стена укорочена
      },
      "room-uuid-kitchen": {
        "polygon": [...],  // Новый контур
        "area": 18.5       // Новая площадь
      }
    },
    
    // Удалённые элементы (массив id)
    "removed": [
      "wall-uuid-2",
      "furniture-uuid-3"
    ]
  },
  
  // Полный снимок сцены (для быстрого доступа)
  // Пересчитывается при изменениях
  "snapshot": {
    // Полная копия elements из scenes
    "elements": { ... },
    "bounds": { ... },
    "stats": { ... }
  },
  
  // Результат compliance для этой ветки
  "complianceResult": { ... },
  
  // AI контекст (если source = ai)
  "aiContext": {
    "prompt": "Объедини кухню и гостиную",
    "model": "anthropic/claude-sonnet-4",
    "generatedAt": ISODate("..."),
    "reasoning": "Для объединения помещений удалена ненесущая перегородка..."
  },
  
  // Превью изображение
  "previewUrl": "s3://granula/previews/branch-xxx.png",
  
  "createdAt": ISODate("..."),
  "updatedAt": ISODate("..."),
  "createdBy": "user-uuid"
}

// Индексы
db.branches.createIndex({ "sceneId": 1 });
db.branches.createIndex({ "parentBranchId": 1 });
db.branches.createIndex({ "sceneId": 1, "isActive": 1 });
db.branches.createIndex({ "createdAt": -1 });
```

### Collection: chat_messages

```javascript
// Сообщения чата с AI
// Database: granula
// Collection: chat_messages

{
  "_id": ObjectId("..."),
  
  // ID сцены
  "sceneId": ObjectId("..."),
  
  // ID ветки (опционально, если чат привязан к ветке)
  "branchId": ObjectId("...") | null,
  
  // Роль отправителя
  // user - сообщение пользователя
  // assistant - ответ AI
  // system - системное сообщение
  "role": "user",
  
  // Текст сообщения
  "content": "Хочу объединить кухню с гостиной, но не знаю можно ли...",
  
  // Тип сообщения
  // text - обычный текст
  // suggestion - предложение AI с вариантами
  // action - выполненное действие
  // error - ошибка
  "messageType": "text",
  
  // Сгенерированные варианты (для suggestion)
  "suggestions": [
    {
      "branchId": ObjectId("..."),
      "title": "Вариант 1: Полное объединение",
      "description": "Снос перегородки между кухней и гостиной",
      "previewUrl": "s3://...",
      "isCompliant": true
    },
    {
      "branchId": ObjectId("..."),
      "title": "Вариант 2: Частичное объединение",
      "description": "Создание широкого проёма с барной стойкой",
      "previewUrl": "s3://...",
      "isCompliant": true
    }
  ],
  
  // Выбранный вариант (если пользователь выбрал)
  "selectedSuggestionIndex": 0,
  
  // Метаданные AI запроса
  "aiMetadata": {
    "model": "anthropic/claude-sonnet-4",
    "promptTokens": 1520,
    "completionTokens": 890,
    "totalTokens": 2410,
    "latencyMs": 2340
  },
  
  // Контекст сцены на момент сообщения
  "sceneContext": {
    "snapshotId": "snapshot-hash",  // Хэш состояния сцены
    "activeBranchId": ObjectId("...")
  },
  
  "createdAt": ISODate("..."),
  "userId": "user-uuid"
}

// Индексы
db.chat_messages.createIndex({ "sceneId": 1, "createdAt": 1 });
db.chat_messages.createIndex({ "sceneId": 1, "branchId": 1 });
db.chat_messages.createIndex({ "userId": 1 });
```

### Collection: ai_contexts

```javascript
// Контексты для AI (кэширование промптов)
// Database: granula
// Collection: ai_contexts

{
  "_id": ObjectId("..."),
  
  "sceneId": ObjectId("..."),
  
  // Тип контекста
  // recognition - распознавание планировки
  // generation - генерация вариантов
  // chat - диалог с пользователем
  // compliance - проверка норм
  "contextType": "chat",
  
  // Системный промпт
  "systemPrompt": "Ты - эксперт по планировке квартир...",
  
  // История сообщений для контекста
  "messageHistory": [
    {
      "role": "user",
      "content": "..."
    },
    {
      "role": "assistant",
      "content": "..."
    }
  ],
  
  // Сжатое описание сцены для контекста
  "sceneDescription": "Квартира 65м², 2 комнаты. Кухня 10м², гостиная 20м²...",
  
  // Embedding вектор для семантического поиска (опционально)
  "embedding": [0.1, -0.2, 0.3, ...],  // 1536 dimensions for text-embedding-3-small
  
  // TTL для автоочистки
  "expiresAt": ISODate("..."),
  
  "createdAt": ISODate("..."),
  "updatedAt": ISODate("...")
}

// Индексы
db.ai_contexts.createIndex({ "sceneId": 1, "contextType": 1 });
db.ai_contexts.createIndex({ "expiresAt": 1 }, { expireAfterSeconds: 0 });
```

---

## Redis Schema

### Структуры данных

```
# Сессии пользователей
# Hash
user:session:{session_id}
  - user_id: string
  - created_at: timestamp
  - expires_at: timestamp
  - device_info: json

# Кэш пользователя
# Hash with TTL
user:cache:{user_id}
  - email: string
  - name: string
  - role: string
  - settings: json
TTL: 15 минут

# Кэш воркспейса
# Hash with TTL
workspace:cache:{workspace_id}
  - name: string
  - owner_id: string
  - settings: json
TTL: 10 минут

# Rate Limiting (Sliding Window)
# Sorted Set
ratelimit:{identifier}:{endpoint}
  - score: timestamp (ms)
  - member: request_id
TTL: динамический (размер окна)

# Очередь задач AI
# List (FIFO queue)
queue:ai:{task_type}
  - json задачи

# Результаты AI задач
# String with TTL
ai:result:{task_id}
  - json результата
TTL: 1 час

# Блокировки (distributed locks)
# String with TTL
lock:{resource_type}:{resource_id}
  - lock_token
TTL: 30 секунд (с автопродлением)

# Онлайн статус пользователей
# Set
online:users
  - user_id

# WebSocket подключения
# Hash
ws:connections:{user_id}
  - connection_id: server_id
  
# Pub/Sub каналы
# Channels
channel:workspace:{workspace_id}  # Обновления воркспейса
channel:scene:{scene_id}          # Обновления сцены в реальном времени
channel:notifications:{user_id}   # Персональные уведомления
```

### Rate Limiting Implementation

```lua
-- scripts/ratelimit.lua
-- Скрипт для sliding window rate limiting

local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local request_id = ARGV[4]

-- Удаляем устаревшие записи
redis.call('ZREMRANGEBYSCORE', key, 0, now - window)

-- Получаем текущее количество запросов
local count = redis.call('ZCARD', key)

if count < limit then
    -- Добавляем новый запрос
    redis.call('ZADD', key, now, request_id)
    redis.call('PEXPIRE', key, window)
    return {1, limit - count - 1}  -- allowed, remaining
else
    -- Превышен лимит
    local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
    local retry_after = window - (now - oldest[2])
    return {0, retry_after}  -- denied, retry_after_ms
end
```

---

## MinIO/S3 Structure

```
granula/
├── floor-plans/
│   └── {workspace_id}/
│       └── {floor_plan_id}/
│           ├── original.{ext}      # Оригинальный файл
│           ├── processed.png       # Обработанное изображение
│           └── thumbnail.png       # Превью
│
├── models/
│   ├── furniture/                  # 3D модели мебели
│   │   ├── sofa-01.glb
│   │   ├── bed-01.glb
│   │   └── ...
│   └── elements/                   # 3D модели элементов
│       ├── door-01.glb
│       ├── window-01.glb
│       └── ...
│
├── renders/
│   └── {workspace_id}/
│       └── {scene_id}/
│           └── {branch_id}/
│               ├── preview.png     # Превью 2D
│               ├── render-3d.png   # 3D рендер
│               └── export.pdf      # Экспорт документа
│
├── avatars/
│   └── {user_id}.{ext}             # Аватары пользователей
│
└── exports/
    └── {workspace_id}/
        └── {export_id}/
            ├── project.pdf         # PDF документация
            ├── drawings.dwg        # CAD чертежи
            └── model.glb           # 3D модель
```

### Bucket Policies

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "PublicReadModels",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::granula/models/*"
    },
    {
      "Sid": "AuthenticatedReadRenders",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::account:role/granula-api"
      },
      "Action": [
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject"
      ],
      "Resource": "arn:aws:s3:::granula/renders/*"
    }
  ]
}
```

---

## Миграции

### PostgreSQL миграции (golang-migrate)

```
migrations/postgres/
├── 000001_init_schema.up.sql
├── 000001_init_schema.down.sql
├── 000002_add_compliance_rules.up.sql
├── 000002_add_compliance_rules.down.sql
├── 000003_add_notifications.up.sql
└── 000003_add_notifications.down.sql
```

### MongoDB миграции (скрипты)

```javascript
// migrations/mongodb/001_create_indexes.js
db = db.getSiblingDB('granula');

// Scenes indexes
db.scenes.createIndex({ "workspaceId": 1 });
db.scenes.createIndex({ "floorPlanId": 1 });
db.scenes.createIndex({ "createdAt": -1 });

// Branches indexes
db.branches.createIndex({ "sceneId": 1 });
db.branches.createIndex({ "parentBranchId": 1 });

// Chat messages indexes
db.chat_messages.createIndex({ "sceneId": 1, "createdAt": 1 });

// AI contexts TTL index
db.ai_contexts.createIndex(
  { "expiresAt": 1 },
  { expireAfterSeconds: 0 }
);

print("Indexes created successfully");
```


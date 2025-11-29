# 🗃️ WORKPLAN-5: Миграции баз данных

> **Приоритет:** 🟠 Важный  
> **Время:** 1-2 часа  
> **Зависимости:** WORKPLAN-1-PROTO.md  
> **Результат:** Все PostgreSQL сервисы имеют схему БД

---

## 🎯 ЦЕЛЬ

Создать SQL миграции для сервисов, у которых они отсутствуют:
1. auth-service (PostgreSQL)
2. user-service (PostgreSQL)
3. notification-service (PostgreSQL)

---

## 📋 ТЕКУЩЕЕ СОСТОЯНИЕ

### ✅ Сервисы с миграциями
| Сервис | БД | Миграции |
|--------|-----|----------|
| workspace-service | PostgreSQL | ✅ Есть |
| request-service | PostgreSQL | ✅ Есть |
| floorplan-service | PostgreSQL | ✅ Есть |
| compliance-service | PostgreSQL | ✅ Есть |

### ❌ Сервисы без миграций
| Сервис | БД | Миграции |
|--------|-----|----------|
| auth-service | PostgreSQL | ❌ Нет |
| user-service | PostgreSQL | ❌ Нет |
| notification-service | PostgreSQL | ❌ Нет |

### ℹ️ Сервисы на MongoDB (миграции не нужны)
| Сервис | БД | Миграции |
|--------|-----|----------|
| scene-service | MongoDB | Schema-less |
| branch-service | MongoDB | Schema-less |
| ai-service | MongoDB | Schema-less |

---

## 📁 СТРУКТУРА МИГРАЦИЙ

```
<service>/migrations/
├── 000001_init.up.sql
└── 000001_init.down.sql
```

Используется библиотека `golang-migrate`.

---

## 🔧 ПОШАГОВАЯ ИНСТРУКЦИЯ

### ШАГ 1: Auth Service миграции

**Документация моделей:** `docs/models/entities.md`

#### 1.1. Создать директорию

```powershell
New-Item -ItemType Directory -Force -Path auth-service/migrations
```

#### 1.2. Создать миграцию init

**Файл:** `auth-service/migrations/000001_init.up.sql`

```sql
-- Auth Service Database Schema
-- Version: 000001
-- Description: Initial schema for authentication

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table (minimal - full profile in user-service)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index on email for fast lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);

-- Refresh tokens table
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    device_info VARCHAR(255),
    ip_address VARCHAR(45),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for refresh_tokens
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- Password reset tokens table
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    used_at TIMESTAMP WITH TIME ZONE
);

-- Create index for password reset tokens
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_token_hash ON password_reset_tokens(token_hash);

-- Email verification tokens table
CREATE TABLE IF NOT EXISTS email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMP WITH TIME ZONE
);

-- Create index for email verification tokens
CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_user_id ON email_verification_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_token_hash ON email_verification_tokens(token_hash);

-- OAuth connections table (for future OAuth support)
CREATE TABLE IF NOT EXISTS oauth_connections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL, -- 'google', 'yandex'
    provider_user_id VARCHAR(255) NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(provider, provider_user_id)
);

-- Create index for oauth connections
CREATE INDEX IF NOT EXISTS idx_oauth_connections_user_id ON oauth_connections(user_id);
CREATE INDEX IF NOT EXISTS idx_oauth_connections_provider ON oauth_connections(provider, provider_user_id);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_oauth_connections_updated_at
    BEFORE UPDATE ON oauth_connections
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

**Файл:** `auth-service/migrations/000001_init.down.sql`

```sql
-- Rollback Auth Service Schema

DROP TRIGGER IF EXISTS update_oauth_connections_updated_at ON oauth_connections;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS oauth_connections;
DROP TABLE IF EXISTS email_verification_tokens;
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;
```

---

### ШАГ 2: User Service миграции

#### 2.1. Создать директорию

```powershell
New-Item -ItemType Directory -Force -Path user-service/migrations
```

#### 2.2. Создать миграцию init

**Файл:** `user-service/migrations/000001_init.up.sql`

```sql
-- User Service Database Schema
-- Version: 000001
-- Description: User profiles and settings

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- User profiles table
CREATE TABLE IF NOT EXISTS user_profiles (
    id UUID PRIMARY KEY, -- Same as auth user_id
    email VARCHAR(255) NOT NULL UNIQUE,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(20),
    avatar_url VARCHAR(500),
    timezone VARCHAR(50) DEFAULT 'Europe/Moscow',
    language VARCHAR(10) DEFAULT 'ru',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE -- Soft delete
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_user_profiles_email ON user_profiles(email);
CREATE INDEX IF NOT EXISTS idx_user_profiles_deleted_at ON user_profiles(deleted_at);

-- User settings table
CREATE TABLE IF NOT EXISTS user_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES user_profiles(id) ON DELETE CASCADE,
    
    -- Notification settings
    email_notifications BOOLEAN NOT NULL DEFAULT true,
    push_notifications BOOLEAN NOT NULL DEFAULT true,
    sms_notifications BOOLEAN NOT NULL DEFAULT false,
    
    -- Notification types
    notify_workspace_updates BOOLEAN NOT NULL DEFAULT true,
    notify_request_updates BOOLEAN NOT NULL DEFAULT true,
    notify_ai_complete BOOLEAN NOT NULL DEFAULT true,
    notify_compliance_warnings BOOLEAN NOT NULL DEFAULT true,
    
    -- Privacy settings
    profile_visibility VARCHAR(20) DEFAULT 'workspace', -- 'public', 'workspace', 'private'
    show_email BOOLEAN NOT NULL DEFAULT false,
    show_phone BOOLEAN NOT NULL DEFAULT false,
    
    -- Display settings
    theme VARCHAR(20) DEFAULT 'system', -- 'light', 'dark', 'system'
    compact_mode BOOLEAN NOT NULL DEFAULT false,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index for user settings
CREATE INDEX IF NOT EXISTS idx_user_settings_user_id ON user_settings(user_id);

-- User sessions table (for tracking active sessions)
CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    device_name VARCHAR(255),
    device_type VARCHAR(50), -- 'web', 'mobile', 'desktop'
    browser VARCHAR(100),
    os VARCHAR(100),
    ip_address VARCHAR(45),
    location VARCHAR(255),
    last_active_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for sessions
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_last_active_at ON user_sessions(last_active_at);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
CREATE TRIGGER update_user_profiles_updated_at
    BEFORE UPDATE ON user_profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_settings_updated_at
    BEFORE UPDATE ON user_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

**Файл:** `user-service/migrations/000001_init.down.sql`

```sql
-- Rollback User Service Schema

DROP TRIGGER IF EXISTS update_user_settings_updated_at ON user_settings;
DROP TRIGGER IF EXISTS update_user_profiles_updated_at ON user_profiles;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS user_sessions;
DROP TABLE IF EXISTS user_settings;
DROP TABLE IF EXISTS user_profiles;
```

---

### ШАГ 3: Notification Service миграции

#### 3.1. Создать директорию

```powershell
New-Item -ItemType Directory -Force -Path notification-service/migrations
```

#### 3.2. Создать миграцию init

**Файл:** `notification-service/migrations/000001_init.up.sql`

```sql
-- Notification Service Database Schema
-- Version: 000001
-- Description: Notifications and subscriptions

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Notification types enum
CREATE TYPE notification_type AS ENUM (
    'system',
    'workspace_member_added',
    'workspace_member_removed',
    'workspace_invitation',
    'request_submitted',
    'request_assigned',
    'request_status_changed',
    'request_completed',
    'compliance_warning',
    'compliance_error',
    'ai_recognition_complete',
    'ai_generation_complete',
    'branch_merged',
    'branch_conflict',
    'comment_added',
    'mention'
);

-- Notification priority enum
CREATE TYPE notification_priority AS ENUM (
    'low',
    'normal',
    'high',
    'urgent'
);

-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL, -- References user-service
    type notification_type NOT NULL,
    priority notification_priority NOT NULL DEFAULT 'normal',
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    data JSONB, -- Additional data (e.g., workspace_id, request_id)
    is_read BOOLEAN NOT NULL DEFAULT false,
    read_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for notifications
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user_is_read ON notifications(user_id, is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);
CREATE INDEX IF NOT EXISTS idx_notifications_data ON notifications USING gin(data);

-- Push subscriptions table (for Web Push)
CREATE TABLE IF NOT EXISTS push_subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    endpoint TEXT NOT NULL UNIQUE,
    p256dh_key TEXT NOT NULL,
    auth_key TEXT NOT NULL,
    user_agent VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for push subscriptions
CREATE INDEX IF NOT EXISTS idx_push_subscriptions_user_id ON push_subscriptions(user_id);

-- Notification settings table
CREATE TABLE IF NOT EXISTS notification_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE,
    
    -- Channel settings
    email_enabled BOOLEAN NOT NULL DEFAULT true,
    push_enabled BOOLEAN NOT NULL DEFAULT true,
    
    -- Type settings (what to notify about)
    workspace_updates BOOLEAN NOT NULL DEFAULT true,
    request_updates BOOLEAN NOT NULL DEFAULT true,
    ai_updates BOOLEAN NOT NULL DEFAULT true,
    compliance_updates BOOLEAN NOT NULL DEFAULT true,
    mentions BOOLEAN NOT NULL DEFAULT true,
    
    -- Quiet hours
    quiet_hours_enabled BOOLEAN NOT NULL DEFAULT false,
    quiet_hours_start TIME,
    quiet_hours_end TIME,
    quiet_hours_timezone VARCHAR(50) DEFAULT 'Europe/Moscow',
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index for notification settings
CREATE INDEX IF NOT EXISTS idx_notification_settings_user_id ON notification_settings(user_id);

-- Email queue table (for async email sending)
CREATE TABLE IF NOT EXISTS email_queue (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    email VARCHAR(255) NOT NULL,
    template VARCHAR(100) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    data JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'sent', 'failed'
    attempts INT NOT NULL DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    sent_at TIMESTAMP WITH TIME ZONE,
    scheduled_for TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for email queue
CREATE INDEX IF NOT EXISTS idx_email_queue_status ON email_queue(status);
CREATE INDEX IF NOT EXISTS idx_email_queue_scheduled ON email_queue(scheduled_for) WHERE status = 'pending';

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger for updated_at
CREATE TRIGGER update_notification_settings_updated_at
    BEFORE UPDATE ON notification_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

**Файл:** `notification-service/migrations/000001_init.down.sql`

```sql
-- Rollback Notification Service Schema

DROP TRIGGER IF EXISTS update_notification_settings_updated_at ON notification_settings;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS email_queue;
DROP TABLE IF EXISTS notification_settings;
DROP TABLE IF EXISTS push_subscriptions;
DROP TABLE IF EXISTS notifications;

DROP TYPE IF EXISTS notification_priority;
DROP TYPE IF EXISTS notification_type;
```

---

### ШАГ 4: Применить миграции

#### 4.1. Установить golang-migrate CLI (если не установлен)

```powershell
# Windows (scoop)
scoop install migrate

# Или через Go
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

#### 4.2. Применить миграции (вручную)

```powershell
# Auth Service
migrate -path auth-service/migrations -database "postgres://postgres:password@localhost:5432/auth_db?sslmode=disable" up

# User Service
migrate -path user-service/migrations -database "postgres://postgres:password@localhost:5433/user_db?sslmode=disable" up

# Notification Service
migrate -path notification-service/migrations -database "postgres://postgres:password@localhost:5439/notification_db?sslmode=disable" up
```

#### 4.3. Автоматические миграции при запуске (рекомендуется)

Добавить в `main.go` каждого сервиса:

```go
import (
    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func runMigrations(db *sql.DB, migrationsPath string) error {
    driver, err := postgres.WithInstance(db, &postgres.Config{})
    if err != nil {
        return err
    }

    m, err := migrate.NewWithDatabaseInstance(
        "file://"+migrationsPath,
        "postgres", 
        driver,
    )
    if err != nil {
        return err
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }

    return nil
}

// В main():
if err := runMigrations(db, "./migrations"); err != nil {
    log.Fatal("failed to run migrations", logger.Err(err))
}
```

---

### ШАГ 5: Обновить docker-compose.yml

**Добавить переменные для миграций:**

```yaml
services:
  auth-service:
    environment:
      - RUN_MIGRATIONS=true
    volumes:
      - ./auth-service/migrations:/app/migrations

  user-service:
    environment:
      - RUN_MIGRATIONS=true
    volumes:
      - ./user-service/migrations:/app/migrations

  notification-service:
    environment:
      - RUN_MIGRATIONS=true
    volumes:
      - ./notification-service/migrations:/app/migrations
```

---

## ✅ КРИТЕРИИ УСПЕХА

- [ ] Директория `auth-service/migrations` создана
- [ ] Миграция `000001_init.up.sql` для auth создана
- [ ] Миграция `000001_init.down.sql` для auth создана
- [ ] Директория `user-service/migrations` создана
- [ ] Миграция `000001_init.up.sql` для user создана
- [ ] Миграция `000001_init.down.sql` для user создана
- [ ] Директория `notification-service/migrations` создана
- [ ] Миграция `000001_init.up.sql` для notification создана
- [ ] Миграция `000001_init.down.sql` для notification создана
- [ ] Миграции успешно применяются командой `migrate up`
- [ ] Миграции успешно откатываются командой `migrate down`

---

## 🐛 ВОЗМОЖНЫЕ ПРОБЛЕМЫ

### Проблема: "relation already exists"
**Решение:** БД уже содержит таблицы. Либо дропнуть БД, либо использовать `IF NOT EXISTS`.

### Проблема: "migration dirty"
**Решение:** 
```powershell
migrate -path ./migrations -database "postgres://..." force VERSION
```

### Проблема: "cannot find package migrate"
**Решение:**
```powershell
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

---

## 📚 СВЯЗАННАЯ ДОКУМЕНТАЦИЯ

| Документ | Путь | Для чего |
|----------|------|----------|
| Entities | `docs/models/entities.md` | Описание моделей |
| Docker Compose | `docker-compose.yml` | БД конфигурация |
| Существующие миграции | `workspace-service/migrations/` | Референс |

---

## 🎉 ЗАВЕРШЕНИЕ

После создания всех миграций:

1. **Перезапустить docker-compose:**
   ```powershell
   docker-compose down
   docker-compose up -d
   ```

2. **Проверить логи сервисов:**
   ```powershell
   docker-compose logs auth-service
   docker-compose logs user-service
   docker-compose logs notification-service
   ```

3. **Подключиться к БД и проверить таблицы:**
   ```powershell
   docker exec -it granula-postgres-auth psql -U postgres -d auth_db -c "\dt"
   ```

---

*После успешного создания миграций, весь API готов к тестированию!*


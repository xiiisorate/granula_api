# Микросервисная архитектура Granula

## Обзор

Система разделена на независимые микросервисы для:
- Независимого масштабирования компонентов
- Параллельной разработки командой
- Изоляции отказов
- Гибкости технологического стека

---

## Схема архитектуры

```
                                    ┌─────────────────┐
                                    │   Load Balancer │
                                    │    (Traefik)    │
                                    └────────┬────────┘
                                             │
                                             ▼
┌────────────────────────────────────────────────────────────────────────────┐
│                              API GATEWAY                                    │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │  • REST API routing                                                   │  │
│  │  • JWT validation                                                     │  │
│  │  • Rate limiting                                                      │  │
│  │  • Request/Response transformation                                    │  │
│  │  • WebSocket proxy                                                    │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                    Port: 8080                               │
└────────────────────────────────────────┬───────────────────────────────────┘
                                         │
                    ┌────────────────────┼────────────────────┐
                    │                    │                    │
        ┌───────────┴──────┐  ┌─────────┴─────────┐  ┌──────┴───────────┐
        │   gRPC (внутр)   │  │   gRPC (внутр)    │  │   gRPC (внутр)   │
        ▼                  ▼  ▼                   ▼  ▼                  ▼
┌───────────────┐  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐
│  AUTH SERVICE │  │  USER SERVICE │  │WORKSPACE SVC  │  │FLOOR PLAN SVC │
│               │  │               │  │               │  │               │
│ • Register    │  │ • Profile     │  │ • CRUD        │  │ • Upload      │
│ • Login       │  │ • Settings    │  │ • Members     │  │ • Recognition │
│ • OAuth       │  │ • Sessions    │  │ • Invites     │  │ • Processing  │
│ • JWT/Refresh │  │ • Avatar      │  │               │  │               │
│               │  │               │  │               │  │               │
│ Port: 50051   │  │ Port: 50052   │  │ Port: 50053   │  │ Port: 50054   │
└───────┬───────┘  └───────┬───────┘  └───────┬───────┘  └───────┬───────┘
        │                  │                  │                  │
        │                  │                  │                  │
        ▼                  ▼                  ▼                  ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                            PostgreSQL                                    │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐                    │
│  │auth_db  │  │users_db │  │worksp_db│  │plans_db │                    │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘                    │
└─────────────────────────────────────────────────────────────────────────┘

┌───────────────┐  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐
│ SCENE SERVICE │  │ BRANCH SERVICE│  │   AI SERVICE  │  │COMPLIANCE SVC │
│               │  │               │  │               │  │               │
│ • CRUD scenes │  │ • CRUD branch │  │ • Recognition │  │ • Rules DB    │
│ • Elements    │  │ • Delta/Merge │  │ • Generation  │  │ • Validation  │
│ • Snapshots   │  │ • Compare     │  │ • Chat        │  │ • Reports     │
│ • Render      │  │ • Activate    │  │ • Streaming   │  │               │
│               │  │               │  │               │  │               │
│ Port: 50055   │  │ Port: 50056   │  │ Port: 50057   │  │ Port: 50058   │
└───────┬───────┘  └───────┬───────┘  └───────┬───────┘  └───────┬───────┘
        │                  │                  │                  │
        ▼                  ▼                  │                  │
┌─────────────────────────────────────┐      │                  │
│              MongoDB                 │      │                  │
│  ┌─────────┐  ┌─────────┐          │      │                  │
│  │scenes_db│  │branches │          │      ▼                  │
│  └─────────┘  └─────────┘          │  ┌───────────┐          │
└─────────────────────────────────────┘  │OpenRouter │          │
                                         │   API     │          │
                                         └───────────┘          │
                                                                │
┌───────────────┐  ┌───────────────┐                           │
│REQUEST SERVICE│  │NOTIFICATION   │                           │
│               │  │    SERVICE    │                           │
│ • Create req  │  │ • In-app      │          ┌────────────────┘
│ • Status flow │  │ • Email       │          │
│ • Expert      │  │ • Push        │          ▼
│ • Documents   │  │ • WebSocket   │  ┌───────────────┐
│               │  │               │  │   Redis       │
│ Port: 50059   │  │ Port: 50060   │  │ • Cache       │
└───────┬───────┘  └───────┬───────┘  │ • Pub/Sub     │
        │                  │          │ • Sessions    │
        ▼                  │          └───────────────┘
┌─────────────────┐        │
│   PostgreSQL    │        │          ┌───────────────┐
│  requests_db    │        └─────────►│   MinIO/S3    │
└─────────────────┘                   │ • Files       │
                                      │ • Renders     │
                                      └───────────────┘
```

---

## Список микросервисов

### 1. API Gateway
**Порт:** 8080 (HTTP/WebSocket)

| Функция | Описание |
|---------|----------|
| Routing | Маршрутизация REST запросов к сервисам |
| Auth | Валидация JWT токенов |
| Rate Limiting | Ограничение частоты запросов |
| CORS | Cross-Origin Resource Sharing |
| Logging | Централизованное логирование запросов |
| WebSocket | Проксирование WS соединений |
| Aggregation | Объединение данных из нескольких сервисов |

**Технологии:** Go, Fiber, gRPC client

---

### 2. Auth Service
**Порт:** 50051 (gRPC)

| Endpoint | Описание |
|----------|----------|
| Register | Регистрация нового пользователя |
| Login | Вход по email/password |
| OAuthLogin | Вход через Google/Yandex |
| RefreshToken | Обновление access токена |
| Logout | Выход (инвалидация токенов) |
| ValidateToken | Валидация токена (для Gateway) |
| ResetPassword | Сброс пароля |
| VerifyEmail | Подтверждение email |

**База данных:** PostgreSQL (auth_db)
- users (id, email, password_hash, oauth)
- refresh_tokens
- email_verifications
- password_resets

---

### 3. User Service
**Порт:** 50052 (gRPC)

| Endpoint | Описание |
|----------|----------|
| GetProfile | Получение профиля |
| UpdateProfile | Обновление профиля |
| UploadAvatar | Загрузка аватара |
| GetSettings | Получение настроек |
| UpdateSettings | Обновление настроек |
| GetSessions | Список сессий |
| RevokeSession | Отзыв сессии |
| DeleteAccount | Удаление аккаунта |

**База данных:** PostgreSQL (users_db)
- user_profiles
- user_settings
- user_sessions

---

### 4. Workspace Service
**Порт:** 50053 (gRPC)

| Endpoint | Описание |
|----------|----------|
| CreateWorkspace | Создание воркспейса |
| GetWorkspace | Получение воркспейса |
| ListWorkspaces | Список воркспейсов пользователя |
| UpdateWorkspace | Обновление воркспейса |
| DeleteWorkspace | Удаление воркспейса |
| InviteMember | Приглашение участника |
| RemoveMember | Удаление участника |
| UpdateMemberRole | Изменение роли участника |

**База данных:** PostgreSQL (workspaces_db)
- workspaces
- workspace_members
- workspace_invites

---

### 5. Floor Plan Service
**Порт:** 50054 (gRPC)

| Endpoint | Описание |
|----------|----------|
| UploadFloorPlan | Загрузка планировки |
| GetFloorPlan | Получение планировки |
| ListFloorPlans | Список планировок воркспейса |
| UpdateFloorPlan | Обновление метаданных |
| DeleteFloorPlan | Удаление планировки |
| ProcessFloorPlan | Запуск распознавания |
| GetProcessingStatus | Статус обработки |
| CreateSceneFromPlan | Создание сцены из планировки |

**База данных:** PostgreSQL (floor_plans_db)
- floor_plans
- processing_jobs

**Интеграции:** AI Service, MinIO

---

### 6. Scene Service
**Порт:** 50055 (gRPC)

| Endpoint | Описание |
|----------|----------|
| CreateScene | Создание сцены |
| GetScene | Получение сцены |
| ListScenes | Список сцен воркспейса |
| UpdateScene | Обновление метаданных |
| DeleteScene | Удаление сцены |
| UpdateElements | Обновление элементов |
| ApplyOperation | Применение операции |
| DuplicateScene | Дублирование сцены |
| RenderScene | Запрос рендера |

**База данных:** MongoDB (scenes_db)
- scenes

**Интеграции:** Compliance Service, Branch Service

---

### 7. Branch Service
**Порт:** 50056 (gRPC)

| Endpoint | Описание |
|----------|----------|
| CreateBranch | Создание ветки |
| GetBranch | Получение ветки |
| ListBranches | Список веток сцены |
| UpdateBranch | Обновление ветки |
| DeleteBranch | Удаление ветки |
| UpdateDelta | Обновление изменений |
| ActivateBranch | Активация ветки |
| CompareBranches | Сравнение веток |
| MergeBranch | Слияние веток |
| DuplicateBranch | Дублирование ветки |

**База данных:** MongoDB (branches_db)
- branches

---

### 8. AI Service
**Порт:** 50057 (gRPC + gRPC Streaming)

| Endpoint | Описание |
|----------|----------|
| RecognizeFloorPlan | Распознавание планировки |
| GenerateVariants | Генерация вариантов |
| SendChatMessage | Отправка сообщения в чат |
| StreamChatResponse | Streaming ответа |
| GetChatHistory | История чата |
| ClearChatHistory | Очистка истории |
| ResetContext | Сброс контекста |

**База данных:** MongoDB (ai_db)
- chat_messages
- ai_contexts
- generation_jobs

**Интеграции:** OpenRouter API

---

### 9. Compliance Service
**Порт:** 50058 (gRPC)

| Endpoint | Описание |
|----------|----------|
| CheckCompliance | Проверка соответствия |
| CheckOperation | Проверка операции |
| GetRules | Справочник правил |
| GetRule | Детали правила |
| GenerateReport | Генерация отчёта |

**База данных:** PostgreSQL (compliance_db)
- compliance_rules
- rule_categories

---

### 10. Request Service
**Порт:** 50059 (gRPC)

| Endpoint | Описание |
|----------|----------|
| CreateRequest | Создание заявки |
| GetRequest | Получение заявки |
| ListRequests | Список заявок |
| UpdateRequest | Обновление заявки |
| CancelRequest | Отмена заявки |
| AssignExpert | Назначение эксперта |
| UpdateStatus | Изменение статуса |
| UploadDocument | Загрузка документа |

**База данных:** PostgreSQL (requests_db)
- expert_requests
- request_documents
- status_history

---

### 11. Notification Service
**Порт:** 50060 (gRPC)

| Endpoint | Описание |
|----------|----------|
| SendNotification | Отправка уведомления |
| GetNotifications | Список уведомлений |
| MarkAsRead | Отметить прочитанным |
| GetUnreadCount | Количество непрочитанных |
| UpdateSettings | Настройки уведомлений |
| SubscribePush | Подписка на push |
| SendEmail | Отправка email |

**База данных:** PostgreSQL (notifications_db)
- notifications
- notification_settings
- push_subscriptions
- email_templates

**Интеграции:** SMTP, WebSocket (через Gateway)

---

## Коммуникация между сервисами

### gRPC

Все внутренние коммуникации через gRPC:

```protobuf
// proto/auth/v1/auth.proto
syntax = "proto3";

package auth.v1;

service AuthService {
  rpc Register(RegisterRequest) returns (RegisterResponse);
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
}
```

### Event Bus (Redis Pub/Sub)

Асинхронная коммуникация через события:

| Event | Publisher | Subscribers |
|-------|-----------|-------------|
| `user.created` | Auth | User, Notification |
| `user.deleted` | User | Workspace, Request |
| `workspace.created` | Workspace | Notification |
| `floor_plan.processed` | Floor Plan | Scene, Notification |
| `scene.updated` | Scene | Branch, Compliance |
| `branch.created` | Branch, AI | Notification |
| `request.status_changed` | Request | Notification |
| `compliance.violation` | Compliance | Notification |

---

## Структура репозитория

```
granula/
├── api-gateway/
│   ├── cmd/
│   ├── internal/
│   ├── proto/
│   ├── Dockerfile
│   └── go.mod
├── auth-service/
│   ├── cmd/
│   ├── internal/
│   ├── proto/
│   ├── migrations/
│   ├── Dockerfile
│   └── go.mod
├── user-service/
│   └── ...
├── workspace-service/
│   └── ...
├── floor-plan-service/
│   └── ...
├── scene-service/
│   └── ...
├── branch-service/
│   └── ...
├── ai-service/
│   └── ...
├── compliance-service/
│   └── ...
├── request-service/
│   └── ...
├── notification-service/
│   └── ...
├── shared/
│   ├── proto/           # Общие proto файлы
│   ├── pkg/             # Общие пакеты
│   └── go.mod
├── deployments/
│   ├── docker-compose.yml
│   ├── docker-compose.dev.yml
│   └── k8s/
├── docs/
└── Makefile
```

---

## Docker Compose (Development)

```yaml
version: '3.8'

services:
  # Infrastructure
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: granula
      POSTGRES_PASSWORD: secret
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-dbs.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"

  mongodb:
    image: mongo:7
    volumes:
      - mongodb_data:/data/db
    ports:
      - "27017:27017"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  minio:
    image: minio/minio
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    ports:
      - "9000:9000"
      - "9001:9001"

  # Services
  api-gateway:
    build: ./api-gateway
    ports:
      - "8080:8080"
    depends_on:
      - auth-service
      - user-service
    environment:
      - AUTH_SERVICE_ADDR=auth-service:50051
      - USER_SERVICE_ADDR=user-service:50052
      # ...

  auth-service:
    build: ./auth-service
    ports:
      - "50051:50051"
    depends_on:
      - postgres
      - redis
    environment:
      - POSTGRES_DSN=postgres://granula:secret@postgres:5432/auth_db
      - REDIS_URL=redis://redis:6379

  user-service:
    build: ./user-service
    ports:
      - "50052:50052"
    depends_on:
      - postgres
      - minio

  # ... остальные сервисы

volumes:
  postgres_data:
  mongodb_data:
```

---

## Масштабирование

| Сервис | Стратегия | Когда масштабировать |
|--------|-----------|---------------------|
| API Gateway | Horizontal | CPU > 70% |
| Auth Service | Horizontal | RPS > 1000 |
| AI Service | Horizontal | Queue > 100 |
| Scene Service | Horizontal | Memory > 80% |
| Notification | Horizontal | Queue > 1000 |


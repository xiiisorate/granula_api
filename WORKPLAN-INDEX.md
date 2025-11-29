# 📋 ГЛАВНЫЙ ПЛАН РАБОТЫ — Granula API

> **Дата создания:** 29 ноября 2024  
> **Общее время:** ~12-16 часов  
> **Приоритет:** Запуск MVP для хакатона

---

## 🎯 ЦЕЛЬ

Довести API до рабочего состояния, где:
1. Все сервисы компилируются и запускаются
2. AI распознаёт планировки из фото
3. AI чат работает с контекстом сцены
4. Генерация вариантов работает
5. Все API endpoints доступны

---

## 📁 СТРУКТУРА ПЛАНА

План разделён на **5 модулей** для параллельной работы:

| # | Файл | Описание | Время | Приоритет |
|---|------|----------|-------|-----------|
| 1 | [WORKPLAN-1-PROTO.md](./WORKPLAN-1-PROTO.md) | Исправление proto и генерация Go кода | 1-2 ч | 🔴 Блокирующий |
| 2 | [WORKPLAN-2-API-GATEWAY.md](./WORKPLAN-2-API-GATEWAY.md) | Создание недостающих HTTP handlers | 3-4 ч | 🔴 Высокий |
| 3 | [WORKPLAN-3-AI-MODULE.md](./WORKPLAN-3-AI-MODULE.md) | Исправление AI модуля (Vision, контекст) | 4-5 ч | 🔴 Критический |
| 4 | [WORKPLAN-4-INTEGRATIONS.md](./WORKPLAN-4-INTEGRATIONS.md) | Интеграции между сервисами | 2-3 ч | 🟠 Важный |
| 5 | [WORKPLAN-5-MIGRATIONS.md](./WORKPLAN-5-MIGRATIONS.md) | Миграции БД для сервисов | 1-2 ч | 🟠 Важный |

---

## 🔄 ПОРЯДОК ВЫПОЛНЕНИЯ

```
┌─────────────────────────────────────────────────────────────────┐
│                     ЭТАП 1: БАЗОВАЯ СБОРКА                      │
│                        (Блокирующий)                            │
├─────────────────────────────────────────────────────────────────┤
│  WORKPLAN-1-PROTO.md                                            │
│  └── Исправить go_package во всех proto                         │
│  └── Сгенерировать Go код (protoc)                              │
│  └── go mod tidy во всех сервисах                               │
│  └── Проверить компиляцию                                       │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                 ЭТАП 2: ПАРАЛЛЕЛЬНАЯ РАБОТА                     │
│                                                                 │
│   ┌─────────────────────┐    ┌─────────────────────┐            │
│   │ WORKPLAN-2-API-GW   │    │ WORKPLAN-3-AI       │            │
│   │ (Dev 1)             │    │ (Dev 2)             │            │
│   │                     │    │                     │            │
│   │ - FloorPlan handler │    │ - Vision API        │            │
│   │ - Branch handler    │    │ - Scene integration │            │
│   │ - Compliance handler│    │ - Recognition fix   │            │
│   │ - Request handler   │    │ - SelectSuggestion  │            │
│   └─────────────────────┘    └─────────────────────┘            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     ЭТАП 3: ИНТЕГРАЦИЯ                          │
│                                                                 │
│   ┌─────────────────────┐    ┌─────────────────────┐            │
│   │ WORKPLAN-4-INTEG    │    │ WORKPLAN-5-MIGR     │            │
│   │                     │    │                     │            │
│   │ - AI → Scene        │    │ - auth migrations   │            │
│   │ - Request → Notif   │    │ - user migrations   │            │
│   │ - Workspace → Notif │    │ - notif migrations  │            │
│   └─────────────────────┘    └─────────────────────┘            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                   ЭТАП 4: ТЕСТИРОВАНИЕ                          │
│                                                                 │
│  1. docker-compose up                                           │
│  2. Тест всех HTTP endpoints                                    │
│  3. Тест распознавания планов (загрузка фото)                   │
│  4. Тест AI чата с контекстом сцены                             │
│  5. Тест генерации вариантов                                    │
└─────────────────────────────────────────────────────────────────┘
```

---

## 📚 КЛЮЧЕВАЯ ДОКУМЕНТАЦИЯ

### Архитектура и обзор
| Документ | Путь | Описание |
|----------|------|----------|
| README | `docs/README.md` | Обзор проекта, стек, структура |
| Архитектура | `docs/architecture/microservices.md` | Микросервисы, порты, взаимодействие |
| Стек | `docs/STACK.md` | Технологии, библиотеки, паттерны |
| Сервис | `docs/SERVICE.md` | Бизнес-логика, user stories |

### API Endpoints
| Документ | Путь | Описание |
|----------|------|----------|
| Аутентификация | `docs/api/authentication.md` | JWT, OAuth, регистрация |
| Пользователи | `docs/api/users.md` | Профиль, сессии, аватары |
| Воркспейсы | `docs/api/workspaces.md` | Проекты, участники, приглашения |
| Планировки | `docs/api/floor-plans.md` | Загрузка, распознавание |
| Сцены | `docs/api/scenes.md` | 3D модель, элементы |
| Ветки | `docs/api/branches.md` | Версионирование, merge |
| AI Чат | `docs/api/chat.md` | Чат, генерация, стриминг |
| Compliance | `docs/api/compliance.md` | Проверка норм |
| Заявки | `docs/api/requests.md` | Экспертные заявки |
| Уведомления | `docs/api/notifications.md` | Push, WebSocket |

### Сущности
| Документ | Путь | Описание |
|----------|------|----------|
| Entities | `docs/models/entities.md` | Доменные модели Go |

---

## 🔍 БЫСТРАЯ НАВИГАЦИЯ ПО КОДУ

### Proto файлы
```
shared/proto/
├── ai/v1/ai.proto
├── auth/v1/auth.proto
├── branch/v1/branch.proto
├── common/v1/common.proto
├── compliance/v1/compliance.proto
├── floorplan/v1/floorplan.proto
├── notification/v1/notification.proto
├── request/v1/request.proto
├── scene/v1/scene.proto
├── user/v1/user.proto
└── workspace/v1/workspace.proto
```

### AI Service (критический)
```
ai-service/internal/
├── service/
│   ├── chat_service.go        # Чат с AI
│   ├── generation_service.go  # Генерация вариантов
│   └── recognition_service.go # ❌ СЛОМАНО: Распознавание
├── openrouter/
│   └── client.go              # OpenRouter API (нужен Vision)
├── prompts/
│   └── prompts.go             # ✅ Готово: 810 строк промптов
└── grpc/
    └── server.go              # gRPC handlers
```

### API Gateway handlers
```
api-gateway/internal/handlers/
├── auth.go                    # ✅ Готово
├── user_handler.go            # ✅ Готово
├── workspace.go               # ✅ Готово
├── scene.go                   # ✅ Готово
├── ai.go                      # ✅ Готово
├── notification_handler.go    # ✅ Готово
├── floorplan.go              # ❌ НЕТ — создать!
├── branch.go                  # ❌ НЕТ — создать!
├── compliance.go              # ❌ НЕТ — создать!
└── request.go                 # ❌ НЕТ — создать!
```

---

## ✅ ЧЕКЛИСТ ГОТОВНОСТИ К ЗАПУСКУ

### Блокирующие задачи
- [ ] Proto файлы сгенерированы в `shared/gen/`
- [ ] Все сервисы компилируются без ошибок
- [ ] `go mod tidy` выполнен во всех сервисах

### AI функциональность
- [ ] Recognition отправляет реальные изображения
- [ ] Chat получает данные сцены
- [ ] Generation получает данные сцены
- [ ] SelectSuggestion реализован

### API Gateway
- [ ] FloorPlan handlers созданы
- [ ] Branch handlers созданы
- [ ] Compliance handlers созданы
- [ ] Request handlers созданы

### Базы данных
- [ ] Миграции auth-service созданы
- [ ] Миграции user-service созданы
- [ ] Миграции notification-service созданы

### Интеграционное тестирование
- [ ] docker-compose up работает
- [ ] Health checks проходят
- [ ] Тест регистрации/логина
- [ ] Тест загрузки планировки
- [ ] Тест распознавания AI
- [ ] Тест чата с AI
- [ ] Тест генерации вариантов

---

## 🚀 БЫСТРЫЙ СТАРТ

```powershell
# 1. Сгенерировать proto (см. WORKPLAN-1-PROTO.md)
cd shared
.\scripts\generate-proto.ps1

# 2. Собрать все сервисы
cd ..
foreach ($svc in Get-ChildItem -Directory *-service) {
    Push-Location $svc.FullName
    go mod tidy
    go build ./...
    Pop-Location
}

# 3. Запустить инфраструктуру
docker-compose up -d postgres-auth postgres-user postgres-workspace postgres-request postgres-floorplan postgres-compliance mongodb redis minio

# 4. Запустить сервисы
docker-compose up -d

# 5. Проверить health
curl http://localhost:8080/health
```

---

*Перейдите к конкретному плану для детальных инструкций*


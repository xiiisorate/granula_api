# Granula API — Быстрый старт для разработчиков

> **Статус:** Ветка `dev/shared` создана, окружение настроено  
> **Команда:** Developer 1 (Core) + Developer 2 (AI/3D)

---

## 📚 Что такое Proto файлы?

### Кратко

**Protocol Buffers (protobuf)** — это формат сериализации данных от Google. Мы используем его для:

1. **Определения API контрактов** между микросервисами
2. **Автоматической генерации Go кода** (клиенты и серверы)
3. **Типобезопасной коммуникации** через gRPC

### Пример

```protobuf
// shared/proto/auth/v1/auth.proto

syntax = "proto3";                              // Версия protobuf
package auth.v1;                                // Пакет (namespace)
option go_package = "github.com/granula/shared/gen/auth/v1;authv1";  // Go import path

// Сервис — определяет какие методы доступны
service AuthService {
  rpc Register(RegisterRequest) returns (RegisterResponse);
  rpc Login(LoginRequest) returns (LoginResponse);
}

// Сообщения — структуры данных
message RegisterRequest {
  string email = 1;      // = 1, = 2 — это номера полей (не значения!)
  string password = 2;
  string name = 3;
}

message RegisterResponse {
  string user_id = 1;
  string access_token = 2;
  string refresh_token = 3;
}
```

### Как это работает

```
┌─────────────────┐     protoc      ┌─────────────────┐
│  auth.proto     │ ───────────────►│  auth.pb.go     │  (структуры)
│  (определение)  │                 │  auth_grpc.pb.go│  (клиент/сервер)
└─────────────────┘                 └─────────────────┘
                                            │
                    ┌───────────────────────┼───────────────────────┐
                    │                       │                       │
                    ▼                       ▼                       ▼
            ┌──────────────┐       ┌──────────────┐       ┌──────────────┐
            │ Auth Service │       │ API Gateway  │       │ Любой другой │
            │   (server)   │       │   (client)   │       │   сервис     │
            └──────────────┘       └──────────────┘       └──────────────┘
```

---

## 🛠️ Инструменты (установка)

### Windows (PowerShell от имени администратора)

```powershell
# 1. Установка protoc (Protocol Buffer Compiler)
winget install Google.Protobuf

# Или скачайте вручную:
# https://github.com/protocolbuffers/protobuf/releases
# Распакуйте в C:\protoc и добавьте C:\protoc\bin в PATH

# 2. Установка Go плагинов для protoc
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 3. Проверка
protoc --version
# libprotoc 25.x

# Проверка что плагины в PATH
where protoc-gen-go
where protoc-gen-go-grpc
```

### Если `protoc-gen-go` не найден

Добавьте Go bin в PATH:

```powershell
# Временно (для текущей сессии)
$env:PATH += ";$env:USERPROFILE\go\bin"

# Постоянно (выполните и перезапустите терминал)
[Environment]::SetEnvironmentVariable("PATH", $env:PATH + ";$env:USERPROFILE\go\bin", "User")
```

---

## 📁 Структура Shared модуля

```
shared/
├── proto/                      # Исходные .proto файлы
│   ├── common/v1/
│   │   └── common.proto        # Общие типы (Pagination, Timestamp, etc.)
│   ├── auth/v1/
│   │   └── auth.proto          # Auth Service API
│   ├── user/v1/
│   │   └── user.proto          # User Service API
│   ├── workspace/v1/
│   │   └── workspace.proto
│   ├── floor_plan/v1/
│   │   └── floor_plan.proto
│   ├── scene/v1/
│   │   └── scene.proto
│   ├── branch/v1/
│   │   └── branch.proto
│   ├── ai/v1/
│   │   └── ai.proto            # С streaming для чата
│   ├── compliance/v1/
│   │   └── compliance.proto
│   ├── request/v1/
│   │   └── request.proto
│   └── notification/v1/
│       └── notification.proto
├── gen/                        # Сгенерированный Go код (НЕ редактировать!)
│   ├── common/v1/
│   │   └── common.pb.go
│   ├── auth/v1/
│   │   ├── auth.pb.go
│   │   └── auth_grpc.pb.go
│   └── ...
├── pkg/                        # Общие Go пакеты
│   ├── logger/
│   │   └── logger.go
│   ├── errors/
│   │   └── errors.go
│   ├── config/
│   │   └── config.go
│   └── grpc/
│       ├── server.go
│       └── client.go
└── go.mod
```

---

## 🚀 Час 0-1: Работа над Shared (СОВМЕСТНО)

### Кто что делает

| Developer 1 (Core) | Developer 2 (AI/3D) |
|--------------------|---------------------|
| `common.proto` | `compliance.proto` |
| `auth.proto` | `ai.proto` (со streaming) |
| `user.proto` | `scene.proto` |
| `workspace.proto` | `branch.proto` |
| `request.proto` | `floor_plan.proto` |
| `notification.proto` | — |
| `shared/pkg/logger` | `shared/pkg/grpc` |
| `shared/pkg/errors` | — |
| `shared/pkg/config` | — |
| `shared/go.mod` | — |

---

### Шаг 1: Developer 1 — Базовая настройка (первые 10 минут)

```powershell
# Убедитесь что вы в ветке dev/shared
git checkout dev/shared
git pull origin dev/shared

# Создайте shared/go.mod
cd shared
go mod init github.com/granula/shared
```

Создайте файл `shared/go.mod`:

```go
module github.com/granula/shared

go 1.22

require (
    google.golang.org/grpc v1.60.0
    google.golang.org/protobuf v1.32.0
    go.uber.org/zap v1.26.0
    github.com/spf13/viper v1.18.2
)
```

```powershell
# Скачайте зависимости
go mod tidy

# Коммит
cd ..
git add shared/go.mod shared/go.sum
git commit -m "feat(shared): initialize go module"
git push origin dev/shared
```

---

### Шаг 2: Developer 1 — common.proto

Создайте `shared/proto/common/v1/common.proto`:

```protobuf
syntax = "proto3";

package common.v1;

option go_package = "github.com/granula/shared/gen/common/v1;commonv1";

import "google/protobuf/timestamp.proto";

// Пагинация для списков
message PaginationRequest {
  int32 page = 1;       // Номер страницы (начиная с 1)
  int32 page_size = 2;  // Элементов на страницу (макс 100)
}

message PaginationResponse {
  int32 total = 1;       // Всего элементов
  int32 page = 2;        // Текущая страница
  int32 page_size = 3;   // Размер страницы
  int32 total_pages = 4; // Всего страниц
}

// UUID wrapper
message UUID {
  string value = 1;
}

// Стандартный ответ об ошибке
message Error {
  string code = 1;      // Код ошибки (например, "VALIDATION_ERROR")
  string message = 2;   // Человекочитаемое сообщение
  map<string, string> details = 3; // Детали (например, какие поля невалидны)
}

// Пустой запрос/ответ
message Empty {}
```

---

### Шаг 3: Developer 1 — auth.proto

Создайте `shared/proto/auth/v1/auth.proto`:

```protobuf
syntax = "proto3";

package auth.v1;

option go_package = "github.com/granula/shared/gen/auth/v1;authv1";

import "common/v1/common.proto";
import "google/protobuf/timestamp.proto";

// ============================================================================
// Auth Service
// ============================================================================

service AuthService {
  // Регистрация нового пользователя
  rpc Register(RegisterRequest) returns (RegisterResponse);
  
  // Вход по email/password
  rpc Login(LoginRequest) returns (LoginResponse);
  
  // Вход через OAuth (Google/Yandex)
  rpc OAuthLogin(OAuthLoginRequest) returns (LoginResponse);
  
  // Валидация JWT токена (используется API Gateway)
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  
  // Обновление access токена
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
  
  // Выход (инвалидация токенов)
  rpc Logout(LogoutRequest) returns (common.v1.Empty);
  
  // Запрос сброса пароля
  rpc RequestPasswordReset(RequestPasswordResetRequest) returns (common.v1.Empty);
  
  // Сброс пароля
  rpc ResetPassword(ResetPasswordRequest) returns (common.v1.Empty);
  
  // Подтверждение email
  rpc VerifyEmail(VerifyEmailRequest) returns (common.v1.Empty);
}

// ============================================================================
// Messages
// ============================================================================

message RegisterRequest {
  string email = 1;
  string password = 2;
  string name = 3;
}

message RegisterResponse {
  string user_id = 1;
  string access_token = 2;
  string refresh_token = 3;
  google.protobuf.Timestamp expires_at = 4;
}

message LoginRequest {
  string email = 1;
  string password = 2;
}

message LoginResponse {
  string user_id = 1;
  string access_token = 2;
  string refresh_token = 3;
  google.protobuf.Timestamp expires_at = 4;
}

message OAuthLoginRequest {
  string provider = 1;  // "google" или "yandex"
  string code = 2;      // Authorization code от OAuth провайдера
  string redirect_uri = 3;
}

message ValidateTokenRequest {
  string access_token = 1;
}

message ValidateTokenResponse {
  bool valid = 1;
  string user_id = 2;
  string role = 3;  // "user", "admin", "expert"
}

message RefreshTokenRequest {
  string refresh_token = 1;
}

message RefreshTokenResponse {
  string access_token = 1;
  string refresh_token = 2;
  google.protobuf.Timestamp expires_at = 3;
}

message LogoutRequest {
  string refresh_token = 1;
}

message RequestPasswordResetRequest {
  string email = 1;
}

message ResetPasswordRequest {
  string token = 1;
  string new_password = 2;
}

message VerifyEmailRequest {
  string token = 1;
}
```

---

### Шаг 4: Developer 2 — compliance.proto (параллельно)

Developer 2 создаёт `shared/proto/compliance/v1/compliance.proto`:

```protobuf
syntax = "proto3";

package compliance.v1;

option go_package = "github.com/granula/shared/gen/compliance/v1;compliancev1";

import "common/v1/common.proto";

// ============================================================================
// Compliance Service — проверка норм СНиП и ЖК РФ
// ============================================================================

service ComplianceService {
  // Полная проверка сцены на соответствие нормам
  rpc CheckCompliance(CheckComplianceRequest) returns (CheckComplianceResponse);
  
  // Проверка конкретной операции (перед применением)
  rpc CheckOperation(CheckOperationRequest) returns (CheckOperationResponse);
  
  // Получить список всех правил
  rpc GetRules(GetRulesRequest) returns (GetRulesResponse);
  
  // Получить детали правила
  rpc GetRule(GetRuleRequest) returns (Rule);
  
  // Сгенерировать отчёт о соответствии
  rpc GenerateReport(GenerateReportRequest) returns (GenerateReportResponse);
}

// ============================================================================
// Messages
// ============================================================================

message CheckComplianceRequest {
  string scene_id = 1;
}

message CheckComplianceResponse {
  bool compliant = 1;              // Соответствует ли нормам
  repeated Violation violations = 2; // Список нарушений
  ComplianceStats stats = 3;       // Статистика
}

message CheckOperationRequest {
  string scene_id = 1;
  Operation operation = 2;
}

message CheckOperationResponse {
  bool allowed = 1;                 // Можно ли выполнить операцию
  repeated Violation violations = 2; // Что будет нарушено
  repeated string warnings = 3;     // Предупреждения (не критичные)
}

message Operation {
  string type = 1;  // "DEMOLISH_WALL", "ADD_WALL", "MOVE_WET_ZONE", etc.
  string element_id = 2;
  map<string, string> params = 3;
}

message Violation {
  string rule_id = 1;
  string rule_code = 2;       // Например "СНиП 31-01-2003 п.9.22"
  string severity = 3;        // "ERROR", "WARNING"
  string message = 4;         // Человекочитаемое описание
  string element_id = 5;      // Какой элемент нарушает
  string suggestion = 6;      // Как исправить
}

message ComplianceStats {
  int32 total_rules_checked = 1;
  int32 violations_count = 2;
  int32 warnings_count = 3;
}

message GetRulesRequest {
  string category = 1;  // Фильтр по категории (опционально)
  common.v1.PaginationRequest pagination = 2;
}

message GetRulesResponse {
  repeated Rule rules = 1;
  common.v1.PaginationResponse pagination = 2;
}

message GetRuleRequest {
  string rule_id = 1;
}

message Rule {
  string id = 1;
  string code = 2;            // "СНиП 31-01-2003 п.9.22"
  string category = 3;        // "load_bearing", "wet_zones", "fire_safety"
  string name = 4;
  string description = 5;
  string severity = 6;        // "ERROR", "WARNING"
  bool active = 7;
}

message GenerateReportRequest {
  string scene_id = 1;
  string format = 2;  // "PDF", "JSON"
}

message GenerateReportResponse {
  bytes report = 1;       // Файл отчёта
  string filename = 2;
  string content_type = 3;
}
```

---

### Шаг 5: Developer 2 — ai.proto (со streaming)

```protobuf
syntax = "proto3";

package ai.v1;

option go_package = "github.com/granula/shared/gen/ai/v1;aiv1";

import "google/protobuf/timestamp.proto";

// ============================================================================
// AI Service — распознавание, генерация, чат
// ============================================================================

service AIService {
  // Распознавание планировки из изображения
  rpc RecognizeFloorPlan(RecognizeFloorPlanRequest) returns (RecognizeFloorPlanResponse);
  
  // Генерация вариантов планировки
  rpc GenerateVariants(GenerateVariantsRequest) returns (GenerateVariantsResponse);
  
  // Отправка сообщения в чат (без streaming)
  rpc SendChatMessage(ChatMessageRequest) returns (ChatMessageResponse);
  
  // Streaming ответ чата (Server-side streaming)
  rpc StreamChatResponse(ChatMessageRequest) returns (stream ChatChunk);
  
  // Получить историю чата
  rpc GetChatHistory(GetChatHistoryRequest) returns (GetChatHistoryResponse);
  
  // Очистить историю чата
  rpc ClearChatHistory(ClearChatHistoryRequest) returns (ClearChatHistoryResponse);
}

// ============================================================================
// Recognition
// ============================================================================

message RecognizeFloorPlanRequest {
  string floor_plan_id = 1;
  bytes image = 2;              // Изображение планировки
  string image_type = 3;        // "jpeg", "png", "pdf"
  RecognitionOptions options = 4;
}

message RecognitionOptions {
  bool detect_load_bearing = 1;  // Определять несущие стены
  bool detect_wet_zones = 2;     // Определять мокрые зоны
  bool detect_furniture = 3;     // Определять мебель
  float scale = 4;               // Масштаб (пикселей на метр)
}

message RecognizeFloorPlanResponse {
  bool success = 1;
  RecognizedScene scene = 2;
  float confidence = 3;          // Уверенность распознавания (0-1)
  repeated string warnings = 4;
}

message RecognizedScene {
  repeated RecognizedWall walls = 1;
  repeated RecognizedRoom rooms = 2;
  repeated RecognizedElement elements = 3;
  Dimensions dimensions = 4;
}

message RecognizedWall {
  string id = 1;
  Point start = 2;
  Point end = 3;
  float thickness = 4;
  bool is_load_bearing = 5;
  float confidence = 6;
}

message RecognizedRoom {
  string id = 1;
  string type = 2;  // "living", "bedroom", "kitchen", "bathroom", etc.
  repeated Point polygon = 3;
  float area = 4;
  float confidence = 5;
}

message RecognizedElement {
  string id = 1;
  string type = 2;  // "door", "window", "sink", "toilet", etc.
  Point position = 3;
  Dimensions size = 4;
  float rotation = 5;
  float confidence = 6;
}

message Point {
  float x = 1;
  float y = 2;
}

message Dimensions {
  float width = 1;
  float height = 2;
  float depth = 3;  // Для 3D
}

// ============================================================================
// Generation
// ============================================================================

message GenerateVariantsRequest {
  string scene_id = 1;
  string branch_id = 2;
  string prompt = 3;          // Описание желаемых изменений
  int32 variants_count = 4;   // Сколько вариантов сгенерировать (1-5)
  GenerationOptions options = 5;
}

message GenerationOptions {
  bool preserve_load_bearing = 1;  // Не трогать несущие стены
  bool check_compliance = 2;       // Проверять нормы
  repeated string room_types = 3;  // Какие комнаты должны быть
}

message GenerateVariantsResponse {
  repeated GeneratedVariant variants = 1;
}

message GeneratedVariant {
  string id = 1;
  string branch_id = 2;       // ID созданной ветки
  string description = 3;     // Описание варианта
  float score = 4;            // Оценка варианта (0-1)
  repeated string changes = 5; // Список изменений
}

// ============================================================================
// Chat
// ============================================================================

message ChatMessageRequest {
  string scene_id = 1;
  string branch_id = 2;
  string message = 3;
  string context_id = 4;  // ID контекста (для продолжения разговора)
}

message ChatMessageResponse {
  string message_id = 1;
  string response = 2;
  string context_id = 3;
  repeated SuggestedAction actions = 4;  // Предложенные действия
}

// Чанк для streaming
message ChatChunk {
  string content = 1;         // Часть ответа
  bool is_final = 2;          // Это последний чанк?
  string message_id = 3;      // ID сообщения (в первом чанке)
  SuggestedAction action = 4; // Предложенное действие (в последнем чанке)
}

message SuggestedAction {
  string type = 1;     // "DEMOLISH_WALL", "ADD_FURNITURE", etc.
  string description = 2;
  map<string, string> params = 3;
}

message GetChatHistoryRequest {
  string scene_id = 1;
  string branch_id = 2;
  int32 limit = 3;
}

message GetChatHistoryResponse {
  repeated ChatMessage messages = 1;
}

message ChatMessage {
  string id = 1;
  string role = 2;  // "user" или "assistant"
  string content = 3;
  google.protobuf.Timestamp created_at = 4;
}

message ClearChatHistoryRequest {
  string scene_id = 1;
  string branch_id = 2;
}

message ClearChatHistoryResponse {
  int32 deleted_count = 1;
}
```

---

### Шаг 6: Генерация Go кода

После создания proto файлов, **один из разработчиков** генерирует код:

```powershell
# Перейти в корень проекта
cd R:\granula\api

# Создать папку для сгенерированного кода
New-Item -ItemType Directory -Force -Path shared/gen

# Сгенерировать Go код для всех proto файлов
# Windows PowerShell:

# Common
protoc --go_out=shared/gen --go_opt=paths=source_relative `
       --go-grpc_out=shared/gen --go-grpc_opt=paths=source_relative `
       -I shared/proto `
       shared/proto/common/v1/common.proto

# Auth
protoc --go_out=shared/gen --go_opt=paths=source_relative `
       --go-grpc_out=shared/gen --go-grpc_opt=paths=source_relative `
       -I shared/proto `
       shared/proto/auth/v1/auth.proto

# Compliance
protoc --go_out=shared/gen --go_opt=paths=source_relative `
       --go-grpc_out=shared/gen --go-grpc_opt=paths=source_relative `
       -I shared/proto `
       shared/proto/compliance/v1/compliance.proto

# AI
protoc --go_out=shared/gen --go_opt=paths=source_relative `
       --go-grpc_out=shared/gen --go-grpc_opt=paths=source_relative `
       -I shared/proto `
       shared/proto/ai/v1/ai.proto
```

**Или используйте команду из Makefile:**

```powershell
# Если make установлен
make proto
```

---

### Шаг 7: Коммит и синхронизация

```powershell
# Developer 1 делает коммит своих proto
git add shared/proto/common shared/proto/auth shared/proto/user
git commit -m "feat(shared): add common, auth, user proto files"
git push origin dev/shared

# Developer 2 делает коммит своих proto
git add shared/proto/compliance shared/proto/ai shared/proto/scene shared/proto/branch
git commit -m "feat(shared): add compliance, ai, scene, branch proto files"
git push origin dev/shared

# Если конфликт — один пулит, резолвит, пушит
git pull origin dev/shared
# ... resolve conflicts ...
git add .
git commit -m "merge: resolve proto conflicts"
git push origin dev/shared
```

---

## 📋 Час 1+: Расходимся по своим сервисам

### Developer 1: Начало работы над Auth Service

```powershell
# Убедитесь что shared готов
git checkout dev/shared
git pull origin dev/shared

# Создайте свою ветку
git checkout -b dev/d1-auth-service

# Инициализируйте сервис
cd auth-service
go mod init github.com/granula/auth-service

# Добавьте зависимость на shared
go mod edit -replace github.com/granula/shared=../shared
go mod tidy

# Создайте базовую структуру
# (уже создана скриптом init-project.ps1)

# Начинайте писать код...
# auth-service/cmd/server/main.go
# auth-service/internal/config/config.go
# и т.д.

# Периодически коммитьте
git add .
git commit -m "feat(auth): implement user registration"
git push origin dev/d1-auth-service
```

### Developer 2: Начало работы над Compliance Service

```powershell
# Убедитесь что shared готов
git checkout dev/shared
git pull origin dev/shared

# Создайте свою ветку
git checkout -b dev/d2-compliance-service

# Инициализируйте сервис
cd compliance-service
go mod init github.com/granula/compliance-service

# Добавьте зависимость на shared
go mod edit -replace github.com/granula/shared=../shared
go mod tidy

# Начинайте писать код...

# Периодически коммитьте
git add .
git commit -m "feat(compliance): add SNiP rules engine"
git push origin dev/d2-compliance-service
```

---

## 🔄 Когда нужны изменения в Shared

Если Developer 2 нужно добавить новое поле в proto:

```powershell
# 1. Переключиться на dev/shared
git checkout dev/shared
git pull origin dev/shared

# 2. Внести изменения в proto файлы
# Отредактировать shared/proto/...

# 3. Перегенерировать Go код
make proto

# 4. Закоммитить
git add shared/
git commit -m "feat(shared): add new field to scene.proto"
git push origin dev/shared

# 5. Вернуться в свою ветку и подтянуть изменения
git checkout dev/d2-scene-service
git merge origin/dev/shared
```

---

## 📊 Чеклист первого часа

### Developer 1 (Core)

- [ ] Создать `shared/go.mod`
- [ ] Создать `shared/proto/common/v1/common.proto`
- [ ] Создать `shared/proto/auth/v1/auth.proto`
- [ ] Создать `shared/proto/user/v1/user.proto`
- [ ] Создать `shared/proto/workspace/v1/workspace.proto`
- [ ] Создать `shared/proto/request/v1/request.proto`
- [ ] Создать `shared/proto/notification/v1/notification.proto`
- [ ] Создать `shared/pkg/logger/logger.go`
- [ ] Создать `shared/pkg/errors/errors.go`
- [ ] Создать `shared/pkg/config/config.go`
- [ ] Сгенерировать Go код (`make proto`)

### Developer 2 (AI/3D)

- [ ] Создать `shared/proto/compliance/v1/compliance.proto`
- [ ] Создать `shared/proto/ai/v1/ai.proto` (со streaming)
- [ ] Создать `shared/proto/scene/v1/scene.proto`
- [ ] Создать `shared/proto/branch/v1/branch.proto`
- [ ] Создать `shared/proto/floor_plan/v1/floor_plan.proto`
- [ ] Создать `shared/pkg/grpc/server.go`
- [ ] Создать `shared/pkg/grpc/client.go`

---

## 🔗 Полезные ссылки

- [Protocol Buffers Documentation](https://protobuf.dev/)
- [gRPC Go Tutorial](https://grpc.io/docs/languages/go/quickstart/)
- [Go Project Layout](https://github.com/golang-standards/project-layout)

---

## ❓ FAQ

### Q: Что если я изменил proto и забыл перегенерировать?

Go компилятор выдаст ошибки о несуществующих типах. Запустите `make proto`.

### Q: Как тестировать gRPC без клиента?

Используйте [grpcurl](https://github.com/fullstorydev/grpcurl) или [Postman](https://www.postman.com/) (поддерживает gRPC).

### Q: Как отлаживать gRPC?

Включите логирование в interceptors и используйте `GRPC_GO_LOG_SEVERITY_LEVEL=info`.


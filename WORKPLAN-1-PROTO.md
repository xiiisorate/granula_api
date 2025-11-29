# 📦 WORKPLAN-1: Proto файлы и генерация кода

> **Приоритет:** 🔴 БЛОКИРУЮЩИЙ — без этого ничего не скомпилируется  
> **Время:** 1-2 часа  
> **Зависимости:** Нет  
> **Результат:** Все сервисы компилируются

---

## 🎯 ЦЕЛЬ

1. Исправить `go_package` во всех proto файлах
2. Сгенерировать Go код из proto в `shared/gen/`
3. Выполнить `go mod tidy` во всех сервисах
4. Убедиться, что все сервисы компилируются

---

## 📋 ПРОБЛЕМА

### Текущее состояние
Папка `shared/gen/` **ПУСТАЯ** — Go код из proto файлов не сгенерирован.

### Почему это критично
Все сервисы импортируют код из `shared/gen/...`:
```go
// Пример из auth-service/internal/grpc/server.go
import (
    pb "github.com/xiiisorate/granula_api/shared/gen/auth/v1"
)
```

Без сгенерированного кода **НИ ОДИН СЕРВИС НЕ СКОМПИЛИРУЕТСЯ**.

### Вторая проблема: неправильные пути
Proto файлы содержат неправильный `go_package`:
```protobuf
// СЕЙЧАС (неправильно):
option go_package = "github.com/granula/shared/gen/auth/v1;authv1";

// НУЖНО:
option go_package = "github.com/xiiisorate/granula_api/shared/gen/auth/v1;authv1";
```

---

## 📁 ФАЙЛЫ ДЛЯ ИЗМЕНЕНИЯ

### Proto файлы (11 штук)
| # | Файл | Текущий go_package | Нужный go_package |
|---|------|-------------------|-------------------|
| 1 | `shared/proto/common/v1/common.proto` | `github.com/granula/...` | `github.com/xiiisorate/granula_api/...` |
| 2 | `shared/proto/auth/v1/auth.proto` | `github.com/granula/...` | `github.com/xiiisorate/granula_api/...` |
| 3 | `shared/proto/user/v1/user.proto` | `github.com/granula/...` | `github.com/xiiisorate/granula_api/...` |
| 4 | `shared/proto/workspace/v1/workspace.proto` | `github.com/granula/...` | `github.com/xiiisorate/granula_api/...` |
| 5 | `shared/proto/scene/v1/scene.proto` | `github.com/granula/...` | `github.com/xiiisorate/granula_api/...` |
| 6 | `shared/proto/branch/v1/branch.proto` | `github.com/granula/...` | `github.com/xiiisorate/granula_api/...` |
| 7 | `shared/proto/ai/v1/ai.proto` | `github.com/granula/...` | `github.com/xiiisorate/granula_api/...` |
| 8 | `shared/proto/compliance/v1/compliance.proto` | `github.com/granula/...` | `github.com/xiiisorate/granula_api/...` |
| 9 | `shared/proto/floorplan/v1/floorplan.proto` | `github.com/granula/...` | `github.com/xiiisorate/granula_api/...` |
| 10 | `shared/proto/request/v1/request.proto` | `github.com/granula/...` | `github.com/xiiisorate/granula_api/...` |
| 11 | `shared/proto/notification/v1/notification.proto` | `github.com/granula/...` | `github.com/xiiisorate/granula_api/...` |

---

## 🔧 ПОШАГОВАЯ ИНСТРУКЦИЯ

### ШАГ 1: Установить инструменты protoc (если не установлены)

**Windows (PowerShell):**
```powershell
# Установить protoc через chocolatey
choco install protobuf -y

# Установить Go плагины
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Убедиться что $GOPATH/bin в PATH
$env:PATH += ";$env:GOPATH\bin"
```

**Проверка:**
```powershell
protoc --version
# libprotoc 3.x.x

protoc-gen-go --version
# protoc-gen-go v1.x.x
```

---

### ШАГ 2: Исправить go_package в proto файлах

#### 2.1. common.proto
**Файл:** `shared/proto/common/v1/common.proto`

**Найти и заменить:**
```protobuf
// БЫЛО:
option go_package = "github.com/granula/shared/gen/common/v1;commonv1";

// СТАЛО:
option go_package = "github.com/xiiisorate/granula_api/shared/gen/common/v1;commonv1";
```

#### 2.2. auth.proto
**Файл:** `shared/proto/auth/v1/auth.proto`

```protobuf
// БЫЛО:
option go_package = "github.com/granula/shared/gen/auth/v1;authv1";

// СТАЛО:
option go_package = "github.com/xiiisorate/granula_api/shared/gen/auth/v1;authv1";
```

#### 2.3. user.proto
**Файл:** `shared/proto/user/v1/user.proto`

```protobuf
// БЫЛО:
option go_package = "github.com/granula/shared/gen/user/v1;userv1";

// СТАЛО:
option go_package = "github.com/xiiisorate/granula_api/shared/gen/user/v1;userv1";
```

#### 2.4. workspace.proto
**Файл:** `shared/proto/workspace/v1/workspace.proto`

```protobuf
// БЫЛО:
option go_package = "github.com/granula/shared/gen/workspace/v1;workspacev1";

// СТАЛО:
option go_package = "github.com/xiiisorate/granula_api/shared/gen/workspace/v1;workspacev1";
```

#### 2.5. scene.proto
**Файл:** `shared/proto/scene/v1/scene.proto`

```protobuf
// БЫЛО:
option go_package = "github.com/granula/shared/gen/scene/v1;scenev1";

// СТАЛО:
option go_package = "github.com/xiiisorate/granula_api/shared/gen/scene/v1;scenev1";
```

#### 2.6. branch.proto
**Файл:** `shared/proto/branch/v1/branch.proto`

```protobuf
// БЫЛО:
option go_package = "github.com/granula/shared/gen/branch/v1;branchv1";

// СТАЛО:
option go_package = "github.com/xiiisorate/granula_api/shared/gen/branch/v1;branchv1";
```

#### 2.7. ai.proto
**Файл:** `shared/proto/ai/v1/ai.proto`

```protobuf
// БЫЛО:
option go_package = "github.com/granula/shared/gen/ai/v1;aiv1";

// СТАЛО:
option go_package = "github.com/xiiisorate/granula_api/shared/gen/ai/v1;aiv1";
```

#### 2.8. compliance.proto
**Файл:** `shared/proto/compliance/v1/compliance.proto`

```protobuf
// БЫЛО:
option go_package = "github.com/granula/shared/gen/compliance/v1;compliancev1";

// СТАЛО:
option go_package = "github.com/xiiisorate/granula_api/shared/gen/compliance/v1;compliancev1";
```

#### 2.9. floorplan.proto
**Файл:** `shared/proto/floorplan/v1/floorplan.proto`

```protobuf
// БЫЛО:
option go_package = "github.com/granula/shared/gen/floorplan/v1;floorplanv1";

// СТАЛО:
option go_package = "github.com/xiiisorate/granula_api/shared/gen/floorplan/v1;floorplanv1";
```

#### 2.10. request.proto
**Файл:** `shared/proto/request/v1/request.proto`

```protobuf
// БЫЛО:
option go_package = "github.com/granula/shared/gen/request/v1;requestv1";

// СТАЛО:
option go_package = "github.com/xiiisorate/granula_api/shared/gen/request/v1;requestv1";
```

#### 2.11. notification.proto
**Файл:** `shared/proto/notification/v1/notification.proto`

```protobuf
// БЫЛО:
option go_package = "github.com/granula/shared/gen/notification/v1;notificationv1";

// СТАЛО:
option go_package = "github.com/xiiisorate/granula_api/shared/gen/notification/v1;notificationv1";
```

---

### ШАГ 3: Создать папки для генерации

```powershell
cd shared

# Создать структуру папок
New-Item -ItemType Directory -Force -Path gen/common/v1
New-Item -ItemType Directory -Force -Path gen/auth/v1
New-Item -ItemType Directory -Force -Path gen/user/v1
New-Item -ItemType Directory -Force -Path gen/workspace/v1
New-Item -ItemType Directory -Force -Path gen/scene/v1
New-Item -ItemType Directory -Force -Path gen/branch/v1
New-Item -ItemType Directory -Force -Path gen/ai/v1
New-Item -ItemType Directory -Force -Path gen/compliance/v1
New-Item -ItemType Directory -Force -Path gen/floorplan/v1
New-Item -ItemType Directory -Force -Path gen/request/v1
New-Item -ItemType Directory -Force -Path gen/notification/v1
```

---

### ШАГ 4: Сгенерировать Go код из proto

**PowerShell скрипт (сохранить как `shared/scripts/generate-proto.ps1`):**

```powershell
#!/usr/bin/env pwsh

$ErrorActionPreference = "Stop"

$PROTO_DIR = "$PSScriptRoot/../proto"
$GEN_DIR = "$PSScriptRoot/../gen"

Write-Host "Generating proto files..." -ForegroundColor Cyan

# Список proto файлов в правильном порядке (common первый - от него зависят другие)
$protos = @(
    "common/v1/common.proto",
    "auth/v1/auth.proto",
    "user/v1/user.proto",
    "workspace/v1/workspace.proto",
    "scene/v1/scene.proto",
    "branch/v1/branch.proto",
    "ai/v1/ai.proto",
    "compliance/v1/compliance.proto",
    "floorplan/v1/floorplan.proto",
    "request/v1/request.proto",
    "notification/v1/notification.proto"
)

foreach ($proto in $protos) {
    Write-Host "  Generating $proto..." -ForegroundColor Yellow
    
    protoc --proto_path="$PROTO_DIR" `
           --go_out="$GEN_DIR" --go_opt=paths=source_relative `
           --go-grpc_out="$GEN_DIR" --go-grpc_opt=paths=source_relative `
           "$PROTO_DIR/$proto"
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  ERROR generating $proto" -ForegroundColor Red
        exit 1
    }
}

Write-Host "Proto generation complete!" -ForegroundColor Green
```

**Запуск:**
```powershell
cd shared
.\scripts\generate-proto.ps1
```

**Или вручную (одной командой):**
```powershell
cd shared

protoc --proto_path=proto `
  --go_out=gen --go_opt=paths=source_relative `
  --go-grpc_out=gen --go-grpc_opt=paths=source_relative `
  proto/common/v1/common.proto `
  proto/auth/v1/auth.proto `
  proto/user/v1/user.proto `
  proto/workspace/v1/workspace.proto `
  proto/scene/v1/scene.proto `
  proto/branch/v1/branch.proto `
  proto/ai/v1/ai.proto `
  proto/compliance/v1/compliance.proto `
  proto/floorplan/v1/floorplan.proto `
  proto/request/v1/request.proto `
  proto/notification/v1/notification.proto
```

---

### ШАГ 5: Проверить сгенерированные файлы

```powershell
# Проверить что файлы созданы
Get-ChildItem -Recurse shared/gen -Filter "*.go" | Select-Object FullName

# Ожидаемый результат:
# shared/gen/common/v1/common.pb.go
# shared/gen/common/v1/common_grpc.pb.go
# shared/gen/auth/v1/auth.pb.go
# shared/gen/auth/v1/auth_grpc.pb.go
# ... и т.д.
```

---

### ШАГ 6: Выполнить go mod tidy во всех сервисах

```powershell
# Из корня проекта
cd R:\granula\api

# Список всех сервисов
$services = @(
    "shared",
    "api-gateway",
    "auth-service",
    "user-service",
    "workspace-service",
    "scene-service",
    "branch-service",
    "ai-service",
    "compliance-service",
    "floorplan-service",
    "request-service",
    "notification-service"
)

foreach ($svc in $services) {
    Write-Host "Running go mod tidy in $svc..." -ForegroundColor Cyan
    Push-Location $svc
    go mod tidy
    if ($LASTEXITCODE -ne 0) {
        Write-Host "ERROR in $svc" -ForegroundColor Red
    }
    Pop-Location
}
```

---

### ШАГ 7: Проверить компиляцию всех сервисов

```powershell
$services = @(
    "api-gateway",
    "auth-service",
    "user-service",
    "workspace-service",
    "scene-service",
    "branch-service",
    "ai-service",
    "compliance-service",
    "floorplan-service",
    "request-service",
    "notification-service"
)

foreach ($svc in $services) {
    Write-Host "Building $svc..." -ForegroundColor Cyan
    Push-Location $svc
    go build ./...
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  OK" -ForegroundColor Green
    } else {
        Write-Host "  FAILED" -ForegroundColor Red
    }
    Pop-Location
}
```

---

## ✅ КРИТЕРИИ УСПЕХА

- [ ] Все 11 proto файлов содержат правильный `go_package`
- [ ] Папка `shared/gen/` содержит сгенерированные `.pb.go` и `_grpc.pb.go` файлы
- [ ] Команда `go mod tidy` выполняется без ошибок во всех сервисах
- [ ] Команда `go build ./...` выполняется без ошибок во всех сервисах

---

## 🐛 ВОЗМОЖНЫЕ ПРОБЛЕМЫ

### Проблема: "protoc-gen-go: program not found"
**Решение:**
```powershell
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
$env:PATH += ";$(go env GOPATH)\bin"
```

### Проблема: "import path does not begin with hostname"
**Решение:** Проверить что `go_package` начинается с `github.com/...`

### Проблема: "cannot find module providing package"
**Решение:** Проверить что `shared/go.mod` содержит правильный module path:
```go
module github.com/xiiisorate/granula_api/shared
```

---

## 📚 СВЯЗАННАЯ ДОКУМЕНТАЦИЯ

- Proto спецификация: `docs/QUICK-START.md` (секция "Proto файлы")
- Архитектура: `docs/architecture/microservices.md`
- Shared модуль: `shared/go.mod`

---

## ➡️ СЛЕДУЮЩИЙ ШАГ

После успешной компиляции всех сервисов, переходите к:
- [WORKPLAN-2-API-GATEWAY.md](./WORKPLAN-2-API-GATEWAY.md) — создание HTTP handlers
- [WORKPLAN-3-AI-MODULE.md](./WORKPLAN-3-AI-MODULE.md) — исправление AI модуля


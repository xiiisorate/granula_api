# 🌐 WORKPLAN-2: API Gateway Handlers

> **Приоритет:** 🔴 Высокий  
> **Время:** 3-4 часа  
> **Зависимости:** WORKPLAN-1-PROTO.md  
> **Результат:** Все HTTP endpoints работают

---

## 🎯 ЦЕЛЬ

Создать недостающие HTTP handlers в API Gateway для:
1. FloorPlan (загрузка и управление планами)
2. Branch (версионирование)
3. Compliance (проверка норм)
4. Request (заявки на экспертизу)

---

## 📋 ПРОБЛЕМА

### Текущее состояние
В файле `api-gateway/cmd/main.go` (строки 244-291) используются **placeholder handlers**:

```go
// ❌ СЕЙЧАС — это заглушки:
floorPlans.Get("/", placeholderHandler("List floor plans"))
floorPlans.Post("/", placeholderHandler("Upload floor plan"))
// ... и т.д.
```

### Какие endpoints не работают

| Группа | Endpoints | Статус |
|--------|-----------|--------|
| `/floor-plans/*` | GET, POST, PATCH, DELETE | ❌ Placeholder |
| `/scenes/:id/branches/*` | GET, POST, PATCH, DELETE | ❌ Placeholder |
| `/compliance/*` | POST check, GET rules | ❌ Placeholder |
| `/requests/*` | GET, POST, PATCH, DELETE | ❌ Placeholder |

---

## 📁 ФАЙЛЫ ДЛЯ СОЗДАНИЯ

| # | Файл | Описание | Документация |
|---|------|----------|--------------|
| 1 | `api-gateway/internal/handlers/floorplan.go` | FloorPlan HTTP handlers | `docs/api/floor-plans.md` |
| 2 | `api-gateway/internal/handlers/branch.go` | Branch HTTP handlers | `docs/api/branches.md` |
| 3 | `api-gateway/internal/handlers/compliance.go` | Compliance HTTP handlers | `docs/api/compliance.md` |
| 4 | `api-gateway/internal/handlers/request.go` | Request HTTP handlers | `docs/api/requests.md` |

---

## 🔧 ПОШАГОВАЯ ИНСТРУКЦИЯ

### ШАГ 1: Изучить существующие handlers (как образец)

**Референсные файлы:**
- `api-gateway/internal/handlers/ai.go` — хороший пример с gRPC клиентом
- `api-gateway/internal/handlers/workspace.go` — пример CRUD операций
- `api-gateway/internal/handlers/scene.go` — пример работы со сценами

**Ключевые паттерны:**
```go
// 1. Структура handler с gRPC клиентом
type AIHandler struct {
    client aipb.AIServiceClient
}

func NewAIHandler(conn *grpc.ClientConn) *AIHandler {
    return &AIHandler{
        client: aipb.NewAIServiceClient(conn),
    }
}

// 2. Метод handler с Swagger annotations
// @Summary Название
// @Description Описание
// @Tags tag
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ...
// @Success 200 {object} ResponseType
// @Failure 400 {object} ErrorResponse
// @Router /path [method]
func (h *Handler) MethodName(c *fiber.Ctx) error {
    // 1. Parse input
    // 2. Create context with timeout
    // 3. Call gRPC
    // 4. Handle errors
    // 5. Return JSON response
}
```

---

### ШАГ 2: Создать FloorPlan Handler

**Файл:** `api-gateway/internal/handlers/floorplan.go`

**Документация:** `docs/api/floor-plans.md`

**Endpoints для реализации:**

| Method | Path | gRPC Method | Описание |
|--------|------|-------------|----------|
| POST | `/floor-plans` | Upload | Загрузка плана |
| GET | `/floor-plans` | List | Список планов |
| GET | `/floor-plans/:id` | Get | Получить план |
| PATCH | `/floor-plans/:id` | Update | Обновить план |
| DELETE | `/floor-plans/:id` | Delete | Удалить план |
| POST | `/floor-plans/:id/recognize` | StartRecognition | Запустить распознавание |
| GET | `/floor-plans/:id/recognition-status` | GetRecognitionStatus | Статус распознавания |
| POST | `/floor-plans/:id/create-scene` | CreateSceneFromFloorPlan | Создать 3D сцену |

**Код handler:**

```go
// Package handlers provides HTTP handlers for API Gateway.
package handlers

import (
    "context"
    "io"
    "time"

    "github.com/gofiber/fiber/v2"
    "google.golang.org/grpc"

    floorplanpb "github.com/xiiisorate/granula_api/shared/gen/floorplan/v1"
)

// FloorPlanHandler handles floor plan HTTP requests.
type FloorPlanHandler struct {
    client floorplanpb.FloorPlanServiceClient
}

// NewFloorPlanHandler creates a new FloorPlanHandler.
func NewFloorPlanHandler(conn *grpc.ClientConn) *FloorPlanHandler {
    return &FloorPlanHandler{
        client: floorplanpb.NewFloorPlanServiceClient(conn),
    }
}

// Upload загружает новый план.
// @Summary Загрузить планировку
// @Description Загрузка изображения планировки (BTI, скан, фото)
// @Tags floor-plans
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Файл планировки"
// @Param workspace_id formData string true "ID воркспейса"
// @Param name formData string false "Название"
// @Param source_type formData string false "Тип источника (bti, scan, photo, sketch)"
// @Success 201 {object} FloorPlanResponse
// @Failure 400 {object} ErrorResponse
// @Router /floor-plans [post]
func (h *FloorPlanHandler) Upload(c *fiber.Ctx) error {
    // Get user ID from context (set by auth middleware)
    userID := c.Locals("user_id").(string)
    
    // Get workspace ID
    workspaceID := c.FormValue("workspace_id")
    if workspaceID == "" {
        return fiber.NewError(fiber.StatusBadRequest, "workspace_id is required")
    }
    
    // Get file
    fileHeader, err := c.FormFile("file")
    if err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "file is required")
    }
    
    file, err := fileHeader.Open()
    if err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "failed to open file")
    }
    defer file.Close()
    
    fileData, err := io.ReadAll(file)
    if err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "failed to read file")
    }
    
    ctx, cancel := context.WithTimeout(c.Context(), 60*time.Second)
    defer cancel()
    
    req := &floorplanpb.UploadFloorPlanRequest{
        WorkspaceId: workspaceID,
        UserId:      userID,
        FileName:    fileHeader.Filename,
        FileData:    fileData,
        ContentType: fileHeader.Header.Get("Content-Type"),
        Name:        c.FormValue("name", fileHeader.Filename),
        SourceType:  c.FormValue("source_type", "scan"),
    }
    
    resp, err := h.client.Upload(ctx, req)
    if err != nil {
        return handleGRPCError(err)
    }
    
    return c.Status(fiber.StatusCreated).JSON(fiber.Map{
        "data":       floorPlanToMap(resp.FloorPlan),
        "request_id": c.GetRespHeader("X-Request-ID"),
    })
}

// List возвращает список планов.
// @Summary Список планировок
// @Description Получить список планировок воркспейса
// @Tags floor-plans
// @Produce json
// @Security BearerAuth
// @Param workspace_id query string true "ID воркспейса"
// @Param limit query int false "Лимит" default(20)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {object} FloorPlansListResponse
// @Router /floor-plans [get]
func (h *FloorPlanHandler) List(c *fiber.Ctx) error {
    workspaceID := c.Query("workspace_id")
    if workspaceID == "" {
        return fiber.NewError(fiber.StatusBadRequest, "workspace_id is required")
    }
    
    ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
    defer cancel()
    
    resp, err := h.client.List(ctx, &floorplanpb.ListFloorPlansRequest{
        WorkspaceId: workspaceID,
        Limit:       int32(c.QueryInt("limit", 20)),
        Offset:      int32(c.QueryInt("offset", 0)),
    })
    if err != nil {
        return handleGRPCError(err)
    }
    
    items := make([]fiber.Map, 0, len(resp.FloorPlans))
    for _, fp := range resp.FloorPlans {
        items = append(items, floorPlanToMap(fp))
    }
    
    return c.JSON(fiber.Map{
        "data": fiber.Map{
            "items": items,
            "total": resp.Total,
        },
        "request_id": c.GetRespHeader("X-Request-ID"),
    })
}

// Get возвращает план по ID.
// @Summary Получить планировку
// @Description Получить детали планировки по ID
// @Tags floor-plans
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID планировки"
// @Success 200 {object} FloorPlanResponse
// @Failure 404 {object} ErrorResponse
// @Router /floor-plans/{id} [get]
func (h *FloorPlanHandler) Get(c *fiber.Ctx) error {
    id := c.Params("id")
    
    ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
    defer cancel()
    
    resp, err := h.client.Get(ctx, &floorplanpb.GetFloorPlanRequest{Id: id})
    if err != nil {
        return handleGRPCError(err)
    }
    
    return c.JSON(fiber.Map{
        "data":       floorPlanToMap(resp.FloorPlan),
        "request_id": c.GetRespHeader("X-Request-ID"),
    })
}

// Update обновляет план.
// @Summary Обновить планировку
// @Description Обновить метаданные планировки
// @Tags floor-plans
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID планировки"
// @Param body body UpdateFloorPlanInput true "Данные для обновления"
// @Success 200 {object} FloorPlanResponse
// @Router /floor-plans/{id} [patch]
func (h *FloorPlanHandler) Update(c *fiber.Ctx) error {
    id := c.Params("id")
    
    var input UpdateFloorPlanInput
    if err := c.BodyParser(&input); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
    }
    
    ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
    defer cancel()
    
    resp, err := h.client.Update(ctx, &floorplanpb.UpdateFloorPlanRequest{
        Id:   id,
        Name: input.Name,
    })
    if err != nil {
        return handleGRPCError(err)
    }
    
    return c.JSON(fiber.Map{
        "data":       floorPlanToMap(resp.FloorPlan),
        "request_id": c.GetRespHeader("X-Request-ID"),
    })
}

// Delete удаляет план.
// @Summary Удалить планировку
// @Description Удалить планировку
// @Tags floor-plans
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID планировки"
// @Success 200 {object} SuccessResponse
// @Router /floor-plans/{id} [delete]
func (h *FloorPlanHandler) Delete(c *fiber.Ctx) error {
    id := c.Params("id")
    
    ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
    defer cancel()
    
    _, err := h.client.Delete(ctx, &floorplanpb.DeleteFloorPlanRequest{Id: id})
    if err != nil {
        return handleGRPCError(err)
    }
    
    return c.JSON(fiber.Map{
        "message":    "Floor plan deleted",
        "request_id": c.GetRespHeader("X-Request-ID"),
    })
}

// StartRecognition запускает распознавание.
// @Summary Запустить распознавание
// @Description Запустить AI распознавание планировки
// @Tags floor-plans
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID планировки"
// @Param body body RecognitionOptionsInput false "Опции распознавания"
// @Success 200 {object} RecognitionJobResponse
// @Router /floor-plans/{id}/recognize [post]
func (h *FloorPlanHandler) StartRecognition(c *fiber.Ctx) error {
    id := c.Params("id")
    
    var input RecognitionOptionsInput
    c.BodyParser(&input) // Optional
    
    ctx, cancel := context.WithTimeout(c.Context(), 30*time.Second)
    defer cancel()
    
    resp, err := h.client.StartRecognition(ctx, &floorplanpb.StartRecognitionRequest{
        FloorPlanId: id,
        Options: &floorplanpb.RecognitionOptions{
            DetectLoadBearing: input.DetectLoadBearing,
            DetectWetZones:    input.DetectWetZones,
            DetectFurniture:   input.DetectFurniture,
        },
    })
    if err != nil {
        return handleGRPCError(err)
    }
    
    return c.JSON(fiber.Map{
        "data": fiber.Map{
            "job_id": resp.JobId,
            "status": resp.Status,
        },
        "request_id": c.GetRespHeader("X-Request-ID"),
    })
}

// GetRecognitionStatus возвращает статус распознавания.
// @Summary Статус распознавания
// @Description Получить статус задачи распознавания
// @Tags floor-plans
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID планировки"
// @Success 200 {object} RecognitionStatusResponse
// @Router /floor-plans/{id}/recognition-status [get]
func (h *FloorPlanHandler) GetRecognitionStatus(c *fiber.Ctx) error {
    id := c.Params("id")
    
    ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
    defer cancel()
    
    resp, err := h.client.GetRecognitionStatus(ctx, &floorplanpb.GetRecognitionStatusRequest{
        FloorPlanId: id,
    })
    if err != nil {
        return handleGRPCError(err)
    }
    
    return c.JSON(fiber.Map{
        "data": fiber.Map{
            "status":   resp.Status,
            "progress": resp.Progress,
            "result":   resp.Result,
            "error":    resp.Error,
        },
        "request_id": c.GetRespHeader("X-Request-ID"),
    })
}

// Helper functions

func floorPlanToMap(fp *floorplanpb.FloorPlan) fiber.Map {
    if fp == nil {
        return nil
    }
    return fiber.Map{
        "id":                fp.Id,
        "workspace_id":      fp.WorkspaceId,
        "user_id":           fp.UserId,
        "name":              fp.Name,
        "source_type":       fp.SourceType,
        "file_url":          fp.FileUrl,
        "thumbnail_url":     fp.ThumbnailUrl,
        "status":            fp.Status,
        "recognition_data":  fp.RecognitionData,
        "created_at":        fp.CreatedAt.AsTime(),
        "updated_at":        fp.UpdatedAt.AsTime(),
    }
}

// Input types

type UpdateFloorPlanInput struct {
    Name string `json:"name"`
}

type RecognitionOptionsInput struct {
    DetectLoadBearing bool `json:"detect_load_bearing"`
    DetectWetZones    bool `json:"detect_wet_zones"`
    DetectFurniture   bool `json:"detect_furniture"`
}
```

---

### ШАГ 3: Создать Branch Handler

**Файл:** `api-gateway/internal/handlers/branch.go`

**Документация:** `docs/api/branches.md`

**Endpoints для реализации:**

| Method | Path | gRPC Method | Описание |
|--------|------|-------------|----------|
| GET | `/scenes/:scene_id/branches` | List | Список веток |
| POST | `/scenes/:scene_id/branches` | Create | Создать ветку |
| GET | `/scenes/:scene_id/branches/:id` | Get | Получить ветку |
| PATCH | `/scenes/:scene_id/branches/:id` | Update | Обновить ветку |
| DELETE | `/scenes/:scene_id/branches/:id` | Delete | Удалить ветку |
| POST | `/scenes/:scene_id/branches/:id/activate` | Activate | Активировать ветку |
| POST | `/scenes/:scene_id/branches/:id/merge` | Merge | Слить ветки |
| GET | `/scenes/:scene_id/branches/:id/compare/:target_id` | Compare | Сравнить ветки |

**Структура (аналогично FloorPlanHandler):**

```go
package handlers

import (
    "context"
    "time"

    "github.com/gofiber/fiber/v2"
    "google.golang.org/grpc"

    branchpb "github.com/xiiisorate/granula_api/shared/gen/branch/v1"
)

type BranchHandler struct {
    client branchpb.BranchServiceClient
}

func NewBranchHandler(conn *grpc.ClientConn) *BranchHandler {
    return &BranchHandler{
        client: branchpb.NewBranchServiceClient(conn),
    }
}

// Реализовать методы: List, Create, Get, Update, Delete, Activate, Merge, Compare
// По аналогии с FloorPlanHandler
```

---

### ШАГ 4: Создать Compliance Handler

**Файл:** `api-gateway/internal/handlers/compliance.go`

**Документация:** `docs/api/compliance.md`

**Endpoints для реализации:**

| Method | Path | gRPC Method | Описание |
|--------|------|-------------|----------|
| POST | `/compliance/check` | CheckCompliance | Полная проверка сцены |
| POST | `/compliance/check-operation` | CheckOperation | Проверка операции |
| GET | `/compliance/rules` | GetRules | Список правил |
| GET | `/compliance/rules/:id` | GetRule | Получить правило |
| POST | `/compliance/report` | GenerateReport | Сгенерировать отчёт |

---

### ШАГ 5: Создать Request Handler

**Файл:** `api-gateway/internal/handlers/request.go`

**Документация:** `docs/api/requests.md`

**Endpoints для реализации:**

| Method | Path | gRPC Method | Описание |
|--------|------|-------------|----------|
| POST | `/requests` | Create | Создать заявку |
| GET | `/requests` | List | Список заявок |
| GET | `/requests/:id` | Get | Получить заявку |
| PATCH | `/requests/:id` | Update | Обновить заявку |
| POST | `/requests/:id/submit` | Submit | Отправить заявку |
| POST | `/requests/:id/cancel` | Cancel | Отменить заявку |
| POST | `/requests/:id/documents` | AddDocument | Добавить документ |
| GET | `/requests/:id/documents` | GetDocuments | Список документов |

---

### ШАГ 6: Зарегистрировать handlers в main.go

**Файл:** `api-gateway/cmd/main.go`

**Изменения:**

```go
// 1. Добавить импорты (если нет)
import (
    "github.com/xiiisorate/granula_api/api-gateway/internal/handlers"
)

// 2. Создать handlers (после создания gRPC клиентов)
floorPlanHandler := handlers.NewFloorPlanHandler(grpcClients.FloorPlanConn)
branchHandler := handlers.NewBranchHandler(grpcClients.BranchConn)
complianceHandler := handlers.NewComplianceHandler(grpcClients.ComplianceConn)
requestHandler := handlers.NewRequestHandler(grpcClients.RequestConn)

// 3. Заменить placeholder routes на реальные handlers

// FloorPlans
floorPlans := api.Group("/floor-plans")
floorPlans.Post("/", floorPlanHandler.Upload)
floorPlans.Get("/", floorPlanHandler.List)
floorPlans.Get("/:id", floorPlanHandler.Get)
floorPlans.Patch("/:id", floorPlanHandler.Update)
floorPlans.Delete("/:id", floorPlanHandler.Delete)
floorPlans.Post("/:id/recognize", floorPlanHandler.StartRecognition)
floorPlans.Get("/:id/recognition-status", floorPlanHandler.GetRecognitionStatus)

// Branches
branches := api.Group("/scenes/:scene_id/branches")
branches.Get("/", branchHandler.List)
branches.Post("/", branchHandler.Create)
branches.Get("/:id", branchHandler.Get)
branches.Patch("/:id", branchHandler.Update)
branches.Delete("/:id", branchHandler.Delete)
branches.Post("/:id/activate", branchHandler.Activate)
branches.Post("/:id/merge", branchHandler.Merge)
branches.Get("/:id/compare/:target_id", branchHandler.Compare)

// Compliance
compliance := api.Group("/compliance")
compliance.Post("/check", complianceHandler.Check)
compliance.Post("/check-operation", complianceHandler.CheckOperation)
compliance.Get("/rules", complianceHandler.GetRules)
compliance.Get("/rules/:id", complianceHandler.GetRule)

// Requests
requests := api.Group("/requests")
requests.Post("/", requestHandler.Create)
requests.Get("/", requestHandler.List)
requests.Get("/:id", requestHandler.Get)
requests.Patch("/:id", requestHandler.Update)
requests.Post("/:id/submit", requestHandler.Submit)
requests.Post("/:id/cancel", requestHandler.Cancel)
requests.Post("/:id/documents", requestHandler.AddDocument)
requests.Get("/:id/documents", requestHandler.GetDocuments)
```

---

### ШАГ 7: Добавить gRPC клиенты в Clients struct

**Файл:** `api-gateway/internal/grpc/clients.go`

**Добавить поля:**
```go
type Clients struct {
    // ... существующие поля ...
    FloorPlanConn  *grpc.ClientConn
    BranchConn     *grpc.ClientConn
    ComplianceConn *grpc.ClientConn
    RequestConn    *grpc.ClientConn
}

func NewClients(cfg *config.Config) (*Clients, error) {
    // ... существующий код ...
    
    // FloorPlan Service
    floorPlanConn, err := grpc.Dial(cfg.FloorPlanServiceAddr, grpc.WithInsecure())
    if err != nil {
        return nil, err
    }
    
    // Branch Service
    branchConn, err := grpc.Dial(cfg.BranchServiceAddr, grpc.WithInsecure())
    if err != nil {
        return nil, err
    }
    
    // Compliance Service
    complianceConn, err := grpc.Dial(cfg.ComplianceServiceAddr, grpc.WithInsecure())
    if err != nil {
        return nil, err
    }
    
    // Request Service
    requestConn, err := grpc.Dial(cfg.RequestServiceAddr, grpc.WithInsecure())
    if err != nil {
        return nil, err
    }
    
    return &Clients{
        // ...
        FloorPlanConn:  floorPlanConn,
        BranchConn:     branchConn,
        ComplianceConn: complianceConn,
        RequestConn:    requestConn,
    }, nil
}
```

---

## ✅ КРИТЕРИИ УСПЕХА

- [ ] Файл `handlers/floorplan.go` создан и компилируется
- [ ] Файл `handlers/branch.go` создан и компилируется
- [ ] Файл `handlers/compliance.go` создан и компилируется
- [ ] Файл `handlers/request.go` создан и компилируется
- [ ] Handlers зарегистрированы в `main.go`
- [ ] gRPC клиенты добавлены в `clients.go`
- [ ] `go build ./...` в api-gateway проходит без ошибок
- [ ] Swagger annotations добавлены

---

## 📚 СВЯЗАННАЯ ДОКУМЕНТАЦИЯ

| Документ | Путь | Для чего |
|----------|------|----------|
| FloorPlans API | `docs/api/floor-plans.md` | Спецификация endpoints |
| Branches API | `docs/api/branches.md` | Спецификация endpoints |
| Compliance API | `docs/api/compliance.md` | Спецификация endpoints |
| Requests API | `docs/api/requests.md` | Спецификация endpoints |
| Существующий ai.go | `api-gateway/internal/handlers/ai.go` | Референс для кода |
| Существующий workspace.go | `api-gateway/internal/handlers/workspace.go` | Референс для кода |

---

## ➡️ СЛЕДУЮЩИЙ ШАГ

После создания handlers, переходите к:
- [WORKPLAN-3-AI-MODULE.md](./WORKPLAN-3-AI-MODULE.md) — исправление AI модуля


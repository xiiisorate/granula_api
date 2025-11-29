# 🔌 WORKPLAN-4: Интеграции между сервисами

> **Приоритет:** 🟠 Важный  
> **Время:** 2-3 часа  
> **Зависимости:** WORKPLAN-1-PROTO.md, WORKPLAN-3-AI-MODULE.md  
> **Результат:** Сервисы взаимодействуют друг с другом

---

## 🎯 ЦЕЛЬ

Реализовать недостающие интеграции между сервисами:
1. AI Service → Scene Service (данные планировки)
2. Request Service → Notification Service (уведомления)
3. Workspace Service → Notification Service (уведомления)
4. Branch Service → Scene Service (элементы веток)

---

## 📋 ТЕКУЩЕЕ СОСТОЯНИЕ ИНТЕГРАЦИЙ

### ✅ Реализованные интеграции
| Источник | Назначение | Тип | Статус |
|----------|------------|-----|--------|
| API Gateway | All Services | gRPC | ✅ Работает |
| Scene Service | Compliance Service | gRPC | ✅ Работает |
| FloorPlan Service | AI Service | gRPC | ✅ Работает |
| FloorPlan Service | MinIO | S3 | ✅ Работает |

### ❌ Отсутствующие интеграции
| Источник | Назначение | Тип | Проблема |
|----------|------------|-----|----------|
| AI Service | Scene Service | gRPC | TODO в коде |
| Request Service | Notification Service | gRPC | TODO в коде |
| Workspace Service | Notification Service | gRPC | TODO в коде |
| Branch Service | Scene Service | gRPC | TODO в коде |

---

## 📁 ФАЙЛЫ С TODO

### AI Service → Scene Service
- `ai-service/internal/service/chat_service.go` (строка 314-318)
- `ai-service/internal/grpc/server.go` (строка 131)

### Request Service → Notification Service
- `request-service/internal/service/request_service.go`:
  - Строка 219: `TODO: Send notification to staff`
  - Строка 309: `TODO: Send notification to user`
  - Строка 353: `TODO: Send notifications to user and expert`
  - Строка 396: `TODO: Send notification to user`
  - Строка 442: `TODO: Send notification to user`

### Workspace Service → Notification Service
- `workspace-service/internal/service/workspace_service.go`:
  - Строка 155: `TODO: Publish workspace.created event`
  - Строка 345: `TODO: Publish workspace.deleted event`
  - Строка 415: `TODO: Send notification to new member`
  - Строка 601: `TODO: Send notification to invitee`

---

## 🔧 ПОШАГОВАЯ ИНСТРУКЦИЯ

### ШАГ 1: AI Service → Scene Service

> **Примечание:** Эта интеграция подробно описана в [WORKPLAN-3-AI-MODULE.md](./WORKPLAN-3-AI-MODULE.md)

**Краткое резюме:**
1. Создать `ai-service/internal/grpc/scene_client.go`
2. Добавить SceneClient в ChatService и GenerationService
3. Обновить `getSceneSummary()` для реального получения данных
4. Передавать `sceneData` в GenerateVariants

---

### ШАГ 2: Request Service → Notification Service

#### 2.1. Создать Notification клиент

**Создать файл:** `request-service/internal/grpc/notification_client.go`

```go
package grpc

import (
    "context"

    notificationpb "github.com/xiiisorate/granula_api/shared/gen/notification/v1"
    "google.golang.org/grpc"
)

// NotificationClient wraps notification service gRPC client.
type NotificationClient struct {
    client notificationpb.NotificationServiceClient
}

// NewNotificationClient creates a new notification client.
func NewNotificationClient(addr string) (*NotificationClient, error) {
    conn, err := grpc.Dial(addr, grpc.WithInsecure())
    if err != nil {
        return nil, err
    }
    return &NotificationClient{
        client: notificationpb.NewNotificationServiceClient(conn),
    }, nil
}

// SendRequestSubmitted sends notification when request is submitted.
func (c *NotificationClient) SendRequestSubmitted(ctx context.Context, userID, requestID string) error {
    _, err := c.client.Create(ctx, &notificationpb.CreateNotificationRequest{
        UserId:  userID,
        Type:    "request_submitted",
        Title:   "Заявка отправлена",
        Message: "Ваша заявка успешно отправлена на рассмотрение",
        Data: map[string]string{
            "request_id": requestID,
        },
    })
    return err
}

// SendRequestAssigned sends notification when expert is assigned.
func (c *NotificationClient) SendRequestAssigned(ctx context.Context, userID, requestID, expertName string) error {
    _, err := c.client.Create(ctx, &notificationpb.CreateNotificationRequest{
        UserId:  userID,
        Type:    "request_assigned",
        Title:   "Назначен эксперт",
        Message: "К вашей заявке назначен эксперт: " + expertName,
        Data: map[string]string{
            "request_id":  requestID,
            "expert_name": expertName,
        },
    })
    return err
}

// SendRequestStatusChanged sends notification when request status changes.
func (c *NotificationClient) SendRequestStatusChanged(ctx context.Context, userID, requestID, status string) error {
    titles := map[string]string{
        "in_review":  "Заявка на рассмотрении",
        "approved":   "Заявка одобрена",
        "rejected":   "Заявка отклонена",
        "completed":  "Заявка завершена",
        "cancelled":  "Заявка отменена",
    }

    title := titles[status]
    if title == "" {
        title = "Статус заявки изменён"
    }

    _, err := c.client.Create(ctx, &notificationpb.CreateNotificationRequest{
        UserId:  userID,
        Type:    "request_status_changed",
        Title:   title,
        Message: "Статус вашей заявки изменён на: " + status,
        Data: map[string]string{
            "request_id": requestID,
            "status":     status,
        },
    })
    return err
}

// NotifyStaff sends notification to staff about new request.
func (c *NotificationClient) NotifyStaff(ctx context.Context, requestID, requestType string) error {
    // В реальности здесь нужно получить список staff пользователей
    // Для MVP можно использовать фиксированный ID или broadcast
    _, err := c.client.Create(ctx, &notificationpb.CreateNotificationRequest{
        UserId:  "staff", // Или broadcast channel
        Type:    "new_request",
        Title:   "Новая заявка",
        Message: "Поступила новая заявка на: " + requestType,
        Data: map[string]string{
            "request_id": requestID,
        },
    })
    return err
}
```

#### 2.2. Добавить клиент в RequestService

**Файл:** `request-service/internal/service/request_service.go`

**Добавить поле:**
```go
type RequestService struct {
    repo              *repository.RequestRepository
    notificationClient *grpc.NotificationClient  // ADD
    log               *logger.Logger
}

func NewRequestService(repo *repository.RequestRepository, notificationClient *grpc.NotificationClient, log *logger.Logger) *RequestService {
    return &RequestService{
        repo:               repo,
        notificationClient: notificationClient,  // ADD
        log:                log,
    }
}
```

#### 2.3. Заменить TODO на реальные вызовы

**Строка ~219 (SubmitRequest):**
```go
func (s *RequestService) SubmitRequest(ctx context.Context, id string, userID string) (*entity.Request, error) {
    // ... existing code ...
    
    // Send notification to staff
    if s.notificationClient != nil {
        if err := s.notificationClient.NotifyStaff(ctx, request.ID, string(request.ServiceType)); err != nil {
            s.log.Warn("failed to notify staff", logger.Err(err))
        }
    }
    
    return request, nil
}
```

**Строка ~309 (UpdateStatus):**
```go
func (s *RequestService) UpdateStatus(ctx context.Context, id string, status entity.RequestStatus, comment string) (*entity.Request, error) {
    // ... existing code ...
    
    // Send notification to user
    if s.notificationClient != nil {
        if err := s.notificationClient.SendRequestStatusChanged(ctx, request.UserID, request.ID, string(status)); err != nil {
            s.log.Warn("failed to notify user", logger.Err(err))
        }
    }
    
    return request, nil
}
```

**Строка ~353 (AssignExpert):**
```go
func (s *RequestService) AssignExpert(ctx context.Context, id string, expertID string, expertName string) (*entity.Request, error) {
    // ... existing code ...
    
    // Send notification to user
    if s.notificationClient != nil {
        if err := s.notificationClient.SendRequestAssigned(ctx, request.UserID, request.ID, expertName); err != nil {
            s.log.Warn("failed to notify user", logger.Err(err))
        }
    }
    
    return request, nil
}
```

---

### ШАГ 3: Workspace Service → Notification Service

#### 3.1. Создать Notification клиент

**Создать файл:** `workspace-service/internal/grpc/notification_client.go`

```go
package grpc

import (
    "context"

    notificationpb "github.com/xiiisorate/granula_api/shared/gen/notification/v1"
    "google.golang.org/grpc"
)

// NotificationClient wraps notification service gRPC client.
type NotificationClient struct {
    client notificationpb.NotificationServiceClient
}

// NewNotificationClient creates a new notification client.
func NewNotificationClient(addr string) (*NotificationClient, error) {
    conn, err := grpc.Dial(addr, grpc.WithInsecure())
    if err != nil {
        return nil, err
    }
    return &NotificationClient{
        client: notificationpb.NewNotificationServiceClient(conn),
    }, nil
}

// SendMemberAdded sends notification when user is added to workspace.
func (c *NotificationClient) SendMemberAdded(ctx context.Context, userID, workspaceID, workspaceName, role string) error {
    _, err := c.client.Create(ctx, &notificationpb.CreateNotificationRequest{
        UserId:  userID,
        Type:    "workspace_member_added",
        Title:   "Вы добавлены в проект",
        Message: "Вы добавлены в проект \"" + workspaceName + "\" с ролью " + role,
        Data: map[string]string{
            "workspace_id":   workspaceID,
            "workspace_name": workspaceName,
            "role":           role,
        },
    })
    return err
}

// SendInvitation sends invitation notification.
func (c *NotificationClient) SendInvitation(ctx context.Context, email, workspaceID, workspaceName, inviterName string) error {
    // Для email инвайтов нужен отдельный механизм
    // Для MVP можно использовать in-app notification если пользователь зарегистрирован
    _, err := c.client.Create(ctx, &notificationpb.CreateNotificationRequest{
        UserId:  email, // В реальности нужно найти userID по email
        Type:    "workspace_invitation",
        Title:   "Приглашение в проект",
        Message: inviterName + " приглашает вас в проект \"" + workspaceName + "\"",
        Data: map[string]string{
            "workspace_id":   workspaceID,
            "workspace_name": workspaceName,
            "inviter_name":   inviterName,
        },
    })
    return err
}

// SendMemberRemoved sends notification when user is removed from workspace.
func (c *NotificationClient) SendMemberRemoved(ctx context.Context, userID, workspaceID, workspaceName string) error {
    _, err := c.client.Create(ctx, &notificationpb.CreateNotificationRequest{
        UserId:  userID,
        Type:    "workspace_member_removed",
        Title:   "Вы удалены из проекта",
        Message: "Вы были удалены из проекта \"" + workspaceName + "\"",
        Data: map[string]string{
            "workspace_id":   workspaceID,
            "workspace_name": workspaceName,
        },
    })
    return err
}
```

#### 3.2. Добавить клиент в WorkspaceService

**Файл:** `workspace-service/internal/service/workspace_service.go`

**Добавить поле и обновить конструктор:**
```go
type WorkspaceService struct {
    repo               *repository.WorkspaceRepository
    notificationClient *grpc.NotificationClient  // ADD
    log                *logger.Logger
}

func NewWorkspaceService(repo *repository.WorkspaceRepository, notificationClient *grpc.NotificationClient, log *logger.Logger) *WorkspaceService {
    return &WorkspaceService{
        repo:               repo,
        notificationClient: notificationClient,  // ADD
        log:                log,
    }
}
```

#### 3.3. Заменить TODO на реальные вызовы

**Строка ~415 (AddMember):**
```go
func (s *WorkspaceService) AddMember(ctx context.Context, workspaceID, userID, role string, addedBy string) (*entity.WorkspaceMember, error) {
    // ... existing code ...
    
    // Send notification to new member
    if s.notificationClient != nil {
        if err := s.notificationClient.SendMemberAdded(ctx, userID, workspace.ID, workspace.Name, role); err != nil {
            s.log.Warn("failed to notify new member", logger.Err(err))
        }
    }
    
    return member, nil
}
```

**Строка ~601 (InviteMember):**
```go
func (s *WorkspaceService) InviteMember(ctx context.Context, workspaceID, email, role string, invitedBy string) (*entity.Invitation, error) {
    // ... existing code ...
    
    // Send notification to invitee
    if s.notificationClient != nil {
        if err := s.notificationClient.SendInvitation(ctx, email, workspace.ID, workspace.Name, inviterName); err != nil {
            s.log.Warn("failed to send invitation notification", logger.Err(err))
        }
    }
    
    return invitation, nil
}
```

---

### ШАГ 4: Branch Service → Scene Service

#### 4.1. Создать Scene клиент

**Создать файл:** `branch-service/internal/grpc/scene_client.go`

```go
package grpc

import (
    "context"

    scenepb "github.com/xiiisorate/granula_api/shared/gen/scene/v1"
    "google.golang.org/grpc"
)

// SceneClient wraps scene service gRPC client.
type SceneClient struct {
    client scenepb.SceneServiceClient
}

// NewSceneClient creates a new scene client.
func NewSceneClient(addr string) (*SceneClient, error) {
    conn, err := grpc.Dial(addr, grpc.WithInsecure())
    if err != nil {
        return nil, err
    }
    return &SceneClient{
        client: scenepb.NewSceneServiceClient(conn),
    }, nil
}

// GetElements returns all elements for a scene/branch.
func (c *SceneClient) GetElements(ctx context.Context, sceneID, branchID string) ([]*scenepb.Element, error) {
    resp, err := c.client.ListElements(ctx, &scenepb.ListElementsRequest{
        SceneId:  sceneID,
        BranchId: branchID,
        Limit:    10000,
    })
    if err != nil {
        return nil, err
    }
    return resp.Elements, nil
}

// CopyElements copies elements from source to target branch.
func (c *SceneClient) CopyElements(ctx context.Context, sceneID, sourceBranchID, targetBranchID string) error {
    elements, err := c.GetElements(ctx, sceneID, sourceBranchID)
    if err != nil {
        return err
    }
    
    for _, el := range elements {
        // Create copy in target branch
        _, err := c.client.CreateElement(ctx, &scenepb.CreateElementRequest{
            SceneId:    sceneID,
            BranchId:   targetBranchID,
            Type:       el.Type,
            Name:       el.Name,
            Properties: el.Properties,
            Geometry:   el.Geometry,
        })
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

#### 4.2. Добавить клиент в BranchService

**Файл:** `branch-service/internal/service/branch_service.go`

**Исправить TODO в CreateBranch (строка ~37):**
```go
func (s *BranchService) CreateBranch(ctx context.Context, req CreateBranchRequest) (*entity.Branch, error) {
    // ... existing code ...
    
    // Copy elements from parent branch if parentID is set
    if req.ParentID != "" && s.sceneClient != nil {
        if err := s.sceneClient.CopyElements(ctx, req.SceneID, req.ParentID, branch.ID); err != nil {
            s.log.Warn("failed to copy elements from parent", logger.Err(err))
        }
    }
    
    return branch, nil
}
```

---

### ШАГ 5: Инициализация клиентов в main.go

#### 5.1. Request Service

**Файл:** `request-service/cmd/main.go`

```go
// Create notification client
notificationClient, err := grpc.NewNotificationClient(cfg.NotificationServiceAddr)
if err != nil {
    log.Warn("notification service unavailable", logger.Err(err))
    // Continue without notifications
}

// Create service with notification client
requestService := service.NewRequestService(requestRepo, notificationClient, log)
```

#### 5.2. Workspace Service

**Файл:** `workspace-service/cmd/main.go`

```go
// Create notification client
notificationClient, err := grpc.NewNotificationClient(cfg.NotificationServiceAddr)
if err != nil {
    log.Warn("notification service unavailable", logger.Err(err))
}

// Create service with notification client
workspaceService := service.NewWorkspaceService(workspaceRepo, notificationClient, log)
```

#### 5.3. Branch Service

**Файл:** `branch-service/cmd/main.go`

```go
// Create scene client
sceneClient, err := grpc.NewSceneClient(cfg.SceneServiceAddr)
if err != nil {
    log.Warn("scene service unavailable", logger.Err(err))
}

// Create service with scene client
branchService := service.NewBranchService(branchRepo, sceneClient, log)
```

---

### ШАГ 6: Добавить адреса сервисов в конфиг

**Файлы конфигов** (пример для request-service):

`request-service/internal/config/config.go`:
```go
type Config struct {
    // ... existing fields ...
    NotificationServiceAddr string `env:"NOTIFICATION_SERVICE_ADDR" envDefault:"notification-service:50051"`
}
```

`.env` или `docker-compose.yml`:
```yaml
environment:
  - NOTIFICATION_SERVICE_ADDR=notification-service:50051
```

---

## ✅ КРИТЕРИИ УСПЕХА

### AI → Scene
- [ ] SceneClient создан в AI Service
- [ ] ChatService использует реальные данные сцены
- [ ] GenerationService использует реальные данные сцены

### Request → Notification
- [ ] NotificationClient создан в Request Service
- [ ] Уведомления отправляются при submit
- [ ] Уведомления отправляются при смене статуса
- [ ] Уведомления отправляются при назначении эксперта

### Workspace → Notification
- [ ] NotificationClient создан в Workspace Service
- [ ] Уведомления отправляются при добавлении участника
- [ ] Уведомления отправляются при приглашении

### Branch → Scene
- [ ] SceneClient создан в Branch Service
- [ ] Элементы копируются при создании ветки из parent

---

## 📚 СВЯЗАННАЯ ДОКУМЕНТАЦИЯ

| Документ | Путь | Для чего |
|----------|------|----------|
| Notifications API | `docs/api/notifications.md` | Типы уведомлений |
| Requests API | `docs/api/requests.md` | Статусы заявок |
| Workspaces API | `docs/api/workspaces.md` | Участники, приглашения |
| Branches API | `docs/api/branches.md` | Создание веток |
| Микросервисы | `docs/architecture/microservices.md` | Порты, связи |

---

## ➡️ СЛЕДУЮЩИЙ ШАГ

После настройки интеграций, переходите к:
- [WORKPLAN-5-MIGRATIONS.md](./WORKPLAN-5-MIGRATIONS.md) — миграции БД


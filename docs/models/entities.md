# Доменные сущности

## Обзор

Доменные сущности представляют основные бизнес-объекты системы. Они не зависят от инфраструктуры и содержат бизнес-логику.

## User

```go
// internal/domain/entity/user.go

// User пользователь системы.
type User struct {
    // ID уникальный идентификатор (UUID v4)
    ID uuid.UUID `json:"id"`
    
    // Email адрес электронной почты (уникальный)
    Email string `json:"email"`
    
    // PasswordHash хэш пароля (bcrypt)
    // Пустой для OAuth пользователей
    PasswordHash string `json:"-"`
    
    // Name отображаемое имя
    Name string `json:"name"`
    
    // AvatarURL URL аватара в S3
    AvatarURL *string `json:"avatar_url,omitempty"`
    
    // Role роль в системе
    Role Role `json:"role"`
    
    // EmailVerified подтверждён ли email
    EmailVerified bool `json:"email_verified"`
    
    // OAuthProvider провайдер OAuth (google, yandex)
    OAuthProvider *string `json:"oauth_provider,omitempty"`
    
    // OAuthID идентификатор в системе OAuth провайдера
    OAuthID *string `json:"-"`
    
    // Settings настройки пользователя
    Settings *UserSettings `json:"settings,omitempty"`
    
    // CreatedAt время создания
    CreatedAt time.Time `json:"created_at"`
    
    // UpdatedAt время последнего обновления
    UpdatedAt time.Time `json:"updated_at"`
    
    // DeletedAt время удаления (soft delete)
    DeletedAt *time.Time `json:"-"`
}

// Role роль пользователя.
type Role string

const (
    RoleUser   Role = "user"
    RoleAdmin  Role = "admin"
    RoleExpert Role = "expert"
)

// UserSettings настройки пользователя.
type UserSettings struct {
    Language      string                `json:"language"`       // ru, en
    Theme         string                `json:"theme"`          // light, dark, system
    Units         string                `json:"units"`          // metric, imperial
    Notifications *NotificationSettings `json:"notifications"`
}

// NotificationSettings настройки уведомлений.
type NotificationSettings struct {
    Email     bool `json:"email"`
    Push      bool `json:"push"`
    Marketing bool `json:"marketing"`
}

// IsAdmin проверяет, является ли пользователь администратором.
func (u *User) IsAdmin() bool {
    return u.Role == RoleAdmin
}

// IsExpert проверяет, является ли пользователь экспертом.
func (u *User) IsExpert() bool {
    return u.Role == RoleExpert
}

// HasOAuth проверяет, использует ли пользователь OAuth.
func (u *User) HasOAuth() bool {
    return u.OAuthProvider != nil && *u.OAuthProvider != ""
}
```

## Workspace

```go
// internal/domain/entity/workspace.go

// Workspace проект ремонта/перепланировки.
type Workspace struct {
    // ID уникальный идентификатор
    ID uuid.UUID `json:"id"`
    
    // OwnerID ID владельца
    OwnerID uuid.UUID `json:"owner_id"`
    
    // Name название проекта
    Name string `json:"name"`
    
    // Description описание
    Description string `json:"description,omitempty"`
    
    // Address адрес квартиры
    Address string `json:"address,omitempty"`
    
    // TotalArea общая площадь в м²
    TotalArea *float64 `json:"total_area,omitempty"`
    
    // RoomsCount количество комнат
    RoomsCount *int `json:"rooms_count,omitempty"`
    
    // Status статус проекта
    Status WorkspaceStatus `json:"status"`
    
    // Settings настройки воркспейса
    Settings *WorkspaceSettings `json:"settings"`
    
    // PreviewURL URL превью изображения
    PreviewURL *string `json:"preview_url,omitempty"`
    
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    DeletedAt *time.Time `json:"-"`
}

// WorkspaceStatus статус воркспейса.
type WorkspaceStatus string

const (
    WorkspaceStatusDraft     WorkspaceStatus = "draft"
    WorkspaceStatusActive    WorkspaceStatus = "active"
    WorkspaceStatusCompleted WorkspaceStatus = "completed"
    WorkspaceStatusArchived  WorkspaceStatus = "archived"
)

// WorkspaceSettings настройки воркспейса.
type WorkspaceSettings struct {
    Units          string  `json:"units"`           // metric, imperial
    GridSize       float64 `json:"grid_size"`       // размер сетки в метрах
    WallHeight     float64 `json:"wall_height"`     // высота стен по умолчанию
    SnapToGrid     bool    `json:"snap_to_grid"`
    ShowDimensions bool    `json:"show_dimensions"`
}

// DefaultWorkspaceSettings возвращает настройки по умолчанию.
func DefaultWorkspaceSettings() *WorkspaceSettings {
    return &WorkspaceSettings{
        Units:          "metric",
        GridSize:       0.1,
        WallHeight:     2.7,
        SnapToGrid:     true,
        ShowDimensions: true,
    }
}
```

## WorkspaceMember

```go
// internal/domain/entity/workspace_member.go

// WorkspaceMember участник воркспейса.
type WorkspaceMember struct {
    ID          uuid.UUID     `json:"id"`
    WorkspaceID uuid.UUID     `json:"workspace_id"`
    UserID      uuid.UUID     `json:"user_id"`
    Role        WorkspaceRole `json:"role"`
    InvitedBy   *uuid.UUID    `json:"invited_by,omitempty"`
    JoinedAt    time.Time     `json:"joined_at"`
}

// WorkspaceRole роль в воркспейсе.
type WorkspaceRole string

const (
    WorkspaceRoleOwner  WorkspaceRole = "owner"
    WorkspaceRoleEditor WorkspaceRole = "editor"
    WorkspaceRoleViewer WorkspaceRole = "viewer"
)

// CanEdit проверяет право на редактирование.
func (r WorkspaceRole) CanEdit() bool {
    return r == WorkspaceRoleOwner || r == WorkspaceRoleEditor
}

// CanDelete проверяет право на удаление.
func (r WorkspaceRole) CanDelete() bool {
    return r == WorkspaceRoleOwner
}

// CanInvite проверяет право на приглашение.
func (r WorkspaceRole) CanInvite() bool {
    return r == WorkspaceRoleOwner || r == WorkspaceRoleEditor
}

// CanManageMembers проверяет право на управление участниками.
func (r WorkspaceRole) CanManageMembers() bool {
    return r == WorkspaceRoleOwner
}
```

## FloorPlan

```go
// internal/domain/entity/floor_plan.go

// FloorPlan загруженная планировка.
type FloorPlan struct {
    ID              uuid.UUID         `json:"id"`
    WorkspaceID     uuid.UUID         `json:"workspace_id"`
    FilePath        string            `json:"file_path"`
    FileType        string            `json:"file_type"`
    OriginalName    string            `json:"original_name"`
    FileSize        int64             `json:"file_size"`
    SourceType      FloorPlanSource   `json:"source_type"`
    Status          FloorPlanStatus   `json:"status"`
    RecognitionData *RecognitionData  `json:"recognition_data,omitempty"`
    ErrorMessage    *string           `json:"error_message,omitempty"`
    CreatedAt       time.Time         `json:"created_at"`
    UpdatedAt       time.Time         `json:"updated_at"`
}

// FloorPlanSource источник планировки.
type FloorPlanSource string

const (
    FloorPlanSourceBTI           FloorPlanSource = "bti"
    FloorPlanSourceTechnicalPlan FloorPlanSource = "technical_plan"
    FloorPlanSourceSketch        FloorPlanSource = "sketch"
    FloorPlanSourceOther         FloorPlanSource = "other"
)

// FloorPlanStatus статус обработки.
type FloorPlanStatus string

const (
    FloorPlanStatusPending    FloorPlanStatus = "pending"
    FloorPlanStatusProcessing FloorPlanStatus = "processing"
    FloorPlanStatusCompleted  FloorPlanStatus = "completed"
    FloorPlanStatusFailed     FloorPlanStatus = "failed"
)

// RecognitionData результат распознавания.
type RecognitionData struct {
    Bounds    *Bounds          `json:"bounds"`
    Walls     []WallData       `json:"walls"`
    Rooms     []RoomData       `json:"rooms"`
    Openings  []OpeningData    `json:"openings"`
    Utilities []UtilityData    `json:"utilities"`
    Metadata  *RecognitionMeta `json:"metadata"`
}

// Bounds габариты помещения.
type Bounds struct {
    Width  float64 `json:"width"`
    Height float64 `json:"height"`
    Depth  float64 `json:"depth"`
}

// WallData данные стены.
type WallData struct {
    ID            string  `json:"id"`
    Start         Point3D `json:"start"`
    End           Point3D `json:"end"`
    Thickness     float64 `json:"thickness"`
    IsLoadBearing bool    `json:"is_load_bearing"`
    Material      string  `json:"material,omitempty"`
    Confidence    float64 `json:"confidence"`
}

// RoomData данные комнаты.
type RoomData struct {
    ID         string    `json:"id"`
    Type       string    `json:"type"`
    Name       string    `json:"name"`
    Polygon    []Point2D `json:"polygon"`
    Area       float64   `json:"area"`
    Confidence float64   `json:"confidence"`
}

// OpeningData данные проёма.
type OpeningData struct {
    ID         string  `json:"id"`
    Type       string  `json:"type"` // door, window
    WallID     string  `json:"wall_id"`
    Position   float64 `json:"position"`
    Width      float64 `json:"width"`
    Height     float64 `json:"height"`
    Elevation  float64 `json:"elevation,omitempty"`
    Confidence float64 `json:"confidence"`
}

// UtilityData данные инженерного элемента.
type UtilityData struct {
    ID          string  `json:"id"`
    Type        string  `json:"type"`
    Position    Point3D `json:"position"`
    CanRelocate bool    `json:"can_relocate"`
    Confidence  float64 `json:"confidence"`
}

// Point2D точка на плоскости.
type Point2D struct {
    X float64 `json:"x"`
    Z float64 `json:"z"`
}

// Point3D точка в пространстве.
type Point3D struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
    Z float64 `json:"z"`
}
```

## Scene (MongoDB)

```go
// internal/domain/entity/scene.go

// Scene 3D сцена квартиры.
type Scene struct {
    ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    WorkspaceID     string             `bson:"workspaceId" json:"workspace_id"`
    FloorPlanID     *string            `bson:"floorPlanId,omitempty" json:"floor_plan_id,omitempty"`
    Name            string             `bson:"name" json:"name"`
    Description     string             `bson:"description,omitempty" json:"description,omitempty"`
    SchemaVersion   int                `bson:"schemaVersion" json:"schema_version"`
    Bounds          *Bounds            `bson:"bounds" json:"bounds"`
    Elements        *SceneElements     `bson:"elements" json:"elements"`
    DisplaySettings *DisplaySettings   `bson:"displaySettings" json:"display_settings"`
    ComplianceResult *ComplianceResult `bson:"complianceResult,omitempty" json:"compliance_result,omitempty"`
    Stats           *SceneStats        `bson:"stats" json:"stats"`
    CreatedAt       time.Time          `bson:"createdAt" json:"created_at"`
    UpdatedAt       time.Time          `bson:"updatedAt" json:"updated_at"`
    CreatedBy       string             `bson:"createdBy" json:"created_by"`
    UpdatedBy       string             `bson:"updatedBy" json:"updated_by"`
}

// SceneElements все элементы сцены.
type SceneElements struct {
    Walls     []WallElement      `bson:"walls" json:"walls"`
    Rooms     []RoomElement      `bson:"rooms" json:"rooms"`
    Furniture []FurnitureElement `bson:"furniture" json:"furniture"`
    Utilities []UtilityElement   `bson:"utilities" json:"utilities"`
}

// WallElement элемент стены.
type WallElement struct {
    ID         string           `bson:"id" json:"id"`
    Type       string           `bson:"type" json:"type"`
    Name       string           `bson:"name" json:"name"`
    Start      Point3D          `bson:"start" json:"start"`
    End        Point3D          `bson:"end" json:"end"`
    Height     float64          `bson:"height" json:"height"`
    Thickness  float64          `bson:"thickness" json:"thickness"`
    Properties *WallProperties  `bson:"properties" json:"properties"`
    Openings   []OpeningElement `bson:"openings,omitempty" json:"openings,omitempty"`
    Metadata   *ElementMetadata `bson:"metadata" json:"metadata"`
}

// WallProperties свойства стены.
type WallProperties struct {
    IsLoadBearing bool   `bson:"isLoadBearing" json:"is_load_bearing"`
    Material      string `bson:"material" json:"material"`
    CanDemolish   bool   `bson:"canDemolish" json:"can_demolish"`
}

// RoomElement элемент комнаты.
type RoomElement struct {
    ID         string          `bson:"id" json:"id"`
    Type       string          `bson:"type" json:"type"`
    Name       string          `bson:"name" json:"name"`
    RoomType   string          `bson:"roomType" json:"room_type"`
    Polygon    []Point2D       `bson:"polygon" json:"polygon"`
    Area       float64         `bson:"area" json:"area"`
    Perimeter  float64         `bson:"perimeter" json:"perimeter"`
    Properties *RoomProperties `bson:"properties" json:"properties"`
}

// FurnitureElement элемент мебели.
type FurnitureElement struct {
    ID            string             `bson:"id" json:"id"`
    Type          string             `bson:"type" json:"type"`
    Name          string             `bson:"name" json:"name"`
    FurnitureType string             `bson:"furnitureType" json:"furniture_type"`
    Position      Point3D            `bson:"position" json:"position"`
    Rotation      Point3D            `bson:"rotation" json:"rotation"`
    Dimensions    *Dimensions        `bson:"dimensions" json:"dimensions"`
    ModelURL      string             `bson:"modelUrl,omitempty" json:"model_url,omitempty"`
    Metadata      *FurnitureMetadata `bson:"metadata" json:"metadata"`
}

// DisplaySettings настройки отображения.
type DisplaySettings struct {
    FloorTexture  string  `bson:"floorTexture" json:"floor_texture"`
    WallColor     string  `bson:"wallColor" json:"wall_color"`
    CeilingColor  string  `bson:"ceilingColor" json:"ceiling_color"`
    AmbientLight  float64 `bson:"ambientLight" json:"ambient_light"`
    ShowGrid      bool    `bson:"showGrid" json:"show_grid"`
    GridSize      float64 `bson:"gridSize" json:"grid_size"`
}

// SceneStats статистика сцены.
type SceneStats struct {
    TotalArea      float64 `bson:"totalArea" json:"total_area"`
    RoomsCount     int     `bson:"roomsCount" json:"rooms_count"`
    WallsCount     int     `bson:"wallsCount" json:"walls_count"`
    FurnitureCount int     `bson:"furnitureCount" json:"furniture_count"`
}
```

## Branch (MongoDB)

```go
// internal/domain/entity/branch.go

// Branch ветка дизайна.
type Branch struct {
    ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    SceneID          primitive.ObjectID `bson:"sceneId" json:"scene_id"`
    ParentBranchID   *primitive.ObjectID `bson:"parentBranchId,omitempty" json:"parent_branch_id,omitempty"`
    Name             string             `bson:"name" json:"name"`
    Description      string             `bson:"description,omitempty" json:"description,omitempty"`
    Source           BranchSource       `bson:"source" json:"source"`
    Order            int                `bson:"order" json:"order"`
    IsActive         bool               `bson:"isActive" json:"is_active"`
    IsFavorite       bool               `bson:"isFavorite" json:"is_favorite"`
    Delta            *BranchDelta       `bson:"delta" json:"delta"`
    Snapshot         *BranchSnapshot    `bson:"snapshot,omitempty" json:"snapshot,omitempty"`
    ComplianceResult *ComplianceResult  `bson:"complianceResult,omitempty" json:"compliance_result,omitempty"`
    AIContext        *AIContext         `bson:"aiContext,omitempty" json:"ai_context,omitempty"`
    PreviewURL       *string            `bson:"previewUrl,omitempty" json:"preview_url,omitempty"`
    CreatedAt        time.Time          `bson:"createdAt" json:"created_at"`
    UpdatedAt        time.Time          `bson:"updatedAt" json:"updated_at"`
    CreatedBy        string             `bson:"createdBy" json:"created_by"`
}

// BranchSource источник ветки.
type BranchSource string

const (
    BranchSourceUser BranchSource = "user"
    BranchSourceAI   BranchSource = "ai"
)

// BranchDelta изменения относительно родителя.
type BranchDelta struct {
    Added    map[string][]interface{}     `bson:"added" json:"added"`
    Modified map[string]interface{}       `bson:"modified" json:"modified"`
    Removed  []string                     `bson:"removed" json:"removed"`
}

// BranchSnapshot полный снимок состояния.
type BranchSnapshot struct {
    Elements *SceneElements `bson:"elements" json:"elements"`
    Bounds   *Bounds        `bson:"bounds" json:"bounds"`
    Stats    *SceneStats    `bson:"stats" json:"stats"`
}

// AIContext контекст AI-генерации.
type AIContext struct {
    Prompt      string    `bson:"prompt" json:"prompt"`
    Model       string    `bson:"model" json:"model"`
    GeneratedAt time.Time `bson:"generatedAt" json:"generated_at"`
    Reasoning   string    `bson:"reasoning,omitempty" json:"reasoning,omitempty"`
}
```

## ExpertRequest

```go
// internal/domain/entity/expert_request.go

// ExpertRequest заявка на эксперта.
type ExpertRequest struct {
    ID                 uuid.UUID           `json:"id"`
    WorkspaceID        uuid.UUID           `json:"workspace_id"`
    UserID             uuid.UUID           `json:"user_id"`
    SceneID            string              `json:"scene_id"`
    BranchID           *string             `json:"branch_id,omitempty"`
    ServiceType        ServiceType         `json:"service_type"`
    Status             RequestStatus       `json:"status"`
    ContactName        string              `json:"contact_name"`
    ContactPhone       string              `json:"contact_phone"`
    ContactEmail       string              `json:"contact_email"`
    PreferredTime      string              `json:"preferred_contact_time,omitempty"`
    Comment            string              `json:"comment,omitempty"`
    AssignedExpertID   *uuid.UUID          `json:"assigned_expert_id,omitempty"`
    EstimatedDate      *time.Time          `json:"estimated_date,omitempty"`
    EstimatedPrice     *float64            `json:"estimated_price,omitempty"`
    RejectionReason    *string             `json:"rejection_reason,omitempty"`
    StatusHistory      []StatusHistoryEntry `json:"status_history"`
    CreatedAt          time.Time           `json:"created_at"`
    UpdatedAt          time.Time           `json:"updated_at"`
}

// ServiceType тип услуги.
type ServiceType string

const (
    ServiceConsultation  ServiceType = "consultation"
    ServiceDocumentation ServiceType = "documentation"
    ServiceExpertVisit   ServiceType = "expert_visit"
    ServiceFullService   ServiceType = "full_service"
)

// RequestStatus статус заявки.
type RequestStatus string

const (
    RequestStatusPending    RequestStatus = "pending"
    RequestStatusReviewing  RequestStatus = "reviewing"
    RequestStatusApproved   RequestStatus = "approved"
    RequestStatusRejected   RequestStatus = "rejected"
    RequestStatusInProgress RequestStatus = "in_progress"
    RequestStatusCompleted  RequestStatus = "completed"
    RequestStatusCancelled  RequestStatus = "cancelled"
)

// StatusHistoryEntry запись истории статусов.
type StatusHistoryEntry struct {
    Status    RequestStatus `json:"status"`
    ChangedAt time.Time     `json:"changed_at"`
    ChangedBy string        `json:"changed_by"`
    Comment   *string       `json:"comment,omitempty"`
}

// CanTransitionTo проверяет допустимость перехода статуса.
func (s RequestStatus) CanTransitionTo(target RequestStatus) bool {
    transitions := map[RequestStatus][]RequestStatus{
        RequestStatusPending:    {RequestStatusReviewing, RequestStatusRejected, RequestStatusCancelled},
        RequestStatusReviewing:  {RequestStatusApproved, RequestStatusRejected, RequestStatusCancelled},
        RequestStatusApproved:   {RequestStatusInProgress, RequestStatusCancelled},
        RequestStatusInProgress: {RequestStatusCompleted, RequestStatusCancelled},
    }
    
    allowed, ok := transitions[s]
    if !ok {
        return false
    }
    
    for _, t := range allowed {
        if t == target {
            return true
        }
    }
    return false
}
```

## ComplianceRule

```go
// internal/domain/entity/compliance_rule.go

// ComplianceRule правило проверки.
type ComplianceRule struct {
    ID          uuid.UUID         `json:"id"`
    Code        string            `json:"code"`
    Name        string            `json:"name"`
    Category    RuleCategory      `json:"category"`
    Severity    RuleSeverity      `json:"severity"`
    Description string            `json:"description"`
    RuleConfig  map[string]interface{} `json:"rule_config"`
    Source      string            `json:"source"`
    SourceURL   *string           `json:"source_url,omitempty"`
    Active      bool              `json:"active"`
    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
}

// RuleCategory категория правила.
type RuleCategory string

const (
    RuleCategoryStructural    RuleCategory = "structural"
    RuleCategoryPlumbing      RuleCategory = "plumbing"
    RuleCategoryElectrical    RuleCategory = "electrical"
    RuleCategoryVentilation   RuleCategory = "ventilation"
    RuleCategoryFireSafety    RuleCategory = "fire_safety"
    RuleCategoryAccessibility RuleCategory = "accessibility"
    RuleCategoryGeneral       RuleCategory = "general"
)

// RuleSeverity критичность правила.
type RuleSeverity string

const (
    RuleSeverityError   RuleSeverity = "error"
    RuleSeverityWarning RuleSeverity = "warning"
    RuleSeverityInfo    RuleSeverity = "info"
)

// ComplianceResult результат проверки.
type ComplianceResult struct {
    LastCheckedAt time.Time            `json:"last_checked_at"`
    IsCompliant   bool                 `json:"is_compliant"`
    Violations    []ComplianceViolation `json:"violations"`
    Warnings      []ComplianceViolation `json:"warnings"`
    Info          []ComplianceViolation `json:"info,omitempty"`
}

// ComplianceViolation нарушение.
type ComplianceViolation struct {
    RuleCode         string            `json:"rule_code"`
    RuleName         string            `json:"rule_name"`
    Category         RuleCategory      `json:"category"`
    Severity         RuleSeverity      `json:"severity"`
    Message          string            `json:"message"`
    AffectedElements []AffectedElement `json:"affected_elements"`
    Suggestion       string            `json:"suggestion,omitempty"`
    Source           string            `json:"source"`
    SourceURL        *string           `json:"source_url,omitempty"`
}

// AffectedElement затронутый элемент.
type AffectedElement struct {
    Type string `json:"type"`
    ID   string `json:"id"`
    Name string `json:"name"`
}
```

## Notification

```go
// internal/domain/entity/notification.go

// Notification уведомление.
type Notification struct {
    ID        uuid.UUID              `json:"id"`
    UserID    uuid.UUID              `json:"user_id"`
    Type      NotificationType       `json:"type"`
    Title     string                 `json:"title"`
    Message   string                 `json:"message"`
    Data      map[string]interface{} `json:"data,omitempty"`
    Read      bool                   `json:"read"`
    ReadAt    *time.Time             `json:"read_at,omitempty"`
    CreatedAt time.Time              `json:"created_at"`
}

// NotificationType тип уведомления.
type NotificationType string

const (
    NotificationRequestStatus     NotificationType = "request_status"
    NotificationComplianceWarning NotificationType = "compliance_warning"
    NotificationWorkspaceInvite   NotificationType = "workspace_invite"
    NotificationAIComplete        NotificationType = "ai_generation_complete"
    NotificationSystem            NotificationType = "system"
)
```


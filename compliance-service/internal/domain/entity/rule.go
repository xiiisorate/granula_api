// Package entity defines domain entities for Compliance Service.
package entity

import (
	"time"

	"github.com/google/uuid"
)

// Severity represents the severity level of a compliance rule violation.
type Severity string

const (
	// SeverityInfo is for informational/recommendation rules.
	SeverityInfo Severity = "INFO"

	// SeverityWarning is for rules that can be worked around with approval.
	SeverityWarning Severity = "WARNING"

	// SeverityError is for critical violations that cannot be approved.
	SeverityError Severity = "ERROR"
)

// RuleCategory represents a category of compliance rules.
type RuleCategory string

const (
	// CategoryLoadBearing rules related to load-bearing structures.
	CategoryLoadBearing RuleCategory = "load_bearing"

	// CategoryWetZones rules for wet zones (bathrooms, kitchens).
	CategoryWetZones RuleCategory = "wet_zones"

	// CategoryFireSafety fire safety related rules.
	CategoryFireSafety RuleCategory = "fire_safety"

	// CategoryVentilation ventilation requirements.
	CategoryVentilation RuleCategory = "ventilation"

	// CategoryMinArea minimum area requirements.
	CategoryMinArea RuleCategory = "min_area"

	// CategoryDaylight natural daylight requirements.
	CategoryDaylight RuleCategory = "daylight"

	// CategoryAccessibility accessibility requirements.
	CategoryAccessibility RuleCategory = "accessibility"

	// CategoryGeneral general planning rules.
	CategoryGeneral RuleCategory = "general"
)

// ElementType represents the type of scene element a rule applies to.
type ElementType string

const (
	ElementTypeWall            ElementType = "WALL"
	ElementTypeLoadBearingWall ElementType = "LOAD_BEARING_WALL"
	ElementTypeRoom            ElementType = "ROOM"
	ElementTypeDoor            ElementType = "DOOR"
	ElementTypeWindow          ElementType = "WINDOW"
	ElementTypeWetZone         ElementType = "WET_ZONE"
	ElementTypeKitchen         ElementType = "KITCHEN"
	ElementTypeBathroom        ElementType = "BATHROOM"
	ElementTypeToilet          ElementType = "TOILET"
	ElementTypeSink            ElementType = "SINK"
	ElementTypeVentilation     ElementType = "VENTILATION"
)

// OperationType represents the type of operation a rule applies to.
type OperationType string

const (
	OperationTypeDemolishWall    OperationType = "DEMOLISH_WALL"
	OperationTypeAddWall         OperationType = "ADD_WALL"
	OperationTypeMoveWall        OperationType = "MOVE_WALL"
	OperationTypeAddOpening      OperationType = "ADD_OPENING"
	OperationTypeCloseOpening    OperationType = "CLOSE_OPENING"
	OperationTypeMergeRooms      OperationType = "MERGE_ROOMS"
	OperationTypeSplitRoom       OperationType = "SPLIT_ROOM"
	OperationTypeChangeRoomType  OperationType = "CHANGE_ROOM_TYPE"
	OperationTypeMoveWetZone     OperationType = "MOVE_WET_ZONE"
	OperationTypeExpandWetZone   OperationType = "EXPAND_WET_ZONE"
	OperationTypeMovePlumbing    OperationType = "MOVE_PLUMBING"
	OperationTypeMoveVentilation OperationType = "MOVE_VENTILATION"
)

// ApprovalType represents the type of approval required.
type ApprovalType string

const (
	// ApprovalTypeNone no approval required.
	ApprovalTypeNone ApprovalType = "NONE"

	// ApprovalTypeNotification simple notification to authorities.
	ApprovalTypeNotification ApprovalType = "NOTIFICATION"

	// ApprovalTypeProject requires a project from a certified designer.
	ApprovalTypeProject ApprovalType = "PROJECT"

	// ApprovalTypeExpertise requires structural expertise.
	ApprovalTypeExpertise ApprovalType = "EXPERTISE"

	// ApprovalTypeProhibited operation is prohibited and cannot be approved.
	ApprovalTypeProhibited ApprovalType = "PROHIBITED"
)

// Rule represents a compliance rule in the system.
type Rule struct {
	// ID is the unique identifier.
	ID uuid.UUID `json:"id" db:"id"`

	// Code is the rule code (e.g., "СНиП 31-01-2003 п.9.22").
	Code string `json:"code" db:"code"`

	// Category of the rule.
	Category RuleCategory `json:"category" db:"category"`

	// Name is a short name for the rule.
	Name string `json:"name" db:"name"`

	// Description is a detailed explanation.
	Description string `json:"description" db:"description"`

	// Severity of violations of this rule.
	Severity Severity `json:"severity" db:"severity"`

	// Active indicates if the rule is currently enforced.
	Active bool `json:"active" db:"active"`

	// AppliesTo is the list of element types this rule applies to.
	AppliesTo []ElementType `json:"applies_to" db:"applies_to"`

	// AppliestoOperations is the list of operations this rule checks.
	AppliesToOperations []OperationType `json:"applies_to_operations" db:"applies_to_operations"`

	// ApprovalRequired is the type of approval needed if this rule is violated.
	ApprovalRequired ApprovalType `json:"approval_required" db:"approval_required"`

	// Parameters are rule-specific parameters (e.g., min_area: 8).
	Parameters map[string]interface{} `json:"parameters" db:"parameters"`

	// References are links to regulatory documents.
	References []DocumentReference `json:"references" db:"references"`

	// Version of the rule.
	Version string `json:"version" db:"version"`

	// CreatedAt timestamp.
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt timestamp.
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// DocumentReference represents a reference to a regulatory document.
type DocumentReference struct {
	// Code of the document (e.g., "СНиП 31-01-2003").
	Code string `json:"code"`

	// Title of the document.
	Title string `json:"title"`

	// Section or paragraph reference.
	Section string `json:"section"`

	// URL to the document text (optional).
	URL string `json:"url,omitempty"`
}

// RuleCategoryInfo provides metadata about a rule category.
type RuleCategoryInfo struct {
	// ID of the category.
	ID RuleCategory `json:"id"`

	// Name for display.
	Name string `json:"name"`

	// Description of what this category covers.
	Description string `json:"description"`

	// Icon name for UI.
	Icon string `json:"icon"`

	// RulesCount is the number of rules in this category.
	RulesCount int `json:"rules_count"`
}

// NewRule creates a new Rule with default values.
func NewRule(code, name string, category RuleCategory, severity Severity) *Rule {
	now := time.Now().UTC()
	return &Rule{
		ID:        uuid.New(),
		Code:      code,
		Name:      name,
		Category:  category,
		Severity:  severity,
		Active:    true,
		Version:   "1.0.0",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Deactivate marks the rule as inactive.
func (r *Rule) Deactivate() {
	r.Active = false
	r.UpdatedAt = time.Now().UTC()
}

// Activate marks the rule as active.
func (r *Rule) Activate() {
	r.Active = true
	r.UpdatedAt = time.Now().UTC()
}

// AddReference adds a document reference to the rule.
func (r *Rule) AddReference(ref DocumentReference) {
	r.References = append(r.References, ref)
	r.UpdatedAt = time.Now().UTC()
}

// SetParameter sets a rule parameter.
func (r *Rule) SetParameter(key string, value interface{}) {
	if r.Parameters == nil {
		r.Parameters = make(map[string]interface{})
	}
	r.Parameters[key] = value
	r.UpdatedAt = time.Now().UTC()
}

// GetParameter gets a rule parameter with type assertion.
func (r *Rule) GetParameter(key string) (interface{}, bool) {
	if r.Parameters == nil {
		return nil, false
	}
	val, ok := r.Parameters[key]
	return val, ok
}

// GetFloatParameter gets a float parameter.
func (r *Rule) GetFloatParameter(key string, defaultVal float64) float64 {
	val, ok := r.GetParameter(key)
	if !ok {
		return defaultVal
	}
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return defaultVal
	}
}

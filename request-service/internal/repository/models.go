// Package repository handles data access for Request Service.
package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Request represents an expert request in the database.
type Request struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key"`
	WorkspaceID uuid.UUID      `gorm:"type:uuid;not null;index"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null;index"`
	Title       string         `gorm:"type:varchar(255);not null"`
	Description string         `gorm:"type:text"`
	Category    string         `gorm:"type:varchar(100);not null"`
	Status      string         `gorm:"type:varchar(50);default:'pending';not null;index"`
	ExpertID    *uuid.UUID     `gorm:"type:uuid;index"`
	Priority    string         `gorm:"type:varchar(20);default:'normal';not null"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relations
	StatusHistory []StatusHistory `gorm:"foreignKey:RequestID"`
}

// StatusHistory tracks request status changes.
type StatusHistory struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	RequestID uuid.UUID `gorm:"type:uuid;not null;index"`
	Status    string    `gorm:"type:varchar(50);not null"`
	Comment   string    `gorm:"type:text"`
	ChangedBy uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// RequestDocument represents a document attached to request.
type RequestDocument struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key"`
	RequestID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Name        string    `gorm:"type:varchar(255);not null"`
	URL         string    `gorm:"type:varchar(500);not null"`
	Size        int64     `gorm:"not null"`
	ContentType string    `gorm:"type:varchar(100);not null"`
	UploadedBy  uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}

// TableName sets the table name for Request.
func (Request) TableName() string {
	return "requests"
}

// TableName sets the table name for StatusHistory.
func (StatusHistory) TableName() string {
	return "request_status_history"
}

// TableName sets the table name for RequestDocument.
func (RequestDocument) TableName() string {
	return "request_documents"
}

// Migrate runs database migrations.
func Migrate(db *gorm.DB) error {
	// Enable UUID extension
	db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)

	return db.AutoMigrate(&Request{}, &StatusHistory{}, &RequestDocument{})
}

// Request status constants
const (
	StatusPending    = "pending"
	StatusReview     = "review"
	StatusApproved   = "approved"
	StatusRejected   = "rejected"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	StatusCancelled  = "cancelled"
)

// Request category constants
const (
	CategoryConsultation   = "consultation"
	CategoryDocumentation  = "documentation"
	CategoryExpertVisit    = "expert_visit"
	CategoryFullService    = "full_service"
)

// Request priority constants
const (
	PriorityLow    = "low"
	PriorityNormal = "normal"
	PriorityHigh   = "high"
	PriorityUrgent = "urgent"
)

// IsValidStatus checks if a status is valid.
func IsValidStatus(status string) bool {
	switch status {
	case StatusPending, StatusReview, StatusApproved, StatusRejected,
		StatusInProgress, StatusCompleted, StatusCancelled:
		return true
	default:
		return false
	}
}

// IsValidCategory checks if a category is valid.
func IsValidCategory(category string) bool {
	switch category {
	case CategoryConsultation, CategoryDocumentation, CategoryExpertVisit, CategoryFullService:
		return true
	default:
		return false
	}
}

// IsValidPriority checks if a priority is valid.
func IsValidPriority(priority string) bool {
	switch priority {
	case PriorityLow, PriorityNormal, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
}

// CanTransitionTo checks if status transition is valid.
func CanTransitionTo(from, to string) bool {
	transitions := map[string][]string{
		StatusPending:    {StatusReview, StatusCancelled},
		StatusReview:     {StatusApproved, StatusRejected, StatusCancelled},
		StatusApproved:   {StatusInProgress, StatusCancelled},
		StatusRejected:   {StatusPending}, // Can resubmit
		StatusInProgress: {StatusCompleted, StatusCancelled},
		StatusCompleted:  {}, // Final state
		StatusCancelled:  {}, // Final state
	}

	allowed, ok := transitions[from]
	if !ok {
		return false
	}

	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}


// Package repository handles data access for Notification Service.
package repository

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationType represents the type of notification.
type NotificationType string

const (
	NotificationTypeRequestStatus     NotificationType = "request_status"
	NotificationTypeComplianceWarning NotificationType = "compliance_warning"
	NotificationTypeWorkspaceInvite   NotificationType = "workspace_invite"
	NotificationTypeAIComplete        NotificationType = "ai_generation_complete"
	NotificationTypeSystem            NotificationType = "system"
)

// Notification represents a notification in the database.
type Notification struct {
	ID        uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID        `gorm:"type:uuid;not null;index"`
	Type      NotificationType `gorm:"size:50;not null;index"`
	Title     string           `gorm:"size:255;not null"`
	Message   string           `gorm:"size:2000;not null"`
	Data      NotificationData `gorm:"type:jsonb;default:'{}'"`
	Read      bool             `gorm:"default:false;not null;index"`
	ReadAt    *time.Time
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// NotificationData holds additional notification data.
type NotificationData map[string]string

// Scan implements sql.Scanner for JSONB.
func (d *NotificationData) Scan(value interface{}) error {
	if value == nil {
		*d = make(NotificationData)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, d)
}

// Value implements driver.Valuer for JSONB.
func (d NotificationData) Value() (driver.Value, error) {
	if d == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(d)
}

// BeforeCreate generates UUID before insert.
func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

// MarkAsRead marks the notification as read.
func (n *Notification) MarkAsRead() {
	now := time.Now()
	n.Read = true
	n.ReadAt = &now
}

// Migrate runs database migrations.
func Migrate(db *gorm.DB) error {
	// Enable UUID extension
	db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)

	return db.AutoMigrate(&Notification{})
}


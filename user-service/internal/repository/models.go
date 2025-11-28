// Package repository handles data access for User Service.
package repository

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserProfile represents a user profile in the database.
type UserProfile struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key"`
	Email         string         `gorm:"uniqueIndex;not null"`
	Name          string         `gorm:"type:varchar(255);not null"`
	Role          string         `gorm:"type:varchar(50);default:'user';not null"`
	EmailVerified bool           `gorm:"default:false;not null"`
	AvatarURL     *string        `gorm:"type:varchar(255)"`
	Settings      *UserSettings  `gorm:"type:jsonb;default:'{}';not null"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

// UserSettings represents user preferences.
type UserSettings struct {
	Language      string                `json:"language"`
	Theme         string                `json:"theme"`
	Units         string                `json:"units"`
	Notifications *NotificationSettings `json:"notifications"`
}

// NotificationSettings represents notification preferences.
type NotificationSettings struct {
	Email     bool `json:"email"`
	Push      bool `json:"push"`
	Marketing bool `json:"marketing"`
}

// Scan implements sql.Scanner for JSONB.
func (s *UserSettings) Scan(value interface{}) error {
	if value == nil {
		*s = UserSettings{
			Language: "ru",
			Theme:    "system",
			Units:    "metric",
			Notifications: &NotificationSettings{
				Email:     true,
				Push:      true,
				Marketing: false,
			},
		}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, s)
}

// Value implements driver.Valuer for JSONB.
func (s UserSettings) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// DefaultSettings returns default user settings.
func DefaultSettings() *UserSettings {
	return &UserSettings{
		Language: "ru",
		Theme:    "system",
		Units:    "metric",
		Notifications: &NotificationSettings{
			Email:     true,
			Push:      true,
			Marketing: false,
		},
	}
}

// Migrate runs database migrations.
func Migrate(db *gorm.DB) error {
	// Enable UUID extension
	db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)

	return db.AutoMigrate(&UserProfile{})
}


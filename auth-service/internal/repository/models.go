// Package repository handles data access for Auth Service.
package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the database.
type User struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email         string         `gorm:"uniqueIndex;not null"`
	PasswordHash  string         `gorm:"not null"`
	Name          string         `gorm:"type:varchar(255);not null"`
	Role          string         `gorm:"type:varchar(50);default:'user';not null"`
	EmailVerified bool           `gorm:"default:false;not null"`
	AvatarURL     *string        `gorm:"type:varchar(255)"`
	Settings      string         `gorm:"type:jsonb;default:'{}';not null"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

// RefreshToken represents a refresh token in the database.
type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Token     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Revoked   bool      `gorm:"default:false;not null"`
	DeviceID  *string   `gorm:"type:varchar(255)"`
	UserAgent *string   `gorm:"type:varchar(512)"`
	IP        *string   `gorm:"type:varchar(45)"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	User      User      `gorm:"foreignKey:UserID"`
}

// BeforeCreate generates UUID before insert.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if u.Settings == "" {
		u.Settings = `{"language":"ru","theme":"system","units":"metric","notifications":{"email":true,"push":true,"marketing":false}}`
	}
	return nil
}

// BeforeCreate generates UUID before insert.
func (t *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// IsExpired checks if the token is expired.
func (t *RefreshToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsValid checks if the token is valid.
func (t *RefreshToken) IsValid() bool {
	return !t.Revoked && !t.IsExpired()
}

// Migrate runs database migrations.
func Migrate(db *gorm.DB) error {
	// Enable UUID extension
	db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)
	
	return db.AutoMigrate(&User{}, &RefreshToken{})
}


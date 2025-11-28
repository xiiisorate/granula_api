// Package repository handles data access for Workspace Service.
package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Workspace represents a workspace in the database.
type Workspace struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key"`
	Name        string         `gorm:"type:varchar(255);not null"`
	Description string         `gorm:"type:text"`
	OwnerID     uuid.UUID      `gorm:"type:uuid;not null;index"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relations
	Members []WorkspaceMember `gorm:"foreignKey:WorkspaceID"`
}

// WorkspaceMember represents a workspace member in the database.
type WorkspaceMember struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key"`
	WorkspaceID uuid.UUID      `gorm:"type:uuid;not null;index"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null;index"`
	Role        string         `gorm:"type:varchar(50);default:'member';not null"`
	JoinedAt    time.Time      `gorm:"autoCreateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relations
	Workspace Workspace `gorm:"foreignKey:WorkspaceID"`
}

// WorkspaceInvite represents a pending invitation.
type WorkspaceInvite struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key"`
	WorkspaceID uuid.UUID      `gorm:"type:uuid;not null;index"`
	Email       string         `gorm:"type:varchar(255);not null"`
	Role        string         `gorm:"type:varchar(50);default:'member';not null"`
	Token       string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	ExpiresAt   time.Time      `gorm:"not null"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TableName sets the table name for Workspace.
func (Workspace) TableName() string {
	return "workspaces"
}

// TableName sets the table name for WorkspaceMember.
func (WorkspaceMember) TableName() string {
	return "workspace_members"
}

// TableName sets the table name for WorkspaceInvite.
func (WorkspaceInvite) TableName() string {
	return "workspace_invites"
}

// Migrate runs database migrations.
func Migrate(db *gorm.DB) error {
	// Enable UUID extension
	db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)

	return db.AutoMigrate(&Workspace{}, &WorkspaceMember{}, &WorkspaceInvite{})
}

// MemberRole constants
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
	RoleViewer = "viewer"
)

// IsValidRole checks if a role is valid.
func IsValidRole(role string) bool {
	switch role {
	case RoleOwner, RoleAdmin, RoleMember, RoleViewer:
		return true
	default:
		return false
	}
}


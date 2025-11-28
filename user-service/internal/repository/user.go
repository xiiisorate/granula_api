// Package repository handles data access for User Service.
package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository handles user database operations.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByID finds a user profile by ID.
func (r *UserRepository) FindByID(id uuid.UUID) (*UserProfile, error) {
	var profile UserProfile
	err := r.db.Where("id = ?", id).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// Create creates a new user profile.
func (r *UserRepository) Create(profile *UserProfile) error {
	return r.db.Create(profile).Error
}

// Update updates a user profile.
func (r *UserRepository) Update(profile *UserProfile) error {
	return r.db.Save(profile).Error
}

// UpdateName updates user name.
func (r *UserRepository) UpdateName(userID uuid.UUID, name string) error {
	return r.db.Model(&UserProfile{}).Where("id = ?", userID).Update("name", name).Error
}

// UpdateSettings updates user settings.
func (r *UserRepository) UpdateSettings(userID uuid.UUID, settings *UserSettings) error {
	return r.db.Model(&UserProfile{}).Where("id = ?", userID).Update("settings", settings).Error
}

// UpdateAvatar updates user avatar URL.
func (r *UserRepository) UpdateAvatar(userID uuid.UUID, avatarURL *string) error {
	return r.db.Model(&UserProfile{}).Where("id = ?", userID).Update("avatar_url", avatarURL).Error
}

// SoftDelete soft deletes a user profile.
func (r *UserRepository) SoftDelete(userID uuid.UUID) error {
	return r.db.Delete(&UserProfile{}, "id = ?", userID).Error
}

// CreateOrUpdate creates or updates a user profile (for sync from auth service).
func (r *UserRepository) CreateOrUpdate(profile *UserProfile) error {
	return r.db.Save(profile).Error
}


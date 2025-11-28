// Package repository handles data access for Auth Service.
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

// Create creates a new user.
func (r *UserRepository) Create(user *User) error {
	return r.db.Create(user).Error
}

// FindByEmail finds a user by email.
func (r *UserRepository) FindByEmail(email string) (*User, error) {
	var user User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID finds a user by ID.
func (r *UserRepository) FindByID(id uuid.UUID) (*User, error) {
	var user User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// EmailExists checks if email exists.
func (r *UserRepository) EmailExists(email string) bool {
	var count int64
	r.db.Model(&User{}).Where("email = ?", email).Count(&count)
	return count > 0
}

// Update updates a user.
func (r *UserRepository) Update(user *User) error {
	return r.db.Save(user).Error
}

// UpdatePassword updates user password.
func (r *UserRepository) UpdatePassword(userID uuid.UUID, passwordHash string) error {
	return r.db.Model(&User{}).Where("id = ?", userID).Update("password_hash", passwordHash).Error
}

// SoftDelete soft deletes a user.
func (r *UserRepository) SoftDelete(userID uuid.UUID) error {
	return r.db.Delete(&User{}, "id = ?", userID).Error
}


// Package repository handles data access for Auth Service.
package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RefreshTokenRepository handles refresh token database operations.
type RefreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new RefreshTokenRepository.
func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create creates a new refresh token.
func (r *RefreshTokenRepository) Create(token *RefreshToken) error {
	return r.db.Create(token).Error
}

// FindByToken finds a refresh token by token string.
func (r *RefreshTokenRepository) FindByToken(token string) (*RefreshToken, error) {
	var refreshToken RefreshToken
	err := r.db.Preload("User").Where("token = ? AND revoked = ?", token, false).First(&refreshToken).Error
	if err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

// RevokeToken revokes a refresh token.
func (r *RefreshTokenRepository) RevokeToken(token string) error {
	return r.db.Model(&RefreshToken{}).Where("token = ?", token).Update("revoked", true).Error
}

// RevokeByUserID revokes all tokens for a user.
func (r *RefreshTokenRepository) RevokeByUserID(userID uuid.UUID) (int64, error) {
	result := r.db.Model(&RefreshToken{}).
		Where("user_id = ? AND revoked = ?", userID, false).
		Update("revoked", true)
	return result.RowsAffected, result.Error
}

// DeleteByUserID deletes all tokens for a user.
func (r *RefreshTokenRepository) DeleteByUserID(userID uuid.UUID) error {
	return r.db.Delete(&RefreshToken{}, "user_id = ?", userID).Error
}

// DeleteExpiredTokens deletes all expired tokens.
func (r *RefreshTokenRepository) DeleteExpiredTokens() error {
	return r.db.Delete(&RefreshToken{}, "expires_at < NOW()").Error
}


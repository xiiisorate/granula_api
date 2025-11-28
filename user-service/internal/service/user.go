// Package service handles business logic for User Service.
package service

import (
	"strings"

	"github.com/xiiisorate/granula_api/user-service/internal/repository"
	"github.com/xiiisorate/granula_api/shared/pkg/errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserService handles user business logic.
type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService creates a new UserService.
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// GetProfile returns a user profile.
func (s *UserService) GetProfile(userID uuid.UUID) (*repository.UserProfile, error) {
	profile, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.NotFound("user", userID.String())
	}
	return profile, nil
}

// UpdateProfileInput contains profile update data.
type UpdateProfileInput struct {
	Name     *string
	Settings *repository.UserSettings
}

// UpdateProfile updates a user profile.
func (s *UserService) UpdateProfile(userID uuid.UUID, input *UpdateProfileInput) (*repository.UserProfile, error) {
	profile, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.NotFound("user", userID.String())
	}

	// Update name if provided
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if len(name) < 2 || len(name) > 255 {
			return nil, errors.InvalidArgument("name", "must be between 2 and 255 characters")
		}
		profile.Name = name
	}

	// Update settings if provided
	if input.Settings != nil {
		// Validate language
		if input.Settings.Language != "" {
			if !isValidLanguage(input.Settings.Language) {
				return nil, errors.InvalidArgument("language", "invalid language")
			}
			profile.Settings.Language = input.Settings.Language
		}

		// Validate theme
		if input.Settings.Theme != "" {
			if !isValidTheme(input.Settings.Theme) {
				return nil, errors.InvalidArgument("theme", "invalid theme")
			}
			profile.Settings.Theme = input.Settings.Theme
		}

		// Validate units
		if input.Settings.Units != "" {
			if !isValidUnits(input.Settings.Units) {
				return nil, errors.InvalidArgument("units", "invalid units")
			}
			profile.Settings.Units = input.Settings.Units
		}

		// Update notification settings
		if input.Settings.Notifications != nil {
			profile.Settings.Notifications = input.Settings.Notifications
		}
	}

	if err := s.userRepo.Update(profile); err != nil {
		return nil, errors.Internal("failed to update profile").WithCause(err)
	}

	return profile, nil
}

// ChangePasswordInput contains password change data.
type ChangePasswordInput struct {
	UserID          uuid.UUID
	CurrentPassword string
	NewPassword     string
	PasswordHash    string // Current password hash from auth service
}

// ChangePassword validates password change (actual change is in auth service).
func (s *UserService) ChangePassword(input *ChangePasswordInput) error {
	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(input.PasswordHash), []byte(input.CurrentPassword)); err != nil {
		return errors.Unauthenticated("invalid current password")
	}

	// Validate new password
	if len(input.NewPassword) < 8 {
		return errors.InvalidArgument("password", "must be at least 8 characters")
	}

	// Check if new password is same as old
	if err := bcrypt.CompareHashAndPassword([]byte(input.PasswordHash), []byte(input.NewPassword)); err == nil {
		return errors.InvalidArgument("password", "new password must be different")
	}

	return nil
}

// DeleteAccount deletes a user account.
func (s *UserService) DeleteAccount(userID uuid.UUID) error {
	return s.userRepo.SoftDelete(userID)
}

// UpdateAvatar updates user avatar.
func (s *UserService) UpdateAvatar(userID uuid.UUID, avatarURL string) error {
	return s.userRepo.UpdateAvatar(userID, &avatarURL)
}

// DeleteAvatar removes user avatar.
func (s *UserService) DeleteAvatar(userID uuid.UUID) error {
	return s.userRepo.UpdateAvatar(userID, nil)
}

// Helper functions
func isValidLanguage(lang string) bool {
	validLanguages := []string{"ru", "en"}
	for _, v := range validLanguages {
		if v == lang {
			return true
		}
	}
	return false
}

func isValidTheme(theme string) bool {
	validThemes := []string{"light", "dark", "system"}
	for _, v := range validThemes {
		if v == theme {
			return true
		}
	}
	return false
}

func isValidUnits(units string) bool {
	validUnits := []string{"metric", "imperial"}
	for _, v := range validUnits {
		if v == units {
			return true
		}
	}
	return false
}


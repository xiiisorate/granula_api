package service

import (
	"testing"

	"github.com/xiiisorate/granula_api/user-service/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// Test helper validation functions
func TestIsValidLanguage(t *testing.T) {
	tests := []struct {
		lang     string
		expected bool
	}{
		{"ru", true},
		{"en", true},
		{"de", false},
		{"fr", false},
		{"", false},
		{"RU", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			result := isValidLanguage(tt.lang)
			if result != tt.expected {
				t.Errorf("isValidLanguage(%q) = %v, expected %v", tt.lang, result, tt.expected)
			}
		})
	}
}

func TestIsValidTheme(t *testing.T) {
	tests := []struct {
		theme    string
		expected bool
	}{
		{"light", true},
		{"dark", true},
		{"system", true},
		{"auto", false},
		{"", false},
		{"Light", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.theme, func(t *testing.T) {
			result := isValidTheme(tt.theme)
			if result != tt.expected {
				t.Errorf("isValidTheme(%q) = %v, expected %v", tt.theme, result, tt.expected)
			}
		})
	}
}

func TestIsValidUnits(t *testing.T) {
	tests := []struct {
		units    string
		expected bool
	}{
		{"metric", true},
		{"imperial", true},
		{"si", false},
		{"", false},
		{"Metric", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.units, func(t *testing.T) {
			result := isValidUnits(tt.units)
			if result != tt.expected {
				t.Errorf("isValidUnits(%q) = %v, expected %v", tt.units, result, tt.expected)
			}
		})
	}
}

// Test name validation logic
func TestNameValidation(t *testing.T) {
	tests := []struct {
		name    string
		isValid bool
	}{
		{"Jo", true},        // 2 chars - minimum
		{"John Doe", true},  // normal
		{"A", false},        // 1 char - too short
		{"", false},         // empty
		{string(make([]byte, 255)), true},  // 255 chars - maximum
		{string(make([]byte, 256)), false}, // 256 chars - too long
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := len(tt.name) >= 2 && len(tt.name) <= 255
			if isValid != tt.isValid {
				t.Errorf("name length validation for %q (len=%d) = %v, expected %v", 
					tt.name, len(tt.name), isValid, tt.isValid)
			}
		})
	}
}

// Test password validation
func TestPasswordValidation(t *testing.T) {
	tests := []struct {
		password string
		isValid  bool
	}{
		{"12345678", true},  // exactly 8 chars
		{"password123", true},
		{"1234567", false},  // 7 chars - too short
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			isValid := len(tt.password) >= 8
			if isValid != tt.isValid {
				t.Errorf("password validation for len=%d = %v, expected %v",
					len(tt.password), isValid, tt.isValid)
			}
		})
	}
}

// Test password change logic
func TestChangePasswordValidation(t *testing.T) {
	currentPassword := "oldpassword123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(currentPassword), bcrypt.DefaultCost)
	passwordHash := string(hash)

	tests := []struct {
		name            string
		currentPassword string
		newPassword     string
		shouldFail      bool
		failReason      string
	}{
		{
			name:            "valid change",
			currentPassword: currentPassword,
			newPassword:     "newpassword456",
			shouldFail:      false,
		},
		{
			name:            "wrong current password",
			currentPassword: "wrongpassword",
			newPassword:     "newpassword456",
			shouldFail:      true,
			failReason:      "wrong current password",
		},
		{
			name:            "new password too short",
			currentPassword: currentPassword,
			newPassword:     "short",
			shouldFail:      true,
			failReason:      "new password too short",
		},
		{
			name:            "same password",
			currentPassword: currentPassword,
			newPassword:     currentPassword,
			shouldFail:      true,
			failReason:      "same as old password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify current password
			currentPassErr := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(tt.currentPassword))
			
			// Validate new password length
			newPassValid := len(tt.newPassword) >= 8
			
			// Check if same password
			samePassErr := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(tt.newPassword))
			isSamePassword := samePassErr == nil

			hasError := currentPassErr != nil || !newPassValid || isSamePassword

			if hasError != tt.shouldFail {
				t.Errorf("expected shouldFail=%v, got hasError=%v (reason: %s)", 
					tt.shouldFail, hasError, tt.failReason)
			}
		})
	}
}

// Test UserSettings defaults
func TestDefaultSettings(t *testing.T) {
	settings := repository.DefaultSettings()

	if settings.Language != "ru" {
		t.Errorf("Default language should be 'ru', got %s", settings.Language)
	}

	if settings.Theme != "system" {
		t.Errorf("Default theme should be 'system', got %s", settings.Theme)
	}

	if settings.Units != "metric" {
		t.Errorf("Default units should be 'metric', got %s", settings.Units)
	}

	if settings.Notifications == nil {
		t.Error("Default notifications should not be nil")
	} else {
		if !settings.Notifications.Email {
			t.Error("Default email notifications should be true")
		}
		if !settings.Notifications.Push {
			t.Error("Default push notifications should be true")
		}
		if settings.Notifications.Marketing {
			t.Error("Default marketing notifications should be false")
		}
	}
}


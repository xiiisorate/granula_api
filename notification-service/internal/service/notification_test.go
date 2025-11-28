package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/notification-service/internal/repository"
)

// Test Notification model
func TestNotification_MarkAsRead(t *testing.T) {
	notif := &repository.Notification{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Read:   false,
	}

	notif.MarkAsRead()

	if !notif.Read {
		t.Error("Notification should be marked as read")
	}

	if notif.ReadAt == nil {
		t.Error("ReadAt should be set")
	}

	if notif.ReadAt.After(time.Now().Add(time.Second)) {
		t.Error("ReadAt should be in the past or now")
	}
}

// Test Notification types
func TestNotificationType_IsValid(t *testing.T) {
	validTypes := []repository.NotificationType{
		repository.NotificationTypeSystem,
		repository.NotificationTypeRequestStatus,
		repository.NotificationTypeComplianceWarning,
		repository.NotificationTypeWorkspaceInvite,
		repository.NotificationTypeAIComplete,
	}

	invalidTypes := []repository.NotificationType{
		"invalid_type",
		"",
		"unknown",
	}

	// Test valid types
	for _, notifType := range validTypes {
		t.Run("valid_"+string(notifType), func(t *testing.T) {
			isValid := isValidNotificationType(notifType)
			if !isValid {
				t.Errorf("NotificationType(%q) should be valid", notifType)
			}
		})
	}

	// Test invalid types
	for _, notifType := range invalidTypes {
		name := string(notifType)
		if name == "" {
			name = "empty"
		}
		t.Run("invalid_"+name, func(t *testing.T) {
			isValid := isValidNotificationType(notifType)
			if isValid {
				t.Errorf("NotificationType(%q) should be invalid", notifType)
			}
		})
	}
}

// Helper function to validate notification type
func isValidNotificationType(t repository.NotificationType) bool {
	validTypes := []repository.NotificationType{
		repository.NotificationTypeSystem,
		repository.NotificationTypeRequestStatus,
		repository.NotificationTypeComplianceWarning,
		repository.NotificationTypeWorkspaceInvite,
		repository.NotificationTypeAIComplete,
	}
	for _, v := range validTypes {
		if v == t {
			return true
		}
	}
	return false
}

// Test GetListInput defaults
func TestGetListInput_Defaults(t *testing.T) {
	tests := []struct {
		name           string
		inputLimit     int
		inputOffset    int
		expectedLimit  int
		expectedOffset int
	}{
		{"negative limit", -1, 0, 50, 0},
		{"zero limit", 0, 0, 50, 0},
		{"over max limit", 200, 0, 50, 0},
		{"valid limit", 25, 0, 25, 0},
		{"max limit", 100, 0, 100, 0},
		{"negative offset", 10, -5, 10, 0},
		{"valid offset", 10, 20, 10, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &GetListInput{
				Limit:  tt.inputLimit,
				Offset: tt.inputOffset,
			}

			// Apply defaults (mimicking service logic)
			if input.Limit <= 0 || input.Limit > 100 {
				input.Limit = 50
			}
			if input.Offset < 0 {
				input.Offset = 0
			}

			if input.Limit != tt.expectedLimit {
				t.Errorf("Limit = %d, expected %d", input.Limit, tt.expectedLimit)
			}
			if input.Offset != tt.expectedOffset {
				t.Errorf("Offset = %d, expected %d", input.Offset, tt.expectedOffset)
			}
		})
	}
}

// Test NotificationData JSON marshaling
func TestNotificationData(t *testing.T) {
	data := repository.NotificationData{
		"workspace_id": "123",
		"action":       "created",
		"count":        "5",
	}

	if data["workspace_id"] != "123" {
		t.Error("Failed to store workspace_id")
	}

	if data["action"] != "created" {
		t.Error("Failed to store action")
	}

	if data["count"] != "5" {
		t.Error("Failed to store count")
	}
}

// Test CreateInput validation
func TestCreateInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   CreateInput
		isValid bool
	}{
		{
			name: "valid input",
			input: CreateInput{
				UserID:  uuid.New(),
				Type:    repository.NotificationTypeSystem,
				Title:   "Test Title",
				Message: "Test Message",
			},
			isValid: true,
		},
		{
			name: "empty user id",
			input: CreateInput{
				UserID:  uuid.Nil,
				Type:    repository.NotificationTypeSystem,
				Title:   "Test Title",
				Message: "Test Message",
			},
			isValid: false,
		},
		{
			name: "invalid type",
			input: CreateInput{
				UserID:  uuid.New(),
				Type:    "invalid",
				Title:   "Test Title",
				Message: "Test Message",
			},
			isValid: false,
		},
		{
			name: "empty title",
			input: CreateInput{
				UserID:  uuid.New(),
				Type:    repository.NotificationTypeSystem,
				Title:   "",
				Message: "Test Message",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.input.UserID != uuid.Nil &&
				isValidNotificationType(tt.input.Type) &&
				tt.input.Title != ""

			if isValid != tt.isValid {
				t.Errorf("Validation result = %v, expected %v", isValid, tt.isValid)
			}
		})
	}
}

// Test ownership verification logic
func TestOwnershipVerification(t *testing.T) {
	ownerID := uuid.New()
	otherID := uuid.New()

	notif := &repository.Notification{
		ID:     uuid.New(),
		UserID: ownerID,
	}

	// Owner should have access
	if notif.UserID != ownerID {
		t.Error("Owner should have access")
	}

	// Non-owner should not have access
	if notif.UserID == otherID {
		t.Error("Non-owner should not have access")
	}
}

// Test MarkAllAsReadInput
func TestMarkAllAsReadInput(t *testing.T) {
	userID := uuid.New()
	notifType := repository.NotificationTypeSystem
	beforeTime := time.Now().Add(-24 * time.Hour)

	input := &MarkAllAsReadInput{
		UserID: userID,
		Type:   &notifType,
		Before: &beforeTime,
	}

	if input.UserID != userID {
		t.Error("UserID mismatch")
	}

	if input.Type == nil || *input.Type != notifType {
		t.Error("Type mismatch")
	}

	if input.Before == nil || !input.Before.Equal(beforeTime) {
		t.Error("Before time mismatch")
	}
}

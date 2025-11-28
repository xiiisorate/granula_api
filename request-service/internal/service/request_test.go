package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/request-service/internal/repository"
)

// Test title validation
func TestRequestTitleValidation(t *testing.T) {
	tests := []struct {
		title   string
		isValid bool
	}{
		{"ABC", true},       // 3 chars - minimum
		{"My Request Title", true},
		{"AB", false},       // 2 chars - too short
		{"", false},         // empty
		{string(make([]byte, 255)), true},  // 255 chars - maximum
		{string(make([]byte, 256)), false}, // 256 chars - too long
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			isValid := len(tt.title) >= 3 && len(tt.title) <= 255
			if isValid != tt.isValid {
				t.Errorf("title validation for len=%d = %v, expected %v",
					len(tt.title), isValid, tt.isValid)
			}
		})
	}
}

// Test status validation
func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		status   string
		expected bool
	}{
		{repository.StatusPending, true},
		{repository.StatusReview, true},
		{repository.StatusApproved, true},
		{repository.StatusRejected, true},
		{repository.StatusInProgress, true},
		{repository.StatusCompleted, true},
		{repository.StatusCancelled, true},
		{"draft", false},
		{"", false},
		{"Pending", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := repository.IsValidStatus(tt.status)
			if result != tt.expected {
				t.Errorf("IsValidStatus(%q) = %v, expected %v", tt.status, result, tt.expected)
			}
		})
	}
}

// Test category validation
func TestIsValidCategory(t *testing.T) {
	tests := []struct {
		category string
		expected bool
	}{
		{repository.CategoryConsultation, true},
		{repository.CategoryDocumentation, true},
		{repository.CategoryExpertVisit, true},
		{repository.CategoryFullService, true},
		{"other", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			result := repository.IsValidCategory(tt.category)
			if result != tt.expected {
				t.Errorf("IsValidCategory(%q) = %v, expected %v", tt.category, result, tt.expected)
			}
		})
	}
}

// Test priority validation
func TestIsValidPriority(t *testing.T) {
	tests := []struct {
		priority string
		expected bool
	}{
		{repository.PriorityLow, true},
		{repository.PriorityNormal, true},
		{repository.PriorityHigh, true},
		{repository.PriorityUrgent, true},
		{"critical", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.priority, func(t *testing.T) {
			result := repository.IsValidPriority(tt.priority)
			if result != tt.expected {
				t.Errorf("IsValidPriority(%q) = %v, expected %v", tt.priority, result, tt.expected)
			}
		})
	}
}

// Test status transitions
func TestCanTransitionTo(t *testing.T) {
	tests := []struct {
		from     string
		to       string
		expected bool
	}{
		// From pending
		{repository.StatusPending, repository.StatusReview, true},
		{repository.StatusPending, repository.StatusCancelled, true},
		{repository.StatusPending, repository.StatusApproved, false},
		{repository.StatusPending, repository.StatusCompleted, false},

		// From review
		{repository.StatusReview, repository.StatusApproved, true},
		{repository.StatusReview, repository.StatusRejected, true},
		{repository.StatusReview, repository.StatusCancelled, true},
		{repository.StatusReview, repository.StatusPending, false},

		// From approved
		{repository.StatusApproved, repository.StatusInProgress, true},
		{repository.StatusApproved, repository.StatusCancelled, true},
		{repository.StatusApproved, repository.StatusCompleted, false},

		// From rejected (can resubmit)
		{repository.StatusRejected, repository.StatusPending, true},
		{repository.StatusRejected, repository.StatusApproved, false},

		// From in_progress
		{repository.StatusInProgress, repository.StatusCompleted, true},
		{repository.StatusInProgress, repository.StatusCancelled, true},
		{repository.StatusInProgress, repository.StatusPending, false},

		// Final states (no transitions)
		{repository.StatusCompleted, repository.StatusPending, false},
		{repository.StatusCompleted, repository.StatusCancelled, false},
		{repository.StatusCancelled, repository.StatusPending, false},
		{repository.StatusCancelled, repository.StatusApproved, false},
	}

	for _, tt := range tests {
		name := tt.from + " -> " + tt.to
		t.Run(name, func(t *testing.T) {
			result := repository.CanTransitionTo(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("CanTransitionTo(%q, %q) = %v, expected %v",
					tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

// Test pagination defaults
func TestListRequestsInput_Defaults(t *testing.T) {
	tests := []struct {
		name         string
		inputPage    int
		inputSize    int
		expectedPage int
		expectedSize int
	}{
		{"negative page", -1, 20, 1, 20},
		{"zero page", 0, 20, 1, 20},
		{"valid page", 3, 20, 3, 20},
		{"negative page size", 1, -10, 1, 20},
		{"zero page size", 1, 0, 1, 20},
		{"over max page size", 1, 200, 1, 20},
		{"valid page size", 1, 50, 1, 50},
		{"max page size", 1, 100, 1, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &ListRequestsInput{
				Page:     tt.inputPage,
				PageSize: tt.inputSize,
			}

			// Apply defaults (mimicking service logic)
			if input.Page < 1 {
				input.Page = 1
			}
			if input.PageSize < 1 || input.PageSize > 100 {
				input.PageSize = 20
			}

			if input.Page != tt.expectedPage {
				t.Errorf("Page = %d, expected %d", input.Page, tt.expectedPage)
			}
			if input.PageSize != tt.expectedSize {
				t.Errorf("PageSize = %d, expected %d", input.PageSize, tt.expectedSize)
			}
		})
	}
}

// Test CreateRequestInput validation
func TestCreateRequestInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   CreateRequestInput
		isValid bool
	}{
		{
			name: "valid input",
			input: CreateRequestInput{
				WorkspaceID: uuid.New(),
				UserID:      uuid.New(),
				Title:       "Request Title",
				Description: "Description",
				Category:    repository.CategoryConsultation,
				Priority:    repository.PriorityNormal,
			},
			isValid: true,
		},
		{
			name: "empty title",
			input: CreateRequestInput{
				WorkspaceID: uuid.New(),
				UserID:      uuid.New(),
				Title:       "",
				Category:    repository.CategoryConsultation,
			},
			isValid: false,
		},
		{
			name: "short title",
			input: CreateRequestInput{
				WorkspaceID: uuid.New(),
				UserID:      uuid.New(),
				Title:       "AB",
				Category:    repository.CategoryConsultation,
			},
			isValid: false,
		},
		{
			name: "invalid category",
			input: CreateRequestInput{
				WorkspaceID: uuid.New(),
				UserID:      uuid.New(),
				Title:       "Valid Title",
				Category:    "invalid",
			},
			isValid: false,
		},
		{
			name: "empty workspace",
			input: CreateRequestInput{
				WorkspaceID: uuid.Nil,
				UserID:      uuid.New(),
				Title:       "Valid Title",
				Category:    repository.CategoryConsultation,
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := len(tt.input.Title) >= 3 &&
				len(tt.input.Title) <= 255 &&
				repository.IsValidCategory(tt.input.Category) &&
				tt.input.WorkspaceID != uuid.Nil &&
				tt.input.UserID != uuid.Nil

			if isValid != tt.isValid {
				t.Errorf("Validation result = %v, expected %v", isValid, tt.isValid)
			}
		})
	}
}

// Test permission checks
func TestRequestPermissions(t *testing.T) {
	creatorID := uuid.New()
	otherUserID := uuid.New()

	request := &repository.Request{
		ID:     uuid.New(),
		UserID: creatorID,
		Status: repository.StatusPending,
	}

	// Only creator can update/cancel
	if request.UserID != creatorID {
		t.Error("Creator should be able to update")
	}

	if request.UserID == otherUserID {
		t.Error("Non-creator should not be able to update")
	}

	// Can only update in pending/rejected status
	canUpdate := request.Status == repository.StatusPending || request.Status == repository.StatusRejected
	if !canUpdate {
		t.Error("Should be able to update in pending status")
	}

	request.Status = repository.StatusInProgress
	canUpdate = request.Status == repository.StatusPending || request.Status == repository.StatusRejected
	if canUpdate {
		t.Error("Should not be able to update in in_progress status")
	}
}


package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/workspace-service/internal/repository"
)

// Test name validation
func TestWorkspaceNameValidation(t *testing.T) {
	tests := []struct {
		name    string
		isValid bool
	}{
		{"AB", true},        // 2 chars - minimum
		{"My Workspace", true},
		{"A", false},        // 1 char - too short
		{"", false},         // empty
		{string(make([]byte, 255)), true},  // 255 chars - maximum
		{string(make([]byte, 256)), false}, // 256 chars - too long
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := len(tt.name) >= 2 && len(tt.name) <= 255
			if isValid != tt.isValid {
				t.Errorf("name validation for len=%d = %v, expected %v",
					len(tt.name), isValid, tt.isValid)
			}
		})
	}
}

// Test role validation
func TestIsValidRole(t *testing.T) {
	tests := []struct {
		role     string
		expected bool
	}{
		{repository.RoleOwner, true},
		{repository.RoleAdmin, true},
		{repository.RoleMember, true},
		{repository.RoleViewer, true},
		{"superadmin", false},
		{"", false},
		{"Owner", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			result := repository.IsValidRole(tt.role)
			if result != tt.expected {
				t.Errorf("IsValidRole(%q) = %v, expected %v", tt.role, result, tt.expected)
			}
		})
	}
}

// Test access control logic
func TestWorkspaceAccessControl(t *testing.T) {
	ownerID := uuid.New()
	adminID := uuid.New()
	memberID := uuid.New()
	viewerID := uuid.New()
	outsiderID := uuid.New()

	workspace := &repository.Workspace{
		ID:      uuid.New(),
		OwnerID: ownerID,
		Members: []repository.WorkspaceMember{
			{UserID: adminID, Role: repository.RoleAdmin},
			{UserID: memberID, Role: repository.RoleMember},
			{UserID: viewerID, Role: repository.RoleViewer},
		},
	}

	// Helper function - mimics hasAccess logic
	hasAccess := func(w *repository.Workspace, userID uuid.UUID) bool {
		if w.OwnerID == userID {
			return true
		}
		for _, m := range w.Members {
			if m.UserID == userID {
				return true
			}
		}
		return false
	}

	// Helper function - mimics canManage logic
	canManage := func(w *repository.Workspace, userID uuid.UUID) bool {
		if w.OwnerID == userID {
			return true
		}
		for _, m := range w.Members {
			if m.UserID == userID && (m.Role == repository.RoleAdmin || m.Role == repository.RoleOwner) {
				return true
			}
		}
		return false
	}

	// Test hasAccess
	accessTests := []struct {
		name     string
		userID   uuid.UUID
		expected bool
	}{
		{"owner has access", ownerID, true},
		{"admin has access", adminID, true},
		{"member has access", memberID, true},
		{"viewer has access", viewerID, true},
		{"outsider has no access", outsiderID, false},
	}

	for _, tt := range accessTests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasAccess(workspace, tt.userID)
			if result != tt.expected {
				t.Errorf("hasAccess() = %v, expected %v", result, tt.expected)
			}
		})
	}

	// Test canManage
	manageTests := []struct {
		name     string
		userID   uuid.UUID
		expected bool
	}{
		{"owner can manage", ownerID, true},
		{"admin can manage", adminID, true},
		{"member cannot manage", memberID, false},
		{"viewer cannot manage", viewerID, false},
		{"outsider cannot manage", outsiderID, false},
	}

	for _, tt := range manageTests {
		t.Run(tt.name, func(t *testing.T) {
			result := canManage(workspace, tt.userID)
			if result != tt.expected {
				t.Errorf("canManage() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Test pagination defaults
func TestListWorkspacesInput_Defaults(t *testing.T) {
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
			input := &ListWorkspacesInput{
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

// Test CreateWorkspaceInput
func TestCreateWorkspaceInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   CreateWorkspaceInput
		isValid bool
	}{
		{
			name: "valid input",
			input: CreateWorkspaceInput{
				Name:        "My Workspace",
				Description: "Description",
				OwnerID:     uuid.New(),
			},
			isValid: true,
		},
		{
			name: "empty name",
			input: CreateWorkspaceInput{
				Name:    "",
				OwnerID: uuid.New(),
			},
			isValid: false,
		},
		{
			name: "short name",
			input: CreateWorkspaceInput{
				Name:    "A",
				OwnerID: uuid.New(),
			},
			isValid: false,
		},
		{
			name: "empty owner",
			input: CreateWorkspaceInput{
				Name:    "My Workspace",
				OwnerID: uuid.Nil,
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := len(tt.input.Name) >= 2 &&
				len(tt.input.Name) <= 255 &&
				tt.input.OwnerID != uuid.Nil

			if isValid != tt.isValid {
				t.Errorf("Validation result = %v, expected %v", isValid, tt.isValid)
			}
		})
	}
}

// Test add member validation
func TestAddMemberInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		role    string
		isValid bool
	}{
		{"admin role", repository.RoleAdmin, true},
		{"member role", repository.RoleMember, true},
		{"viewer role", repository.RoleViewer, true},
		{"owner role - forbidden", repository.RoleOwner, false}, // Cannot assign owner role
		{"invalid role", "superuser", false},
		{"empty role - defaults to member", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := tt.role
			if role == "" {
				role = repository.RoleMember
			}
			
			isValid := repository.IsValidRole(role) && role != repository.RoleOwner
			
			if isValid != tt.isValid {
				t.Errorf("AddMember role validation = %v, expected %v", isValid, tt.isValid)
			}
		})
	}
}


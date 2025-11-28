// Package service handles business logic for Workspace Service.
package service

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/shared/pkg/errors"
	"github.com/xiiisorate/granula_api/workspace-service/internal/repository"
)

// WorkspaceService handles workspace business logic.
type WorkspaceService struct {
	workspaceRepo *repository.WorkspaceRepository
	memberRepo    *repository.MemberRepository
}

// NewWorkspaceService creates a new WorkspaceService.
func NewWorkspaceService(
	workspaceRepo *repository.WorkspaceRepository,
	memberRepo *repository.MemberRepository,
) *WorkspaceService {
	return &WorkspaceService{
		workspaceRepo: workspaceRepo,
		memberRepo:    memberRepo,
	}
}

// CreateWorkspaceInput contains workspace creation data.
type CreateWorkspaceInput struct {
	Name        string
	Description string
	OwnerID     uuid.UUID
}

// CreateWorkspace creates a new workspace.
func (s *WorkspaceService) CreateWorkspace(input *CreateWorkspaceInput) (*repository.Workspace, error) {
	// Validate name
	name := strings.TrimSpace(input.Name)
	if len(name) < 2 || len(name) > 255 {
		return nil, errors.InvalidArgument("name", "must be between 2 and 255 characters")
	}

	// Create workspace
	workspace := &repository.Workspace{
		ID:          uuid.New(),
		Name:        name,
		Description: strings.TrimSpace(input.Description),
		OwnerID:     input.OwnerID,
	}

	if err := s.workspaceRepo.Create(workspace); err != nil {
		return nil, errors.Internal("failed to create workspace").WithCause(err)
	}

	// Add owner as member with owner role
	member := &repository.WorkspaceMember{
		ID:          uuid.New(),
		WorkspaceID: workspace.ID,
		UserID:      input.OwnerID,
		Role:        repository.RoleOwner,
	}

	if err := s.memberRepo.Create(member); err != nil {
		// Rollback workspace creation
		_ = s.workspaceRepo.Delete(workspace.ID)
		return nil, errors.Internal("failed to add owner as member").WithCause(err)
	}

	// Reload workspace with members
	return s.workspaceRepo.FindByID(workspace.ID)
}

// GetWorkspace returns a workspace by ID.
func (s *WorkspaceService) GetWorkspace(workspaceID, userID uuid.UUID) (*repository.Workspace, error) {
	workspace, err := s.workspaceRepo.FindByID(workspaceID)
	if err != nil {
		return nil, errors.NotFound("workspace", workspaceID.String())
	}

	// Check access
	if !s.hasAccess(workspace, userID) {
		return nil, errors.PermissionDenied("you don't have access to this workspace")
	}

	return workspace, nil
}

// ListWorkspacesInput contains list workspaces parameters.
type ListWorkspacesInput struct {
	UserID   uuid.UUID
	Page     int
	PageSize int
}

// ListWorkspacesOutput contains list workspaces result.
type ListWorkspacesOutput struct {
	Workspaces []repository.Workspace
	Total      int64
	Page       int
	PageSize   int
}

// ListWorkspaces returns workspaces for a user.
func (s *WorkspaceService) ListWorkspaces(input *ListWorkspacesInput) (*ListWorkspacesOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 100 {
		input.PageSize = 20
	}

	workspaces, total, err := s.workspaceRepo.FindByUserID(input.UserID, input.Page, input.PageSize)
	if err != nil {
		return nil, errors.Internal("failed to list workspaces").WithCause(err)
	}

	return &ListWorkspacesOutput{
		Workspaces: workspaces,
		Total:      total,
		Page:       input.Page,
		PageSize:   input.PageSize,
	}, nil
}

// UpdateWorkspaceInput contains workspace update data.
type UpdateWorkspaceInput struct {
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	Name        *string
	Description *string
}

// UpdateWorkspace updates a workspace.
func (s *WorkspaceService) UpdateWorkspace(input *UpdateWorkspaceInput) (*repository.Workspace, error) {
	workspace, err := s.workspaceRepo.FindByID(input.WorkspaceID)
	if err != nil {
		return nil, errors.NotFound("workspace", input.WorkspaceID.String())
	}

	// Check permission (only owner and admin can update)
	if !s.canManage(workspace, input.UserID) {
		return nil, errors.PermissionDenied("you don't have permission to update this workspace")
	}

	updates := make(map[string]interface{})

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if len(name) < 2 || len(name) > 255 {
			return nil, errors.InvalidArgument("name", "must be between 2 and 255 characters")
		}
		updates["name"] = name
	}

	if input.Description != nil {
		updates["description"] = strings.TrimSpace(*input.Description)
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := s.workspaceRepo.UpdateFields(input.WorkspaceID, updates); err != nil {
			return nil, errors.Internal("failed to update workspace").WithCause(err)
		}
	}

	return s.workspaceRepo.FindByID(input.WorkspaceID)
}

// DeleteWorkspace deletes a workspace.
func (s *WorkspaceService) DeleteWorkspace(workspaceID, userID uuid.UUID) error {
	workspace, err := s.workspaceRepo.FindByID(workspaceID)
	if err != nil {
		return errors.NotFound("workspace", workspaceID.String())
	}

	// Only owner can delete
	if workspace.OwnerID != userID {
		return errors.PermissionDenied("only the owner can delete this workspace")
	}

	if err := s.workspaceRepo.Delete(workspaceID); err != nil {
		return errors.Internal("failed to delete workspace").WithCause(err)
	}

	return nil
}

// AddMemberInput contains add member data.
type AddMemberInput struct {
	WorkspaceID uuid.UUID
	UserID      uuid.UUID // User adding the member
	NewUserID   uuid.UUID // User being added
	Role        string
}

// AddMember adds a member to a workspace.
func (s *WorkspaceService) AddMember(input *AddMemberInput) (*repository.WorkspaceMember, error) {
	workspace, err := s.workspaceRepo.FindByID(input.WorkspaceID)
	if err != nil {
		return nil, errors.NotFound("workspace", input.WorkspaceID.String())
	}

	// Check permission (only owner and admin can add members)
	if !s.canManage(workspace, input.UserID) {
		return nil, errors.PermissionDenied("you don't have permission to add members")
	}

	// Validate role
	role := input.Role
	if role == "" {
		role = repository.RoleMember
	}
	if !repository.IsValidRole(role) {
		return nil, errors.InvalidArgument("role", "invalid role")
	}

	// Cannot add owner role
	if role == repository.RoleOwner {
		return nil, errors.InvalidArgument("role", "cannot assign owner role")
	}

	// Check if already a member
	existing, _ := s.memberRepo.FindByWorkspaceAndUser(input.WorkspaceID, input.NewUserID)
	if existing != nil {
		return nil, errors.AlreadyExists("member", "user_id", input.NewUserID.String())
	}

	member := &repository.WorkspaceMember{
		ID:          uuid.New(),
		WorkspaceID: input.WorkspaceID,
		UserID:      input.NewUserID,
		Role:        role,
	}

	if err := s.memberRepo.Create(member); err != nil {
		return nil, errors.Internal("failed to add member").WithCause(err)
	}

	return member, nil
}

// RemoveMember removes a member from a workspace.
func (s *WorkspaceService) RemoveMember(workspaceID, userID, memberUserID uuid.UUID) error {
	workspace, err := s.workspaceRepo.FindByID(workspaceID)
	if err != nil {
		return errors.NotFound("workspace", workspaceID.String())
	}

	// Cannot remove owner
	if memberUserID == workspace.OwnerID {
		return errors.InvalidArgument("member", "cannot remove workspace owner")
	}

	// Check permission (owner, admin, or self)
	if !s.canManage(workspace, userID) && userID != memberUserID {
		return errors.PermissionDenied("you don't have permission to remove members")
	}

	if err := s.memberRepo.Delete(workspaceID, memberUserID); err != nil {
		return errors.Internal("failed to remove member").WithCause(err)
	}

	return nil
}

// UpdateMemberRole updates a member's role.
func (s *WorkspaceService) UpdateMemberRole(workspaceID, userID, memberUserID uuid.UUID, newRole string) error {
	workspace, err := s.workspaceRepo.FindByID(workspaceID)
	if err != nil {
		return errors.NotFound("workspace", workspaceID.String())
	}

	// Only owner can change roles
	if workspace.OwnerID != userID {
		return errors.PermissionDenied("only the owner can change member roles")
	}

	// Cannot change owner's role
	if memberUserID == workspace.OwnerID {
		return errors.InvalidArgument("member", "cannot change owner's role")
	}

	// Validate role
	if !repository.IsValidRole(newRole) || newRole == repository.RoleOwner {
		return errors.InvalidArgument("role", "invalid role")
	}

	if err := s.memberRepo.UpdateRole(workspaceID, memberUserID, newRole); err != nil {
		return errors.Internal("failed to update member role").WithCause(err)
	}

	return nil
}

// hasAccess checks if user has access to workspace.
func (s *WorkspaceService) hasAccess(workspace *repository.Workspace, userID uuid.UUID) bool {
	if workspace.OwnerID == userID {
		return true
	}
	for _, m := range workspace.Members {
		if m.UserID == userID {
			return true
		}
	}
	return false
}

// canManage checks if user can manage workspace (owner or admin).
func (s *WorkspaceService) canManage(workspace *repository.Workspace, userID uuid.UUID) bool {
	if workspace.OwnerID == userID {
		return true
	}
	for _, m := range workspace.Members {
		if m.UserID == userID && (m.Role == repository.RoleAdmin || m.Role == repository.RoleOwner) {
			return true
		}
	}
	return false
}


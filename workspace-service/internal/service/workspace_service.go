// =============================================================================
// Package service provides business logic for Workspace Service.
// =============================================================================
// This package implements the core business operations for workspace management.
// It coordinates between the repository layer and gRPC handlers, enforcing
// business rules and orchestrating complex operations.
//
// Architecture:
//
//	gRPC Server → Service (this) → Repository → Database
//
// Responsibilities:
//   - Business rule validation
//   - Authorization checks
//   - Transaction coordination
//   - Event publishing
//   - Cross-service communication
//
// =============================================================================
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
	"github.com/xiiisorate/granula_api/workspace-service/internal/domain/entity"
	"github.com/xiiisorate/granula_api/workspace-service/internal/repository/postgres"
)

// =============================================================================
// Service Errors
// =============================================================================

// ServiceError represents a business logic error with additional context.
type ServiceError struct {
	Code    string // Machine-readable error code
	Message string // Human-readable message
	Err     error  // Underlying error (may be nil)
}

func (e *ServiceError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *ServiceError) Unwrap() error {
	return e.Err
}

// Common service errors
var (
	ErrUnauthorized = &ServiceError{Code: "UNAUTHORIZED", Message: "not authorized to perform this action"}
	ErrForbidden    = &ServiceError{Code: "FORBIDDEN", Message: "access denied"}
)

// =============================================================================
// WorkspaceService
// =============================================================================

// WorkspaceService provides business operations for workspace management.
// It encapsulates all business logic and coordinates with the repository layer.
//
// Thread Safety:
//
//	The service is safe for concurrent use. All state is managed by the
//	repository layer through proper database transactions.
//
// Usage Example:
//
//	svc := service.NewWorkspaceService(repo, log)
//	ws, err := svc.CreateWorkspace(ctx, userID, "My Workspace", "Description")
type WorkspaceService struct {
	repo *postgres.WorkspaceRepository
	log  *logger.Logger
}

// NewWorkspaceService creates a new WorkspaceService instance.
//
// Parameters:
//   - repo: PostgreSQL repository for data persistence
//   - log: Logger instance for operational logging
//
// Returns:
//   - *WorkspaceService: Initialized service ready for use
func NewWorkspaceService(repo *postgres.WorkspaceRepository, log *logger.Logger) *WorkspaceService {
	return &WorkspaceService{
		repo: repo,
		log:  log,
	}
}

// =============================================================================
// Workspace CRUD Operations
// =============================================================================

// CreateWorkspace creates a new workspace owned by the specified user.
// The owner is automatically added as the first member with RoleOwner.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - ownerID: UUID of the user creating the workspace
//   - name: Display name (2-100 characters)
//   - description: Optional description (max 1000 characters)
//
// Returns:
//   - *entity.Workspace: Created workspace with ID assigned
//   - error: Validation or database error
//
// Business Rules:
//   - Name must be 2-100 characters
//   - Owner becomes first member with owner role
//   - Workspace is created in a single transaction
//
// Events Published:
//   - workspace.created (TODO: implement event publishing)
func (s *WorkspaceService) CreateWorkspace(ctx context.Context, ownerID uuid.UUID, name, description string) (*entity.Workspace, error) {
	s.log.Info("creating workspace",
		logger.String("owner_id", ownerID.String()),
		logger.String("name", name),
	)

	// Create workspace entity (validates name)
	ws, err := entity.NewWorkspace(ownerID, name)
	if err != nil {
		s.log.Warn("workspace creation failed: invalid name",
			logger.String("name", name),
			logger.Err(err),
		)
		return nil, err
	}

	// Set description
	ws.UpdateDescription(description)

	// Persist to database
	if err := s.repo.Create(ctx, ws); err != nil {
		s.log.Error("failed to create workspace in database",
			logger.String("workspace_id", ws.ID.String()),
			logger.Err(err),
		)
		return nil, fmt.Errorf("create workspace: %w", err)
	}

	s.log.Info("workspace created successfully",
		logger.String("workspace_id", ws.ID.String()),
		logger.String("owner_id", ownerID.String()),
	)

	// TODO: Publish workspace.created event

	return ws, nil
}

// GetWorkspace retrieves a workspace by ID.
// Access control: User must be a member of the workspace.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - workspaceID: UUID of the workspace to retrieve
//   - userID: UUID of the requesting user (for access control)
//
// Returns:
//   - *entity.Workspace: Found workspace
//   - error: entity.ErrWorkspaceNotFound or access denied
func (s *WorkspaceService) GetWorkspace(ctx context.Context, workspaceID, userID uuid.UUID) (*entity.Workspace, error) {
	// Check membership first
	isMember, err := s.repo.IsMember(ctx, workspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		s.log.Warn("access denied: user is not a member",
			logger.String("workspace_id", workspaceID.String()),
			logger.String("user_id", userID.String()),
		)
		return nil, ErrForbidden
	}

	// Get workspace with members
	ws, err := s.repo.GetByIDWithMembers(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	return ws, nil
}

// GetWorkspaceBasic retrieves a workspace without loading members.
// Use this for performance when member details are not needed.
func (s *WorkspaceService) GetWorkspaceBasic(ctx context.Context, workspaceID, userID uuid.UUID) (*entity.Workspace, error) {
	// Check membership
	isMember, err := s.repo.IsMember(ctx, workspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return nil, ErrForbidden
	}

	return s.repo.GetByID(ctx, workspaceID)
}

// ListWorkspaces returns all workspaces where the user is a member.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - userID: UUID of the user
//   - nameFilter: Optional filter by name (case-insensitive contains)
//   - limit: Maximum results (1-100, default 20)
//   - offset: Pagination offset
//
// Returns:
//   - []*entity.Workspace: List of workspaces
//   - int: Total count for pagination
//   - error: Database error
func (s *WorkspaceService) ListWorkspaces(ctx context.Context, userID uuid.UUID, nameFilter string, limit, offset int) ([]*entity.Workspace, int, error) {
	opts := postgres.ListOptions{
		UserID:     userID,
		NameFilter: nameFilter,
		Limit:      limit,
		Offset:     offset,
	}

	return s.repo.List(ctx, opts)
}

// UpdateWorkspace updates workspace name and/or description.
// Access control: Only owner or admin can update.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - workspaceID: UUID of the workspace to update
//   - userID: UUID of the requesting user
//   - name: New name (empty string to keep current)
//   - description: New description (empty string to keep current)
//
// Returns:
//   - *entity.Workspace: Updated workspace
//   - error: Validation, access denied, or database error
func (s *WorkspaceService) UpdateWorkspace(ctx context.Context, workspaceID, userID uuid.UUID, name, description string) (*entity.Workspace, error) {
	s.log.Info("updating workspace",
		logger.String("workspace_id", workspaceID.String()),
		logger.String("user_id", userID.String()),
	)

	// Check authorization
	role, err := s.repo.GetMemberRole(ctx, workspaceID, userID)
	if err != nil {
		if err == entity.ErrMemberNotFound {
			return nil, ErrForbidden
		}
		return nil, fmt.Errorf("check role: %w", err)
	}

	if !role.CanManageMembers() {
		s.log.Warn("access denied: insufficient permissions",
			logger.String("workspace_id", workspaceID.String()),
			logger.String("user_id", userID.String()),
			logger.String("role", string(role)),
		)
		return nil, ErrForbidden
	}

	// Get current workspace
	ws, err := s.repo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if name != "" {
		if err := ws.UpdateName(name); err != nil {
			return nil, err
		}
	}
	if description != "" {
		ws.UpdateDescription(description)
	}

	// Save changes
	if err := s.repo.Update(ctx, ws); err != nil {
		return nil, fmt.Errorf("update workspace: %w", err)
	}

	s.log.Info("workspace updated successfully",
		logger.String("workspace_id", workspaceID.String()),
	)

	return ws, nil
}

// DeleteWorkspace permanently removes a workspace and all its data.
// Access control: Only owner can delete.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - workspaceID: UUID of the workspace to delete
//   - userID: UUID of the requesting user (must be owner)
//
// Returns:
//   - error: Access denied or database error
//
// Warning:
//
//	This operation is irreversible. All associated floor plans, scenes,
//	branches, and other data will be permanently deleted.
func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, workspaceID, userID uuid.UUID) error {
	s.log.Info("deleting workspace",
		logger.String("workspace_id", workspaceID.String()),
		logger.String("user_id", userID.String()),
	)

	// Check authorization (only owner can delete)
	ws, err := s.repo.GetByID(ctx, workspaceID)
	if err != nil {
		return err
	}

	if !ws.IsOwner(userID) {
		s.log.Warn("access denied: only owner can delete workspace",
			logger.String("workspace_id", workspaceID.String()),
			logger.String("user_id", userID.String()),
		)
		return ErrForbidden
	}

	// Delete workspace
	if err := s.repo.Delete(ctx, workspaceID); err != nil {
		s.log.Error("failed to delete workspace",
			logger.String("workspace_id", workspaceID.String()),
			logger.Err(err),
		)
		return fmt.Errorf("delete workspace: %w", err)
	}

	s.log.Info("workspace deleted successfully",
		logger.String("workspace_id", workspaceID.String()),
	)

	// TODO: Publish workspace.deleted event

	return nil
}

// =============================================================================
// Member Management Operations
// =============================================================================

// AddMember adds a user to a workspace with the specified role.
// Access control: Only owner or admin can add members.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - workspaceID: UUID of the workspace
//   - userID: UUID of the requesting user (must be owner/admin)
//   - memberUserID: UUID of the user to add
//   - role: Role to assign (cannot be owner)
//
// Returns:
//   - *entity.Member: Created member
//   - error: Validation, access denied, or already exists
func (s *WorkspaceService) AddMember(ctx context.Context, workspaceID, userID, memberUserID uuid.UUID, role entity.MemberRole) (*entity.Member, error) {
	s.log.Info("adding member to workspace",
		logger.String("workspace_id", workspaceID.String()),
		logger.String("member_user_id", memberUserID.String()),
		logger.String("role", string(role)),
	)

	// Check authorization
	callerRole, err := s.repo.GetMemberRole(ctx, workspaceID, userID)
	if err != nil {
		if err == entity.ErrMemberNotFound {
			return nil, ErrForbidden
		}
		return nil, fmt.Errorf("check role: %w", err)
	}

	if !callerRole.CanManageMembers() {
		return nil, ErrForbidden
	}

	// Validate role
	if role == entity.RoleOwner || !role.IsValid() {
		return nil, entity.ErrInvalidRole
	}

	// Create member
	member := &entity.Member{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		UserID:      memberUserID,
		Role:        role,
		JoinedAt:    timeNow(),
		InvitedBy:   userID,
	}

	// Add to database
	if err := s.repo.AddMember(ctx, member); err != nil {
		if err == postgres.ErrDuplicateMember {
			return nil, entity.ErrMemberAlreadyExists
		}
		return nil, fmt.Errorf("add member: %w", err)
	}

	s.log.Info("member added successfully",
		logger.String("workspace_id", workspaceID.String()),
		logger.String("member_id", member.ID.String()),
	)

	// TODO: Send notification to new member

	return member, nil
}

// RemoveMember removes a user from a workspace.
// Access control: Owner/admin can remove anyone except owner.
// Users can remove themselves (leave workspace).
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - workspaceID: UUID of the workspace
//   - userID: UUID of the requesting user
//   - memberUserID: UUID of the user to remove
//
// Returns:
//   - error: Access denied, cannot remove owner, or not found
func (s *WorkspaceService) RemoveMember(ctx context.Context, workspaceID, userID, memberUserID uuid.UUID) error {
	s.log.Info("removing member from workspace",
		logger.String("workspace_id", workspaceID.String()),
		logger.String("member_user_id", memberUserID.String()),
	)

	// Get workspace to check ownership
	ws, err := s.repo.GetByID(ctx, workspaceID)
	if err != nil {
		return err
	}

	// Cannot remove owner
	if ws.IsOwner(memberUserID) {
		return entity.ErrCannotRemoveOwner
	}

	// Authorization check
	if userID != memberUserID {
		// Removing someone else - need admin/owner permission
		callerRole, err := s.repo.GetMemberRole(ctx, workspaceID, userID)
		if err != nil {
			return ErrForbidden
		}
		if !callerRole.CanManageMembers() {
			return ErrForbidden
		}
	}

	// Remove from database
	if err := s.repo.RemoveMember(ctx, workspaceID, memberUserID); err != nil {
		return fmt.Errorf("remove member: %w", err)
	}

	s.log.Info("member removed successfully",
		logger.String("workspace_id", workspaceID.String()),
		logger.String("member_user_id", memberUserID.String()),
	)

	return nil
}

// UpdateMemberRole changes a member's role in a workspace.
// Access control: Only owner can change roles.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - workspaceID: UUID of the workspace
//   - userID: UUID of the requesting user (must be owner)
//   - memberUserID: UUID of the member to update
//   - role: New role (cannot be owner)
//
// Returns:
//   - *entity.Member: Updated member
//   - error: Access denied, invalid role, or not found
func (s *WorkspaceService) UpdateMemberRole(ctx context.Context, workspaceID, userID, memberUserID uuid.UUID, role entity.MemberRole) (*entity.Member, error) {
	s.log.Info("updating member role",
		logger.String("workspace_id", workspaceID.String()),
		logger.String("member_user_id", memberUserID.String()),
		logger.String("new_role", string(role)),
	)

	// Only owner can change roles
	ws, err := s.repo.GetByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	if !ws.IsOwner(userID) {
		return nil, ErrForbidden
	}

	// Cannot change owner's role or make someone else owner
	if ws.IsOwner(memberUserID) || role == entity.RoleOwner {
		return nil, entity.ErrInvalidRole
	}

	// Validate role
	if !role.IsValid() {
		return nil, entity.ErrInvalidRole
	}

	// Update in database
	if err := s.repo.UpdateMemberRole(ctx, workspaceID, memberUserID, role); err != nil {
		return nil, fmt.Errorf("update role: %w", err)
	}

	s.log.Info("member role updated successfully",
		logger.String("workspace_id", workspaceID.String()),
		logger.String("member_user_id", memberUserID.String()),
	)

	// Return updated member
	member, err := s.repo.GetMember(ctx, workspaceID, memberUserID)
	if err != nil {
		return nil, fmt.Errorf("get updated member: %w", err)
	}

	return member, nil
}

// GetMembers returns all members of a workspace.
// Access control: User must be a member.
func (s *WorkspaceService) GetMembers(ctx context.Context, workspaceID, userID uuid.UUID) ([]*entity.Member, error) {
	// Check membership
	isMember, err := s.repo.IsMember(ctx, workspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return nil, ErrForbidden
	}

	members, err := s.repo.GetMembers(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	// Convert to pointer slice
	result := make([]*entity.Member, len(members))
	for i := range members {
		result[i] = &members[i]
	}
	return result, nil
}

// =============================================================================
// Invite Management Operations
// =============================================================================

// InviteMember creates an invitation for a user to join a workspace.
// Access control: Only owner or admin can invite.
func (s *WorkspaceService) InviteMember(ctx context.Context, workspaceID, inviterID, inviteeID uuid.UUID, role entity.MemberRole) (*entity.WorkspaceInvite, error) {
	s.log.Info("inviting user to workspace",
		logger.String("workspace_id", workspaceID.String()),
		logger.String("invitee_id", inviteeID.String()),
	)

	// Check authorization
	callerRole, err := s.repo.GetMemberRole(ctx, workspaceID, inviterID)
	if err != nil {
		return nil, ErrForbidden
	}
	if !callerRole.CanManageMembers() {
		return nil, ErrForbidden
	}

	// Check if already a member
	isMember, _ := s.repo.IsMember(ctx, workspaceID, inviteeID)
	if isMember {
		return nil, entity.ErrMemberAlreadyExists
	}

	// Check for existing pending invite
	existingInvite, _ := s.repo.GetPendingInvite(ctx, workspaceID, inviteeID)
	if existingInvite != nil {
		return existingInvite, nil // Return existing invite
	}

	// Create invite
	invite := entity.NewWorkspaceInvite(workspaceID, inviteeID, inviterID, role)
	if err := s.repo.CreateInvite(ctx, invite); err != nil {
		return nil, fmt.Errorf("create invite: %w", err)
	}

	s.log.Info("invite created",
		logger.String("invite_id", invite.ID.String()),
	)

	// TODO: Send notification to invitee

	return invite, nil
}

// =============================================================================
// Helper Functions
// =============================================================================

func timeNow() time.Time {
	return time.Now().UTC()
}

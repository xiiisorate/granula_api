// =============================================================================
// Package entity defines domain entities for Workspace Service.
// =============================================================================
// This package contains the core business entities that represent workspaces
// and their members. These entities encapsulate business logic and validation
// rules, ensuring data integrity at the domain level.
//
// Entity Design Principles:
//   - Immutable IDs: UUIDs are assigned at creation and never change
//   - Timestamps: CreatedAt/UpdatedAt are managed automatically
//   - Validation: Business rules are enforced in entity methods
//   - Rich Domain Model: Entities contain behavior, not just data
//
// =============================================================================
package entity

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// Domain Errors
// =============================================================================
// These errors represent business rule violations and should be returned
// when validation fails. They are distinct from infrastructure errors.

var (
	// ErrWorkspaceNotFound is returned when a workspace cannot be found by ID.
	ErrWorkspaceNotFound = errors.New("workspace not found")

	// ErrMemberNotFound is returned when a member cannot be found in a workspace.
	ErrMemberNotFound = errors.New("member not found")

	// ErrMemberAlreadyExists is returned when trying to add an existing member.
	ErrMemberAlreadyExists = errors.New("member already exists in workspace")

	// ErrCannotRemoveOwner is returned when trying to remove the workspace owner.
	ErrCannotRemoveOwner = errors.New("cannot remove workspace owner")

	// ErrInvalidWorkspaceName is returned when workspace name fails validation.
	ErrInvalidWorkspaceName = errors.New("workspace name must be 2-100 characters")

	// ErrInvalidRole is returned when an invalid member role is specified.
	ErrInvalidRole = errors.New("invalid member role")

	// ErrOwnerCannotLeave is returned when owner tries to leave workspace.
	ErrOwnerCannotLeave = errors.New("owner cannot leave workspace, transfer ownership first")
)

// =============================================================================
// Member Role Constants
// =============================================================================

// MemberRole represents the role of a member within a workspace.
// Roles determine permissions and capabilities within the workspace.
type MemberRole string

const (
	// RoleOwner has full control over the workspace including deletion
	// and ownership transfer. Only one owner per workspace.
	RoleOwner MemberRole = "owner"

	// RoleAdmin can manage members, projects, and settings but cannot
	// delete the workspace or transfer ownership.
	RoleAdmin MemberRole = "admin"

	// RoleEditor can create, edit, and delete projects and scenes
	// but cannot manage members or workspace settings.
	RoleEditor MemberRole = "editor"

	// RoleViewer has read-only access to all workspace content.
	// Cannot create or modify any resources.
	RoleViewer MemberRole = "viewer"
)

// ValidRoles contains all valid member roles for validation.
var ValidRoles = []MemberRole{RoleOwner, RoleAdmin, RoleEditor, RoleViewer}

// IsValid checks if the role is one of the predefined valid roles.
func (r MemberRole) IsValid() bool {
	for _, valid := range ValidRoles {
		if r == valid {
			return true
		}
	}
	return false
}

// String returns the string representation of the role.
func (r MemberRole) String() string {
	return string(r)
}

// CanManageMembers returns true if this role has permission to manage members.
func (r MemberRole) CanManageMembers() bool {
	return r == RoleOwner || r == RoleAdmin
}

// CanEditContent returns true if this role has permission to edit content.
func (r MemberRole) CanEditContent() bool {
	return r == RoleOwner || r == RoleAdmin || r == RoleEditor
}

// CanDeleteWorkspace returns true if this role can delete the workspace.
func (r MemberRole) CanDeleteWorkspace() bool {
	return r == RoleOwner
}

// =============================================================================
// Workspace Entity
// =============================================================================

// Workspace represents a collaborative workspace where users can create
// and manage floor plans, scenes, and design variants.
//
// A workspace is the top-level organizational unit in Granula. It contains:
//   - Floor plans (uploaded/recognized)
//   - 3D scenes (derived from floor plans)
//   - Design branches (variations of scenes)
//   - Expert requests (for professional consultation)
//
// Business Rules:
//   - Every workspace has exactly one owner
//   - Name must be 2-100 characters
//   - Description is optional, max 1000 characters
//   - Workspace can have unlimited members
type Workspace struct {
	// ID is the unique identifier for the workspace (UUID v4).
	// Assigned at creation, immutable.
	ID uuid.UUID `json:"id" db:"id"`

	// OwnerID is the UUID of the user who owns this workspace.
	// The owner has full control and cannot be removed.
	OwnerID uuid.UUID `json:"owner_id" db:"owner_id"`

	// Name is the display name of the workspace.
	// Required, 2-100 characters, trimmed of whitespace.
	Name string `json:"name" db:"name"`

	// Description provides additional context about the workspace.
	// Optional, max 1000 characters.
	Description string `json:"description" db:"description"`

	// Members contains all users with access to this workspace.
	// Includes the owner as a member with RoleOwner.
	Members []Member `json:"members" db:"-"`

	// MemberCount is the total number of members (including owner).
	// Used for display without loading all members.
	MemberCount int `json:"member_count" db:"member_count"`

	// ProjectCount is the number of projects/floor plans in this workspace.
	// Cached for quick display in workspace lists.
	ProjectCount int `json:"project_count" db:"project_count"`

	// CreatedAt is the timestamp when the workspace was created.
	// Set automatically, immutable.
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt is the timestamp of the last modification.
	// Updated automatically on any change.
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// NewWorkspace creates a new workspace with the given owner and name.
// The workspace is initialized with the owner as the first member.
//
// Parameters:
//   - ownerID: UUID of the user creating the workspace
//   - name: Display name (will be trimmed and validated)
//
// Returns:
//   - *Workspace: New workspace entity
//   - error: ErrInvalidWorkspaceName if name validation fails
//
// Example:
//
//	ws, err := entity.NewWorkspace(userID, "My Apartment Redesign")
//	if err != nil {
//	    log.Fatal("Invalid workspace name")
//	}
func NewWorkspace(ownerID uuid.UUID, name string) (*Workspace, error) {
	// Trim and validate name
	name = strings.TrimSpace(name)
	if len(name) < 2 || len(name) > 100 {
		return nil, ErrInvalidWorkspaceName
	}

	now := time.Now().UTC()
	ws := &Workspace{
		ID:          uuid.New(),
		OwnerID:     ownerID,
		Name:        name,
		Description: "",
		Members:     make([]Member, 0, 1),
		MemberCount: 1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Add owner as first member
	ws.Members = append(ws.Members, Member{
		ID:          uuid.New(),
		WorkspaceID: ws.ID,
		UserID:      ownerID,
		Role:        RoleOwner,
		JoinedAt:    now,
	})

	return ws, nil
}

// UpdateName updates the workspace name after validation.
// Returns ErrInvalidWorkspaceName if validation fails.
func (w *Workspace) UpdateName(name string) error {
	name = strings.TrimSpace(name)
	if len(name) < 2 || len(name) > 100 {
		return ErrInvalidWorkspaceName
	}
	w.Name = name
	w.UpdatedAt = time.Now().UTC()
	return nil
}

// UpdateDescription updates the workspace description.
// Truncates to 1000 characters if longer.
func (w *Workspace) UpdateDescription(description string) {
	description = strings.TrimSpace(description)
	if len(description) > 1000 {
		description = description[:1000]
	}
	w.Description = description
	w.UpdatedAt = time.Now().UTC()
}

// AddMember adds a new member to the workspace with the specified role.
//
// Parameters:
//   - userID: UUID of the user to add
//   - role: Role to assign (cannot be owner)
//
// Returns:
//   - *Member: The newly created member
//   - error: ErrMemberAlreadyExists if user is already a member
//   - error: ErrInvalidRole if role is invalid or owner
func (w *Workspace) AddMember(userID uuid.UUID, role MemberRole) (*Member, error) {
	// Check if already a member
	if w.HasMember(userID) {
		return nil, ErrMemberAlreadyExists
	}

	// Validate role (cannot add another owner)
	if role == RoleOwner {
		return nil, ErrInvalidRole
	}
	if !role.IsValid() {
		return nil, ErrInvalidRole
	}

	member := Member{
		ID:          uuid.New(),
		WorkspaceID: w.ID,
		UserID:      userID,
		Role:        role,
		JoinedAt:    time.Now().UTC(),
	}

	w.Members = append(w.Members, member)
	w.MemberCount++
	w.UpdatedAt = time.Now().UTC()

	return &member, nil
}

// RemoveMember removes a member from the workspace by user ID.
//
// Parameters:
//   - userID: UUID of the user to remove
//
// Returns:
//   - error: ErrCannotRemoveOwner if trying to remove owner
//   - error: ErrMemberNotFound if user is not a member
func (w *Workspace) RemoveMember(userID uuid.UUID) error {
	// Cannot remove owner
	if userID == w.OwnerID {
		return ErrCannotRemoveOwner
	}

	// Find and remove member
	for i, m := range w.Members {
		if m.UserID == userID {
			w.Members = append(w.Members[:i], w.Members[i+1:]...)
			w.MemberCount--
			w.UpdatedAt = time.Now().UTC()
			return nil
		}
	}

	return ErrMemberNotFound
}

// UpdateMemberRole changes the role of an existing member.
//
// Parameters:
//   - userID: UUID of the member to update
//   - role: New role to assign (cannot be owner)
//
// Returns:
//   - error: ErrMemberNotFound if user is not a member
//   - error: ErrCannotRemoveOwner if trying to change owner's role
//   - error: ErrInvalidRole if role is invalid
func (w *Workspace) UpdateMemberRole(userID uuid.UUID, role MemberRole) error {
	// Cannot change owner's role
	if userID == w.OwnerID {
		return ErrCannotRemoveOwner
	}

	// Validate role
	if role == RoleOwner || !role.IsValid() {
		return ErrInvalidRole
	}

	// Find and update member
	for i := range w.Members {
		if w.Members[i].UserID == userID {
			w.Members[i].Role = role
			w.UpdatedAt = time.Now().UTC()
			return nil
		}
	}

	return ErrMemberNotFound
}

// HasMember returns true if the user is a member of this workspace.
func (w *Workspace) HasMember(userID uuid.UUID) bool {
	for _, m := range w.Members {
		if m.UserID == userID {
			return true
		}
	}
	return false
}

// GetMember returns the member details for a user if they are a member.
// Returns nil if the user is not a member.
func (w *Workspace) GetMember(userID uuid.UUID) *Member {
	for _, m := range w.Members {
		if m.UserID == userID {
			return &m
		}
	}
	return nil
}

// IsOwner returns true if the given user is the workspace owner.
func (w *Workspace) IsOwner(userID uuid.UUID) bool {
	return w.OwnerID == userID
}

// TransferOwnership transfers workspace ownership to another member.
// The previous owner becomes an admin.
//
// Parameters:
//   - newOwnerID: UUID of the member to become the new owner
//
// Returns:
//   - error: ErrMemberNotFound if new owner is not a member
func (w *Workspace) TransferOwnership(newOwnerID uuid.UUID) error {
	// Verify new owner is a member
	if !w.HasMember(newOwnerID) {
		return ErrMemberNotFound
	}

	// Update old owner to admin
	for i := range w.Members {
		if w.Members[i].UserID == w.OwnerID {
			w.Members[i].Role = RoleAdmin
		}
		if w.Members[i].UserID == newOwnerID {
			w.Members[i].Role = RoleOwner
		}
	}

	w.OwnerID = newOwnerID
	w.UpdatedAt = time.Now().UTC()
	return nil
}

// =============================================================================
// Member Entity
// =============================================================================

// Member represents a user's membership in a workspace.
// Each member has a specific role that determines their permissions.
type Member struct {
	// ID is the unique identifier for this membership record.
	ID uuid.UUID `json:"id" db:"id"`

	// WorkspaceID links this membership to a workspace.
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`

	// UserID is the UUID of the user who is a member.
	UserID uuid.UUID `json:"user_id" db:"user_id"`

	// Role determines the member's permissions in the workspace.
	Role MemberRole `json:"role" db:"role"`

	// JoinedAt is when the user became a member.
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`

	// InvitedBy is the UUID of the user who invited this member.
	// Nil for workspace owner.
	InvitedBy *uuid.UUID `json:"invited_by,omitempty" db:"invited_by"`
}

// =============================================================================
// Workspace Invite Entity
// =============================================================================

// InviteStatus represents the current state of a workspace invitation.
type InviteStatus string

const (
	// InviteStatusPending means the invite is awaiting response.
	InviteStatusPending InviteStatus = "pending"

	// InviteStatusAccepted means the user accepted the invite.
	InviteStatusAccepted InviteStatus = "accepted"

	// InviteStatusDeclined means the user declined the invite.
	InviteStatusDeclined InviteStatus = "declined"

	// InviteStatusExpired means the invite has expired.
	InviteStatusExpired InviteStatus = "expired"

	// InviteStatusCancelled means the invite was cancelled by the sender.
	InviteStatusCancelled InviteStatus = "cancelled"
)

// WorkspaceInvite represents an invitation to join a workspace.
// Invites have a limited lifetime and can be accepted/declined.
type WorkspaceInvite struct {
	// ID is the unique identifier for this invitation.
	ID uuid.UUID `json:"id" db:"id"`

	// WorkspaceID is the workspace being invited to.
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`

	// InvitedUserID is the UUID of the user being invited.
	InvitedUserID uuid.UUID `json:"invited_user_id" db:"invited_user_id"`

	// InvitedByUserID is the UUID of the user who sent the invite.
	InvitedByUserID uuid.UUID `json:"invited_by_user_id" db:"invited_by_user_id"`

	// Role is the role that will be assigned when accepting.
	Role MemberRole `json:"role" db:"role"`

	// Status is the current state of the invitation.
	Status InviteStatus `json:"status" db:"status"`

	// ExpiresAt is when the invitation expires.
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`

	// CreatedAt is when the invitation was created.
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// RespondedAt is when the user responded (accepted/declined).
	RespondedAt *time.Time `json:"responded_at,omitempty" db:"responded_at"`
}

// NewWorkspaceInvite creates a new workspace invitation.
// The invite expires after 7 days by default.
func NewWorkspaceInvite(workspaceID, invitedUserID, invitedByUserID uuid.UUID, role MemberRole) *WorkspaceInvite {
	now := time.Now().UTC()
	return &WorkspaceInvite{
		ID:              uuid.New(),
		WorkspaceID:     workspaceID,
		InvitedUserID:   invitedUserID,
		InvitedByUserID: invitedByUserID,
		Role:            role,
		Status:          InviteStatusPending,
		ExpiresAt:       now.Add(7 * 24 * time.Hour), // 7 days
		CreatedAt:       now,
	}
}

// IsExpired returns true if the invitation has expired.
func (i *WorkspaceInvite) IsExpired() bool {
	return time.Now().UTC().After(i.ExpiresAt)
}

// IsPending returns true if the invitation is still pending.
func (i *WorkspaceInvite) IsPending() bool {
	return i.Status == InviteStatusPending && !i.IsExpired()
}

// Accept marks the invitation as accepted.
func (i *WorkspaceInvite) Accept() {
	now := time.Now().UTC()
	i.Status = InviteStatusAccepted
	i.RespondedAt = &now
}

// Decline marks the invitation as declined.
func (i *WorkspaceInvite) Decline() {
	now := time.Now().UTC()
	i.Status = InviteStatusDeclined
	i.RespondedAt = &now
}

// Cancel marks the invitation as cancelled.
func (i *WorkspaceInvite) Cancel() {
	now := time.Now().UTC()
	i.Status = InviteStatusCancelled
	i.RespondedAt = &now
}

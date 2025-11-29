// =============================================================================
// Package postgres provides PostgreSQL implementations of repository interfaces.
// =============================================================================
// This package contains the data access layer for workspace-related entities.
// It uses the pgx driver for high-performance PostgreSQL operations and
// implements proper connection pooling, prepared statements, and error handling.
//
// Design Decisions:
//   - Uses pgxpool for connection management
//   - Supports transactions for multi-step operations
//   - Returns domain errors for business logic failures
//   - Implements pagination for list operations
//
// =============================================================================
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xiiisorate/granula_api/workspace-service/internal/domain/entity"
)

// =============================================================================
// Repository Errors
// =============================================================================

var (
	// ErrDuplicateWorkspace is returned when creating a workspace with duplicate name for same owner.
	ErrDuplicateWorkspace = errors.New("workspace with this name already exists")

	// ErrDuplicateMember is returned when adding a member who is already in the workspace.
	ErrDuplicateMember = errors.New("user is already a member of this workspace")
)

// =============================================================================
// WorkspaceRepository
// =============================================================================

// WorkspaceRepository provides data access operations for workspaces.
// It handles CRUD operations and member management in PostgreSQL.
//
// The repository uses a connection pool and all methods accept a context
// for cancellation and timeout support.
//
// Thread Safety:
//
//	The repository is safe for concurrent use as it uses a connection pool.
//	Each method acquires a connection from the pool, executes the query,
//	and returns the connection to the pool.
type WorkspaceRepository struct {
	// pool is the database connection pool.
	// Must be initialized before use.
	pool *pgxpool.Pool
}

// NewWorkspaceRepository creates a new WorkspaceRepository with the given connection pool.
//
// Parameters:
//   - pool: PostgreSQL connection pool (must not be nil)
//
// Returns:
//   - *WorkspaceRepository: Initialized repository ready for use
//
// Example:
//
//	pool, _ := pgxpool.New(ctx, dsn)
//	repo := postgres.NewWorkspaceRepository(pool)
func NewWorkspaceRepository(pool *pgxpool.Pool) *WorkspaceRepository {
	return &WorkspaceRepository{pool: pool}
}

// =============================================================================
// Workspace CRUD Operations
// =============================================================================

// Create inserts a new workspace and its owner as the first member.
// This operation is transactional - if either insert fails, both are rolled back.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - workspace: Workspace entity to create (ID should be pre-assigned)
//
// Returns:
//   - error: Database or constraint violation error
//
// Transaction:
//
//  1. INSERT INTO workspaces
//  2. INSERT INTO workspace_members (owner)
//  3. COMMIT or ROLLBACK
func (r *WorkspaceRepository) Create(ctx context.Context, workspace *entity.Workspace) error {
	// Start transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // Rollback if not committed

	// Insert workspace
	query := `
		INSERT INTO workspaces (id, owner_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.Exec(ctx, query,
		workspace.ID,
		workspace.OwnerID,
		workspace.Name,
		workspace.Description,
		workspace.CreatedAt,
		workspace.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert workspace: %w", err)
	}

	// Insert owner as first member
	memberQuery := `
		INSERT INTO workspace_members (id, workspace_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	if len(workspace.Members) > 0 {
		owner := workspace.Members[0]
		_, err = tx.Exec(ctx, memberQuery,
			owner.ID,
			owner.WorkspaceID,
			owner.UserID,
			owner.Role,
			owner.JoinedAt,
		)
		if err != nil {
			return fmt.Errorf("insert owner member: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a workspace by its unique identifier.
// Members are NOT loaded by default - use GetByIDWithMembers for that.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - id: Workspace UUID
//
// Returns:
//   - *entity.Workspace: Found workspace entity
//   - error: entity.ErrWorkspaceNotFound if not found, or database error
func (r *WorkspaceRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Workspace, error) {
	query := `
		SELECT 
			id, owner_id, name, description, created_at, updated_at,
			(SELECT COUNT(*) FROM workspace_members WHERE workspace_id = w.id) as member_count,
			(SELECT COUNT(*) FROM floor_plans WHERE workspace_id = w.id) as project_count
		FROM workspaces w
		WHERE id = $1
	`

	var ws entity.Workspace
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&ws.ID,
		&ws.OwnerID,
		&ws.Name,
		&ws.Description,
		&ws.CreatedAt,
		&ws.UpdatedAt,
		&ws.MemberCount,
		&ws.ProjectCount,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrWorkspaceNotFound
		}
		return nil, fmt.Errorf("query workspace: %w", err)
	}

	return &ws, nil
}

// GetByIDWithMembers retrieves a workspace with all its members loaded.
// Use this when you need member information; otherwise use GetByID.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - id: Workspace UUID
//
// Returns:
//   - *entity.Workspace: Found workspace with Members populated
//   - error: entity.ErrWorkspaceNotFound if not found
func (r *WorkspaceRepository) GetByIDWithMembers(ctx context.Context, id uuid.UUID) (*entity.Workspace, error) {
	// First get the workspace
	ws, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Then load members
	members, err := r.GetMembers(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load members: %w", err)
	}
	ws.Members = members

	return ws, nil
}

// Update saves changes to an existing workspace.
// Only Name and Description can be updated. Other fields are ignored.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - workspace: Workspace with updated fields
//
// Returns:
//   - error: entity.ErrWorkspaceNotFound if workspace doesn't exist
func (r *WorkspaceRepository) Update(ctx context.Context, workspace *entity.Workspace) error {
	query := `
		UPDATE workspaces 
		SET name = $2, description = $3, updated_at = $4
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		workspace.ID,
		workspace.Name,
		workspace.Description,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update workspace: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrWorkspaceNotFound
	}

	return nil
}

// Delete removes a workspace and all associated data.
// This includes members, invites, and any cascade-deleted related entities.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - id: Workspace UUID to delete
//
// Returns:
//   - error: entity.ErrWorkspaceNotFound if workspace doesn't exist
//
// Warning:
//
//	This operation is destructive and cannot be undone.
//	Consider implementing soft delete for production use.
func (r *WorkspaceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Start transaction for cascading deletes
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete members first (foreign key constraint)
	_, err = tx.Exec(ctx, "DELETE FROM workspace_members WHERE workspace_id = $1", id)
	if err != nil {
		return fmt.Errorf("delete members: %w", err)
	}

	// Delete invites
	_, err = tx.Exec(ctx, "DELETE FROM workspace_invites WHERE workspace_id = $1", id)
	if err != nil {
		return fmt.Errorf("delete invites: %w", err)
	}

	// Delete workspace
	result, err := tx.Exec(ctx, "DELETE FROM workspaces WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete workspace: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrWorkspaceNotFound
	}

	return tx.Commit(ctx)
}

// =============================================================================
// List and Search Operations
// =============================================================================

// ListOptions defines parameters for listing workspaces.
type ListOptions struct {
	// UserID filters to workspaces where this user is a member.
	// Required for non-admin queries.
	UserID uuid.UUID

	// NameFilter filters workspaces by name (case-insensitive contains).
	// Optional.
	NameFilter string

	// Limit is the maximum number of results to return.
	// Default: 20, Max: 100
	Limit int

	// Offset is the number of results to skip for pagination.
	// Default: 0
	Offset int
}

// List returns workspaces matching the given options.
// Results are ordered by updated_at descending (most recent first).
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - opts: Filtering and pagination options
//
// Returns:
//   - []*entity.Workspace: List of matching workspaces
//   - int: Total count of matching workspaces (for pagination)
//   - error: Database error
func (r *WorkspaceRepository) List(ctx context.Context, opts ListOptions) ([]*entity.Workspace, int, error) {
	// Validate and apply defaults
	if opts.Limit <= 0 || opts.Limit > 100 {
		opts.Limit = 20
	}

	// Build query
	query := `
		SELECT 
			w.id, w.owner_id, w.name, w.description, w.created_at, w.updated_at,
			(SELECT COUNT(*) FROM workspace_members WHERE workspace_id = w.id) as member_count,
			(SELECT COUNT(*) FROM floor_plans WHERE workspace_id = w.id) as project_count
		FROM workspaces w
		INNER JOIN workspace_members wm ON wm.workspace_id = w.id
		WHERE wm.user_id = $1
	`
	countQuery := `
		SELECT COUNT(DISTINCT w.id)
		FROM workspaces w
		INNER JOIN workspace_members wm ON wm.workspace_id = w.id
		WHERE wm.user_id = $1
	`
	args := []interface{}{opts.UserID}
	argIndex := 2

	// Add name filter if provided
	if opts.NameFilter != "" {
		query += fmt.Sprintf(" AND w.name ILIKE $%d", argIndex)
		countQuery += fmt.Sprintf(" AND w.name ILIKE $%d", argIndex)
		args = append(args, "%"+opts.NameFilter+"%")
		argIndex++
	}

	// Add ordering and pagination
	query += fmt.Sprintf(" ORDER BY w.updated_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, opts.Limit, opts.Offset)

	// Execute count query
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args[:argIndex-1]...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count workspaces: %w", err)
	}

	// Execute list query
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list workspaces: %w", err)
	}
	defer rows.Close()

	// Scan results
	workspaces := make([]*entity.Workspace, 0)
	for rows.Next() {
		var ws entity.Workspace
		err := rows.Scan(
			&ws.ID,
			&ws.OwnerID,
			&ws.Name,
			&ws.Description,
			&ws.CreatedAt,
			&ws.UpdatedAt,
			&ws.MemberCount,
			&ws.ProjectCount,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan workspace: %w", err)
		}
		workspaces = append(workspaces, &ws)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate rows: %w", err)
	}

	return workspaces, total, nil
}

// =============================================================================
// Member Management Operations
// =============================================================================

// GetMembers retrieves all members of a workspace.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - workspaceID: UUID of the workspace
//
// Returns:
//   - []entity.Member: List of all members (including owner)
//   - error: Database error
func (r *WorkspaceRepository) GetMembers(ctx context.Context, workspaceID uuid.UUID) ([]entity.Member, error) {
	query := `
		SELECT id, workspace_id, user_id, role, joined_at, invited_by
		FROM workspace_members
		WHERE workspace_id = $1
		ORDER BY joined_at ASC
	`

	rows, err := r.pool.Query(ctx, query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("query members: %w", err)
	}
	defer rows.Close()

	members := make([]entity.Member, 0)
	for rows.Next() {
		var m entity.Member
		err := rows.Scan(
			&m.ID,
			&m.WorkspaceID,
			&m.UserID,
			&m.Role,
			&m.JoinedAt,
			&m.InvitedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		members = append(members, m)
	}

	return members, rows.Err()
}

// AddMember adds a new member to a workspace.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - member: Member entity to add
//
// Returns:
//   - error: ErrDuplicateMember if already a member, or database error
func (r *WorkspaceRepository) AddMember(ctx context.Context, member *entity.Member) error {
	query := `
		INSERT INTO workspace_members (id, workspace_id, user_id, role, joined_at, invited_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (workspace_id, user_id) DO NOTHING
	`

	result, err := r.pool.Exec(ctx, query,
		member.ID,
		member.WorkspaceID,
		member.UserID,
		member.Role,
		member.JoinedAt,
		member.InvitedBy,
	)
	if err != nil {
		return fmt.Errorf("insert member: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrDuplicateMember
	}

	return nil
}

// RemoveMember removes a member from a workspace.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - workspaceID: UUID of the workspace
//   - userID: UUID of the user to remove
//
// Returns:
//   - error: entity.ErrMemberNotFound if not a member
func (r *WorkspaceRepository) RemoveMember(ctx context.Context, workspaceID, userID uuid.UUID) error {
	query := `DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`

	result, err := r.pool.Exec(ctx, query, workspaceID, userID)
	if err != nil {
		return fmt.Errorf("delete member: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrMemberNotFound
	}

	return nil
}

// UpdateMemberRole changes a member's role in a workspace.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - workspaceID: UUID of the workspace
//   - userID: UUID of the member
//   - role: New role to assign
//
// Returns:
//   - error: entity.ErrMemberNotFound if not a member
func (r *WorkspaceRepository) UpdateMemberRole(ctx context.Context, workspaceID, userID uuid.UUID, role entity.MemberRole) error {
	query := `UPDATE workspace_members SET role = $3 WHERE workspace_id = $1 AND user_id = $2`

	result, err := r.pool.Exec(ctx, query, workspaceID, userID, role)
	if err != nil {
		return fmt.Errorf("update member role: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrMemberNotFound
	}

	return nil
}

// IsMember checks if a user is a member of a workspace.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - workspaceID: UUID of the workspace
//   - userID: UUID of the user
//
// Returns:
//   - bool: true if user is a member
//   - error: Database error
func (r *WorkspaceRepository) IsMember(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM workspace_members WHERE workspace_id = $1 AND user_id = $2)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, workspaceID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check membership: %w", err)
	}

	return exists, nil
}

// GetMember retrieves a single member from a workspace.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - workspaceID: UUID of the workspace
//   - userID: UUID of the user
//
// Returns:
//   - *entity.Member: Member entity if found
//   - error: entity.ErrMemberNotFound if not a member
func (r *WorkspaceRepository) GetMember(ctx context.Context, workspaceID, userID uuid.UUID) (*entity.Member, error) {
	query := `
		SELECT id, workspace_id, user_id, role, joined_at, invited_by
		FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2
	`

	var m entity.Member
	err := r.pool.QueryRow(ctx, query, workspaceID, userID).Scan(
		&m.ID,
		&m.WorkspaceID,
		&m.UserID,
		&m.Role,
		&m.JoinedAt,
		&m.InvitedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrMemberNotFound
		}
		return nil, fmt.Errorf("get member: %w", err)
	}

	return &m, nil
}

// GetMemberRole retrieves the role of a member in a workspace.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - workspaceID: UUID of the workspace
//   - userID: UUID of the user
//
// Returns:
//   - entity.MemberRole: The member's role
//   - error: entity.ErrMemberNotFound if not a member
func (r *WorkspaceRepository) GetMemberRole(ctx context.Context, workspaceID, userID uuid.UUID) (entity.MemberRole, error) {
	query := `SELECT role FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`

	var role entity.MemberRole
	err := r.pool.QueryRow(ctx, query, workspaceID, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", entity.ErrMemberNotFound
		}
		return "", fmt.Errorf("get member role: %w", err)
	}

	return role, nil
}

// =============================================================================
// Invite Management Operations
// =============================================================================

// CreateInvite creates a new workspace invitation.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - invite: Invite entity to create
//
// Returns:
//   - error: Database error
func (r *WorkspaceRepository) CreateInvite(ctx context.Context, invite *entity.WorkspaceInvite) error {
	query := `
		INSERT INTO workspace_invites 
			(id, workspace_id, invited_user_id, invited_by_user_id, role, status, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		invite.ID,
		invite.WorkspaceID,
		invite.InvitedUserID,
		invite.InvitedByUserID,
		invite.Role,
		invite.Status,
		invite.ExpiresAt,
		invite.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create invite: %w", err)
	}

	return nil
}

// GetPendingInvite retrieves a pending invite for a user to a workspace.
func (r *WorkspaceRepository) GetPendingInvite(ctx context.Context, workspaceID, userID uuid.UUID) (*entity.WorkspaceInvite, error) {
	query := `
		SELECT id, workspace_id, invited_user_id, invited_by_user_id, role, status, expires_at, created_at, responded_at
		FROM workspace_invites
		WHERE workspace_id = $1 AND invited_user_id = $2 AND status = 'pending' AND expires_at > NOW()
	`

	var inv entity.WorkspaceInvite
	err := r.pool.QueryRow(ctx, query, workspaceID, userID).Scan(
		&inv.ID,
		&inv.WorkspaceID,
		&inv.InvitedUserID,
		&inv.InvitedByUserID,
		&inv.Role,
		&inv.Status,
		&inv.ExpiresAt,
		&inv.CreatedAt,
		&inv.RespondedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No pending invite
		}
		return nil, fmt.Errorf("get pending invite: %w", err)
	}

	return &inv, nil
}

// UpdateInvite updates an existing invitation (e.g., accept/decline).
func (r *WorkspaceRepository) UpdateInvite(ctx context.Context, invite *entity.WorkspaceInvite) error {
	query := `UPDATE workspace_invites SET status = $2, responded_at = $3 WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, invite.ID, invite.Status, invite.RespondedAt)
	if err != nil {
		return fmt.Errorf("update invite: %w", err)
	}

	return nil
}

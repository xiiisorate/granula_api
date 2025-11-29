// =============================================================================
// Package postgres provides PostgreSQL implementations of repository interfaces.
// =============================================================================
// This package contains the data access layer for request-related entities.
// It uses the pgx driver for high-performance PostgreSQL operations.
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
	"github.com/xiiisorate/granula_api/request-service/internal/domain/entity"
)

// =============================================================================
// RequestRepository
// =============================================================================

// RequestRepository provides data access operations for expert requests.
type RequestRepository struct {
	pool *pgxpool.Pool
}

// NewRequestRepository creates a new RequestRepository.
func NewRequestRepository(pool *pgxpool.Pool) *RequestRepository {
	return &RequestRepository{pool: pool}
}

// =============================================================================
// Request CRUD Operations
// =============================================================================

// Create inserts a new request into the database.
func (r *RequestRepository) Create(ctx context.Context, req *entity.Request) error {
	query := `
		INSERT INTO requests (
			id, workspace_id, user_id, title, description, category, priority, 
			status, estimated_cost, contact_phone, contact_email, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.pool.Exec(ctx, query,
		req.ID,
		req.WorkspaceID,
		req.UserID,
		req.Title,
		req.Description,
		req.Category,
		req.Priority,
		req.Status,
		req.EstimatedCost,
		req.ContactPhone,
		req.ContactEmail,
		req.CreatedAt,
		req.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert request: %w", err)
	}

	// Insert initial status history entry
	if len(req.StatusHistory) > 0 {
		if err := r.insertStatusChange(ctx, &req.StatusHistory[0]); err != nil {
			return fmt.Errorf("insert initial status: %w", err)
		}
	}

	return nil
}

// GetByID retrieves a request by its unique identifier.
func (r *RequestRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Request, error) {
	query := `
		SELECT 
			id, workspace_id, user_id, title, description, category, priority, status,
			expert_id, assigned_at, estimated_cost, final_cost, rejection_reason, notes,
			contact_phone, contact_email, created_at, updated_at, completed_at
		FROM requests
		WHERE id = $1
	`

	var req entity.Request
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&req.ID,
		&req.WorkspaceID,
		&req.UserID,
		&req.Title,
		&req.Description,
		&req.Category,
		&req.Priority,
		&req.Status,
		&req.ExpertID,
		&req.AssignedAt,
		&req.EstimatedCost,
		&req.FinalCost,
		&req.RejectionReason,
		&req.Notes,
		&req.ContactPhone,
		&req.ContactEmail,
		&req.CreatedAt,
		&req.UpdatedAt,
		&req.CompletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrRequestNotFound
		}
		return nil, fmt.Errorf("query request: %w", err)
	}

	return &req, nil
}

// GetByIDWithHistory retrieves a request with its status history.
func (r *RequestRepository) GetByIDWithHistory(ctx context.Context, id uuid.UUID) (*entity.Request, error) {
	req, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Load status history
	history, err := r.GetStatusHistory(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load history: %w", err)
	}
	req.StatusHistory = history

	return req, nil
}

// Update saves changes to an existing request.
func (r *RequestRepository) Update(ctx context.Context, req *entity.Request) error {
	query := `
		UPDATE requests SET
			title = $2, description = $3, priority = $4, status = $5,
			expert_id = $6, assigned_at = $7, estimated_cost = $8, final_cost = $9,
			rejection_reason = $10, notes = $11, contact_phone = $12, contact_email = $13,
			updated_at = $14, completed_at = $15
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		req.ID,
		req.Title,
		req.Description,
		req.Priority,
		req.Status,
		req.ExpertID,
		req.AssignedAt,
		req.EstimatedCost,
		req.FinalCost,
		req.RejectionReason,
		req.Notes,
		req.ContactPhone,
		req.ContactEmail,
		time.Now().UTC(),
		req.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("update request: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrRequestNotFound
	}

	return nil
}

// Delete removes a request (hard delete - use with caution).
func (r *RequestRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete documents
	_, err = tx.Exec(ctx, "DELETE FROM request_documents WHERE request_id = $1", id)
	if err != nil {
		return fmt.Errorf("delete documents: %w", err)
	}

	// Delete status history
	_, err = tx.Exec(ctx, "DELETE FROM request_status_history WHERE request_id = $1", id)
	if err != nil {
		return fmt.Errorf("delete history: %w", err)
	}

	// Delete request
	result, err := tx.Exec(ctx, "DELETE FROM requests WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete request: %w", err)
	}

	if result.RowsAffected() == 0 {
		return entity.ErrRequestNotFound
	}

	return tx.Commit(ctx)
}

// =============================================================================
// List Operations
// =============================================================================

// ListOptions defines parameters for listing requests.
type ListOptions struct {
	// WorkspaceID filters to requests in this workspace.
	WorkspaceID *uuid.UUID

	// UserID filters to requests created by this user.
	UserID *uuid.UUID

	// ExpertID filters to requests assigned to this expert.
	ExpertID *uuid.UUID

	// Status filters to specific status.
	Status *entity.RequestStatus

	// Category filters to specific category.
	Category *entity.RequestCategory

	// Limit and offset for pagination.
	Limit  int
	Offset int
}

// List returns requests matching the given options.
func (r *RequestRepository) List(ctx context.Context, opts ListOptions) ([]*entity.Request, int, error) {
	// Validate and apply defaults
	if opts.Limit <= 0 || opts.Limit > 100 {
		opts.Limit = 20
	}

	// Build dynamic query
	baseQuery := `FROM requests WHERE 1=1`
	countQuery := `SELECT COUNT(*) ` + baseQuery
	listQuery := `
		SELECT id, workspace_id, user_id, title, description, category, priority, status,
		       expert_id, assigned_at, estimated_cost, final_cost, rejection_reason, notes,
		       contact_phone, contact_email, created_at, updated_at, completed_at
	` + baseQuery

	args := make([]interface{}, 0)
	argIndex := 1

	// Add filters
	if opts.WorkspaceID != nil {
		filter := fmt.Sprintf(" AND workspace_id = $%d", argIndex)
		countQuery += filter
		listQuery += filter
		args = append(args, *opts.WorkspaceID)
		argIndex++
	}

	if opts.UserID != nil {
		filter := fmt.Sprintf(" AND user_id = $%d", argIndex)
		countQuery += filter
		listQuery += filter
		args = append(args, *opts.UserID)
		argIndex++
	}

	if opts.ExpertID != nil {
		filter := fmt.Sprintf(" AND expert_id = $%d", argIndex)
		countQuery += filter
		listQuery += filter
		args = append(args, *opts.ExpertID)
		argIndex++
	}

	if opts.Status != nil {
		filter := fmt.Sprintf(" AND status = $%d", argIndex)
		countQuery += filter
		listQuery += filter
		args = append(args, *opts.Status)
		argIndex++
	}

	if opts.Category != nil {
		filter := fmt.Sprintf(" AND category = $%d", argIndex)
		countQuery += filter
		listQuery += filter
		args = append(args, *opts.Category)
		argIndex++
	}

	// Add pagination
	listQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	listArgs := append(args, opts.Limit, opts.Offset)

	// Execute count query
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count requests: %w", err)
	}

	// Execute list query
	rows, err := r.pool.Query(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list requests: %w", err)
	}
	defer rows.Close()

	requests := make([]*entity.Request, 0)
	for rows.Next() {
		var req entity.Request
		err := rows.Scan(
			&req.ID, &req.WorkspaceID, &req.UserID, &req.Title, &req.Description,
			&req.Category, &req.Priority, &req.Status, &req.ExpertID, &req.AssignedAt,
			&req.EstimatedCost, &req.FinalCost, &req.RejectionReason, &req.Notes,
			&req.ContactPhone, &req.ContactEmail, &req.CreatedAt, &req.UpdatedAt, &req.CompletedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan request: %w", err)
		}
		requests = append(requests, &req)
	}

	return requests, total, rows.Err()
}

// =============================================================================
// Status History Operations
// =============================================================================

// insertStatusChange adds a status change record.
func (r *RequestRepository) insertStatusChange(ctx context.Context, change *entity.StatusChange) error {
	query := `
		INSERT INTO request_status_history (id, request_id, from_status, to_status, comment, changed_by, changed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.pool.Exec(ctx, query,
		change.ID,
		change.RequestID,
		change.FromStatus,
		change.ToStatus,
		change.Comment,
		change.ChangedBy,
		change.ChangedAt,
	)
	return err
}

// AddStatusChange records a status change in history.
func (r *RequestRepository) AddStatusChange(ctx context.Context, change *entity.StatusChange) error {
	return r.insertStatusChange(ctx, change)
}

// GetStatusHistory retrieves all status changes for a request.
func (r *RequestRepository) GetStatusHistory(ctx context.Context, requestID uuid.UUID) ([]entity.StatusChange, error) {
	query := `
		SELECT id, request_id, from_status, to_status, comment, changed_by, changed_at
		FROM request_status_history
		WHERE request_id = $1
		ORDER BY changed_at ASC
	`

	rows, err := r.pool.Query(ctx, query, requestID)
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}
	defer rows.Close()

	history := make([]entity.StatusChange, 0)
	for rows.Next() {
		var change entity.StatusChange
		err := rows.Scan(
			&change.ID, &change.RequestID, &change.FromStatus, &change.ToStatus,
			&change.Comment, &change.ChangedBy, &change.ChangedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan history: %w", err)
		}
		history = append(history, change)
	}

	return history, rows.Err()
}

// =============================================================================
// Document Operations
// =============================================================================

// AddDocument attaches a document to a request.
func (r *RequestRepository) AddDocument(ctx context.Context, doc *entity.Document) error {
	query := `
		INSERT INTO request_documents (id, request_id, type, name, storage_path, mime_type, size, uploaded_by, uploaded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pool.Exec(ctx, query,
		doc.ID, doc.RequestID, doc.Type, doc.Name, doc.StoragePath,
		doc.MimeType, doc.Size, doc.UploadedBy, doc.UploadedAt,
	)
	return err
}

// GetDocuments retrieves all documents for a request.
func (r *RequestRepository) GetDocuments(ctx context.Context, requestID uuid.UUID) ([]entity.Document, error) {
	query := `
		SELECT id, request_id, type, name, storage_path, mime_type, size, uploaded_by, uploaded_at
		FROM request_documents
		WHERE request_id = $1
		ORDER BY uploaded_at ASC
	`

	rows, err := r.pool.Query(ctx, query, requestID)
	if err != nil {
		return nil, fmt.Errorf("query documents: %w", err)
	}
	defer rows.Close()

	docs := make([]entity.Document, 0)
	for rows.Next() {
		var doc entity.Document
		err := rows.Scan(
			&doc.ID, &doc.RequestID, &doc.Type, &doc.Name, &doc.StoragePath,
			&doc.MimeType, &doc.Size, &doc.UploadedBy, &doc.UploadedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}
		docs = append(docs, doc)
	}

	return docs, rows.Err()
}

// DeleteDocument removes a document.
func (r *RequestRepository) DeleteDocument(ctx context.Context, documentID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM request_documents WHERE id = $1", documentID)
	return err
}


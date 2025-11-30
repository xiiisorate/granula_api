// =============================================================================
// Package service provides business logic for Request Service.
// =============================================================================
// This package implements the core business operations for expert request
// management, including request creation, status transitions, expert
// assignment, and document handling.
//
// =============================================================================
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/request-service/internal/domain/entity"
	"github.com/xiiisorate/granula_api/request-service/internal/repository/postgres"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
)

// =============================================================================
// Service Errors
// =============================================================================

var (
	// ErrForbidden is returned when user lacks permission for the operation.
	ErrForbidden = errors.New("forbidden: insufficient permissions")

	// ErrUnauthorized is returned when user is not authenticated.
	ErrUnauthorized = errors.New("unauthorized: authentication required")
)

// =============================================================================
// RequestService
// =============================================================================

// NotificationSender is an interface for sending notifications.
// This allows for easier testing and decoupling from the notification client implementation.
type NotificationSender interface {
	SendRequestSubmitted(ctx context.Context, userID, requestID, serviceType string) error
	SendRequestStatusChanged(ctx context.Context, userID, requestID, oldStatus, newStatus, comment string) error
	SendRequestAssigned(ctx context.Context, userID, requestID, expertID, expertName string) error
	SendRequestRejected(ctx context.Context, userID, requestID, reason string) error
	SendRequestCompleted(ctx context.Context, userID, requestID string, finalCost int) error
	NotifyStaff(ctx context.Context, requestID, serviceType, workspaceID string) error
}

// RequestService provides business operations for expert request management.
// It integrates with Notification Service to send status updates to users.
type RequestService struct {
	repo         *postgres.RequestRepository
	notification NotificationSender // Optional: nil if notifications are disabled
	log          *logger.Logger
}

// NewRequestService creates a new RequestService instance.
//
// Parameters:
//   - repo: PostgreSQL repository for data persistence
//   - notification: Notification client for sending alerts (can be nil)
//   - log: Logger instance for operational logging
//
// Returns:
//   - *RequestService: Initialized service ready for use
//
// Note: If notification is nil, notifications will be silently skipped.
func NewRequestService(repo *postgres.RequestRepository, notification NotificationSender, log *logger.Logger) *RequestService {
	return &RequestService{
		repo:         repo,
		notification: notification,
		log:          log,
	}
}

// =============================================================================
// Request CRUD Operations
// =============================================================================

// CreateRequest creates a new expert request.
//
// Parameters:
//   - ctx: Context for cancellation
//   - workspaceID: UUID of the associated workspace
//   - userID: UUID of the user creating the request
//   - title: Brief summary (5-200 characters)
//   - description: Detailed description
//   - category: Type of service requested
//
// Returns:
//   - *entity.Request: Created request in draft status
//   - error: Validation or database error
func (s *RequestService) CreateRequest(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	title, description string,
	category entity.RequestCategory,
) (*entity.Request, error) {
	s.log.Info("creating request",
		logger.String("workspace_id", workspaceID.String()),
		logger.String("user_id", userID.String()),
		logger.String("category", string(category)),
	)

	// Create request entity
	req, err := entity.NewRequest(workspaceID, userID, title, category)
	if err != nil {
		return nil, err
	}

	// Set description
	if err := req.UpdateDescription(description); err != nil {
		return nil, err
	}

	// Persist to database
	if err := s.repo.Create(ctx, req); err != nil {
		s.log.Error("failed to create request", logger.Err(err))
		return nil, fmt.Errorf("create request: %w", err)
	}

	s.log.Info("request created",
		logger.String("request_id", req.ID.String()),
	)

	return req, nil
}

// GetRequest retrieves a request by ID.
// Access control: User must be request owner, expert, or workspace admin.
func (s *RequestService) GetRequest(ctx context.Context, requestID, userID uuid.UUID) (*entity.Request, error) {
	req, err := s.repo.GetByIDWithHistory(ctx, requestID)
	if err != nil {
		return nil, err
	}

	// Access check - for now, allow owner or assigned expert
	if req.UserID != userID && (req.ExpertID == nil || *req.ExpertID != userID) {
		// TODO: Add workspace admin check
		// For now, return the request (add proper RBAC later)
	}

	return req, nil
}

// ListRequests returns requests with filtering.
func (s *RequestService) ListRequests(
	ctx context.Context,
	workspaceID *uuid.UUID,
	userID *uuid.UUID,
	status *entity.RequestStatus,
	limit, offset int,
) ([]*entity.Request, int, error) {
	opts := postgres.ListOptions{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Status:      status,
		Limit:       limit,
		Offset:      offset,
	}

	return s.repo.List(ctx, opts)
}

// UpdateRequest updates request title and description.
// Only allowed for draft requests by the owner.
func (s *RequestService) UpdateRequest(
	ctx context.Context,
	requestID, userID uuid.UUID,
	title, description string,
) (*entity.Request, error) {
	s.log.Info("updating request",
		logger.String("request_id", requestID.String()),
	)

	req, err := s.repo.GetByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	// Check authorization
	if !req.CanUserModify(userID) {
		return nil, entity.ErrCannotModifyRequest
	}

	// Update fields
	if title != "" {
		req.Title = title
	}
	if description != "" {
		req.Description = description
	}
	req.UpdatedAt = time.Now().UTC()

	// Save
	if err := s.repo.Update(ctx, req); err != nil {
		return nil, fmt.Errorf("update request: %w", err)
	}

	return req, nil
}

// SubmitRequest submits a draft request for review.
func (s *RequestService) SubmitRequest(ctx context.Context, requestID, userID uuid.UUID) (*entity.Request, error) {
	s.log.Info("submitting request",
		logger.String("request_id", requestID.String()),
	)

	req, err := s.repo.GetByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	// Check authorization
	if req.UserID != userID {
		return nil, entity.ErrCannotModifyRequest
	}

	// Submit
	if err := req.Submit(); err != nil {
		return nil, err
	}

	// Save status change
	if err := s.repo.Update(ctx, req); err != nil {
		return nil, fmt.Errorf("update request: %w", err)
	}

	// Record status change
	if len(req.StatusHistory) > 0 {
		lastChange := req.StatusHistory[len(req.StatusHistory)-1]
		if err := s.repo.AddStatusChange(ctx, &lastChange); err != nil {
			s.log.Warn("failed to record status change", logger.Err(err))
		}
	}

	s.log.Info("request submitted",
		logger.String("request_id", requestID.String()),
	)

	// Send notification to staff about new request
	if s.notification != nil {
		go func() {
			// Use background context as the request context may be cancelled
			ctx := context.Background()
			if err := s.notification.NotifyStaff(ctx, req.ID.String(), string(req.Category), req.WorkspaceID.String()); err != nil {
				s.log.Warn("failed to notify staff about new request", logger.Err(err))
			}
			// Also notify the user that their request was submitted
			if err := s.notification.SendRequestSubmitted(ctx, req.UserID.String(), req.ID.String(), string(req.Category)); err != nil {
				s.log.Warn("failed to notify user about request submission", logger.Err(err))
			}
		}()
	}

	return req, nil
}

// CancelRequest cancels a request.
func (s *RequestService) CancelRequest(ctx context.Context, requestID, userID uuid.UUID) error {
	s.log.Info("cancelling request",
		logger.String("request_id", requestID.String()),
	)

	req, err := s.repo.GetByID(ctx, requestID)
	if err != nil {
		return err
	}

	// Check authorization
	if !req.CanUserCancel(userID) {
		return entity.ErrCannotModifyRequest
	}

	// Cancel
	if err := req.Cancel(userID); err != nil {
		return err
	}

	// Save
	if err := s.repo.Update(ctx, req); err != nil {
		return fmt.Errorf("update request: %w", err)
	}

	// Record status change
	if len(req.StatusHistory) > 0 {
		lastChange := req.StatusHistory[len(req.StatusHistory)-1]
		if err := s.repo.AddStatusChange(ctx, &lastChange); err != nil {
			s.log.Warn("failed to record status change", logger.Err(err))
		}
	}

	s.log.Info("request cancelled",
		logger.String("request_id", requestID.String()),
	)

	return nil
}

// =============================================================================
// Staff/Admin Operations
// =============================================================================

// UpdateStatus changes the request status (staff operation).
func (s *RequestService) UpdateStatus(
	ctx context.Context,
	requestID uuid.UUID,
	newStatus entity.RequestStatus,
	comment string,
	changedBy uuid.UUID,
) (*entity.Request, error) {
	s.log.Info("updating request status",
		logger.String("request_id", requestID.String()),
		logger.String("new_status", string(newStatus)),
	)

	req, err := s.repo.GetByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	// Validate and change status
	if err := req.ChangeStatus(newStatus, comment, &changedBy); err != nil {
		return nil, err
	}

	// Save
	if err := s.repo.Update(ctx, req); err != nil {
		return nil, fmt.Errorf("update request: %w", err)
	}

	// Record status change
	if len(req.StatusHistory) > 0 {
		lastChange := req.StatusHistory[len(req.StatusHistory)-1]
		if err := s.repo.AddStatusChange(ctx, &lastChange); err != nil {
			s.log.Warn("failed to record status change", logger.Err(err))
		}
	}

	s.log.Info("request status updated",
		logger.String("request_id", requestID.String()),
		logger.String("status", string(newStatus)),
	)

	// Send notification to user about status change
	if s.notification != nil {
		go func() {
			ctx := context.Background()
			oldStatus := "" // Previous status before the change
			// Get the previous status from history if available
			if len(req.StatusHistory) > 1 {
				oldStatus = string(req.StatusHistory[len(req.StatusHistory)-2].FromStatus)
			} else if len(req.StatusHistory) == 1 {
				oldStatus = string(req.StatusHistory[0].FromStatus)
			}
			if err := s.notification.SendRequestStatusChanged(ctx, req.UserID.String(), req.ID.String(), oldStatus, string(newStatus), comment); err != nil {
				s.log.Warn("failed to notify user about status change", logger.Err(err))
			}
		}()
	}

	return req, nil
}

// AssignExpert assigns an expert to a request.
func (s *RequestService) AssignExpert(
	ctx context.Context,
	requestID, expertID uuid.UUID,
	assignedBy uuid.UUID,
) (*entity.Request, error) {
	s.log.Info("assigning expert",
		logger.String("request_id", requestID.String()),
		logger.String("expert_id", expertID.String()),
	)

	req, err := s.repo.GetByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	// Assign expert
	if err := req.AssignExpert(expertID); err != nil {
		return nil, err
	}

	// Save
	if err := s.repo.Update(ctx, req); err != nil {
		return nil, fmt.Errorf("update request: %w", err)
	}

	// Record status change
	if len(req.StatusHistory) > 0 {
		lastChange := req.StatusHistory[len(req.StatusHistory)-1]
		if err := s.repo.AddStatusChange(ctx, &lastChange); err != nil {
			s.log.Warn("failed to record status change", logger.Err(err))
		}
	}

	s.log.Info("expert assigned",
		logger.String("request_id", requestID.String()),
		logger.String("expert_id", expertID.String()),
	)

	// Send notifications to user and expert about assignment
	if s.notification != nil {
		go func() {
			ctx := context.Background()
			// Note: In production, fetch expert name from User Service
			// For now, use expert ID as placeholder
			expertName := expertID.String()
			if err := s.notification.SendRequestAssigned(ctx, req.UserID.String(), req.ID.String(), expertID.String(), expertName); err != nil {
				s.log.Warn("failed to notify about expert assignment", logger.Err(err))
			}
		}()
	}

	return req, nil
}

// RejectRequest rejects a request with a reason.
func (s *RequestService) RejectRequest(
	ctx context.Context,
	requestID uuid.UUID,
	reason string,
	rejectedBy uuid.UUID,
) (*entity.Request, error) {
	s.log.Info("rejecting request",
		logger.String("request_id", requestID.String()),
	)

	req, err := s.repo.GetByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	// Reject
	if err := req.Reject(reason, rejectedBy); err != nil {
		return nil, err
	}

	// Save
	if err := s.repo.Update(ctx, req); err != nil {
		return nil, fmt.Errorf("update request: %w", err)
	}

	// Record status change
	if len(req.StatusHistory) > 0 {
		lastChange := req.StatusHistory[len(req.StatusHistory)-1]
		if err := s.repo.AddStatusChange(ctx, &lastChange); err != nil {
			s.log.Warn("failed to record status change", logger.Err(err))
		}
	}

	s.log.Info("request rejected",
		logger.String("request_id", requestID.String()),
	)

	// Send notification to user about rejection
	if s.notification != nil {
		go func() {
			ctx := context.Background()
			if err := s.notification.SendRequestRejected(ctx, req.UserID.String(), req.ID.String(), reason); err != nil {
				s.log.Warn("failed to notify user about request rejection", logger.Err(err))
			}
		}()
	}

	return req, nil
}

// CompleteRequest marks a request as completed.
func (s *RequestService) CompleteRequest(
	ctx context.Context,
	requestID uuid.UUID,
	finalCost int,
	completedBy uuid.UUID,
) (*entity.Request, error) {
	s.log.Info("completing request",
		logger.String("request_id", requestID.String()),
	)

	req, err := s.repo.GetByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	// Set final cost
	req.FinalCost = &finalCost

	// Complete
	if err := req.ChangeStatus(entity.StatusCompleted, "Request completed", &completedBy); err != nil {
		return nil, err
	}

	// Save
	if err := s.repo.Update(ctx, req); err != nil {
		return nil, fmt.Errorf("update request: %w", err)
	}

	// Record status change
	if len(req.StatusHistory) > 0 {
		lastChange := req.StatusHistory[len(req.StatusHistory)-1]
		if err := s.repo.AddStatusChange(ctx, &lastChange); err != nil {
			s.log.Warn("failed to record status change", logger.Err(err))
		}
	}

	s.log.Info("request completed",
		logger.String("request_id", requestID.String()),
	)

	// Send notification to user about completion
	if s.notification != nil {
		go func() {
			ctx := context.Background()
			cost := 0
			if req.FinalCost != nil {
				cost = *req.FinalCost
			}
			if err := s.notification.SendRequestCompleted(ctx, req.UserID.String(), req.ID.String(), cost); err != nil {
				s.log.Warn("failed to notify user about request completion", logger.Err(err))
			}
		}()
	}

	return req, nil
}

// =============================================================================
// Document Operations
// =============================================================================

// AddDocument attaches a document to a request.
func (s *RequestService) AddDocument(
	ctx context.Context,
	requestID uuid.UUID,
	docType entity.DocumentType,
	name, storagePath, mimeType string,
	size int64,
	uploadedBy uuid.UUID,
) (*entity.Document, error) {
	// Verify request exists
	_, err := s.repo.GetByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	doc := entity.NewDocument(requestID, docType, name, storagePath, mimeType, size, uploadedBy)

	if err := s.repo.AddDocument(ctx, doc); err != nil {
		return nil, fmt.Errorf("add document: %w", err)
	}

	s.log.Info("document added",
		logger.String("request_id", requestID.String()),
		logger.String("document_id", doc.ID.String()),
	)

	return doc, nil
}

// GetDocuments retrieves all documents for a request.
func (s *RequestService) GetDocuments(ctx context.Context, requestID uuid.UUID) ([]entity.Document, error) {
	return s.repo.GetDocuments(ctx, requestID)
}

// GetStatusHistory retrieves the status change history.
func (s *RequestService) GetStatusHistory(ctx context.Context, requestID uuid.UUID) ([]entity.StatusChange, error) {
	return s.repo.GetStatusHistory(ctx, requestID)
}

// Package service handles business logic for Request Service.
package service

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/request-service/internal/repository"
	"github.com/xiiisorate/granula_api/shared/pkg/errors"
)

// RequestService handles request business logic.
type RequestService struct {
	requestRepo *repository.RequestRepository
	historyRepo *repository.StatusHistoryRepository
}

// NewRequestService creates a new RequestService.
func NewRequestService(
	requestRepo *repository.RequestRepository,
	historyRepo *repository.StatusHistoryRepository,
) *RequestService {
	return &RequestService{
		requestRepo: requestRepo,
		historyRepo: historyRepo,
	}
}

// CreateRequestInput contains request creation data.
type CreateRequestInput struct {
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	Title       string
	Description string
	Category    string
	Priority    string
}

// CreateRequest creates a new request.
func (s *RequestService) CreateRequest(input *CreateRequestInput) (*repository.Request, error) {
	// Validate title
	title := strings.TrimSpace(input.Title)
	if len(title) < 3 || len(title) > 255 {
		return nil, errors.InvalidArgument("title", "must be between 3 and 255 characters")
	}

	// Validate category
	if !repository.IsValidCategory(input.Category) {
		return nil, errors.InvalidArgument("category", "invalid category")
	}

	// Validate priority (default to normal)
	priority := input.Priority
	if priority == "" {
		priority = repository.PriorityNormal
	}
	if !repository.IsValidPriority(priority) {
		return nil, errors.InvalidArgument("priority", "invalid priority")
	}

	// Create request
	request := &repository.Request{
		ID:          uuid.New(),
		WorkspaceID: input.WorkspaceID,
		UserID:      input.UserID,
		Title:       title,
		Description: strings.TrimSpace(input.Description),
		Category:    input.Category,
		Status:      repository.StatusPending,
		Priority:    priority,
	}

	if err := s.requestRepo.Create(request); err != nil {
		return nil, errors.Internal("failed to create request").WithCause(err)
	}

	// Create initial status history
	history := &repository.StatusHistory{
		ID:        uuid.New(),
		RequestID: request.ID,
		Status:    repository.StatusPending,
		Comment:   "Request created",
		ChangedBy: input.UserID,
	}

	if err := s.historyRepo.Create(history); err != nil {
		// Log but don't fail
	}

	return s.requestRepo.FindByID(request.ID)
}

// GetRequest returns a request by ID.
func (s *RequestService) GetRequest(requestID uuid.UUID) (*repository.Request, error) {
	request, err := s.requestRepo.FindByID(requestID)
	if err != nil {
		return nil, errors.NotFound("request", requestID.String())
	}
	return request, nil
}

// ListRequestsInput contains list requests parameters.
type ListRequestsInput struct {
	WorkspaceID uuid.UUID
	Status      string
	Page        int
	PageSize    int
}

// ListRequestsOutput contains list requests result.
type ListRequestsOutput struct {
	Requests []repository.Request
	Total    int64
	Page     int
	PageSize int
}

// ListRequests returns requests for a workspace.
func (s *RequestService) ListRequests(input *ListRequestsInput) (*ListRequestsOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 100 {
		input.PageSize = 20
	}

	// Validate status if provided
	if input.Status != "" && !repository.IsValidStatus(input.Status) {
		return nil, errors.InvalidArgument("status", "invalid status")
	}

	requests, total, err := s.requestRepo.FindByWorkspaceID(input.WorkspaceID, input.Status, input.Page, input.PageSize)
	if err != nil {
		return nil, errors.Internal("failed to list requests").WithCause(err)
	}

	return &ListRequestsOutput{
		Requests: requests,
		Total:    total,
		Page:     input.Page,
		PageSize: input.PageSize,
	}, nil
}

// UpdateRequestInput contains request update data.
type UpdateRequestInput struct {
	RequestID   uuid.UUID
	UserID      uuid.UUID
	Title       *string
	Description *string
}

// UpdateRequest updates a request.
func (s *RequestService) UpdateRequest(input *UpdateRequestInput) (*repository.Request, error) {
	request, err := s.requestRepo.FindByID(input.RequestID)
	if err != nil {
		return nil, errors.NotFound("request", input.RequestID.String())
	}

	// Only creator can update (and only in pending/rejected status)
	if request.UserID != input.UserID {
		return nil, errors.PermissionDenied("only the creator can update this request")
	}

	if request.Status != repository.StatusPending && request.Status != repository.StatusRejected {
		return nil, errors.PreconditionFailed("request can only be updated in pending or rejected status")
	}

	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if len(title) < 3 || len(title) > 255 {
			return nil, errors.InvalidArgument("title", "must be between 3 and 255 characters")
		}
		request.Title = title
	}

	if input.Description != nil {
		request.Description = strings.TrimSpace(*input.Description)
	}

	request.UpdatedAt = time.Now()

	if err := s.requestRepo.Update(request); err != nil {
		return nil, errors.Internal("failed to update request").WithCause(err)
	}

	return request, nil
}

// CancelRequest cancels a request.
func (s *RequestService) CancelRequest(requestID, userID uuid.UUID) error {
	request, err := s.requestRepo.FindByID(requestID)
	if err != nil {
		return errors.NotFound("request", requestID.String())
	}

	// Only creator can cancel
	if request.UserID != userID {
		return errors.PermissionDenied("only the creator can cancel this request")
	}

	// Check if cancellation is allowed
	if !repository.CanTransitionTo(request.Status, repository.StatusCancelled) {
		return errors.PreconditionFailed("request cannot be cancelled in current status")
	}

	if err := s.requestRepo.UpdateStatus(requestID, repository.StatusCancelled); err != nil {
		return errors.Internal("failed to cancel request").WithCause(err)
	}

	// Add to history
	s.addStatusHistory(requestID, repository.StatusCancelled, "Request cancelled by user", userID)

	return nil
}

// AssignExpertInput contains expert assignment data.
type AssignExpertInput struct {
	RequestID uuid.UUID
	ExpertID  uuid.UUID
	AssignedBy uuid.UUID
}

// AssignExpert assigns an expert to a request.
func (s *RequestService) AssignExpert(input *AssignExpertInput) (*repository.Request, error) {
	request, err := s.requestRepo.FindByID(input.RequestID)
	if err != nil {
		return nil, errors.NotFound("request", input.RequestID.String())
	}

	// Can only assign in approved status
	if request.Status != repository.StatusApproved {
		return nil, errors.PreconditionFailed("expert can only be assigned to approved requests")
	}

	if err := s.requestRepo.UpdateExpert(input.RequestID, &input.ExpertID); err != nil {
		return nil, errors.Internal("failed to assign expert").WithCause(err)
	}

	// Update status to in_progress
	if err := s.requestRepo.UpdateStatus(input.RequestID, repository.StatusInProgress); err != nil {
		return nil, errors.Internal("failed to update status").WithCause(err)
	}

	// Add to history
	s.addStatusHistory(input.RequestID, repository.StatusInProgress, "Expert assigned", input.AssignedBy)

	return s.requestRepo.FindByID(input.RequestID)
}

// UpdateStatusInput contains status update data.
type UpdateStatusInput struct {
	RequestID uuid.UUID
	Status    string
	Comment   string
	ChangedBy uuid.UUID
}

// UpdateStatus updates request status.
func (s *RequestService) UpdateStatus(input *UpdateStatusInput) (*repository.Request, error) {
	request, err := s.requestRepo.FindByID(input.RequestID)
	if err != nil {
		return nil, errors.NotFound("request", input.RequestID.String())
	}

	// Validate new status
	if !repository.IsValidStatus(input.Status) {
		return nil, errors.InvalidArgument("status", "invalid status")
	}

	// Check transition is valid
	if !repository.CanTransitionTo(request.Status, input.Status) {
		return nil, errors.PreconditionFailed("invalid status transition from " + request.Status + " to " + input.Status)
	}

	if err := s.requestRepo.UpdateStatus(input.RequestID, input.Status); err != nil {
		return nil, errors.Internal("failed to update status").WithCause(err)
	}

	// Add to history
	s.addStatusHistory(input.RequestID, input.Status, input.Comment, input.ChangedBy)

	return s.requestRepo.FindByID(input.RequestID)
}

// GetStatusHistory returns status history for a request.
func (s *RequestService) GetStatusHistory(requestID uuid.UUID) ([]repository.StatusHistory, error) {
	return s.historyRepo.FindByRequestID(requestID)
}

// addStatusHistory adds a status history entry.
func (s *RequestService) addStatusHistory(requestID uuid.UUID, status, comment string, changedBy uuid.UUID) {
	history := &repository.StatusHistory{
		ID:        uuid.New(),
		RequestID: requestID,
		Status:    status,
		Comment:   comment,
		ChangedBy: changedBy,
	}
	_ = s.historyRepo.Create(history)
}


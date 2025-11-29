// =============================================================================
// Package grpc provides gRPC server implementation for Request Service.
// =============================================================================
// This package implements the RequestService gRPC interface for managing
// expert consultation requests from users for BTI services.
//
// Request Types:
//   - consultation: Online consultation (from 2000 ₽)
//   - documentation: Document preparation (from 15000 ₽)
//   - expert_visit: On-site expert visit (from 5000 ₽)
//   - full_package: Complete package (from 30000 ₽)
//
// Status Flow:
//
//	draft → pending → in_review → approved → assigned → in_progress → completed
//	                           ↓
//	                       rejected
//	(any) → cancelled
//
// =============================================================================
package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/request-service/internal/domain/entity"
	"github.com/xiiisorate/granula_api/request-service/internal/service"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// =============================================================================
// RequestServer Implementation
// =============================================================================

// RequestServer implements the gRPC RequestService interface.
// It delegates all business logic to the service layer.
type RequestServer struct {
	service *service.RequestService
	log     *logger.Logger
}

// NewRequestServer creates a new RequestServer instance.
func NewRequestServer(svc *service.RequestService, log *logger.Logger) *RequestServer {
	return &RequestServer{
		service: svc,
		log:     log,
	}
}

// =============================================================================
// DTOs
// =============================================================================

// CreateRequestDTO represents input for creating a request.
type CreateRequestDTO struct {
	WorkspaceID  string
	UserID       string
	Title        string
	Description  string
	Category     string
	ContactPhone string
	ContactEmail string
}

// RequestResponseDTO represents request data in responses.
type RequestResponseDTO struct {
	ID              string
	WorkspaceID     string
	UserID          string
	Title           string
	Description     string
	Category        string
	Priority        string
	Status          string
	ExpertID        string
	AssignedAt      int64
	EstimatedCost   int32
	FinalCost       int32
	RejectionReason string
	ContactPhone    string
	ContactEmail    string
	CreatedAt       int64
	UpdatedAt       int64
	CompletedAt     int64
}

// =============================================================================
// CRUD Methods
// =============================================================================

// CreateRequest creates a new expert request.
func (s *RequestServer) CreateRequest(ctx context.Context, req *CreateRequestDTO) (*RequestResponseDTO, error) {
	s.log.Info("CreateRequest called",
		logger.String("workspace_id", req.WorkspaceID),
		logger.String("user_id", req.UserID),
	)

	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace_id")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	category := entity.RequestCategory(req.Category)
	if !category.IsValid() {
		return nil, status.Error(codes.InvalidArgument, "invalid category")
	}

	request, err := s.service.CreateRequest(ctx, workspaceID, userID, req.Title, req.Description, category)
	if err != nil {
		return nil, s.mapError(err)
	}

	// Set contact info
	request.ContactPhone = req.ContactPhone
	request.ContactEmail = req.ContactEmail

	return requestToDTO(request), nil
}

// GetRequest retrieves a request by ID.
func (s *RequestServer) GetRequest(ctx context.Context, requestID, userID string) (*RequestResponseDTO, error) {
	id, err := uuid.Parse(requestID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request_id")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	request, err := s.service.GetRequest(ctx, id, uid)
	if err != nil {
		return nil, s.mapError(err)
	}

	return requestToDTO(request), nil
}

// ListRequestsDTO represents list request parameters.
type ListRequestsDTO struct {
	UserID      string
	WorkspaceID string
	Status      string
	Limit       int32
	Offset      int32
}

// ListRequestsResponseDTO represents list response.
type ListRequestsResponseDTO struct {
	Requests []*RequestResponseDTO
	Total    int32
}

// ListRequests returns requests based on filters.
func (s *RequestServer) ListRequests(ctx context.Context, req *ListRequestsDTO) (*ListRequestsResponseDTO, error) {
	var userID *uuid.UUID
	if req.UserID != "" {
		id, err := uuid.Parse(req.UserID)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid user_id")
		}
		userID = &id
	}

	var workspaceID *uuid.UUID
	if req.WorkspaceID != "" {
		id, err := uuid.Parse(req.WorkspaceID)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid workspace_id")
		}
		workspaceID = &id
	}

	var statusFilter *entity.RequestStatus
	if req.Status != "" {
		st := entity.RequestStatus(req.Status)
		if !st.IsValid() {
			return nil, status.Error(codes.InvalidArgument, "invalid status")
		}
		statusFilter = &st
	}

	requests, total, err := s.service.ListRequests(ctx, workspaceID, userID, statusFilter, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, s.mapError(err)
	}

	dtos := make([]*RequestResponseDTO, 0, len(requests))
	for _, r := range requests {
		dtos = append(dtos, requestToDTO(r))
	}

	return &ListRequestsResponseDTO{
		Requests: dtos,
		Total:    int32(total),
	}, nil
}

// =============================================================================
// Status Transition Methods
// =============================================================================

// SubmitRequest transitions from draft to pending.
func (s *RequestServer) SubmitRequest(ctx context.Context, requestID, userID string) (*RequestResponseDTO, error) {
	id, err := uuid.Parse(requestID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request_id")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	request, err := s.service.SubmitRequest(ctx, id, uid)
	if err != nil {
		return nil, s.mapError(err)
	}

	return requestToDTO(request), nil
}

// CancelRequest cancels the request.
func (s *RequestServer) CancelRequest(ctx context.Context, requestID, userID string) error {
	id, err := uuid.Parse(requestID)
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid request_id")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid user_id")
	}

	if err := s.service.CancelRequest(ctx, id, uid); err != nil {
		return s.mapError(err)
	}

	return nil
}

// =============================================================================
// Expert Methods
// =============================================================================

// AssignExpert assigns an expert to the request.
func (s *RequestServer) AssignExpert(ctx context.Context, requestID, expertID, assignedBy string) (*RequestResponseDTO, error) {
	id, err := uuid.Parse(requestID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request_id")
	}

	eid, err := uuid.Parse(expertID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid expert_id")
	}

	assignedByID, err := uuid.Parse(assignedBy)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid assigned_by")
	}

	request, err := s.service.AssignExpert(ctx, id, eid, assignedByID)
	if err != nil {
		return nil, s.mapError(err)
	}

	return requestToDTO(request), nil
}

// RejectRequest rejects the request with a reason.
func (s *RequestServer) RejectRequest(ctx context.Context, requestID, reason, rejectedBy string) error {
	id, err := uuid.Parse(requestID)
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid request_id")
	}

	if reason == "" {
		return status.Error(codes.InvalidArgument, "rejection reason is required")
	}

	rejectedByID, err := uuid.Parse(rejectedBy)
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid rejected_by")
	}

	_, err = s.service.RejectRequest(ctx, id, reason, rejectedByID)
	if err != nil {
		return s.mapError(err)
	}

	return nil
}

// CompleteRequest marks the request as completed.
func (s *RequestServer) CompleteRequest(ctx context.Context, requestID string, finalCost int, completedBy string) (*RequestResponseDTO, error) {
	id, err := uuid.Parse(requestID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request_id")
	}

	completedByID, err := uuid.Parse(completedBy)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid completed_by")
	}

	request, err := s.service.CompleteRequest(ctx, id, finalCost, completedByID)
	if err != nil {
		return nil, s.mapError(err)
	}

	return requestToDTO(request), nil
}

// =============================================================================
// Pricing Info
// =============================================================================

// PricingInfo represents pricing for a service category.
type PricingInfo struct {
	Category      string
	Name          string
	Description   string
	BasePrice     int32
	EstimatedDays string
	Includes      []string
}

// GetPricing returns pricing information.
func (s *RequestServer) GetPricing(_ context.Context) ([]*PricingInfo, error) {
	return []*PricingInfo{
		{
			Category:      "consultation",
			Name:          "Консультация",
			Description:   "Онлайн консультация с экспертом БТИ",
			BasePrice:     2000,
			EstimatedDays: "1-2 дня",
			Includes:      []string{"Анализ планировки", "Рекомендации", "Оценка рисков"},
		},
		{
			Category:      "documentation",
			Name:          "Оформление документации",
			Description:   "Подготовка пакета документов",
			BasePrice:     15000,
			EstimatedDays: "7-14 дней",
			Includes:      []string{"Проект перепланировки", "Техническое заключение"},
		},
		{
			Category:      "expert_visit",
			Name:          "Выезд эксперта",
			Description:   "Выезд эксперта на объект",
			BasePrice:     5000,
			EstimatedDays: "1-3 дня",
			Includes:      []string{"Осмотр объекта", "Замеры", "Заключение"},
		},
		{
			Category:      "full_package",
			Name:          "Полный комплекс",
			Description:   "Полное сопровождение под ключ",
			BasePrice:     30000,
			EstimatedDays: "14-30 дней",
			Includes:      []string{"Все услуги", "Согласование", "Ввод в эксплуатацию"},
		},
	}, nil
}

// =============================================================================
// Error Mapping
// =============================================================================

func (s *RequestServer) mapError(err error) error {
	switch {
	case errors.Is(err, entity.ErrRequestNotFound):
		return status.Error(codes.NotFound, "request not found")
	case errors.Is(err, entity.ErrInvalidRequestStatus):
		return status.Error(codes.FailedPrecondition, "invalid status transition")
	case errors.Is(err, entity.ErrCannotModifyRequest):
		return status.Error(codes.FailedPrecondition, "cannot modify request")
	case errors.Is(err, entity.ErrInvalidRequestTitle):
		return status.Error(codes.InvalidArgument, "invalid title")
	case errors.Is(err, entity.ErrExpertAlreadyAssigned):
		return status.Error(codes.AlreadyExists, "expert already assigned")
	case errors.Is(err, service.ErrForbidden):
		return status.Error(codes.PermissionDenied, "access denied")
	case errors.Is(err, service.ErrUnauthorized):
		return status.Error(codes.Unauthenticated, "unauthorized")
	default:
		s.log.Error("internal error", logger.Err(err))
		return status.Error(codes.Internal, "internal server error")
	}
}

// =============================================================================
// DTO Conversion
// =============================================================================

func requestToDTO(r *entity.Request) *RequestResponseDTO {
	dto := &RequestResponseDTO{
		ID:            r.ID.String(),
		WorkspaceID:   r.WorkspaceID.String(),
		UserID:        r.UserID.String(),
		Title:         r.Title,
		Description:   r.Description,
		Category:      string(r.Category),
		Priority:      string(r.Priority),
		Status:        string(r.Status),
		EstimatedCost: int32(r.EstimatedCost),
		ContactPhone:  r.ContactPhone,
		ContactEmail:  r.ContactEmail,
		CreatedAt:     r.CreatedAt.Unix(),
		UpdatedAt:     r.UpdatedAt.Unix(),
	}

	if r.ExpertID != nil {
		dto.ExpertID = r.ExpertID.String()
	}
	if r.AssignedAt != nil {
		dto.AssignedAt = r.AssignedAt.Unix()
	}
	if r.FinalCost != nil {
		dto.FinalCost = int32(*r.FinalCost)
	}
	if r.RejectionReason != "" {
		dto.RejectionReason = r.RejectionReason
	}
	if r.CompletedAt != nil {
		dto.CompletedAt = r.CompletedAt.Unix()
	}

	return dto
}

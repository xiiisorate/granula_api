// Package server implements gRPC server for Request Service.
package server

import (
	"context"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/request-service/internal/repository"
	"github.com/xiiisorate/granula_api/request-service/internal/service"
	"github.com/xiiisorate/granula_api/shared/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RequestServiceServer is the interface for Request gRPC service.
type RequestServiceServer interface {
	CreateRequest(ctx context.Context, req *CreateRequestRequest) (*CreateRequestResponse, error)
	GetRequest(ctx context.Context, req *GetRequestRequest) (*GetRequestResponse, error)
	ListRequests(ctx context.Context, req *ListRequestsRequest) (*ListRequestsResponse, error)
	UpdateRequest(ctx context.Context, req *UpdateRequestRequest) (*UpdateRequestResponse, error)
	CancelRequest(ctx context.Context, req *CancelRequestRequest) (*CancelRequestResponse, error)
	AssignExpert(ctx context.Context, req *AssignExpertRequest) (*AssignExpertResponse, error)
	UpdateStatus(ctx context.Context, req *UpdateStatusRequest) (*UpdateStatusResponse, error)
}

// Request/Response types
type CreateRequestRequest struct {
	WorkspaceID string
	UserID      string
	Title       string
	Description string
	Category    string
	Priority    string
}

type CreateRequestResponse struct {
	Request *Request
}

type GetRequestRequest struct {
	RequestID string
}

type GetRequestResponse struct {
	Request *Request
}

type ListRequestsRequest struct {
	WorkspaceID string
	Status      string
	Page        int32
	PageSize    int32
}

type ListRequestsResponse struct {
	Requests []*Request
	Total    int64
	Page     int32
	PageSize int32
}

type UpdateRequestRequest struct {
	RequestID   string
	UserID      string
	Title       *string
	Description *string
}

type UpdateRequestResponse struct {
	Request *Request
}

type CancelRequestRequest struct {
	RequestID string
	UserID    string
}

type CancelRequestResponse struct {
	Success bool
}

type AssignExpertRequest struct {
	RequestID  string
	ExpertID   string
	AssignedBy string
}

type AssignExpertResponse struct {
	Request *Request
}

type UpdateStatusRequest struct {
	RequestID string
	Status    string
	Comment   string
	ChangedBy string
}

type UpdateStatusResponse struct {
	Request *Request
}

// Request DTO for responses
type Request struct {
	ID          string
	WorkspaceID string
	UserID      string
	Title       string
	Description string
	Category    string
	Status      string
	ExpertID    string
	Priority    string
	CreatedAt   int64
	UpdatedAt   int64
}

// RequestServer implements RequestServiceServer.
type RequestServer struct {
	requestService *service.RequestService
}

// NewRequestServer creates a new RequestServer.
func NewRequestServer(requestService *service.RequestService) *RequestServer {
	return &RequestServer{requestService: requestService}
}

// CreateRequest creates a new request.
func (s *RequestServer) CreateRequest(ctx context.Context, req *CreateRequestRequest) (*CreateRequestResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace_id")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	request, err := s.requestService.CreateRequest(&service.CreateRequestInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Priority:    req.Priority,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &CreateRequestResponse{
		Request: requestToProto(request),
	}, nil
}

// GetRequest returns a request.
func (s *RequestServer) GetRequest(ctx context.Context, req *GetRequestRequest) (*GetRequestResponse, error) {
	requestID, err := uuid.Parse(req.RequestID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request_id")
	}

	request, err := s.requestService.GetRequest(requestID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &GetRequestResponse{
		Request: requestToProto(request),
	}, nil
}

// ListRequests returns requests for a workspace.
func (s *RequestServer) ListRequests(ctx context.Context, req *ListRequestsRequest) (*ListRequestsResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace_id")
	}

	result, err := s.requestService.ListRequests(&service.ListRequestsInput{
		WorkspaceID: workspaceID,
		Status:      req.Status,
		Page:        int(req.Page),
		PageSize:    int(req.PageSize),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	requests := make([]*Request, len(result.Requests))
	for i, r := range result.Requests {
		requests[i] = requestToProto(&r)
	}

	return &ListRequestsResponse{
		Requests: requests,
		Total:    result.Total,
		Page:     int32(result.Page),
		PageSize: int32(result.PageSize),
	}, nil
}

// UpdateRequest updates a request.
func (s *RequestServer) UpdateRequest(ctx context.Context, req *UpdateRequestRequest) (*UpdateRequestResponse, error) {
	requestID, err := uuid.Parse(req.RequestID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request_id")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	request, err := s.requestService.UpdateRequest(&service.UpdateRequestInput{
		RequestID:   requestID,
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &UpdateRequestResponse{
		Request: requestToProto(request),
	}, nil
}

// CancelRequest cancels a request.
func (s *RequestServer) CancelRequest(ctx context.Context, req *CancelRequestRequest) (*CancelRequestResponse, error) {
	requestID, err := uuid.Parse(req.RequestID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request_id")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	if err := s.requestService.CancelRequest(requestID, userID); err != nil {
		return nil, toGRPCError(err)
	}

	return &CancelRequestResponse{Success: true}, nil
}

// AssignExpert assigns an expert to a request.
func (s *RequestServer) AssignExpert(ctx context.Context, req *AssignExpertRequest) (*AssignExpertResponse, error) {
	requestID, err := uuid.Parse(req.RequestID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request_id")
	}

	expertID, err := uuid.Parse(req.ExpertID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid expert_id")
	}

	assignedBy, err := uuid.Parse(req.AssignedBy)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid assigned_by")
	}

	request, err := s.requestService.AssignExpert(&service.AssignExpertInput{
		RequestID:  requestID,
		ExpertID:   expertID,
		AssignedBy: assignedBy,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &AssignExpertResponse{
		Request: requestToProto(request),
	}, nil
}

// UpdateStatus updates request status.
func (s *RequestServer) UpdateStatus(ctx context.Context, req *UpdateStatusRequest) (*UpdateStatusResponse, error) {
	requestID, err := uuid.Parse(req.RequestID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request_id")
	}

	changedBy, err := uuid.Parse(req.ChangedBy)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid changed_by")
	}

	request, err := s.requestService.UpdateStatus(&service.UpdateStatusInput{
		RequestID: requestID,
		Status:    req.Status,
		Comment:   req.Comment,
		ChangedBy: changedBy,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &UpdateStatusResponse{
		Request: requestToProto(request),
	}, nil
}

// RegisterRequestServiceServer registers the request service server.
func RegisterRequestServiceServer(s *grpc.Server, srv RequestServiceServer) {
	// Will be generated from proto
}

// Helper functions
func requestToProto(r *repository.Request) *Request {
	expertID := ""
	if r.ExpertID != nil {
		expertID = r.ExpertID.String()
	}

	return &Request{
		ID:          r.ID.String(),
		WorkspaceID: r.WorkspaceID.String(),
		UserID:      r.UserID.String(),
		Title:       r.Title,
		Description: r.Description,
		Category:    r.Category,
		Status:      r.Status,
		ExpertID:    expertID,
		Priority:    r.Priority,
		CreatedAt:   r.CreatedAt.Unix(),
		UpdatedAt:   r.UpdatedAt.Unix(),
	}
}

func toGRPCError(err error) error {
	if e, ok := err.(*errors.Error); ok {
		return e.ToGRPCError()
	}
	return status.Error(codes.Internal, err.Error())
}


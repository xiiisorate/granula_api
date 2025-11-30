// =============================================================================
// Package grpc provides gRPC adapter for Request Service.
// =============================================================================
// This file adapts the RequestServer to implement the proto-generated
// RequestServiceServer interface.
// =============================================================================
package grpc

import (
	"context"
	"time"

	"github.com/xiiisorate/granula_api/request-service/internal/service"
	commonpb "github.com/xiiisorate/granula_api/shared/gen/common/v1"
	requestpb "github.com/xiiisorate/granula_api/shared/gen/request/v1"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// =============================================================================
// Proto Adapter
// =============================================================================

// RequestServiceAdapter implements the proto RequestServiceServer interface.
type RequestServiceAdapter struct {
	requestpb.UnimplementedRequestServiceServer
	server *RequestServer
	log    *logger.Logger
}

// NewRequestServiceAdapter creates a new adapter.
func NewRequestServiceAdapter(svc *service.RequestService, log *logger.Logger) *RequestServiceAdapter {
	return &RequestServiceAdapter{
		server: NewRequestServer(svc, log),
		log:    log,
	}
}

// CreateRequest creates a new expert request.
func (a *RequestServiceAdapter) CreateRequest(ctx context.Context, req *requestpb.CreateRequestRequest) (*requestpb.CreateRequestResponse, error) {
	a.log.Info("CreateRequest adapter called",
		logger.String("workspace_id", req.GetWorkspaceId()),
		logger.String("user_id", req.GetUserId()),
	)

	// Validate required fields
	if req.GetWorkspaceId() == "" {
		return nil, status.Error(codes.InvalidArgument, "workspace_id is required")
	}
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	// Convert category from proto enum to string
	category := protoCategoryToString(req.GetCategory())

	// Convert to internal DTO
	dto := &CreateRequestDTO{
		WorkspaceID:  req.GetWorkspaceId(),
		UserID:       req.GetUserId(),
		Title:        req.GetTitle(),
		Description:  req.GetDescription(),
		Category:     category,
		ContactPhone: "",
		ContactEmail: "",
	}

	// Extract contact info
	if req.GetContact() != nil {
		dto.ContactPhone = req.GetContact().GetPhone()
		dto.ContactEmail = req.GetContact().GetEmail()
	}

	// Call internal server
	result, err := a.server.CreateRequest(ctx, dto)
	if err != nil {
		return nil, err
	}

	// Convert to proto response
	return &requestpb.CreateRequestResponse{
		Request: dtoToProtoRequest(result),
	}, nil
}

// GetRequest retrieves a request by ID.
func (a *RequestServiceAdapter) GetRequest(ctx context.Context, req *requestpb.GetRequestRequest) (*requestpb.GetRequestResponse, error) {
	// GetRequest requires both request_id and user_id for authorization
	// For now we pass empty user_id, service will handle it
	result, err := a.server.GetRequest(ctx, req.GetRequestId(), "")
	if err != nil {
		return nil, err
	}

	return &requestpb.GetRequestResponse{
		Request: dtoToProtoRequest(result),
	}, nil
}

// ListRequests lists requests with pagination.
func (a *RequestServiceAdapter) ListRequests(ctx context.Context, req *requestpb.ListRequestsRequest) (*requestpb.ListRequestsResponse, error) {
	// Determine page and limit
	var limit int32 = 20
	var offset int32 = 0
	if req.GetPagination() != nil {
		if req.GetPagination().GetPageSize() > 0 {
			limit = req.GetPagination().GetPageSize()
		}
		if req.GetPagination().GetPage() > 0 {
			offset = (req.GetPagination().GetPage() - 1) * limit
		}
	}

	// Build list DTO
	listDTO := &ListRequestsDTO{
		UserID:      req.GetUserId(),
		WorkspaceID: req.GetWorkspaceId(),
		Limit:       limit,
		Offset:      offset,
	}

	result, err := a.server.ListRequests(ctx, listDTO)
	if err != nil {
		return nil, err
	}

	// Convert results
	items := make([]*requestpb.Request, len(result.Requests))
	for i, r := range result.Requests {
		items[i] = dtoToProtoRequest(r)
	}

	return &requestpb.ListRequestsResponse{
		Requests: items,
		Pagination: &commonpb.PaginationResponse{
			Total:      result.Total,
			TotalPages: (result.Total + limit - 1) / limit,
			Page:       offset/limit + 1,
			PageSize:   limit,
		},
	}, nil
}

// CancelRequest cancels a request.
func (a *RequestServiceAdapter) CancelRequest(ctx context.Context, req *requestpb.CancelRequestRequest) (*requestpb.CancelRequestResponse, error) {
	// TODO: Implement when CancelRequest is added to RequestServer
	return nil, status.Error(codes.Unimplemented, "CancelRequest not implemented")
}

// UpdateRequest updates request fields.
func (a *RequestServiceAdapter) UpdateRequest(ctx context.Context, req *requestpb.UpdateRequestRequest) (*requestpb.UpdateRequestResponse, error) {
	return nil, status.Error(codes.Unimplemented, "UpdateRequest not implemented")
}

// AssignExpert assigns an expert to a request.
func (a *RequestServiceAdapter) AssignExpert(ctx context.Context, req *requestpb.AssignExpertRequest) (*requestpb.AssignExpertResponse, error) {
	return nil, status.Error(codes.Unimplemented, "AssignExpert not implemented")
}

// UpdateStatus updates request status.
func (a *RequestServiceAdapter) UpdateStatus(ctx context.Context, req *requestpb.UpdateStatusRequest) (*requestpb.UpdateStatusResponse, error) {
	return nil, status.Error(codes.Unimplemented, "UpdateStatus not implemented")
}

// RejectRequest rejects a request.
func (a *RequestServiceAdapter) RejectRequest(ctx context.Context, req *requestpb.RejectRequestRequest) (*requestpb.RejectRequestResponse, error) {
	return nil, status.Error(codes.Unimplemented, "RejectRequest not implemented")
}

// CompleteRequest marks request as complete.
func (a *RequestServiceAdapter) CompleteRequest(ctx context.Context, req *requestpb.CompleteRequestRequest) (*requestpb.CompleteRequestResponse, error) {
	return nil, status.Error(codes.Unimplemented, "CompleteRequest not implemented")
}

// UploadDocument uploads a document for a request.
func (a *RequestServiceAdapter) UploadDocument(ctx context.Context, req *requestpb.UploadDocumentRequest) (*requestpb.UploadDocumentResponse, error) {
	return nil, status.Error(codes.Unimplemented, "UploadDocument not implemented")
}

// GetDocuments gets all documents for a request.
func (a *RequestServiceAdapter) GetDocuments(ctx context.Context, req *requestpb.GetDocumentsRequest) (*requestpb.GetDocumentsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetDocuments not implemented")
}

// GetStatusHistory gets status change history.
func (a *RequestServiceAdapter) GetStatusHistory(ctx context.Context, req *requestpb.GetStatusHistoryRequest) (*requestpb.GetStatusHistoryResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetStatusHistory not implemented")
}

// =============================================================================
// Helpers
// =============================================================================

// dtoToProtoRequest converts internal DTO to proto Request message.
func dtoToProtoRequest(dto *RequestResponseDTO) *requestpb.Request {
	if dto == nil {
		return nil
	}

	req := &requestpb.Request{
		Id:          dto.ID,
		WorkspaceId: dto.WorkspaceID,
		UserId:      dto.UserID,
		Title:       dto.Title,
		Description: dto.Description,
		Category:    stringToProtoCategory(dto.Category),
		Priority:    stringToProtoPriority(dto.Priority),
		Status:      stringToProtoStatus(dto.Status),
		Contact: &requestpb.Contact{
			Phone: dto.ContactPhone,
			Email: dto.ContactEmail,
		},
	}

	// Convert timestamps
	if dto.CreatedAt > 0 {
		req.CreatedAt = timestamppb.New(time.Unix(dto.CreatedAt, 0))
	}
	if dto.UpdatedAt > 0 {
		req.UpdatedAt = timestamppb.New(time.Unix(dto.UpdatedAt, 0))
	}

	return req
}

func protoCategoryToString(c requestpb.RequestCategory) string {
	switch c {
	case requestpb.RequestCategory_REQUEST_CATEGORY_CONSULTATION:
		return "consultation"
	case requestpb.RequestCategory_REQUEST_CATEGORY_VERIFICATION:
		return "verification"
	case requestpb.RequestCategory_REQUEST_CATEGORY_PROJECT:
		return "project"
	case requestpb.RequestCategory_REQUEST_CATEGORY_APPROVAL:
		return "approval"
	default:
		return "consultation"
	}
}

func stringToProtoCategory(s string) requestpb.RequestCategory {
	switch s {
	case "consultation":
		return requestpb.RequestCategory_REQUEST_CATEGORY_CONSULTATION
	case "verification":
		return requestpb.RequestCategory_REQUEST_CATEGORY_VERIFICATION
	case "project":
		return requestpb.RequestCategory_REQUEST_CATEGORY_PROJECT
	case "approval":
		return requestpb.RequestCategory_REQUEST_CATEGORY_APPROVAL
	default:
		return requestpb.RequestCategory_REQUEST_CATEGORY_UNSPECIFIED
	}
}

func stringToProtoPriority(s string) requestpb.RequestPriority {
	switch s {
	case "low":
		return requestpb.RequestPriority_REQUEST_PRIORITY_LOW
	case "normal":
		return requestpb.RequestPriority_REQUEST_PRIORITY_NORMAL
	case "high":
		return requestpb.RequestPriority_REQUEST_PRIORITY_HIGH
	case "urgent":
		return requestpb.RequestPriority_REQUEST_PRIORITY_URGENT
	default:
		return requestpb.RequestPriority_REQUEST_PRIORITY_UNSPECIFIED
	}
}

func stringToProtoStatus(s string) requestpb.RequestStatus {
	switch s {
	case "draft":
		return requestpb.RequestStatus_REQUEST_STATUS_DRAFT
	case "pending":
		return requestpb.RequestStatus_REQUEST_STATUS_PENDING
	case "in_review":
		return requestpb.RequestStatus_REQUEST_STATUS_IN_REVIEW
	case "approved":
		return requestpb.RequestStatus_REQUEST_STATUS_APPROVED
	case "rejected":
		return requestpb.RequestStatus_REQUEST_STATUS_REJECTED
	case "completed":
		return requestpb.RequestStatus_REQUEST_STATUS_COMPLETED
	case "cancelled":
		return requestpb.RequestStatus_REQUEST_STATUS_CANCELLED
	default:
		return requestpb.RequestStatus_REQUEST_STATUS_UNSPECIFIED
	}
}

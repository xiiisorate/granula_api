// =============================================================================
// Package grpc provides gRPC server implementation for Workspace Service.
// =============================================================================
// This package implements the WorkspaceService gRPC interface defined in
// shared/proto/workspace/v1/workspace.proto. It handles request validation,
// conversion between protobuf and domain types, and error mapping.
//
// Error Handling:
//   - Domain errors are mapped to appropriate gRPC status codes
//   - All errors include descriptive messages
//   - Internal errors are logged but not exposed to clients
//
// =============================================================================
package grpc

import (
	"context"

	"github.com/google/uuid"
	apperrors "github.com/xiiisorate/granula_api/shared/pkg/errors"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
	"github.com/xiiisorate/granula_api/workspace-service/internal/domain/entity"
	"github.com/xiiisorate/granula_api/workspace-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// =============================================================================
// WorkspaceServer Implementation
// =============================================================================

// WorkspaceServer implements the gRPC WorkspaceService interface.
// It delegates all business logic to the service layer and handles
// request/response conversion.
type WorkspaceServer struct {
	// UnimplementedWorkspaceServiceServer provides forward compatibility
	// UnimplementedWorkspaceServiceServer

	// service contains the business logic
	service *service.WorkspaceService

	// log is used for request logging and debugging
	log *logger.Logger
}

// NewWorkspaceServer creates a new WorkspaceServer instance.
//
// Parameters:
//   - svc: Business logic service
//   - log: Logger for operational logging
//
// Returns:
//   - *WorkspaceServer: Ready to register with gRPC server
func NewWorkspaceServer(svc *service.WorkspaceService, log *logger.Logger) *WorkspaceServer {
	return &WorkspaceServer{
		service: svc,
		log:     log,
	}
}

// =============================================================================
// Workspace CRUD Methods
// =============================================================================

// CreateWorkspaceRequest represents a request to create a workspace.
type CreateWorkspaceRequest struct {
	Name        string
	Description string
	OwnerID     string
}

// CreateWorkspaceResponse represents the response from creating a workspace.
type CreateWorkspaceResponse struct {
	Workspace *WorkspaceDTO
}

// WorkspaceDTO is a data transfer object for workspace data.
type WorkspaceDTO struct {
	ID           string
	OwnerID      string
	Name         string
	Description  string
	MemberCount  int32
	ProjectCount int32
	CreatedAt    int64
	UpdatedAt    int64
	Members      []*MemberDTO
}

// MemberDTO is a data transfer object for member data.
type MemberDTO struct {
	ID       string
	UserID   string
	Role     string
	JoinedAt int64
}

// CreateWorkspace handles workspace creation requests.
//
// Request Fields:
//   - name: Workspace name (required, 2-100 chars)
//   - description: Optional description
//   - owner_id: UUID of the creating user (from auth context)
//
// Response Fields:
//   - workspace: Created workspace with ID
//
// Errors:
//   - INVALID_ARGUMENT: Invalid name or owner_id
//   - INTERNAL: Database error
func (s *WorkspaceServer) CreateWorkspace(ctx context.Context, req *CreateWorkspaceRequest) (*CreateWorkspaceResponse, error) {
	s.log.Info("CreateWorkspace called",
		logger.String("name", req.Name),
		logger.String("owner_id", req.OwnerID),
	)

	// Validate owner_id
	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner_id: must be valid UUID")
	}

	// Call service
	ws, err := s.service.CreateWorkspace(ctx, ownerID, req.Name, req.Description)
	if err != nil {
		return nil, s.mapError(err)
	}

	return &CreateWorkspaceResponse{
		Workspace: workspaceToDTO(ws),
	}, nil
}

// GetWorkspaceRequest represents a request to get a workspace.
type GetWorkspaceRequest struct {
	WorkspaceID string
	UserID      string // For access control
}

// GetWorkspaceResponse represents the response from getting a workspace.
type GetWorkspaceResponse struct {
	Workspace *WorkspaceDTO
}

// GetWorkspace retrieves a workspace by ID.
//
// Request Fields:
//   - workspace_id: UUID of the workspace
//   - user_id: UUID of requesting user (from auth context)
//
// Response Fields:
//   - workspace: Workspace with members
//
// Errors:
//   - INVALID_ARGUMENT: Invalid workspace_id
//   - NOT_FOUND: Workspace doesn't exist
//   - PERMISSION_DENIED: User is not a member
func (s *WorkspaceServer) GetWorkspace(ctx context.Context, req *GetWorkspaceRequest) (*GetWorkspaceResponse, error) {
	// Validate IDs
	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace_id")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	// Call service
	ws, err := s.service.GetWorkspace(ctx, workspaceID, userID)
	if err != nil {
		return nil, s.mapError(err)
	}

	return &GetWorkspaceResponse{
		Workspace: workspaceToDTO(ws),
	}, nil
}

// ListWorkspacesRequest represents a request to list workspaces.
type ListWorkspacesRequest struct {
	UserID     string
	NameFilter string
	Limit      int32
	Offset     int32
}

// ListWorkspacesResponse represents the response from listing workspaces.
type ListWorkspacesResponse struct {
	Workspaces []*WorkspaceDTO
	Total      int32
}

// ListWorkspaces returns workspaces where the user is a member.
//
// Request Fields:
//   - user_id: UUID of the user
//   - name_filter: Optional name filter (case-insensitive contains)
//   - limit: Max results (1-100, default 20)
//   - offset: Pagination offset
//
// Response Fields:
//   - workspaces: List of workspaces
//   - total: Total count for pagination
func (s *WorkspaceServer) ListWorkspaces(ctx context.Context, req *ListWorkspacesRequest) (*ListWorkspacesResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	workspaces, total, err := s.service.ListWorkspaces(ctx, userID, req.NameFilter, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, s.mapError(err)
	}

	dtos := make([]*WorkspaceDTO, 0, len(workspaces))
	for _, ws := range workspaces {
		dtos = append(dtos, workspaceToDTO(ws))
	}

	return &ListWorkspacesResponse{
		Workspaces: dtos,
		Total:      int32(total),
	}, nil
}

// UpdateWorkspaceRequest represents a request to update a workspace.
type UpdateWorkspaceRequest struct {
	WorkspaceID string
	UserID      string
	Name        string
	Description string
}

// UpdateWorkspaceResponse represents the response from updating a workspace.
type UpdateWorkspaceResponse struct {
	Workspace *WorkspaceDTO
}

// UpdateWorkspace updates workspace name and/or description.
//
// Request Fields:
//   - workspace_id: UUID of the workspace
//   - user_id: UUID of requesting user (must be owner/admin)
//   - name: New name (empty to keep current)
//   - description: New description (empty to keep current)
//
// Errors:
//   - INVALID_ARGUMENT: Invalid IDs or name
//   - NOT_FOUND: Workspace doesn't exist
//   - PERMISSION_DENIED: User lacks permission
func (s *WorkspaceServer) UpdateWorkspace(ctx context.Context, req *UpdateWorkspaceRequest) (*UpdateWorkspaceResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace_id")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	ws, err := s.service.UpdateWorkspace(ctx, workspaceID, userID, req.Name, req.Description)
	if err != nil {
		return nil, s.mapError(err)
	}

	return &UpdateWorkspaceResponse{
		Workspace: workspaceToDTO(ws),
	}, nil
}

// DeleteWorkspaceRequest represents a request to delete a workspace.
type DeleteWorkspaceRequest struct {
	WorkspaceID string
	UserID      string
}

// DeleteWorkspaceResponse represents the response from deleting a workspace.
type DeleteWorkspaceResponse struct {
	Success bool
}

// DeleteWorkspace permanently removes a workspace.
//
// Request Fields:
//   - workspace_id: UUID of the workspace
//   - user_id: UUID of requesting user (must be owner)
//
// Errors:
//   - INVALID_ARGUMENT: Invalid IDs
//   - NOT_FOUND: Workspace doesn't exist
//   - PERMISSION_DENIED: User is not the owner
func (s *WorkspaceServer) DeleteWorkspace(ctx context.Context, req *DeleteWorkspaceRequest) (*DeleteWorkspaceResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace_id")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	if err := s.service.DeleteWorkspace(ctx, workspaceID, userID); err != nil {
		return nil, s.mapError(err)
	}

	return &DeleteWorkspaceResponse{Success: true}, nil
}

// =============================================================================
// Member Management Methods
// =============================================================================

// AddMemberRequest represents a request to add a member.
type AddMemberRequest struct {
	WorkspaceID  string
	UserID       string
	MemberUserID string
	Role         string
}

// AddMemberResponse represents the response from adding a member.
type AddMemberResponse struct {
	Member *MemberDTO
}

// AddMember adds a user to a workspace.
//
// Request Fields:
//   - workspace_id: UUID of the workspace
//   - user_id: UUID of requesting user (must be owner/admin)
//   - member_user_id: UUID of user to add
//   - role: Role to assign (admin, editor, viewer)
//
// Errors:
//   - INVALID_ARGUMENT: Invalid IDs or role
//   - ALREADY_EXISTS: User is already a member
//   - PERMISSION_DENIED: Caller lacks permission
func (s *WorkspaceServer) AddMember(ctx context.Context, req *AddMemberRequest) (*AddMemberResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace_id")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	memberUserID, err := uuid.Parse(req.MemberUserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid member_user_id")
	}

	role := entity.MemberRole(req.Role)
	if !role.IsValid() || role == entity.RoleOwner {
		return nil, status.Error(codes.InvalidArgument, "invalid role: must be admin, editor, or viewer")
	}

	member, err := s.service.AddMember(ctx, workspaceID, userID, memberUserID, role)
	if err != nil {
		return nil, s.mapError(err)
	}

	return &AddMemberResponse{
		Member: memberToDTO(member),
	}, nil
}

// RemoveMemberRequest represents a request to remove a member.
type RemoveMemberRequest struct {
	WorkspaceID  string
	UserID       string
	MemberUserID string
}

// RemoveMemberResponse represents the response from removing a member.
type RemoveMemberResponse struct {
	Success bool
}

// RemoveMember removes a user from a workspace.
//
// Request Fields:
//   - workspace_id: UUID of the workspace
//   - user_id: UUID of requesting user
//   - member_user_id: UUID of user to remove
//
// Errors:
//   - INVALID_ARGUMENT: Invalid IDs
//   - FAILED_PRECONDITION: Cannot remove owner
//   - NOT_FOUND: Member not found
//   - PERMISSION_DENIED: Caller lacks permission
func (s *WorkspaceServer) RemoveMember(ctx context.Context, req *RemoveMemberRequest) (*RemoveMemberResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace_id")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	memberUserID, err := uuid.Parse(req.MemberUserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid member_user_id")
	}

	if err := s.service.RemoveMember(ctx, workspaceID, userID, memberUserID); err != nil {
		return nil, s.mapError(err)
	}

	return &RemoveMemberResponse{Success: true}, nil
}

// =============================================================================
// Error Mapping
// =============================================================================

// mapError converts domain errors to gRPC status errors.
func (s *WorkspaceServer) mapError(err error) error {
	switch err {
	case entity.ErrWorkspaceNotFound:
		return status.Error(codes.NotFound, "workspace not found")
	case entity.ErrMemberNotFound:
		return status.Error(codes.NotFound, "member not found")
	case entity.ErrMemberAlreadyExists:
		return status.Error(codes.AlreadyExists, "user is already a member")
	case entity.ErrCannotRemoveOwner:
		return status.Error(codes.FailedPrecondition, "cannot remove workspace owner")
	case entity.ErrInvalidWorkspaceName:
		return status.Error(codes.InvalidArgument, "invalid workspace name: must be 2-100 characters")
	case entity.ErrInvalidRole:
		return status.Error(codes.InvalidArgument, "invalid role")
	case entity.ErrOwnerCannotLeave:
		return status.Error(codes.FailedPrecondition, "owner cannot leave workspace")
	case service.ErrForbidden, service.ErrUnauthorized:
		return status.Error(codes.PermissionDenied, "access denied")
	default:
		// Check for wrapped app errors
		if appErr, ok := err.(*apperrors.Error); ok {
			return appErr.ToGRPCError()
		}
		// Log and return generic error
		s.log.Error("internal error", logger.Err(err))
		return status.Error(codes.Internal, "internal server error")
	}
}

// =============================================================================
// DTO Conversion Functions
// =============================================================================

// workspaceToDTO converts a domain workspace to a DTO.
func workspaceToDTO(ws *entity.Workspace) *WorkspaceDTO {
	dto := &WorkspaceDTO{
		ID:           ws.ID.String(),
		OwnerID:      ws.OwnerID.String(),
		Name:         ws.Name,
		Description:  ws.Description,
		MemberCount:  int32(ws.MemberCount),
		ProjectCount: int32(ws.ProjectCount),
		CreatedAt:    ws.CreatedAt.Unix(),
		UpdatedAt:    ws.UpdatedAt.Unix(),
	}

	if len(ws.Members) > 0 {
		dto.Members = make([]*MemberDTO, 0, len(ws.Members))
		for _, m := range ws.Members {
			dto.Members = append(dto.Members, memberToDTO(&m))
		}
	}

	return dto
}

// memberToDTO converts a domain member to a DTO.
func memberToDTO(m *entity.Member) *MemberDTO {
	return &MemberDTO{
		ID:       m.ID.String(),
		UserID:   m.UserID.String(),
		Role:     string(m.Role),
		JoinedAt: m.JoinedAt.Unix(),
	}
}

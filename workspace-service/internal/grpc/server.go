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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonv1 "github.com/xiiisorate/granula_api/shared/gen/common/v1"
	workspacev1 "github.com/xiiisorate/granula_api/shared/gen/workspace/v1"
	apperrors "github.com/xiiisorate/granula_api/shared/pkg/errors"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
	"github.com/xiiisorate/granula_api/workspace-service/internal/domain/entity"
	"github.com/xiiisorate/granula_api/workspace-service/internal/service"
)

// =============================================================================
// WorkspaceServer Implementation
// =============================================================================

// WorkspaceServer implements the gRPC WorkspaceServiceServer interface.
// It delegates all business logic to the service layer and handles
// request/response conversion.
type WorkspaceServer struct {
	// Embed UnimplementedWorkspaceServiceServer for forward compatibility
	workspacev1.UnimplementedWorkspaceServiceServer

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
// CreateWorkspace - creates a new workspace
// =============================================================================

// CreateWorkspace handles workspace creation requests.
//
// Request Fields:
//   - name: Workspace name (required, 2-100 chars)
//   - description: Optional description
//
// Response Fields:
//   - workspace: Created workspace with ID
//
// Errors:
//   - INVALID_ARGUMENT: Invalid name or owner_id
//   - INTERNAL: Database error
func (s *WorkspaceServer) CreateWorkspace(ctx context.Context, req *workspacev1.CreateWorkspaceRequest) (*workspacev1.CreateWorkspaceResponse, error) {
	s.log.Info("CreateWorkspace called",
		logger.String("name", req.Name),
	)

	// Get owner ID from context (would be set by auth middleware in production)
	// For now, we'll generate one as placeholder
	ownerID := uuid.New()

	// Call service
	ws, err := s.service.CreateWorkspace(ctx, ownerID, req.Name, req.Description)
	if err != nil {
		return nil, s.mapError(err)
	}

	return &workspacev1.CreateWorkspaceResponse{
		Workspace: entityToProto(ws),
	}, nil
}

// =============================================================================
// GetWorkspace - retrieves a workspace by ID
// =============================================================================

// GetWorkspace retrieves a workspace by ID.
//
// Request Fields:
//   - workspace_id: UUID of the workspace
//   - include_members: Whether to include member list
//
// Response Fields:
//   - workspace: Workspace with members (if requested)
//
// Errors:
//   - INVALID_ARGUMENT: Invalid workspace_id
//   - NOT_FOUND: Workspace doesn't exist
//   - PERMISSION_DENIED: User is not a member
func (s *WorkspaceServer) GetWorkspace(ctx context.Context, req *workspacev1.GetWorkspaceRequest) (*workspacev1.GetWorkspaceResponse, error) {
	// Validate workspace ID
	workspaceID, err := uuid.Parse(req.WorkspaceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace_id")
	}

	// Get user ID from context (placeholder)
	userID := uuid.New()

	// Call service
	ws, err := s.service.GetWorkspace(ctx, workspaceID, userID)
	if err != nil {
		return nil, s.mapError(err)
	}

	return &workspacev1.GetWorkspaceResponse{
		Workspace: entityToProto(ws),
	}, nil
}

// =============================================================================
// ListWorkspaces - returns workspaces for a user
// =============================================================================

// ListWorkspaces returns workspaces where the user is a member.
//
// Request Fields:
//   - user_id: UUID of the user
//   - status: Optional status filter
//   - search: Optional name filter
//   - pagination: Pagination parameters
//
// Response Fields:
//   - workspaces: List of workspaces
//   - pagination: Pagination metadata
func (s *WorkspaceServer) ListWorkspaces(ctx context.Context, req *workspacev1.ListWorkspacesRequest) (*workspacev1.ListWorkspacesResponse, error) {
	// Parse user ID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	// Pagination defaults
	pageSize := int32(20)
	page := int32(1)
	if req.Pagination != nil {
		if req.Pagination.PageSize > 0 {
			pageSize = req.Pagination.PageSize
		}
		if req.Pagination.Page > 0 {
			page = req.Pagination.Page
		}
	}
	offset := int((page - 1) * pageSize)

	// Call service
	workspaces, total, err := s.service.ListWorkspaces(ctx, userID, req.Search, int(pageSize), offset)
	if err != nil {
		return nil, s.mapError(err)
	}

	// Convert to proto
	protoWorkspaces := make([]*workspacev1.Workspace, 0, len(workspaces))
	for _, ws := range workspaces {
		protoWorkspaces = append(protoWorkspaces, entityToProto(ws))
	}

	// Calculate pagination metadata
	totalPages := int32(total) / pageSize
	if int32(total)%pageSize != 0 {
		totalPages++
	}

	return &workspacev1.ListWorkspacesResponse{
		Workspaces: protoWorkspaces,
		Pagination: &commonv1.PaginationResponse{
			Page:       page,
			PageSize:   pageSize,
			Total:      int32(total),
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
	}, nil
}

// =============================================================================
// UpdateWorkspace - updates workspace metadata
// =============================================================================

// UpdateWorkspace updates workspace name and/or description.
//
// Request Fields:
//   - workspace_id: UUID of the workspace
//   - name: New name (empty to keep current)
//   - description: New description (empty to keep current)
//
// Errors:
//   - INVALID_ARGUMENT: Invalid IDs or name
//   - NOT_FOUND: Workspace doesn't exist
//   - PERMISSION_DENIED: User lacks permission
func (s *WorkspaceServer) UpdateWorkspace(ctx context.Context, req *workspacev1.UpdateWorkspaceRequest) (*workspacev1.UpdateWorkspaceResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace_id")
	}

	// Get user ID from context (placeholder)
	userID := uuid.New()

	ws, err := s.service.UpdateWorkspace(ctx, workspaceID, userID, req.Name, req.Description)
	if err != nil {
		return nil, s.mapError(err)
	}

	return &workspacev1.UpdateWorkspaceResponse{
		Workspace: entityToProto(ws),
	}, nil
}

// =============================================================================
// DeleteWorkspace - deletes a workspace
// =============================================================================

// DeleteWorkspace removes a workspace.
//
// Request Fields:
//   - workspace_id: UUID of the workspace
//   - permanent: Whether to permanently delete
//
// Errors:
//   - INVALID_ARGUMENT: Invalid IDs
//   - NOT_FOUND: Workspace doesn't exist
//   - PERMISSION_DENIED: User is not the owner
func (s *WorkspaceServer) DeleteWorkspace(ctx context.Context, req *workspacev1.DeleteWorkspaceRequest) (*workspacev1.DeleteWorkspaceResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace_id")
	}

	// Get user ID from context (placeholder)
	userID := uuid.New()

	if err := s.service.DeleteWorkspace(ctx, workspaceID, userID); err != nil {
		return nil, s.mapError(err)
	}

	return &workspacev1.DeleteWorkspaceResponse{Success: true}, nil
}

// =============================================================================
// AddMember - adds a member to workspace
// =============================================================================

// AddMember adds a user to a workspace.
//
// Request Fields:
//   - workspace_id: UUID of the workspace
//   - user_id_or_email: UUID or email of user to add
//   - role: Role to assign
//
// Errors:
//   - INVALID_ARGUMENT: Invalid IDs or role
//   - ALREADY_EXISTS: User is already a member
//   - PERMISSION_DENIED: Caller lacks permission
func (s *WorkspaceServer) AddMember(ctx context.Context, req *workspacev1.AddMemberRequest) (*workspacev1.AddMemberResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace_id")
	}

	// Get caller user ID from context (placeholder)
	callerUserID := uuid.New()

	// Try to parse as UUID, otherwise treat as email (for future invite support)
	memberUserID, err := uuid.Parse(req.UserIdOrEmail)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid member user_id")
	}

	role := protoRoleToEntity(req.Role)

	member, err := s.service.AddMember(ctx, workspaceID, callerUserID, memberUserID, role)
	if err != nil {
		return nil, s.mapError(err)
	}

	return &workspacev1.AddMemberResponse{
		Member: memberToProto(member),
	}, nil
}

// =============================================================================
// RemoveMember - removes a member from workspace
// =============================================================================

// RemoveMember removes a user from a workspace.
//
// Request Fields:
//   - workspace_id: UUID of the workspace
//   - user_id: UUID of user to remove
//
// Errors:
//   - INVALID_ARGUMENT: Invalid IDs
//   - FAILED_PRECONDITION: Cannot remove owner
//   - NOT_FOUND: Member not found
//   - PERMISSION_DENIED: Caller lacks permission
func (s *WorkspaceServer) RemoveMember(ctx context.Context, req *workspacev1.RemoveMemberRequest) (*workspacev1.RemoveMemberResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace_id")
	}

	// Get caller user ID from context (placeholder)
	callerUserID := uuid.New()

	memberUserID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	if err := s.service.RemoveMember(ctx, workspaceID, callerUserID, memberUserID); err != nil {
		return nil, s.mapError(err)
	}

	return &workspacev1.RemoveMemberResponse{Success: true}, nil
}

// =============================================================================
// UpdateMemberRole - changes a member's role
// =============================================================================

// UpdateMemberRole changes a member's role in a workspace.
func (s *WorkspaceServer) UpdateMemberRole(ctx context.Context, req *workspacev1.UpdateMemberRoleRequest) (*workspacev1.UpdateMemberRoleResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace_id")
	}

	// Get caller user ID from context (placeholder)
	callerUserID := uuid.New()

	memberUserID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	role := protoRoleToEntity(req.Role)

	member, err := s.service.UpdateMemberRole(ctx, workspaceID, callerUserID, memberUserID, role)
	if err != nil {
		return nil, s.mapError(err)
	}

	return &workspacev1.UpdateMemberRoleResponse{
		Member: memberToProto(member),
	}, nil
}

// =============================================================================
// GetMembers - retrieves workspace members
// =============================================================================

// GetMembers retrieves all members of a workspace.
func (s *WorkspaceServer) GetMembers(ctx context.Context, req *workspacev1.GetMembersRequest) (*workspacev1.GetMembersResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace_id")
	}

	// Get caller user ID from context (placeholder)
	userID := uuid.New()

	members, err := s.service.GetMembers(ctx, workspaceID, userID)
	if err != nil {
		return nil, s.mapError(err)
	}

	protoMembers := make([]*workspacev1.Member, 0, len(members))
	for _, m := range members {
		protoMembers = append(protoMembers, memberToProto(m))
	}

	return &workspacev1.GetMembersResponse{
		Members: protoMembers,
	}, nil
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
// Proto Conversion Functions
// =============================================================================

// entityToProto converts a domain workspace to a protobuf workspace.
func entityToProto(ws *entity.Workspace) *workspacev1.Workspace {
	proto := &workspacev1.Workspace{
		Id:              ws.ID.String(),
		Name:            ws.Name,
		Description:     ws.Description,
		OwnerId:         ws.OwnerID.String(),
		Address:         ws.Address,
		Status:          entityStatusToProto(ws.Status),
		FloorPlansCount: int32(ws.FloorPlansCount),
		ScenesCount:     int32(ws.ScenesCount),
		CreatedAt:       timestamppb.New(ws.CreatedAt),
		UpdatedAt:       timestamppb.New(ws.UpdatedAt),
	}

	// Handle optional pointer fields
	if ws.TotalArea != nil {
		proto.TotalArea = *ws.TotalArea
	}
	if ws.RoomsCount != nil {
		proto.RoomsCount = int32(*ws.RoomsCount)
	}

	// Add settings if present
	if ws.Settings != nil {
		proto.Settings = &workspacev1.WorkspaceSettings{
			PropertyType:         ws.Settings.PropertyType,
			ProjectType:          ws.Settings.ProjectType,
			Units:                ws.Settings.Units,
			DefaultCeilingHeight: ws.Settings.DefaultCeilingHeight,
			DefaultWallThickness: ws.Settings.DefaultWallThickness,
			Currency:             ws.Settings.Currency,
			Region:               ws.Settings.Region,
			AutoComplianceCheck:  ws.Settings.AutoComplianceCheck,
			NotificationsEnabled: ws.Settings.NotificationsEnabled,
		}
	}

	// Add members if present
	if len(ws.Members) > 0 {
		proto.Members = make([]*workspacev1.Member, 0, len(ws.Members))
		for _, m := range ws.Members {
			proto.Members = append(proto.Members, memberToProto(&m))
		}
	}

	return proto
}

// memberToProto converts a domain member to a protobuf member.
func memberToProto(m *entity.Member) *workspacev1.Member {
	return &workspacev1.Member{
		UserId:    m.UserID.String(),
		Role:      entityRoleToProto(m.Role),
		Name:      m.Name,
		Email:     m.Email,
		AvatarUrl: m.AvatarURL,
		JoinedAt:  timestamppb.New(m.JoinedAt),
		InvitedBy: m.InvitedBy.String(),
	}
}

// entityStatusToProto converts domain status to proto status.
func entityStatusToProto(s entity.WorkspaceStatus) workspacev1.WorkspaceStatus {
	switch s {
	case entity.StatusActive:
		return workspacev1.WorkspaceStatus_WORKSPACE_STATUS_ACTIVE
	case entity.StatusArchived:
		return workspacev1.WorkspaceStatus_WORKSPACE_STATUS_ARCHIVED
	case entity.StatusDeleted:
		return workspacev1.WorkspaceStatus_WORKSPACE_STATUS_DELETED
	default:
		return workspacev1.WorkspaceStatus_WORKSPACE_STATUS_UNSPECIFIED
	}
}

// entityRoleToProto converts domain role to proto role.
func entityRoleToProto(r entity.MemberRole) workspacev1.MemberRole {
	switch r {
	case entity.RoleOwner:
		return workspacev1.MemberRole_MEMBER_ROLE_OWNER
	case entity.RoleAdmin:
		return workspacev1.MemberRole_MEMBER_ROLE_ADMIN
	case entity.RoleEditor:
		return workspacev1.MemberRole_MEMBER_ROLE_EDITOR
	case entity.RoleViewer:
		return workspacev1.MemberRole_MEMBER_ROLE_VIEWER
	default:
		return workspacev1.MemberRole_MEMBER_ROLE_UNSPECIFIED
	}
}

// protoRoleToEntity converts proto role to domain role.
func protoRoleToEntity(r workspacev1.MemberRole) entity.MemberRole {
	switch r {
	case workspacev1.MemberRole_MEMBER_ROLE_OWNER:
		return entity.RoleOwner
	case workspacev1.MemberRole_MEMBER_ROLE_ADMIN:
		return entity.RoleAdmin
	case workspacev1.MemberRole_MEMBER_ROLE_EDITOR:
		return entity.RoleEditor
	case workspacev1.MemberRole_MEMBER_ROLE_VIEWER:
		return entity.RoleViewer
	default:
		return entity.RoleViewer
	}
}

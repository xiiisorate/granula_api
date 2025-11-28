// Package server implements gRPC server for Workspace Service.
package server

import (
	"context"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/shared/pkg/errors"
	"github.com/xiiisorate/granula_api/workspace-service/internal/repository"
	"github.com/xiiisorate/granula_api/workspace-service/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WorkspaceServiceServer is the interface for Workspace gRPC service.
type WorkspaceServiceServer interface {
	CreateWorkspace(ctx context.Context, req *CreateWorkspaceRequest) (*CreateWorkspaceResponse, error)
	GetWorkspace(ctx context.Context, req *GetWorkspaceRequest) (*GetWorkspaceResponse, error)
	ListWorkspaces(ctx context.Context, req *ListWorkspacesRequest) (*ListWorkspacesResponse, error)
	UpdateWorkspace(ctx context.Context, req *UpdateWorkspaceRequest) (*UpdateWorkspaceResponse, error)
	DeleteWorkspace(ctx context.Context, req *DeleteWorkspaceRequest) (*DeleteWorkspaceResponse, error)
	AddMember(ctx context.Context, req *AddMemberRequest) (*AddMemberResponse, error)
	RemoveMember(ctx context.Context, req *RemoveMemberRequest) (*RemoveMemberResponse, error)
}

// Request/Response types
type CreateWorkspaceRequest struct {
	Name        string
	Description string
	OwnerID     string // Extracted from context/token
}

type CreateWorkspaceResponse struct {
	Workspace *Workspace
}

type GetWorkspaceRequest struct {
	WorkspaceID string
	UserID      string
}

type GetWorkspaceResponse struct {
	Workspace *Workspace
}

type ListWorkspacesRequest struct {
	UserID   string
	Page     int32
	PageSize int32
}

type ListWorkspacesResponse struct {
	Workspaces []*Workspace
	Total      int64
	Page       int32
	PageSize   int32
}

type UpdateWorkspaceRequest struct {
	WorkspaceID string
	UserID      string
	Name        *string
	Description *string
}

type UpdateWorkspaceResponse struct {
	Workspace *Workspace
}

type DeleteWorkspaceRequest struct {
	WorkspaceID string
	UserID      string
}

type DeleteWorkspaceResponse struct {
	Success bool
}

type AddMemberRequest struct {
	WorkspaceID string
	UserID      string // User making the request
	NewUserID   string // User being added
	Role        string
}

type AddMemberResponse struct {
	Member *Member
}

type RemoveMemberRequest struct {
	WorkspaceID  string
	UserID       string // User making the request
	MemberUserID string // User being removed
}

type RemoveMemberResponse struct {
	Success bool
}

// Workspace DTO for responses
type Workspace struct {
	ID          string
	Name        string
	Description string
	OwnerID     string
	Members     []*Member
	CreatedAt   int64
	UpdatedAt   int64
}

// Member DTO for responses
type Member struct {
	UserID   string
	Role     string
	JoinedAt int64
}

// WorkspaceServer implements WorkspaceServiceServer.
type WorkspaceServer struct {
	workspaceService *service.WorkspaceService
}

// NewWorkspaceServer creates a new WorkspaceServer.
func NewWorkspaceServer(workspaceService *service.WorkspaceService) *WorkspaceServer {
	return &WorkspaceServer{workspaceService: workspaceService}
}

// CreateWorkspace creates a new workspace.
func (s *WorkspaceServer) CreateWorkspace(ctx context.Context, req *CreateWorkspaceRequest) (*CreateWorkspaceResponse, error) {
	ownerID, err := uuid.Parse(req.OwnerID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner_id")
	}

	workspace, err := s.workspaceService.CreateWorkspace(&service.CreateWorkspaceInput{
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     ownerID,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &CreateWorkspaceResponse{
		Workspace: workspaceToProto(workspace),
	}, nil
}

// GetWorkspace returns a workspace.
func (s *WorkspaceServer) GetWorkspace(ctx context.Context, req *GetWorkspaceRequest) (*GetWorkspaceResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace_id")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	workspace, err := s.workspaceService.GetWorkspace(workspaceID, userID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &GetWorkspaceResponse{
		Workspace: workspaceToProto(workspace),
	}, nil
}

// ListWorkspaces returns workspaces for a user.
func (s *WorkspaceServer) ListWorkspaces(ctx context.Context, req *ListWorkspacesRequest) (*ListWorkspacesResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	result, err := s.workspaceService.ListWorkspaces(&service.ListWorkspacesInput{
		UserID:   userID,
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	workspaces := make([]*Workspace, len(result.Workspaces))
	for i, w := range result.Workspaces {
		workspaces[i] = workspaceToProto(&w)
	}

	return &ListWorkspacesResponse{
		Workspaces: workspaces,
		Total:      result.Total,
		Page:       int32(result.Page),
		PageSize:   int32(result.PageSize),
	}, nil
}

// UpdateWorkspace updates a workspace.
func (s *WorkspaceServer) UpdateWorkspace(ctx context.Context, req *UpdateWorkspaceRequest) (*UpdateWorkspaceResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace_id")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	workspace, err := s.workspaceService.UpdateWorkspace(&service.UpdateWorkspaceInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &UpdateWorkspaceResponse{
		Workspace: workspaceToProto(workspace),
	}, nil
}

// DeleteWorkspace deletes a workspace.
func (s *WorkspaceServer) DeleteWorkspace(ctx context.Context, req *DeleteWorkspaceRequest) (*DeleteWorkspaceResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace_id")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	if err := s.workspaceService.DeleteWorkspace(workspaceID, userID); err != nil {
		return nil, toGRPCError(err)
	}

	return &DeleteWorkspaceResponse{Success: true}, nil
}

// AddMember adds a member to a workspace.
func (s *WorkspaceServer) AddMember(ctx context.Context, req *AddMemberRequest) (*AddMemberResponse, error) {
	workspaceID, err := uuid.Parse(req.WorkspaceID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid workspace_id")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	newUserID, err := uuid.Parse(req.NewUserID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid new_user_id")
	}

	member, err := s.workspaceService.AddMember(&service.AddMemberInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
		NewUserID:   newUserID,
		Role:        req.Role,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &AddMemberResponse{
		Member: memberToProto(member),
	}, nil
}

// RemoveMember removes a member from a workspace.
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

	if err := s.workspaceService.RemoveMember(workspaceID, userID, memberUserID); err != nil {
		return nil, toGRPCError(err)
	}

	return &RemoveMemberResponse{Success: true}, nil
}

// RegisterWorkspaceServiceServer registers the workspace service server.
func RegisterWorkspaceServiceServer(s *grpc.Server, srv WorkspaceServiceServer) {
	// Will be generated from proto
}

// Helper functions
func workspaceToProto(w *repository.Workspace) *Workspace {
	members := make([]*Member, len(w.Members))
	for i, m := range w.Members {
		members[i] = memberToProto(&m)
	}

	return &Workspace{
		ID:          w.ID.String(),
		Name:        w.Name,
		Description: w.Description,
		OwnerID:     w.OwnerID.String(),
		Members:     members,
		CreatedAt:   w.CreatedAt.Unix(),
		UpdatedAt:   w.UpdatedAt.Unix(),
	}
}

func memberToProto(m *repository.WorkspaceMember) *Member {
	return &Member{
		UserID:   m.UserID.String(),
		Role:     m.Role,
		JoinedAt: m.JoinedAt.Unix(),
	}
}

func toGRPCError(err error) error {
	if e, ok := err.(*errors.Error); ok {
		return e.ToGRPCError()
	}
	return status.Error(codes.Internal, err.Error())
}


// Package grpc provides gRPC handlers for Branch Service.
package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/branch-service/internal/domain/entity"
	"github.com/xiiisorate/granula_api/branch-service/internal/service"
	pb "github.com/xiiisorate/granula_api/shared/gen/branch/v1"
	apperrors "github.com/xiiisorate/granula_api/shared/pkg/errors"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// BranchServer implements the gRPC Branch Service.
type BranchServer struct {
	pb.UnimplementedBranchServiceServer
	service *service.BranchService
	log     *logger.Logger
}

// NewBranchServer creates a new BranchServer.
func NewBranchServer(svc *service.BranchService, log *logger.Logger) *BranchServer {
	return &BranchServer{service: svc, log: log}
}

// CreateBranch creates a new branch.
func (s *BranchServer) CreateBranch(ctx context.Context, req *pb.CreateBranchRequest) (*pb.CreateBranchResponse, error) {
	sceneID, err := uuid.Parse(req.SceneId)
	if err != nil {
		return nil, apperrors.InvalidArgument("scene_id", "invalid UUID").ToGRPCError()
	}

	var parentID *uuid.UUID
	if req.ParentBranchId != "" {
		pid, err := uuid.Parse(req.ParentBranchId)
		if err != nil {
			return nil, apperrors.InvalidArgument("parent_branch_id", "invalid UUID").ToGRPCError()
		}
		parentID = &pid
	}

	branch, err := s.service.CreateBranch(ctx, sceneID, req.Name, req.Description, parentID)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.CreateBranchResponse{Success: true, Branch: convertBranchToPB(branch)}, nil
}

// GetBranch retrieves a branch.
func (s *BranchServer) GetBranch(ctx context.Context, req *pb.GetBranchRequest) (*pb.GetBranchResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperrors.InvalidArgument("id", "invalid UUID").ToGRPCError()
	}

	branch, err := s.service.GetBranch(ctx, id)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.GetBranchResponse{Branch: convertBranchToPB(branch)}, nil
}

// ListBranches lists branches for a scene.
func (s *BranchServer) ListBranches(ctx context.Context, req *pb.ListBranchesRequest) (*pb.ListBranchesResponse, error) {
	sceneID, err := uuid.Parse(req.SceneId)
	if err != nil {
		return nil, apperrors.InvalidArgument("scene_id", "invalid UUID").ToGRPCError()
	}

	branches, err := s.service.ListBranches(ctx, sceneID)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	pbBranches := make([]*pb.Branch, 0, len(branches))
	for _, b := range branches {
		pbBranches = append(pbBranches, convertBranchToPB(b))
	}

	return &pb.ListBranchesResponse{Branches: pbBranches}, nil
}

// DeleteBranch deletes a branch.
func (s *BranchServer) DeleteBranch(ctx context.Context, req *pb.DeleteBranchRequest) (*pb.DeleteBranchResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperrors.InvalidArgument("id", "invalid UUID").ToGRPCError()
	}

	if err := s.service.DeleteBranch(ctx, id); err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.DeleteBranchResponse{Success: true}, nil
}

// MergeBranch merges branches.
func (s *BranchServer) MergeBranch(ctx context.Context, req *pb.MergeBranchRequest) (*pb.MergeBranchResponse, error) {
	sourceID, err := uuid.Parse(req.SourceBranchId)
	if err != nil {
		return nil, apperrors.InvalidArgument("source_branch_id", "invalid UUID").ToGRPCError()
	}

	targetID, err := uuid.Parse(req.TargetBranchId)
	if err != nil {
		return nil, apperrors.InvalidArgument("target_branch_id", "invalid UUID").ToGRPCError()
	}

	result, err := s.service.MergeBranch(ctx, sourceID, targetID, req.DeleteSource)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.MergeBranchResponse{
		Success:       result.Success,
		ChangesMerged: int32(result.ChangesMerged),
		Conflicts:     result.Conflicts,
	}, nil
}

// GetDiff gets diff between branches.
func (s *BranchServer) GetDiff(ctx context.Context, req *pb.GetDiffRequest) (*pb.GetDiffResponse, error) {
	sourceID, err := uuid.Parse(req.SourceBranchId)
	if err != nil {
		return nil, apperrors.InvalidArgument("source_branch_id", "invalid UUID").ToGRPCError()
	}

	targetID, err := uuid.Parse(req.TargetBranchId)
	if err != nil {
		return nil, apperrors.InvalidArgument("target_branch_id", "invalid UUID").ToGRPCError()
	}

	result, err := s.service.GetDiff(ctx, sourceID, targetID)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.GetDiffResponse{
		Diff: &pb.BranchDiff{TotalChanges: int32(result.TotalChanges)},
	}, nil
}

// CreateSnapshot creates a snapshot.
func (s *BranchServer) CreateSnapshot(ctx context.Context, req *pb.CreateSnapshotRequest) (*pb.CreateSnapshotResponse, error) {
	branchID, err := uuid.Parse(req.BranchId)
	if err != nil {
		return nil, apperrors.InvalidArgument("branch_id", "invalid UUID").ToGRPCError()
	}

	snapshot, err := s.service.CreateSnapshot(ctx, branchID, req.Name)
	if err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.CreateSnapshotResponse{
		Success: true,
		Snapshot: &pb.Snapshot{
			Id:        snapshot.ID.String(),
			BranchId:  snapshot.BranchID.String(),
			Name:      snapshot.Name,
			Version:   int32(snapshot.Version),
			CreatedAt: timestamppb.New(snapshot.CreatedAt),
		},
	}, nil
}

// RestoreSnapshot restores a snapshot.
func (s *BranchServer) RestoreSnapshot(ctx context.Context, req *pb.RestoreSnapshotRequest) (*pb.RestoreSnapshotResponse, error) {
	snapshotID, err := uuid.Parse(req.SnapshotId)
	if err != nil {
		return nil, apperrors.InvalidArgument("snapshot_id", "invalid UUID").ToGRPCError()
	}

	if err := s.service.RestoreSnapshot(ctx, snapshotID); err != nil {
		return nil, apperrors.FromGRPCError(err).ToGRPCError()
	}

	return &pb.RestoreSnapshotResponse{Success: true}, nil
}

func convertBranchToPB(b *entity.Branch) *pb.Branch {
	pbBranch := &pb.Branch{
		Id:          b.ID.String(),
		SceneId:     b.SceneID.String(),
		Name:        b.Name,
		Description: b.Description,
		IsMain:      b.IsMain,
		Status:      convertStatusToPB(b.Status),
		CreatedAt:   timestamppb.New(b.CreatedAt),
		UpdatedAt:   timestamppb.New(b.UpdatedAt),
	}
	if b.ParentBranchID != nil {
		pbBranch.ParentBranchId = b.ParentBranchID.String()
	}
	return pbBranch
}

func convertStatusToPB(s entity.BranchStatus) pb.BranchStatus {
	switch s {
	case entity.BranchStatusActive:
		return pb.BranchStatus_BRANCH_STATUS_ACTIVE
	case entity.BranchStatusMerged:
		return pb.BranchStatus_BRANCH_STATUS_MERGED
	case entity.BranchStatusArchived:
		return pb.BranchStatus_BRANCH_STATUS_ARCHIVED
	default:
		return pb.BranchStatus_BRANCH_STATUS_UNSPECIFIED
	}
}


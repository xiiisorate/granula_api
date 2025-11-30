// =============================================================================
// Package service provides business logic for Branch Service.
// =============================================================================
// This package implements the core business operations for branch management,
// including branch creation from parent branches, element copying, and merging.
//
// Integrates with Scene Service for element operations.
// =============================================================================
package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/branch-service/internal/domain/entity"
	"github.com/xiiisorate/granula_api/branch-service/internal/repository/mongodb"
	apperrors "github.com/xiiisorate/granula_api/shared/pkg/errors"
	"github.com/xiiisorate/granula_api/shared/pkg/logger"
)

// =============================================================================
// Interfaces
// =============================================================================

// SceneElementsProvider is an interface for managing scene elements.
// This allows for easier testing and decoupling from the scene client implementation.
type SceneElementsProvider interface {
	CopyElementsToBranch(ctx context.Context, sceneID, sourceBranchID, targetBranchID uuid.UUID) (int, error)
	DeleteBranchElements(ctx context.Context, branchID uuid.UUID) (int, error)
	CheckSceneCompliance(ctx context.Context, sceneID, branchID uuid.UUID) (bool, []string, error)
}

// =============================================================================
// BranchService
// =============================================================================

// BranchService handles branch operations.
// It integrates with Scene Service for element copying and compliance checks.
//
// Thread Safety: Safe for concurrent use.
type BranchService struct {
	repo        *mongodb.BranchRepository
	sceneClient SceneElementsProvider // Optional: nil if scene service is unavailable
	log         *logger.Logger
}

// NewBranchService creates a new BranchService.
//
// Parameters:
//   - repo: MongoDB repository for branch persistence
//   - sceneClient: Scene Service client for element operations (can be nil)
//   - log: Logger instance for operational logging
//
// Returns:
//   - *BranchService: Initialized service ready for use
//
// Note: If sceneClient is nil, element copying will be skipped.
func NewBranchService(repo *mongodb.BranchRepository, sceneClient SceneElementsProvider, log *logger.Logger) *BranchService {
	return &BranchService{
		repo:        repo,
		sceneClient: sceneClient,
		log:         log,
	}
}

// CreateBranch creates a new branch.
// If parentID is provided, copies all elements from the parent branch.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - sceneID: UUID of the scene containing this branch
//   - name: Display name for the branch
//   - description: Optional description of branch purpose
//   - parentID: Optional UUID of parent branch to copy elements from
//
// Returns:
//   - *entity.Branch: Created branch entity
//   - error: Validation or database error
func (s *BranchService) CreateBranch(ctx context.Context, sceneID uuid.UUID, name, description string, parentID *uuid.UUID) (*entity.Branch, error) {
	s.log.Info("creating branch",
		logger.String("scene_id", sceneID.String()),
		logger.String("name", name),
	)

	branch := entity.NewBranch(sceneID, name, parentID)
	branch.Description = description

	if err := s.repo.Create(ctx, branch); err != nil {
		return nil, err
	}

	// Copy elements from parent branch if parentID is set
	if parentID != nil && s.sceneClient != nil {
		s.log.Info("copying elements from parent branch",
			logger.String("parent_id", parentID.String()),
			logger.String("new_branch_id", branch.ID.String()),
		)

		// Use goroutine to not block branch creation on element copy
		go func() {
			// Use background context as the request context may be cancelled
			ctx := context.Background()
			copiedCount, err := s.sceneClient.CopyElementsToBranch(ctx, sceneID, *parentID, branch.ID)
			if err != nil {
				s.log.Warn("failed to copy elements from parent branch",
					logger.Err(err),
					logger.String("parent_id", parentID.String()),
					logger.String("new_branch_id", branch.ID.String()),
				)
			} else {
				s.log.Info("elements copied from parent branch",
					logger.String("parent_id", parentID.String()),
					logger.String("new_branch_id", branch.ID.String()),
					logger.Int("copied_count", copiedCount),
				)
			}
		}()
	}

	return branch, nil
}

// GetBranch retrieves a branch by ID.
func (s *BranchService) GetBranch(ctx context.Context, id uuid.UUID) (*entity.Branch, error) {
	return s.repo.GetByID(ctx, id)
}

// ListBranches lists branches for a scene.
func (s *BranchService) ListBranches(ctx context.Context, sceneID uuid.UUID) ([]*entity.Branch, error) {
	return s.repo.ListByScene(ctx, sceneID)
}

// DeleteBranch deletes a branch.
func (s *BranchService) DeleteBranch(ctx context.Context, id uuid.UUID) error {
	branch, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if branch.IsMain {
		return apperrors.InvalidArgument("branch", "cannot delete main branch")
	}
	return s.repo.Delete(ctx, id)
}

// MergeBranch merges source branch into target.
func (s *BranchService) MergeBranch(ctx context.Context, sourceID, targetID uuid.UUID, deleteSource bool) (*MergeResult, error) {
	s.log.Info("merging branches",
		logger.String("source", sourceID.String()),
		logger.String("target", targetID.String()),
	)

	source, err := s.repo.GetByID(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	// TODO: Implement actual merge logic with conflict detection

	if deleteSource {
		source.Status = entity.BranchStatusMerged
		_ = s.repo.Update(ctx, source)
	}

	return &MergeResult{Success: true, ChangesMerged: 0}, nil
}

// GetDiff gets differences between two branches.
func (s *BranchService) GetDiff(ctx context.Context, sourceID, targetID uuid.UUID) (*DiffResult, error) {
	// TODO: Implement diff logic
	return &DiffResult{TotalChanges: 0}, nil
}

// CreateSnapshot creates a snapshot of a branch.
func (s *BranchService) CreateSnapshot(ctx context.Context, branchID uuid.UUID, name string) (*entity.Snapshot, error) {
	branch, err := s.repo.GetByID(ctx, branchID)
	if err != nil {
		return nil, err
	}

	// TODO: Serialize current elements
	snapshot := entity.NewSnapshot(branchID, name, branch.Version, nil)

	if err := s.repo.CreateSnapshot(ctx, snapshot); err != nil {
		return nil, err
	}

	return snapshot, nil
}

// RestoreSnapshot restores a branch to a snapshot.
func (s *BranchService) RestoreSnapshot(ctx context.Context, snapshotID uuid.UUID) error {
	snapshot, err := s.repo.GetSnapshot(ctx, snapshotID)
	if err != nil {
		return err
	}

	// TODO: Restore elements from snapshot data
	_ = snapshot

	return nil
}

// MergeResult from branch merge.
type MergeResult struct {
	Success       bool
	ChangesMerged int
	Conflicts     []string
}

// DiffResult between branches.
type DiffResult struct {
	Added        []ElementChange
	Modified     []ElementChange
	Deleted      []ElementChange
	TotalChanges int
}

// ElementChange represents a change to an element.
type ElementChange struct {
	ElementID   string
	ElementType string
	Description string
}


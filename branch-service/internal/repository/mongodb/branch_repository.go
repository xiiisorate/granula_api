// Package mongodb provides MongoDB implementations of repositories.
package mongodb

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/branch-service/internal/domain/entity"
	apperrors "github.com/xiiisorate/granula_api/shared/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BranchRepository handles branch persistence.
type BranchRepository struct {
	branches  *mongo.Collection
	snapshots *mongo.Collection
}

// NewBranchRepository creates a new BranchRepository.
func NewBranchRepository(db *mongo.Database) *BranchRepository {
	branches := db.Collection("branches")
	snapshots := db.Collection("snapshots")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, _ = branches.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "scene_id", Value: 1}}},
		{Keys: bson.D{{Key: "scene_id", Value: 1}, {Key: "is_main", Value: 1}}},
	})
	_, _ = snapshots.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "branch_id", Value: 1}, {Key: "created_at", Value: -1}},
	})

	return &BranchRepository{branches: branches, snapshots: snapshots}
}

// Create creates a new branch.
func (r *BranchRepository) Create(ctx context.Context, branch *entity.Branch) error {
	_, err := r.branches.InsertOne(ctx, branch)
	if err != nil {
		return apperrors.Internal("failed to create branch").WithCause(err)
	}
	return nil
}

// GetByID retrieves a branch by ID.
func (r *BranchRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Branch, error) {
	var branch entity.Branch
	err := r.branches.FindOne(ctx, bson.M{"_id": id}).Decode(&branch)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, apperrors.NotFound("branch", id.String())
		}
		return nil, apperrors.Internal("failed to get branch").WithCause(err)
	}
	return &branch, nil
}

// Update updates a branch.
func (r *BranchRepository) Update(ctx context.Context, branch *entity.Branch) error {
	branch.UpdatedAt = time.Now().UTC()
	result, err := r.branches.ReplaceOne(ctx, bson.M{"_id": branch.ID}, branch)
	if err != nil {
		return apperrors.Internal("failed to update branch").WithCause(err)
	}
	if result.MatchedCount == 0 {
		return apperrors.NotFound("branch", branch.ID.String())
	}
	return nil
}

// Delete deletes a branch.
func (r *BranchRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.branches.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return apperrors.Internal("failed to delete branch").WithCause(err)
	}
	if result.DeletedCount == 0 {
		return apperrors.NotFound("branch", id.String())
	}
	return nil
}

// ListByScene lists branches for a scene.
func (r *BranchRepository) ListByScene(ctx context.Context, sceneID uuid.UUID) ([]*entity.Branch, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.branches.Find(ctx, bson.M{"scene_id": sceneID}, opts)
	if err != nil {
		return nil, apperrors.Internal("failed to list branches").WithCause(err)
	}
	defer cursor.Close(ctx)

	var branches []*entity.Branch
	if err := cursor.All(ctx, &branches); err != nil {
		return nil, apperrors.Internal("failed to decode branches").WithCause(err)
	}
	return branches, nil
}

// GetMainBranch gets the main branch for a scene.
func (r *BranchRepository) GetMainBranch(ctx context.Context, sceneID uuid.UUID) (*entity.Branch, error) {
	var branch entity.Branch
	err := r.branches.FindOne(ctx, bson.M{"scene_id": sceneID, "is_main": true}).Decode(&branch)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, apperrors.NotFoundMsg("main branch not found")
		}
		return nil, apperrors.Internal("failed to get main branch").WithCause(err)
	}
	return &branch, nil
}

// CreateSnapshot creates a snapshot.
func (r *BranchRepository) CreateSnapshot(ctx context.Context, snapshot *entity.Snapshot) error {
	_, err := r.snapshots.InsertOne(ctx, snapshot)
	if err != nil {
		return apperrors.Internal("failed to create snapshot").WithCause(err)
	}
	return nil
}

// GetSnapshot retrieves a snapshot by ID.
func (r *BranchRepository) GetSnapshot(ctx context.Context, id uuid.UUID) (*entity.Snapshot, error) {
	var snapshot entity.Snapshot
	err := r.snapshots.FindOne(ctx, bson.M{"_id": id}).Decode(&snapshot)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, apperrors.NotFound("snapshot", id.String())
		}
		return nil, apperrors.Internal("failed to get snapshot").WithCause(err)
	}
	return &snapshot, nil
}

// ListSnapshots lists snapshots for a branch.
func (r *BranchRepository) ListSnapshots(ctx context.Context, branchID uuid.UUID) ([]*entity.Snapshot, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.snapshots.Find(ctx, bson.M{"branch_id": branchID}, opts)
	if err != nil {
		return nil, apperrors.Internal("failed to list snapshots").WithCause(err)
	}
	defer cursor.Close(ctx)

	var snapshots []*entity.Snapshot
	if err := cursor.All(ctx, &snapshots); err != nil {
		return nil, apperrors.Internal("failed to decode snapshots").WithCause(err)
	}
	return snapshots, nil
}


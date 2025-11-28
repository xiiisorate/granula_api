// Package mongodb provides MongoDB implementations of repositories.
package mongodb

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/scene-service/internal/domain/entity"
	apperrors "github.com/xiiisorate/granula_api/shared/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ElementRepository handles element persistence.
type ElementRepository struct {
	collection *mongo.Collection
}

// NewElementRepository creates a new ElementRepository.
func NewElementRepository(db *mongo.Database) *ElementRepository {
	collection := db.Collection("elements")

	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "scene_id", Value: 1}, {Key: "branch_id", Value: 1}}},
		{Keys: bson.D{{Key: "scene_id", Value: 1}, {Key: "type", Value: 1}}},
		{Keys: bson.D{{Key: "branch_id", Value: 1}, {Key: "is_deleted", Value: 1}}},
		{Keys: bson.D{{Key: "parent_id", Value: 1}}},
	}
	_, _ = collection.Indexes().CreateMany(ctx, indexes)

	return &ElementRepository{collection: collection}
}

// Create creates a new element.
func (r *ElementRepository) Create(ctx context.Context, element *entity.Element) error {
	_, err := r.collection.InsertOne(ctx, element)
	if err != nil {
		return apperrors.Internal("failed to create element").WithCause(err)
	}
	return nil
}

// CreateMany creates multiple elements.
func (r *ElementRepository) CreateMany(ctx context.Context, elements []*entity.Element) error {
	if len(elements) == 0 {
		return nil
	}

	docs := make([]interface{}, len(elements))
	for i, e := range elements {
		docs[i] = e
	}

	_, err := r.collection.InsertMany(ctx, docs)
	if err != nil {
		return apperrors.Internal("failed to create elements").WithCause(err)
	}
	return nil
}

// GetByID retrieves an element by ID.
func (r *ElementRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Element, error) {
	var element entity.Element
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&element)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, apperrors.NotFound("element", id.String())
		}
		return nil, apperrors.Internal("failed to get element").WithCause(err)
	}
	return &element, nil
}

// Update updates an element with optimistic locking.
func (r *ElementRepository) Update(ctx context.Context, element *entity.Element) error {
	oldVersion := element.Version
	element.Update()

	result, err := r.collection.ReplaceOne(ctx,
		bson.M{"_id": element.ID, "version": oldVersion},
		element,
	)
	if err != nil {
		return apperrors.Internal("failed to update element").WithCause(err)
	}
	if result.MatchedCount == 0 {
		return apperrors.Conflict("element was modified by another process")
	}
	return nil
}

// Delete soft-deletes an element.
func (r *ElementRepository) Delete(ctx context.Context, id uuid.UUID) error {
	element, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	element.SoftDelete()
	return r.Update(ctx, element)
}

// HardDelete permanently deletes an element.
func (r *ElementRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return apperrors.Internal("failed to delete element").WithCause(err)
	}
	if result.DeletedCount == 0 {
		return apperrors.NotFound("element", id.String())
	}
	return nil
}

// ListByBranch lists active elements in a branch.
func (r *ElementRepository) ListByBranch(ctx context.Context, branchID uuid.UUID, opts ListOptions) ([]*entity.Element, error) {
	filter := bson.M{
		"branch_id":  branchID,
		"is_deleted": false,
	}

	if opts.Type != "" {
		filter["type"] = opts.Type
	}

	findOpts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})
	if opts.Limit > 0 {
		findOpts.SetLimit(int64(opts.Limit))
	}
	if opts.Offset > 0 {
		findOpts.SetSkip(int64(opts.Offset))
	}

	cursor, err := r.collection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, apperrors.Internal("failed to list elements").WithCause(err)
	}
	defer cursor.Close(ctx)

	var elements []*entity.Element
	if err := cursor.All(ctx, &elements); err != nil {
		return nil, apperrors.Internal("failed to decode elements").WithCause(err)
	}

	return elements, nil
}

// ListByScene lists all elements in a scene (across all branches).
func (r *ElementRepository) ListByScene(ctx context.Context, sceneID uuid.UUID) ([]*entity.Element, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"scene_id": sceneID})
	if err != nil {
		return nil, apperrors.Internal("failed to list elements").WithCause(err)
	}
	defer cursor.Close(ctx)

	var elements []*entity.Element
	if err := cursor.All(ctx, &elements); err != nil {
		return nil, apperrors.Internal("failed to decode elements").WithCause(err)
	}

	return elements, nil
}

// GetWalls retrieves all walls in a branch.
func (r *ElementRepository) GetWalls(ctx context.Context, branchID uuid.UUID) ([]*entity.Element, error) {
	return r.ListByBranch(ctx, branchID, ListOptions{Type: string(entity.ElementTypeWall)})
}

// GetRooms retrieves all rooms in a branch.
func (r *ElementRepository) GetRooms(ctx context.Context, branchID uuid.UUID) ([]*entity.Element, error) {
	return r.ListByBranch(ctx, branchID, ListOptions{Type: string(entity.ElementTypeRoom)})
}

// CountByBranch counts elements in a branch.
func (r *ElementRepository) CountByBranch(ctx context.Context, branchID uuid.UUID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{
		"branch_id":  branchID,
		"is_deleted": false,
	})
}

// CopyToBranch copies elements from one branch to another.
func (r *ElementRepository) CopyToBranch(ctx context.Context, sourceBranchID, targetBranchID uuid.UUID) error {
	// Get all elements from source branch
	elements, err := r.ListByBranch(ctx, sourceBranchID, ListOptions{})
	if err != nil {
		return err
	}

	// Create copies for target branch
	for _, e := range elements {
		copy := *e
		copy.ID = uuid.New()
		copy.BranchID = targetBranchID
		copy.Version = 1
		copy.CreatedAt = time.Now().UTC()
		copy.UpdatedAt = copy.CreatedAt

		if err := r.Create(ctx, &copy); err != nil {
			return err
		}
	}

	return nil
}

// ListOptions for filtering elements.
type ListOptions struct {
	Type   string
	Limit  int
	Offset int
}


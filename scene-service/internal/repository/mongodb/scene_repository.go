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

// SceneRepository handles scene persistence.
type SceneRepository struct {
	collection *mongo.Collection
}

// NewSceneRepository creates a new SceneRepository.
func NewSceneRepository(db *mongo.Database) *SceneRepository {
	collection := db.Collection("scenes")

	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "workspace_id", Value: 1}}},
		{Keys: bson.D{{Key: "owner_id", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	}
	_, _ = collection.Indexes().CreateMany(ctx, indexes)

	return &SceneRepository{collection: collection}
}

// Create creates a new scene.
func (r *SceneRepository) Create(ctx context.Context, scene *entity.Scene) error {
	_, err := r.collection.InsertOne(ctx, scene)
	if err != nil {
		return apperrors.Internal("failed to create scene").WithCause(err)
	}
	return nil
}

// GetByID retrieves a scene by ID.
func (r *SceneRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Scene, error) {
	var scene entity.Scene
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&scene)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, apperrors.NotFound("scene", id.String())
		}
		return nil, apperrors.Internal("failed to get scene").WithCause(err)
	}
	return &scene, nil
}

// Update updates a scene.
func (r *SceneRepository) Update(ctx context.Context, scene *entity.Scene) error {
	scene.UpdatedAt = time.Now().UTC()

	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": scene.ID}, scene)
	if err != nil {
		return apperrors.Internal("failed to update scene").WithCause(err)
	}
	if result.MatchedCount == 0 {
		return apperrors.NotFound("scene", scene.ID.String())
	}
	return nil
}

// Delete deletes a scene.
func (r *SceneRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return apperrors.Internal("failed to delete scene").WithCause(err)
	}
	if result.DeletedCount == 0 {
		return apperrors.NotFound("scene", id.String())
	}
	return nil
}

// ListByWorkspace lists scenes in a workspace.
func (r *SceneRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*entity.Scene, int64, error) {
	filter := bson.M{"workspace_id": workspaceID}

	// Count total
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, apperrors.Internal("failed to count scenes").WithCause(err)
	}

	// Find with pagination
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, apperrors.Internal("failed to list scenes").WithCause(err)
	}
	defer cursor.Close(ctx)

	var scenes []*entity.Scene
	if err := cursor.All(ctx, &scenes); err != nil {
		return nil, 0, apperrors.Internal("failed to decode scenes").WithCause(err)
	}

	return scenes, total, nil
}


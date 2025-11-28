// Package mongodb provides MongoDB implementations of repositories.
package mongodb

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/ai-service/internal/domain/entity"
	apperrors "github.com/xiiisorate/granula_api/shared/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// JobRepository handles job persistence for recognition and generation.
type JobRepository struct {
	recognitionCollection *mongo.Collection
	generationCollection  *mongo.Collection
}

// NewJobRepository creates a new JobRepository.
func NewJobRepository(db *mongo.Database) *JobRepository {
	recognitionCollection := db.Collection("recognition_jobs")
	generationCollection := db.Collection("generation_jobs")

	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Recognition indexes
	recIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "floor_plan_id", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	}
	_, _ = recognitionCollection.Indexes().CreateMany(ctx, recIndexes)

	// Generation indexes
	genIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "scene_id", Value: 1}, {Key: "branch_id", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	}
	_, _ = generationCollection.Indexes().CreateMany(ctx, genIndexes)

	return &JobRepository{
		recognitionCollection: recognitionCollection,
		generationCollection:  generationCollection,
	}
}

// =============================================================================
// Recognition Jobs
// =============================================================================

// SaveRecognitionJob saves a recognition job.
func (r *JobRepository) SaveRecognitionJob(ctx context.Context, job *entity.RecognitionJob) error {
	_, err := r.recognitionCollection.InsertOne(ctx, job)
	if err != nil {
		return apperrors.Internal("failed to save recognition job").WithCause(err)
	}
	return nil
}

// GetRecognitionJob retrieves a recognition job by ID.
func (r *JobRepository) GetRecognitionJob(ctx context.Context, id uuid.UUID) (*entity.RecognitionJob, error) {
	var job entity.RecognitionJob
	err := r.recognitionCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&job)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, apperrors.NotFound("recognition_job", id.String())
		}
		return nil, apperrors.Internal("failed to get recognition job").WithCause(err)
	}
	return &job, nil
}

// UpdateRecognitionJob updates a recognition job.
func (r *JobRepository) UpdateRecognitionJob(ctx context.Context, job *entity.RecognitionJob) error {
	job.UpdatedAt = time.Now().UTC()

	result, err := r.recognitionCollection.ReplaceOne(ctx, bson.M{"_id": job.ID}, job)
	if err != nil {
		return apperrors.Internal("failed to update recognition job").WithCause(err)
	}

	if result.MatchedCount == 0 {
		return apperrors.NotFound("recognition_job", job.ID.String())
	}

	return nil
}

// GetRecognitionJobByFloorPlan gets the latest job for a floor plan.
func (r *JobRepository) GetRecognitionJobByFloorPlan(ctx context.Context, floorPlanID string) (*entity.RecognitionJob, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})

	var job entity.RecognitionJob
	err := r.recognitionCollection.FindOne(ctx, bson.M{"floor_plan_id": floorPlanID}, opts).Decode(&job)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, apperrors.NotFoundMsg("no recognition job found for floor plan")
		}
		return nil, apperrors.Internal("failed to get recognition job").WithCause(err)
	}
	return &job, nil
}

// ListPendingRecognitionJobs lists pending recognition jobs.
func (r *JobRepository) ListPendingRecognitionJobs(ctx context.Context, limit int) ([]*entity.RecognitionJob, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: 1}}).
		SetLimit(int64(limit))

	cursor, err := r.recognitionCollection.Find(ctx, bson.M{"status": entity.JobStatusPending}, opts)
	if err != nil {
		return nil, apperrors.Internal("failed to list pending jobs").WithCause(err)
	}
	defer cursor.Close(ctx)

	var jobs []*entity.RecognitionJob
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, apperrors.Internal("failed to decode jobs").WithCause(err)
	}

	return jobs, nil
}

// =============================================================================
// Generation Jobs
// =============================================================================

// SaveGenerationJob saves a generation job.
func (r *JobRepository) SaveGenerationJob(ctx context.Context, job *entity.GenerationJob) error {
	_, err := r.generationCollection.InsertOne(ctx, job)
	if err != nil {
		return apperrors.Internal("failed to save generation job").WithCause(err)
	}
	return nil
}

// GetGenerationJob retrieves a generation job by ID.
func (r *JobRepository) GetGenerationJob(ctx context.Context, id uuid.UUID) (*entity.GenerationJob, error) {
	var job entity.GenerationJob
	err := r.generationCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&job)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, apperrors.NotFound("generation_job", id.String())
		}
		return nil, apperrors.Internal("failed to get generation job").WithCause(err)
	}
	return &job, nil
}

// UpdateGenerationJob updates a generation job.
func (r *JobRepository) UpdateGenerationJob(ctx context.Context, job *entity.GenerationJob) error {
	job.UpdatedAt = time.Now().UTC()

	result, err := r.generationCollection.ReplaceOne(ctx, bson.M{"_id": job.ID}, job)
	if err != nil {
		return apperrors.Internal("failed to update generation job").WithCause(err)
	}

	if result.MatchedCount == 0 {
		return apperrors.NotFound("generation_job", job.ID.String())
	}

	return nil
}

// ListPendingGenerationJobs lists pending generation jobs.
func (r *JobRepository) ListPendingGenerationJobs(ctx context.Context, limit int) ([]*entity.GenerationJob, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: 1}}).
		SetLimit(int64(limit))

	cursor, err := r.generationCollection.Find(ctx, bson.M{"status": entity.JobStatusPending}, opts)
	if err != nil {
		return nil, apperrors.Internal("failed to list pending jobs").WithCause(err)
	}
	defer cursor.Close(ctx)

	var jobs []*entity.GenerationJob
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, apperrors.Internal("failed to decode jobs").WithCause(err)
	}

	return jobs, nil
}

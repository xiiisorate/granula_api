// Package postgres provides PostgreSQL implementations of repositories.
package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xiiisorate/granula_api/floorplan-service/internal/domain/entity"
	apperrors "github.com/xiiisorate/granula_api/shared/pkg/errors"
)

// FloorPlanRepository handles floor plan persistence.
type FloorPlanRepository struct {
	pool *pgxpool.Pool
}

// NewFloorPlanRepository creates a new FloorPlanRepository.
func NewFloorPlanRepository(pool *pgxpool.Pool) *FloorPlanRepository {
	return &FloorPlanRepository{pool: pool}
}

// Create creates a new floor plan.
func (r *FloorPlanRepository) Create(ctx context.Context, fp *entity.FloorPlan) error {
	query := `
		INSERT INTO floor_plans (id, workspace_id, owner_id, name, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		fp.ID,
		fp.WorkspaceID,
		fp.OwnerID,
		fp.Name,
		fp.Description,
		fp.Status,
		fp.CreatedAt,
		fp.UpdatedAt,
	)
	if err != nil {
		return apperrors.Internal("failed to create floor plan").WithCause(err)
	}

	return nil
}

// GetByID retrieves a floor plan by ID.
func (r *FloorPlanRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.FloorPlan, error) {
	query := `
		SELECT id, workspace_id, owner_id, name, description, status, 
		       recognition_job_id, scene_id, created_at, updated_at
		FROM floor_plans
		WHERE id = $1
	`

	fp := &entity.FloorPlan{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&fp.ID,
		&fp.WorkspaceID,
		&fp.OwnerID,
		&fp.Name,
		&fp.Description,
		&fp.Status,
		&fp.RecognitionJobID,
		&fp.SceneID,
		&fp.CreatedAt,
		&fp.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.NotFound("floor_plan", id.String())
		}
		return nil, apperrors.Internal("failed to get floor plan").WithCause(err)
	}

	// Load file info
	fileInfo, err := r.GetFileInfo(ctx, fp.ID)
	if err == nil {
		fp.FileInfo = fileInfo
	}

	return fp, nil
}

// Update updates a floor plan.
func (r *FloorPlanRepository) Update(ctx context.Context, fp *entity.FloorPlan) error {
	fp.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE floor_plans
		SET name = $2, description = $3, status = $4, 
		    recognition_job_id = $5, scene_id = $6, updated_at = $7
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		fp.ID,
		fp.Name,
		fp.Description,
		fp.Status,
		fp.RecognitionJobID,
		fp.SceneID,
		fp.UpdatedAt,
	)
	if err != nil {
		return apperrors.Internal("failed to update floor plan").WithCause(err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("floor_plan", fp.ID.String())
	}

	return nil
}

// Delete soft deletes a floor plan.
func (r *FloorPlanRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM floor_plans WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return apperrors.Internal("failed to delete floor plan").WithCause(err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("floor_plan", id.String())
	}

	return nil
}

// ListByWorkspace lists floor plans in a workspace.
func (r *FloorPlanRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, opts ListOptions) ([]*entity.FloorPlan, int64, error) {
	countQuery := `SELECT COUNT(*) FROM floor_plans WHERE workspace_id = $1`
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, workspaceID).Scan(&total); err != nil {
		return nil, 0, apperrors.Internal("failed to count floor plans").WithCause(err)
	}

	query := `
		SELECT id, workspace_id, owner_id, name, description, status, 
		       recognition_job_id, scene_id, created_at, updated_at
		FROM floor_plans
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, workspaceID, opts.Limit, opts.Offset)
	if err != nil {
		return nil, 0, apperrors.Internal("failed to list floor plans").WithCause(err)
	}
	defer rows.Close()

	floorPlans := make([]*entity.FloorPlan, 0)
	for rows.Next() {
		fp := &entity.FloorPlan{}
		if err := rows.Scan(
			&fp.ID,
			&fp.WorkspaceID,
			&fp.OwnerID,
			&fp.Name,
			&fp.Description,
			&fp.Status,
			&fp.RecognitionJobID,
			&fp.SceneID,
			&fp.CreatedAt,
			&fp.UpdatedAt,
		); err != nil {
			return nil, 0, apperrors.Internal("failed to scan floor plan").WithCause(err)
		}
		floorPlans = append(floorPlans, fp)
	}

	return floorPlans, total, nil
}

// CreateFileInfo creates file info for a floor plan.
func (r *FloorPlanRepository) CreateFileInfo(ctx context.Context, fi *entity.FileInfo) error {
	query := `
		INSERT INTO floor_plan_files (id, floor_plan_id, original_name, storage_path, mime_type, size, checksum, width, height, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.pool.Exec(ctx, query,
		fi.ID,
		fi.FloorPlanID,
		fi.OriginalName,
		fi.StoragePath,
		fi.MimeType,
		fi.Size,
		fi.Checksum,
		fi.Width,
		fi.Height,
		fi.CreatedAt,
	)
	if err != nil {
		return apperrors.Internal("failed to create file info").WithCause(err)
	}

	return nil
}

// GetFileInfo retrieves file info for a floor plan.
func (r *FloorPlanRepository) GetFileInfo(ctx context.Context, floorPlanID uuid.UUID) (*entity.FileInfo, error) {
	query := `
		SELECT id, floor_plan_id, original_name, storage_path, mime_type, size, checksum, width, height, created_at
		FROM floor_plan_files
		WHERE floor_plan_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	fi := &entity.FileInfo{}
	err := r.pool.QueryRow(ctx, query, floorPlanID).Scan(
		&fi.ID,
		&fi.FloorPlanID,
		&fi.OriginalName,
		&fi.StoragePath,
		&fi.MimeType,
		&fi.Size,
		&fi.Checksum,
		&fi.Width,
		&fi.Height,
		&fi.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.NotFound("file_info", floorPlanID.String())
		}
		return nil, apperrors.Internal("failed to get file info").WithCause(err)
	}

	return fi, nil
}

// CreateThumbnail creates a thumbnail record.
func (r *FloorPlanRepository) CreateThumbnail(ctx context.Context, th *entity.Thumbnail) error {
	query := `
		INSERT INTO floor_plan_thumbnails (id, floor_plan_id, size, storage_path, width, height, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.pool.Exec(ctx, query,
		th.ID,
		th.FloorPlanID,
		th.Size,
		th.StoragePath,
		th.Width,
		th.Height,
		th.CreatedAt,
	)
	if err != nil {
		return apperrors.Internal("failed to create thumbnail").WithCause(err)
	}

	return nil
}

// GetThumbnails retrieves thumbnails for a floor plan.
func (r *FloorPlanRepository) GetThumbnails(ctx context.Context, floorPlanID uuid.UUID) ([]*entity.Thumbnail, error) {
	query := `
		SELECT id, floor_plan_id, size, storage_path, width, height, created_at
		FROM floor_plan_thumbnails
		WHERE floor_plan_id = $1
	`

	rows, err := r.pool.Query(ctx, query, floorPlanID)
	if err != nil {
		return nil, apperrors.Internal("failed to get thumbnails").WithCause(err)
	}
	defer rows.Close()

	thumbnails := make([]*entity.Thumbnail, 0)
	for rows.Next() {
		th := &entity.Thumbnail{}
		if err := rows.Scan(
			&th.ID,
			&th.FloorPlanID,
			&th.Size,
			&th.StoragePath,
			&th.Width,
			&th.Height,
			&th.CreatedAt,
		); err != nil {
			return nil, apperrors.Internal("failed to scan thumbnail").WithCause(err)
		}
		thumbnails = append(thumbnails, th)
	}

	return thumbnails, nil
}

// ListOptions for pagination.
type ListOptions struct {
	Limit  int
	Offset int
}

// Package entity defines domain entities for AI Service.
package entity

import (
	"time"

	"github.com/google/uuid"
)

// GenerationJob represents a layout generation job.
type GenerationJob struct {
	// ID is the job identifier.
	ID uuid.UUID `json:"id" bson:"_id"`

	// SceneID is the source scene.
	SceneID string `json:"scene_id" bson:"scene_id"`

	// BranchID is the source branch.
	BranchID string `json:"branch_id" bson:"branch_id"`

	// Prompt from the user.
	Prompt string `json:"prompt" bson:"prompt"`

	// VariantsCount is how many variants to generate.
	VariantsCount int `json:"variants_count" bson:"variants_count"`

	// Options for generation.
	Options GenerationOptions `json:"options" bson:"options"`

	// Status of the job.
	Status JobStatus `json:"status" bson:"status"`

	// Progress percentage (0-100).
	Progress int `json:"progress" bson:"progress"`

	// Variants generated (if completed).
	Variants []GeneratedVariant `json:"variants,omitempty" bson:"variants,omitempty"`

	// Error message (if failed).
	Error string `json:"error,omitempty" bson:"error,omitempty"`

	// CreatedAt timestamp.
	CreatedAt time.Time `json:"created_at" bson:"created_at"`

	// UpdatedAt timestamp.
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`

	// CompletedAt timestamp.
	CompletedAt *time.Time `json:"completed_at,omitempty" bson:"completed_at,omitempty"`
}

// GenerationOptions are options for layout generation.
type GenerationOptions struct {
	// PreserveLoadBearing prevents changes to load-bearing walls.
	PreserveLoadBearing bool `json:"preserve_load_bearing" bson:"preserve_load_bearing"`

	// CheckCompliance validates against building codes.
	CheckCompliance bool `json:"check_compliance" bson:"check_compliance"`

	// PreserveWetZones keeps wet zones in place.
	PreserveWetZones bool `json:"preserve_wet_zones" bson:"preserve_wet_zones"`

	// RequiredRooms that must be in the layout.
	RequiredRooms []string `json:"required_rooms,omitempty" bson:"required_rooms,omitempty"`

	// MinRoomAreas maps room type to minimum area.
	MinRoomAreas map[string]float64 `json:"min_room_areas,omitempty" bson:"min_room_areas,omitempty"`

	// Style of generation.
	Style GenerationStyle `json:"style" bson:"style"`

	// Budget constraint (optional).
	Budget float64 `json:"budget,omitempty" bson:"budget,omitempty"`
}

// GenerationStyle is the style of generation.
type GenerationStyle string

const (
	GenerationStyleMinimal  GenerationStyle = "MINIMAL"
	GenerationStyleModerate GenerationStyle = "MODERATE"
	GenerationStyleCreative GenerationStyle = "CREATIVE"
)

// GeneratedVariant represents a generated layout variant.
type GeneratedVariant struct {
	// ID of the variant.
	ID string `json:"id" bson:"id"`

	// BranchID of the created branch.
	BranchID string `json:"branch_id" bson:"branch_id"`

	// Name of the variant.
	Name string `json:"name" bson:"name"`

	// Description of changes.
	Description string `json:"description" bson:"description"`

	// Score of the variant (0-1).
	Score float64 `json:"score" bson:"score"`

	// Changes made in this variant.
	Changes []VariantChange `json:"changes" bson:"changes"`

	// IsCompliant with building codes.
	IsCompliant bool `json:"is_compliant" bson:"is_compliant"`

	// EstimatedCost for implementation.
	EstimatedCost float64 `json:"estimated_cost" bson:"estimated_cost"`
}

// VariantChange represents a single change in a variant.
type VariantChange struct {
	// Type of change.
	Type string `json:"type" bson:"type"`

	// Description of the change.
	Description string `json:"description" bson:"description"`

	// ElementIDs affected.
	ElementIDs []string `json:"element_ids,omitempty" bson:"element_ids,omitempty"`
}

// NewGenerationJob creates a new generation job.
func NewGenerationJob(sceneID, branchID, prompt string, variantsCount int, options GenerationOptions) *GenerationJob {
	now := time.Now().UTC()
	return &GenerationJob{
		ID:            uuid.New(),
		SceneID:       sceneID,
		BranchID:      branchID,
		Prompt:        prompt,
		VariantsCount: variantsCount,
		Options:       options,
		Status:        JobStatusPending,
		Progress:      0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// Start marks the job as processing.
func (j *GenerationJob) Start() {
	j.Status = JobStatusProcessing
	j.UpdatedAt = time.Now().UTC()
}

// UpdateProgress updates the job progress.
func (j *GenerationJob) UpdateProgress(progress int) {
	j.Progress = progress
	j.UpdatedAt = time.Now().UTC()
}

// Complete marks the job as completed.
func (j *GenerationJob) Complete(variants []GeneratedVariant) {
	now := time.Now().UTC()
	j.Status = JobStatusCompleted
	j.Progress = 100
	j.Variants = variants
	j.UpdatedAt = now
	j.CompletedAt = &now
}

// Fail marks the job as failed.
func (j *GenerationJob) Fail(err string) {
	now := time.Now().UTC()
	j.Status = JobStatusFailed
	j.Error = err
	j.UpdatedAt = now
	j.CompletedAt = &now
}


// Package entity defines domain entities for Floor Plan Service.
package entity

import (
	"time"

	"github.com/google/uuid"
)

// FloorPlanStatus represents the status of a floor plan.
type FloorPlanStatus string

const (
	FloorPlanStatusUploaded    FloorPlanStatus = "UPLOADED"
	FloorPlanStatusProcessing  FloorPlanStatus = "PROCESSING"
	FloorPlanStatusRecognized  FloorPlanStatus = "RECOGNIZED"
	FloorPlanStatusConfirmed   FloorPlanStatus = "CONFIRMED"
	FloorPlanStatusFailed      FloorPlanStatus = "FAILED"
)

// FloorPlan represents a floor plan document.
type FloorPlan struct {
	// ID is the unique identifier.
	ID uuid.UUID `json:"id" db:"id"`

	// WorkspaceID is the workspace this floor plan belongs to.
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`

	// OwnerID is the user who uploaded this floor plan.
	OwnerID uuid.UUID `json:"owner_id" db:"owner_id"`

	// Name is the display name.
	Name string `json:"name" db:"name"`

	// Description is an optional description.
	Description string `json:"description" db:"description"`

	// Status of the floor plan.
	Status FloorPlanStatus `json:"status" db:"status"`

	// FileInfo contains file metadata.
	FileInfo *FileInfo `json:"file_info" db:"-"`

	// RecognitionJobID if recognition is in progress.
	RecognitionJobID *uuid.UUID `json:"recognition_job_id,omitempty" db:"recognition_job_id"`

	// SceneID of the created scene (after recognition).
	SceneID *uuid.UUID `json:"scene_id,omitempty" db:"scene_id"`

	// Metadata contains additional info.
	Metadata map[string]string `json:"metadata" db:"-"`

	// CreatedAt timestamp.
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt timestamp.
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// FileInfo contains file metadata.
type FileInfo struct {
	// ID is the file record ID.
	ID uuid.UUID `json:"id" db:"id"`

	// FloorPlanID is the parent floor plan.
	FloorPlanID uuid.UUID `json:"floor_plan_id" db:"floor_plan_id"`

	// OriginalName is the original file name.
	OriginalName string `json:"original_name" db:"original_name"`

	// StoragePath is the path in object storage.
	StoragePath string `json:"storage_path" db:"storage_path"`

	// MimeType is the file MIME type.
	MimeType string `json:"mime_type" db:"mime_type"`

	// Size is the file size in bytes.
	Size int64 `json:"size" db:"size"`

	// Checksum is MD5/SHA256 hash.
	Checksum string `json:"checksum" db:"checksum"`

	// Width in pixels (for images).
	Width int `json:"width,omitempty" db:"width"`

	// Height in pixels (for images).
	Height int `json:"height,omitempty" db:"height"`

	// CreatedAt timestamp.
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// NewFloorPlan creates a new floor plan.
func NewFloorPlan(workspaceID, ownerID uuid.UUID, name string) *FloorPlan {
	now := time.Now().UTC()
	return &FloorPlan{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		OwnerID:     ownerID,
		Name:        name,
		Status:      FloorPlanStatusUploaded,
		Metadata:    make(map[string]string),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewFileInfo creates file info for a floor plan.
func NewFileInfo(floorPlanID uuid.UUID, originalName, storagePath, mimeType string, size int64) *FileInfo {
	return &FileInfo{
		ID:           uuid.New(),
		FloorPlanID:  floorPlanID,
		OriginalName: originalName,
		StoragePath:  storagePath,
		MimeType:     mimeType,
		Size:         size,
		CreatedAt:    time.Now().UTC(),
	}
}

// StartProcessing marks the floor plan as processing.
func (f *FloorPlan) StartProcessing(jobID uuid.UUID) {
	f.Status = FloorPlanStatusProcessing
	f.RecognitionJobID = &jobID
	f.UpdatedAt = time.Now().UTC()
}

// CompleteRecognition marks recognition as complete.
func (f *FloorPlan) CompleteRecognition(sceneID uuid.UUID) {
	f.Status = FloorPlanStatusRecognized
	f.SceneID = &sceneID
	f.UpdatedAt = time.Now().UTC()
}

// ConfirmRecognition confirms the recognition result.
func (f *FloorPlan) ConfirmRecognition() {
	f.Status = FloorPlanStatusConfirmed
	f.UpdatedAt = time.Now().UTC()
}

// FailRecognition marks recognition as failed.
func (f *FloorPlan) FailRecognition() {
	f.Status = FloorPlanStatusFailed
	f.RecognitionJobID = nil
	f.UpdatedAt = time.Now().UTC()
}

// Thumbnail represents a generated thumbnail.
type Thumbnail struct {
	ID          uuid.UUID `json:"id" db:"id"`
	FloorPlanID uuid.UUID `json:"floor_plan_id" db:"floor_plan_id"`
	Size        string    `json:"size" db:"size"` // e.g., "128x128", "256x256"
	StoragePath string    `json:"storage_path" db:"storage_path"`
	Width       int       `json:"width" db:"width"`
	Height      int       `json:"height" db:"height"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// NewThumbnail creates a new thumbnail.
func NewThumbnail(floorPlanID uuid.UUID, size, storagePath string, width, height int) *Thumbnail {
	return &Thumbnail{
		ID:          uuid.New(),
		FloorPlanID: floorPlanID,
		Size:        size,
		StoragePath: storagePath,
		Width:       width,
		Height:      height,
		CreatedAt:   time.Now().UTC(),
	}
}


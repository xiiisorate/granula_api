// Package entity defines domain entities for Branch Service.
package entity

import (
	"time"

	"github.com/google/uuid"
)

// BranchStatus represents the status of a branch.
type BranchStatus string

const (
	BranchStatusActive   BranchStatus = "ACTIVE"
	BranchStatusMerged   BranchStatus = "MERGED"
	BranchStatusArchived BranchStatus = "ARCHIVED"
)

// Branch represents a scene version branch.
type Branch struct {
	ID             uuid.UUID    `json:"id" bson:"_id"`
	SceneID        uuid.UUID    `json:"scene_id" bson:"scene_id"`
	Name           string       `json:"name" bson:"name"`
	Description    string       `json:"description" bson:"description"`
	ParentBranchID *uuid.UUID   `json:"parent_branch_id,omitempty" bson:"parent_branch_id,omitempty"`
	IsMain         bool         `json:"is_main" bson:"is_main"`
	Status         BranchStatus `json:"status" bson:"status"`
	Version        int          `json:"version" bson:"version"`
	CreatedAt      time.Time    `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at" bson:"updated_at"`
}

// NewBranch creates a new branch.
func NewBranch(sceneID uuid.UUID, name string, parentID *uuid.UUID) *Branch {
	now := time.Now().UTC()
	return &Branch{
		ID:             uuid.New(),
		SceneID:        sceneID,
		Name:           name,
		ParentBranchID: parentID,
		IsMain:         parentID == nil,
		Status:         BranchStatusActive,
		Version:        1,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// Snapshot represents a point-in-time snapshot of a branch.
type Snapshot struct {
	ID        uuid.UUID `json:"id" bson:"_id"`
	BranchID  uuid.UUID `json:"branch_id" bson:"branch_id"`
	Name      string    `json:"name" bson:"name"`
	Version   int       `json:"version" bson:"version"`
	Data      []byte    `json:"data" bson:"data"` // Serialized elements
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

// NewSnapshot creates a new snapshot.
func NewSnapshot(branchID uuid.UUID, name string, version int, data []byte) *Snapshot {
	return &Snapshot{
		ID:        uuid.New(),
		BranchID:  branchID,
		Name:      name,
		Version:   version,
		Data:      data,
		CreatedAt: time.Now().UTC(),
	}
}


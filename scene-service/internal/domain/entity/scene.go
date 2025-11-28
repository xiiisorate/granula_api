// Package entity defines domain entities for Scene Service.
package entity

import (
	"time"

	"github.com/google/uuid"
)

// Scene represents a 3D scene containing walls, rooms, and furniture.
type Scene struct {
	// ID is the unique identifier.
	ID uuid.UUID `json:"id" bson:"_id"`

	// WorkspaceID is the workspace this scene belongs to.
	WorkspaceID uuid.UUID `json:"workspace_id" bson:"workspace_id"`

	// OwnerID is the user who owns this scene.
	OwnerID uuid.UUID `json:"owner_id" bson:"owner_id"`

	// FloorPlanID is the source floor plan (if any).
	FloorPlanID *uuid.UUID `json:"floor_plan_id,omitempty" bson:"floor_plan_id,omitempty"`

	// Name is the display name.
	Name string `json:"name" bson:"name"`

	// Description is an optional description.
	Description string `json:"description" bson:"description"`

	// Dimensions are the scene dimensions.
	Dimensions Dimensions3D `json:"dimensions" bson:"dimensions"`

	// MainBranchID is the ID of the main branch.
	MainBranchID uuid.UUID `json:"main_branch_id" bson:"main_branch_id"`

	// Metadata contains additional info.
	Metadata map[string]string `json:"metadata" bson:"metadata"`

	// CreatedAt timestamp.
	CreatedAt time.Time `json:"created_at" bson:"created_at"`

	// UpdatedAt timestamp.
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// Dimensions3D represents 3D dimensions.
type Dimensions3D struct {
	Width  float64 `json:"width" bson:"width"`   // X axis
	Depth  float64 `json:"depth" bson:"depth"`   // Y axis
	Height float64 `json:"height" bson:"height"` // Z axis
}

// NewScene creates a new scene.
func NewScene(workspaceID, ownerID uuid.UUID, name string) *Scene {
	now := time.Now().UTC()
	id := uuid.New()
	return &Scene{
		ID:           id,
		WorkspaceID:  workspaceID,
		OwnerID:      ownerID,
		Name:         name,
		MainBranchID: uuid.New(), // Will be set by Branch Service
		Metadata:     make(map[string]string),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// Point3D represents a 3D point.
type Point3D struct {
	X float64 `json:"x" bson:"x"`
	Y float64 `json:"y" bson:"y"`
	Z float64 `json:"z" bson:"z"`
}

// Point2D represents a 2D point.
type Point2D struct {
	X float64 `json:"x" bson:"x"`
	Y float64 `json:"y" bson:"y"`
}

// Rotation3D represents 3D rotation in Euler angles (degrees).
type Rotation3D struct {
	X float64 `json:"x" bson:"x"` // Pitch
	Y float64 `json:"y" bson:"y"` // Yaw
	Z float64 `json:"z" bson:"z"` // Roll
}


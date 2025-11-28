// Package entity defines domain entities for Scene Service.
package entity

import (
	"time"

	"github.com/google/uuid"
)

// ElementType represents the type of scene element.
type ElementType string

const (
	ElementTypeWall      ElementType = "WALL"
	ElementTypeRoom      ElementType = "ROOM"
	ElementTypeDoor      ElementType = "DOOR"
	ElementTypeWindow    ElementType = "WINDOW"
	ElementTypeFurniture ElementType = "FURNITURE"
	ElementTypeFixture   ElementType = "FIXTURE"
	ElementTypeDecor     ElementType = "DECOR"
)

// Element represents a scene element (wall, room, furniture, etc.).
type Element struct {
	// ID is the unique identifier.
	ID uuid.UUID `json:"id" bson:"_id"`

	// SceneID is the parent scene.
	SceneID uuid.UUID `json:"scene_id" bson:"scene_id"`

	// BranchID is the branch this element belongs to.
	BranchID uuid.UUID `json:"branch_id" bson:"branch_id"`

	// Type is the element type.
	Type ElementType `json:"type" bson:"type"`

	// Name is the display name.
	Name string `json:"name" bson:"name"`

	// Position in 3D space.
	Position Point3D `json:"position" bson:"position"`

	// Rotation in 3D space.
	Rotation Rotation3D `json:"rotation" bson:"rotation"`

	// Dimensions of the element.
	Dimensions Dimensions3D `json:"dimensions" bson:"dimensions"`

	// Properties specific to the element type.
	Properties ElementProperties `json:"properties" bson:"properties"`

	// ParentID for hierarchical elements.
	ParentID *uuid.UUID `json:"parent_id,omitempty" bson:"parent_id,omitempty"`

	// IsDeleted for soft delete (important for branching).
	IsDeleted bool `json:"is_deleted" bson:"is_deleted"`

	// Version for optimistic locking.
	Version int64 `json:"version" bson:"version"`

	// CreatedAt timestamp.
	CreatedAt time.Time `json:"created_at" bson:"created_at"`

	// UpdatedAt timestamp.
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// ElementProperties holds type-specific properties.
type ElementProperties struct {
	// Wall properties
	IsLoadBearing bool     `json:"is_load_bearing,omitempty" bson:"is_load_bearing,omitempty"`
	Thickness     float64  `json:"thickness,omitempty" bson:"thickness,omitempty"`
	Material      string   `json:"material,omitempty" bson:"material,omitempty"`
	StartPoint    *Point2D `json:"start_point,omitempty" bson:"start_point,omitempty"`
	EndPoint      *Point2D `json:"end_point,omitempty" bson:"end_point,omitempty"`

	// Room properties
	RoomType  string    `json:"room_type,omitempty" bson:"room_type,omitempty"`
	Area      float64   `json:"area,omitempty" bson:"area,omitempty"`
	IsWetZone bool      `json:"is_wet_zone,omitempty" bson:"is_wet_zone,omitempty"`
	Boundary  []Point2D `json:"boundary,omitempty" bson:"boundary,omitempty"`

	// Opening properties (door/window)
	OpeningType string `json:"opening_type,omitempty" bson:"opening_type,omitempty"`
	WallID      string `json:"wall_id,omitempty" bson:"wall_id,omitempty"`
	WallOffset  float64 `json:"wall_offset,omitempty" bson:"wall_offset,omitempty"`

	// Furniture properties
	CatalogID   string `json:"catalog_id,omitempty" bson:"catalog_id,omitempty"`
	ModelURL    string `json:"model_url,omitempty" bson:"model_url,omitempty"`
	Brand       string `json:"brand,omitempty" bson:"brand,omitempty"`
	Price       float64 `json:"price,omitempty" bson:"price,omitempty"`

	// Common properties
	Color       string            `json:"color,omitempty" bson:"color,omitempty"`
	TextureURL  string            `json:"texture_url,omitempty" bson:"texture_url,omitempty"`
	CustomData  map[string]string `json:"custom_data,omitempty" bson:"custom_data,omitempty"`
}

// NewElement creates a new element.
func NewElement(sceneID, branchID uuid.UUID, elemType ElementType, name string) *Element {
	now := time.Now().UTC()
	return &Element{
		ID:        uuid.New(),
		SceneID:   sceneID,
		BranchID:  branchID,
		Type:      elemType,
		Name:      name,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewWall creates a new wall element.
func NewWall(sceneID, branchID uuid.UUID, name string, start, end Point2D, thickness float64, isLoadBearing bool) *Element {
	wall := NewElement(sceneID, branchID, ElementTypeWall, name)
	wall.Properties = ElementProperties{
		StartPoint:    &start,
		EndPoint:      &end,
		Thickness:     thickness,
		IsLoadBearing: isLoadBearing,
	}
	return wall
}

// NewRoom creates a new room element.
func NewRoom(sceneID, branchID uuid.UUID, name, roomType string, boundary []Point2D, isWetZone bool) *Element {
	room := NewElement(sceneID, branchID, ElementTypeRoom, name)
	room.Properties = ElementProperties{
		RoomType:  roomType,
		Boundary:  boundary,
		IsWetZone: isWetZone,
	}
	return room
}

// NewFurniture creates a new furniture element.
func NewFurniture(sceneID, branchID uuid.UUID, name, catalogID string, position Point3D, dimensions Dimensions3D) *Element {
	furniture := NewElement(sceneID, branchID, ElementTypeFurniture, name)
	furniture.Position = position
	furniture.Dimensions = dimensions
	furniture.Properties = ElementProperties{
		CatalogID: catalogID,
	}
	return furniture
}

// Update updates element and increments version.
func (e *Element) Update() {
	e.Version++
	e.UpdatedAt = time.Now().UTC()
}

// SoftDelete marks element as deleted.
func (e *Element) SoftDelete() {
	e.IsDeleted = true
	e.Update()
}


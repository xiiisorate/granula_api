// Package entity defines domain entities for AI Service.
// These entities are used for floor plan recognition, AI chat, and scene generation.
package entity

import (
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// JOB STATUS
// =============================================================================

// JobStatus represents the status of an async job.
type JobStatus string

const (
	JobStatusPending    JobStatus = "PENDING"
	JobStatusProcessing JobStatus = "PROCESSING"
	JobStatusCompleted  JobStatus = "COMPLETED"
	JobStatusFailed     JobStatus = "FAILED"
	JobStatusCancelled  JobStatus = "CANCELLED"
)

// =============================================================================
// RECOGNITION JOB
// =============================================================================

// RecognitionJob represents a floor plan recognition job.
// Jobs are stored in MongoDB and processed asynchronously.
type RecognitionJob struct {
	// ID is the job identifier.
	ID uuid.UUID `json:"id" bson:"_id"`

	// FloorPlanID is the floor plan being recognized.
	FloorPlanID string `json:"floor_plan_id" bson:"floor_plan_id"`

	// Status of the job.
	Status JobStatus `json:"status" bson:"status"`

	// Progress percentage (0-100).
	Progress int `json:"progress" bson:"progress"`

	// Options used for recognition.
	Options RecognitionOptions `json:"options" bson:"options"`

	// Result of recognition (if completed).
	Result *RecognitionResult `json:"result,omitempty" bson:"result,omitempty"`

	// Error message (if failed).
	Error string `json:"error,omitempty" bson:"error,omitempty"`

	// CreatedAt timestamp.
	CreatedAt time.Time `json:"created_at" bson:"created_at"`

	// UpdatedAt timestamp.
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`

	// CompletedAt timestamp.
	CompletedAt *time.Time `json:"completed_at,omitempty" bson:"completed_at,omitempty"`
}

// RecognitionOptions are options for floor plan recognition.
type RecognitionOptions struct {
	DetectLoadBearing bool    `json:"detect_load_bearing" bson:"detect_load_bearing"`
	DetectWetZones    bool    `json:"detect_wet_zones" bson:"detect_wet_zones"`
	DetectFurniture   bool    `json:"detect_furniture" bson:"detect_furniture"`
	Scale             float64 `json:"scale" bson:"scale"`
	Orientation       int     `json:"orientation" bson:"orientation"`
	DetailLevel       int     `json:"detail_level" bson:"detail_level"`
}

// =============================================================================
// RECOGNITION RESULT — Full 3D Model Data
// =============================================================================

// RecognitionResult is the complete result of floor plan recognition.
// This data is used to create a 3D scene in MongoDB.
// Format follows MongoDB scenes collection schema.
type RecognitionResult struct {
	// Confidence of the recognition (0-1).
	Confidence float64 `json:"confidence" bson:"confidence"`

	// Bounds — 3D dimensions of the apartment.
	Bounds Bounds3D `json:"bounds" bson:"bounds"`

	// Dimensions — 2D dimensions (legacy, for backward compatibility).
	Dimensions Dimensions2D `json:"dimensions" bson:"dimensions"`

	// TotalArea in square meters.
	TotalArea float64 `json:"total_area" bson:"total_area"`

	// Stats — summary statistics.
	Stats RecognitionStats `json:"stats" bson:"stats"`

	// Elements contains all scene elements.
	Elements SceneElements `json:"elements" bson:"elements"`

	// Walls detected (legacy, moved to Elements.Walls).
	Walls []RecognizedWall `json:"walls" bson:"walls"`

	// Rooms detected (legacy, moved to Elements.Rooms).
	Rooms []RecognizedRoom `json:"rooms" bson:"rooms"`

	// Openings detected (legacy, moved to Elements.Walls[].Openings).
	Openings []RecognizedOpening `json:"openings" bson:"openings"`

	// Recognition metadata.
	Recognition RecognitionMeta `json:"recognition" bson:"recognition"`

	// Warnings during recognition.
	Warnings []string `json:"warnings,omitempty" bson:"warnings,omitempty"`

	// Notes about the plan.
	Notes []string `json:"notes,omitempty" bson:"notes,omitempty"`

	// ModelVersion used.
	ModelVersion string `json:"model_version" bson:"model_version"`

	// ProcessingTimeMs is the processing time in milliseconds.
	ProcessingTimeMs int64 `json:"processing_time_ms" bson:"processing_time_ms"`

	// Metadata from AI response (legacy).
	Metadata *RecognitionMetadata `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// Bounds3D represents 3D dimensions of the apartment.
type Bounds3D struct {
	Width  float64 `json:"width" bson:"width"`   // X axis (meters)
	Height float64 `json:"height" bson:"height"` // Y axis - ceiling height (meters)
	Depth  float64 `json:"depth" bson:"depth"`   // Z axis (meters)
}

// RecognitionStats contains summary statistics.
type RecognitionStats struct {
	TotalArea      float64 `json:"totalArea" bson:"totalArea"`
	RoomsCount     int     `json:"roomsCount" bson:"roomsCount"`
	WallsCount     int     `json:"wallsCount" bson:"wallsCount"`
	FurnitureCount int     `json:"furnitureCount" bson:"furnitureCount"`
}

// SceneElements contains all elements of the scene.
type SceneElements struct {
	Walls     []Wall3D     `json:"walls" bson:"walls"`
	Rooms     []Room3D     `json:"rooms" bson:"rooms"`
	Furniture []Furniture  `json:"furniture" bson:"furniture"`
	Utilities []Utility    `json:"utilities" bson:"utilities"`
}

// RecognitionMeta contains metadata about the recognition process.
type RecognitionMeta struct {
	SourceType     string   `json:"sourceType" bson:"sourceType"`         // BTI, technical_passport, sketch
	Quality        string   `json:"quality" bson:"quality"`               // high, medium, low
	Scale          string   `json:"scale" bson:"scale"`                   // 1:50, 1:100, 1:200
	Orientation    int      `json:"orientation" bson:"orientation"`       // 0, 90, 180, 270
	HasDimensions  bool     `json:"hasDimensions" bson:"hasDimensions"`   // Has dimension lines
	HasAnnotations bool     `json:"hasAnnotations" bson:"hasAnnotations"` // Has text labels
	BuildingType   string   `json:"buildingType" bson:"buildingType"`     // panel, brick, monolith
	FloorNumber    *int     `json:"floorNumber" bson:"floorNumber"`       // Floor number if detected
	Warnings       []string `json:"warnings" bson:"warnings"`
	Notes          []string `json:"notes" bson:"notes"`
}

// =============================================================================
// WALL 3D — Full Wall Definition for 3D Model
// =============================================================================

// Wall3D represents a wall with full 3D properties.
type Wall3D struct {
	ID        string     `json:"id" bson:"id"`
	Type      string     `json:"type" bson:"type"` // Always "wall"
	Name      string     `json:"name" bson:"name"`
	Start     Point3D    `json:"start" bson:"start"`
	End       Point3D    `json:"end" bson:"end"`
	Height    float64    `json:"height" bson:"height"`
	Thickness float64    `json:"thickness" bson:"thickness"`

	// Properties of the wall.
	Properties WallProperties `json:"properties" bson:"properties"`

	// Openings in this wall (doors, windows).
	Openings []Opening3D `json:"openings" bson:"openings"`

	// Metadata for editor.
	Metadata ElementMetadata `json:"metadata" bson:"metadata"`
}

// WallProperties contains properties of a wall.
type WallProperties struct {
	IsLoadBearing  bool   `json:"isLoadBearing" bson:"isLoadBearing"`
	Material       string `json:"material" bson:"material"`             // brick, concrete, drywall, block, panel
	CanDemolish    bool   `json:"canDemolish" bson:"canDemolish"`       // Can this wall be demolished
	StructuralType string `json:"structuralType" bson:"structuralType"` // external, internal_bearing, partition
}

// Opening3D represents a door or window in a wall.
type Opening3D struct {
	ID            string   `json:"id" bson:"id"`
	Type          string   `json:"type" bson:"type"`                   // door, window, passage, arch
	Subtype       string   `json:"subtype" bson:"subtype"`             // single_swing, double, sliding, entrance
	Position      float64  `json:"position" bson:"position"`           // Distance from wall start (meters)
	Width         float64  `json:"width" bson:"width"`
	Height        float64  `json:"height" bson:"height"`
	Elevation     float64  `json:"elevation" bson:"elevation"`         // Height from floor (0 for doors)
	OpensTo       string   `json:"opens_to" bson:"opens_to"`           // left, right, inward, outward
	HasDoor       bool     `json:"has_door" bson:"has_door"`           // Has door leaf
	ConnectsRooms []string `json:"connects_rooms" bson:"connects_rooms"` // Room IDs
}

// =============================================================================
// ROOM 3D — Full Room Definition for 3D Model
// =============================================================================

// Room3D represents a room with full 3D properties.
type Room3D struct {
	ID        string          `json:"id" bson:"id"`
	Type      string          `json:"type" bson:"type"`           // Always "room"
	Name      string          `json:"name" bson:"name"`           // "Кухня", "Спальня 1"
	RoomType  string          `json:"roomType" bson:"roomType"`   // kitchen, bedroom, bathroom, etc.
	Polygon   []Point2D       `json:"polygon" bson:"polygon"`     // Closed polygon (clockwise)
	Area      float64         `json:"area" bson:"area"`           // Square meters
	Perimeter float64         `json:"perimeter" bson:"perimeter"` // Meters

	// Properties of the room.
	Properties RoomProperties `json:"properties" bson:"properties"`

	// Wall IDs that form this room.
	WallIDs []string `json:"wallIds" bson:"wallIds"`

	// Metadata.
	Metadata RoomMetadata `json:"metadata" bson:"metadata"`
}

// RoomProperties contains properties of a room.
type RoomProperties struct {
	HasWetZone     bool    `json:"hasWetZone" bson:"hasWetZone"`
	HasVentilation bool    `json:"hasVentilation" bson:"hasVentilation"`
	HasWindow      bool    `json:"hasWindow" bson:"hasWindow"`
	MinAllowedArea float64 `json:"minAllowedArea" bson:"minAllowedArea"` // SNiP requirement
	CeilingHeight  float64 `json:"ceilingHeight" bson:"ceilingHeight"`  // If different from default
}

// RoomMetadata contains metadata for a room.
type RoomMetadata struct {
	Confidence  float64 `json:"confidence" bson:"confidence"`
	LabelOnPlan string  `json:"labelOnPlan" bson:"labelOnPlan"` // Text from plan
	AreaOnPlan  float64 `json:"areaOnPlan" bson:"areaOnPlan"`   // Area written on plan
}

// =============================================================================
// FURNITURE — Equipment and Furniture
// =============================================================================

// Furniture represents furniture or equipment in the scene.
type Furniture struct {
	ID            string             `json:"id" bson:"id"`
	Type          string             `json:"type" bson:"type"` // Always "furniture"
	Name          string             `json:"name" bson:"name"`
	FurnitureType string             `json:"furnitureType" bson:"furnitureType"` // sink, toilet, bathtub, gas_stove
	Position      Point3D            `json:"position" bson:"position"`
	Rotation      Rotation3D         `json:"rotation" bson:"rotation"`
	Dimensions    Dimensions3D       `json:"dimensions" bson:"dimensions"`
	RoomID        string             `json:"roomId" bson:"roomId"`
	Properties    FurnitureProps     `json:"properties" bson:"properties"`
	Metadata      ElementMetadata    `json:"metadata" bson:"metadata"`
}

// FurnitureProps contains properties of furniture.
type FurnitureProps struct {
	CanRelocate   bool   `json:"canRelocate" bson:"canRelocate"`
	Category      string `json:"category" bson:"category"`           // bathroom, kitchen, living
	RequiresWater bool   `json:"requiresWater" bson:"requiresWater"`
	RequiresGas   bool   `json:"requiresGas" bson:"requiresGas"`
	RequiresDrain bool   `json:"requiresDrain" bson:"requiresDrain"`
}

// =============================================================================
// UTILITIES — Engineering Systems
// =============================================================================

// Utility represents engineering utilities (risers, vents, etc.).
type Utility struct {
	ID          string          `json:"id" bson:"id"`
	Type        string          `json:"type" bson:"type"` // Always "utility"
	Name        string          `json:"name" bson:"name"`
	UtilityType string          `json:"utilityType" bson:"utilityType"` // water_riser, sewer_riser, ventilation
	Position    Point3D         `json:"position" bson:"position"`
	Dimensions  UtilityDims     `json:"dimensions" bson:"dimensions"`
	RoomID      string          `json:"roomId" bson:"roomId"`
	Properties  UtilityProps    `json:"properties" bson:"properties"`
	Metadata    ElementMetadata `json:"metadata" bson:"metadata"`
}

// UtilityDims contains dimensions of a utility element.
type UtilityDims struct {
	Diameter float64 `json:"diameter" bson:"diameter"` // For round utilities
	Width    float64 `json:"width" bson:"width"`       // For rectangular
	Depth    float64 `json:"depth" bson:"depth"`       // For rectangular
}

// UtilityProps contains properties of a utility.
type UtilityProps struct {
	CanRelocate         bool    `json:"canRelocate" bson:"canRelocate"`
	ProtectionZone      float64 `json:"protectionZone" bson:"protectionZone"` // Radius in meters
	SharedWithNeighbors bool    `json:"sharedWithNeighbors" bson:"sharedWithNeighbors"`
}

// =============================================================================
// COMMON TYPES
// =============================================================================

// Dimensions2D represents 2D dimensions.
type Dimensions2D struct {
	Width  float64 `json:"width" bson:"width"`
	Height float64 `json:"height" bson:"height"`
}

// Dimensions3D represents 3D dimensions.
type Dimensions3D struct {
	Width  float64 `json:"width" bson:"width"`
	Height float64 `json:"height" bson:"height"`
	Depth  float64 `json:"depth" bson:"depth"`
}

// Point2D represents a 2D point.
type Point2D struct {
	X float64 `json:"x" bson:"x"`
	Y float64 `json:"y" bson:"y"`
}

// Point3D represents a 3D point.
type Point3D struct {
	X float64 `json:"x" bson:"x"`
	Y float64 `json:"y" bson:"y"`
	Z float64 `json:"z" bson:"z"`
}

// Rotation3D represents rotation in 3D space.
type Rotation3D struct {
	X float64 `json:"x" bson:"x"`
	Y float64 `json:"y" bson:"y"`
	Z float64 `json:"z" bson:"z"`
}

// ElementMetadata contains common metadata for scene elements.
type ElementMetadata struct {
	Confidence float64 `json:"confidence" bson:"confidence"`
	Source     string  `json:"source" bson:"source"` // recognition, manual, generated
	Locked     bool    `json:"locked" bson:"locked"`
	Visible    bool    `json:"visible" bson:"visible"`
	ModelURL   string  `json:"modelUrl,omitempty" bson:"modelUrl,omitempty"` // 3D model URL
}

// =============================================================================
// LEGACY TYPES — For backward compatibility
// =============================================================================

// RecognizedWall represents a detected wall (legacy format).
type RecognizedWall struct {
	TempID                string  `json:"temp_id" bson:"temp_id"`
	Start                 Point2D `json:"start" bson:"start"`
	End                   Point2D `json:"end" bson:"end"`
	Thickness             float64 `json:"thickness" bson:"thickness"`
	IsLoadBearing         bool    `json:"is_load_bearing" bson:"is_load_bearing"`
	Material              string  `json:"material,omitempty" bson:"material,omitempty"`
	CanDemolish           bool    `json:"can_demolish,omitempty" bson:"can_demolish,omitempty"`
	Confidence            float64 `json:"confidence" bson:"confidence"`
	LoadBearingConfidence float64 `json:"load_bearing_confidence" bson:"load_bearing_confidence"`
}

// RecognizedRoom represents a detected room (legacy format).
type RecognizedRoom struct {
	TempID     string    `json:"temp_id" bson:"temp_id"`
	Type       string    `json:"type" bson:"type"`
	Name       string    `json:"name,omitempty" bson:"name,omitempty"`
	Boundary   []Point2D `json:"boundary" bson:"boundary"`
	Polygon    []Point2D `json:"polygon,omitempty" bson:"polygon,omitempty"`
	Area       float64   `json:"area" bson:"area"`
	Perimeter  float64   `json:"perimeter,omitempty" bson:"perimeter,omitempty"`
	IsWetZone  bool      `json:"is_wet_zone" bson:"is_wet_zone"`
	HasWindow  bool      `json:"has_window,omitempty" bson:"has_window,omitempty"`
	Confidence float64   `json:"confidence" bson:"confidence"`
	WallIDs    []string  `json:"wall_ids" bson:"wall_ids"`
}

// RecognizedOpening represents a detected door or window (legacy format).
type RecognizedOpening struct {
	TempID        string   `json:"temp_id" bson:"temp_id"`
	Type          string   `json:"type" bson:"type"` // "door", "window", "arch"
	Subtype       string   `json:"subtype,omitempty" bson:"subtype,omitempty"`
	Position      Point2D  `json:"position" bson:"position"`
	Width         float64  `json:"width" bson:"width"`
	Height        float64  `json:"height,omitempty" bson:"height,omitempty"`
	Elevation     float64  `json:"elevation,omitempty" bson:"elevation,omitempty"`
	WallID        string   `json:"wall_id" bson:"wall_id"`
	OpensTo       string   `json:"opens_to,omitempty" bson:"opens_to,omitempty"`
	ConnectsRooms []string `json:"connects_rooms,omitempty" bson:"connects_rooms,omitempty"`
	Confidence    float64  `json:"confidence" bson:"confidence"`
}

// RecognizedElement represents detected furniture or equipment (legacy format).
type RecognizedElement struct {
	TempID        string       `json:"temp_id" bson:"temp_id"`
	Type          string       `json:"type" bson:"type"`                               // furniture, utility
	ElementType   string       `json:"element_type,omitempty" bson:"element_type,omitempty"` // sink, toilet, etc.
	FurnitureType string       `json:"furniture_type,omitempty" bson:"furniture_type,omitempty"`
	Name          string       `json:"name,omitempty" bson:"name,omitempty"`
	Position      Point2D      `json:"position" bson:"position"`
	Dimensions    Dimensions2D `json:"dimensions" bson:"dimensions"`
	Rotation      float64      `json:"rotation" bson:"rotation"`
	RoomID        string       `json:"room_id" bson:"room_id"`
	CanRelocate   bool         `json:"can_relocate,omitempty" bson:"can_relocate,omitempty"`
	Confidence    float64      `json:"confidence" bson:"confidence"`
}

// RecognitionMetadata contains metadata about recognition process (legacy).
type RecognitionMetadata struct {
	ModelVersion        string `json:"model_version" bson:"model_version"`
	ProcessingTimeMs    int64  `json:"processing_time_ms" bson:"processing_time_ms"`
	DetectedScale       string `json:"detected_scale" bson:"detected_scale"`
	DetectedOrientation int    `json:"detected_orientation" bson:"detected_orientation"`
}

// =============================================================================
// JOB LIFECYCLE METHODS
// =============================================================================

// NewRecognitionJob creates a new recognition job.
func NewRecognitionJob(floorPlanID string, options RecognitionOptions) *RecognitionJob {
	now := time.Now().UTC()
	return &RecognitionJob{
		ID:          uuid.New(),
		FloorPlanID: floorPlanID,
		Status:      JobStatusPending,
		Progress:    0,
		Options:     options,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Start marks the job as processing.
func (j *RecognitionJob) Start() {
	j.Status = JobStatusProcessing
	j.UpdatedAt = time.Now().UTC()
}

// UpdateProgress updates the job progress.
func (j *RecognitionJob) UpdateProgress(progress int) {
	j.Progress = progress
	j.UpdatedAt = time.Now().UTC()
}

// Complete marks the job as completed.
func (j *RecognitionJob) Complete(result *RecognitionResult) {
	now := time.Now().UTC()
	j.Status = JobStatusCompleted
	j.Progress = 100
	j.Result = result
	j.UpdatedAt = now
	j.CompletedAt = &now
}

// Fail marks the job as failed.
func (j *RecognitionJob) Fail(err string) {
	now := time.Now().UTC()
	j.Status = JobStatusFailed
	j.Error = err
	j.UpdatedAt = now
	j.CompletedAt = &now
}

// Cancel marks the job as cancelled.
func (j *RecognitionJob) Cancel() {
	now := time.Now().UTC()
	j.Status = JobStatusCancelled
	j.UpdatedAt = now
	j.CompletedAt = &now
}

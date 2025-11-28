// Package entity defines domain entities for AI Service.
package entity

import (
	"time"

	"github.com/google/uuid"
)

// JobStatus represents the status of an async job.
type JobStatus string

const (
	JobStatusPending    JobStatus = "PENDING"
	JobStatusProcessing JobStatus = "PROCESSING"
	JobStatusCompleted  JobStatus = "COMPLETED"
	JobStatusFailed     JobStatus = "FAILED"
	JobStatusCancelled  JobStatus = "CANCELLED"
)

// RecognitionJob represents a floor plan recognition job.
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

// RecognitionResult is the result of floor plan recognition.
type RecognitionResult struct {
	// Confidence of the recognition (0-1).
	Confidence float64 `json:"confidence" bson:"confidence"`

	// Dimensions of the floor plan.
	Dimensions Dimensions2D `json:"dimensions" bson:"dimensions"`

	// TotalArea in square meters.
	TotalArea float64 `json:"total_area" bson:"total_area"`

	// Walls detected.
	Walls []RecognizedWall `json:"walls" bson:"walls"`

	// Rooms detected.
	Rooms []RecognizedRoom `json:"rooms" bson:"rooms"`

	// Openings detected (doors, windows).
	Openings []RecognizedOpening `json:"openings" bson:"openings"`

	// Elements detected (furniture, equipment).
	Elements []RecognizedElement `json:"elements" bson:"elements"`

	// Warnings during recognition.
	Warnings []string `json:"warnings,omitempty" bson:"warnings,omitempty"`

	// ModelVersion used.
	ModelVersion string `json:"model_version" bson:"model_version"`

	// ProcessingTimeMs is the processing time in milliseconds.
	ProcessingTimeMs int64 `json:"processing_time_ms" bson:"processing_time_ms"`
}

// Dimensions2D represents 2D dimensions.
type Dimensions2D struct {
	Width  float64 `json:"width" bson:"width"`
	Height float64 `json:"height" bson:"height"`
}

// Point2D represents a 2D point.
type Point2D struct {
	X float64 `json:"x" bson:"x"`
	Y float64 `json:"y" bson:"y"`
}

// RecognizedWall represents a detected wall.
type RecognizedWall struct {
	TempID                string  `json:"temp_id" bson:"temp_id"`
	Start                 Point2D `json:"start" bson:"start"`
	End                   Point2D `json:"end" bson:"end"`
	Thickness             float64 `json:"thickness" bson:"thickness"`
	IsLoadBearing         bool    `json:"is_load_bearing" bson:"is_load_bearing"`
	Confidence            float64 `json:"confidence" bson:"confidence"`
	LoadBearingConfidence float64 `json:"load_bearing_confidence" bson:"load_bearing_confidence"`
}

// RecognizedRoom represents a detected room.
type RecognizedRoom struct {
	TempID     string    `json:"temp_id" bson:"temp_id"`
	Type       string    `json:"type" bson:"type"`
	Boundary   []Point2D `json:"boundary" bson:"boundary"`
	Area       float64   `json:"area" bson:"area"`
	IsWetZone  bool      `json:"is_wet_zone" bson:"is_wet_zone"`
	Confidence float64   `json:"confidence" bson:"confidence"`
	WallIDs    []string  `json:"wall_ids" bson:"wall_ids"`
}

// RecognizedOpening represents a detected door or window.
type RecognizedOpening struct {
	TempID     string  `json:"temp_id" bson:"temp_id"`
	Type       string  `json:"type" bson:"type"` // "door", "window", "arch"
	Position   Point2D `json:"position" bson:"position"`
	Width      float64 `json:"width" bson:"width"`
	WallID     string  `json:"wall_id" bson:"wall_id"`
	Confidence float64 `json:"confidence" bson:"confidence"`
}

// RecognizedElement represents detected furniture or equipment.
type RecognizedElement struct {
	TempID     string       `json:"temp_id" bson:"temp_id"`
	Type       string       `json:"type" bson:"type"`
	Position   Point2D      `json:"position" bson:"position"`
	Dimensions Dimensions2D `json:"dimensions" bson:"dimensions"`
	Rotation   float64      `json:"rotation" bson:"rotation"`
	RoomID     string       `json:"room_id" bson:"room_id"`
	Confidence float64      `json:"confidence" bson:"confidence"`
}

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


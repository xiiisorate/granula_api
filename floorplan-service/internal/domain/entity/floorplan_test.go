// Package entity provides tests for floor plan domain entities.
package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestNewFloorPlan verifies floor plan creation.
func TestNewFloorPlan(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	ownerID := uuid.New()
	name := "Квартира на Пушкинской"

	fp := NewFloorPlan(workspaceID, ownerID, name)

	if fp == nil {
		t.Fatal("expected non-nil floor plan")
	}

	if fp.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}

	if fp.WorkspaceID != workspaceID {
		t.Errorf("expected workspace_id %s, got %s", workspaceID, fp.WorkspaceID)
	}

	if fp.OwnerID != ownerID {
		t.Errorf("expected owner_id %s, got %s", ownerID, fp.OwnerID)
	}

	if fp.Name != name {
		t.Errorf("expected name '%s', got '%s'", name, fp.Name)
	}

	if fp.Status != FloorPlanStatusUploaded {
		t.Errorf("expected status %s, got %s", FloorPlanStatusUploaded, fp.Status)
	}

	if fp.Metadata == nil {
		t.Error("expected initialized metadata map")
	}

	if fp.CreatedAt.IsZero() {
		t.Error("expected non-zero created_at")
	}

	if fp.UpdatedAt.IsZero() {
		t.Error("expected non-zero updated_at")
	}
}

// TestFloorPlan_StartProcessing verifies transition to processing status.
func TestFloorPlan_StartProcessing(t *testing.T) {
	t.Parallel()

	fp := NewFloorPlan(uuid.New(), uuid.New(), "Test")
	jobID := uuid.New()

	beforeUpdate := fp.UpdatedAt
	time.Sleep(1 * time.Millisecond) // Ensure time difference

	fp.StartProcessing(jobID)

	if fp.Status != FloorPlanStatusProcessing {
		t.Errorf("expected status %s, got %s", FloorPlanStatusProcessing, fp.Status)
	}

	if fp.RecognitionJobID == nil {
		t.Fatal("expected non-nil recognition_job_id")
	}

	if *fp.RecognitionJobID != jobID {
		t.Errorf("expected recognition_job_id %s, got %s", jobID, *fp.RecognitionJobID)
	}

	if !fp.UpdatedAt.After(beforeUpdate) {
		t.Error("expected updated_at to be updated")
	}
}

// TestFloorPlan_CompleteRecognition verifies transition to recognized status.
func TestFloorPlan_CompleteRecognition(t *testing.T) {
	t.Parallel()

	fp := NewFloorPlan(uuid.New(), uuid.New(), "Test")
	fp.StartProcessing(uuid.New())

	sceneID := uuid.New()
	fp.CompleteRecognition(sceneID)

	if fp.Status != FloorPlanStatusRecognized {
		t.Errorf("expected status %s, got %s", FloorPlanStatusRecognized, fp.Status)
	}

	if fp.SceneID == nil {
		t.Fatal("expected non-nil scene_id")
	}

	if *fp.SceneID != sceneID {
		t.Errorf("expected scene_id %s, got %s", sceneID, *fp.SceneID)
	}
}

// TestFloorPlan_ConfirmRecognition verifies transition to confirmed status.
func TestFloorPlan_ConfirmRecognition(t *testing.T) {
	t.Parallel()

	fp := NewFloorPlan(uuid.New(), uuid.New(), "Test")
	fp.StartProcessing(uuid.New())
	fp.CompleteRecognition(uuid.New())
	fp.ConfirmRecognition()

	if fp.Status != FloorPlanStatusConfirmed {
		t.Errorf("expected status %s, got %s", FloorPlanStatusConfirmed, fp.Status)
	}
}

// TestFloorPlan_FailRecognition verifies transition to failed status.
func TestFloorPlan_FailRecognition(t *testing.T) {
	t.Parallel()

	fp := NewFloorPlan(uuid.New(), uuid.New(), "Test")
	jobID := uuid.New()
	fp.StartProcessing(jobID)

	fp.FailRecognition()

	if fp.Status != FloorPlanStatusFailed {
		t.Errorf("expected status %s, got %s", FloorPlanStatusFailed, fp.Status)
	}

	if fp.RecognitionJobID != nil {
		t.Error("expected recognition_job_id to be cleared")
	}
}

// TestNewFileInfo verifies file info creation.
func TestNewFileInfo(t *testing.T) {
	t.Parallel()

	floorPlanID := uuid.New()
	originalName := "plan.pdf"
	storagePath := "/floor-plans/workspace-123/plan-456.pdf"
	mimeType := "application/pdf"
	size := int64(1024000)

	fi := NewFileInfo(floorPlanID, originalName, storagePath, mimeType, size)

	if fi == nil {
		t.Fatal("expected non-nil file info")
	}

	if fi.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}

	if fi.FloorPlanID != floorPlanID {
		t.Errorf("expected floor_plan_id %s, got %s", floorPlanID, fi.FloorPlanID)
	}

	if fi.OriginalName != originalName {
		t.Errorf("expected original_name '%s', got '%s'", originalName, fi.OriginalName)
	}

	if fi.StoragePath != storagePath {
		t.Errorf("expected storage_path '%s', got '%s'", storagePath, fi.StoragePath)
	}

	if fi.MimeType != mimeType {
		t.Errorf("expected mime_type '%s', got '%s'", mimeType, fi.MimeType)
	}

	if fi.Size != size {
		t.Errorf("expected size %d, got %d", size, fi.Size)
	}
}

// TestNewThumbnail verifies thumbnail creation.
func TestNewThumbnail(t *testing.T) {
	t.Parallel()

	floorPlanID := uuid.New()
	size := "256x256"
	storagePath := "/floor-plans/workspace-123/plan-456_256x256.jpg"
	width := 256
	height := 192

	th := NewThumbnail(floorPlanID, size, storagePath, width, height)

	if th == nil {
		t.Fatal("expected non-nil thumbnail")
	}

	if th.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}

	if th.FloorPlanID != floorPlanID {
		t.Errorf("expected floor_plan_id %s, got %s", floorPlanID, th.FloorPlanID)
	}

	if th.Size != size {
		t.Errorf("expected size '%s', got '%s'", size, th.Size)
	}

	if th.Width != width {
		t.Errorf("expected width %d, got %d", width, th.Width)
	}

	if th.Height != height {
		t.Errorf("expected height %d, got %d", height, th.Height)
	}
}

// TestFloorPlanStatusStrings verifies status string values.
func TestFloorPlanStatusStrings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status FloorPlanStatus
		str    string
	}{
		{FloorPlanStatusUploaded, "UPLOADED"},
		{FloorPlanStatusProcessing, "PROCESSING"},
		{FloorPlanStatusRecognized, "RECOGNIZED"},
		{FloorPlanStatusConfirmed, "CONFIRMED"},
		{FloorPlanStatusFailed, "FAILED"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if string(tt.status) != tt.str {
				t.Errorf("expected '%s', got '%s'", tt.str, string(tt.status))
			}
		})
	}
}

// TestFloorPlan_StatusWorkflow verifies complete status workflow.
func TestFloorPlan_StatusWorkflow(t *testing.T) {
	t.Parallel()

	fp := NewFloorPlan(uuid.New(), uuid.New(), "Test Plan")

	// Initial status
	if fp.Status != FloorPlanStatusUploaded {
		t.Errorf("step 1: expected %s, got %s", FloorPlanStatusUploaded, fp.Status)
	}

	// Start processing
	jobID := uuid.New()
	fp.StartProcessing(jobID)
	if fp.Status != FloorPlanStatusProcessing {
		t.Errorf("step 2: expected %s, got %s", FloorPlanStatusProcessing, fp.Status)
	}

	// Complete recognition
	sceneID := uuid.New()
	fp.CompleteRecognition(sceneID)
	if fp.Status != FloorPlanStatusRecognized {
		t.Errorf("step 3: expected %s, got %s", FloorPlanStatusRecognized, fp.Status)
	}

	// Confirm
	fp.ConfirmRecognition()
	if fp.Status != FloorPlanStatusConfirmed {
		t.Errorf("step 4: expected %s, got %s", FloorPlanStatusConfirmed, fp.Status)
	}
}

// TestFloorPlan_Metadata verifies metadata operations.
func TestFloorPlan_Metadata(t *testing.T) {
	t.Parallel()

	fp := NewFloorPlan(uuid.New(), uuid.New(), "Test")

	// Metadata should be initialized
	if fp.Metadata == nil {
		t.Fatal("expected non-nil metadata")
	}

	// Add metadata
	fp.Metadata["source"] = "mobile_upload"
	fp.Metadata["building_type"] = "apartment"

	if fp.Metadata["source"] != "mobile_upload" {
		t.Errorf("expected metadata source='mobile_upload', got '%s'", fp.Metadata["source"])
	}
}

// BenchmarkNewFloorPlan benchmarks floor plan creation.
func BenchmarkNewFloorPlan(b *testing.B) {
	workspaceID := uuid.New()
	ownerID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewFloorPlan(workspaceID, ownerID, "Test Floor Plan")
	}
}

// BenchmarkStatusWorkflow benchmarks status transitions.
func BenchmarkStatusWorkflow(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fp := NewFloorPlan(uuid.New(), uuid.New(), "Test")
		fp.StartProcessing(uuid.New())
		fp.CompleteRecognition(uuid.New())
		fp.ConfirmRecognition()
	}
}

// BenchmarkNewFileInfo benchmarks file info creation.
func BenchmarkNewFileInfo(b *testing.B) {
	floorPlanID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewFileInfo(floorPlanID, "test.pdf", "/path/to/file.pdf", "application/pdf", 1024000)
	}
}

// Package entity provides tests for branch domain entities.
//
// Tests cover:
// - Branch creation (main and child branches)
// - Status transitions
// - Snapshot creation
// - Version management
package entity

import (
	"testing"

	"github.com/google/uuid"
)

// TestNewBranch_Main verifies main branch creation.
func TestNewBranch_Main(t *testing.T) {
	t.Parallel()

	sceneID := uuid.New()
	name := "main"

	branch := NewBranch(sceneID, name, nil)

	if branch == nil {
		t.Fatal("expected non-nil branch")
	}

	if branch.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}

	if branch.SceneID != sceneID {
		t.Errorf("expected scene_id %s, got %s", sceneID, branch.SceneID)
	}

	if branch.Name != name {
		t.Errorf("expected name '%s', got '%s'", name, branch.Name)
	}

	if branch.ParentBranchID != nil {
		t.Error("expected nil parent_branch_id for main branch")
	}

	if !branch.IsMain {
		t.Error("expected is_main to be true for main branch")
	}

	if branch.Status != BranchStatusActive {
		t.Errorf("expected status %s, got %s", BranchStatusActive, branch.Status)
	}

	if branch.Version != 1 {
		t.Errorf("expected version 1, got %d", branch.Version)
	}

	if branch.CreatedAt.IsZero() {
		t.Error("expected non-zero created_at")
	}
}

// TestNewBranch_Child verifies child branch creation.
func TestNewBranch_Child(t *testing.T) {
	t.Parallel()

	sceneID := uuid.New()
	parentID := uuid.New()
	name := "AI Вариант #1"

	branch := NewBranch(sceneID, name, &parentID)

	if branch == nil {
		t.Fatal("expected non-nil branch")
	}

	if branch.ParentBranchID == nil {
		t.Fatal("expected non-nil parent_branch_id for child branch")
	}

	if *branch.ParentBranchID != parentID {
		t.Errorf("expected parent_branch_id %s, got %s", parentID, *branch.ParentBranchID)
	}

	if branch.IsMain {
		t.Error("expected is_main to be false for child branch")
	}
}

// TestBranchStatus verifies branch status values.
func TestBranchStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status BranchStatus
		str    string
	}{
		{BranchStatusActive, "ACTIVE"},
		{BranchStatusMerged, "MERGED"},
		{BranchStatusArchived, "ARCHIVED"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if string(tt.status) != tt.str {
				t.Errorf("expected '%s', got '%s'", tt.str, string(tt.status))
			}
		})
	}
}

// TestNewSnapshot verifies snapshot creation.
func TestNewSnapshot(t *testing.T) {
	t.Parallel()

	branchID := uuid.New()
	name := "Сохранение перед изменениями"
	version := 5
	data := []byte(`{"elements":[{"id":"elem-1","type":"WALL"}]}`)

	snapshot := NewSnapshot(branchID, name, version, data)

	if snapshot == nil {
		t.Fatal("expected non-nil snapshot")
	}

	if snapshot.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}

	if snapshot.BranchID != branchID {
		t.Errorf("expected branch_id %s, got %s", branchID, snapshot.BranchID)
	}

	if snapshot.Name != name {
		t.Errorf("expected name '%s', got '%s'", name, snapshot.Name)
	}

	if snapshot.Version != version {
		t.Errorf("expected version %d, got %d", version, snapshot.Version)
	}

	if len(snapshot.Data) != len(data) {
		t.Errorf("expected data length %d, got %d", len(data), len(snapshot.Data))
	}

	if snapshot.CreatedAt.IsZero() {
		t.Error("expected non-zero created_at")
	}
}

// TestSnapshot_DataIntegrity verifies snapshot data is preserved.
func TestSnapshot_DataIntegrity(t *testing.T) {
	t.Parallel()

	originalData := []byte(`{"elements":[{"id":"wall-1","type":"WALL","properties":{"is_load_bearing":true}}]}`)

	snapshot := NewSnapshot(uuid.New(), "test", 1, originalData)

	// Verify data is exactly the same
	if string(snapshot.Data) != string(originalData) {
		t.Error("snapshot data does not match original")
	}
}

// TestBranch_Hierarchy verifies branch hierarchy creation.
func TestBranch_Hierarchy(t *testing.T) {
	t.Parallel()

	sceneID := uuid.New()

	// Create main branch
	main := NewBranch(sceneID, "main", nil)
	if !main.IsMain {
		t.Fatal("expected main branch")
	}

	// Create child branch
	child1 := NewBranch(sceneID, "AI вариант 1", &main.ID)
	if child1.IsMain {
		t.Fatal("child1 should not be main")
	}
	if *child1.ParentBranchID != main.ID {
		t.Errorf("child1 parent should be main")
	}

	// Create grandchild branch
	grandchild := NewBranch(sceneID, "Модификация варианта 1", &child1.ID)
	if grandchild.IsMain {
		t.Fatal("grandchild should not be main")
	}
	if *grandchild.ParentBranchID != child1.ID {
		t.Errorf("grandchild parent should be child1")
	}
}

// TestBranch_StatusTransition verifies status changes.
func TestBranch_StatusTransition(t *testing.T) {
	t.Parallel()

	branch := NewBranch(uuid.New(), "test", nil)

	// Initial status
	if branch.Status != BranchStatusActive {
		t.Errorf("initial status should be ACTIVE, got %s", branch.Status)
	}

	// Transition to merged
	branch.Status = BranchStatusMerged
	if branch.Status != BranchStatusMerged {
		t.Errorf("status should be MERGED, got %s", branch.Status)
	}

	// Transition to archived
	branch.Status = BranchStatusArchived
	if branch.Status != BranchStatusArchived {
		t.Errorf("status should be ARCHIVED, got %s", branch.Status)
	}
}

// TestBranch_Description verifies description field.
func TestBranch_Description(t *testing.T) {
	t.Parallel()

	branch := NewBranch(uuid.New(), "test", nil)
	description := "Вариант планировки с объединённой кухней-гостиной"

	branch.Description = description

	if branch.Description != description {
		t.Errorf("expected description '%s', got '%s'", description, branch.Description)
	}
}

// TestSnapshot_EmptyData verifies snapshot with empty data.
func TestSnapshot_EmptyData(t *testing.T) {
	t.Parallel()

	snapshot := NewSnapshot(uuid.New(), "empty", 0, nil)

	if snapshot.Data != nil {
		t.Error("expected nil data")
	}
}

// BenchmarkNewBranch benchmarks branch creation.
func BenchmarkNewBranch(b *testing.B) {
	sceneID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewBranch(sceneID, "Test Branch", nil)
	}
}

// BenchmarkNewBranch_Child benchmarks child branch creation.
func BenchmarkNewBranch_Child(b *testing.B) {
	sceneID := uuid.New()
	parentID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewBranch(sceneID, "Child Branch", &parentID)
	}
}

// BenchmarkNewSnapshot benchmarks snapshot creation.
func BenchmarkNewSnapshot(b *testing.B) {
	branchID := uuid.New()
	data := make([]byte, 10240) // 10KB data

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewSnapshot(branchID, "Snapshot", 1, data)
	}
}


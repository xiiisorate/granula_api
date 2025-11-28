// Package entity provides tests for scene domain entities.
//
// Tests cover:
// - Scene creation and management
// - Element creation (walls, rooms, furniture)
// - Status transitions and versioning
// - Property handling
package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestNewScene verifies scene creation.
func TestNewScene(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	ownerID := uuid.New()
	name := "Квартира-студия"

	scene := NewScene(workspaceID, ownerID, name)

	if scene == nil {
		t.Fatal("expected non-nil scene")
	}

	if scene.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}

	if scene.WorkspaceID != workspaceID {
		t.Errorf("expected workspace_id %s, got %s", workspaceID, scene.WorkspaceID)
	}

	if scene.OwnerID != ownerID {
		t.Errorf("expected owner_id %s, got %s", ownerID, scene.OwnerID)
	}

	if scene.Name != name {
		t.Errorf("expected name '%s', got '%s'", name, scene.Name)
	}

	if scene.MainBranchID == uuid.Nil {
		t.Error("expected non-nil main_branch_id")
	}

	if scene.Metadata == nil {
		t.Error("expected initialized metadata map")
	}

	if scene.CreatedAt.IsZero() {
		t.Error("expected non-zero created_at")
	}
}

// TestNewElement verifies element creation.
func TestNewElement(t *testing.T) {
	t.Parallel()

	sceneID := uuid.New()
	branchID := uuid.New()
	elemType := ElementTypeWall
	name := "North Wall"

	elem := NewElement(sceneID, branchID, elemType, name)

	if elem == nil {
		t.Fatal("expected non-nil element")
	}

	if elem.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}

	if elem.SceneID != sceneID {
		t.Errorf("expected scene_id %s, got %s", sceneID, elem.SceneID)
	}

	if elem.BranchID != branchID {
		t.Errorf("expected branch_id %s, got %s", branchID, elem.BranchID)
	}

	if elem.Type != elemType {
		t.Errorf("expected type %s, got %s", elemType, elem.Type)
	}

	if elem.Name != name {
		t.Errorf("expected name '%s', got '%s'", name, elem.Name)
	}

	if elem.Version != 1 {
		t.Errorf("expected version 1, got %d", elem.Version)
	}

	if elem.IsDeleted {
		t.Error("expected is_deleted to be false")
	}
}

// TestNewWall verifies wall element creation.
func TestNewWall(t *testing.T) {
	t.Parallel()

	sceneID := uuid.New()
	branchID := uuid.New()
	name := "Load Bearing Wall"
	start := Point2D{X: 0, Y: 0}
	end := Point2D{X: 5, Y: 0}
	thickness := 0.3
	isLoadBearing := true

	wall := NewWall(sceneID, branchID, name, start, end, thickness, isLoadBearing)

	if wall == nil {
		t.Fatal("expected non-nil wall")
	}

	if wall.Type != ElementTypeWall {
		t.Errorf("expected type %s, got %s", ElementTypeWall, wall.Type)
	}

	if wall.Properties.IsLoadBearing != isLoadBearing {
		t.Errorf("expected is_load_bearing %v, got %v", isLoadBearing, wall.Properties.IsLoadBearing)
	}

	if wall.Properties.Thickness != thickness {
		t.Errorf("expected thickness %f, got %f", thickness, wall.Properties.Thickness)
	}

	if wall.Properties.StartPoint == nil {
		t.Fatal("expected non-nil start_point")
	}

	if wall.Properties.StartPoint.X != start.X || wall.Properties.StartPoint.Y != start.Y {
		t.Errorf("expected start_point (%f, %f), got (%f, %f)",
			start.X, start.Y, wall.Properties.StartPoint.X, wall.Properties.StartPoint.Y)
	}

	if wall.Properties.EndPoint == nil {
		t.Fatal("expected non-nil end_point")
	}

	if wall.Properties.EndPoint.X != end.X || wall.Properties.EndPoint.Y != end.Y {
		t.Errorf("expected end_point (%f, %f), got (%f, %f)",
			end.X, end.Y, wall.Properties.EndPoint.X, wall.Properties.EndPoint.Y)
	}
}

// TestNewRoom verifies room element creation.
func TestNewRoom(t *testing.T) {
	t.Parallel()

	sceneID := uuid.New()
	branchID := uuid.New()
	name := "Гостиная"
	roomType := "LIVING"
	boundary := []Point2D{
		{X: 0, Y: 0},
		{X: 5, Y: 0},
		{X: 5, Y: 4},
		{X: 0, Y: 4},
	}
	isWetZone := false

	room := NewRoom(sceneID, branchID, name, roomType, boundary, isWetZone)

	if room == nil {
		t.Fatal("expected non-nil room")
	}

	if room.Type != ElementTypeRoom {
		t.Errorf("expected type %s, got %s", ElementTypeRoom, room.Type)
	}

	if room.Properties.RoomType != roomType {
		t.Errorf("expected room_type '%s', got '%s'", roomType, room.Properties.RoomType)
	}

	if len(room.Properties.Boundary) != len(boundary) {
		t.Errorf("expected %d boundary points, got %d", len(boundary), len(room.Properties.Boundary))
	}

	if room.Properties.IsWetZone != isWetZone {
		t.Errorf("expected is_wet_zone %v, got %v", isWetZone, room.Properties.IsWetZone)
	}
}

// TestNewFurniture verifies furniture element creation.
func TestNewFurniture(t *testing.T) {
	t.Parallel()

	sceneID := uuid.New()
	branchID := uuid.New()
	name := "Диван IKEA"
	catalogID := "ikea-sofa-123"
	position := Point3D{X: 2.5, Y: 1.5, Z: 0}
	dimensions := Dimensions3D{Width: 2.0, Depth: 0.9, Height: 0.8}

	furniture := NewFurniture(sceneID, branchID, name, catalogID, position, dimensions)

	if furniture == nil {
		t.Fatal("expected non-nil furniture")
	}

	if furniture.Type != ElementTypeFurniture {
		t.Errorf("expected type %s, got %s", ElementTypeFurniture, furniture.Type)
	}

	if furniture.Properties.CatalogID != catalogID {
		t.Errorf("expected catalog_id '%s', got '%s'", catalogID, furniture.Properties.CatalogID)
	}

	if furniture.Position.X != position.X || furniture.Position.Y != position.Y || furniture.Position.Z != position.Z {
		t.Errorf("expected position (%f, %f, %f), got (%f, %f, %f)",
			position.X, position.Y, position.Z,
			furniture.Position.X, furniture.Position.Y, furniture.Position.Z)
	}

	if furniture.Dimensions.Width != dimensions.Width {
		t.Errorf("expected width %f, got %f", dimensions.Width, furniture.Dimensions.Width)
	}
}

// TestElement_Update verifies element update and version increment.
func TestElement_Update(t *testing.T) {
	t.Parallel()

	elem := NewElement(uuid.New(), uuid.New(), ElementTypeWall, "Test")
	originalVersion := elem.Version
	originalUpdatedAt := elem.UpdatedAt

	time.Sleep(1 * time.Millisecond) // Ensure time difference
	elem.Update()

	if elem.Version != originalVersion+1 {
		t.Errorf("expected version %d, got %d", originalVersion+1, elem.Version)
	}

	if !elem.UpdatedAt.After(originalUpdatedAt) {
		t.Error("expected updated_at to be updated")
	}
}

// TestElement_SoftDelete verifies soft deletion.
func TestElement_SoftDelete(t *testing.T) {
	t.Parallel()

	elem := NewElement(uuid.New(), uuid.New(), ElementTypeRoom, "Test Room")
	originalVersion := elem.Version

	elem.SoftDelete()

	if !elem.IsDeleted {
		t.Error("expected is_deleted to be true")
	}

	if elem.Version != originalVersion+1 {
		t.Errorf("expected version to increment, got %d", elem.Version)
	}
}

// TestElementTypeStrings verifies element type string values.
func TestElementTypeStrings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		elemType ElementType
		str      string
	}{
		{ElementTypeWall, "WALL"},
		{ElementTypeRoom, "ROOM"},
		{ElementTypeDoor, "DOOR"},
		{ElementTypeWindow, "WINDOW"},
		{ElementTypeFurniture, "FURNITURE"},
		{ElementTypeFixture, "FIXTURE"},
		{ElementTypeDecor, "DECOR"},
	}

	for _, tt := range tests {
		t.Run(string(tt.elemType), func(t *testing.T) {
			if string(tt.elemType) != tt.str {
				t.Errorf("expected '%s', got '%s'", tt.str, string(tt.elemType))
			}
		})
	}
}

// TestDimensions3D verifies 3D dimensions.
func TestDimensions3D(t *testing.T) {
	t.Parallel()

	dim := Dimensions3D{
		Width:  5.0,
		Depth:  4.0,
		Height: 2.7,
	}

	if dim.Width != 5.0 {
		t.Errorf("expected width 5.0, got %f", dim.Width)
	}

	if dim.Depth != 4.0 {
		t.Errorf("expected depth 4.0, got %f", dim.Depth)
	}

	if dim.Height != 2.7 {
		t.Errorf("expected height 2.7, got %f", dim.Height)
	}
}

// TestPoint3D verifies 3D point.
func TestPoint3D(t *testing.T) {
	t.Parallel()

	point := Point3D{X: 1.5, Y: 2.5, Z: 0.5}

	if point.X != 1.5 || point.Y != 2.5 || point.Z != 0.5 {
		t.Errorf("expected (1.5, 2.5, 0.5), got (%f, %f, %f)", point.X, point.Y, point.Z)
	}
}

// TestPoint2D verifies 2D point.
func TestPoint2D(t *testing.T) {
	t.Parallel()

	point := Point2D{X: 3.0, Y: 4.0}

	if point.X != 3.0 || point.Y != 4.0 {
		t.Errorf("expected (3.0, 4.0), got (%f, %f)", point.X, point.Y)
	}
}

// TestScene_Metadata verifies metadata operations.
func TestScene_Metadata(t *testing.T) {
	t.Parallel()

	scene := NewScene(uuid.New(), uuid.New(), "Test")

	// Metadata should be initialized
	if scene.Metadata == nil {
		t.Fatal("expected non-nil metadata")
	}

	// Add metadata
	scene.Metadata["source"] = "ai_generated"
	scene.Metadata["accuracy"] = "high"

	if scene.Metadata["source"] != "ai_generated" {
		t.Errorf("expected metadata source='ai_generated', got '%s'", scene.Metadata["source"])
	}
}

// BenchmarkNewScene benchmarks scene creation.
func BenchmarkNewScene(b *testing.B) {
	workspaceID := uuid.New()
	ownerID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewScene(workspaceID, ownerID, "Test Scene")
	}
}

// BenchmarkNewWall benchmarks wall creation.
func BenchmarkNewWall(b *testing.B) {
	sceneID := uuid.New()
	branchID := uuid.New()
	start := Point2D{X: 0, Y: 0}
	end := Point2D{X: 5, Y: 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewWall(sceneID, branchID, "Wall", start, end, 0.2, false)
	}
}

// BenchmarkNewRoom benchmarks room creation.
func BenchmarkNewRoom(b *testing.B) {
	sceneID := uuid.New()
	branchID := uuid.New()
	boundary := []Point2D{
		{X: 0, Y: 0}, {X: 5, Y: 0}, {X: 5, Y: 4}, {X: 0, Y: 4},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewRoom(sceneID, branchID, "Room", "LIVING", boundary, false)
	}
}

// BenchmarkElementUpdate benchmarks element updates.
func BenchmarkElementUpdate(b *testing.B) {
	elem := NewElement(uuid.New(), uuid.New(), ElementTypeWall, "Test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		elem.Update()
	}
}

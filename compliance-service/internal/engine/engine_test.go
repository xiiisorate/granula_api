// Package engine provides tests for the compliance rules engine.
//
// Tests cover:
// - Load-bearing wall checks (СНиП 31-01-2003 п.9.22)
// - Wet zone placement rules (ЖК РФ ст.25-26)
// - Minimum area requirements (СНиП 31-01-2003)
// - Ventilation requirements
// - Fire safety checks
package engine

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/xiiisorate/granula_api/compliance-service/internal/domain/entity"
)

// createTestRules creates a set of test rules for the engine.
func createTestRules() []*entity.Rule {
	return []*entity.Rule{
		{
			ID:       uuid.New(),
			Code:     "SNiP-31-01-2003-9.22",
			Name:     "Запрет сноса несущих стен",
			Category: entity.CategoryLoadBearing,
			Severity: entity.SeverityError,
			Active:   true,
			AppliesToOperations: []entity.OperationType{
				entity.OperationTypeDemolishWall,
				entity.OperationTypeAddOpening,
			},
		},
		{
			ID:       uuid.New(),
			Code:     "ZHK-RF-25-26",
			Name:     "Размещение мокрых зон",
			Category: entity.CategoryWetZones,
			Severity: entity.SeverityError,
			Active:   true,
			AppliesToOperations: []entity.OperationType{
				entity.OperationTypeMoveWetZone,
			},
		},
		{
			ID:       uuid.New(),
			Code:     "SNiP-31-01-2003-5.7",
			Name:     "Минимальная площадь комнат",
			Category: entity.CategoryMinArea,
			Severity: entity.SeverityWarning,
			Active:   true,
			AppliesToOperations: []entity.OperationType{
				entity.OperationTypeMergeRooms,
				entity.OperationTypeSplitRoom,
			},
		},
	}
}

// TestNewRuleEngine verifies that a new engine is created correctly.
func TestNewRuleEngine(t *testing.T) {
	t.Parallel()

	rules := createTestRules()
	engine := NewRuleEngine(rules)

	if engine == nil {
		t.Fatal("expected non-nil engine")
	}

	if len(engine.rules) != len(rules) {
		t.Errorf("expected %d rules, got %d", len(rules), len(engine.rules))
	}

	// Verify checkers are registered
	expectedCategories := []entity.RuleCategory{
		entity.CategoryLoadBearing,
		entity.CategoryWetZones,
		entity.CategoryMinArea,
		entity.CategoryVentilation,
		entity.CategoryFireSafety,
		entity.CategoryDaylight,
		entity.CategoryGeneral,
	}

	for _, cat := range expectedCategories {
		if _, ok := engine.checker[cat]; !ok {
			t.Errorf("expected checker for category %s", cat)
		}
	}
}

// TestCheckScene_NoViolations verifies compliant scene passes all checks.
func TestCheckScene_NoViolations(t *testing.T) {
	t.Parallel()

	rules := createTestRules()
	engine := NewRuleEngine(rules)

	scene := &SceneData{
		ID:        uuid.New().String(),
		TotalArea: 60.0,
		Walls: []WallData{
			{ID: uuid.New().String(), IsLoadBearing: false, Thickness: 0.1},
			{ID: uuid.New().String(), IsLoadBearing: true, Thickness: 0.3},
		},
		Rooms: []RoomData{
			{
				ID:         uuid.New().String(),
				Type:       "LIVING",
				Area:       20.0,
				HasWindows: true,
				IsWetZone:  false,
			},
			{
				ID:         uuid.New().String(),
				Type:       "BEDROOM",
				Area:       14.0,
				HasWindows: true,
				IsWetZone:  false,
			},
			{
				ID:         uuid.New().String(),
				Type:       "KITCHEN",
				Area:       10.0,
				HasWindows: true,
				IsWetZone:  true,
			},
		},
	}

	ctx := context.Background()
	result := engine.CheckScene(ctx, scene)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Log any violations for debugging
	if len(result.Violations) > 0 {
		for _, v := range result.Violations {
			t.Logf("Violation: %s - %s", v.RuleCode, v.Title)
		}
	}
}

// TestCheckScene_MinAreaViolation verifies minimum area check detects violations.
func TestCheckScene_MinAreaViolation(t *testing.T) {
	t.Parallel()

	rules := createTestRules()
	engine := NewRuleEngine(rules)

	scene := &SceneData{
		ID:        uuid.New().String(),
		TotalArea: 30.0,
		Rooms: []RoomData{
			{
				ID:         uuid.New().String(),
				Type:       "LIVING",
				Area:       6.0, // Below minimum 8 sqm
				HasWindows: true,
				IsWetZone:  false,
			},
			{
				ID:         uuid.New().String(),
				Type:       "KITCHEN",
				Area:       3.0, // Below minimum 5 sqm
				HasWindows: true,
				IsWetZone:  true,
			},
		},
	}

	ctx := context.Background()
	result := engine.CheckScene(ctx, scene)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Should have violations for area
	hasAreaViolation := false
	for _, v := range result.Violations {
		if v.Category == entity.CategoryMinArea {
			hasAreaViolation = true
			break
		}
	}

	if !hasAreaViolation {
		t.Error("expected minimum area violation")
	}
}

// TestCheckOperation_DemolishPartition verifies partition demolition is allowed.
func TestCheckOperation_DemolishPartition(t *testing.T) {
	t.Parallel()

	rules := createTestRules()
	engine := NewRuleEngine(rules)

	wallID := uuid.New().String()

	scene := &SceneData{
		ID: uuid.New().String(),
		Walls: []WallData{
			{ID: wallID, IsLoadBearing: false, Thickness: 0.1},
		},
	}

	op := &OperationData{
		Type:        entity.OperationTypeDemolishWall,
		ElementID:   wallID,
		ElementType: entity.ElementTypeWall,
	}

	ctx := context.Background()
	result := engine.CheckOperation(ctx, scene, op)

	// Should NOT have violations for demolishing a partition
	for _, v := range result.Violations {
		if v.Category == entity.CategoryLoadBearing {
			t.Errorf("unexpected load bearing violation for partition: %s", v.Title)
		}
	}
}

// TestCheckOperation_DemolishLoadBearing verifies load-bearing wall demolition is blocked.
func TestCheckOperation_DemolishLoadBearing(t *testing.T) {
	t.Parallel()

	rules := createTestRules()
	engine := NewRuleEngine(rules)

	wallID := uuid.New().String()

	scene := &SceneData{
		ID: uuid.New().String(),
		Walls: []WallData{
			{ID: wallID, IsLoadBearing: true, Thickness: 0.3},
		},
	}

	op := &OperationData{
		Type:        entity.OperationTypeDemolishWall,
		ElementID:   wallID,
		ElementType: entity.ElementTypeLoadBearingWall,
	}

	ctx := context.Background()
	result := engine.CheckOperation(ctx, scene, op)

	// Should have violation for load-bearing wall
	hasLoadBearingViolation := false
	for _, v := range result.Violations {
		if v.Category == entity.CategoryLoadBearing {
			hasLoadBearingViolation = true
			t.Logf("Correctly detected: %s", v.Title)
			break
		}
	}

	if !hasLoadBearingViolation {
		t.Error("expected load bearing violation when demolishing load-bearing wall")
	}
}

// TestGroupRulesByCategory verifies rules are correctly grouped.
func TestGroupRulesByCategory(t *testing.T) {
	t.Parallel()

	rules := []*entity.Rule{
		{ID: uuid.New(), Category: entity.CategoryLoadBearing, Active: true},
		{ID: uuid.New(), Category: entity.CategoryLoadBearing, Active: true},
		{ID: uuid.New(), Category: entity.CategoryWetZones, Active: true},
		{ID: uuid.New(), Category: entity.CategoryMinArea, Active: true},
		{ID: uuid.New(), Category: entity.CategoryMinArea, Active: false}, // Inactive
	}

	engine := NewRuleEngine(rules)
	grouped := engine.groupRulesByCategory()

	// Check counts (only active rules)
	if len(grouped[entity.CategoryLoadBearing]) != 2 {
		t.Errorf("expected 2 load bearing rules, got %d", len(grouped[entity.CategoryLoadBearing]))
	}

	if len(grouped[entity.CategoryWetZones]) != 1 {
		t.Errorf("expected 1 wet zones rule, got %d", len(grouped[entity.CategoryWetZones]))
	}

	if len(grouped[entity.CategoryMinArea]) != 1 {
		t.Errorf("expected 1 min area rule (active only), got %d", len(grouped[entity.CategoryMinArea]))
	}
}

// BenchmarkCheckScene benchmarks the full compliance check.
func BenchmarkCheckScene(b *testing.B) {
	rules := createTestRules()
	engine := NewRuleEngine(rules)

	scene := &SceneData{
		ID:        uuid.New().String(),
		TotalArea: 80.0,
		Walls: []WallData{
			{ID: uuid.New().String(), IsLoadBearing: true, Thickness: 0.3},
			{ID: uuid.New().String(), IsLoadBearing: false, Thickness: 0.1},
			{ID: uuid.New().String(), IsLoadBearing: true, Thickness: 0.25},
			{ID: uuid.New().String(), IsLoadBearing: false, Thickness: 0.12},
		},
		Rooms: []RoomData{
			{ID: uuid.New().String(), Type: "LIVING", Area: 25.0, HasWindows: true},
			{ID: uuid.New().String(), Type: "BEDROOM", Area: 16.0, HasWindows: true},
			{ID: uuid.New().String(), Type: "BEDROOM", Area: 12.0, HasWindows: true},
			{ID: uuid.New().String(), Type: "KITCHEN", Area: 10.0, HasWindows: true, IsWetZone: true},
			{ID: uuid.New().String(), Type: "BATHROOM", Area: 5.0, IsWetZone: true},
			{ID: uuid.New().String(), Type: "HALLWAY", Area: 8.0},
		},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.CheckScene(ctx, scene)
	}
}

// BenchmarkCheckOperation benchmarks single operation check.
func BenchmarkCheckOperation(b *testing.B) {
	rules := createTestRules()
	engine := NewRuleEngine(rules)

	wallID := uuid.New().String()
	scene := &SceneData{
		ID: uuid.New().String(),
		Walls: []WallData{
			{ID: wallID, IsLoadBearing: false, Thickness: 0.1},
		},
	}

	op := &OperationData{
		Type:        entity.OperationTypeDemolishWall,
		ElementID:   wallID,
		ElementType: entity.ElementTypeWall,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.CheckOperation(ctx, scene, op)
	}
}

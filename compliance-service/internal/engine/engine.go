// Package engine provides the compliance rule checking engine.
package engine

import (
	"context"
	"fmt"

	"github.com/xiiisorate/granula_api/compliance-service/internal/domain/entity"
)

// SceneData represents the scene data for compliance checking.
// This is a simplified representation - actual data comes from Scene Service.
type SceneData struct {
	ID        string            `json:"id"`
	Rooms     []RoomData        `json:"rooms"`
	Walls     []WallData        `json:"walls"`
	Equipment []EquipmentData   `json:"equipment"`
	TotalArea float64           `json:"total_area"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// RoomData represents room information for checking.
type RoomData struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Area         float64        `json:"area"`
	HasWindows   bool           `json:"has_windows"`
	IsWetZone    bool           `json:"is_wet_zone"`
	WallIDs      []string       `json:"wall_ids"`
	EquipmentIDs []string       `json:"equipment_ids"`
	Position     entity.Point2D `json:"position"`
}

// WallData represents wall information for checking.
type WallData struct {
	ID            string  `json:"id"`
	IsLoadBearing bool    `json:"is_load_bearing"`
	Thickness     float64 `json:"thickness"`
	HasOpenings   bool    `json:"has_openings"`
}

// EquipmentData represents equipment information for checking.
type EquipmentData struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	RoomID string `json:"room_id"`
}

// OperationData represents an operation to be checked.
type OperationData struct {
	Type        entity.OperationType `json:"type"`
	ElementID   string               `json:"element_id"`
	ElementType entity.ElementType   `json:"element_type"`
	Params      map[string]string    `json:"params"`
	NewPosition *entity.Point2D      `json:"new_position,omitempty"`
}

// RuleEngine checks compliance rules against scene data.
type RuleEngine struct {
	rules   []*entity.Rule
	checker map[entity.RuleCategory]CategoryChecker
}

// CategoryChecker is the interface for category-specific rule checking.
type CategoryChecker interface {
	Check(ctx context.Context, scene *SceneData, rules []*entity.Rule) []*entity.Violation
	CheckOperation(ctx context.Context, scene *SceneData, op *OperationData, rules []*entity.Rule) []*entity.Violation
}

// NewRuleEngine creates a new rule engine with the given rules.
func NewRuleEngine(rules []*entity.Rule) *RuleEngine {
	engine := &RuleEngine{
		rules:   rules,
		checker: make(map[entity.RuleCategory]CategoryChecker),
	}

	// Register category checkers
	engine.checker[entity.CategoryLoadBearing] = &LoadBearingChecker{}
	engine.checker[entity.CategoryWetZones] = &WetZoneChecker{}
	engine.checker[entity.CategoryMinArea] = &MinAreaChecker{}
	engine.checker[entity.CategoryVentilation] = &VentilationChecker{}
	engine.checker[entity.CategoryFireSafety] = &FireSafetyChecker{}
	engine.checker[entity.CategoryDaylight] = &DaylightChecker{}
	engine.checker[entity.CategoryGeneral] = &GeneralChecker{}

	return engine
}

// CheckScene performs a full compliance check on a scene.
func (e *RuleEngine) CheckScene(ctx context.Context, scene *SceneData) *entity.ComplianceResult {
	result := entity.NewComplianceResult()

	// Group rules by category
	rulesByCategory := e.groupRulesByCategory()

	// Check each category
	for category, rules := range rulesByCategory {
		checker, ok := e.checker[category]
		if !ok {
			continue
		}

		violations := checker.Check(ctx, scene, rules)
		for _, v := range violations {
			result.AddViolation(v)
		}
	}

	result.Finalize(len(e.rules))
	return result
}

// CheckOperation checks if an operation is allowed.
func (e *RuleEngine) CheckOperation(ctx context.Context, scene *SceneData, op *OperationData) *entity.ComplianceResult {
	result := entity.NewComplianceResult()

	// Get rules that apply to this operation type
	applicableRules := e.getRulesForOperation(op.Type)

	// Group by category and check
	rulesByCategory := make(map[entity.RuleCategory][]*entity.Rule)
	for _, rule := range applicableRules {
		rulesByCategory[rule.Category] = append(rulesByCategory[rule.Category], rule)
	}

	for category, rules := range rulesByCategory {
		checker, ok := e.checker[category]
		if !ok {
			continue
		}

		violations := checker.CheckOperation(ctx, scene, op, rules)
		for _, v := range violations {
			result.AddViolation(v)
		}
	}

	result.Finalize(len(applicableRules))
	return result
}

// groupRulesByCategory groups rules by their category.
func (e *RuleEngine) groupRulesByCategory() map[entity.RuleCategory][]*entity.Rule {
	grouped := make(map[entity.RuleCategory][]*entity.Rule)
	for _, rule := range e.rules {
		if rule.Active {
			grouped[rule.Category] = append(grouped[rule.Category], rule)
		}
	}
	return grouped
}

// getRulesForOperation returns rules that apply to the given operation.
func (e *RuleEngine) getRulesForOperation(opType entity.OperationType) []*entity.Rule {
	result := make([]*entity.Rule, 0)
	for _, rule := range e.rules {
		if !rule.Active {
			continue
		}
		for _, op := range rule.AppliesToOperations {
			if op == opType {
				result = append(result, rule)
				break
			}
		}
	}
	return result
}

// =============================================================================
// Category Checkers
// =============================================================================

// LoadBearingChecker checks rules related to load-bearing structures.
type LoadBearingChecker struct{}

// Check performs checks for load-bearing rules.
func (c *LoadBearingChecker) Check(ctx context.Context, scene *SceneData, rules []*entity.Rule) []*entity.Violation {
	// Full scene check doesn't check operations, just structure
	return nil
}

// CheckOperation checks if operation is allowed on load-bearing elements.
func (c *LoadBearingChecker) CheckOperation(ctx context.Context, scene *SceneData, op *OperationData, rules []*entity.Rule) []*entity.Violation {
	violations := make([]*entity.Violation, 0)

	// Find the wall being operated on
	var wall *WallData
	for i := range scene.Walls {
		if scene.Walls[i].ID == op.ElementID {
			wall = &scene.Walls[i]
			break
		}
	}

	if wall == nil || !wall.IsLoadBearing {
		return violations
	}

	// Check rules for this operation on load-bearing wall
	for _, rule := range rules {
		if op.Type == entity.OperationTypeDemolishWall {
			v := entity.NewViolation(rule, op.ElementID,
				"Снос несущей стены запрещён. Несущие стены обеспечивают устойчивость здания.")
			v.WithElementType(entity.ElementTypeLoadBearingWall)
			v.WithSuggestion("Рассмотрите устройство проёма вместо полного сноса (требуется проект и экспертиза)")
			violations = append(violations, v)
		}

		if op.Type == entity.OperationTypeAddOpening {
			maxWidth := rule.GetFloatParameter("max_opening_width", 0.9)
			v := entity.NewViolation(rule, op.ElementID,
				fmt.Sprintf("Устройство проёма в несущей стене требует проекта. Максимальная ширина: %.1f м.", maxWidth))
			v.WithElementType(entity.ElementTypeLoadBearingWall)
			v.WithSuggestion("Закажите проект усиления от сертифицированного проектировщика")
			violations = append(violations, v)
		}
	}

	return violations
}

// WetZoneChecker checks rules related to wet zones.
type WetZoneChecker struct{}

// Check checks wet zone rules.
func (c *WetZoneChecker) Check(ctx context.Context, scene *SceneData, rules []*entity.Rule) []*entity.Violation {
	violations := make([]*entity.Violation, 0)

	for _, room := range scene.Rooms {
		if !room.IsWetZone {
			continue
		}

		for _, rule := range rules {
			// Check if wet zone is over living area (simplified check)
			if rule.Code == "ZHK-RF-25" {
				// In real implementation, would check floor below
				// This is a placeholder for the actual check
			}
		}
	}

	return violations
}

// CheckOperation checks wet zone operation rules.
func (c *WetZoneChecker) CheckOperation(ctx context.Context, scene *SceneData, op *OperationData, rules []*entity.Rule) []*entity.Violation {
	violations := make([]*entity.Violation, 0)

	if op.Type == entity.OperationTypeMoveWetZone || op.Type == entity.OperationTypeExpandWetZone {
		for _, rule := range rules {
			if rule.Code == "ZHK-RF-25" {
				v := entity.NewViolation(rule, op.ElementID,
					"Перемещение санузла может привести к его расположению над жилыми помещениями соседей.")
				v.WithElementType(op.ElementType)
				v.WithSuggestion("Уточните расположение помещений на нижнем этаже в БТИ")
				violations = append(violations, v)
			}

			if rule.Code == "SP-54.13330-5.8" && op.Type == entity.OperationTypeExpandWetZone {
				v := entity.NewViolation(rule, op.ElementID,
					"При расширении мокрой зоны требуется устройство гидроизоляции.")
				v.WithElementType(op.ElementType)
				v.WithSuggestion("Включите в проект работы по гидроизоляции пола и стен на высоту 20 см")
				violations = append(violations, v)
			}
		}
	}

	return violations
}

// MinAreaChecker checks minimum area rules.
type MinAreaChecker struct{}

// minAreaRequirements defines minimum area requirements by room type (in sqm).
// Based on СП 54.13330.2016 and СНиП 31-01-2003.
var minAreaRequirements = map[string]float64{
	"LIVING":   14.0, // Жилая комната (единственная) - 14 кв.м
	"BEDROOM":  8.0,  // Спальня - 8 кв.м
	"KITCHEN":  5.0,  // Кухня - 5 кв.м
	"BATHROOM": 3.8,  // Совмещённый санузел - 3.8 кв.м
	"TOILET":   1.2,  // Отдельный туалет - 1.2 кв.м
	"HALLWAY":  0.0,  // Коридор - нет минимума
	"BALCONY":  0.0,  // Балкон - нет минимума
}

// Check checks minimum area rules.
func (c *MinAreaChecker) Check(ctx context.Context, scene *SceneData, rules []*entity.Rule) []*entity.Violation {
	violations := make([]*entity.Violation, 0)

	// Skip if no rules or no rooms
	if len(rules) == 0 || len(scene.Rooms) == 0 {
		return violations
	}

	// Use first rule for creating violations (or find specific rule)
	rule := rules[0]

	for _, room := range scene.Rooms {
		minArea, exists := minAreaRequirements[room.Type]
		if !exists || minArea <= 0 {
			continue
		}

		// Also check rule parameters for custom limits
		if ruleMin := rule.GetFloatParameter("min_area_"+room.Type, 0); ruleMin > 0 {
			minArea = ruleMin
		}

		if room.Area < minArea {
			v := entity.NewViolation(rule, room.ID,
				fmt.Sprintf("Площадь помещения '%s' (%.1f кв.м) меньше минимальной (%.1f кв.м)",
					room.Type, room.Area, minArea))
			v.WithElementType(entity.ElementTypeRoom)
			v.WithPosition(room.Position.X, room.Position.Y)
			v.WithSuggestion(fmt.Sprintf("Увеличьте площадь помещения минимум до %.1f кв.м", minArea))
			violations = append(violations, v)
		}
	}

	return violations
}

// CheckOperation checks area constraints for operations.
func (c *MinAreaChecker) CheckOperation(ctx context.Context, scene *SceneData, op *OperationData, rules []*entity.Rule) []*entity.Violation {
	// Operations that might reduce area would be checked here
	return nil
}

// VentilationChecker checks ventilation rules.
type VentilationChecker struct{}

// Check checks ventilation rules.
func (c *VentilationChecker) Check(ctx context.Context, scene *SceneData, rules []*entity.Rule) []*entity.Violation {
	return nil
}

// CheckOperation checks ventilation operation rules.
func (c *VentilationChecker) CheckOperation(ctx context.Context, scene *SceneData, op *OperationData, rules []*entity.Rule) []*entity.Violation {
	violations := make([]*entity.Violation, 0)

	if op.Type == entity.OperationTypeMoveVentilation {
		for _, rule := range rules {
			if rule.Code == "SP-54-VENT-CHANNELS" {
				v := entity.NewViolation(rule, op.ElementID,
					"Перенос или изменение сечения вентиляционных каналов категорически запрещён.")
				v.WithElementType(entity.ElementTypeVentilation)
				v.WithSuggestion("Оставьте вентиляционные каналы без изменений")
				violations = append(violations, v)
			}
		}
	}

	return violations
}

// FireSafetyChecker checks fire safety rules.
type FireSafetyChecker struct{}

// Check checks fire safety rules.
func (c *FireSafetyChecker) Check(ctx context.Context, scene *SceneData, rules []*entity.Rule) []*entity.Violation {
	return nil
}

// CheckOperation checks fire safety for operations.
func (c *FireSafetyChecker) CheckOperation(ctx context.Context, scene *SceneData, op *OperationData, rules []*entity.Rule) []*entity.Violation {
	return nil
}

// DaylightChecker checks daylight requirements.
type DaylightChecker struct{}

// Check checks daylight rules.
func (c *DaylightChecker) Check(ctx context.Context, scene *SceneData, rules []*entity.Rule) []*entity.Violation {
	violations := make([]*entity.Violation, 0)

	for _, room := range scene.Rooms {
		// Living rooms must have windows
		if (room.Type == "LIVING" || room.Type == "BEDROOM") && !room.HasWindows {
			for _, rule := range rules {
				if rule.Code == "SP-54-DAYLIGHT" {
					v := entity.NewViolation(rule, room.ID,
						"Жилая комната должна иметь естественное освещение (окно)")
					v.WithElementType(entity.ElementTypeRoom)
					v.WithPosition(room.Position.X, room.Position.Y)
					v.WithSuggestion("Добавьте окно или измените тип помещения")
					violations = append(violations, v)
				}
			}
		}
	}

	return violations
}

// CheckOperation checks daylight for operations.
func (c *DaylightChecker) CheckOperation(ctx context.Context, scene *SceneData, op *OperationData, rules []*entity.Rule) []*entity.Violation {
	return nil
}

// GeneralChecker checks general planning rules.
type GeneralChecker struct{}

// Check checks general rules.
func (c *GeneralChecker) Check(ctx context.Context, scene *SceneData, rules []*entity.Rule) []*entity.Violation {
	return nil
}

// CheckOperation checks general rules for operations.
func (c *GeneralChecker) CheckOperation(ctx context.Context, scene *SceneData, op *OperationData, rules []*entity.Rule) []*entity.Violation {
	violations := make([]*entity.Violation, 0)

	for _, rule := range rules {
		if rule.Code == "ZHK-RF-26" {
			v := entity.NewViolation(rule, op.ElementID,
				"Данная перепланировка требует уведомления органов местного самоуправления")
			v.WithSuggestion("Подайте уведомление о перепланировке в МФЦ")
			violations = append(violations, v)
		}
	}

	return violations
}

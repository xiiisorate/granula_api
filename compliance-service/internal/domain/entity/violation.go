// Package entity defines domain entities for Compliance Service.
package entity

import (
	"github.com/google/uuid"
)

// Point2D represents a 2D point.
type Point2D struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Violation represents a compliance violation found during checking.
type Violation struct {
	// ID is the unique identifier.
	ID uuid.UUID `json:"id"`

	// RuleID of the violated rule.
	RuleID uuid.UUID `json:"rule_id"`

	// RuleCode for quick reference.
	RuleCode string `json:"rule_code"`

	// Severity of the violation.
	Severity Severity `json:"severity"`

	// Category of the violation.
	Category RuleCategory `json:"category"`

	// Title is a short description.
	Title string `json:"title"`

	// Description is a detailed explanation.
	Description string `json:"description"`

	// ElementID of the element causing the violation.
	ElementID string `json:"element_id,omitempty"`

	// ElementType of the problematic element.
	ElementType ElementType `json:"element_type,omitempty"`

	// Position where the violation occurs.
	Position *Point2D `json:"position,omitempty"`

	// Suggestion for how to fix the violation.
	Suggestion string `json:"suggestion,omitempty"`

	// References to regulatory documents.
	References []DocumentReference `json:"references,omitempty"`

	// ApprovalRequired if the violation can be waived.
	ApprovalRequired ApprovalType `json:"approval_required"`
}

// NewViolation creates a new Violation from a Rule.
func NewViolation(rule *Rule, elementID string, description string) *Violation {
	return &Violation{
		ID:               uuid.New(),
		RuleID:           rule.ID,
		RuleCode:         rule.Code,
		Severity:         rule.Severity,
		Category:         rule.Category,
		Title:            rule.Name,
		Description:      description,
		ElementID:        elementID,
		References:       rule.References,
		ApprovalRequired: rule.ApprovalRequired,
	}
}

// WithElementType sets the element type.
func (v *Violation) WithElementType(t ElementType) *Violation {
	v.ElementType = t
	return v
}

// WithPosition sets the position.
func (v *Violation) WithPosition(x, y float64) *Violation {
	v.Position = &Point2D{X: x, Y: y}
	return v
}

// WithSuggestion sets the suggestion.
func (v *Violation) WithSuggestion(suggestion string) *Violation {
	v.Suggestion = suggestion
	return v
}

// IsBlocking returns true if the violation cannot be approved.
func (v *Violation) IsBlocking() bool {
	return v.ApprovalRequired == ApprovalTypeProhibited || v.Severity == SeverityError
}

// ComplianceResult holds the result of a compliance check.
type ComplianceResult struct {
	// Compliant is true if there are no blocking violations.
	Compliant bool `json:"compliant"`

	// Violations found during the check.
	Violations []*Violation `json:"violations"`

	// Stats about the check.
	Stats ComplianceStats `json:"stats"`

	// RulesVersion used for the check.
	RulesVersion string `json:"rules_version"`
}

// ComplianceStats holds statistics about a compliance check.
type ComplianceStats struct {
	// TotalRulesChecked is the number of rules evaluated.
	TotalRulesChecked int `json:"total_rules_checked"`

	// ErrorsCount is the number of ERROR severity violations.
	ErrorsCount int `json:"errors_count"`

	// WarningsCount is the number of WARNING severity violations.
	WarningsCount int `json:"warnings_count"`

	// InfoCount is the number of INFO severity messages.
	InfoCount int `json:"info_count"`

	// ComplianceScore is a 0-100 score based on violations.
	ComplianceScore int `json:"compliance_score"`
}

// NewComplianceResult creates a new result.
func NewComplianceResult() *ComplianceResult {
	return &ComplianceResult{
		Compliant:    true,
		Violations:   make([]*Violation, 0),
		RulesVersion: "1.0.0",
	}
}

// AddViolation adds a violation and updates stats.
func (r *ComplianceResult) AddViolation(v *Violation) {
	r.Violations = append(r.Violations, v)

	// Update compliance status
	if v.IsBlocking() {
		r.Compliant = false
	}

	// Update stats
	switch v.Severity {
	case SeverityError:
		r.Stats.ErrorsCount++
	case SeverityWarning:
		r.Stats.WarningsCount++
	case SeverityInfo:
		r.Stats.InfoCount++
	}
}

// Finalize calculates final stats.
func (r *ComplianceResult) Finalize(totalRulesChecked int) {
	r.Stats.TotalRulesChecked = totalRulesChecked

	// Calculate compliance score
	// Start with 100, deduct points for violations
	score := 100
	score -= r.Stats.ErrorsCount * 25   // -25 per error
	score -= r.Stats.WarningsCount * 10 // -10 per warning
	score -= r.Stats.InfoCount * 2      // -2 per info

	if score < 0 {
		score = 0
	}
	r.Stats.ComplianceScore = score
}

// HasErrors returns true if there are any ERROR severity violations.
func (r *ComplianceResult) HasErrors() bool {
	return r.Stats.ErrorsCount > 0
}

// HasWarnings returns true if there are any WARNING severity violations.
func (r *ComplianceResult) HasWarnings() bool {
	return r.Stats.WarningsCount > 0
}

// FilterBySeverity returns violations of specific severity.
func (r *ComplianceResult) FilterBySeverity(severity Severity) []*Violation {
	result := make([]*Violation, 0)
	for _, v := range r.Violations {
		if v.Severity == severity {
			result = append(result, v)
		}
	}
	return result
}

// FilterByCategory returns violations of specific category.
func (r *ComplianceResult) FilterByCategory(category RuleCategory) []*Violation {
	result := make([]*Violation, 0)
	for _, v := range r.Violations {
		if v.Category == category {
			result = append(result, v)
		}
	}
	return result
}

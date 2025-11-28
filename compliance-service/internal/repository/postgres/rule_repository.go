// Package postgres provides PostgreSQL implementations of repositories.
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xiiisorate/granula_api/compliance-service/internal/domain/entity"
	apperrors "github.com/xiiisorate/granula_api/shared/pkg/errors"
)

// RuleRepository handles database operations for compliance rules.
type RuleRepository struct {
	pool *pgxpool.Pool
}

// NewRuleRepository creates a new RuleRepository.
func NewRuleRepository(pool *pgxpool.Pool) *RuleRepository {
	return &RuleRepository{pool: pool}
}

// Create creates a new rule in the database.
func (r *RuleRepository) Create(ctx context.Context, rule *entity.Rule) error {
	appliesToJSON, err := json.Marshal(rule.AppliesTo)
	if err != nil {
		return apperrors.Internal("failed to marshal applies_to").WithCause(err)
	}

	appliesToOpsJSON, err := json.Marshal(rule.AppliesToOperations)
	if err != nil {
		return apperrors.Internal("failed to marshal applies_to_operations").WithCause(err)
	}

	parametersJSON, err := json.Marshal(rule.Parameters)
	if err != nil {
		return apperrors.Internal("failed to marshal parameters").WithCause(err)
	}

	referencesJSON, err := json.Marshal(rule.References)
	if err != nil {
		return apperrors.Internal("failed to marshal references").WithCause(err)
	}

	query := `
		INSERT INTO compliance_rules (
			id, code, category, name, description, severity, active,
			applies_to, applies_to_operations, approval_required,
			parameters, "references", version, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)`

	_, err = r.pool.Exec(ctx, query,
		rule.ID,
		rule.Code,
		rule.Category,
		rule.Name,
		rule.Description,
		rule.Severity,
		rule.Active,
		appliesToJSON,
		appliesToOpsJSON,
		rule.ApprovalRequired,
		parametersJSON,
		referencesJSON,
		rule.Version,
		rule.CreatedAt,
		rule.UpdatedAt,
	)

	if err != nil {
		return apperrors.Internal("failed to create rule").WithCause(err)
	}

	return nil
}

// GetByID retrieves a rule by its ID.
func (r *RuleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Rule, error) {
	query := `
		SELECT 
			id, code, category, name, description, severity, active,
			applies_to, applies_to_operations, approval_required,
			parameters, "references", version, created_at, updated_at
		FROM compliance_rules
		WHERE id = $1`

	row := r.pool.QueryRow(ctx, query, id)
	return r.scanRule(row)
}

// GetByCode retrieves a rule by its code.
func (r *RuleRepository) GetByCode(ctx context.Context, code string) (*entity.Rule, error) {
	query := `
		SELECT 
			id, code, category, name, description, severity, active,
			applies_to, applies_to_operations, approval_required,
			parameters, "references", version, created_at, updated_at
		FROM compliance_rules
		WHERE code = $1`

	row := r.pool.QueryRow(ctx, query, code)
	return r.scanRule(row)
}

// List retrieves rules with optional filtering.
func (r *RuleRepository) List(ctx context.Context, opts ListOptions) ([]*entity.Rule, int, error) {
	// Build query with filters
	query := `
		SELECT 
			id, code, category, name, description, severity, active,
			applies_to, applies_to_operations, approval_required,
			parameters, "references", version, created_at, updated_at
		FROM compliance_rules
		WHERE 1=1`

	countQuery := `SELECT COUNT(*) FROM compliance_rules WHERE 1=1`
	args := make([]interface{}, 0)
	argIndex := 1

	if opts.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND category = $%d", argIndex)
		args = append(args, opts.Category)
		argIndex++
	}

	if opts.Severity != "" {
		query += fmt.Sprintf(" AND severity = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND severity = $%d", argIndex)
		args = append(args, opts.Severity)
		argIndex++
	}

	if opts.ActiveOnly {
		query += " AND active = true"
		countQuery += " AND active = true"
	}

	// Get total count
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, apperrors.Internal("failed to count rules").WithCause(err)
	}

	// Add ordering and pagination
	query += " ORDER BY category, code"
	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, opts.Limit)
		argIndex++
	}
	if opts.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, opts.Offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, apperrors.Internal("failed to list rules").WithCause(err)
	}
	defer rows.Close()

	rules := make([]*entity.Rule, 0)
	for rows.Next() {
		rule, err := r.scanRuleFromRows(rows)
		if err != nil {
			return nil, 0, err
		}
		rules = append(rules, rule)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, apperrors.Internal("error iterating rules").WithCause(err)
	}

	return rules, total, nil
}

// ListByElementType retrieves rules that apply to a specific element type.
func (r *RuleRepository) ListByElementType(ctx context.Context, elementType entity.ElementType) ([]*entity.Rule, error) {
	query := `
		SELECT 
			id, code, category, name, description, severity, active,
			applies_to, applies_to_operations, approval_required,
			parameters, "references", version, created_at, updated_at
		FROM compliance_rules
		WHERE active = true AND applies_to @> $1::jsonb`

	elementTypeJSON, _ := json.Marshal([]entity.ElementType{elementType})

	rows, err := r.pool.Query(ctx, query, string(elementTypeJSON))
	if err != nil {
		return nil, apperrors.Internal("failed to list rules by element type").WithCause(err)
	}
	defer rows.Close()

	rules := make([]*entity.Rule, 0)
	for rows.Next() {
		rule, err := r.scanRuleFromRows(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// ListByOperation retrieves rules that apply to a specific operation.
func (r *RuleRepository) ListByOperation(ctx context.Context, opType entity.OperationType) ([]*entity.Rule, error) {
	query := `
		SELECT 
			id, code, category, name, description, severity, active,
			applies_to, applies_to_operations, approval_required,
			parameters, "references", version, created_at, updated_at
		FROM compliance_rules
		WHERE active = true AND applies_to_operations @> $1::jsonb`

	opTypeJSON, _ := json.Marshal([]entity.OperationType{opType})

	rows, err := r.pool.Query(ctx, query, string(opTypeJSON))
	if err != nil {
		return nil, apperrors.Internal("failed to list rules by operation").WithCause(err)
	}
	defer rows.Close()

	rules := make([]*entity.Rule, 0)
	for rows.Next() {
		rule, err := r.scanRuleFromRows(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// GetCategories retrieves all rule categories with counts.
func (r *RuleRepository) GetCategories(ctx context.Context) ([]*entity.RuleCategoryInfo, error) {
	query := `
		SELECT category, COUNT(*) as rules_count
		FROM compliance_rules
		WHERE active = true
		GROUP BY category
		ORDER BY category`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, apperrors.Internal("failed to get categories").WithCause(err)
	}
	defer rows.Close()

	categories := make([]*entity.RuleCategoryInfo, 0)
	for rows.Next() {
		var cat entity.RuleCategoryInfo
		if err := rows.Scan(&cat.ID, &cat.RulesCount); err != nil {
			return nil, apperrors.Internal("failed to scan category").WithCause(err)
		}
		// Add metadata
		cat.Name, cat.Description, cat.Icon = getCategoryMetadata(cat.ID)
		categories = append(categories, &cat)
	}

	return categories, nil
}

// Update updates an existing rule.
func (r *RuleRepository) Update(ctx context.Context, rule *entity.Rule) error {
	appliesToJSON, _ := json.Marshal(rule.AppliesTo)
	appliesToOpsJSON, _ := json.Marshal(rule.AppliesToOperations)
	parametersJSON, _ := json.Marshal(rule.Parameters)
	referencesJSON, _ := json.Marshal(rule.References)

	query := `
		UPDATE compliance_rules SET
			code = $2,
			category = $3,
			name = $4,
			description = $5,
			severity = $6,
			active = $7,
			applies_to = $8,
			applies_to_operations = $9,
			approval_required = $10,
			parameters = $11,
			"references" = $12,
			version = $13,
			updated_at = $14
		WHERE id = $1`

	result, err := r.pool.Exec(ctx, query,
		rule.ID,
		rule.Code,
		rule.Category,
		rule.Name,
		rule.Description,
		rule.Severity,
		rule.Active,
		appliesToJSON,
		appliesToOpsJSON,
		rule.ApprovalRequired,
		parametersJSON,
		referencesJSON,
		rule.Version,
		rule.UpdatedAt,
	)

	if err != nil {
		return apperrors.Internal("failed to update rule").WithCause(err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("rule", rule.ID.String())
	}

	return nil
}

// Delete deletes a rule by ID.
func (r *RuleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM compliance_rules WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return apperrors.Internal("failed to delete rule").WithCause(err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.NotFound("rule", id.String())
	}

	return nil
}

// CountActive returns the count of active rules.
func (r *RuleRepository) CountActive(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM compliance_rules WHERE active = true").Scan(&count)
	if err != nil {
		return 0, apperrors.Internal("failed to count active rules").WithCause(err)
	}
	return count, nil
}

// ListOptions for filtering rules.
type ListOptions struct {
	Category   entity.RuleCategory
	Severity   entity.Severity
	ActiveOnly bool
	Limit      int
	Offset     int
}

// scanRule scans a single row into a Rule entity.
func (r *RuleRepository) scanRule(row pgx.Row) (*entity.Rule, error) {
	var rule entity.Rule
	var appliesToJSON, appliesToOpsJSON, parametersJSON, referencesJSON []byte

	err := row.Scan(
		&rule.ID,
		&rule.Code,
		&rule.Category,
		&rule.Name,
		&rule.Description,
		&rule.Severity,
		&rule.Active,
		&appliesToJSON,
		&appliesToOpsJSON,
		&rule.ApprovalRequired,
		&parametersJSON,
		&referencesJSON,
		&rule.Version,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFoundMsg("rule not found")
		}
		return nil, apperrors.Internal("failed to scan rule").WithCause(err)
	}

	// Unmarshal JSON fields
	if len(appliesToJSON) > 0 {
		_ = json.Unmarshal(appliesToJSON, &rule.AppliesTo)
	}
	if len(appliesToOpsJSON) > 0 {
		_ = json.Unmarshal(appliesToOpsJSON, &rule.AppliesToOperations)
	}
	if len(parametersJSON) > 0 {
		_ = json.Unmarshal(parametersJSON, &rule.Parameters)
	}
	if len(referencesJSON) > 0 {
		_ = json.Unmarshal(referencesJSON, &rule.References)
	}

	return &rule, nil
}

// scanRuleFromRows scans from pgx.Rows.
func (r *RuleRepository) scanRuleFromRows(rows pgx.Rows) (*entity.Rule, error) {
	var rule entity.Rule
	var appliesToJSON, appliesToOpsJSON, parametersJSON, referencesJSON []byte

	err := rows.Scan(
		&rule.ID,
		&rule.Code,
		&rule.Category,
		&rule.Name,
		&rule.Description,
		&rule.Severity,
		&rule.Active,
		&appliesToJSON,
		&appliesToOpsJSON,
		&rule.ApprovalRequired,
		&parametersJSON,
		&referencesJSON,
		&rule.Version,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err != nil {
		return nil, apperrors.Internal("failed to scan rule").WithCause(err)
	}

	// Unmarshal JSON fields
	if len(appliesToJSON) > 0 {
		_ = json.Unmarshal(appliesToJSON, &rule.AppliesTo)
	}
	if len(appliesToOpsJSON) > 0 {
		_ = json.Unmarshal(appliesToOpsJSON, &rule.AppliesToOperations)
	}
	if len(parametersJSON) > 0 {
		_ = json.Unmarshal(parametersJSON, &rule.Parameters)
	}
	if len(referencesJSON) > 0 {
		_ = json.Unmarshal(referencesJSON, &rule.References)
	}

	return &rule, nil
}

// getCategoryMetadata returns metadata for a category.
func getCategoryMetadata(cat entity.RuleCategory) (name, description, icon string) {
	switch cat {
	case entity.CategoryLoadBearing:
		return "Несущие конструкции", "Правила, связанные с несущими стенами и конструкциями", "wall"
	case entity.CategoryWetZones:
		return "Мокрые зоны", "Правила размещения санузлов, ванных, кухонь", "water"
	case entity.CategoryFireSafety:
		return "Пожарная безопасность", "Правила пожарной безопасности", "fire"
	case entity.CategoryVentilation:
		return "Вентиляция", "Требования к вентиляции помещений", "wind"
	case entity.CategoryMinArea:
		return "Минимальные площади", "Минимальные площади помещений", "square"
	case entity.CategoryDaylight:
		return "Естественное освещение", "Требования к естественному освещению", "sun"
	case entity.CategoryAccessibility:
		return "Доступность", "Требования доступности для маломобильных граждан", "accessibility"
	default:
		return "Общие правила", "Общие правила планировки", "check"
	}
}

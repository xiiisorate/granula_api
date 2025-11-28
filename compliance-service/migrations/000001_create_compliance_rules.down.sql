-- =============================================================================
-- Migration: Drop compliance_rules table
-- =============================================================================

DROP TRIGGER IF EXISTS update_compliance_rules_updated_at ON compliance_rules;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS compliance_rules;
DROP TYPE IF EXISTS approval_type;
DROP TYPE IF EXISTS rule_category;
DROP TYPE IF EXISTS rule_severity;


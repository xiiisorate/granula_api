-- =============================================================================
-- Migration Rollback: Drop Requests Tables
-- =============================================================================
-- WARNING: This will permanently delete all request data!
-- =============================================================================

-- Drop trigger
DROP TRIGGER IF EXISTS trigger_request_updated_at ON requests;

-- Drop function
DROP FUNCTION IF EXISTS update_request_updated_at();

-- Drop tables
DROP TABLE IF EXISTS request_documents;
DROP TABLE IF EXISTS request_status_history;
DROP TABLE IF EXISTS requests;


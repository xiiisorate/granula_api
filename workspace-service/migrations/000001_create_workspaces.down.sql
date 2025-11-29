-- =============================================================================
-- Migration Rollback: Drop Workspaces Tables
-- =============================================================================
-- This rollback removes all workspace-related tables and functions.
-- WARNING: This will permanently delete all workspace data!
-- =============================================================================

-- Drop trigger first
DROP TRIGGER IF EXISTS trigger_workspace_updated_at ON workspaces;

-- Drop function
DROP FUNCTION IF EXISTS update_workspace_updated_at();

-- Drop tables in correct order (respecting foreign keys)
DROP TABLE IF EXISTS workspace_invites;
DROP TABLE IF EXISTS workspace_members;
DROP TABLE IF EXISTS workspaces;


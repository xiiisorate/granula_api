-- =============================================================================
-- Migration: Create Workspaces Tables
-- =============================================================================
-- This migration creates the core tables for workspace management:
-- - workspaces: Main workspace table
-- - workspace_members: Membership/roles relationship
-- - workspace_invites: Pending invitations
--
-- Author: Granula Development Team
-- Date: 2024-11-29
-- =============================================================================

-- Enable UUID extension if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- Table: workspaces
-- =============================================================================
-- Stores workspace metadata. Each workspace is a container for projects,
-- floor plans, scenes, and branches.
--
-- Indexes:
-- - Primary key on id
-- - Index on owner_id for listing user's workspaces
-- - Index on updated_at for sorting by recent activity

CREATE TABLE IF NOT EXISTS workspaces (
    -- Primary key (UUID v4)
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Owner's user ID (references auth-service users)
    owner_id UUID NOT NULL,
    
    -- Display name (2-100 characters)
    name VARCHAR(100) NOT NULL,
    
    -- Optional description (up to 1000 characters)
    description TEXT DEFAULT '',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Soft delete support (optional, for future use)
    deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

-- Index for listing workspaces by owner
CREATE INDEX idx_workspaces_owner_id ON workspaces(owner_id);

-- Index for sorting by recent activity
CREATE INDEX idx_workspaces_updated_at ON workspaces(updated_at DESC);

-- Index for soft deletes (partial index for active workspaces)
CREATE INDEX idx_workspaces_active ON workspaces(id) WHERE deleted_at IS NULL;

-- =============================================================================
-- Table: workspace_members
-- =============================================================================
-- Stores workspace membership and roles. A user can be a member of multiple
-- workspaces, and each membership has a specific role.
--
-- Roles:
-- - owner: Full control, can delete workspace, transfer ownership
-- - admin: Can manage members and settings, cannot delete workspace
-- - editor: Can create and edit content
-- - viewer: Read-only access

CREATE TABLE IF NOT EXISTS workspace_members (
    -- Primary key
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Foreign key to workspace
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    
    -- User ID (references auth-service users)
    user_id UUID NOT NULL,
    
    -- Member role (owner, admin, editor, viewer)
    role VARCHAR(20) NOT NULL DEFAULT 'viewer'
        CHECK (role IN ('owner', 'admin', 'editor', 'viewer')),
    
    -- When the user joined
    joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Who invited this member (NULL for owner)
    invited_by UUID DEFAULT NULL,
    
    -- Unique constraint: one membership per user per workspace
    UNIQUE (workspace_id, user_id)
);

-- Index for listing members of a workspace
CREATE INDEX idx_workspace_members_workspace_id ON workspace_members(workspace_id);

-- Index for listing workspaces a user belongs to
CREATE INDEX idx_workspace_members_user_id ON workspace_members(user_id);

-- Index for role-based queries
CREATE INDEX idx_workspace_members_role ON workspace_members(workspace_id, role);

-- =============================================================================
-- Table: workspace_invites
-- =============================================================================
-- Stores pending workspace invitations. Invites expire after 7 days.
--
-- Status:
-- - pending: Waiting for response
-- - accepted: User joined the workspace
-- - declined: User declined the invite
-- - expired: Invite expired without response
-- - cancelled: Invite was cancelled by sender

CREATE TABLE IF NOT EXISTS workspace_invites (
    -- Primary key
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Foreign key to workspace
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    
    -- User being invited
    invited_user_id UUID NOT NULL,
    
    -- User who sent the invite
    invited_by_user_id UUID NOT NULL,
    
    -- Role to assign when accepted
    role VARCHAR(20) NOT NULL DEFAULT 'viewer'
        CHECK (role IN ('admin', 'editor', 'viewer')),
    
    -- Invite status
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'accepted', 'declined', 'expired', 'cancelled')),
    
    -- When the invite expires
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    responded_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

-- Index for finding pending invites for a user
CREATE INDEX idx_workspace_invites_user_pending ON workspace_invites(invited_user_id, status)
    WHERE status = 'pending';

-- Index for finding invites by workspace
CREATE INDEX idx_workspace_invites_workspace ON workspace_invites(workspace_id);

-- =============================================================================
-- Trigger: Auto-update updated_at timestamp
-- =============================================================================
-- Automatically updates the updated_at column when a workspace is modified.

CREATE OR REPLACE FUNCTION update_workspace_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_workspace_updated_at
    BEFORE UPDATE ON workspaces
    FOR EACH ROW
    EXECUTE FUNCTION update_workspace_updated_at();

-- =============================================================================
-- Comments for documentation
-- =============================================================================

COMMENT ON TABLE workspaces IS 'Collaborative workspaces for managing floor plans and scenes';
COMMENT ON TABLE workspace_members IS 'Workspace membership and role assignments';
COMMENT ON TABLE workspace_invites IS 'Pending workspace invitations';

COMMENT ON COLUMN workspaces.owner_id IS 'UUID of the user who owns this workspace';
COMMENT ON COLUMN workspace_members.role IS 'Member role: owner, admin, editor, or viewer';
COMMENT ON COLUMN workspace_invites.expires_at IS 'Invite expiration time (default: 7 days from creation)';


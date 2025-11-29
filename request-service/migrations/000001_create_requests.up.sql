-- =============================================================================
-- Migration: Create Requests Tables
-- =============================================================================
-- This migration creates the core tables for expert request management:
-- - requests: Main request table
-- - request_status_history: Status change audit trail
-- - request_documents: Attached files
--
-- Author: Granula Development Team
-- Date: 2024-11-29
-- =============================================================================

-- Enable UUID extension if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- Table: requests
-- =============================================================================
-- Stores expert service requests from users.

CREATE TABLE IF NOT EXISTS requests (
    -- Primary key
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Links to workspace
    workspace_id UUID NOT NULL,
    
    -- User who created the request
    user_id UUID NOT NULL,
    
    -- Brief summary (5-200 characters)
    title VARCHAR(200) NOT NULL,
    
    -- Detailed description
    description TEXT DEFAULT '',
    
    -- Service category
    category VARCHAR(50) NOT NULL DEFAULT 'consultation'
        CHECK (category IN ('consultation', 'documentation', 'expert_visit', 'full_package')),
    
    -- Priority level
    priority VARCHAR(20) NOT NULL DEFAULT 'normal'
        CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
    
    -- Current status
    status VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'pending', 'in_review', 'approved', 'rejected', 
                         'assigned', 'in_progress', 'completed', 'cancelled')),
    
    -- Assigned expert (nullable)
    expert_id UUID DEFAULT NULL,
    
    -- When expert was assigned
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    
    -- Estimated cost in rubles
    estimated_cost INTEGER NOT NULL DEFAULT 0,
    
    -- Final cost after completion (nullable)
    final_cost INTEGER DEFAULT NULL,
    
    -- Reason for rejection (if rejected)
    rejection_reason TEXT DEFAULT '',
    
    -- Internal notes (staff only)
    notes TEXT DEFAULT '',
    
    -- Contact information
    contact_phone VARCHAR(20) DEFAULT '',
    contact_email VARCHAR(255) DEFAULT '',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE DEFAULT NULL
);

-- Indexes for common queries
CREATE INDEX idx_requests_workspace_id ON requests(workspace_id);
CREATE INDEX idx_requests_user_id ON requests(user_id);
CREATE INDEX idx_requests_expert_id ON requests(expert_id) WHERE expert_id IS NOT NULL;
CREATE INDEX idx_requests_status ON requests(status);
CREATE INDEX idx_requests_created_at ON requests(created_at DESC);

-- Composite index for listing user's requests in a workspace
CREATE INDEX idx_requests_workspace_user ON requests(workspace_id, user_id, status);

-- =============================================================================
-- Table: request_status_history
-- =============================================================================
-- Audit trail of all status changes.

CREATE TABLE IF NOT EXISTS request_status_history (
    -- Primary key
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Foreign key to request
    request_id UUID NOT NULL REFERENCES requests(id) ON DELETE CASCADE,
    
    -- Previous status
    from_status VARCHAR(20) NOT NULL,
    
    -- New status
    to_status VARCHAR(20) NOT NULL,
    
    -- Reason/comment for the change
    comment TEXT DEFAULT '',
    
    -- User who made the change (NULL for system)
    changed_by UUID DEFAULT NULL,
    
    -- When the change occurred
    changed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for retrieving history by request
CREATE INDEX idx_request_status_history_request_id ON request_status_history(request_id);

-- Index for auditing changes by user
CREATE INDEX idx_request_status_history_changed_by ON request_status_history(changed_by) 
    WHERE changed_by IS NOT NULL;

-- =============================================================================
-- Table: request_documents
-- =============================================================================
-- Attached documents/files.

CREATE TABLE IF NOT EXISTS request_documents (
    -- Primary key
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Foreign key to request
    request_id UUID NOT NULL REFERENCES requests(id) ON DELETE CASCADE,
    
    -- Document type
    type VARCHAR(50) NOT NULL DEFAULT 'other'
        CHECK (type IN ('floor_plan', 'bti_certificate', 'ownership', 'other')),
    
    -- Original filename
    name VARCHAR(255) NOT NULL,
    
    -- Path in object storage
    storage_path VARCHAR(500) NOT NULL,
    
    -- MIME type
    mime_type VARCHAR(100) NOT NULL,
    
    -- File size in bytes
    size BIGINT NOT NULL,
    
    -- User who uploaded
    uploaded_by UUID NOT NULL,
    
    -- Upload timestamp
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for retrieving documents by request
CREATE INDEX idx_request_documents_request_id ON request_documents(request_id);

-- =============================================================================
-- Trigger: Auto-update updated_at timestamp
-- =============================================================================

CREATE OR REPLACE FUNCTION update_request_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_request_updated_at
    BEFORE UPDATE ON requests
    FOR EACH ROW
    EXECUTE FUNCTION update_request_updated_at();

-- =============================================================================
-- Comments
-- =============================================================================

COMMENT ON TABLE requests IS 'Expert service requests from users';
COMMENT ON TABLE request_status_history IS 'Audit trail of request status changes';
COMMENT ON TABLE request_documents IS 'Documents attached to requests';

COMMENT ON COLUMN requests.category IS 'Service type: consultation (2000₽), documentation (15000₽), expert_visit (5000₽), full_package (30000₽)';
COMMENT ON COLUMN requests.priority IS 'Urgency level affecting processing time';
COMMENT ON COLUMN requests.status IS 'Workflow state: draft → pending → in_review → approved/rejected → assigned → in_progress → completed';


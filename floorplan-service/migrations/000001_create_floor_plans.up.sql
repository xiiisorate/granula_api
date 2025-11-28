-- Floor Plans table
-- Хранит метаданные планировок

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Floor plan status enum
CREATE TYPE floor_plan_status AS ENUM (
    'UPLOADED',
    'PROCESSING',
    'RECOGNIZED',
    'CONFIRMED',
    'FAILED'
);

-- Floor plans table
CREATE TABLE IF NOT EXISTS floor_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id UUID NOT NULL,
    owner_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    status floor_plan_status NOT NULL DEFAULT 'UPLOADED',
    recognition_job_id UUID,
    scene_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_floor_plans_workspace_id ON floor_plans(workspace_id);
CREATE INDEX idx_floor_plans_owner_id ON floor_plans(owner_id);
CREATE INDEX idx_floor_plans_status ON floor_plans(status);
CREATE INDEX idx_floor_plans_created_at ON floor_plans(created_at DESC);

-- File info table
CREATE TABLE IF NOT EXISTS floor_plan_files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    floor_plan_id UUID NOT NULL REFERENCES floor_plans(id) ON DELETE CASCADE,
    original_name VARCHAR(512) NOT NULL,
    storage_path VARCHAR(1024) NOT NULL,
    mime_type VARCHAR(128) NOT NULL,
    size BIGINT NOT NULL,
    checksum VARCHAR(64),
    width INTEGER DEFAULT 0,
    height INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_floor_plan_files_floor_plan_id ON floor_plan_files(floor_plan_id);

-- Thumbnails table
CREATE TABLE IF NOT EXISTS floor_plan_thumbnails (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    floor_plan_id UUID NOT NULL REFERENCES floor_plans(id) ON DELETE CASCADE,
    size VARCHAR(32) NOT NULL, -- e.g., "128x128", "256x256"
    storage_path VARCHAR(1024) NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_floor_plan_thumbnails_floor_plan_id ON floor_plan_thumbnails(floor_plan_id);

-- Updated at trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_floor_plans_updated_at
    BEFORE UPDATE ON floor_plans
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE floor_plans IS 'Floor plan documents uploaded by users';
COMMENT ON TABLE floor_plan_files IS 'File metadata for floor plans stored in object storage';
COMMENT ON TABLE floor_plan_thumbnails IS 'Generated thumbnails for floor plans';

COMMENT ON COLUMN floor_plans.workspace_id IS 'Workspace this floor plan belongs to';
COMMENT ON COLUMN floor_plans.recognition_job_id IS 'AI recognition job ID (if processing)';
COMMENT ON COLUMN floor_plans.scene_id IS 'Created scene ID (after successful recognition)';
COMMENT ON COLUMN floor_plan_files.storage_path IS 'Path in MinIO/S3 object storage';
COMMENT ON COLUMN floor_plan_files.checksum IS 'MD5 checksum for integrity verification';


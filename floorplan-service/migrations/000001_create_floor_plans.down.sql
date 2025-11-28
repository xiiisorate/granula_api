-- Rollback floor plans tables

DROP TRIGGER IF EXISTS update_floor_plans_updated_at ON floor_plans;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS floor_plan_thumbnails;
DROP TABLE IF EXISTS floor_plan_files;
DROP TABLE IF EXISTS floor_plans;

DROP TYPE IF EXISTS floor_plan_status;


-- Drop indices
DROP INDEX IF EXISTS experiment_segment;
DROP INDEX IF EXISTS experiment_time_range;

-- Drop constraints
ALTER TABLE settings DROP CONSTRAINT settings_project_id_positive;
ALTER TABLE experiments DROP CONSTRAINT experiments_interval_null_or_positive;

-- Drop tables
DROP TABLE IF EXISTS experiments;
DROP TABLE IF EXISTS settings;

-- Drop types
DROP TYPE IF EXISTS experiment_status;
DROP TYPE IF EXISTS experiment_type;
DROP TYPE IF EXISTS experiment_tier;

-- Removes treatment_schema and validation_url from table.
-- Hence, this migration involves data loss for project settings with treatment_schema and/or validation_url configured
ALTER TABLE settings DROP COLUMN treatment_schema;
ALTER TABLE settings DROP COLUMN validation_url;


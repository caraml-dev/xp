-- Drop constraint
ALTER TABLE treatments DROP CONSTRAINT treatment_unique_name;

-- Drop table
DROP TABLE IF EXISTS treatments;

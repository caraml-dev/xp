-- Drop constraint
ALTER TABLE segments DROP CONSTRAINT segment_unique_name;

-- Drop table
DROP TABLE IF EXISTS segments;

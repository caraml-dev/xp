-- Drop constraint
ALTER TABLE custom_segmenters DROP CONSTRAINT segmenter_unique_name;

-- Drop table
DROP TABLE IF EXISTS custom_segmenters;

-- Drop segmenter_type enum
DROP TYPE IF EXISTS segmenter_type;

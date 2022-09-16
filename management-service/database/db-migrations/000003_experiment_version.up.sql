ALTER TABLE experiments ADD version integer NOT NULL DEFAULT 1;

-- Set the version number for existing records
WITH agg_data AS (
    SELECT t1.id AS id, count(*) AS num_history
    FROM experiments t1 INNER JOIN experiment_history t2 ON t1.id = t2.experiment_id
    GROUP BY t1.id
) UPDATE experiments SET version = num_history+1
FROM agg_data WHERE experiments.id = agg_data.id;

-- Settings Table
CREATE TABLE IF NOT EXISTS settings
(
   project_id      integer       PRIMARY KEY,

   username        varchar(50)   NOT NULL UNIQUE,
   passkey         varchar(128)  NOT NULL,
   config          jsonb,

   created_at      timestamp     NOT NULL default current_timestamp,
   updated_at      timestamp     NOT NULL default current_timestamp,

   CONSTRAINT settings_project_id_positive CHECK (project_id > 0)
);


-- Experiments Table
CREATE TYPE experiment_status as ENUM ('active', 'inactive');
CREATE TYPE experiment_type as ENUM ('A/B', 'Switchback');
CREATE TYPE experiment_tier as ENUM ('default', 'override');
 
CREATE TABLE IF NOT EXISTS experiments
(
   id              bigserial           PRIMARY KEY,
   name            varchar(64)         NOT NULL,
   description     text,
 
   type            experiment_type     NOT NULL,
   tier            experiment_tier     NOT NULL default 'default',
   interval        int,
   treatments      jsonb,
   segment         jsonb,
      
   status          experiment_status   NOT NULL default 'active',
   start_time      timestamp           WITH TIME ZONE NOT NULL,
   end_time        timestamp           WITH TIME ZONE NOT NULL,
  
   project_id      integer             NOT NULL references settings (project_id) ON DELETE CASCADE,
   
   created_at      timestamp           NOT NULL default current_timestamp,
   updated_at      timestamp           NOT NULL default current_timestamp,
   updated_by      varchar(255),

   CONSTRAINT experiments_interval_null_or_positive CHECK (interval ISNULL or interval > 0)
);
 
CREATE INDEX experiment_segment ON experiments USING gin (segment);
-- Timestamp with timezone range index, inclusive of start time (>=), exclusive of end time (<)
CREATE INDEX experiment_time_range ON experiments USING gist (tstzrange(start_time, end_time, '[)'));

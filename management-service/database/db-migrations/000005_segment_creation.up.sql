CREATE TABLE IF NOT EXISTS segments
(
   id              bigserial        PRIMARY KEY,
   name            varchar(64)      NOT NULL,
   segment         jsonb            NOT NULL,

   project_id      integer          NOT NULL references settings (project_id) ON DELETE CASCADE,

   created_at      timestamp        NOT NULL default current_timestamp,
   updated_at      timestamp        NOT NULL default current_timestamp,
   updated_by      varchar(255),

   CONSTRAINT segment_unique_name UNIQUE (name, project_id)
);

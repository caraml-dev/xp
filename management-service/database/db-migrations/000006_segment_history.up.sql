CREATE TABLE IF NOT EXISTS segment_history
(
   id              serial             PRIMARY KEY,
   segment_id      integer            NOT NULL references segments (id) ON DELETE CASCADE,
   version         integer,

   name            varchar(64)        NOT NULL,
   segment         jsonb,
  
   created_at      timestamp          NOT NULL default current_timestamp,
   updated_at      timestamp          NOT NULL default current_timestamp,
   updated_by      varchar(255)
);

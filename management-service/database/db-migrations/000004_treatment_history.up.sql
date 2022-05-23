CREATE TABLE IF NOT EXISTS treatment_history
(
   id              serial             PRIMARY KEY,
   treatment_id    integer            NOT NULL references treatments (id) ON DELETE CASCADE,
   version         integer,

   name            varchar(64)        NOT NULL,
   configuration   jsonb,
  
   created_at      timestamp          NOT NULL default current_timestamp,
   updated_at      timestamp          NOT NULL default current_timestamp,
   updated_by      varchar(255)
);

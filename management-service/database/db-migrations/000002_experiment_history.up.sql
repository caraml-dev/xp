CREATE TABLE IF NOT EXISTS experiment_history
(
   id              bigserial           PRIMARY KEY,
   experiment_id   integer             NOT NULL references experiments (id) ON DELETE CASCADE,
   version         integer,

   name            varchar(64)         NOT NULL,
   description     text,

   type            experiment_type     NOT NULL,
   tier            experiment_tier     NOT NULL default 'default',
   interval        integer,
   treatments      jsonb,
   segment         jsonb,
      
   status          experiment_status   NOT NULL default 'active',
   start_time      timestamp           WITH TIME ZONE NOT NULL,
   end_time        timestamp           WITH TIME ZONE NOT NULL,
  
   created_at      timestamp           NOT NULL default current_timestamp,
   updated_at      timestamp           NOT NULL default current_timestamp,
   updated_by      varchar(255)
);

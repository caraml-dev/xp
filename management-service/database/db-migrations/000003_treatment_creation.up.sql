CREATE TABLE IF NOT EXISTS treatments
(
    id              serial              PRIMARY KEY,
    name            varchar(64)         NOT NULL,
    configuration   jsonb,

    project_id      integer             NOT NULL references settings (project_id) ON DELETE CASCADE,

    created_at      timestamp           NOT NULL default current_timestamp,
    updated_at      timestamp           NOT NULL default current_timestamp,
    updated_by      varchar(255),
    CONSTRAINT treatment_unique_name UNIQUE (name, project_id)
);

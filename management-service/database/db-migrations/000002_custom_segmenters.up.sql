CREATE TYPE segmenter_type as ENUM ('STRING', 'BOOL', 'INTEGER', 'REAL');

CREATE TABLE IF NOT EXISTS custom_segmenters
(
    project_id integer NOT NULL references settings (project_id) ON DELETE CASCADE,

    name            varchar(64)     NOT NULL,
    type            segmenter_type  NOT NULL,
    description     text,

    required        boolean,
    multi_valued    boolean,
    options         jsonb,
    constraints     jsonb,

    created_at      timestamp     NOT NULL default current_timestamp,
    updated_at      timestamp     NOT NULL default current_timestamp,

    CONSTRAINT segmenter_unique_name UNIQUE (name, project_id)
);

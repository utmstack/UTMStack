CREATE TABLE IF NOT EXISTS utm_incident_actions (
    id             BIGSERIAL PRIMARY KEY,
    action_command VARCHAR(255),
    action_description VARCHAR(255),
    action_params  VARCHAR(255),
    action_type    INTEGER,
    action_editable BOOLEAN NOT NULL DEFAULT FALSE,
    created_date   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    modified_date  TIMESTAMPTZ,
    created_user   VARCHAR(50) NOT NULL DEFAULT 'system',
    modified_user  VARCHAR(50)
);

CREATE TABLE IF NOT EXISTS utm_incident_action_command (
    id           BIGSERIAL PRIMARY KEY,
    action_id    BIGINT NOT NULL,
    os_platform  VARCHAR(50),
    command      VARCHAR(1024)
);

CREATE TABLE IF NOT EXISTS utm_incident_jobs (
    id            BIGSERIAL PRIMARY KEY,
    action_id     BIGINT,
    params        VARCHAR(1024),
    agent         VARCHAR(255),
    status        INTEGER,
    job_result    TEXT,
    origin_id     INTEGER NOT NULL DEFAULT 0,
    origin_type   VARCHAR(50) NOT NULL DEFAULT '',
    created_date  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    modified_date TIMESTAMPTZ,
    created_user  VARCHAR(50) NOT NULL DEFAULT 'system',
    modified_user VARCHAR(50)
);

CREATE TABLE IF NOT EXISTS utm_incident_variables (
    id                  BIGSERIAL PRIMARY KEY,
    variable_name       VARCHAR(255) UNIQUE,
    variable_value      TEXT,
    variable_description TEXT,
    is_secret           BOOLEAN NOT NULL DEFAULT FALSE,
    created_by          VARCHAR(255),
    last_modified_date  TIMESTAMPTZ,
    last_modified_by    VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS utm_incident (
    id BIGSERIAL PRIMARY KEY,
    incident_name VARCHAR(255) NOT NULL UNIQUE,
    incident_description TEXT,
    incident_status VARCHAR(50) NOT NULL,
    incident_severity INTEGER,
    incident_assigned_to TEXT,
    incident_solution TEXT,
    incident_created_date TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS utm_incident_alert (
    id BIGSERIAL PRIMARY KEY,
    incident_id BIGINT NOT NULL REFERENCES utm_incident(id),
    alert_id VARCHAR(255) NOT NULL,
    alert_name VARCHAR(255) NOT NULL,
    alert_severity INTEGER NOT NULL,
    alert_status INTEGER,
    CONSTRAINT uq_incident_alert_id UNIQUE (alert_id)
);

CREATE TABLE IF NOT EXISTS utm_incident_note (
    id BIGSERIAL PRIMARY KEY,
    incident_id BIGINT NOT NULL REFERENCES utm_incident(id),
    note_text VARCHAR(1000) NOT NULL,
    note_send_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    note_send_by VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS utm_incident_history (
    id BIGSERIAL PRIMARY KEY,
    incident_id BIGINT NOT NULL REFERENCES utm_incident(id),
    action VARCHAR(255),
    action_type VARCHAR(100) NOT NULL,
    action_detail TEXT,
    action_created_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    action_created_by VARCHAR(255)
);

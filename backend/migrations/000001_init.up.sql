-- Baseline cleanup: drop the legacy JHipster persistent-audit tables.
--
-- The Go audit module persists to a fresh `audit_logs` table (created by
-- GORM AutoMigrate) with a tamper-evident hash chain. The old JHipster
-- audit schema is obsolete: it was effectively dead in the panel (the
-- "User access audit" sidebar link was removed), and the canonical audit
-- trail lives in OpenSearch. Drop the child table first (FK), then the
-- parent and its sequence.

DROP TABLE IF EXISTS jhi_persistent_audit_evt_data;
DROP TABLE IF EXISTS jhi_persistent_audit_event;
DROP SEQUENCE IF EXISTS jhi_persistent_audit_event_event_id_seq;

-- Drop legacy OpenVAS columns on jhi_user — not modelled in the Go User.
ALTER TABLE jhi_user DROP COLUMN IF EXISTS openvas_user_uuid;
ALTER TABLE jhi_user DROP COLUMN IF EXISTS openvas_user_id;

-- Drop the config-parameter section grouping. The Go appconfig module updates
-- seeded parameters by key (conf_param_short) and never reads the section, which
-- only existed so the legacy frontend could group settings. Drop the column
-- first (removes its FK), then the now-unreferenced table.
ALTER TABLE utm_configuration_parameter DROP COLUMN IF EXISTS section_id;
DROP TABLE IF EXISTS utm_configuration_section;

-- Drop the SOAR rule config-change history. The snapshot table was written on
-- every rule update but its UI (the rule-history view) was never wired into the
-- legacy frontend, so the data is unused. The Go soar module no longer models it.
DROP TABLE IF EXISTS utm_alert_response_rule_history;

-- Drop the alerts checkpoint table. The Go scheduler computes its evaluation
-- window from time.Now() each tick and never reads this row back — it was
-- write-only dead state. The Go alerts module no longer models it.
DROP TABLE IF EXISTS utm_alert_last;

-- Drop the alert change-history table. History now lives embedded in each alert
-- document in OpenSearch (the alert's `history` array), written from deploy time
-- onward. Pre-migration history is intentionally not carried over.
DROP TABLE IF EXISTS utm_alert_log;

-- Drop the asset-metrics table. Its only producers were the legacy network_scan
-- services (UtmNetworkScanService / UtmAssetGroupService), which are not ported,
-- so nothing in the Go stack reads or writes it. When network_scan is ported it
-- can reintroduce its own metrics storage.
DROP TABLE IF EXISTS utm_asset_metrics;

-- Drop the agent-manager registry table. It was dead JHipster scaffolding in the
-- legacy backend: an empty @SuppressWarnings("unused") repository that was never
-- injected, never seeded, never read or written. Nothing in the Go stack (backend,
-- agent-manager service or agents) references it.
DROP TABLE IF EXISTS utm_agent_manager;

-- Drop the SOC AI processing-request table. It was dead code in the legacy backend
-- (entity/repo/service with zero references) and was deliberately not ported: the
-- Go socai module is a pure HTTP passthrough to SOC_AI_BASE_URL with no DB.
DROP TABLE IF EXISTS utm_alert_socai_processing_request;

-- Drop the licensing/client table. The Go backend validates the license against
-- the LICENSE file directly instead of a DB row, so utm_client is obsolete. It
-- was created/seeded by the installer (now removed there too) and only consumed
-- by the legacy Angular license module; the new stack does not reference it.
DROP TABLE IF EXISTS utm_client;

-- Drop the incident-response action/command/job/variable tables. This was a
-- designed-but-never-finished feature: the predefined "actions" (BLOCK_IP,
-- ISOLATE_HOST, RUN_CMD, …), their per-OS commands and automation variables were
-- seeded and CRUD-able, jobs could be created, but no processor ever dispatched a
-- job and the CRUD screens were unreachable in the legacy frontend. The only live
-- command path is the /soar/ws/command WebSocket, which runs raw commands against
-- the agent-manager — it no longer interpolates variables or masks secrets, so
-- none of these tables have any consumer left.
DROP TABLE IF EXISTS utm_incident_jobs;
DROP TABLE IF EXISTS utm_incident_action_command;
DROP TABLE IF EXISTS utm_incident_actions;
DROP TABLE IF EXISTS utm_incident_variables;

-- utm_incident_history: the Go model renamed the legacy column action_date to
-- action_created_date. AutoMigrate (which runs before this) adds the new column
-- but leaves the old action_date (NOT NULL, no default) in place on upgraded
-- installs, so every history INSERT would violate its NOT NULL constraint.
-- Copy the legacy timestamps into the new column and drop the obsolete one.
-- Guarded so it is a no-op on fresh installs where action_date never existed.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'utm_incident_history' AND column_name = 'action_date'
    ) THEN
        UPDATE utm_incident_history SET action_created_date = action_date WHERE action_date IS NOT NULL;
        ALTER TABLE utm_incident_history DROP COLUMN action_date;
    END IF;
END $$;

-- IAM: roles live in jhi_authority (name = the role id). Membership is read
-- straight from jhi_user_authority (no break on existing installs).
INSERT INTO jhi_authority (name, name_show, description) VALUES
    ('ROLE_ADMIN', 'Administrator', 'Full access to all resources'),
    ('ROLE_USER',  'User',          'Standard user with module-level read access')
ON CONFLICT (name) DO NOTHING;

INSERT INTO permissions (name, description, resource, action) VALUES
    ('users.read',       'List and view users',           'users',       'read'),
    ('users.write',      'Create and update users',       'users',       'write'),
    ('users.delete',     'Deactivate users',              'users',       'delete'),
    ('roles.read',       'List and view roles',           'roles',       'read'),
    ('config.read',      'Read application config',             'config', 'read'),
    ('config.write',     'Update or delete application config', 'config', 'write'),
    ('audit.read',       'List and view audit log entries', 'audit',     'read'),
    ('soar.read',        'List and view SOAR response rules',     'soar', 'read'),
    ('soar.write',       'Create and update SOAR response rules', 'soar', 'write'),
    ('alerts.read',      'List and view alerts, tags and tag rules',          'alerts', 'read'),
    ('alerts.write',     'Update alerts and manage alert tags and tag rules', 'alerts', 'write'),
    ('incidents.read',   'List and view incidents, alerts, notes and history', 'incidents', 'read'),
    ('incidents.write',  'Create and manage incidents, their alerts and notes', 'incidents', 'write')
ON CONFLICT (name) DO NOTHING;

-- Bind ROLE_ADMIN to every permission currently in the catalog. Re-run this
-- scoped to new permissions in each module's seed as it ships.
INSERT INTO authority_permissions (authority_name, permission_id)
SELECT a.name, p.id
FROM jhi_authority a CROSS JOIN permissions p
WHERE a.name = 'ROLE_ADMIN'
ON CONFLICT DO NOTHING;

-- Bind ROLE_USER to every read-only permission in the catalog (module-level read
-- access). Re-run this scoped to new read permissions in each module's seed.
INSERT INTO authority_permissions (authority_name, permission_id)
SELECT 'ROLE_USER', p.id
FROM permissions p
WHERE p.action = 'read'
ON CONFLICT DO NOTHING;

-- Seed the system-owned "False positive" alert tag.
--
-- The legacy backend seeded this row in data.sql (id=1, system_owner=true). It
-- is the canonical tag the rules engine applies to auto-complete alerts: the
-- OpenSearch Painless scripts and the Go usecase reference it by NAME
-- ("False positive"), so the managed tag must exist for the UI tag picker and
-- for parity with legacy installs.
--
-- On an in-place upgrade the legacy row already exists (utm_alert_tag is not
-- dropped), so this is a no-op via ON CONFLICT. On a fresh install GORM creates
-- an empty table and this seeds it.
--
-- We intentionally omit an explicit id and let the GORM-created sequence assign
-- it: forcing id=1 would not advance the sequence and the next auto-insert would
-- collide. The Go code keys off tag_name, not the id, so the value is irrelevant.
INSERT INTO utm_alert_tag (tag_name, tag_color, system_owner)
VALUES ('False positive', '#f44336', true)
ON CONFLICT (tag_name) DO NOTHING;

-- Seed default mail configuration rows in utm_configuration_parameter so the
-- settings UI has stable keys to bind to before an admin fills them in. Keys
-- mirror pkg/constants/mail_configuration.go and the values/metadata mirror the
-- legacy Liquibase data.sql so the Go appconfig module (which only ever UPDATEs
-- pre-seeded params, never creates them) and the legacy panel stay at parity.
--
-- smtp.auth is the encryption type (TLS/SSL/NONE), not a boolean: toEmailConfig
-- maps a non-empty/non-"none"/non-"false" value to SMTP auth enabled.
--
-- conf_param_short has no unique index, so we guard each row with NOT EXISTS
-- instead of ON CONFLICT. On an in-place upgrade the legacy rows already exist
-- (this is a no-op); on a fresh install GORM creates an empty table and this
-- seeds it.
INSERT INTO utm_configuration_parameter
    (conf_param_short, conf_param_large, conf_param_description, conf_param_value, conf_param_required, conf_param_datatype, conf_param_option)
SELECT v.conf_param_short, v.conf_param_large, v.conf_param_description, v.conf_param_value, v.conf_param_required, v.conf_param_datatype, v.conf_param_option
FROM (VALUES
    ('utmstack.mail.host',                      'Mail Server Host',         'SMTP server host. For instance, smtp.example.com.', '',    true,  'text',     NULL),
    ('utmstack.mail.port',                      'Mail Server Port',         'SMTP server port',                                  '587', true,  'number',   NULL),
    ('utmstack.mail.username',                  'Mail Server Username',     'Login user of the SMTP server',                     '',    true,  'text',     NULL),
    ('utmstack.mail.password',                  'Mail Server Password',     'Login password of the SMTP server',                 '',    true,  'password', NULL),
    ('utmstack.mail.from',                      'Utmstack email address',   'Address from which emails are sent',                '',    true,  'email',    NULL),
    ('utmstack.mail.baseUrl',                   'Utmstack base url',        'Base url of Utmstack',                              '',    true,  'text',     NULL),
    ('utmstack.mail.organization',              'Organization Name',        'This field helps identify the organization name in incident and alert notification emails.', '', false, 'text', NULL),
    ('utmstack.mail.properties.mail.smtp.auth', 'Encryption type',          'Select the encryption type used by the SMTP server', 'TLS', true, 'radio',    'TLS,SSL,NONE')
) AS v(conf_param_short, conf_param_large, conf_param_description, conf_param_value, conf_param_required, conf_param_datatype, conf_param_option)
WHERE NOT EXISTS (
    SELECT 1 FROM utm_configuration_parameter c WHERE c.conf_param_short = v.conf_param_short
);

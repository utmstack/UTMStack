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
    ('audit.read',       'List and view audit log entries', 'audit',     'read'),
    ('soar.read',        'List and view SOAR response rules',     'soar', 'read'),
    ('soar.write',       'Create and update SOAR response rules', 'soar', 'write')
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

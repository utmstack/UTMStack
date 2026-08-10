-- Seeds the product needs on a fresh install.
--
-- There is no upgrade path in here. v11 installs are not migrated in place —
-- v12 is installed clean and the data is carried over separately — so the
-- table drops, column reshapes and tenant backfills that used to live here are
-- gone. What remains is content the product ships and cannot start without.
--
-- Ordering: GORM AutoMigrate creates the tables before this runs.
-- Idempotent: every statement is ON CONFLICT DO NOTHING or guarded by NOT EXISTS.

-- IAM catalogue. A permission is identified by its name: it is what the code
-- checks (RequirePermission("soar.write")), so a surrogate id would only add a
-- join to every authorisation query.
INSERT INTO permissions (name, description) VALUES
    ('users.read',            'List and view users'),
    ('users.write',           'Create and update users, and assign their roles'),
    ('users.delete',          'Deactivate users'),
    ('roles.read',            'List and view roles and the permission catalogue'),
    ('roles.write',           'Create, update and delete the tenant''s own roles'),
    ('idp.read',              'List and view identity-provider configurations'),
    ('idp.write',             'Create, update and delete identity-provider configurations'),
    ('tenant.read',           'List and view tenants'),
    ('tenant.write',          'Provision, update and terminate tenants'),
    ('config.read',           'Read application config'),
    ('config.write',          'Update or delete application config'),
    ('audit.read',            'List and view audit log entries'),
    ('notifications.read',    'List and view notifications'),
    ('notifications.write',   'Mark notifications read and delete them'),
    ('alerts.read',           'List and view alerts, tags and tag rules'),
    ('alerts.write',          'Update alerts and manage alert tags and tag rules'),
    ('incidents.read',        'List and view incidents, alerts, notes and history'),
    ('incidents.write',       'Create and manage incidents, their alerts and notes'),
    ('soar.read',             'List and view SOAR response flows'),
    ('soar.write',            'Create and run SOAR response flows'),
    ('compliance.read',       'List and view frameworks, controls, reports and score history'),
    ('compliance.write',      'Manage frameworks and controls, run evaluations, record verdicts on a report and manage report schedules'),
    ('dashboards.read',       'List and view dashboards, visualizations and their layouts'),
    ('dashboards.write',      'Create, update and delete dashboards, visualizations and layouts'),
    ('loganalyzer.read',      'Explore logs and view saved queries'),
    ('loganalyzer.write',     'Create, update and delete saved log-analyzer queries'),
    ('datasources.read',      'List and view datasources'),
    ('datasources.write',     'Update datasource labels and asset sensitivity, and delete datasources'),
    ('eventprocessing.read',  'List and view correlation rules and regex patterns'),
    ('eventprocessing.write', 'Create, update and delete correlation rules and regex patterns'),
    ('integrations.read',     'List integrations and their configuration'),
    ('integrations.write',    'Activate integrations and manage their configuration'),
    ('adaudit.read',          'List and view the audited Active Directory user inventory'),
    ('threatintel.read',      'Search threat intelligence and its feeds'),
    ('opensearch.read',       'Query OpenSearch (search, properties, cluster status)'),
    ('opensearch.write',      'Destructive OpenSearch ops (delete index)'),
    -- These spend money against a quota the whole instance shares, so they are
    -- permissions rather than something every authenticated user simply has.
    ('socai.read',            'Use the SOC AI assistant'),
    ('socai.write',           'Configure the SOC AI provider')
ON CONFLICT (name) DO NOTHING;

-- The three roles every tenant gets. They belong to no tenant and carry
-- system_owner, which is what puts them in every tenant's reads and out of every
-- tenant's writes: a customer assigns them and builds its own beside them, but
-- cannot edit or delete these. The ids are fixed so re-running is a no-op.
INSERT INTO role (id, tenant_id, name, display_name, description, system_owner, created_at, updated_at) VALUES
    ('00000000-0000-0000-0000-00000000a001', '00000000-0000-0000-0000-000000000000',
     'ROLE_ADMIN',   'Administrator',
     'Full access, including users, roles and configuration', TRUE, NOW(), NOW()),
    ('00000000-0000-0000-0000-00000000a002', '00000000-0000-0000-0000-000000000000',
     'ROLE_ANALYST', 'Analyst',
     'Runs the SOC: triages alerts, works incidents and launches response actions', TRUE, NOW(), NOW()),
    ('00000000-0000-0000-0000-00000000a003', '00000000-0000-0000-0000-000000000000',
     'ROLE_VIEWER',  'Viewer',
     'Reads everything and changes nothing', TRUE, NOW(), NOW())
ON CONFLICT (tenant_id, name) DO NOTHING;

INSERT INTO role_permission (role_id, permission_name)
SELECT '00000000-0000-0000-0000-00000000a001', p.name FROM permissions p
ON CONFLICT DO NOTHING;

-- The analyst reads everything and writes only where the work happens. Not
-- users or roles, because assigning a role is how you grant yourself the rest;
-- not config, tenants or identity providers, which are the instance's shape
-- rather than its operation.
INSERT INTO role_permission (role_id, permission_name)
SELECT '00000000-0000-0000-0000-00000000a002', p.name FROM permissions p
WHERE p.name LIKE '%.read'
   OR p.name IN ('alerts.write', 'incidents.write', 'soar.write',
                 'loganalyzer.write', 'dashboards.write', 'notifications.write')
ON CONFLICT DO NOTHING;

INSERT INTO role_permission (role_id, permission_name)
SELECT '00000000-0000-0000-0000-00000000a003', p.name FROM permissions p
WHERE p.name LIKE '%.read'
ON CONFLICT DO NOTHING;

-- Seed the system-owned "False positive" alert tag.

--
-- Idempotent: every statement is ON CONFLICT DO NOTHING or a delete.


-- "False positive" is the tag the rules engine applies to auto-complete an
-- alert. Both the engine and this service reference it by NAME, never by id,
-- so what matters is that a row with this name exists before either runs.
--
-- It is system-owned, which is how it belongs to every tenant and to none: the
-- tenancy callbacks read a scoped table as "tenant_id = mine OR system_owner",
-- so the tenant on this row is never matched against. It still has to hold a
-- value — the column is NOT NULL — and the nil UUID is the one that says "no
-- tenant" without naming one.
--
-- The id is left to the column's own gen_random_uuid(). The conflict target is
-- (tenant_id, tag_name) because that is the unique index: a tag name is unique
-- inside a tenant, not across the table.
INSERT INTO alert_tag (tenant_id, tag_name, tag_color, system_owner)
VALUES ('00000000-0000-0000-0000-000000000000', 'False positive', '#f44336', true)
ON CONFLICT (tenant_id, tag_name) DO NOTHING;

-- Seed default mail configuration rows in app_config so the
-- settings UI has stable keys to bind to before an admin fills them in. Keys

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
INSERT INTO app_config
    (tenant_id, conf_param_short, conf_param_large, conf_param_description, conf_param_value, conf_param_required, conf_param_datatype, conf_param_option)
SELECT 'ce66672c-e36d-4761-a8c8-90058fee1a24', v.conf_param_short, v.conf_param_large, v.conf_param_description, v.conf_param_value, v.conf_param_required, v.conf_param_datatype, v.conf_param_option
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
    SELECT 1 FROM app_config c WHERE c.conf_param_short = v.conf_param_short
);

-- White-label branding is stored as a single JSON value on this config row
-- (the appconfig branding usecase marshals/unmarshals it). Reuses the config
-- table instead of a dedicated table. Empty value => default UTMStack brand.
INSERT INTO app_config
    (tenant_id, conf_param_short, conf_param_large, conf_param_description, conf_param_value, conf_param_required, conf_param_datatype, conf_param_option)
SELECT 'ce66672c-e36d-4761-a8c8-90058fee1a24', 'branding', 'White-label branding', 'White-label branding (logo, product name, colors) as JSON', '', false, 'text', NULL
WHERE NOT EXISTS (
    SELECT 1 FROM app_config c WHERE c.conf_param_short = 'branding'
);

-- Platform-default UI language. Drives the pre-login screens, system emails and
-- the default for users without a personal lang_key (each user can override it
-- from their profile). The appconfig module only UPDATEs pre-seeded params, so
-- it must exist here for GET/PUT /config/utmstack.system.language to work.
INSERT INTO app_config
    (tenant_id, conf_param_short, conf_param_large, conf_param_description, conf_param_value, conf_param_required, conf_param_datatype, conf_param_option)
SELECT 'ce66672c-e36d-4761-a8c8-90058fee1a24', 'utmstack.system.language', 'Platform Language', 'Default UI language for the platform (pre-login screens, system emails and new users). Each user can override it in their profile.', 'en', false, 'radio', 'en,es,pt,fr,de'
WHERE NOT EXISTS (
    SELECT 1 FROM app_config c WHERE c.conf_param_short = 'utmstack.system.language'
);

-- Org-wide timestamp display preference. Data is always stored in UTC; these only
-- control how timestamps are rendered in the UI (timezone + format). Read app-wide
-- via the public GET /date-format, edited by an admin via /config.
INSERT INTO app_config
    (tenant_id, conf_param_short, conf_param_large, conf_param_description, conf_param_value, conf_param_required, conf_param_datatype, conf_param_option)
SELECT 'ce66672c-e36d-4761-a8c8-90058fee1a24', v.conf_param_short, v.conf_param_large, v.conf_param_description, v.conf_param_value, v.conf_param_required, v.conf_param_datatype, v.conf_param_option
FROM (VALUES
    ('utmstack.time.zone',       'Default Time Zone', 'Time zone used to display timestamps. Logs remain stored in UTC.', 'UTC',    false, 'text', NULL),
    ('utmstack.time.dateformat', 'Date Format',       'Format used to display dates and times.',                          'medium', false, 'radio', 'short,medium,long,full')
) AS v(conf_param_short, conf_param_large, conf_param_description, conf_param_value, conf_param_required, conf_param_datatype, conf_param_option)
WHERE NOT EXISTS (
    SELECT 1 FROM app_config c WHERE c.conf_param_short = v.conf_param_short
);

-- ThreatWinds (feeds plugin) integration credentials. Seeded as empty so the
-- appconfig module (which only UPDATEs pre-seeded params) can serve/store them
-- and the feeds plugin reads them via GET /api/v1/config/<key>. apiSecret is a
-- secret (encrypted at rest, decrypted on per-key read).
INSERT INTO app_config
    (tenant_id, conf_param_short, conf_param_large, conf_param_description, conf_param_value, conf_param_required, conf_param_datatype, conf_param_option)
SELECT 'ce66672c-e36d-4761-a8c8-90058fee1a24', v.conf_param_short, v.conf_param_large, v.conf_param_description, v.conf_param_value, v.conf_param_required, v.conf_param_datatype, v.conf_param_option
FROM (VALUES
    ('utmstack.tw.enabled',   'ThreatWinds Enabled',    'Whether the ThreatWinds intelligence integration is enabled', '', false, 'text',     NULL),
    ('utmstack.tw.apiKey',    'ThreatWinds API Key',    'API Key for ThreatWinds integration.',                        '', true,  'text',     NULL),
    ('utmstack.tw.apiSecret', 'ThreatWinds API Secret', 'API Secret for ThreatWinds integration.',                     '', true,  'password', NULL)
) AS v(conf_param_short, conf_param_large, conf_param_description, conf_param_value, conf_param_required, conf_param_datatype, conf_param_option)
WHERE NOT EXISTS (
    SELECT 1 FROM app_config c WHERE c.conf_param_short = v.conf_param_short
);

-- Named regex patterns and asset CIA are no longer tables: the patterns ship
-- baked into the pipeline bootstrap, and asset sensitivity lives on the
-- datasource row that owns it.


-- The integrations the release ships. They carry system_owner and the nil
-- tenant, which is what puts them in every tenant's reads and out of every
-- tenant's writes: a customer configures them and adds its own beside them, but
-- cannot edit or delete these. Description and icon are deliberately absent —
-- the frontend owns them so they can be translated and themed (see
-- i18n integrations.modules.<NAME> and constants/systemModules.ts).
INSERT INTO integrations (tenant_id, name, data_type, ingest_type, system_owner) VALUES
    ('00000000-0000-0000-0000-000000000000', 'WINDOWS_AGENT',      'wineventlog',                'agent', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'LINUX_AGENT',        'linux',                      'agent', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'MACOS',              'macos',                      'agent', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'VMWARE',             'vmware-esxi',                'forwarder', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'NETFLOW',            'netflow',                    'forwarder', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'AWS_IAM_USER',       'aws',                        'plugin', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'AZURE',              'azure',                      'plugin', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'O365',               'o365',                       'plugin', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'GCP',                'google',                     'plugin', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'KASPERSKY',          'antivirus-kaspersky',        'forwarder', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'ESET',               'antivirus-esmc-eset',        'forwarder', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'BITDEFENDER',        'antivirus-bitdefender-gz',   'plugin', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'SENTINEL_ONE',       'antivirus-sentinel-one',     'forwarder', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'SOPHOS',             'sophos-central',             'plugin', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'CROWDSTRIKE',        'crowdstrike',                'plugin', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'CISCO',              'firewall-cisco-asa',         'forwarder', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'MERAKI',             'firewall-meraki',            'forwarder', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'FIRE_POWER',         'firewall-cisco-firepower',   'forwarder', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'CISCO_SWITCH',       'cisco-switch',               'forwarder', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'FORTIGATE',          'firewall-fortigate-traffic', 'forwarder', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'FORTIWEB',           'firewall-fortiweb',          'forwarder', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'SOPHOS_XG',          'firewall-sophos-xg',         'forwarder', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'PALO_ALTO',          'firewall-paloalto',          'forwarder', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'SONIC_WALL',         'firewall-sonicwall',         'forwarder', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'PFSENSE',            'firewall-pfsense',           'forwarder', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'MIKROTIK',           'firewall-mikrotik',          'forwarder', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'AIX',                'ibm-aix',                    'forwarder', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'AS_400',             'ibm-as400',                  'collector', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'ORACLE',             'oracle',                     'forwarder', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'SURICATA',           'suricata',                   'forwarder', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'DECEPTIVE_BYTES',    'deceptive-bytes',            'forwarder', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'GITHUB',             'github',                     'forwarder', TRUE),
    ('00000000-0000-0000-0000-000000000000', 'UTMSTACK',           'utmstack',                   'collector', TRUE)
ON CONFLICT (tenant_id, name) DO UPDATE SET
    data_type   = EXCLUDED.data_type,
    ingest_type = EXCLUDED.ingest_type;

-- Seed the system index patterns. Uses v11- prefix directly (no UPDATE step).

-- preserved as-is; on fresh installs all rows are inserted.
-- Active rows match the current production set.
INSERT INTO utm_index_pattern (id, pattern, pattern_module, pattern_system, is_active) VALUES
    (1,  'v11-log-*',                              NULL,                                                                                    true, true),
    (2,  'v11-alert-*',                            NULL,                                                                                    true, true),
    (8,  'v11-log-wineventlog-*',                  'WINDOWS_AGENT',                                                                         true, true),
    (10, 'v11-log-aws-*',                          'AWS_IAM_USER',                                                                          true, false),
    (11, 'v11-log-azure-*',                        'AZURE',                                                                                  true, false),
    (12, 'v11-log-o365-*',                         'O365',                                                                                   true, false),
    (13, 'v11-log-firewall-meraki-*',              'MERAKI',                                                                                 true, false),
    (14, 'v11-log-firewall-*',                     'MERAKI,SOPHOS_XG,CISCO,FORTIGATE,FIRE_POWER,MIKROTIK,PALO_ALTO,SONIC_WALL,PFSENSE,FORTIWEB', true, true),
    (15, 'v11-log-firewall-cisco-asa-*',           'CISCO',                                                                                  true, false),
    (17, 'v11-log-firewall-sophos-xg-*',           'SOPHOS_XG',                                                                              true, true),
    (19, 'v11-log-generic-*',                      NULL,                                                                                    true, true),
    (21, 'v11-log-firewall-fortigate-traffic-*',   'FORTIGATE',                                                                              true, true),
    (24, 'v11-log-vmware-esxi-*',                  'VMWARE',                                                                                 true, false),
    (25, 'v11-log-google-*',                       'GCP',                                                                                    true, false),
    (26, 'v11-log-firewall-cisco-firepower-*',     'FIRE_POWER',                                                                             true, false),
    (39, 'v11-log-linux-*',                        'LINUX_AGENT',                                                                            true, true),
    (40, 'v11-log-antivirus-*',                    'ESET,SENTINEL_ONE,KASPERSKY',                                                            true, false),
    (41, 'v11-log-antivirus-esmc-eset-*',          'ESET',                                                                                   true, false),
    (42, 'v11-log-antivirus-kaspersky-*',          'KASPERSKY',                                                                              true, false),
    (43, 'v11-log-antivirus-sentinel-one-*',       'SENTINEL_ONE',                                                                           true, false),
    (44, 'v11-log-sophos-central-*',               'SOPHOS',                                                                                 true, false),
    (45, 'v11-log-github-*',                       'GITHUB',                                                                                 true, true),
    (47, 'v11-log-macos-*',                        'MACOS',                                                                                  true, false),
    (48, 'v11-log-firewall-mikrotik-*',            'MIKROTIK',                                                                               true, false),
    (49, 'v11-log-firewall-paloalto-*',            'PALO_ALTO',                                                                              true, false),
    (50, 'v11-log-cisco-switch-*',                 'CISCO_SWITCH',                                                                           true, false),
    (51, 'v11-log-firewall-sonicwall-*',           'SONIC_WALL',                                                                             true, false),
    (52, 'v11-log-deceptive-bytes-*',              'DECEPTIVE_BYTES',                                                                        true, false),
    (53, 'v11-log-antivirus-bitdefender-gz-*',     'BITDEFENDER',                                                                            true, false),
    (56, 'v11-soc-ai',                             'SOC_AI',                                                                                 true, true),
    (60, 'v11-log-json-input-*',                   'JSON',                                                                                   true, true),
    (62, 'v11-log-syslog-*',                       'SYSLOG',                                                                                 true, true),
    (63, 'v11-log-firewall-pfsense-*',             'PFSENSE',                                                                                true, false),
    (64, 'v11-log-netflow-*',                      'NETFLOW',                                                                                true, false),
    (65, 'v11-log-firewall-fortiweb-*',            'FORTIWEB',                                                                               true, false),
    (66, 'v11-log-ibm-aix-*',                      'AIX',                                                                                    true, false),
    (67, 'v11-log-ibm-as400-*',                    'AS_400',                                                                                 true, true),
    (68, 'v11-log-suricata-*',                     'SURICATA',                                                                               true, false),
    (69, 'v11-log-utmstack-*',                     'UTMSTACK',                                                                               true, true),
    (70, 'v11-log-crowdstrike-*',                  'CrowdStrike',                                                                            true, true)
ON CONFLICT (id) DO NOTHING;

-- Advance the sequence above the highest seeded id so new user-created patterns
-- don't collide with the system rows.
SELECT setval(pg_get_serial_sequence('utm_index_pattern', 'id'), GREATEST((SELECT MAX(id) FROM utm_index_pattern), 1000));


-- ---- visualizations carry a spec, not SQL ------------------------------
-- Purpose:
--   A visualization used to store a statement built against OpenSearch. It
--   stores the question now — dataset, aggregation, breakdown, filters — and
--   the statement is built when the query runs, by the driver that knows the
--   dialect. That is also what applies the tenant: the query is built with a
--   scope carrying it, rather than depending on whoever wrote the SQL.
--
--   The old column is dropped rather than carried: its contents are the wrong
--   dialect against the wrong store, so nothing in it can be salvaged.
--
-- Ordering: AutoMigrate runs BEFORE this and adds spec.
-- Idempotent: IF EXISTS.

ALTER TABLE visualization DROP COLUMN IF EXISTS sql_query;


-- ---- saved queries name a dataset ---------------------------------------
-- Purpose:
--   A saved log-analyzer query pointed at a row in the index-pattern registry.
--   The event store has two datasets and no registry, so the query names one
--   directly and the foreign key goes.
--
-- Ordering: AutoMigrate runs BEFORE this.
-- Idempotent: IF EXISTS.

ALTER TABLE saved_query DROP COLUMN IF EXISTS id_pattern;

-- ---- the saved-query table is ours now -----------------------------------
-- Purpose:
--   utm_log_analyzer_query, with its la_ column prefixes, was the Java ORM's
--   naming. AutoMigrate creates the new table from the model but never drops
--   what it replaced, so on a database that predates the rename both exist and
--   the rows are in the old one. Move them across, once, and drop it.
--
-- Ordering: AutoMigrate runs BEFORE this, so saved_query already exists.
-- Idempotent: keyed on the old table still being there, and copies nothing
--   when the new one already holds rows.

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables
             WHERE table_schema = 'public' AND table_name = 'utm_log_analyzer_query') THEN

    INSERT INTO saved_query (id, tenant_id, name, description, owner,
                             created_at, updated_at, columns, filters, dataset)
    SELECT id, tenant_id, la_name, la_description, la_owner,
           la_creation_date, la_modification_date, la_columns, la_filters, la_data_origin
    FROM utm_log_analyzer_query
    ON CONFLICT (id) DO NOTHING;

    -- Keep the sequence ahead of the ids just inserted, or the next save
    -- collides with a row that came over.
    PERFORM setval(pg_get_serial_sequence('saved_query', 'id'),
                   GREATEST((SELECT COALESCE(MAX(id), 0) FROM saved_query), 1));

    DROP TABLE utm_log_analyzer_query;
  END IF;
END $$;

-- Email recipients, kept apart for alerts and for incidents because they are
-- different audiences at different volumes: an incident is raised by a person
-- and there are few, while alerts arrive on their own and at whatever rate the
-- environment produces. One list for both means whoever wants incident mail
-- also gets every alert.
--
-- All four optional and comma-separated. The appconfig usecase only ever
-- UPDATEs pre-seeded rows, so they have to exist here for GET/PUT
-- /config/<key> to answer.
INSERT INTO app_config
    (tenant_id, conf_param_short, conf_param_large, conf_param_description, conf_param_value, conf_param_required, conf_param_datatype, conf_param_option)
SELECT 'ce66672c-e36d-4761-a8c8-90058fee1a24', v.conf_param_short, v.conf_param_large, v.conf_param_description, v.conf_param_value, v.conf_param_required, v.conf_param_datatype, v.conf_param_option
FROM (VALUES
    ('utmstack.alerts.notification_to',    'Alerts notification To',    'Comma-separated addresses that receive new-alert notifications. Empty means alerts are not emailed.', '', false, 'text', NULL),
    ('utmstack.alerts.notification_cc',    'Alerts notification Cc',    'Comma-separated addresses copied on new-alert notifications.',                                     '', false, 'text', NULL),
    ('utmstack.incidents.notification_to', 'Incidents notification To', 'Comma-separated addresses that receive incident notifications. Empty means incidents are not emailed.', '', false, 'text', NULL),
    ('utmstack.incidents.notification_cc', 'Incidents notification Cc', 'Comma-separated addresses copied on incident notifications.',                                      '', false, 'text', NULL)
) AS v(conf_param_short, conf_param_large, conf_param_description, conf_param_value, conf_param_required, conf_param_datatype, conf_param_option)
WHERE NOT EXISTS (
    SELECT 1 FROM app_config c WHERE c.conf_param_short = v.conf_param_short
);

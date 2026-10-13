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
    ('storage.read',          'Read how long the event store keeps records and what it holds'),
    ('storage.write',         'Change retention and configure cold storage'),
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

--
-- smtp.auth is the encryption type (TLS/SSL/NONE), not a boolean: toEmailConfig
-- maps a non-empty/non-"none"/non-"false" value to SMTP auth enabled.
INSERT INTO app_config (tenant_id, key, label, description, value, is_secret)
VALUES
    ('ce66672c-e36d-4761-a8c8-90058fee1a24', 'utmstack.mail.host',     'Mail Server Host',       'SMTP server host. For instance, smtp.example.com.', '',    false),
    ('ce66672c-e36d-4761-a8c8-90058fee1a24', 'utmstack.mail.port',     'Mail Server Port',       'SMTP server port',                                  '587', false),
    ('ce66672c-e36d-4761-a8c8-90058fee1a24', 'utmstack.mail.username', 'Mail Server Username',   'Login user of the SMTP server',                     '',    false),
    ('ce66672c-e36d-4761-a8c8-90058fee1a24', 'utmstack.mail.password', 'Mail Server Password',   'Login password of the SMTP server',                 '',    true),
    ('ce66672c-e36d-4761-a8c8-90058fee1a24', 'utmstack.mail.from',     'Sender email address',   'Address from which emails are sent',                '',    false),
    ('ce66672c-e36d-4761-a8c8-90058fee1a24', 'utmstack.mail.baseUrl',  'Base URL',               'Base URL of this installation, used in email links', '',   false),
    ('ce66672c-e36d-4761-a8c8-90058fee1a24', 'utmstack.mail.organization', 'Organization Name',  'Identifies the organization in incident and alert notification emails.', '', false),
    ('ce66672c-e36d-4761-a8c8-90058fee1a24', 'utmstack.mail.properties.mail.smtp.auth', 'Encryption type', 'Encryption used by the SMTP server: TLS, SSL or NONE', 'TLS', false),

    -- White-label branding is one JSON value on one row: logo, product name and
    -- colors, empty meaning the shipped brand.
    ('ce66672c-e36d-4761-a8c8-90058fee1a24', 'branding', 'White-label branding', 'Logo, product name and colors, as JSON. Empty means the shipped brand.', '', false),

    -- Platform-default UI language. Drives the pre-login screens, system emails
    -- and users with no personal lang_key; each user can override it.
    ('ce66672c-e36d-4761-a8c8-90058fee1a24', 'utmstack.system.language', 'Platform Language', 'Default UI language: en, es, pt, fr, de.', 'en', false),

    -- Timestamps are stored in UTC; these only decide how they are displayed.
    --
    -- The ThreatWinds contribution is deliberately absent: its credentials are
    -- secrets, and the config API withholds a secret by design, so they live
    -- encrypted in the feeds plugin's own config file instead.
    ('ce66672c-e36d-4761-a8c8-90058fee1a24', 'utmstack.time.zone',       'Default Time Zone', 'Time zone used to display timestamps. Records stay stored in UTC.', 'UTC',    false),
    ('ce66672c-e36d-4761-a8c8-90058fee1a24', 'utmstack.time.dateformat', 'Date Format',       'How dates and times are displayed: short, medium, long or full.',   'medium', false)
ON CONFLICT (tenant_id, key) DO NOTHING;


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


-- Email recipients, kept apart for alerts and for incidents because they are
-- different audiences at different volumes: an incident is raised by a person
-- and there are few, while alerts arrive on their own and at whatever rate the
-- environment produces. One list for both means whoever wants incident mail
-- also gets every alert.
--
-- All four optional and comma-separated. The appconfig usecase only ever
-- UPDATEs pre-seeded rows, so they have to exist here for GET/PUT
-- /config/<key> to answer.
INSERT INTO app_config (tenant_id, key, label, description, value, is_secret)
VALUES
    ('ce66672c-e36d-4761-a8c8-90058fee1a24', 'utmstack.alerts.notification_to',    'Alerts notification To',    'Comma-separated addresses that receive new-alert notifications. Empty means alerts are not emailed.', '', false),
    ('ce66672c-e36d-4761-a8c8-90058fee1a24', 'utmstack.alerts.notification_cc',    'Alerts notification Cc',    'Comma-separated addresses copied on new-alert notifications.', '', false),
    ('ce66672c-e36d-4761-a8c8-90058fee1a24', 'utmstack.incidents.notification_to', 'Incidents notification To', 'Comma-separated addresses that receive incident notifications. Empty means incidents are not emailed.', '', false),
    ('ce66672c-e36d-4761-a8c8-90058fee1a24', 'utmstack.incidents.notification_cc', 'Incidents notification Cc', 'Comma-separated addresses copied on incident notifications.', '', false)
ON CONFLICT (tenant_id, key) DO NOTHING;

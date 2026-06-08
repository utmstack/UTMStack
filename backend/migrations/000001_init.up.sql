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

-- Drop the data-source ingestion-status tables. utm_data_input_status was a
-- materialized cache of OpenSearch `v11-statistics-*` (latest event timestamp per
-- source+data_type) and utm_data_input_status_checkpoint only held the sync poll
-- cursor. Nothing in the Go stack reads them: no frontend calls the endpoints, no
-- module depends on them, and the liveness fields (median/isDown) were unused.
-- Source liveness is derived from OpenSearch directly; the datainput module and
-- the network_scan sync writer are removed. Drop the checkpoint first.
DROP TABLE IF EXISTS utm_data_input_status_checkpoint;
DROP TABLE IF EXISTS utm_data_input_status;

-- Drop the Liquibase bookkeeping tables. The Go backend manages schema versioning
-- with golang-migrate (schema_migrations / schema_migrations_pre), so Liquibase's
-- internal changelog tracking is obsolete. These linger on databases upgraded from
-- the Java backend; nothing in the Go stack reads or writes them.
DROP TABLE IF EXISTS databasechangelog;
DROP TABLE IF EXISTS databasechangeloglock;

-- Drop the dashboard per-role ACL table. The Go dashboards module ports
-- utm_dashboard / utm_visualization / utm_dashboard_visualization but NOT this
-- authority table — dashboards are not gated per-role in the new stack (module
-- permission dashboards.read/write governs access). No Go entity references it.
DROP TABLE IF EXISTS utm_dashboard_authority;

-- Drop the agent-manager registry table. It was dead JHipster scaffolding in the
-- legacy backend: an empty @SuppressWarnings("unused") repository that was never
-- injected, never seeded, never read or written. Nothing in the Go stack (backend,
-- agent-manager service or agents) references it.
DROP TABLE IF EXISTS utm_agent_manager;

-- Drop utm_ports and its sequence. The table stored open ports discovered by the
-- network probe scanner, but the only Angular components that could write to it
-- (AssetCreateComponent / AssetPortCreateComponent) were dead code — their
-- selectors were never mounted in any reachable template. No automated job writes
-- to it either. The Go network_scan module no longer models this table.
DROP TABLE IF EXISTS public.utm_ports;
DROP SEQUENCE IF EXISTS public.utm_ports_id_seq;

-- Drop utm_network_scan and utm_asset_types. The datasources module recreates
-- utm_network_scan from a clean, slimmed schema in 000003 (asset_type_id and the dead
-- sync/probe columns removed, `label` added, group_id kept). Dropping the dirty table
-- here — before 000003 — makes the clean CREATE take effect on in-place upgrades too,
-- where the legacy table would otherwise survive 000003's CREATE ... IF NOT EXISTS.
-- This is a documented BREAKING CHANGE: per-asset curation (group assignments, notes,
-- alias) is discarded. Group definitions in utm_asset_group are kept, and alerts already
-- enriched with their group retain it in OpenSearch. utm_asset_types is removed for good,
-- replaced by the free-text `label` column on utm_network_scan. CASCADE clears the FK from
-- the dropped table; utm_network_scan_id_seq is left for 000003 to reuse.
DROP TABLE IF EXISTS public.utm_network_scan CASCADE;
DROP TABLE IF EXISTS public.utm_asset_types;
DROP SEQUENCE IF EXISTS public.utm_asset_types_id_seq;

-- Drop utm_collectors. It was a GORM-AutoMigrated local cache of the agent-manager's
-- collector registry (synced on demand by the collectors module). Collectors are folding
-- into datasources (SourceKind=collector, registered via ping), so the cache is redundant.
-- No data is lost — the agent-manager remains the source of truth. Removed from Models()
-- so AutoMigrate no longer recreates it.
DROP TABLE IF EXISTS public.utm_collectors;

-- Drop utm_asset_group.type, the legacy ASSET/COLLECTOR discriminator added by
-- liquibase 20240506001. It only powered the old split asset-groups vs collector-groups
-- UI; the Go datasources model never mapped it and the new stack uses a single group
-- namespace. CASCADE so any unique constraint that still includes the column (the legacy
-- composite over group_name+type) is removed with it, whatever its name. Single-name
-- uniqueness is then governed by the GORM uniqueIndex on group_name (UtmAssetGroup model).
-- NOTE on in-place upgrades: group_name now must be globally unique — if an install has the
-- same group_name across an ASSET and a COLLECTOR group, those rows must be deduplicated
-- before this runs.
ALTER TABLE utm_asset_group DROP COLUMN IF EXISTS type CASCADE;

-- Drop the SOC AI processing-request table. It was dead code in the legacy backend
-- (entity/repo/service with zero references) and was deliberately not ported: the
-- Go socai module is a pure HTTP passthrough to SOC_AI_BASE_URL with no DB.
DROP TABLE IF EXISTS utm_alert_socai_processing_request;

-- Drop the licensing/client table. The Go backend validates the license against
-- the LICENSE file directly instead of a DB row, so utm_client is obsolete. It
-- was created/seeded by the installer (now removed there too) and only consumed
-- by the legacy Angular license module; the new stack does not reference it.
DROP TABLE IF EXISTS utm_client;

-- Drop the logstash pipeline tables (utm_group_logstash_pipeline_filters,
-- utm_logstash_pipeline): Logstash-era health artifacts. The native event
-- processor doesn't fail silently; ingestion health is now live via
-- /eventprocessing/ingestion-stats (v11-statistics-*).
DROP TABLE IF EXISTS utm_group_logstash_pipeline_filters CASCADE;
DROP TABLE IF EXISTS utm_logstash_pipeline CASCADE;

-- Drop the filter-group catalog (utm_logstash_filter_group): never populated,
-- no UI, no YAML references it. utm_logstash_filter is NOT dropped here —
-- FilterBootstrap reads it first, migrates user filters to YAML, then drops
-- it. Dropping here (before the bootstrap runs) would lose user filter data.
DROP TABLE IF EXISTS utm_logstash_filter_group CASCADE;

-- The incident-response automation tables (utm_incident_actions,
-- utm_incident_action_command, utm_incident_jobs, utm_incident_variables) are
-- KEPT and migrated to the Go soar module:
--   - variables   → interpolated into /soar/ws/command commands, secrets masked
--   - actions     → predefined responses (SHUTDOWN_SERVER, RUN_CMD, …)
--   - jobs        → responses run against agents, surfaced in incident reports
-- All four are (re)created/managed by AutoMigrate, so their legacy data must
-- survive an in-place upgrade — nothing is dropped here.

-- Seed the predefined incident-response actions and their per-OS commands
-- (ported from the legacy Liquibase data.sql). Idempotent: ON CONFLICT keeps any
-- existing rows from an upgraded install untouched. AutoMigrate created the
-- tables before this runs.
INSERT INTO utm_incident_actions (id, action_command, action_description, action_params, action_type, action_editable, created_date, created_user, modified_date, modified_user) VALUES
    (1, 'SHUTDOWN_SERVER',   'Shutdown server',                              NULL, 1, false, '2021-07-09 19:06:50.578', 'system', NULL, NULL),
    (2, 'DISABLE_USER',      'Kick out and disable user',                    NULL, 2, false, '2020-03-19 23:21:23.444', 'system', NULL, NULL),
    (3, 'BLOCK_IP',          'Block ip and disconnect any traffic from IP',  NULL, 3, false, '2021-07-09 19:06:50.578', 'system', NULL, NULL),
    (4, 'ISOLATE_HOST',      'Isolate host (disconnect from network)',       NULL, 4, false, '2021-07-09 19:06:50.578', 'system', NULL, NULL),
    (5, 'RESTART_SERVER',    'Restart server',                               NULL, 5, false, '2021-07-09 19:06:50.578', 'system', NULL, NULL),
    (6, 'KILL_PROCESS',      'Kill process',                                 NULL, 6, false, '2021-07-09 19:06:50.578', 'system', NULL, NULL),
    (7, 'UNINSTALL_PROGRAM', 'Uninstall program',                            NULL, 7, false, '2021-07-09 19:06:50.578', 'system', NULL, NULL),
    (8, 'RUN_CMD',           'Run shell command',                            NULL, 8, false, '2021-07-09 19:06:50.578', 'system', NULL, NULL)
ON CONFLICT (id) DO NOTHING;

INSERT INTO utm_incident_action_command (id, action_id, os_platform, command) VALUES
    (9,  1, 'windows', 'cmd.exe /c shutdown -s -t 0 -f'),
    (10, 1, 'linux',   'init 0')
ON CONFLICT (id) DO NOTHING;

-- Keep the identity sequences ahead of the explicitly-seeded ids so future
-- API inserts don't collide with them.
SELECT setval(pg_get_serial_sequence('utm_incident_actions', 'id'), GREATEST((SELECT MAX(id) FROM utm_incident_actions), 1));
SELECT setval(pg_get_serial_sequence('utm_incident_action_command', 'id'), GREATEST((SELECT MAX(id) FROM utm_incident_action_command), 1));

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

-- Drop the utm_module columns the integrations model removed. AutoMigrate now
-- uses integrations.UtmModule (no server_id/lite_version/needs_restart/
-- is_activatable), but it never drops columns, so the legacy ones linger on
-- upgrade — remove them. data_type/is_system were added pre-AutoMigrate.
ALTER TABLE utm_module DROP COLUMN IF EXISTS server_id;
ALTER TABLE utm_module DROP COLUMN IF EXISTS lite_version;
ALTER TABLE utm_module DROP COLUMN IF EXISTS needs_restart;
ALTER TABLE utm_module DROP COLUMN IF EXISTS is_activatable;

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
    ('incidents.write',  'Create and manage incidents, their alerts and notes', 'incidents', 'write'),
    ('opensearch.read',  'Query OpenSearch (search, properties, cluster status)', 'opensearch', 'read'),
    ('opensearch.write', 'Destructive OpenSearch ops (delete index)',            'opensearch', 'write'),
    ('eventprocessing.read',  'List and view correlation rules, regex patterns and tenant config',      'eventprocessing', 'read'),
    ('eventprocessing.write', 'Create, update and delete correlation rules, regex patterns and tenant config', 'eventprocessing', 'write'),
    ('integrations.read',  'List integrations, their tenants and configuration',          'integrations', 'read'),
    ('integrations.write', 'Activate integrations and create/update/delete their tenants', 'integrations', 'write'),
    ('compliance.read',  'List and view compliance standards, controls, reports and evaluation history', 'compliance', 'read'),
    ('compliance.write', 'Create, update and delete compliance standards, controls and report schedules', 'compliance', 'write'),
    ('datasources.read',  'List and view datasources and their groups',                 'datasources', 'read'),
    ('datasources.write', 'Create, update and delete datasources, their groups and group assignments', 'datasources', 'write'),
    ('dashboards.read',  'List and view dashboards, visualizations and their layouts',  'dashboards', 'read'),
    ('dashboards.write', 'Create, update and delete dashboards, visualizations and layouts', 'dashboards', 'write'),
    ('loganalyzer.read',  'Explore logs (top values, chart view) and view saved queries', 'loganalyzer', 'read'),
    ('loganalyzer.write', 'Create, update and delete saved log-analyzer queries',          'loganalyzer', 'write'),
    ('idp.read',  'List and view SAML identity-provider configurations', 'idp', 'read'),
    ('idp.write', 'Create, update and delete SAML identity-provider configurations', 'idp', 'write')
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

-- utm_regex_pattern and utm_tenant_config are now file-backed (patterns.yaml
-- utm_regex_pattern and utm_tenant_config are NOT dropped here — PipelineBootstrap
-- reads them first, migrates to YAML files, then drops them. Dropping here
-- (before the bootstrap runs) would lose user data.

-- Seed the built-in (system) integrations into utm_module.
--
-- On a fresh install GORM creates an empty utm_module and this populates the
-- catalog (name, icon, category, data_type, is_system=true). On an in-place
-- upgrade these rows already exist (the pre-AutoMigrate step backfilled their
-- data_type/is_system), so ON CONFLICT (module_name) makes this a no-op.
--
-- SOC_AI and FILE_INTEGRITY are intentionally omitted: they are not data-source
-- integrations in the new model. The 7 pullers (AWS/AZURE/GCP/O365/BITDEFENDER/
-- CROWDSTRIKE/SOPHOS) carry their field schema in integrations/<name>.yaml.
INSERT INTO utm_module (module_name, pretty_name, module_description, module_active, module_icon, module_category, data_type, is_system) VALUES
    ('WINDOWS_AGENT',   'Windows Agent',                 'Send Windows operating-system logs to UTMStack via the installed agent.',                        true,  'windows-agent.svg', 'Agents & Syslog', 'wineventlog',                true),
    ('LINUX_AGENT',     'Linux agent',                   'Send Linux operating-system logs to UTMStack via the installed agent.',                          true,  'linux_agent.svg',   'Agents & Syslog', 'linux',                      true),
    ('MACOS',           'MacOS',                         'macOS is a series of operating systems for the Apple Mac family of computers.',                  true,  'macos.svg',         'Agents & Syslog', 'macos',                      true),
    ('SYSLOG',          'Syslog',                        'Accept syslog from firewalls and other network devices that support it.',                        true,  'log-file.svg',      'Agents & Syslog', 'syslog',                     true),
    ('VMWARE',          'VMWare Syslog',                 'Redirect and store VMware ESXi syslog messages in UTMStack.',                                    false, 'vmware.svg',        'Agents & Syslog', 'vmware-esxi',                true),
    ('JSON',            'Json Input',                    'Send your JSON-format logs to be processed by UTMStack.',                                        true,  'json.svg',          'Other',           'json-input',                 true),
    ('NETFLOW',         'Netflow',                       'Redirect network-traffic NetFlow logs to UTMStack for monitoring and analysis.',                 false, 'netflow.svg',       'Network',         'netflow',                    true),
    ('AWS_IAM_USER',    'AWS Cloudwatch',                'Audit AWS account activity and API usage via CloudTrail/CloudWatch.',                            false, 'aws-cloudtrail.svg', 'Cloud',          'aws',                        true),
    ('AZURE',           'Azure',                         'Microsoft Azure public cloud computing platform (IaaS, PaaS, SaaS).',                            false, 'azure.svg',         'Cloud',           'azure',                      true),
    ('O365',            'Microsoft 365',                 'Microsoft 365 subscription services (Exchange Online, Teams, SharePoint, OneDrive).',            false, 'office.svg',        'Cloud',           'o365',                       true),
    ('GCP',             'Google Cloud Platform',         'Google Cloud Platform hosted services for compute, storage and application development.',         false, 'gcp.svg',           'Cloud',           'google',                     true),
    ('KASPERSKY',       'Kaspersky Security',            'Comprehensive information about the devices and applications running on your network.',           false, 'kaspersky.svg',     'Antivirus',       'antivirus-kaspersky',        true),
    ('ESET',            'ESET Endpoint Protection',      'Multilayered endpoint protection with machine learning and easy management.',                    false, 'eset.svg',          'Antivirus',       'antivirus-esmc-eset',        true),
    ('BITDEFENDER',     'Bitdefender',                   'Bitdefender GravityZone cybersecurity for businesses and consumers.',                            false, 'bitdefender.svg',   'Antivirus',       'antivirus-bitdefender-gz',   true),
    ('SENTINEL_ONE',    'SentinelOne Endpoint Security', 'SentinelOne endpoint security (Core, Control and Complete tiers).',                              false, 'sentinelone.svg',   'XDR',             'antivirus-sentinel-one',     true),
    ('SOPHOS',          'Sophos Central',                'Single cloud management for Sophos endpoint, server, mobile, firewall, ZTNA and email.',         false, 'sophos.svg',        'XDR',             'sophos-central',             true),
    ('CROWDSTRIKE',     'CrowdStrike',                   'CrowdStrike Falcon cloud-native endpoint and workload protection.',                              true,  'crowdstrike.svg',   'Device',          'crowdstrike',                true),
    ('CISCO',           'Cisco ASA',                     'Cisco ASA adaptive security appliances — ingest device logs into UTMStack.',                     false, 'cisco.svg',         'Device',          'firewall-cisco-asa',         true),
    ('MERAKI',          'Cisco Meraki',                  'Syslog integration with Cisco Meraki firewalls.',                                               false, 'meraki.svg',        'Device',          'firewall-meraki',            true),
    ('FIRE_POWER',      'Fire Power',                    'Cisco Firepower next-generation firewalls — threat protection and IPS.',                          false, 'fire-power.svg',    'Device',          'firewall-cisco-firepower',   true),
    ('CISCO_SWITCH',    'Cisco Switch',                  'Cisco network switches — performance, flexibility and security telemetry.',                       false, 'cisco.svg',         'Device',          'cisco-switch',               true),
    ('FORTIGATE',       'FortiGate',                     'Fortinet FortiGate next-generation firewalls (NGFW).',                                           false, 'fortigate.svg',     'Device',          'firewall-fortigate-traffic', true),
    ('FORTIWEB',        'FortiWeb',                      'Fortinet FortiWeb web application firewall (WAF).',                                              false, 'fortigate.svg',     'Device',          'firewall-fortiweb',          true),
    ('SOPHOS_XG',       'Sophos XG',                     'Sophos XG firewall — anti-malware, web filtering, IPS, VPN and reporting.',                       false, 'sophosxg.svg',      'Device',          'firewall-sophos-xg',         true),
    ('PALO_ALTO',       'Palo Alto',                     'Palo Alto Networks next-generation firewalls — application and threat inspection.',               false, 'palo-alto.svg',     'Device',          'firewall-paloalto',          true),
    ('SONIC_WALL',      'SonicWall',                     'SonicWall next-generation firewalls (NGFW).',                                                    false, 'sonicwall.svg',     'Device',          'firewall-sonicwall',         true),
    ('PFSENSE',         'Pfsense',                       'pfSense open-source firewall and router with unified threat management.',                         false, 'pfsense.svg',       'Device',          'firewall-pfsense',           true),
    ('MIKROTIK',        'MikroTik',                      'MikroTik RouterOS routing, switching, firewall and wireless equipment.',                          false, 'mikrotik.svg',      'Device',          'firewall-mikrotik',          true),
    ('AIX',             'IBM AIX',                       'IBM AIX proprietary Unix operating systems for IBM platforms.',                                 false, 'aix.svg',           'Device',          'ibm-aix',                    true),
    ('AS_400',          'AS/400',                        'IBM AS/400 (IBM i) midrange systems running OS/400.',                                            false, 'ibm-as-400.svg',    'Device',          'ibm-as400',                  true),
    ('ORACLE',          'Oracle',                        'Oracle Database multi-model database management system.',                                        false, 'oracle.svg',        'Device',          'oracle',                     true),
    ('SURICATA',        'Suricata',                      'Suricata open-source intrusion detection and prevention system (IDS/IPS).',                       false, 'suricata.png',      'Device',          'suricata',                   true),
    ('DECEPTIVE_BYTES', 'Deceptive Bytes',               'Deceptive Bytes active endpoint deception platform.',                                            false, 'deceptive-b.svg',   'Other',           'deceptive-bytes',            true),
    ('GITHUB',          'GitHub',                        'GitHub source-hosting and version control — audit log ingestion.',                               false, 'github.svg',        'Other',           'github',                     true),
    ('UTMSTACK',        'Utmstack',                      'UTMStack open-source SIEM and XDR self-telemetry.',                                              false, 'utmstack.png',      'Device',          'utmstack',                   true)
ON CONFLICT (module_name) DO NOTHING;

-- Seed the system index patterns. Uses v11- prefix directly (no UPDATE step).
-- ON CONFLICT (id) DO NOTHING: on in-place upgrades the existing rows are
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

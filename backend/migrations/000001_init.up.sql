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
    ('correlation.read',  'List and view correlation rules, regex patterns, data types and tenant config',      'correlation', 'read'),
    ('correlation.write', 'Create, update and delete correlation rules, regex patterns, data types and tenant config', 'correlation', 'write'),
    ('integrations.read',  'List integrations, their tenants and configuration',          'integrations', 'read'),
    ('integrations.write', 'Activate integrations and create/update/delete their tenants', 'integrations', 'write')
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

-- Seed the system-owned regex patterns the correlation engine relies on.
-- Ported from legacy Liquibase changeset 20250616001_insert_utm_regex_pattern.xml.
--
-- These are the canonical named patterns ({{.ipv4}}, {{.syslogDate}}, ...) the
-- parsing filters reference by name; the correlation module exposes them through
-- a CRUD with system_owner protection, so they must pre-exist. All rows are
-- system_owner=true and cannot be modified or deleted via the API.
--
-- On an in-place upgrade the legacy rows already exist (utm_regex_pattern is not
-- dropped), so this is a no-op via ON CONFLICT (id). On a fresh install GORM
-- creates an empty table and this seeds it. We keep the legacy explicit ids so
-- in-place upgrades match on conflict, then advance the GORM-created sequence to
-- MAX(id) so the next user-created pattern does not collide.
INSERT INTO utm_regex_pattern (id, pattern_id, pattern_description, pattern_definition, system_owner) VALUES
    (1,  'ciscoMacAddr',   'Matches with CISCO MAC address',                                                                                               '(?:(?:[A-Fa-f0-9]{4}\.){2}[A-Fa-f0-9]{4})',                                                                                                                                                                                                                                                                           true),
    (2,  'syslogDate',     'Matches with syslog date format. Ex: Jun 16 12:34:56',                                                                         '[A-Z][a-z]{2} \d{1,2} \d{2}:\d{2}:\d{2}',                                                                                                                                                                                                                                                                            true),
    (3,  'winMacAddr',     'Matches with windows MAC address',                                                                                             '(?:(?:[A-Fa-f0-9]{2}-){5}[A-Fa-f0-9]{2})',                                                                                                                                                                                                                                                                           true),
    (4,  'commonMacAddr',  'Matches with common MAC address',                                                                                              '(?:(?:[A-Fa-f0-9]{2}:){5}[A-Fa-f0-9]{2})|(?:(?:[A-Fa-f0-9]{2}-){5}[A-Fa-f0-9]{2})',                                                                                                                                                                                                                               true),
    (5,  'integer',        'Matches with groups of signed or unsigned numbers. Ex: 0, 54, +23, -11.',                                                      '(?:[+-]?(?:[0-9]+))',                                                                                                                                                                                                                                                                                                  true),
    (6,  'day',            'Matches with any known variant of day name. Ex: Monday, Mon',                                                                  '(?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?)',                                                                                                                                                                                                                     true),
    (7,  'word',           'Matches with complete words without spaces, a word can contain _, -.',                                                          '\b\w+\b',                                                                                                                                                                                                                                                                                                               true),
    (8,  'greedy',         'Matches with the full string',                                                                                                  '.*',                                                                                                                                                                                                                                                                                                                    true),
    (9,  'space',          'Matches with one or more spaces',                                                                                              '\s+',                                                                                                                                                                                                                                                                                                                  true),
    (10, 'notSpace',       'Matches with one or more non spaces',                                                                                          '\S+',                                                                                                                                                                                                                                                                                                                  true),
    (11, 'monthName',      'Matches with any known variant of month name. Ex: January, Feb, feb',                                                          '\b(?:[Jj]an(?:uary|uar)?|[Ff]eb(?:ruary|ruar)?|[Mm](?:a|ä)?r(?:ch|z)?|[Aa]pr(?:il)?|[Mm]a(?:y|i)?|[Jj]un(?:e|i)?|[Jj]ul(?:y|i)?|[Aa]ug(?:ust)?|[Ss]ep(?:tember)?|[Oo](?:c|k)?t(?:ober)?|[Nn]ov(?:ember)?|[Dd]e(?:c|z)(?:ember)?)\b',                                                                              true),
    (12, 'ipv4',           'Matches with IPv4 address',                                                                                                    '(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\\.)){3}((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)))',                                                                                                                                                                                                                    true),
    (13, 'email',          'Matches with email address',                                                                                                   '((?P<name>[a-zA-Z0-9.!#$%&''*+/=?^_`{|}~-]+)@(?P<domain>[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*))',                                                                                                                                                    true),
    (14, 'domain',         'Matches with a domain server',                                                                                                 '((?:[_a-z0-9](?:[_a-z0-9-]{0,61}[a-z0-9])?\.)+(?:[a-z](?:[a-z0-9-]{0,61}[a-z0-9])?)?)',                                                                                                                                                                                                                          true),
    (15, 'hostname',       'Matches with a hostname',                                                                                                      '(\b(?:[0-9A-Za-z][0-9A-Za-z-]{0,62})(?:\.(?:[0-9A-Za-z][0-9A-Za-z-]{0,62}))*(\.?|\b))',                                                                                                                                                                                                                          true),
    (16, 'data',           'Matches with the full string until reaches the next pattern',                                                                  '(.*?)',                                                                                                                                                                                                                                                                                                                 true),
    (17, 'ipv6',           'Matches with IPv6 address',                                                                                                    '([0-9a-fA-F]{1,4}(:[0-9a-fA-F]{0,4}){1,7}|::[0-1]?)',                                                                                                                                                                                                                                                               true),
    (18, 'uuid',           'Matches with UUID values. Ex: 550e8400-e29b-41d4-a716-446655440000',                                                           '([A-Fa-f0-9]{8}-(?:[A-Fa-f0-9]{4}-){3}[A-Fa-f0-9]{12})',                                                                                                                                                                                                                                                         true),
    (19, 'monthNumber',    'Matches with the month number. Ex: 01,10,12',                                                                                  '(?:0[1-9]|1[0-2])',                                                                                                                                                                                                                                                                                                 true),
    (20, 'monthDay',       'Matches with day number within a month. Ex: 01, 1, 14, 31',                                                                   '(?:(?:0[1-9])|(?:[12][0-9])|(?:3[01])|[1-9])',                                                                                                                                                                                                                                                                    true),
    (21, 'year',           'Matches with any year value between 1000 and 9999',                                                                            '(([1-9])[0-9]{1,3})',                                                                                                                                                                                                                                                                                               true),
    (22, 'hour',           'Matches with H24 hour format. Ex: 07, 18, 23.',                                                                               '(([01][0-9])|2[0-4])',                                                                                                                                                                                                                                                                                              true),
    (23, 'minute',         'Matches with mm minute format. Ex: 02, 10, 59.',                                                                              '(?:[0-5][0-9])',                                                                                                                                                                                                                                                                                                     true),
    (24, 'seconds',        'Matches with SS; SS.sss; SS:sss; SS,sss seconds with milliseconds (optional). Ex: 00.000, 12.1000000.',                        '(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)',                                                                                                                                                                                                                                                                           true),
    (25, 'time',           'Matches with full time part of a timestamp as: H24:mm:SS',                                                                    '((([01][0-9])|2[0-4]):(?:[0-5][0-9])(?::(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)))',                                                                                                                                                                                                                                true),
    (26, 'iso8601Timezone','Matches with ISO8601 timezone standard. Ex: Z or -07:00 or +07:00 or -0700 or +0700',                                         '(Z|([+-](([01][0-9])|2[0-4]):?([0-5][0-9])))',                                                                                                                                                                                                                                                                     true)
ON CONFLICT (id) DO NOTHING;

-- Advance the GORM-created sequence past the seeded ids (COALESCE guards the
-- impossible empty-table case). Uses pg_get_serial_sequence so it works whatever
-- the sequence is named.
SELECT setval(pg_get_serial_sequence('utm_regex_pattern', 'id'), (SELECT COALESCE(MAX(id), 1) FROM utm_regex_pattern));

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

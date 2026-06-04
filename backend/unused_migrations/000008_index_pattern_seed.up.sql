-- Seed utm_index_pattern with the 51 system rows from the Java data.sql.
-- Idempotent: ON CONFLICT (id) DO UPDATE so re-running on a populated table is safe.
-- All rows have pattern_system = true.
-- Active rows: 1, 2, 8, 19, 39, 60, 61, 62.  All others: false.

INSERT INTO utm_index_pattern (id, pattern, pattern_module, pattern_system, is_active)
VALUES
    (1,  'log-*',                              NULL,                                                                         true, true),
    (2,  'alert-*',                            NULL,                                                                         true, true),
    (4,  'log-netflow-*',                      'NETFLOW',                                                                    true, false),
    (8,  'log-wineventlog-*',                  'WINDOWS_AGENT',                                                              true, true),
    (10, 'log-aws-*',                          'AWS_IAM_USER',                                                               true, false),
    (11, 'log-azure-*',                        'AZURE',                                                                      true, false),
    (12, 'log-o365-*',                         'O365',                                                                       true, false),
    (13, 'log-firewall-meraki-*',              'MERAKI',                                                                     true, false),
    (14, 'log-firewall-*',                     'MERAKI,SOPHOS_XG,CISCO,FORTIGATE,FIRE_POWER,UFW,MIKROTIK,PALO_ALTO,SONIC_WALL', true, false),
    (15, 'log-firewall-cisco-asa-*',           'CISCO',                                                                      true, false),
    (17, 'log-firewall-sophos-xg-*',           'SOPHOS_XG',                                                                  true, false),
    (18, 'log-iis-*',                          'IIS',                                                                        true, false),
    (19, 'log-generic-*',                      NULL,                                                                         true, true),
    (21, 'log-firewall-fortigate-traffic-*',   'FORTIGATE',                                                                  true, false),
    (24, 'log-vmware-esxi-*',                  'VMWARE',                                                                     true, false),
    (25, 'log-google-*',                       'GCP',                                                                        true, false),
    (26, 'log-firewall-cisco-firepower-*',     'FIRE_POWER',                                                                 true, false),
    (27, 'log-redis-*',                        'REDIS',                                                                      true, false),
    (28, 'log-postgresql-*',                   'POSTGRESQL',                                                                 true, false),
    (29, 'log-osquery-*',                      'OSQUERY',                                                                    true, false),
    (30, 'log-nginx-*',                        'NGINX',                                                                      true, false),
    (31, 'log-mysql-*',                        'MYSQL',                                                                      true, false),
    (32, 'log-mongodb-*',                      'MONGODB',                                                                    true, false),
    (33, 'log-logstash-*',                     'LOGSTASH',                                                                   true, false),
    (34, 'log-kibana-*',                       'KIBANA',                                                                     true, false),
    (35, 'log-kafka-*',                        'KAFKA',                                                                      true, false),
    (36, 'log-elasticsearch-*',                'ELASTICSEARCH',                                                              true, false),
    (37, 'log-auditd-*',                       'AUDITD',                                                                     true, false),
    (38, 'log-apache*',                        'APACHE',                                                                     true, false),
    (39, 'log-linux-*',                        'LINUX_AGENT',                                                                true, true),
    (40, 'log-antivirus-*',                    'ESET,SENTINEL_ONE,KASPERSKY',                                                true, false),
    (41, 'log-antivirus-esmc-eset-*',          'ESET',                                                                       true, false),
    (42, 'log-antivirus-kaspersky-*',          'KASPERSKY',                                                                  true, false),
    (43, 'log-antivirus-sentinel-one-*',       'SENTINEL_ONE',                                                               true, false),
    (44, 'log-sophos-central-*',               'SOPHOS',                                                                     true, false),
    (45, 'log-github-*',                       'GITHUB',                                                                     true, false),
    (46, 'log-firewall-ufw-*',                 'UFW',                                                                        true, false),
    (47, 'log-macos-*',                        'MACOS',                                                                      true, false),
    (48, 'log-firewall-mikrotik-*',            'MIKROTIK',                                                                   true, false),
    (49, 'log-firewall-paloalto-*',            'PALO_ALTO',                                                                  true, false),
    (50, 'log-cisco-switch-*',                 'CISCO_SWITCH',                                                               true, false),
    (51, 'log-firewall-sonicwall-*',           'SONIC_WALL',                                                                 true, false),
    (52, 'log-deceptive-bytes-*',              'DECEPTIVE_BYTES',                                                            true, false),
    (53, 'log-antivirus-bitdefender-gz-*',     'BITDEFENDER',                                                                true, false),
    (56, 'soc-ai',                             'SOC_AI',                                                                     true, false),
    (57, 'log-haproxy-*',                      'HAPROXY',                                                                    true, false),
    (58, 'log-nats-*',                         'NATS',                                                                       true, false),
    (59, 'log-traefik-*',                      'TRAEFIK',                                                                    true, false),
    (60, 'log-json-input-*',                   'JSON',                                                                       true, true),
    (61, 'log-rsyslog-linux-*',                'LINUX_LOGS',                                                                 true, true),
    (62, 'log-syslog-*',                       'SYSLOG',                                                                     true, true)
ON CONFLICT (id) DO UPDATE
    SET pattern        = EXCLUDED.pattern,
        pattern_module = EXCLUDED.pattern_module;

-- Apply the Java Liquibase changeset 20241227001: prepend 'v11-' to all
-- system patterns (same transformation Java does after initial seed).
-- Result: 'alert-*' → 'v11-alert-*', 'log-*' → 'v11-log-*', etc.
-- 'soc-ai' (no glob, no 'log-' prefix) is also prefixed: 'soc-ai' → 'v11-soc-ai'.
UPDATE utm_index_pattern
SET pattern = CONCAT('v11-', pattern)
WHERE pattern_system = true
  AND pattern NOT LIKE 'v11-%';

-- Reset sequence above the seed range so new POST inserts start at 10000,
-- never colliding with the system rows occupying ids 1..62.
SELECT setval('utm_index_pattern_id_seq', 10000);

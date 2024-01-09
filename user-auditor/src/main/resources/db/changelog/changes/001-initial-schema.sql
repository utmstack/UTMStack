create table utm_user_source
(
    id            bigint not null
        primary key,
    created_date  timestamp,
    modified_date timestamp,
    index_name    varchar(255),
    index_pattern varchar(255),
    active        boolean
);

INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (1, now(), null, 'WINDOWS_AGENT', 'log-wineventlog-*', true);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (2, now(), null, 'SYSLOG', 'log-syslog-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (3, now(), null, 'VMWARE', 'log-vmware-esxi-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (4, now(), null, 'LINUX_AGENT', 'log-linux-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (5, now(), null, 'AUDITD', 'log-auditd-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (6, now(), null, 'ELASTICSEARCH', 'log-elasticsearch-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (7, now(), null, 'HAPROXY', 'log-haproxy-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (8, now(), null, 'KAFKA', 'log-kafka-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (9, now(), null, 'KIBANA', 'log-kibana-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (10, now(), null, 'LOGSTASH', 'log-logstash-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (11, now(), null, 'MONGODB', 'log-mongodb-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (12, now(), null, 'MYSQL', 'log-mysql-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (13, now(), null, 'NATS', 'log-nats-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (14, now(), null, 'NGINX', 'log-nginx-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (15, now(), null, 'OSQUERY', 'log-osquery-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (16, now(), null, 'POSTGRESQL', 'log-postgresql-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (17, now(), null, 'REDIS', 'log-redis-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (18, now(), null, 'TRAEFIK', 'log-traefik-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (19, now(), null, 'FIRE_POWER', 'log-firewall-cisco-asa-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (20, now(), null, 'MERAKI', 'log-firewall-meraki-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (21, now(), null, 'JSON', 'log-json-input-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (22, now(), null, 'IIS', 'log-iis-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (23, now(), null, 'KASPERSKY', 'log-antivirus-kaspersky-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (24, now(), null, 'ESET', 'log-antivirus-esmc-eset-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (25, now(), null, 'SENTINEL_ONE', 'log-antivirus-sentinel-one-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (26, now(), null, 'SOPHOS_XG', 'log-firewall-sophos-xg-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (27, now(), null, 'MACOS', 'log-macos-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (28, now(), null, 'AZURE', 'log-azure-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (29, now(), null, 'O365', 'log-o365-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (30, now(), null, 'AWS_IAM_USER', 'log-aws-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (31, now(), null, 'SOPHOS', 'log-sophos-central-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (32, now(), null, 'GCP', 'log-google-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (33, now(), null, 'MIKROTIK', 'log-firewall-mikrotik-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (34, now(), null, 'PALO_ALTO', 'log-firewall-paloalto-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (35, now(), null, 'CISCO_SWITCH', 'log-cisco-switch-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (36, now(), null, 'SONIC_WALL', 'log-firewall-sonicwall-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (37, now(), null, 'DECEPTIVE_BYTES', 'log-deceptive-bytes-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (38, now(), null, 'GITHUB', 'log-github-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (39, now(), null, 'FORTIGATE', 'log-firewall-fortigate-traffic-*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (40, now(), null, 'APACHE & APACHE2', 'log-apache*', false);
INSERT INTO public.utm_user_source (id, created_date, modified_date, index_name, index_pattern, active) VALUES (41, now(), null, 'BITDEFENDER', 'log-antivirus-bitdefender-gz-*', false);

create table utm_source_filter
(
    id              bigint not null
        constraint utm_source_query_pkey
            primary key,
    created_date    timestamp,
    modified_date   timestamp,
    field           varchar(255),
    value           varchar(255),
    user_sources_id bigint
        constraint f_key_utm_user_source_filter
            references utm_user_source,
    operator        numeric,
    active          boolean default true
);

INSERT INTO public.utm_source_filter (id, created_date, modified_date, field, value, user_sources_id, operator, active) VALUES (1, now(), null, 'logx.wineventlog.event_id', '4624', 1, 0, true);
INSERT INTO public.utm_source_filter (id, created_date, modified_date, field, value, user_sources_id, operator, active) VALUES (2, now(), null, 'logx.wineventlog.event_id', '4726', 1, 0, true);
INSERT INTO public.utm_source_filter (id, created_date, modified_date, field, value, user_sources_id, operator, active) VALUES (3, now(), null, 'logx.wineventlog.event_id', '4720', 1, 0, true);


create table utm_source_scan
(
    id             bigint not null
        primary key,
    active         boolean default true,
    last_execution_date date,
    created_date   timestamp,
    modified_date  timestamp,
    user_sources_id bigint
        constraint fk_utm_user_source_scan
            references utm_user_source
);

create table utm_user
(
    id             bigint not null
        primary key,
    created_date   timestamp,
    modified_date  timestamp,
    user_sources_id bigint
        constraint f_key_utm_user_source_scan_user
            references utm_user_source,
    active         boolean default true,
    name           varchar(255),
    sid            varchar(255)
);

create table utm_user_attribute
(
    id              bigint not null
        primary key,
    active          boolean default true,
    attribute_key   varchar(255),
    attribute_value varchar(255),
    created_date    timestamp,
    modified_date   timestamp,
    user_id         bigint
        constraint fk_utm_user_attribute
            references utm_user
);
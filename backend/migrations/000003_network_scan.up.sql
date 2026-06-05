CREATE SEQUENCE IF NOT EXISTS public.utm_asset_group_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE SEQUENCE IF NOT EXISTS public.utm_asset_types_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE SEQUENCE IF NOT EXISTS public.utm_network_scan_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE SEQUENCE IF NOT EXISTS public.utm_ports_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE IF NOT EXISTS public.utm_asset_group
(
    id                bigint DEFAULT nextval('public.utm_asset_group_id_seq'::regclass) NOT NULL,
    group_name        character varying(100) NOT NULL,
    group_description character varying(255),
    created_date      timestamp without time zone,
    CONSTRAINT pk_utm_asset_group PRIMARY KEY (id),
    CONSTRAINT ux_group_name UNIQUE (group_name)
);

CREATE TABLE IF NOT EXISTS public.utm_asset_types
(
    id        bigint DEFAULT nextval('public.utm_asset_types_id_seq'::regclass) NOT NULL,
    type_name character varying(100),
    CONSTRAINT pk_utm_asset_tags PRIMARY KEY (id),
    CONSTRAINT ux_utm_asset_tags_tag_name UNIQUE (type_name)
);

CREATE TABLE IF NOT EXISTS public.utm_network_scan
(
    id                     bigint DEFAULT nextval('public.utm_network_scan_id_seq'::regclass) NOT NULL,
    asset_ip               character varying(255),
    asset_addresses        text,
    asset_mac              character varying(255),
    asset_os               character varying(255),
    asset_name             character varying(255),
    asset_aliases          text,
    asset_alive            boolean,
    asset_status           character varying(255),
    asset_type_id          bigint,
    discovered_at          timestamp without time zone,
    modified_at            timestamp without time zone,
    asset_severity         character varying(50),
    asset_notes            text,
    asset_severity_metric  double precision,
    server_name            character varying,
    group_id               bigint,
    registered_mode        character varying(20),
    asset_alias            character varying(500),
    is_agent               boolean,
    register_ip            character varying(50),
    asset_os_arch          character varying(100),
    asset_os_major_version character varying(20),
    asset_os_minor_version character varying(20),
    asset_os_platform      character varying(100),
    asset_os_version       character varying(100),
    update_level           character varying(50),
    CONSTRAINT pk_utm_network_scan PRIMARY KEY (id),
    CONSTRAINT uk_asset_name UNIQUE (asset_name),
    CONSTRAINT fk_group_id FOREIGN KEY (group_id) REFERENCES public.utm_asset_group (id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS public.utm_ports
(
    id      bigint DEFAULT nextval('public.utm_ports_id_seq'::regclass) NOT NULL,
    scan_id bigint,
    port    integer,
    tcp     character varying(255),
    udp     character varying(255),
    CONSTRAINT pk_utm_ports PRIMARY KEY (id),
    CONSTRAINT fk_network_scan_id FOREIGN KEY (scan_id) REFERENCES public.utm_network_scan (id) ON DELETE CASCADE
);

-- Tables needed for the data-source sync job. These are conceptually owned by the
-- datainput module but no Go migration provides them yet; creating here with IF NOT
-- EXISTS keeps existing prod DBs (already migrated from the Java backend) untouched.
CREATE TABLE IF NOT EXISTS public.utm_data_input_status
(
    source      character varying(256) NOT NULL,
    data_type   character varying(50)  NOT NULL,
    "timestamp" bigint                 NOT NULL,
    median      bigint,
    id          character varying(300) NOT NULL,
    CONSTRAINT pk_utm_data_input_status PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.utm_data_input_status_checkpoint
(
    id                       bigserial PRIMARY KEY,
    last_processed_timestamp timestamp with time zone NOT NULL DEFAULT NOW() - INTERVAL '12 hours'
);

INSERT INTO public.utm_data_input_status_checkpoint (id, last_processed_timestamp)
VALUES (1, NOW() - INTERVAL '12 hours')
ON CONFLICT (id) DO NOTHING;

ALTER TABLE IF EXISTS public.utm_menu DROP CONSTRAINT IF EXISTS fk_module_id CASCADE;
ALTER TABLE IF EXISTS public.utm_menu DROP COLUMN IF EXISTS module_id;

DROP TABLE IF EXISTS public.utm_ad_notification_conf;
DROP TABLE IF EXISTS public.utm_ad_report;
DROP TABLE IF EXISTS public.utm_ad_tracker;
DROP TABLE IF EXISTS public.utm_app_module;
DROP TABLE IF EXISTS public.utm_application_update_components;
DROP TABLE IF EXISTS public.utm_application_update_info;
DROP TABLE IF EXISTS public.utm_category_description;
DROP TABLE IF EXISTS public.utm_gvm_scan_result;
DROP TABLE IF EXISTS public.utm_gvm_task;
DROP TABLE IF EXISTS public.utm_module_modal;
DROP TABLE IF EXISTS public.utm_system_restart;

CREATE TABLE IF NOT EXISTS public.jhi_authority
(
    name      character varying(50) NOT NULL,
    name_show character varying(50),
    CONSTRAINT pk_jhi_authority PRIMARY KEY (name)
);

CREATE TABLE IF NOT EXISTS public.jhi_persistent_audit_event
(
    event_id   bigint DEFAULT nextval('public.jhi_persistent_audit_event_event_id_seq'::regclass) NOT NULL,
    principal  character varying(50)                                                              NOT NULL,
    event_date timestamp without time zone,
    event_type character varying(255),
    CONSTRAINT pk_jhi_persistent_audit_event PRIMARY KEY (event_id)
);
CREATE INDEX IF NOT EXISTS idx_persistent_audit_event ON public.jhi_persistent_audit_event USING btree (principal, event_date);

CREATE TABLE IF NOT EXISTS public.jhi_persistent_audit_evt_data
(
    event_id bigint                 NOT NULL,
    name     character varying(150) NOT NULL,
    value    character varying(255),
    CONSTRAINT jhi_persistent_audit_evt_data_pkey PRIMARY KEY (event_id, name),
    CONSTRAINT fk_evt_pers_audit_evt_data FOREIGN KEY (event_id) REFERENCES public.jhi_persistent_audit_event (event_id)
);
CREATE INDEX IF NOT EXISTS idx_persistent_audit_evt_data ON public.jhi_persistent_audit_evt_data USING btree (event_id);

CREATE TABLE IF NOT EXISTS public.jhi_user
(
    id                 bigint DEFAULT nextval('public.jhi_user_id_seq'::regclass) NOT NULL,
    login              character varying(50)                                      NOT NULL,
    password_hash      character varying(60)                                      NOT NULL,
    first_name         character varying(50),
    last_name          character varying(50),
    email              character varying(191),
    image_url          character varying(256),
    activated          boolean                                                    NOT NULL,
    lang_key           character varying(6),
    activation_key     character varying(20),
    reset_key          character varying(20),
    created_by         character varying(50)                                      NOT NULL,
    created_date       timestamp without time zone,
    reset_date         timestamp without time zone,
    last_modified_by   character varying(50),
    last_modified_date timestamp without time zone,
    openvas_user_uuid  character varying(50),
    openvas_user_id    bigint,
    tfa_secret         character varying(100),
    fs_manager         boolean,
    default_password   boolean,
    CONSTRAINT pk_jhi_user PRIMARY KEY (id),
    CONSTRAINT ux_user_email UNIQUE (email),
    CONSTRAINT ux_user_login UNIQUE (login)
);

CREATE TABLE IF NOT EXISTS public.jhi_user_authority
(
    user_id        bigint                NOT NULL,
    authority_name character varying(50) NOT NULL,
    CONSTRAINT jhi_user_authority_pkey PRIMARY KEY (user_id, authority_name),
    CONSTRAINT fk_authority_name FOREIGN KEY (authority_name) REFERENCES public.jhi_authority ("name") ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES public.jhi_user (id)
);

CREATE TABLE IF NOT EXISTS public.utm_index_pattern
(
    id             bigint DEFAULT nextval('public.utm_index_pattern_id_seq'::regclass) NOT NULL,
    pattern        character varying(100)                                              NOT NULL,
    pattern_module character varying(500),
    pattern_system boolean,
    is_active      boolean,
    CONSTRAINT pk_utm_index_pattern PRIMARY KEY (id),
    CONSTRAINT utm_index_pattern_pattern_key UNIQUE (pattern)
);

CREATE TABLE IF NOT EXISTS public.utm_logstash_filter_group
(
    id                bigint DEFAULT nextval('public.utm_logstash_filter_group_id_seq'::regclass) NOT NULL,
    group_name        character varying(150)                                                      NOT NULL,
    group_description text,
    system_owner      boolean,
    CONSTRAINT pk_utm_logstash_filter_group PRIMARY KEY (id),
    CONSTRAINT ux_logstash_filter_group_name UNIQUE (group_name)
);

CREATE TABLE IF NOT EXISTS public.utm_logstash_filter
(
    id              bigint DEFAULT nextval('public.utm_logstash_filter_id_seq'::regclass) NOT NULL,
    logstash_filter text,
    filter_name     character varying(100),
    filter_group_id bigint,
    system_owner    boolean,
    module_name     character varying(500),
    is_active       boolean,
    filter_version  character varying(100),
    CONSTRAINT pk_utm_logstash_filter PRIMARY KEY (id),
    CONSTRAINT fk_filter_group_id FOREIGN KEY (filter_group_id) REFERENCES public.utm_logstash_filter_group (id) ON DELETE SET NULL
);

ALTER TABLE public.utm_logstash_filter ADD COLUMN IF NOT EXISTS filter_version character varying(100);

CREATE TABLE IF NOT EXISTS public.utm_dashboard
(
    id             bigint DEFAULT nextval('public.utm_dashboard_id_seq'::regclass) NOT NULL,
    name           character varying(100)                                          NOT NULL,
    description    character varying(255),
    created_date   timestamp without time zone                                     NOT NULL,
    modified_date  timestamp without time zone,
    user_created   character varying(50)                                           NOT NULL,
    user_modified  character varying(50),
    filters        text,
    refresh_time   bigint,
    dashboard_type character varying(100),
    system_owner   boolean,
    CONSTRAINT pk_utm_dashboard PRIMARY KEY (id),
    CONSTRAINT uk_dash_name_sys_owner UNIQUE (name, system_owner)
);

CREATE TABLE IF NOT EXISTS public.utm_visualization
(
    id            bigint DEFAULT nextval('public.utm_visualization_id_seq'::regclass) NOT NULL,
    name          character varying(100)                                              NOT NULL,
    description   character varying(255),
    created_date  timestamp without time zone                                         NOT NULL,
    modified_date timestamp without time zone,
    user_created  character varying(50)                                               NOT NULL,
    user_modified character varying(50),
    query         text,
    chart_config  text,
    chart_type    character varying(50),
    filters       text,
    id_pattern    bigint,
    aggregation   text,
    event_type    character varying(50),
    chart_action  text,
    system_owner  boolean,
    CONSTRAINT pk_utm_visualization PRIMARY KEY (id),
    CONSTRAINT uk_vis_name_ch_type_sys_owner UNIQUE (name, chart_type, system_owner),
    CONSTRAINT fk_id_pattern FOREIGN KEY (id_pattern) REFERENCES public.utm_index_pattern(id) ON DELETE CASCADE ON UPDATE CASCADE

);

CREATE TABLE IF NOT EXISTS public.utm_agent_manager
(
    id              bigint DEFAULT nextval('public.utm_agent_manager_id_seq'::regclass) NOT NULL,
    manager_host    character varying(100)                                              NOT NULL,
    manager_version character varying(50)                                               NOT NULL,
    manager_port    integer                                                             NOT NULL,
    CONSTRAINT utm_agent_manager_pkey PRIMARY KEY (id),
    CONSTRAINT ux_utm_agent_manager_manager_host UNIQUE (manager_host)
);

CREATE TABLE IF NOT EXISTS public.utm_alert_last
(
    id                   bigint NOT NULL,
    last_alert_timestamp timestamp without time zone,
    CONSTRAINT pk_utm_alert_last PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.utm_alert_log
(
    id            bigint DEFAULT nextval('public.utm_alert_log_id_seq'::regclass) NOT NULL,
    alert_id      character varying(100)                                          NOT NULL,
    log_user      character varying(50)                                           NOT NULL,
    log_date      timestamp without time zone                                     NOT NULL,
    log_action    character varying(255)                                          NOT NULL,
    log_old_value text,
    log_new_value text,
    log_message   text,
    CONSTRAINT pk_utm_alert_log PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.utm_alert_response_rule
(
    id                 bigint DEFAULT nextval('public.utm_alert_response_rule_id_seq'::regclass) NOT NULL,
    rule_name          character varying(150)                                                    NOT NULL,
    rule_description   character varying(512),
    rule_conditions    text                                                                      NOT NULL,
    rule_cmd           text                                                                      NOT NULL,
    rule_active        boolean                                                                   NOT NULL,
    agent_platform     character varying(50)                                                     NOT NULL,
    excluded_agents    text,
    created_by         character varying(50)                                                     NOT NULL,
    created_date       timestamp without time zone                                               NOT NULL,
    last_modified_by   character varying(50),
    last_modified_date timestamp without time zone,
    CONSTRAINT utm_alert_response_rule_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.utm_alert_response_rule_execution
(
    id                  bigint  DEFAULT nextval('public.utm_alert_response_rule_execution_id_seq'::regclass) NOT NULL,
    rule_id             bigint                                                                               NOT NULL,
    alert_id            character varying(100)                                                               NOT NULL,
    command             text                                                                                 NOT NULL,
    command_result      text,
    agent               character varying(150)                                                               NOT NULL,
    execution_date      timestamp without time zone                                                          NOT NULL,
    execution_status    character varying(100)                                                               NOT NULL,
    non_execution_cause character varying(100),
    execution_retries   integer DEFAULT 0,
    CONSTRAINT utm_alert_response_rule_execution_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.utm_alert_response_rule_history
(
    id             bigint DEFAULT nextval('public.utm_alert_response_rule_history_id_seq'::regclass) NOT NULL,
    rule_id        bigint                                                                            NOT NULL,
    created_date   timestamp without time zone                                                       NOT NULL,
    previous_state text                                                                              NOT NULL,
    CONSTRAINT utm_alert_response_rule_history_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.utm_alert_socai_processing_request
(
    alert_id character varying(255) NOT NULL,
    CONSTRAINT utm_alert_socai_processing_request_pkey PRIMARY KEY (alert_id)
);

CREATE TABLE IF NOT EXISTS public.utm_alert_tag
(
    id           bigint DEFAULT nextval('public.utm_alert_tag_id_seq'::regclass) NOT NULL,
    tag_name     character varying(50)                                           NOT NULL,
    tag_color    character varying(15),
    system_owner boolean,
    CONSTRAINT pk_utm_alert_category PRIMARY KEY (id),
    CONSTRAINT uk_tag_name UNIQUE (tag_name)
);

CREATE TABLE IF NOT EXISTS public.utm_alert_tag_rule
(
    id                 bigint DEFAULT nextval('public.utm_alert_tag_rule_id_seq'::regclass) NOT NULL,
    rule_name          character varying(255)                                               NOT NULL,
    rule_description   text                                                                 NOT NULL,
    rule_conditions    text                                                                 NOT NULL,
    rule_applied_tags  character varying(255)                                               NOT NULL,
    created_by         character varying(50)                                                NOT NULL,
    created_date       timestamp without time zone,
    last_modified_by   character varying(50),
    last_modified_date timestamp without time zone,
    rule_active        boolean                                                              NOT NULL,
    rule_deleted       boolean                                                              NOT NULL,
    CONSTRAINT utm_alert_tag_rule_pkey PRIMARY KEY (id),
    CONSTRAINT ux_rule_name UNIQUE (rule_name)
);

CREATE TABLE IF NOT EXISTS public.utm_asset_group
(
    id                bigint DEFAULT nextval('public.utm_asset_group_id_seq'::regclass) NOT NULL,
    group_name        character varying(100)                                            NOT NULL,
    group_description character varying(255),
    created_date      timestamp without time zone,
    CONSTRAINT pk_utm_asset_group PRIMARY KEY (id),
    CONSTRAINT ux_group_name UNIQUE (group_name)
);

CREATE TABLE IF NOT EXISTS public.utm_asset_metrics
(
    id         character varying(255) NOT NULL,
    asset_name character varying(255) NOT NULL,
    metric     character varying(255),
    amount     bigint,
    CONSTRAINT utm_asset_metrics_pkey PRIMARY KEY (id),
    CONSTRAINT utm_asset_metrics_un UNIQUE (asset_name, metric)
);

CREATE TABLE IF NOT EXISTS public.utm_asset_types
(
    id        bigint DEFAULT nextval('public.utm_asset_types_id_seq'::regclass) NOT NULL,
    type_name character varying(100),
    CONSTRAINT pk_utm_asset_tags PRIMARY KEY (id),
    CONSTRAINT ux_utm_asset_tags_tag_name UNIQUE (type_name)
);

CREATE TABLE IF NOT EXISTS public.utm_compliance_standard
(
    id                   bigint DEFAULT nextval('public.utm_compliance_standard_id_seq'::regclass) NOT NULL,
    standard_name        character varying(50),
    standard_description character varying(1000),
    system_owner         boolean,
    CONSTRAINT pk_utm_compliance_standard PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.utm_compliance_standard_section
(
    id                           bigint DEFAULT nextval('public.utm_compliance_standard_section_id_seq'::regclass) NOT NULL,
    standard_id                  bigint,
    standard_section_name        character varying(255),
    standard_section_description character varying(1000),
    CONSTRAINT pk_utm_compliance_standard_section PRIMARY KEY (id),
    CONSTRAINT fk_standard_id FOREIGN KEY (standard_id) REFERENCES public.utm_compliance_standard (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.utm_compliance_report_config
(
    id                           bigint DEFAULT nextval('public.utm_compliance_report_config_id_seq'::regclass) NOT NULL,
    config_solution              text,
    config_report_columns        text,
    config_report_req_body       text,
    config_report_req_params     text,
    config_report_resource_url   character varying(512),
    config_report_request_type   character varying(12),
    config_report_pageable       boolean,
    config_report_filter_by_time boolean,
    config_report_data_origin    character varying(50),
    config_report_export_csv_url character varying(512),
    standard_section_id          bigint,
    config_report_editable       boolean,
    dashboard_id                 bigint,
    config_type                  character varying(50),
    config_url                   character varying(255),
    CONSTRAINT pk_utm_compliance_report_config PRIMARY KEY (id),
    CONSTRAINT uk_solution_section_dashboard UNIQUE (config_solution, standard_section_id, dashboard_id),
    CONSTRAINT fk_dashboard_id FOREIGN KEY (dashboard_id) REFERENCES public.utm_dashboard (id) ON DELETE CASCADE,
    CONSTRAINT fk_standard_section_id FOREIGN KEY (standard_section_id) REFERENCES public.utm_compliance_standard_section (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.utm_configuration_section
(
    id                bigint DEFAULT nextval('public.utm_configuration_section_id_seq'::regclass) NOT NULL,
    section           character varying(255),
    description       character varying(255),
    module_name_short character varying(50),
    CONSTRAINT utm_configuration_section_pkey PRIMARY KEY (id),
    CONSTRAINT utm_configuration_section_section_module_name_short_key UNIQUE (section, module_name_short)
);

CREATE TABLE IF NOT EXISTS public.utm_configuration_parameter
(
    id                     bigint DEFAULT nextval('public.utm_configuration_parameter_id_seq'::regclass) NOT NULL,
    section_id             bigint                                                                        NOT NULL,
    conf_param_short       character varying(100),
    conf_param_large       character varying(255),
    conf_param_description character varying(255),
    conf_param_value       text,
    conf_param_required    boolean,
    conf_param_datatype    character varying(20),
    modification_time      timestamp(6) without time zone,
    modification_user      character varying(255),
    conf_param_option      text,
    CONSTRAINT utm_configuration_parameter_pkey PRIMARY KEY (id),
    CONSTRAINT utm_configuration_parameter_section_id_conf_param_short_key UNIQUE (section_id, conf_param_short),
    CONSTRAINT utm_configuration_parameter_section_id_fkey FOREIGN KEY (section_id) REFERENCES public.utm_configuration_section (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.utm_dashboard_authority
(
    id             bigint DEFAULT nextval('public.utm_dashboard_authority_id_seq'::regclass) NOT NULL,
    id_dashboard   bigint                                                                    NOT NULL,
    authority_name character varying(50)                                                     NOT NULL,
    CONSTRAINT pk_utm_dashboard_authority PRIMARY KEY (id),
    CONSTRAINT fk_dashboard_id FOREIGN KEY (id_dashboard) REFERENCES public.utm_dashboard (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.utm_dashboard_visualization
(
    id                    bigint DEFAULT nextval('public.utm_dashboard_visualization_id_seq'::regclass) NOT NULL,
    id_visualization      bigint                                                                        NOT NULL,
    id_dashboard          bigint                                                                        NOT NULL,
    dv_order              smallint,
    dv_width              double precision,
    dv_height             double precision,
    dv_top                double precision,
    dv_left               double precision,
    dv_show_time_filter   boolean,
    dv_default_time_range character varying(255),
    dv_grid_info          text,
    CONSTRAINT pk_utm_dashboard_visualization PRIMARY KEY (id),
    CONSTRAINT fk_dashboard_id FOREIGN KEY (id_dashboard) REFERENCES public.utm_dashboard (id) ON DELETE CASCADE,
    CONSTRAINT fk_visualization_id FOREIGN KEY (id_visualization) REFERENCES public.utm_visualization (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.utm_data_input_status
(
    source      character varying(256) NOT NULL,
    data_type   character varying(50)  NOT NULL,
    "timestamp" bigint                 NOT NULL,
    median      bigint,
    id          character varying(300) NOT NULL,
    CONSTRAINT pk_utm_data_input_status PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.utm_data_source_config
(
    id             uuid DEFAULT gen_random_uuid() NOT NULL,
    data_type      character varying(255),
    data_type_name character varying(255),
    system_owner   boolean,
    included       boolean,
    CONSTRAINT utm_data_source_config_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.utm_federation_service_client
(
    id              bigint DEFAULT nextval('public.utm_federation_service_client_id_seq'::regclass) NOT NULL,
    fs_client_token text,
    CONSTRAINT pk_utm_federation_service_client PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.utm_getting_started
(
    id         bigint DEFAULT nextval('public.getting_started_seq'::regclass) NOT NULL,
    step_short character varying(255)                                         NOT NULL,
    step_order integer                                                        NOT NULL,
    completed  boolean,
    CONSTRAINT utm_getting_started_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.utm_logstash_pipeline
(
    id bigint  DEFAULT nextval('public.utm_logstash_pipeline_id_seq'::regclass) NOT NULL,
    pipeline_id varchar(255) NULL,
	pipeline_name varchar(255) NULL,
	parent_pipeline int4 NULL,
	pipeline_status varchar(255) NULL,
	module_name varchar(255) NULL,
	system_owner bool NULL,
	pipeline_description varchar(2000) NULL,
	pipeline_internal bool NULL DEFAULT false,
	events_in int8 NULL DEFAULT 0,
	events_filtered int8 NULL DEFAULT 0,
	events_out int8 NULL DEFAULT 0,
	reloads_successes int8 NULL DEFAULT 0,
	reloads_failures int8 NULL DEFAULT 0,
	reloads_last_failure_timestamp varchar(50) NULL,
	reloads_last_error text NULL,
	reloads_last_success_timestamp varchar(50) NULL,
    CONSTRAINT utm_logstash_pipeline_pkey PRIMARY KEY (id),
    CONSTRAINT fk_utm_logstash_pipeline_parent FOREIGN KEY (parent_pipeline) REFERENCES public.utm_logstash_pipeline (id)
);

CREATE TABLE IF NOT EXISTS public.utm_group_logstash_pipeline_filters
(
    id          bigint DEFAULT nextval('public.utm_group_logstash_pipeline_filters_id_seq'::regclass) NOT NULL,
    filter_id   integer,
    pipeline_id integer,
    relation    character varying(50),
    CONSTRAINT utm_group_logstash_pipeline_filters_pkey PRIMARY KEY (id),
    CONSTRAINT fk_utm_logstash_filter FOREIGN KEY (filter_id) REFERENCES public.utm_logstash_filter (id),
    CONSTRAINT fk_utm_logstash_pipeline FOREIGN KEY (pipeline_id) REFERENCES public.utm_logstash_pipeline (id)
);

CREATE TABLE IF NOT EXISTS public.utm_images
(
    short_name   character varying(20) NOT NULL,
    tooltip_text character varying(255),
    system_img   text,
    user_img     text,
    CONSTRAINT pk_utm_images PRIMARY KEY (short_name)
);

CREATE TABLE IF NOT EXISTS public.utm_incident
(
    id                    bigint DEFAULT nextval('public.utm_incident_id_seq'::regclass) NOT NULL,
    incident_name         character varying(250)                                         NOT NULL,
    incident_description  character varying(2000)                                        NOT NULL,
    incident_status       character varying(255)                                         NOT NULL,
    incident_assigned_to  text,
    incident_created_date timestamp without time zone                                    NOT NULL,
    incident_severity     integer                                                        NOT NULL,
    incident_solution     character varying(1000),
    CONSTRAINT utm_incident_pkey PRIMARY KEY (id),
    CONSTRAINT ux_utm_incident_incident_name UNIQUE (incident_name)
);

ALTER TABLE public.utm_incident ALTER COLUMN incident_description TYPE character varying(2000);

CREATE TABLE IF NOT EXISTS public.utm_incident_actions
(
    id                 bigint DEFAULT nextval('public.utm_incident_actions_id_seq'::regclass) NOT NULL,
    action_command     character varying,
    action_description text,
    action_params      character varying,
    action_type        integer,
    action_editable    boolean                                                                NOT NULL,
    created_date       timestamp without time zone                                            NOT NULL,
    created_user       character varying(50)                                                  NOT NULL,
    modified_date      timestamp without time zone,
    modified_user      character varying(50),
    CONSTRAINT pk_utm_incident_actions PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.utm_incident_action_command
(
    id          bigint DEFAULT nextval('public.utm_incident_action_command_id_seq'::regclass) NOT NULL,
    action_id   bigint                                                                        NOT NULL,
    os_platform character varying(255),
    command     text,
    CONSTRAINT utm_incident_action_command_pkey PRIMARY KEY (id),
    CONSTRAINT fk_action_id FOREIGN KEY (action_id) REFERENCES public.utm_incident_actions (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.utm_incident_alert
(
    id             bigint DEFAULT nextval('public.utm_incident_alert_id_seq'::regclass) NOT NULL,
    incident_id    bigint                                                               NOT NULL,
    alert_id       character varying(255)                                               NOT NULL,
    alert_name     character varying(255)                                               NOT NULL,
    alert_status   integer                                                              NOT NULL,
    alert_severity integer                                                              NOT NULL,
    CONSTRAINT utm_incident_alert_pkey PRIMARY KEY (id),
    CONSTRAINT ux_utm_incident_alert_alert_id UNIQUE (alert_id),
    CONSTRAINT fk_utm_incident_alert_incident_id FOREIGN KEY (incident_id) REFERENCES public.utm_incident (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.utm_incident_history
(
    id                bigint DEFAULT nextval('public.utm_incident_history_id_seq'::regclass) NOT NULL,
    incident_id       bigint                                                                 NOT NULL,
    action_date       timestamp without time zone                                            NOT NULL,
    action_type       character varying(255)                                                 NOT NULL,
    action_created_by character varying(255)                                                 NOT NULL,
    action            character varying(255)                                                 NOT NULL,
    action_detail     character varying(255),
    CONSTRAINT utm_incident_history_pkey PRIMARY KEY (id),
    CONSTRAINT fk_utm_incident_history_incident_id FOREIGN KEY (incident_id) REFERENCES public.utm_incident (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.utm_incident_jobs
(
    id            bigint DEFAULT nextval('public.utm_incident_jobs_id_seq'::regclass) NOT NULL,
    action_id     bigint,
    params        character varying,
    agent         character varying,
    status        integer,
    job_result    text,
    origin_id     character varying(100)                                              NOT NULL,
    origin_type   character varying(30)                                               NOT NULL,
    created_date  timestamp without time zone                                         NOT NULL,
    created_user  character varying(50)                                               NOT NULL,
    modified_date timestamp without time zone,
    modified_user character varying(50),
    CONSTRAINT pk_utm_incident_jobs PRIMARY KEY (id),
    CONSTRAINT fk_action_id FOREIGN KEY (action_id) REFERENCES public.utm_incident_actions (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.utm_incident_note
(
    id             bigint DEFAULT nextval('public.utm_incident_note_id_seq'::regclass) NOT NULL,
    incident_id    bigint                                                              NOT NULL,
    note_text      character varying(1000)                                             NOT NULL,
    note_send_date timestamp without time zone                                         NOT NULL,
    note_send_by   character varying(255)                                              NOT NULL,
    CONSTRAINT utm_incident_note_pkey PRIMARY KEY (id),
    CONSTRAINT fk_utm_incident_note_incident_id FOREIGN KEY (incident_id) REFERENCES public.utm_incident (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.utm_server
(
    id          bigint DEFAULT nextval('public.utm_server_id_seq'::regclass) NOT NULL,
    server_name character varying(255),
    server_type character varying(255),
    CONSTRAINT utm_server_pkey PRIMARY KEY (id),
    CONSTRAINT utm_server_un UNIQUE (server_name)
);

CREATE TABLE IF NOT EXISTS public.utm_server_module
(
    id            bigint  DEFAULT nextval('public.utm_server_module_id_seq'::regclass) NOT NULL,
    server_id     integer                                                              NOT NULL,
    module_name   character varying                                                    NOT NULL,
    needs_restart boolean DEFAULT false                                                NOT NULL,
    pretty_name   character varying                                                    NOT NULL,
    CONSTRAINT utm_server_module_pk PRIMARY KEY (id),
    CONSTRAINT utm_server_module_un UNIQUE (module_name, server_id),
    CONSTRAINT utm_server_module_fk FOREIGN KEY (server_id) REFERENCES public.utm_server(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.utm_integration
(
    id                      bigint DEFAULT nextval('public.utm_integration_id_seq'::regclass) NOT NULL,
    module_id               integer,
    integration_name        character varying(255),
    integration_description character varying(255),
    url                     character varying(255),
    integration_icon_path   character varying(255),
    CONSTRAINT utm_integration_pkey PRIMARY KEY (id),
    CONSTRAINT utm_integration_un UNIQUE (module_id, integration_name),
    CONSTRAINT utm_integration_fk FOREIGN KEY (module_id) REFERENCES public.utm_server_module (id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.utm_integration_conf
(
    id               bigint DEFAULT nextval('public.utm_integration_conf_id_seq'::regclass) NOT NULL,
    integration_id   integer                                                                NOT NULL,
    conf_short       character varying(100),
    conf_large       character varying(255),
    conf_description character varying(255),
    conf_value       text,
    conf_datatype    character varying(20),
    CONSTRAINT utm_integration_conf_pkey PRIMARY KEY (id),
    CONSTRAINT utm_integration_conf_un UNIQUE (integration_id, conf_short),
    CONSTRAINT utm_integration_conf_integration_id_fkey FOREIGN KEY (integration_id) REFERENCES public.utm_integration (id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS public.utm_log_analyzer_query
(
    id                   bigint DEFAULT nextval('public.utm_log_analyzer_query_id_seq'::regclass) NOT NULL,
    la_name              character varying(255)                                                   NOT NULL,
    la_description       character varying(255),
    la_owner             character varying(50)                                                    NOT NULL,
    la_creation_date     timestamp without time zone                                              NOT NULL,
    la_modification_date timestamp without time zone,
    la_columns           text,
    la_filters           text,
    la_data_origin       character varying(50),
    id_pattern           bigint,
    CONSTRAINT pk_utm_log_analyzer_query PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.utm_logstash_input
(
    id                bigint DEFAULT nextval('public.utm_logstash_input_id_seq'::regclass) NOT NULL,
    pipeline_id       integer,
    input_pretty_name character varying(250),
    input_plugin      character varying(100),
    input_with_ssl    boolean,
    system_owner      boolean,
    CONSTRAINT utm_logstash_input_pkey PRIMARY KEY (id),
    CONSTRAINT fk_utm_logstash_input FOREIGN KEY (pipeline_id) REFERENCES public.utm_logstash_pipeline (id)
);

CREATE TABLE IF NOT EXISTS public.utm_logstash_input_configuration
(
    id                    bigint DEFAULT nextval('public.utm_logstash_input_configuration_id_seq'::regclass) NOT NULL,
    input_id              integer,
    conf_key              character varying(50),
    conf_value            character varying(255),
    conf_type             character varying(50),
    conf_required         boolean,
    conf_validation_regex character varying(400),
    system_owner          boolean,
    CONSTRAINT utm_logstash_input_configuration_pkey PRIMARY KEY (id),
    CONSTRAINT fk_utm_logstash_input FOREIGN KEY (input_id) REFERENCES public.utm_logstash_input (id)
);

CREATE TABLE IF NOT EXISTS public.utm_logstash_ports_configuration
(
    id            bigint DEFAULT nextval('public.utm_logstash_ports_configuration_id_seq'::regclass) NOT NULL,
    protocol      character varying(50),
    port_or_range character varying(50),
    system_owner  boolean,
    CONSTRAINT utm_logstash_ports_configuration_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.utm_menu
(
    id                bigint  DEFAULT nextval('public.utm_menu_id_seq'::regclass) NOT NULL,
    name              character varying(50)                                       NOT NULL,
    url               text,
    parent_id         bigint,
    type              smallint,
    dashboard_id      bigint,
    "position"        smallint,
    menu_active       boolean DEFAULT true                                        NOT NULL,
    menu_action       boolean DEFAULT false,
    menu_icon         character varying(100),
    module_name_short character varying(500),
    CONSTRAINT pk_utm_menu PRIMARY KEY (id),
    CONSTRAINT fk_dashboard_id FOREIGN KEY (dashboard_id) REFERENCES public.utm_dashboard(id) ON DELETE CASCADE,
    CONSTRAINT fk_parent_id FOREIGN KEY (parent_id) REFERENCES public.utm_menu(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.utm_menu_authority
(
    id             bigint DEFAULT nextval('public.utm_menu_authority_id_seq'::regclass) NOT NULL,
    menu_id        bigint                                                               NOT NULL,
    authority_name character varying(50)                                                NOT NULL,
    CONSTRAINT menu_authority_uk UNIQUE (menu_id, authority_name),
    CONSTRAINT pk_utm_menu_authority PRIMARY KEY (id),
    CONSTRAINT fk_menu_auth_authority_name FOREIGN KEY (authority_name) REFERENCES public.jhi_authority("name") ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_menu_auth_menu_id FOREIGN KEY (menu_id) REFERENCES public.utm_menu(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.utm_module
(
    id                 bigint DEFAULT nextval('public.utm_module_id_seq'::regclass) NOT NULL,
    pretty_name        character varying(255),
    module_description text,
    module_active      boolean,
    module_icon        character varying(255),
    module_name        character varying(50),
    server_id          bigint,
    module_category    character varying(100),
    needs_restart      boolean,
    lite_version       boolean,
    is_activatable     boolean,
    CONSTRAINT pk_utm_module PRIMARY KEY (id),
    CONSTRAINT uk_module_name_server_id UNIQUE (module_name, server_id),
    CONSTRAINT fk_server_id FOREIGN KEY (server_id) REFERENCES public.utm_server(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.utm_module_group
(
    id                bigint DEFAULT nextval('public.utm_module_group_id_seq'::regclass) NOT NULL,
    module_id         bigint                                                             NOT NULL,
    group_name        character varying(255)                                             NOT NULL,
    group_description character varying(512),
    CONSTRAINT uk_module_name_group_name UNIQUE (module_id, group_name),
    CONSTRAINT utm_module_group_pkey PRIMARY KEY (id),
    CONSTRAINT fk_module_id FOREIGN KEY (module_id) REFERENCES public.utm_module(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.utm_module_group_configuration
(
    id               bigint DEFAULT nextval('public.utm_module_group_configuration_id_seq'::regclass) NOT NULL,
    group_id         bigint                                                                           NOT NULL,
    conf_key         character varying(100)                                                           NOT NULL,
    conf_value       text,
    conf_name        character varying(255)                                                           NOT NULL,
    conf_description character varying(512),
    conf_data_type   character varying(50)                                                            NOT NULL,
    conf_required    boolean                                                                          NOT NULL,
    CONSTRAINT uk_group_id_conf_key UNIQUE (group_id, conf_key),
    CONSTRAINT utm_module_group_configuration_pkey PRIMARY KEY (id),
    CONSTRAINT fk_group_id FOREIGN KEY (group_id) REFERENCES public.utm_module_group(id) ON DELETE CASCADE ON UPDATE CASCADE

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
    CONSTRAINT fk_group_id FOREIGN KEY (group_id) REFERENCES public.utm_asset_group(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS public.utm_ports
(
    id      bigint DEFAULT nextval('public.utm_ports_id_seq'::regclass) NOT NULL,
    scan_id bigint,
    port    integer,
    tcp     character varying(255),
    udp     character varying(255),
    CONSTRAINT pk_utm_ports PRIMARY KEY (id),
    CONSTRAINT fk_network_scan_id FOREIGN KEY (scan_id) REFERENCES public.utm_network_scan(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS public.utm_report_section
(
    id                  bigint DEFAULT nextval('public.utm_report_section_id_seq'::regclass) NOT NULL,
    rep_sec_name        character varying(255),
    rep_sec_description character varying(1000)                                              NOT NULL,
    creation_user       character varying(255),
    creation_date       timestamp without time zone,
    modification_user   character varying(255),
    modification_date   timestamp without time zone,
    rep_sec_system      boolean,
    rep_sec_short_name  character varying(100),
    CONSTRAINT pk_utm_report_section PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.utm_report
(
    id                bigint DEFAULT nextval('public.utm_report_id_seq'::regclass) NOT NULL,
    rep_name          character varying(255),
    rep_description   character varying(1000),
    report_section_id bigint,
    dashboard_id      bigint,
    creation_user     character varying(255),
    creation_date     timestamp without time zone,
    modification_user character varying(255),
    modification_date timestamp without time zone,
    rep_url           character varying,
    rep_type          character varying(20),
    rep_module        character varying(100),
    rep_short_name    character varying(50),
    rep_http_method   character varying(10),
    CONSTRAINT pk_utm_report PRIMARY KEY (id),
    CONSTRAINT fk_report_dashboard_id FOREIGN KEY (dashboard_id) REFERENCES public.utm_dashboard(id) ON DELETE CASCADE,
    CONSTRAINT fk_report_section_id FOREIGN KEY (report_section_id) REFERENCES public.utm_report_section(id) ON DELETE CASCADE
);

CREATE OR REPLACE VIEW public.utm_server_configurations AS
SELECT utm_module_group_configuration.conf_key AS conf_short,
       utm_module_group_configuration.conf_value,
       utm_module.module_name,
       utm_module_group.group_name,
       utm_server.server_name,
       utm_module.needs_restart
FROM (((public.utm_server
    JOIN public.utm_module ON ((utm_server.id = utm_module.server_id)))
    JOIN public.utm_module_group ON ((utm_module.id = utm_module_group.module_id)))
    JOIN public.utm_module_group_configuration ON ((utm_module_group.id = utm_module_group_configuration.group_id)));

CREATE TABLE IF NOT EXISTS public.utm_space_notification_control
(
    id                bigint DEFAULT nextval('public.utm_space_notification_control_id_seq'::regclass) NOT NULL,
    next_notification timestamp without time zone                                                      NOT NULL,
    CONSTRAINT utm_space_notification_control_pkey PRIMARY KEY (id)
);

CREATE OR REPLACE FUNCTION public.execute_register_integration_function() RETURNS void
    LANGUAGE plpgsql
AS
$$
declare
    s record;
begin
    ---------------------------------------------------
    -- Creating server
    ---------------------------------------------------
    INSERT INTO utm_server (server_name, server_type)
    VALUES ('master', 'aio')
    ON CONFLICT (server_name) DO UPDATE SET server_type = 'aio';

    for s in (select utm_server.* from utm_server)
        loop
            perform public.register_integrations(cast(s.id as integer), s.server_type);
        end loop;
end;
$$;

CREATE OR REPLACE FUNCTION public.register_integration_ad_audit(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin

    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon",
                                      "module_name", "server_id", "needs_restart", "module_category",
                                      "lite_version", "module_description", "is_activatable")
    VALUES ('AD Audit', 'f', 'active-directory-federation-services.svg', 'AD_AUDIT', srv_id, 'f',
            'UTMStack Modules', TRUE,
            'Track and manage accounts access and permission changes. Get alerted when suspicious activity happens',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'AD Audit',
                                                      module_icon        = 'active-directory-federation-services.svg',
                                                      module_name        = 'AD_AUDIT',
                                                      module_category    = 'UTMStack Modules',
                                                      module_description = 'Track and manage accounts access and permission changes. Get alerted when suspicious activity happens',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_apache(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin

    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Apache', 'f', 'apache.svg', 'APACHE', srv_id, 'f', 'Web Server', 't',
            'As a Web server, Apache is responsible for accepting directory (HTTP) requests from Internet users and sending them their desired information in the form of files and Web pages.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Apache',
                                                      module_icon        = 'apache.svg',
                                                      module_name        = 'APACHE',
                                                      module_category    = 'Web Server',
                                                      module_description = 'As a Web server, Apache is responsible for accepting directory (HTTP) requests from Internet users and sending them their desired information in the form of files and Web pages',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_apache2(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin

    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Apache2', 'f', 'apache2.svg', 'APACHE2', srv_id, 'f', 'Web Server', 't',
            'HTTPD - Apache2 Web Server. Apache is the most commonly used Web server on Linux systems. Web servers are used to serve Web pages requested by client computers.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Apache2',
                                                      module_icon        = 'apache2.svg',
                                                      module_name        = 'APACHE2',
                                                      module_category    = 'Web Server',
                                                      module_description = 'HTTPD - Apache2 Web Server. Apache is the most commonly used Web server on Linux systems. Web servers are used to serve Web pages requested by client computers.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_aws(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('AWS Cloudwatch', 'f', 'aws-cloudtrail.svg', 'AWS_IAM_USER', srv_id, 'f', 'Cloud', 't',
            'AWS Cloudwatch enables auditing, security monitoring, and operational troubleshooting by tracking user activity and API usage. CloudTrail logs, continuously monitors, and retains account activity related to actions across your AWS infrastructure, giving you control over storage, analysis, and remediation actions.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'AWS Cloudwatch',
                                                      module_icon        = 'aws-cloudtrail.svg',
                                                      module_name        = 'AWS_IAM_USER',
                                                      module_category    = 'Cloud',
                                                      module_description = 'AWS Cloudwatch enables auditing, security monitoring, and operational troubleshooting by tracking user activity and API usage. CloudTrail logs, continuously monitors, and retains account activity related to actions across your AWS infrastructure, giving you control over storage, analysis, and remediation actions.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_azure(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Azure', 'f', 'azure.svg', 'AZURE', srv_id, 'f', 'Cloud', 't',
            'At its core, Azure is a public cloud computing platform—with solutions including Infrastructure as a Service (IaaS), Platform as a Service (PaaS), and Software as a Service (SaaS) that can be used for services such as analytics, virtual computing, storage, networking, and much more',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Azure',
                                                      module_icon        = 'azure.svg',
                                                      module_name        = 'AZURE',
                                                      module_category    = 'Cloud',
                                                      module_description = 'At its core, Azure is a public cloud computing platform—with solutions including Infrastructure as a Service (IaaS), Platform as a Service (PaaS), and Software as a Service (SaaS) that can be used for services such as analytics, virtual computing, storage, networking, and much more',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_bitdefender(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO utm_module (pretty_name, module_description, module_active, module_icon, module_name,
                            server_id, module_category, needs_restart, lite_version, is_activatable)
    VALUES ('Bitdefender',
            'Bitdefender provides cyber-security solutions with leading security efficacy, performance, and ease of use to small and medium businesses, mid-market enterprises, and consumers',
            FALSE, 'bitdefender.svg', 'BITDEFENDER', srv_id, 'Antivirus', FALSE, TRUE,
            TRUE)
    ON CONFLICT (module_name, server_id) DO UPDATE SET pretty_name        = 'Bitdefender',
                                                       module_icon        = 'bitdefender.svg',
                                                       module_name        = 'BITDEFENDER',
                                                       module_category    = 'Antivirus',
                                                       module_description = 'Bitdefender provides cyber-security solutions with leading security efficacy, performance, and ease of use to small and medium businesses, mid-market enterprises, and consumers',
                                                       lite_version       = TRUE,
                                                       is_activatable     = TRUE,
                                                       server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_cisco(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Cisco ASA', 'f', 'cisco.svg', 'CISCO', srv_id, 'f', 'Device', 't',
            'Adaptive Security Appliances, or simply Cisco ASA, is Cisco''s line of network security devices. This integration configures UTMStack ingestion of logs from this device.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Cisco ASA',
                                                      module_icon        = 'cisco.svg',
                                                      module_name        = 'CISCO',
                                                      module_category    = 'Device',
                                                      module_description = 'Adaptive Security Appliances, or simply Cisco ASA, is Cisco''s line of network security devices. This integration configures UTMStack ingestion of logs from this device.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_cisco_meraki(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Cisco Meraki', 'f', 'meraki.svg', 'MERAKI', srv_id, 'f', 'Device', 't',
            'This integration enables a Syslog integration with CISCO Meraki firewalls. Configure your Meraki device to send logs to complete the integration.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Cisco Meraki',
                                                      module_icon        = 'meraki.svg',
                                                      module_name        = 'MERAKI',
                                                      module_category    = 'Device',
                                                      module_description = 'This integration enables a Syslog integration with CISCO Meraki firewalls. Configure your Meraki device to send logs to complete the integration.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_cisco_switch(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Cisco Switch', 'f', 'cisco.svg', 'CISCO_SWITCH', srv_id, 'f', 'Device', 't',
            'Cisco network switches deliver performance, flexibility, and security. Cisco switches are scalable and cost-efficient and meet the demands of hybrid work.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Cisco Switch',
                                                      module_icon        = 'cisco.svg',
                                                      module_name        = 'CISCO_SWITCH',
                                                      module_category    = 'Device',
                                                      module_description = 'Cisco network switches deliver performance, flexibility, and security. Cisco switches are scalable and cost-efficient and meet the demands of hybrid work.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_deceptive_bytes(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Deceptive Bytes', 'f', 'deceptive-b.svg', 'DECEPTIVE_BYTES', srv_id, 'f', 'Other', 't',
            'A leader in endpoint deception technology, provides an Active Endpoint Deception platform to enterprises & MSSPs which enables them real-time prevention of unknown and sophisticated threats.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Deceptive Bytes',
                                                      module_icon        = 'deceptive-b.svg',
                                                      module_name        = 'DECEPTIVE_BYTES',
                                                      module_category    = 'Other',
                                                      module_description = 'A leader in endpoint deception technology, provides an Active Endpoint Deception platform to enterprises & MSSPs which enables them real-time prevention of unknown and sophisticated threats.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_elasticsearch(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin

    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Elasticsearch', 'f', 'elasticsearch.svg', 'ELASTICSEARCH', srv_id, 'f', 'Database', 't',
            'Elasticsearch is a highly scalable open-source full-text search and analytics engine. It allows you to store, search, and analyze big volumes of data quickly and in near real time. It is generally used as the underlying engine/technology that powers applications that have complex search features and requirements.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Elasticsearch',
                                                      module_icon        = 'elasticsearch.svg',
                                                      module_name        = 'ELASTICSEARCH',
                                                      module_category    = 'Database',
                                                      module_description = 'Elasticsearch is a highly scalable open-source full-text search and analytics engine. It allows you to store, search, and analyze big volumes of data quickly and in near real time. It is generally used as the underlying engine/technology that powers applications that have complex search features and requirements.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_eset(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('ESET Endpoint Protection', 'f', 'eset.svg', 'ESET', srv_id, 'f', 'Antivirus', 't',
            'Modern multilayered endpoint protection featuring strong machine learning and easy-to-use management.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'ESET Endpoint Protection',
                                                      module_icon        = 'eset.svg',
                                                      module_name        = 'ESET',
                                                      module_category    = 'Antivirus',
                                                      module_description = 'Modern multilayered endpoint protection featuring strong machine learning and easy-to-use management.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_file_integrity(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon",
                                      "module_name", "server_id", "needs_restart", "module_category",
                                      "lite_version", "module_description", is_activatable)
    VALUES ('File Classification', 'f', 'file-integrity.svg', 'FILE_INTEGRITY', srv_id, 'f', 'UTMStack Modules', 't',
            'Keep track of changes and access to classified information', TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'File Classification',
                                                      module_icon        = 'file-integrity.svg',
                                                      module_name        = 'FILE_INTEGRITY',
                                                      module_category    = 'UTMStack Modules',
                                                      module_description = 'Keep track of changes and access to classified information',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_fire_power(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Fire Power', 'f', 'fire-power.svg', 'FIRE_POWER', srv_id, 'f', 'Device', 't',
            'Cisco Networks Firepower Next-Generation Firewalls (NGFW) offer superior cyber threat protection, intrusion prevention, and enterprise security management controls for organizations of all sizes and deployments.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Fire Power',
                                                      module_icon        = 'fire-power.svg',
                                                      module_name        = 'FIRE_POWER',
                                                      module_category    = 'Device',
                                                      module_description = 'Cisco Networks Firepower Next-Generation Firewalls (NGFW) offer superior cyber threat protection, intrusion prevention, and enterprise security management controls for organizations of all sizes and deployments.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_fortigate(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('FortiGate', 'f', 'fortigate.svg', 'FORTIGATE', srv_id, 'f', 'Device', 't',
            'Fortinet''s FortiGate next-generation firewalls (NGFW) provide organizations supreme protection against web-based network threats, including known and unknown threats and intrusion strategies',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'FortiGate',
                                                      module_icon        = 'fortigate.svg',
                                                      module_name        = 'FORTIGATE',
                                                      module_category    = 'Device',
                                                      module_description = 'Fortinet''s FortiGate next-generation firewalls (NGFW) provide organizations supreme protection against web-based network threats, including known and unknown threats and intrusion strategies',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_gcp(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Google Cloud Platform', 'f', 'gcp.svg', 'GCP', srv_id, 'f', 'Cloud', 't',
            'Google Cloud Platform (GCP) is a suite of public cloud computing services offered by Google. The platform includes a range of hosted services for compute, storage and application development that run on Google hardware',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Google Cloud Platform',
                                                      module_icon        = 'gcp.svg',
                                                      module_name        = 'GCP',
                                                      module_category    = 'Cloud',
                                                      module_description = 'Google Cloud Platform is a suite of public cloud computing services offered by Google. The platform includes a range of hosted services for compute, storage and application development that run on Google hardware',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_github(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('GitHub', 'f', 'github.svg', 'GITHUB', srv_id, 'f', 'Other', 't',
            'GitHub, is an Internet hosting service for software development and version control using Git. It provides the distributed version control of Git plus access control, bug tracking, software feature requests, task management, continuous integration, and wikis for every project.',
            TRUE)

    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'GitHub',
                                                      module_icon        = 'github.svg',
                                                      module_name        = 'GITHUB',
                                                      module_category    = 'Other',
                                                      module_description = 'GitHub, is an Internet hosting service for software development and version control using Git. It provides the distributed version control of Git plus access control, bug tracking, software feature requests, task management, continuous integration, and wikis for every project.',
                                                      lite_version       = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_hap(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin

    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('High Availability Proxy', 'f', 'haproxy.svg', 'HAPROXY', srv_id, 'f', 'Proxy', 't',
            'HAProxy (High Availability Proxy) is open source proxy and load balancing server software. It provides high availability at the network (TCP) and application (HTTP/S) layers, improving speed and performance by distributing workload across multiple servers.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'High Availability Proxy',
                                                      module_icon        = 'haproxy.svg',
                                                      module_name        = 'HAPROXY',
                                                      module_category    = 'Proxy',
                                                      module_description = 'HAProxy (High Availability Proxy) is open source proxy and load balancing server software. It provides high availability at the network (TCP) and application (HTTP/S) layers, improving speed and performance by distributing workload across multiple servers.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_ibm_as_400(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO utm_module (pretty_name, module_description, module_active, module_icon, module_name,
                            server_id, module_category, needs_restart, lite_version, is_activatable)
    VALUES ('IBM AS/400',
            'The IBM AS/400 is a family of midrange computers from IBM announced in June 1988 and released in August 1988. It was the successor to the System/36 and System/38 platforms, and ran the OS/400 operating system',
            FALSE, 'ibm-as-400.svg', 'IBM_AS_400', srv_id, 'Device', FALSE, TRUE,
            TRUE)
    ON CONFLICT (module_name, server_id) DO UPDATE SET pretty_name        = 'IBM AS/400',
                                                       module_icon        = 'ibm-as-400.svg',
                                                       module_name        = 'IBM_AS_400',
                                                       module_category    = 'Device',
                                                       module_description = 'The IBM AS/400 is a family of midrange computers from IBM announced in June 1988 and released in August 1988. It was the successor to the System/36 and System/38 platforms, and ran the OS/400 operating system',
                                                       lite_version       = TRUE,
                                                       is_activatable     = TRUE,
                                                       server_id          = srv_id;


end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_iis(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Internet Information Services', 'f', 'iis.svg', 'IIS', srv_id, 'f', 'Web Server', 't',
            'Internet Information Services (IIS) is a flexible, general-purpose web server from Microsoft that runs on Windows systems to serve requested HTML pages or files. An IIS web server accepts requests from remote client computers and returns the appropriate response',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Internet Information Services',
                                                      module_icon        = 'iis.svg',
                                                      module_name        = 'IIS',
                                                      module_category    = 'Web Server',
                                                      module_description = 'Internet Information Services (IIS) is a flexible, general-purpose web server from Microsoft that runs on Windows systems to serve requested HTML pages or files. An IIS web server accepts requests from remote client computers and returns the appropriate response',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_json(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Json Input', true, 'json.svg', 'JSON', srv_id, 'f', 'Other', 't',
            'Activating this module you can send your json format logs to be processed by UTMStack', FALSE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Json Input',
                                                      module_icon        = 'json.svg',
                                                      module_name        = 'JSON',
                                                      module_category    = 'Other',
                                                      module_description = 'Activating this module you can send your json format logs to be processed by UTMStack',
                                                      lite_version       = TRUE,
                                                      is_activatable     = FALSE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_kafka(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin

    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Kafka', 'f', 'kafka.svg', 'KAFKA', srv_id, 'f', 'Database', 't',
            'Kafka is primarily used to build real-time streaming data pipelines and applications that adapt to the data streams. It combines messaging, storage, and stream processing to allow storage and analysis of both historical and real-time data.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Kafka',
                                                      module_icon        = 'kafka.svg',
                                                      module_name        = 'KAFKA',
                                                      module_category    = 'Database',
                                                      module_description = 'Kafka is primarily used to build real-time streaming data pipelines and applications that adapt to the data streams. It combines messaging, storage, and stream processing to allow storage and analysis of both historical and real-time data.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_kaspersky(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Kaspersky Security', 'f', 'kaspersky.svg', 'KASPERSKY', srv_id, 'f', 'Antivirus', 't',
            'Kaspersky Security provides comprehensive information about the devices and applications running on your network.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Kaspersky Security',
                                                      module_icon        = 'kaspersky.svg',
                                                      module_name        = 'KASPERSKY',
                                                      module_category    = 'Antivirus',
                                                      module_description = 'Kaspersky Security provides comprehensive information about the devices and applications running on your network.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_kibana(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Kibana', 'f', 'kibana.svg', 'KIBANA', srv_id, 'f', 'Database', 't',
            'Kibana is a data visualization and exploration tool used for log and time-series analytics, application monitoring, and operational intelligence use cases. It offers powerful and easy-to-use features such as histograms, line graphs, pie charts, heat maps, and built-in geospatial support.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Kibana',
                                                      module_icon        = 'kibana.svg',
                                                      module_name        = 'KIBANA',
                                                      module_category    = 'Database',
                                                      module_description = 'Kibana is a data visualization and exploration tool used for log and time-series analytics, application monitoring, and operational intelligence use cases. It offers powerful and easy-to-use features such as histograms, line graphs, pie charts, heat maps, and built-in geospatial support.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_linux_agent(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Linux agent', true, 'linux_agent.svg', 'LINUX_AGENT', srv_id, 'f', 'Agents & Syslog', 't',
            'By installing and configuring this agent on the Linux systems family you can send the logs generated by this operating system to UTMStack',
            FALSE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Linux agent',
                                                      module_icon        = 'linux_agent.svg',
                                                      module_name        = 'LINUX_AGENT',
                                                      module_category    = 'Agents & Syslog',
                                                      module_description = 'By installing and configuring this agent on the Linux systems family you can send the logs generated by this operating system to UTMStack',
                                                      lite_version       = TRUE,
                                                      is_activatable     = FALSE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_linux_audit_demon(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin

    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Linux Auditing Demon', 'f', 'auditd.svg', 'AUDITD', srv_id, 'f', 'Other', 't',
            'The job of the Linux Auditing Demon is to collect and write log files of audit to the disk as a background service.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Linux Auditing Demon',
                                                      module_icon        = 'auditd.svg',
                                                      module_name        = 'AUDITD',
                                                      module_category    = 'Other',
                                                      module_description = 'The job of the Linux Auditing Demon is to collect and write log files of audit to the disk as a background service.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_linux_logs(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin

    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Linux Logs', true, 'linux-tux.svg', 'LINUX_LOGS', srv_id, 'f', 'Agents & Syslog', 't',
            'Linux logs provide a timeline of events for the Linux operating system, ' ||
            'applications and system and are a valuable troubleshooting tool when you ' ||
            'encounter issues.This is an alternative for collecting Linux logs using ' ||
            'rsyslog when the UTMStack Agent can not be used.',
            FALSE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Linux Logs',
                                                      module_icon        = 'linux-tux.svg',
                                                      module_name        = 'LINUX_LOGS',
                                                      module_category    = 'Agents & Syslog',
                                                      module_description =
                                                                      'Linux logs provide a timeline of events for the Linux operating system, ' ||
                                                                      'applications and system and are a valuable troubleshooting tool when you ' ||
                                                                      'encounter issues.This is an alternative for collecting Linux logs using ' ||
                                                                      'rsyslog when the UTMStack Agent can not be used.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = FALSE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_logstash(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Logstash', 'f', 'logstash.svg', 'LOGSTASH', srv_id, 'f', 'Database', 't',
            'Logstash is a light-weight, open-source, server-side data processing pipeline that allows you to collect data from a variety of sources, transform it on the fly, and send it to your desired destination. It is most often used as a data pipeline for Elasticsearch, an open-source analytics and search engine.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Logstash',
                                                      module_icon        = 'logstash.svg',
                                                      module_name        = 'LOGSTASH',
                                                      module_category    = 'Database',
                                                      module_description = 'Logstash is a light-weight, open-source, server-side data processing pipeline that allows you to collect data from a variety of sources, transform it on the fly, and send it to your desired destination. It is most often used as a data pipeline for Elasticsearch, an open-source analytics and search engine.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_macos(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('MacOS', 'f', 'macos.svg', 'MACOS', srv_id, 'f', 'Other', 't',
            'macOS (formerly Mac OS X, later OS X) is a series of operating systems, primarily for Apple Mac family of computers.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'MacOS',
                                                      module_icon        = 'macos.svg',
                                                      module_name        = 'MACOS',
                                                      module_category    = 'Other',
                                                      module_description = 'macOS (formerly Mac OS X, later OS X) is a series of operating systems, primarily for Apple Mac family of computers.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_mikrotik(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('MikroTik', 'f', 'mikrotik.svg', 'MIKROTIK', srv_id, 'f', 'Cloud', 't',
            'MikroTik is company provides routing, switching and wireless equipment for all possible
             uses - from the customer location, up to high end data centres.MikroTik uses RouterOS, is an operating system and software capable of acting as a Router, Bridge, Firewall, Bandwidth Management, Wireless AP & Client and many other functions.RouterOS can perform almost all networking functions as well as some server functions.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'MikroTik',
                                                      module_icon        = 'mikrotik.svg',
                                                      module_name        = 'MIKROTIK',
                                                      module_category    = 'Device',
                                                      module_description = 'MikroTik is company provides routing, switching and wireless equipment for all possible
                                                                             uses - from the customer location, up to high end data centres.MikroTik uses RouterOS, is an operating system and software capable of acting as a Router, Bridge, Firewall, Bandwidth Management, Wireless AP & Client and many other functions.RouterOS can perform almost all networking functions as well as some server functions.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_mongodb(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin

    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('MongoDB', 'f', 'mongodb.svg', 'MONGODB', srv_id, 'f', 'Database', 't',
            'MongoDB is a document database used to build highly available and scalable internet applications.', TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'MongoDB',
                                                      module_icon        = 'mongodb.svg',
                                                      module_name        = 'MONGODB',
                                                      module_category    = 'Database',
                                                      module_description = 'MongoDB is a document database used to build highly available and scalable internet applications.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_mysql(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin

    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('MySQL', 'f', 'mysql.svg', 'MYSQL', srv_id, 'f', 'Database', 't',
            'MySQL is a relational database management system based on SQL. The most common use for mySQL however, is for the purpose of a web database. It can be used to store anything from a single record of information to an entire inventory of available products for an online store.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'MySQL',
                                                      module_icon        = 'mysql.svg',
                                                      module_name        = 'MYSQL',
                                                      module_category    = 'Database',
                                                      module_description = 'MySQL is a relational database management system based on SQL. The most common use for mySQL however, is for the purpose of a web database. It can be used to store anything from a single record of information to an entire inventory of available products for an online store.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_nats(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Nats', 'f', 'nats.svg', 'NATS', srv_id, 'f', 'Other', 't',
            'NATS is an open-source messaging system. The core design principles of NATS are performance, scalability, and ease of use.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Nats',
                                                      module_icon        = 'nats.svg',
                                                      module_name        = 'NATS',
                                                      module_category    = 'Other',
                                                      module_description = 'NATS is an open-source messaging system. The core design principles of NATS are performance, scalability, and ease of use.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_netflow(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin
    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon",
                                      "module_name", "server_id", "needs_restart", "module_category",
                                      "lite_version", "module_description", "is_activatable")
    VALUES ('Netflow', 'f', 'netflow.svg', 'NETFLOW', srv_id, 'f', 'Network', 't',
            'Integrating NetFlow you can redirect all logs of the network traffic to UTMStack, allowing you to monitor and analyze this logs more efficiently and effectively',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Netflow',
                                                      module_icon        = 'netflow.svg',
                                                      module_name        = 'NETFLOW',
                                                      module_category    = 'Network',
                                                      module_description = 'Integrating NetFlow you can redirect all logs of the network traffic to UTMStack, allowing you to monitor and analyze this logs more efficiently and effectively',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;
end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_nginx(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin

    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Nginx', 'f', 'nginx.svg', 'NGINX', srv_id, 'f', 'Web Server', 't',
            'NGINX is open source software for web serving, reverse proxying, caching, load balancing, media streaming, and more.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Nginx',
                                                      module_icon        = 'nginx.svg',
                                                      module_name        = 'NGINX',
                                                      module_category    = 'Web Server',
                                                      module_description = 'NGINX is open source software for web serving, reverse proxying, caching, load balancing, media streaming, and more.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_o365(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Office365', 'f', 'office.svg', 'O365', srv_id, 'f', 'Cloud', 't',
            'Microsoft 365, formerly Office 365, is a line of subscription services offered by Microsoft which adds to and includes the Microsoft Office product line',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Office365',
                                                      module_icon        = 'office.svg',
                                                      module_name        = 'O365',
                                                      module_category    = 'Cloud',
                                                      module_description = 'Microsoft 365, formerly Office 365, is a line of subscription services offered by Microsoft which adds to and includes the Microsoft Office product line',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_osquery(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('OsQuery', 'f', 'osquery.svg', 'OSQUERY', srv_id, 'f', 'Other', 't',
            'OsQuery allows you to craft your system queries using SQL statements, making it easy to use by security engineers that are already familiar with SQL. Osquery is a flexible tool and can be used for a variety of use cases to troubleshoot performance and operational issues.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'OsQuery',
                                                      module_icon        = 'osquery.svg',
                                                      module_name        = 'OSQUERY',
                                                      module_category    = 'Other',
                                                      module_description = 'OsQuery allows you to craft your system queries using SQL statements, making it easy to use by security engineers that are already familiar with SQL. osquery is a flexible tool and can be used for a variety of use cases to troubleshoot performance and operational issues.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_palo_alto(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Palo Alto', 'f', 'palo-alto.svg', 'PALO_ALTO', srv_id, 'f', 'Device', 't',
            'Palo Alto Networks® next-generation firewalls inspect all traffic (including applications, threats, and content), and tie that traffic to the user, regardless of location or device type. The user, application, and content—the elements that run your business—become integral components of your enterprise security policy. This allows you to align security with your business policies, as well as write rules that are easy to understand and maintain.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Palo Alto',
                                                      module_icon        = 'palo-alto.svg',
                                                      module_name        = 'PALO_ALTO',
                                                      module_category    = 'Device',
                                                      module_description = 'Palo Alto Networks® next-generation firewalls inspect all traffic (including applications, threats, and content), and tie that traffic to the user, regardless of location or device type. The user, application, and content—the elements that run your business—become integral components of your enterprise security policy. This allows you to align security with your business policies, as well as write rules that are easy to understand and maintain.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_postgresql(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('PostgreSQL', 'f', 'postgresql.svg', 'POSTGRESQL', srv_id, 'f', 'Database', 't',
            'PostgreSQL is used as the primary data store or data warehouse for many web, mobile, geospatial, and analytics applications.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'PostgreSQL',
                                                      module_icon        = 'postgresql.svg',
                                                      module_name        = 'POSTGRESQL',
                                                      module_category    = 'Database',
                                                      module_description = 'PostgreSQL is used as the primary data store or data warehouse for many web, mobile, geospatial, and analytics applications.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_redis(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Redis', 'f', 'redis.svg', 'REDIS', srv_id, 'f', 'Database', 't',
            'Redis can be used with streaming solutions such as Apache Kafka and Amazon Kinesis as an in-memory data store to ingest, process, and analyze real-time data with sub-millisecond latency. Redis is an ideal choice for real-time analytics use cases such as social media analytics, ad targeting, personalization, and IoT.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Redis',
                                                      module_icon        = 'redis.svg',
                                                      module_name        = 'REDIS',
                                                      module_category    = 'Database',
                                                      module_description = 'Redis can be used with streaming solutions such as Apache Kafka and Amazon Kinesis as an in-memory data store to ingest, process, and analyze real-time data with sub-millisecond latency. Redis is an ideal choice for real-time analytics use cases such as social media analytics, ad targeting, personalization, and IoT.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_salesforce(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin

    INSERT INTO utm_module (pretty_name, module_description, module_active, module_icon, module_name,
                            server_id, module_category, needs_restart, lite_version, is_activatable)
    VALUES ('Salesforce',
            'Salesforce provides customer relationship management software and applications focused on sales, customer service, marketing automation, e-commerce, analytics, and application development',
            FALSE, 'salesforce.svg', 'SALESFORCE', srv_id, 'Cloud', FALSE, TRUE,
            TRUE)
    ON CONFLICT (module_name, server_id) DO UPDATE SET pretty_name        = 'Salesforce',
                                                       module_icon        = 'salesforce.svg',
                                                       module_name        = 'SALESFORCE',
                                                       module_category    = 'Cloud',
                                                       module_description = 'Salesforce provides customer relationship management software and applications focused on sales, customer service, marketing automation, e-commerce, analytics, and application development',
                                                       lite_version       = TRUE,
                                                       is_activatable     = TRUE,
                                                       server_id          = srv_id;


end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_sentinel_one(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('SentinelOne Endpoint Security', 'f', 'sentinelone.svg', 'SENTINEL_ONE', srv_id, 'f', 'XDR', 't',
            'SentinelOne Endpoint Security technology provides solutions with three different tiers of functionality, Core, Control and Complete.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'SentinelOne Endpoint Security',
                                                      module_icon        = 'sentinelone.svg',
                                                      module_name        = 'SENTINEL_ONE',
                                                      module_category    = 'XDR',
                                                      module_description = 'SentinelOne Endpoint Security technology provides solutions with three different tiers of functionality, Core, Control and Complete.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_soc_ai(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
declare
    grp_id integer;
    mod_id
           bigint;

begin
    INSERT INTO utm_module (pretty_name, module_description, module_active, module_icon, module_name,
                            server_id, module_category, needs_restart, lite_version, is_activatable)
    VALUES ('SOC AI',
            'The Generative Pre-trained Transformer (GPT) is an advanced AI language model developed by OpenAI, designed to understand and generate human-like text. GPT can be used to investigate Security Operations Center (SOC) alerts by analyzing large volumes of data, identifying patterns, and providing insights into potential security incidents. By leveraging GPT is natural language processing capabilities, security analysts can efficiently triage alerts, prioritize threats, and gather contextual information to aid in incident response and remediation.',
            FALSE, 'soc-ai.svg',
            'SOC_AI', srv_id,
            'AI',
            FALSE,
            TRUE,
            TRUE)
    ON CONFLICT (module_name, server_id) DO UPDATE SET pretty_name        = 'SOC AI',
                                                       module_icon        = 'soc-ai.svg',
                                                       module_name        = 'SOC_AI',
                                                       module_category    = 'AI',
                                                       module_description = 'The Generative Pre-trained Transformer (GPT) is an advanced AI language model developed by OpenAI, designed to understand and generate human-like text. GPT can be used to investigate Security Operations Center (SOC) alerts by analyzing large volumes of data, identifying patterns, and providing insights into potential security incidents. By leveraging GPT is natural language processing capabilities, security analysts can efficiently triage alerts, prioritize threats, and gather contextual information to aid in incident response and remediation.',
                                                       lite_version       = TRUE,
                                                       server_id          = srv_id
    RETURNING id
        INTO mod_id;

    insert
    into utm_module_group (module_id,
                           group_name,
                           group_description)
    values (mod_id,
            'Configuration',
            'Add the connection key to connect with OpenAI')
    on conflict (module_id ,
        group_name) do update
        set group_name = 'Configuration'
    returning id
        into
            grp_id;

    ----------------------------------------------
    -- DELETING PREVIOUS DATA BEFORE INSERT
    ----------------------------------------------
    delete from utm_module_group_configuration where group_id = grp_id;

    insert
    into utm_module_group_configuration (group_id,
                                         conf_key,
                                         conf_name,
                                         conf_description,
                                         conf_value,
                                         conf_required,
                                         conf_data_type)
    values (grp_id,
            'utmstack.socai.key',
            'Key',
            'OpenAI Connection key',
            null,
            't',
            'password')
    on conflict do nothing;
    insert
    into utm_module_group_configuration (group_id,
                                         conf_key,
                                         conf_name,
                                         conf_description,
                                         conf_value,
                                         conf_required,
                                         conf_data_type)
    values (grp_id,
            'utmstack.socai.incidentCreation',
            'Automatic Incident creation',
            'If set to "true", the system will create incidents based on analysis of alerts.',
            'true',
            'f',
            'bool')
    on conflict do nothing;

    insert
    into utm_module_group_configuration (group_id,
                                         conf_key,
                                         conf_name,
                                         conf_description,
                                         conf_value,
                                         conf_required,
                                         conf_data_type)
    values (grp_id,
            'utmstack.socai.changeAlertStatus',
            'Change Alert Status',
            'If set to "true", SOC Ai will automatically change the status of alerts. Analysts should investigate those with the status "In Review".',
            'true',
            'f',
            'bool')
    on conflict do nothing;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_sonic_wall(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('SonicWall', 'f', 'sonicwall.svg', 'SONIC_WALL', srv_id, 'f', 'Device', 't',
            'SonicWall next-generation firewalls (NGFW) provide the security, control and visibility you need to maintain an effective cybersecurity posture. SonicWall’s award-winning hardware and advanced technology are built into each firewall to give you the edge on evolving threats',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'SonicWall',
                                                      module_icon        = 'sonicwall.svg',
                                                      module_name        = 'SONIC_WALL',
                                                      module_category    = 'Device',
                                                      module_description = 'SonicWall next-generation firewalls (NGFW) provide the security, control and visibility you need to maintain an effective cybersecurity posture. SonicWall’s award-winning hardware and advanced technology are built into each firewall to give you the edge on evolving threats',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_sophos_central(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Sophos Central', 'f', 'sophos.svg', 'SOPHOS', srv_id, 'f', 'XDR', 't',
            'Sophos Central is a single cloud management solution for all your Sophos next-gen technologies: endpoint, server, mobile, firewall, ZTNA, email, and so much more',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Sophos Central',
                                                      module_icon        = 'sophos.svg',
                                                      module_name        = 'SOPHOS',
                                                      module_category    = 'XDR',
                                                      module_description = 'Sophos Central is a single cloud management solution for all your Sophos next-gen technologies: endpoint, server, mobile, firewall, ZTNA, email, and so much more',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_sophosxg(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Sophos XG', 'f', 'sophosxg.svg', 'SOPHOS_XG', srv_id, 'f', 'Device', 't',
            'Features full protection for your home network, including anti-malware, web security and URL filtering, application control, IPS, traffic shaping, VPN, reporting and monitoring, and much more.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Sophos XG',
                                                      module_icon        = 'sophosxg.svg',
                                                      module_name        = 'SOPHOS_XG',
                                                      module_category    = 'Device',
                                                      module_description = 'Features full protection for your home network, including anti-malware, web security and URL filtering, application control, IPS, traffic shaping, VPN, reporting and monitoring, and much more.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_syslog(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin
    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Syslog', true, 'log-file.svg', 'SYSLOG', srv_id, 'f', 'Agents & Syslog', 't',
            'Syslog is a standard for sending and receiving notification messages, in a particular format, from various network devices. UTMStack accepts syslog from firewalls and other devices that support it',
            FALSE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Syslog',
                                                      module_icon        = 'log-file.svg',
                                                      module_name        = 'SYSLOG',
                                                      module_category    = 'Agents & Syslog',
                                                      module_description = 'Syslog is a standard for sending and receiving notification messages, in a particular format, from various network devices. UTMStack accepts syslog from firewalls and other devices that support it',
                                                      lite_version       = TRUE,
                                                      is_activatable     = FALSE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_traefik(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('Traefik', 'f', 'traefik.svg', 'TRAEFIK', srv_id, 'f', 'Cloud', 't',
            'Traefik is the modern standard for Routing, Load Balancing, and Proxies for the Cloud, On-Prem, and Hybrid workloads.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Traefik',
                                                      module_icon        = 'traefik.svg',
                                                      module_name        = 'TRAEFIK',
                                                      module_category    = 'Cloud',
                                                      module_description = 'Traefik is the modern standard for Routing, Load Balancing, and Proxies for the Cloud, On-Prem, and Hybrid workloads.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_ufw(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin


    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('UFW', 'f', 'ubuntu-fw.svg', 'UFW', srv_id, 'f', 'Other', 't',
            'The default firewall configuration tool for Ubuntu is ufw. Developed to ease iptables firewall configuration, ufw provides a user-friendly way to create an IPv4 or IPv6 host-based firewall.',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'UFW',
                                                      module_icon        = 'ubuntu-fw.svg',
                                                      module_name        = 'UFW',
                                                      module_category    = 'Other',
                                                      module_description = 'The default firewall configuration tool for Ubuntu is ufw. Developed to ease iptables firewall configuration, ufw provides a user-friendly way to create an IPv4 or IPv6 host-based firewall.',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_vmware(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin

    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      is_activatable)
    VALUES ('VMWare Syslog', 'f', 'vmware.svg', 'VMWARE', srv_id, 'f', 'Agents & Syslog', 't',
            'VMWare allows businesses to run multiple application and operating system workloads on the one server. You can use the Syslog Service to redirect and store ESXi messages to UTMStack',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'VMWare Syslog',
                                                      module_icon        = 'vmware.svg',
                                                      module_name        = 'VMWARE',
                                                      module_category    = 'Agents & Syslog',
                                                      module_description = 'VMWare allows businesses to run multiple application and operating system workloads on the one server. You can use the Syslog Service to redirect and store ESXi messages to UTMStack',
                                                      lite_version       = TRUE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_vulnerabilities(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
BEGIN
    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon",
                                      "module_name", "server_id", "needs_restart", "module_category",
                                      "lite_version", "module_description", "is_activatable")
    VALUES ('Vulnerabilities', 'f', 'vulnerabilities.svg', 'VULNERABILITIES', srv_id, 'f', 'UTMStack Modules',
            FALSE,
            'Active and passive vulnerability scanners for early detection, with of the box reports for compliance audits',
            TRUE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Vulnerabilities',
                                                      module_icon        = 'vulnerabilities.svg',
                                                      module_name        = 'VULNERABILITIES',
                                                      module_category    = 'UTMStack Modules',
                                                      module_description = 'Active and passive vulnerability scanners for early detection, with of the box reports for compliance audits',
                                                      lite_version       = FALSE,
                                                      is_activatable     = TRUE,
                                                      server_id          = srv_id;
END;

$$;

CREATE OR REPLACE FUNCTION public.register_integration_window_agent(srv_id integer) RETURNS void
    LANGUAGE plpgsql
AS
$$
begin

    INSERT INTO "public"."utm_module"("pretty_name", "module_active", "module_icon", "module_name", "server_id",
                                      "needs_restart", "module_category", "lite_version", "module_description",
                                      "is_activatable")
    VALUES ('Windows Agent', true, 'windows-agent.svg', 'WINDOWS_AGENT', srv_id, 'f', 'Agents & Syslog', 't',
            'By installing and configuring this agent on windows systems you can send the logs generated by this operating system to UTMStack',
            FALSE)
    ON CONFLICT (module_name,server_id) DO UPDATE SET pretty_name        = 'Windows Agent',
                                                      module_icon        = 'windows-agent.svg',
                                                      module_name        = 'WINDOWS_AGENT',
                                                      module_category    = 'Agents & Syslog',
                                                      module_description = 'By installing and configuring this agent on windows systems you can send the logs generated by this operating system to UTMStack',
                                                      lite_version       = TRUE,
                                                      is_activatable     = FALSE,
                                                      server_id          = srv_id;

end;

$$;

CREATE OR REPLACE FUNCTION public.register_integrations(srv_id integer, srv_type character varying) RETURNS void
    LANGUAGE plpgsql
AS
$$

BEGIN
    ------------------------------------------------------------------------------------------------------
    -- Creating modules
    ------------------------------------------------------------------------------------------------------

    ---------------------------------------------------
    -- Netflow
    ---------------------------------------------------
    perform public.register_integration_netflow(srv_id);

    ---------------------------------------------------
    -- Windows Agent
    ---------------------------------------------------
    perform public.register_integration_window_agent(srv_id);

    ---------------------------------------------------
    -- Syslog
    ---------------------------------------------------
    perform public.register_integration_syslog(srv_id);

    ---------------------------------------------------
    -- Linux Logs
    ---------------------------------------------------
    perform public.register_integration_linux_logs(srv_id);

    ---------------------------------------------------
    -- VMWare Syslog
    ---------------------------------------------------
    perform public.register_integration_vmware(srv_id);

    ---------------------------------------------------
    -- Linux agent
    ---------------------------------------------------
    perform public.register_integration_linux_agent(srv_id);

    ---------------------------------------------------
    -- Apache
    ---------------------------------------------------
    perform public.register_integration_apache(srv_id);

    ---------------------------------------------------
    -- Apache2
    ---------------------------------------------------
    --perform public.register_integration_apache2(srv_id);

    ---------------------------------------------------
    -- Linux Auditing Demon
    ---------------------------------------------------
    perform public.register_integration_linux_audit_demon(srv_id);

    ---------------------------------------------------
    -- Elasticsearch
    ---------------------------------------------------
    perform public.register_integration_elasticsearch(srv_id);

    ---------------------------------------------------
    -- High Availability Proxy
    ---------------------------------------------------
    perform public.register_integration_hap(srv_id);

    ---------------------------------------------------
    -- Kafka
    ---------------------------------------------------
    perform public.register_integration_kafka(srv_id);

    ---------------------------------------------------
    -- Kibana
    ---------------------------------------------------
    perform public.register_integration_kibana(srv_id);

    ---------------------------------------------------
    -- Logstash
    ---------------------------------------------------
    perform public.register_integration_logstash(srv_id);

    ---------------------------------------------------
    -- MongoDB
    ---------------------------------------------------
    perform public.register_integration_mongodb(srv_id);

    ---------------------------------------------------
    -- MySQL
    ---------------------------------------------------
    perform public.register_integration_mysql(srv_id);

    ---------------------------------------------------
    -- Nats
    ---------------------------------------------------
    perform public.register_integration_nats(srv_id);

    ---------------------------------------------------
    -- Nginx
    ---------------------------------------------------
    perform public.register_integration_nginx(srv_id);

    ---------------------------------------------------
    -- OsQuery
    ---------------------------------------------------
    perform public.register_integration_osquery(srv_id);

    ---------------------------------------------------
    -- PostgreSQL
    ---------------------------------------------------
    perform public.register_integration_postgresql(srv_id);

    ---------------------------------------------------
    -- Redis
    ---------------------------------------------------
    perform public.register_integration_redis(srv_id);

    ---------------------------------------------------
    -- Traefik
    ---------------------------------------------------
    perform public.register_integration_traefik(srv_id);

    ---------------------------------------------------
    -- Cisco
    ---------------------------------------------------
    perform public.register_integration_cisco(srv_id);

    ---------------------------------------------------
    -- Cisco Meraki
    ---------------------------------------------------
    perform public.register_integration_cisco_meraki(srv_id);

    ---------------------------------------------------
    -- Json Input
    ---------------------------------------------------
    perform public.register_integration_json(srv_id);

    ---------------------------------------------------
    -- Internet Information Services (IIS)
    ---------------------------------------------------
    perform public.register_integration_iis(srv_id);

    ---------------------------------------------------
    -- Kaspersky Security
    ---------------------------------------------------
    perform public.register_integration_kaspersky(srv_id);

    --------------------------------------------------------------
    -- ESET Endpoint Protection
    --------------------------------------------------------------
    perform public.register_integration_eset(srv_id);

    ------------------------------------------------------------
    -- SentinelOne
    ------------------------------------------------------------
    perform public.register_integration_sentinel_one(srv_id);

    ---------------------------------------------------
    -- FortiGate
    ---------------------------------------------------
    perform public.register_integration_fortigate(srv_id);

    ---------------------------------------------------
    -- Sophos XG
    ---------------------------------------------------
    perform public.register_integration_sophosxg(srv_id);

    ---------------------------------------------------
    -- MacOS
    ---------------------------------------------------
    perform public.register_integration_macos(srv_id);

    ---------------------------------------------------
    -- IBM AS/400
    ---------------------------------------------------
    --perform public.register_integration_ibm_as_400(srv_id);

    IF srv_type = 'aio' THEN
        ---------------------------------------------------
        -- File Classification
        ---------------------------------------------------
        perform public.register_integration_file_integrity(srv_id);

        ---------------------------------------------------
        -- Azure
        ---------------------------------------------------
        perform public.register_integration_azure(srv_id);

        ---------------------------------------------------
        -- Office365
        ---------------------------------------------------
        perform public.register_integration_o365(srv_id);

        ---------------------------------------------------
        -- AWS Cloudwatch
        ---------------------------------------------------
        perform public.register_integration_aws(srv_id);

        ---------------------------------------------------
        -- Sophos Central
        ---------------------------------------------------
        perform public.register_integration_sophos_central(srv_id);

        ---------------------------------------------------
        -- Google Cloud Platform (GCP)
        ---------------------------------------------------
        perform public.register_integration_gcp(srv_id);

        ---------------------------------------------------
        -- Fire Power
        ---------------------------------------------------
        perform public.register_integration_fire_power(srv_id);

        ---------------------------------------------------
        -- UFW
        ---------------------------------------------------
        perform public.register_integration_ufw(srv_id);

        ---------------------------------------------------
        -- Mikrotik
        ---------------------------------------------------
        perform public.register_integration_mikrotik(srv_id);

        ---------------------------------------------------
        -- Palo Alto
        ---------------------------------------------------
        perform public.register_integration_palo_alto(srv_id);

        ---------------------------------------------------
        -- Cisco Switch
        ---------------------------------------------------
        perform public.register_integration_cisco_switch(srv_id);

        ---------------------------------------------------
        -- SonicWall
        ---------------------------------------------------
        perform public.register_integration_sonic_wall(srv_id);

        ---------------------------------------------------
        -- Deceptive Bytes
        ---------------------------------------------------
        perform public.register_integration_deceptive_bytes(srv_id);

        ---------------------------------------------------
        -- GitHub
        ---------------------------------------------------
        perform public.register_integration_github(srv_id);

        ---------------------------------------------------
        -- Bitdefender
        ---------------------------------------------------
        perform public.register_integration_bitdefender(srv_id);

        ---------------------------------------------------
        -- Salesforce
        ---------------------------------------------------
        --perform public.register_integration_salesforce(srv_id);

        ---------------------------------------------------
        -- GPT
        ---------------------------------------------------
        perform public.register_integration_soc_ai(srv_id);

    END IF;

    perform public.update_module_dependencies();
END;
$$;

CREATE OR REPLACE FUNCTION public.update_module_dependencies() RETURNS void
    LANGUAGE plpgsql
AS
$$
declare
    m           record;
    m_instances int;
begin
    for m in (select distinct utm_module.module_name from utm_module)
        loop
            m_instances :=
                (select count(*) from utm_module where module_name = m.module_name and module_active is true);
            update utm_menu
            set menu_active=(m_instances > 0)
            where m.module_name = any (string_to_array(module_name_short, ','));
            update utm_index_pattern
            set is_active=(m_instances > 0)
            where m.module_name = any (string_to_array(pattern_module, ','));
            update utm_logstash_filter
            set is_active=(m_instances > 0)
            where m.module_name = any (string_to_array(module_name, ','));
        end loop;

    if m is null then
        update utm_menu set menu_active= false where module_name_short is not null;
        update utm_index_pattern set is_active= false where pattern_module is not null;
        update utm_logstash_filter set is_active= false where module_name is not null;
    end if;
end;
$$;


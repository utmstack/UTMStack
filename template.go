package main

const (
	stackName = "utmstack"
	composerFile = "utmstack.yml"
	composerTemplate = `version: "3.8"
volumes:
  logstash_pipeline:
    external: false
  rsyslog_logs:
    external: false
  wazuh_logs:
    external: false

services:
  elasticsearch:
    image: "utmstack.azurecr.io/opendistro:1.11.0"
    volumes:
      - ${ES_DATA}:/usr/share/elasticsearch/data
      - ${ES_BACKUPS}:/usr/share/elasticsearch/backups
    environment:
      - node.name=elasticsearch
      - discovery.seed_hosts=elasticsearch
      - cluster.initial_master_nodes=elasticsearch
      - "ES_JAVA_OPTS=-Xms${ES_MEM}g -Xmx${ES_MEM}g"
      - path.repo=/usr/share/elasticsearch/backups

  openvas:
    image: "utmstack.azurecr.io/openvas:11"
    ports:
      - "5432:5432"
      - "9390:9390"
    environment:
      - USERNAME=${DB_USER}
      - PASSWORD=${DB_PASS}
      - DB_PASSWORD=${DB_PASS}
      - HTTPS=0

# Configure pipeline from mutate
  logstash:
    image: "utmstack.azurecr.io/logstash:7.9.3"
    volumes:
      - logstash_pipeline:/usr/share/logstash/pipeline
    ports:
      - 5044:5044
    environment:
      - CONFIG_RELOAD_AUTOMATIC=true

  kibana:
    depends_on:
      - elasticsearch
    image: "utmstack.azurecr.io/opendistro-kibana:1.11.0"
    environment:
      - ELASTICSEARCH_URL=http://elasticsearch:9200
      - ELASTICSEARCH_HOSTS=http://elasticsearch:9200

  rsyslog:
    image: "utmstack.azurecr.io/rsyslog:8.36.0"
    volumes:
      - rsyslog_logs:/logs
    ports:
      - 514:514
      - 514:514/udp

  wazuh:
    image: "utmstack.azurecr.io/wazuh:3.11.1"
    volumes:
      - wazuh_logs:/var/ossec/logs
    ports:
      - 1514:1514
      - 1514:1514/udp

  scanner:
    image: "utmstack.azurecr.io/scanner:1.0.0"

# Add configs to image in /etc/nginx/conf
  nginx:
    image: "utmstack.azurecr.io/nginx:1.19.5"
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ${NGINX_CERT}:/etc/nginx/cert

  panel:
    depends_on:
      - openvas
      - elasticsearch
    image: "utmstack.azurecr.io/panel:7.0.0"
    environment:
      - TOMCAT_ADMIN_USER=${DB_USER}
      - TOMCAT_ADMIN_PASSWORD=${DB_PASS}
      - POSTGRESQL_USER=gvm
      - POSTGRESQL_PASSWORD=${DB_PASS}
      - POSTGRESQL_HOST=openvas
      - POSTGRESQL_PORT=5432
      - POSTGRESQL_DATABASE=utmstack
      - ELASTICSEARCH_HOST=elasticsearch
      - ELASTICSEARCH_PORT=9200
      - ELASTICSEARCH_SECRET=${CLIENT_SECRET}
      - OPENVAS_HOST=openvas
      - OPENVAS_PORT=9390
      - OPENVAS_USER=${DB_USER}
      - OPENVAS_PASSWORD=${DB_PASS}
      - OPENVAS_PG_PORT=5432
      - OPENVAS_PG_DATABASE=gvmd
      - OPENVAS_PG_USER=gvm
      - OPENVAS_PG_PASSWORD=${DB_PASS}
      - JRE_HOME=/opt/tomcat/bin/jre
      - JAVA_HOME=/opt/tomcat/bin/jre
      - CATALINA_BASE=/opt/tomcat/
      - CATALINA_HOME=/opt/tomcat/
      - LD_LIBRARY_PATH=/usr/lib/x86_64-linux-gnu

  utmstack_aws:
    depends_on:
      - logan
    image: "utmstack.azurecr.io/datasources:7.0.0"
    environment:
      - SERVER_NAME
      - DB_USER
      - DB_PASS
    command: ["python3", "-m", "utmstack.aws"]

  utmstack_azure:
    depends_on:
      - logan
    image: "utmstack.azurecr.io/datasources:7.0.0"
    environment:
      - SERVER_NAME
      - DB_USER
      - DB_PASS
    command: ["python3", "-m", "utmstack.azure"]

  utmstack_mutate:
    depends_on:
      - logan
    image: "utmstack.azurecr.io/datasources:7.0.0"
    environment:
      - SERVER_NAME
      - DB_USER
      - DB_PASS
    command: ["python3", "-m", "utmstack.mutate"]

  utmstack_office365:
    depends_on:
      - logan
    image: "utmstack.azurecr.io/datasources:7.0.0"
    environment:
      - SERVER_NAME
      - DB_USER
      - DB_PASS
    command: ["python3", "-m", "utmstack.office365"]

  utmstack_openvas:
    depends_on:
      - logan
    image: "utmstack.azurecr.io/datasources:7.0.0"
    environment:
      - SERVER_NAME
      - DB_USER
      - DB_PASS
    command: ["python3", "-m", "utmstack.openvas"]

  utmstack_suric:
    depends_on:
      - logan
    image: "utmstack.azurecr.io/datasources:7.0.0"
    environment:
      - SERVER_NAME
      - DB_USER
      - DB_PASS
    command: ["python3", "-m", "utmstack.suric"]

  utmstack_transporter:
    depends_on:
      - logan
    image: "utmstack.azurecr.io/datasources:7.0.0"
    environment:
      - SERVER_NAME
      - DB_USER
      - DB_PASS
    command: ["python3", "-m", "utmstack.transporter"]
`)

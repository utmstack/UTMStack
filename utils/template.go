package utils

const (
	stackName    = "utmstack"
	composerFile = "utmstack.yml"
	baseTemplate = `version: "3.8"

volumes:
  rsyslog_logs:
    external: false
  postgres_data:
    external: false
  wazuh_logs:
    external: false

services:
  openvas:
    image: "utmstack.azurecr.io/openvas:11"
    ports:
      - "8888:5432"
      - "9390:9390"
    environment:
      - USERNAME=admin
      - PASSWORD=${DB_PASS}
      - DB_PASSWORD=${DB_PASS}
      - HTTPS=0

  logstash:
    image: "utmstack.azurecr.io/logstash:7.9.3"
    volumes:
      - ${LOGSTASH_PIPELINE}:/usr/share/logstash/pipeline
    ports:
      - 5044:5044
    environment:
      - CONFIG_RELOAD_AUTOMATIC=true

  rsyslog:
    image: "utmstack.azurecr.io/rsyslog:8.36.0"
    volumes:
      - ${RSYSLOG_LOGS}:/logs
    ports:
      - 514:514
      - 514:514/udp

  scanner:
    image: "utmstack.azurecr.io/scanner:1.0.0-alpha.1"
    ports:
      - "5000:5000"
      - "8000:8000"

  datasources_mutate:
    image: "utmstack.azurecr.io/datasources:testing"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
      - ${LOGSTASH_PIPELINE}:/usr/share/logstash/pipeline
    environment:
      - SERVER_NAME
      - SERVER_TYPE
      - DB_HOST
      - DB_PASS
    command: ["python3", "-m", "utmstack.mutate"]

  datasources_openvas:
    image: "utmstack.azurecr.io/datasources:testing"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
    environment:
      - DB_NAME=gvmd
      - SERVER_NAME
      - SERVER_TYPE
      - DB_HOST
      - DB_PASS
    command: ["python3", "-m", "utmstack.openvas"]

  datasources_transporter:
    image: "utmstack.azurecr.io/datasources:testing"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
      - ${RSYSLOG_LOGS}:/logs
      - wazuh:/var/ossec/
      - ${WAZUH_LOGS}:/var/ossec/logs
    environment:
      - SERVER_NAME
      - SERVER_TYPE
      - DB_HOST
      - DB_PASS
    command: ["python3", "-m", "utmstack.transporter"]

  datasources_probe_api:
    image: "utmstack.azurecr.io/datasources:testing"
    volumes:
      - wazuh:/var/ossec/
      - ${WAZUH_LOGS}:/var/ossec/logs
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
    environment:
      - SERVER_NAME
      - SERVER_TYPE
      - DB_HOST
      - DB_PASS
      - SCANNER_IP
    ports:
      - 23949:23949
      - 1514:1514
      - 1514:1514/udp
      - 1515:1515
      - 1516:1516
      - 55000:55000
    command: ["/pw.sh"]
`
	masterTemplate = baseTemplate + `
  elasticsearch:
    image: "utmstack.azurecr.io/opendistro:1.11.0"
    ports:
      - "9200:9200"
    volumes:
      - ${ES_DATA}:/usr/share/elasticsearch/data
      - ${ES_BACKUPS}:/usr/share/elasticsearch/backups
    environment:
      - node.name=elasticsearch
      - discovery.seed_hosts=elasticsearch
      - cluster.initial_master_nodes=elasticsearch
      - "ES_JAVA_OPTS=-Xms${ES_MEM}g -Xmx${ES_MEM}g"
      - path.repo=/usr/share/elasticsearch

  postgres:
    image: "utmstack.azurecr.io/postgres:13"
    environment:
      - "POSTGRES_PASSWORD=${DB_PASS}"
      - "PGDATA=/var/lib/postgresql/data/pgdata"
      - "shared_buffers=256MB"
      - "max_connections=1000000"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  nginx:
    image: "utmstack.azurecr.io/nginx:1.19.5"
    ports:
      - "443:443"
    volumes:
      - ${NGINX_CERT}:/etc/nginx/cert

  datasources_aws:
    depends_on:
      - logan
    image: "utmstack.azurecr.io/datasources:testing"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
    environment:
      - SERVER_NAME
      - DB_PASS
    command: ["python3", "-m", "utmstack.aws"]

  datasources_azure:
    depends_on:
      - logan
    image: "utmstack.azurecr.io/datasources:testing"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
    environment:
      - SERVER_NAME
      - DB_PASS
    command: ["python3", "-m", "utmstack.azure"]

  datasources_office365:
    depends_on:
      - logan
    image: "utmstack.azurecr.io/datasources:testing"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
    environment:
      - SERVER_NAME
      - DB_PASS
    command: ["python3", "-m", "utmstack.office365"]

  datasources_webroot:
    depends_on:
      - logan
    image: "utmstack.azurecr.io/datasources:testing"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
    environment:
      - SERVER_NAME
      - DB_PASS
    command: ["python3", "-m", "utmstack.webroot"]

  datasources_logan:
    depends_on:
      - panel
    image: "utmstack.azurecr.io/datasources:testing"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
    environment:
      - SERVER_NAME
      - DB_PASS
    ports:
      - "50051:50051"
    command: ["python3", "-m", "utmstack.logan"]

  panel:
    depends_on:
      - postgres
      - openvas
      - elasticsearch
    image: "utmstack.azurecr.io/panel:7.0.0-1"
    environment:
      - TOMCAT_ADMIN_USER=admin
      - TOMCAT_ADMIN_PASSWORD=${DB_PASS}
      - POSTGRESQL_USER=postgres
      - POSTGRESQL_PASSWORD=${DB_PASS}
      - POSTGRESQL_HOST=postgres
      - POSTGRESQL_PORT=5432
      - POSTGRESQL_DATABASE=utmstack
      - ELASTICSEARCH_HOST=elasticsearch
      - ELASTICSEARCH_PORT=9200
      - ELASTICSEARCH_SECRET=${CLIENT_SECRET}
      - OPENVAS_HOST=openvas
      - OPENVAS_PORT=9390
      - OPENVAS_USER=admin
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
`
)

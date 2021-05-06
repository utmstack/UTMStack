package utils

const (
	stackName    = "utmstack"
	composerFile = "utmstack.yml"
	baseTemplate = `version: "3.8"

volumes:
  postgres_data:
    external: false
  wazuh_etc:
    external: false
  wazuh_var:
    external: false
  wazuh_logs:
    external: false
  openvas_data:
    external: false

services:
  openvas:
    image: "utmstack.azurecr.io/openvas:latest"
    volumes:
      - openvas_data:/data
    ports:
      - "8888:5432"
      - "9390:9390"
    environment:
      - USERNAME=admin
      - PASSWORD=${DB_PASS}
      - DB_PASSWORD=${DB_PASS}
      - HTTPS=0
    deploy:
      resources:
        limits:
          cpus: '2.00'
          memory: 1024M

  logstash:
    image: "utmstack.azurecr.io/logstash:latest"
    volumes:
      - ${LOGSTASH_PIPELINE}:/usr/share/logstash/pipeline
    ports:
      - 5044:5044
    environment:
      - CONFIG_RELOAD_AUTOMATIC=true
    deploy:
      resources:
        limits:
          cpus: '4.00'
          memory: 2048M

  rsyslog:
    image: "utmstack.azurecr.io/rsyslog:latest"
    volumes:
      - ${RSYSLOG_LOGS}:/logs
    ports:
      - 514:514
      - 514:514/udp
    deploy:
      resources:
        limits:
          cpus: '2.00'
          memory: 512M

  datasources_mutate:
    image: "utmstack.azurecr.io/datasources:latest"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
      - ${LOGSTASH_PIPELINE}:/usr/share/logstash/pipeline
    environment:
      - SERVER_NAME
      - SERVER_TYPE
      - DB_HOST
      - DB_PASS
      - CORRELATION_URL
    deploy:
      resources:
        limits:
          cpus: '1.00'
          memory: 512M
    command: ["python3", "-m", "utmstack.mutate"]


  datasources_transporter:
    image: "utmstack.azurecr.io/datasources:latest"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
      - ${RSYSLOG_LOGS}:/logs
      - /var/log/suricata:/var/log/suricata
      - wazuh_etc:/var/ossec/etc
      - wazuh_var:/var/ossec/var
      - wazuh_logs:/var/ossec/logs
    environment:
      - SERVER_NAME
      - SERVER_TYPE
      - DB_HOST
      - DB_PASS
    deploy:
      resources:
        limits:
          cpus: '1.00'
          memory: 512M
    command: ["python3", "-m", "utmstack.transporter"]

  datasources_netflow:
    image: "utmstack.azurecr.io/datasources:latest"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
    ports:
      - "2055:2055/udp"
    environment:
      - SERVER_NAME
      - SERVER_TYPE
      - DB_HOST
      - DB_PASS
    deploy:
      resources:
        limits:
          cpus: '1.00'
          memory: 512M
    command: ["python3", "-m", "utmstack.netflow"]

  datasources_probe_api:
    image: "utmstack.azurecr.io/datasources:latest"
    volumes:
      - wazuh_etc:/var/ossec/etc
      - wazuh_var:/var/ossec/var
      - wazuh_logs:/var/ossec/logs
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
    environment:
      - SERVER_NAME
      - SERVER_TYPE
      - DB_HOST
      - DB_PASS
      - SCANNER_IP
      - SCANNER_IFACE
    ports:
      - 23949:23949
      - 1514:1514
      - 1514:1514/udp
      - 1515:1515
      - 1516:1516
      - 55000:55000
    deploy:
      resources:
        limits:
          cpus: '1.00'
          memory: 512M
    command: ["/pw.sh"]
`
	masterTemplate = baseTemplate + `
  elasticsearch:
    image: "utmstack.azurecr.io/opendistro:latest"
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
    image: "utmstack.azurecr.io/postgres:latest"
    environment:
      - "POSTGRES_PASSWORD=${DB_PASS}"
      - "PGDATA=/var/lib/postgresql/data/pgdata"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    command: ["postgres", "-c", "shared_buffers=256MB", "-c", "max_connections=1000"]

  nginx:
    image: "utmstack.azurecr.io/nginx:latest"
    ports:
      - "443:443"
    volumes:
      - ${NGINX_CERT}:/etc/nginx/cert
    deploy:
      resources:
        limits:
          cpus: '1.00'
          memory: 512M

  datasources_aws:
    image: "utmstack.azurecr.io/datasources:latest"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
    environment:
      - SERVER_NAME
      - DB_PASS
    deploy:
      resources:
        limits:
          cpus: '1.00'
          memory: 512M
    command: ["python3", "-m", "utmstack.aws"]

  datasources_azure:
    image: "utmstack.azurecr.io/datasources:latest"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
    environment:
      - SERVER_NAME
      - DB_PASS
    deploy:
      resources:
        limits:
          cpus: '1.00'
          memory: 512M
    command: ["python3", "-m", "utmstack.azure"]

  datasources_office365:
    image: "utmstack.azurecr.io/datasources:latest"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
    environment:
      - SERVER_NAME
      - DB_PASS
    deploy:
      resources:
        limits:
          cpus: '1.00'
          memory: 512M
    command: ["python3", "-m", "utmstack.office365"]

  datasources_webroot:
    image: "utmstack.azurecr.io/datasources:latest"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
    environment:
      - SERVER_NAME
      - DB_PASS
    deploy:
      resources:
        limits:
          cpus: '1.00'
          memory: 512M
    command: ["python3", "-m", "utmstack.webroot"]

  datasources_sophos:
    image: "utmstack.azurecr.io/datasources:latest"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
    environment:
      - SERVER_NAME
      - DB_PASS
    deploy:
      resources:
        limits:
          cpus: '1.00'
          memory: 512M
    command: ["python3", "-m", "utmstack.sophos"]

  datasources_logan:
    image: "utmstack.azurecr.io/datasources:latest"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
    environment:
      - SERVER_NAME
      - DB_PASS
    ports:
      - "50051:50051"
    deploy:
      resources:
        limits:
          cpus: '4.00'
          memory: 2048M
    command: ["python3", "-m", "utmstack.logan"]

  panel:
    image: "utmstack.azurecr.io/panel:latest"
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
    deploy:
      resources:
        limits:
          cpus: '1.00'
          memory: 2048M
  
  zapier:
    image: "utmstack.azurecr.io/zapier:latest"
    environment:
      - POSTGRESQL_USER=postgres
      - POSTGRESQL_PASSWORD=${DB_PASS}
      - POSTGRESQL_HOST=postgres
      - POSTGRESQL_PORT=5432
      - POSTGRESQL_DATABASE=utmstack
      - ELASTICSEARCH_HOST=elasticsearch
      - ELASTICSEARCH_PORT=9200
    ports:
      - "9999:8080"
    deploy:
      resources:
        limits:
          cpus: '1.00'
          memory: 512M
  correlation:
    image: "utmstack.azurecr.io/correlation:latest"
    ports:
      - "9090:8080"
    environment:
      - POSTGRESQL_USER=postgres
      - POSTGRESQL_PASSWORD=${DB_PASS}
      - POSTGRESQL_HOST=postgres
      - POSTGRESQL_PORT=5432
      - POSTGRESQL_DATABASE=utmstack
      - ELASTICSEARCH_HOST=elasticsearch
      - ELASTICSEARCH_PORT=9200
`
)

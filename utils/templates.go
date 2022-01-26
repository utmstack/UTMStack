package utils

const (
	stackName     = "utmstack"
	composerFile  = "utmstack.yml"
	probeTemplate = `version: "3.8"

volumes:
  postgres_data:
    external: false
  ossec_logs:
    external: false
  ossec_var:
    external: false
  openvas_data:
    external: false
  geoip_data:
    external: false

services:
  watchtower:
    image: containrrr/watchtower
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - /root/.docker/config.json:/config.json
    environment:
      - WATCHTOWER_NO_RESTART=true
      - WATCHTOWER_POLL_INTERVAL=${UPDATES}

  logstash:
    image: "utmstack.azurecr.io/logstash:${TAG}"
    volumes:
	    - ${UTMSTACK_DATASOURCES}:/etc/utmstack
      - ${LOGSTASH_PIPELINE}:/usr/share/logstash/pipeline
      - /var/log/suricata:/var/log/suricata
      - ossec_logs:/var/ossec/logs
      - ${CERT}:/cert
    ports:
      - 5044:5044
      - 8089:8089
      - 514:514
      - 1470:1470
      - 2056:2056
      - 2055:2055/udp
    environment:
      - CONFIG_RELOAD_AUTOMATIC=true
      - "LS_JAVA_OPTS=-Xms${LS_MEM}g -Xmx${LS_MEM}g"
      - PIPELINE_WORKERS=6

  datasources_mutate:
    image: "utmstack.azurecr.io/datasources:${TAG}"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
      - ${LOGSTASH_PIPELINE}:/usr/share/logstash/pipeline
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - SERVER_NAME
      - SERVER_TYPE
      - DB_HOST
      - DB_PASS
      - CORRELATION_URL
    command: ["python3", "-m", "utmstack.mutate"]

  datasources_cleaner:
    image: "utmstack.azurecr.io/datasources:${TAG}"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
      - /var/log/suricata:/var/log/suricata
      - ossec_var:/var/ossec/var
      - ossec_logs:/var/ossec/logs
    environment:
      - SERVER_NAME
      - SERVER_TYPE
      - DB_HOST
      - DB_PASS
    command: ["python3", "-m", "utmstack.cleaner"]

  datasources_probe_api:
    image: "utmstack.azurecr.io/datasources:${TAG}"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
      - ${CERT}:/cert
    environment:
      - SERVER_NAME
      - SERVER_TYPE
      - DB_HOST
      - DB_PASS
      - SCANNER_IP
      - SCANNER_IFACE
    command: ["/pw.sh"]

  datasources_agent_manager:
    image: "utmstack.azurecr.io/agent-manager:${TAG}"
    volumes:
      - ${CERT}:/cert
    ports:
      - "9000:9000"
    environment:
      - DB_HOST
      - DB_PASS
    depends_on:
      - "node1"
      - "postgres"
      - "panel"
      - "wazuh"
    command: ["/run.sh"]

  wazuh:
    image: "utmstack.azurecr.io/wazuh:${TAG}"
    ports:
      - "1514:1514"
      - "1515:1515"
      - "55000:55000"
    volumes:
      - ossec_logs:/var/ossec/logs
      - ossec_var:/var/ossec
`
	masterTemplate = probeTemplate + `
  node1:
    image: "utmstack.azurecr.io/opendistro:${TAG}"
    ports:
      - "9200:9200"
    volumes:
      - ${ES_DATA}:/usr/share/elasticsearch/data
      - ${ES_BACKUPS}:/usr/share/elasticsearch/backups
    environment:
      - node.name=node1
      - discovery.seed_hosts=node1
      - cluster.initial_master_nodes=node1
      - "ES_JAVA_OPTS=-Xms${ES_MEM}g -Xmx${ES_MEM}g"
      - path.repo=/usr/share/elasticsearch

  postgres:
    image: "utmstack.azurecr.io/postgres:${TAG}"
    environment:
      - "POSTGRES_PASSWORD=${DB_PASS}"
      - "PGDATA=/var/lib/postgresql/data/pgdata"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    command: ["postgres", "-c", "shared_buffers=256MB", "-c", "max_connections=1000"]

  nginx:
    image: "utmstack.azurecr.io/utmstack_frontend:${TAG}"
    depends_on:
      - "panel"
      - "filebrowser"
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ${CERT}:/etc/nginx/cert

  datasources_aws:
    image: "utmstack.azurecr.io/datasources:${TAG}"
    depends_on:
      - "node1"
      - "postgres"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
    environment:
      - SERVER_NAME
      - DB_HOST
      - DB_PASS
    command: ["python3", "-m", "utmstack.aws"]

  datasources_office365:
    image: "utmstack.azurecr.io/datasources:${TAG}"
    depends_on:
      - "node1"
      - "postgres"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
    environment:
      - SERVER_NAME
      - DB_HOST
      - DB_PASS
    command: ["python3", "-m", "utmstack.office365"]

  datasources_webroot:
    image: "utmstack.azurecr.io/datasources:${TAG}"
    depends_on:
      - "node1"
      - "postgres"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
    environment:
      - SERVER_NAME
      - DB_HOST
      - DB_PASS
    command: ["python3", "-m", "utmstack.webroot"]

  datasources_sophos:
    image: "utmstack.azurecr.io/datasources:${TAG}"
    depends_on:
      - "node1"
      - "postgres"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
    environment:
      - SERVER_NAME
      - DB_HOST
      - DB_PASS
    command: ["python3", "-m", "utmstack.sophos"]

  datasources_logan:
    image: "utmstack.azurecr.io/datasources:${TAG}"
    depends_on:
      - "node1"
      - "postgres"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
    environment:
      - SERVER_NAME
      - DB_HOST
      - DB_PASS
    ports:
      - "50051:50051"
    command: ["python3", "-m", "utmstack.logan"]

  panel:
    image: "utmstack.azurecr.io/utmstack_backend:${TAG}"
    depends_on:
      - "node1"
      - "postgres"
    environment:
      - SERVER_NAME
      - LITE
      - DB_USER=postgres
      - DB_PASS
      - DB_HOST
      - DB_PORT=5432
      - DB_NAME=utmstack
      - ELASTICSEARCH_HOST=${DB_HOST}
      - ELASTICSEARCH_PORT=9200

  correlation:
    image: "utmstack.azurecr.io/correlation:${TAG}"
    volumes:
      - geoip_data:/app/geosets
      - ${UTMSTACK_RULES}:/app/rulesets/custom
    ports:
      - "9090:8080"
    environment:
      - SERVER_NAME
      - POSTGRESQL_USER=postgres
      - POSTGRESQL_PASSWORD=${DB_PASS}
      - POSTGRESQL_HOST=${DB_HOST}
      - POSTGRESQL_PORT=5432
      - POSTGRESQL_DATABASE=utmstack
      - ELASTICSEARCH_HOST=${DB_HOST}
      - ELASTICSEARCH_PORT=9200
      - ERROR_LEVEL=info
    depends_on:
      - "node1"
      - "postgres"

  filebrowser:
    image: "utmstack.azurecr.io/filebrowser:${TAG}"
    volumes:
      - ${UTMSTACK_RULES}:/srv
    environment:
      - PASSWORD=${DB_PASS}
    deploy:
      resources:
        limits:
          cpus: '1.00'
          memory: 512M
`
	openvasTemplate = `
  openvas:
    image: "utmstack.azurecr.io/openvas:${TAG}"
    volumes:
      - openvas_data:/data
    ports:
      - "8888:5432"
      - "9390:9390"
      - "9392:9392"
    environment:
      - USERNAME=admin
      - PASSWORD=${DB_PASS}
      - DB_PASSWORD=${DB_PASS}
      - HTTPS=0
    deploy:
      resources:
        limits:
          cpus: '2.00'
          memory: 2048M`

	probeTemplateStandard  = probeTemplate + openvasTemplate
	masterTemplateStandard = probeTemplateStandard + masterTemplate
)

package utils

const (
	stackName     = "utmstack"
	composerFile  = "utmstack.yml"
	probeTemplate = `version: "3.8"

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
      - ${LOGSTASH_PIPELINE}:/usr/share/logstash/pipeline
      - /var/log/suricata:/var/log/suricata
      - wazuh_logs:/var/ossec/logs
      - ${CERT}:/cert
    ports:
      - 5044:5044
      - 8089:8089
      - 514:514
      - 514:514/udp
      - 2055:2055/udp
    environment:
      - CONFIG_RELOAD_AUTOMATIC=true
    deploy:
      resources:
        limits:
          cpus: '2.00'
          memory: 2048M

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
    deploy:
      resources:
        limits:
          cpus: '1.00'
          memory: 512M
    command: ["python3", "-m", "utmstack.mutate"]

  datasources_transporter:
    image: "utmstack.azurecr.io/datasources:${TAG}"
    volumes:
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
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

  datasources_probe_api:
    image: "utmstack.azurecr.io/datasources:${TAG}"
    volumes:
      - wazuh_etc:/var/ossec/etc
      - wazuh_var:/var/ossec/var
      - wazuh_logs:/var/ossec/logs
      - ${UTMSTACK_DATASOURCES}:/etc/utmstack
      - ${CERT}:/cert
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
          cpus: '2.00'
          memory: 1024M
    command: ["/pw.sh"]
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
    deploy:
      resources:
        limits:
          cpus: '2.00'
          memory: 1024M

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
    deploy:
      resources:
        limits:
          cpus: '1.00'
          memory: 1024M
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
    deploy:
      resources:
        limits:
          cpus: '1.00'
          memory: 1024M
    command: ["python3", "-m", "utmstack.office365"]
  
  datasources_azure:
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
    deploy:
      resources:
        limits:
          cpus: '1.00'
          memory: 1024M
    command: ["python3", "-m", "utmstack.azure"]

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
    deploy:
      resources:
        limits:
          cpus: '1.00'
          memory: 1024M
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
    deploy:
      resources:
        limits:
          cpus: '1.00'
          memory: 1024M
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
    deploy:
      resources:
        limits:
          cpus: '1.00'
          memory: 1024M
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
    deploy:
      resources:
        limits:
          cpus: '2.00'
          memory: 2048M

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

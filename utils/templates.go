package utils

const (
	composerFile  = "docker-compose.yml"
	probeTemplateLite = `version: "3"

volumes:
  postgres_data:
  ossec_logs:
  ossec_var:
  openvas_data:
  geoip_data:

services:
  watchtower:
    container_name: watchtower
    restart: always
    image: containrrr/watchtower
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - /root/.docker/config.json:/config.json
    environment:
      - WATCHTOWER_NO_RESTART=false
      - WATCHTOWER_POLL_INTERVAL=${UPDATES}

  logstash:
    container_name: logstash
    restart: always
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
      - 514:514/udp
      - 1470:1470
      - 2056:2056
      - 2055:2055/udp
    environment:
      - CONFIG_RELOAD_AUTOMATIC=true
      - "LS_JAVA_OPTS=-Xms${LS_MEM}g -Xmx${LS_MEM}g -Xss100m"
      - PIPELINE_WORKERS=4
    depends_on:
      - "datasources_mutate"

  datasources_mutate:
    container_name: datasources_mutate
    restart: always
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
    container_name: datasources_transporter
    restart: always
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
    container_name: datasources_probe_api
    restart: always
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
    container_name: datasources_agent_manager
    restart: always
    image: "utmstack.azurecr.io/agent-manager:${TAG}"
    volumes:
      - ${CERT}:/cert
    ports:
      - "9000:9000"
    environment:
      - DB_HOST
      - DB_PASS
      - SCANNER_IP
    depends_on:
      - "wazuh"
    command: ["/run.sh"]

  wazuh:
    container_name: wazuh
    restart: always
    image: "utmstack.azurecr.io/wazuh:${TAG}"
    ports:
      - "1514:1514"
      - "1515:1515"
      - "55000:55000"
    volumes:
      - ossec_logs:/var/ossec/logs
      - ossec_var:/var/ossec
`
	masterTemplate = `
  node1:
    container_name: node1
    restart: always
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
    container_name: postgres
    restart: always
    image: "utmstack.azurecr.io/postgres:${TAG}"
    environment:
      - "POSTGRES_PASSWORD=${DB_PASS}"
      - "PGDATA=/var/lib/postgresql/data/pgdata"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    command: ["postgres", "-c", "shared_buffers=256MB", "-c", "max_connections=1000"]

  frontend:
    container_name: frontend
    restart: always
    image: "utmstack.azurecr.io/utmstack_frontend:${TAG}"
    depends_on:
      - "backend"
      - "filebrowser"
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ${CERT}:/etc/nginx/cert

  datasources_aws:
    container_name: datasources_aws
    restart: always
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
    container_name: datasources_office365
    restart: always
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
    container_name: datasources_webroot
    restart: always
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
    container_name: datasources_sophos
    restart: always
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
    container_name: datasources_logan
    restart: always
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

  backend:
    container_name: backend
    restart: always
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
    container_name: correlation
    restart: always
    image: "utmstack.azurecr.io/correlation:${TAG}"
    volumes:
      - geoip_data:/app/geosets
      - ${UTMSTACK_RULES}:/app/rulesets/custom
    ports:
      - "9090:8080"
    environment:
      - SERVER_NAME
      - POSTGRESQL_USER=postgres
      - "POSTGRESQL_PASSWORD=${DB_PASS}"
      - POSTGRESQL_HOST=${DB_HOST}
      - POSTGRESQL_PORT=5432
      - POSTGRESQL_DATABASE=utmstack
      - ELASTICSEARCH_HOST=${DB_HOST}
      - ELASTICSEARCH_PORT=9200
      - ERROR_LEVEL=info
    depends_on:
      - "node1"
      - "postgres"
      - "backend"

  filebrowser:
    container_name: filebrowser
    restart: always
    image: "utmstack.azurecr.io/filebrowser:${TAG}"
    volumes:
      - ${UTMSTACK_RULES}:/srv
    environment:
      - "PASSWORD=${DB_PASS}"
`
	openvasTemplate = `
  openvas:
    container_name: openvas
    restart: always
    image: "utmstack.azurecr.io/openvas:${TAG}"
    volumes:
      - openvas_data:/data
    ports:
      - "8888:5432"
      - "9390:9390"
      - "9392:9392"
    environment:
      - USERNAME=admin
      - "PASSWORD=${DB_PASS}"
      - "DB_PASSWORD=${DB_PASS}"
      - HTTPS=0
`
  probeTemplateStandard  = probeTemplateLite + openvasTemplate
  masterTemplateStandard = probeTemplateLite + masterTemplate + openvasTemplate
  masterTemplateLite     = probeTemplateLite + masterTemplate
  ufw = `# rules.input-after
#
# Rules that should be run after the ufw command line added rules. Custom
# rules should be added to one of these chains:
#   ufw-after-input
#   ufw-after-output
#   ufw-after-forward
#

# Don't delete these required lines, otherwise there will be errors
*filter
:ufw-after-input - [0:0]
:ufw-after-output - [0:0]
:ufw-after-forward - [0:0]
:ufw-user-forward - [0:0]
:DOCKER-USER - [0:0]
# End required lines

# ALLOW SSH AND OPENVPN PORTS
-A INPUT -p tcp -m tcp --dport 22 -j ACCEPT

# ALLOW ALL FROM VPN AND DOCKER SUBNETS
-A INPUT -s 10.21.199.0/24 -j ACCEPT
-A INPUT -s 172.17.0.0/16 -j ACCEPT
-A INPUT -s 172.18.0.0/16 -j ACCEPT
-A DOCKER-USER -s 10.21.199.0/24 -j ACCEPT
-A DOCKER-USER -s 172.17.0.0/16 -j ACCEPT
-A DOCKER-USER -s 172.18.0.0/16 -j ACCEPT

# SECURING ELASTIC AND POSTGRES
-A DOCKER-USER -p tcp -m tcp --dport 5432 -j DROP
-A DOCKER-USER -p tcp -m tcp --dport 9200 -j DROP
-A DOCKER-USER -p tcp -m tcp --dport 8888 -j DROP
-A DOCKER-USER -p tcp -m tcp --dport 9390 -j DROP
-A DOCKER-USER -p tcp -m tcp --dport 9392 -j DROP
-A DOCKER-USER -p tcp -m tcp --dport 55000 -j DROP

# don't log noisy services by default
-A ufw-after-input -p udp --dport 137 -j ufw-skip-to-policy-input
-A ufw-after-input -p udp --dport 138 -j ufw-skip-to-policy-input
-A ufw-after-input -p tcp --dport 139 -j ufw-skip-to-policy-input
-A ufw-after-input -p tcp --dport 445 -j ufw-skip-to-policy-input
-A ufw-after-input -p udp --dport 67 -j ufw-skip-to-policy-input
-A ufw-after-input -p udp --dport 68 -j ufw-skip-to-policy-input

# don't log noisy broadcast
-A ufw-after-input -m addrtype --dst-type BROADCAST -j ufw-skip-to-policy-input

# don't delete the 'COMMIT' line or these rules won't be processed
COMMIT
`
)

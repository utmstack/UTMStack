package utils

const (
	composerFile          = "compose.yml"
	probeTemplateStandard = `version: "3"

volumes:
  postgres_data:
  
  ossec_logs:
  ossec_var:
  openvas_data:
  geoip_data:
  updates:

services:
  watchtower:
    restart: always
    image: containrrr/watchtower
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - /root/.docker/config.json:/config.json
    environment:
      - WATCHTOWER_NO_RESTART=false
      - WATCHTOWER_POLL_INTERVAL={{.Updates}}
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
    deploy:
      placement:
        constraints:
          - node.role == manager

  logstash:
    restart: always
    image: "utmstack.azurecr.io/logstash:{{.Tag}}"
    volumes:
      - {{.Datasources}}:/etc/utmstack
      - {{.LogstashPipeline}}:/usr/share/logstash/pipeline
      - /var/log/suricata:/var/log/suricata
      - ossec_logs:/var/ossec/logs
      - {{.Cert}}:/cert
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
      - "LS_JAVA_OPTS=-Xms{{.LSMem}}g -Xmx{{.LSMem}}g -Xss100m"
      - PIPELINE_WORKERS=4
    depends_on:
      - "datasources_mutate"
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
    deploy:
      placement:
        constraints:
          - node.role == manager

  datasources_mutate:
    restart: always
    image: "utmstack.azurecr.io/datasources:{{.Tag}}"
    volumes:
      - {{.Datasources}}:/etc/utmstack
      - {{.LogstashPipeline}}:/usr/share/logstash/pipeline
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - SERVER_NAME={{.ServerName}}
      - SERVER_TYPE={{.ServerType}}
      - DB_HOST={{.DBHost}}
      - "DB_PASS={{.DBPass}}"
      - CORRELATION_URL={{.Correlation}}
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
    deploy:
      placement:
        constraints:
          - node.role == manager
    command: ["python3", "-m", "utmstack.mutate"]

  datasources_cleaner:
    restart: always
    image: "utmstack.azurecr.io/datasources:{{.Tag}}"
    volumes:
      - {{.Datasources}}:/etc/utmstack
      - /var/log/suricata:/var/log/suricata
      - ossec_var:/var/ossec/var
      - ossec_logs:/var/ossec/logs
    environment:
      - SERVER_NAME={{.ServerName}}
      - SERVER_TYPE={{.ServerType}}
      - DB_HOST={{.DBHost}}
      - "DB_PASS={{.DBPass}}"
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
    deploy:
      placement:
        constraints:
          - node.role == manager
    command: ["python3", "-m", "utmstack.cleaner"]

  datasources_probe_api:
    restart: always
    image: "utmstack.azurecr.io/datasources:{{.Tag}}"
    volumes:
      - {{.Datasources}}:/etc/utmstack
      - {{.Cert}}:/cert
    environment:
      - SERVER_NAME={{.ServerName}}
      - SERVER_TYPE={{.ServerType}}
      - DB_HOST={{.DBHost}}
      - "DB_PASS={{.DBPass}}"
      - SCANNER_IP={{.ScannerIP}}
      - SCANNER_IFACE={{.ScannerIface}}
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
    deploy:
      placement:
        constraints:
          - node.role == manager
    command: ["/pw.sh"]

  datasources_agent_manager:
    restart: always
    image: "utmstack.azurecr.io/agent-manager:{{.Tag}}"
    volumes:
      - {{.Cert}}:/cert
      - {{.Datasources}}:/etc/utmstack
    ports:
      - "9000:9000"
    environment:
      - DB_HOST={{.DBHost}}
      - "DB_PASS={{.DBPass}}"
      - SCANNER_IP={{.ScannerIP}}
    depends_on:
      - "wazuh"
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
    deploy:
      placement:
        constraints:
          - node.role == manager
    command: ["/run.sh"]
`
	masterTemplate = `
  node1:
    restart: always
    image: "utmstack.azurecr.io/opensearch:{{.Tag}}"
    ports:
      - "127.0.0.1:9200:9200"
    volumes:
      - {{.ESData}}:/usr/share/elasticsearch/data
      - {{.ESBackups}}:/usr/share/elasticsearch/backups
    environment:
      - node.name=node1
      - discovery.seed_hosts=node1
      - cluster.initial_master_nodes=node1
      - "ES_JAVA_OPTS=-Xms{{.ESMem}}g -Xmx{{.ESMem}}g"
      - path.repo=/usr/share/elasticsearch
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
    deploy:
      placement:
        constraints:
          - node.role == manager

  postgres:
    restart: always
    image: "utmstack.azurecr.io/postgres:{{.Tag}}"
    environment:
      - "POSTGRES_PASSWORD={{.DBPass}}"
      - "PGDATA=/var/lib/postgresql/data/pgdata"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "127.0.0.1:5432:5432"
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
    deploy:
      placement:
        constraints:
          - node.role == manager
    command: ["postgres", "-c", "shared_buffers=256MB", "-c", "max_connections=1000"]

  frontend:
    restart: always
    image: "utmstack.azurecr.io/utmstack_frontend:{{.Tag}}"
    depends_on:
      - "backend"
      - "filebrowser"
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - {{.Cert}}:/etc/nginx/cert
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
    deploy:
      placement:
        constraints:
          - node.role == manager

  datasources_aws:
    restart: always
    image: "utmstack.azurecr.io/datasources:{{.Tag}}"
    depends_on:
      - "node1"
      - "postgres"
    volumes:
      - {{.Datasources}}:/etc/utmstack
    environment:
      - SERVER_NAME={{.ServerName}}
      - "DB_PASS={{.DBPass}}"
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
    deploy:
      placement:
        constraints:
          - node.role == manager
    command: ["python3", "-m", "utmstack.aws"]

  datasources_office365:
    restart: always
    image: "utmstack.azurecr.io/datasources:{{.Tag}}"
    depends_on:
      - "node1"
      - "postgres"
    volumes:
      - {{.Datasources}}:/etc/utmstack
    environment:
      - SERVER_NAME={{.ServerName}}
      - "DB_PASS={{.DBPass}}"
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
    deploy:
      placement:
        constraints:
          - node.role == manager
    command: ["python3", "-m", "utmstack.office365"]

  datasources_sophos:
    restart: always
    image: "utmstack.azurecr.io/datasources:{{.Tag}}"
    depends_on:
      - "node1"
      - "postgres"
    volumes:
      - {{.Datasources}}:/etc/utmstack
    environment:
      - SERVER_NAME={{.ServerName}}
      - "DB_PASS={{.DBPass}}"
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
    deploy:
      placement:
        constraints:
          - node.role == manager
    command: ["python3", "-m", "utmstack.sophos"]

  datasources_logan:
    restart: always
    image: "utmstack.azurecr.io/datasources:{{.Tag}}"
    depends_on:
      - "node1"
      - "postgres"
    volumes:
      - {{.Datasources}}:/etc/utmstack
    environment:
      - SERVER_NAME={{.ServerName}}
      - "DB_PASS={{.DBPass}}"
    ports:
      - "50051:50051"
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
    deploy:
      placement:
        constraints:
          - node.role == manager
    command: ["python3", "-m", "utmstack.logan"]

  backend:
    restart: always
    image: "utmstack.azurecr.io/utmstack_backend:{{.Tag}}"
    depends_on:
      - "node1"
      - "postgres"
    environment:
      - SERVER_NAME={{.ServerName}}
      - LITE={{.Lite}}
      - DB_USER=postgres
      - "DB_PASS={{.DBPass}}"
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=utmstack
      - ELASTICSEARCH_HOST=node1
      - ELASTICSEARCH_PORT=9200
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
    deploy:
      placement:
        constraints:
          - node.role == manager

  correlation:
    restart: always
    image: "utmstack.azurecr.io/correlation:{{.Tag}}"
    volumes:
      - geoip_data:/app/geosets
      - {{.Rules}}:/app/rulesets
    ports:
      - "9090:8080"
    environment:
      - SERVER_NAME={{.ServerName}}
      - POSTGRESQL_USER=postgres
      - "POSTGRESQL_PASSWORD={{.DBPass}}"
      - POSTGRESQL_HOST=postgres
      - POSTGRESQL_PORT=5432
      - POSTGRESQL_DATABASE=utmstack
      - ELASTICSEARCH_HOST=node1
      - ELASTICSEARCH_PORT=9200
      - ERROR_LEVEL=info
    depends_on:
      - "node1"
      - "postgres"
      - "backend"
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
    deploy:
      placement:
        constraints:
          - node.role == manager

  filebrowser:
    restart: always
    image: "utmstack.azurecr.io/filebrowser:{{.Tag}}"
    volumes:
      - {{.Rules}}:/srv
    environment:
      - "PASSWORD={{.DBPass}}"
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
    deploy:
      placement:
        constraints:
          - node.role == manager
`
	masterTemplateStandard = probeTemplateStandard + masterTemplate
)

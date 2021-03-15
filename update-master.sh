#!/bin/bash

docker pull utmstack.azurecr.io/opendistro:latest
docker pull utmstack.azurecr.io/openvas:latest
docker pull utmstack.azurecr.io/postgres:latest
docker pull utmstack.azurecr.io/logstash:latest
docker pull utmstack.azurecr.io/rsyslog:latest
docker pull utmstack.azurecr.io/scanner:latest
docker pull utmstack.azurecr.io/nginx:latest
docker pull utmstack.azurecr.io/panel:latest
docker pull utmstack.azurecr.io/datasources:latest
docker pull utmstack.azurecr.io/zapier:latest

docker service update --image utmstack.azurecr.io/opendistro:latest utmstack_elasticsearch
docker service update --image utmstack.azurecr.io/openvas:latest utmstack_openvas
docker service update --image utmstack.azurecr.io/postgres:latest utmstack_postgres
docker service update --image utmstack.azurecr.io/logstash:latest utmstack_logstash
docker service update --image utmstack.azurecr.io/rsyslog:latest utmstack_rsyslog
docker service update --image utmstack.azurecr.io/scanner:latest utmstack_scanner
docker service update --image utmstack.azurecr.io/nginx:latest utmstack_nginx
docker service update --image utmstack.azurecr.io/panel:latest utmstack_panel
docker service update --image utmstack.azurecr.io/zapier:latest utmstack_zapier
docker service update --image utmstack.azurecr.io/datasources:latest utmstack_datasources_logan
docker service update --image utmstack.azurecr.io/datasources:latest utmstack_datasources_webroot
docker service update --image utmstack.azurecr.io/datasources:latest utmstack_datasources_office365
docker service update --image utmstack.azurecr.io/datasources:latest utmstack_datasources_azure
docker service update --image utmstack.azurecr.io/datasources:latest utmstack_datasources_aws
docker service update --image utmstack.azurecr.io/datasources:latest utmstack_datasources_probe_api
docker service update --image utmstack.azurecr.io/datasources:latest utmstack_datasources_transporter
docker service update --image utmstack.azurecr.io/datasources:latest utmstack_datasources_mutate
docker service update --image utmstack.azurecr.io/datasources:latest utmstack_datasources_openvas


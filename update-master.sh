#!/bin/bash

docker pull opendistro:latest
docker pull openvas:latest
docker pull postgres:latest
docker pull logstash:latest
docker pull rsyslog:latest
docker pull scanner:latest
docker pull nginx:latest
docker pull panel:latest
docker pull datasources:latest
docker pull zapier:latest

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


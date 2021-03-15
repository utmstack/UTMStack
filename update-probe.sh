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

docker service update --image utmstack.azurecr.io/openvas:latest utmstack_openvas
docker service update --image utmstack.azurecr.io/logstash:latest utmstack_logstash
docker service update --image utmstack.azurecr.io/rsyslog:latest utmstack_rsyslog
docker service update --image utmstack.azurecr.io/scanner:latest utmstack_scanner
docker service update --image utmstack.azurecr.io/datasources:latest utmstack_datasources_probe_api
docker service update --image utmstack.azurecr.io/datasources:latest utmstack_datasources_transporter
docker service update --image utmstack.azurecr.io/datasources:latest utmstack_datasources_mutate
docker service update --image utmstack.azurecr.io/datasources:latest utmstack_datasources_openvas

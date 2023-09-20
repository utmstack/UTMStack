FROM docker.elastic.co/logstash/logstash-oss:7.17.4-amd64
USER root
RUN bin/logstash-plugin install logstash-input-http
RUN bin/logstash-plugin install logstash-output-http
RUN bin/logstash-plugin install logstash-input-syslog
RUN bin/logstash-plugin install logstash-codec-netflow
RUN bin/logstash-plugin install logstash-input-google_pubsub
RUN bin/logstash-plugin install logstash-input-azure_event_hubs

HEALTHCHECK --start-period=300s --interval=60s --timeout=60s --retries=1 \
  CMD if curl -v --silent 'localhost:9600/_node/hot_threads?human' 2>&1|grep -Fwq "[main]>worker0"; then exit 0; exit 1; fi || exit 1
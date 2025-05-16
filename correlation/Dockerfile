FROM ubuntu:24.04
RUN apt-get update
RUN apt-get install -y ca-certificates git wget
COPY correlation /app/
COPY docs/swagger.json /app/docs/
COPY docs/swagger.yaml /app/docs/
COPY config.yml.prod /app/config.yml
COPY run.sh /
RUN chmod +x /app/correlation
RUN chmod +x /run.sh
RUN update-ca-certificates
RUN wget -O /app/asn-blocks-v4.csv https://cdn.utmstack.com/geoip/asn-blocks-v4.csv
RUN wget -O /app/asn-blocks-v6.csv https://cdn.utmstack.com/geoip/asn-blocks-v6.csv
RUN wget -O /app/blocks-v4.csv https://cdn.utmstack.com/geoip/blocks-v4.csv
RUN wget -O /app/blocks-v6.csv https://cdn.utmstack.com/geoip/blocks-v6.csv
RUN wget -O /app/locations-en.csv https://cdn.utmstack.com/geoip/locations-en.csv
RUN wget -O /app/ip_level1.list.tar.gz https://intelligence.threatwinds.com/api/feeds/v1/download/list/level1/accumulative/ip && cd /app && tar -xf ip_level1.list.tar.gz
RUN wget -O /app/ip_level2.list.tar.gz https://intelligence.threatwinds.com/api/feeds/v1/download/list/level2/accumulative/ip && cd /app && tar -xf ip_level2.list.tar.gz
RUN wget -O /app/ip_level3.list.tar.gz https://intelligence.threatwinds.com/api/feeds/v1/download/list/level3/accumulative/ip && cd /app && tar -xf ip_level3.list.tar.gz
RUN mkdir -p /app/rulesets && git clone --depth 1 https://github.com/utmstack/rules.git /app/rulesets/system
ENTRYPOINT [ "/run.sh" ]

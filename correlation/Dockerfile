FROM ubuntu:20.04
RUN apt update
RUN apt install -y ca-certificates git
COPY UTMStackCorrelationEngine /app/
COPY docs/swagger.json /app/docs/
COPY docs/swagger.yaml /app/docs/
COPY config.yml.prod /app/config.yml
COPY run.sh /
RUN chmod +x /app/UTMStackCorrelationEngine
RUN chmod +x /run.sh
RUN update-ca-certificates
ENTRYPOINT [ "/run.sh" ]
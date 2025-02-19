FROM ubuntu:22.04
RUN apt update
RUN apt install -y ca-certificates git
COPY correlation /app/
COPY docs/swagger.json /app/docs/
COPY docs/swagger.yaml /app/docs/
COPY config.yml.prod /app/config.yml
COPY run.sh /
RUN chmod +x /app/correlation
RUN chmod +x /run.sh
RUN update-ca-certificates
ENTRYPOINT [ "/run.sh" ]
FROM ubuntu:24.04

COPY soc-ai /app/

RUN apt-get update && \
    apt-get install -y ca-certificates jq wget && \
    update-ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

EXPOSE 8080

CMD ["/app/soc-ai"]
FROM ubuntu:24.04

RUN apt-get update && \
    apt-get install -y ca-certificates curl && \
    update-ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

COPY azure /app/azure

WORKDIR /app

# The module configuration comes from NATS (NATS_URL, decrypted with ENCRYPTION_KEY), and the logs
# go straight to NATS instead of through the event processor. No configuration folder is mounted.
ENV LOGS_TRANSPORT=nats

CMD ["/app/azure"]

FROM ubuntu:24.04

RUN apt-get update && \
    apt-get install -y ca-certificates curl && \
    update-ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

COPY log-input /app/log-input

WORKDIR /app

# Where agents, collectors and direct pushers connect.
EXPOSE 50051

# Health is plain HTTP on its own port: the gRPC port is behind the auth
# interceptors, so a probe there would have to carry credentials.
EXPOSE 8080

# Fails while draining, so nothing new is routed to a replica on its way out.
HEALTHCHECK --start-period=15s --interval=30s --timeout=10s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

CMD ["/app/log-input"]

FROM ubuntu:24.04

COPY agent-manager /app/
COPY ./dependencies/agent/ /dependencies/agent/
COPY ./dependencies/collector/ /dependencies/collector/

# Install jq
RUN apt-get update && \
    apt-get install -y ca-certificates jq wget && \
    update-ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install grpcurl
RUN GRPCURL_VERSION=1.8.1 && \
    wget -O /tmp/grpcurl.tar.gz https://github.com/fullstorydev/grpcurl/releases/download/v${GRPCURL_VERSION}/grpcurl_${GRPCURL_VERSION}_linux_x86_64.tar.gz && \
    tar -xzf /tmp/grpcurl.tar.gz -C /usr/local/bin && \
    rm /tmp/grpcurl.tar.gz && \
    chmod +x /usr/local/bin/grpcurl


# Expose the gRPC agent-manager port
EXPOSE 50051
EXPOSE 8080

# Set the health check
# HEALTHCHECK --interval=60s --timeout=5s --start-period=5s --retries=3 CMD grpcurl -insecure -plaintext -d '{"service": ""}' localhost:50051 grpc.health.v1.Health/Check | jq -e '.status == "SERVING"' || exit 1


# Run the agent-manager binary
CMD ["/app/agent-manager"]

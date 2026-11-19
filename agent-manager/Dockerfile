FROM debian:trixie-slim

COPY agent-manager /app/
COPY ./dependencies/agent/ /dependencies/agent/
COPY ./dependencies/collector/ /dependencies/collector/

# Install jq + wget (ca-certificates for grpcurl's TLS)
RUN apt-get update -qq && \
    apt-get install -y --no-install-recommends ca-certificates jq wget && \
    update-ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Install grpcurl
RUN GRPCURL_VERSION=1.8.1 && \
    wget -O /tmp/grpcurl.tar.gz https://github.com/fullstorydev/grpcurl/releases/download/v${GRPCURL_VERSION}/grpcurl_${GRPCURL_VERSION}_linux_x86_64.tar.gz && \
    tar -xzf /tmp/grpcurl.tar.gz -C /usr/local/bin && \
    rm /tmp/grpcurl.tar.gz && \
    chmod +x /usr/local/bin/grpcurl

# Expose the gRPC agent-manager port
EXPOSE 9000
EXPOSE 9001

# Run the agent-manager binary
CMD ["/app/agent-manager"]

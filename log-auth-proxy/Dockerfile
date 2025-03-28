FROM ubuntu:24.04

COPY log-auth-proxy /app/

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


# Expose the gRPC log-auth-proxy port
EXPOSE 50051
EXPOSE 8080

# Run the log-auth-proxy binary
CMD ["/app/log-auth-proxy"]

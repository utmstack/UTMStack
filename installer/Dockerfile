# Builds the UTMStack installer and runs it inside a privileged Ubuntu container
# with Docker-in-Docker. The host nginx inside the container terminates TLS on
# 443 and proxies to the docker-swarm frontend service on 10001.

############################
# Stage 1: build installer #
############################
FROM golang:1.25-bookworm AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG INSTALLER_VERSION=v11.0.0-dev.1
ARG DEFAULT_BRANCH=release/v12.0.0
ARG CM_ENCRYPT_SALT=6f2d9c4a8b1e7f3d5a0c6e9b2d4f8a1c
ARG CM_SIGN_PUBLIC_KEY=""
RUN GOOS=linux GOARCH=amd64 go build -o /installer -v -ldflags "\
      -X 'github.com/utmstack/UTMStack/installer/config.DEFAULT_BRANCH=${DEFAULT_BRANCH}' \
      -X 'github.com/utmstack/UTMStack/installer/config.INSTALLER_VERSION=${INSTALLER_VERSION}' \
      -X 'github.com/utmstack/UTMStack/installer/config.REPLACE=${CM_ENCRYPT_SALT}' \
      -X 'github.com/utmstack/UTMStack/installer/config.PUBLIC_KEY=${CM_SIGN_PUBLIC_KEY}'" \
    .

##########################
# Stage 2: runtime image #
##########################
FROM ubuntu:22.04
ENV DEBIAN_FRONTEND=noninteractive
ENV LANG=C.UTF-8

RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates curl gnupg lsb-release \
        nginx openssl iproute2 net-tools iptables uidmap kmod \
        systemd systemd-sysv dbus \
    && install -m 0755 -d /etc/apt/keyrings \
    && curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
       | gpg --dearmor -o /etc/apt/keyrings/docker.gpg \
    && chmod a+r /etc/apt/keyrings/docker.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" \
       > /etc/apt/sources.list.d/docker.list \
    && apt-get update && apt-get install -y --no-install-recommends \
        docker-ce docker-ce-cli containerd.io \
        docker-buildx-plugin docker-compose-plugin docker-compose \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

COPY --from=builder /installer /usr/local/bin/utmstack_installer
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/utmstack_installer /usr/local/bin/docker-entrypoint.sh

# 443/80 = host nginx (inside the container) → docker-swarm frontend
# 10001  = docker-swarm frontend (direct, bypassing host nginx)
EXPOSE 80 443 10001 9000 9001

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]

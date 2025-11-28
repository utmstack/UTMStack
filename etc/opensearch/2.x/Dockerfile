# Copyright OpenSearch Contributors
# SPDX-License-Identifier: Apache-2.0


# This dockerfile generates an AmazonLinux-based image containing an OpenSearch installation.
# It assumes that the working directory contains these files: an OpenSearch tarball (opensearch.tgz), log4j2.properties, opensearch.yml, opensearch-docker-entrypoint.sh, opensearch-onetime-setup.sh.
# Build arguments:
#   VERSION: Required. Used to label the image.
#   UID: Optional. Specify the opensearch userid. Defaults to 1000.
#   GID: Optional. Specify the opensearch groupid. Defaults to 1000.
#   OPENSEARCH_HOME: Optional. Specify the opensearch root directory. Defaults to /usr/share/opensearch.


########################### Stage 0 ########################
FROM opensearchproject/opensearch:latest

# Set the working directory
WORKDIR /usr/share/opensearch/

# Create the .utm_geoip folder inside the image
RUN mkdir -p /usr/share/opensearch/.utm_geoip 

# Switch to root user to execute commands with permissions
USER root

# Download the utm-geoip.tar.xz file and extract it
RUN yum update -y && \
    yum install -y xz wget && \
    wget -O /tmp/utm-geoip.tar.xz https://cdn.utmstack.com/geoip/utm-geoip.tar.xz && \
    tar -xJf /tmp/utm-geoip.tar.xz -C /usr/share/opensearch/.utm_geoip && \
    rm /tmp/utm-geoip.tar.xz && \
    yum remove -y xz wget && \
    yum clean all && \
    rm -rf /var/cache/yum && \
    rm -rf /usr/share/man /usr/share/doc /usr/share/info /tmp/* /var/tmp/*

# Assign permissions and ownership to the extracted folder
RUN chmod -R 755 /usr/share/opensearch/.utm_geoip && \
    chown -R opensearch:opensearch /usr/share/opensearch/.utm_geoip

# Restore OpenSearch user
USER opensearch

# Expose default OpenSearch ports
EXPOSE 9200 9600

# Startup command
CMD ["/usr/share/opensearch/opensearch-docker-entrypoint.sh"]

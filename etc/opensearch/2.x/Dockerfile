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
FROM amazonlinux:2 AS linux_stage_0

ARG UID=1000
ARG GID=1000
ARG TEMP_DIR=/tmp/opensearch
ARG OPENSEARCH_HOME=/usr/share/opensearch
ARG OPENSEARCH_PATH_CONF=$OPENSEARCH_HOME/config
ARG SECURITY_PLUGIN_DIR=$OPENSEARCH_HOME/plugins/opensearch-security
ARG PERFORMANCE_ANALYZER_PLUGIN_CONFIG_DIR=$OPENSEARCH_PATH_CONF/opensearch-performance-analyzer
ARG OS_VERSION=2.5.0
# Update packages
# Install the tools we need: tar and gzip to unpack the OpenSearch tarball, and shadow-utils to give us `groupadd` and `useradd`.
# Install which to allow running of securityadmin.sh
RUN yum update -y && yum install -y tar gzip shadow-utils which && yum clean all

# Create an opensearch user, group, and directory
RUN groupadd -g $GID opensearch && \
    adduser -u $UID -g $GID -d $OPENSEARCH_HOME opensearch && \
    mkdir $TEMP_DIR

RUN mkdir /usr/share/elasticsearch
WORKDIR /usr/share/elasticsearch

RUN set -eux ; \
    cur_arch="" ; \
    case "$(arch)" in \
        aarch64) cur_arch='arm64' ;; \
        x86_64)  cur_arch='x64' ;; \
        *) echo >&2 ; echo >&2 "Unsupported architecture $(arch)" ; echo >&2 ; exit 1 ;; \
    esac ; \
    curl --retry 10 -S -L --output $TEMP_DIR/opensearch.tar.gz https://artifacts.opensearch.org/releases/bundle/opensearch/$OS_VERSION/opensearch-$OS_VERSION-linux-$cur_arch.tar.gz; \
    curl --output $TEMP_DIR/opensearch.pgp https://artifacts.opensearch.org/publickeys/opensearch.pgp; \
    gpg --import $TEMP_DIR/opensearch.pgp; \
    curl --output $TEMP_DIR/opensearch.tar.gz.sig https://artifacts.opensearch.org/releases/bundle/opensearch/$OS_VERSION/opensearch-$OS_VERSION-linux-$cur_arch.tar.gz.sig; \
    gpg --verify $TEMP_DIR/opensearch.tar.gz.sig $TEMP_DIR/opensearch.tar.gz;

RUN tar --warning=no-timestamp -zxf $TEMP_DIR/opensearch.tar.gz -C $OPENSEARCH_HOME --strip-components=1 && \
    mkdir -p $OPENSEARCH_HOME/data && chown -Rv $UID:$GID $OPENSEARCH_HOME/data && \
    if [[ -d $SECURITY_PLUGIN_DIR ]] ; then chmod -v 750 $SECURITY_PLUGIN_DIR/tools/* ; fi && \
    rm -rf $TEMP_DIR

COPY config/* $OPENSEARCH_PATH_CONF/
COPY bin/* $OPENSEARCH_HOME/
RUN if [[ -d $PERFORMANCE_ANALYZER_PLUGIN_CONFIG_DIR ]] ; then mv $OPENSEARCH_PATH_CONF/performance-analyzer.properties $PERFORMANCE_ANALYZER_PLUGIN_CONFIG_DIR/ ; fi
########################### Stage 1 ########################
# Copy working directory to the actual release docker images
FROM amazonlinux:2

ARG UID=1000
ARG GID=1000
ARG OPENSEARCH_HOME=/usr/share/opensearch
ARG OS_VERSION=2.5.0

RUN yum update -y && yum install -y tar gzip shadow-utils which && yum clean all

# Create an opensearch user, group
RUN groupadd -g $GID opensearch && \
    adduser -u $UID -g $GID -d $OPENSEARCH_HOME opensearch

# Copy from Stage0
COPY --from=linux_stage_0 --chown=$UID:$GID $OPENSEARCH_HOME $OPENSEARCH_HOME
WORKDIR $OPENSEARCH_HOME

# Set $JAVA_HOME
RUN echo "export JAVA_HOME=$OPENSEARCH_HOME/jdk" >> /etc/profile.d/java_home.sh && \
    echo "export PATH=\$PATH:\$JAVA_HOME/bin" >> /etc/profile.d/java_home.sh
ENV JAVA_HOME=$OPENSEARCH_HOME/jdk
ENV PATH=$PATH:$JAVA_HOME/bin:$OPENSEARCH_HOME/bin

# Add k-NN lib directory to library loading path variable
ENV LD_LIBRARY_PATH="$LD_LIBRARY_PATH:$OPENSEARCH_HOME/plugins/opensearch-knn/lib"

# Change user
USER $UID


# Setup OpenSearch
# Disable security demo installation during image build, and allow user to disable during startup of the container
# Enable security plugin during image build, and allow user to disable during startup of the container
ARG DISABLE_INSTALL_DEMO_CONFIG=true
ARG DISABLE_SECURITY_PLUGIN=false
RUN ./opensearch-onetime-setup.sh

EXPOSE 9200 9300 9600 9650

# Label
LABEL org.label-schema.schema-version="1.0" \
  org.label-schema.name="opensearch" \
  org.label-schema.version="$OS_VERSION" \
  org.label-schema.url="https://opensearch.org" \
  org.label-schema.vcs-url="https://github.com/OpenSearch" \
  org.label-schema.license="Apache-2.0" \
  org.label-schema.vendor="OpenSearch"

# CMD to run
 ENTRYPOINT ["./opensearch-docker-entrypoint.sh"]
 CMD ["opensearch"]

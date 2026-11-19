# Debian 13 (trixie) ships no JDK 11/17 (only 21/25), so we keep the same Temurin 17
# JRE we run today and layer it onto a Debian 13 slim base instead of the Ubuntu-26.04
# eclipse-temurin image. liblcms2-2 / libfontconfig1 are the runtime natives the JRE
# needs for font/color handling; ca-certificates for TLS; curl for the healthcheck.
FROM eclipse-temurin:17-jre AS jdk

FROM debian:trixie-slim
# ca-certificates/curl for TLS + healthcheck; liblcms2-2 + libfontconfig1 are the
# JRE runtime natives; fonts-dejavu-core/-mono + fontconfig are what the server-side
# PDF generators (openpdf / flying-saucer: network-scan + compliance reports) render
# with — the Ubuntu base we replaced shipped exactly these.
RUN apt-get update -qq && \
    apt-get install -y --no-install-recommends \
        ca-certificates curl liblcms2-2 libfontconfig1 fontconfig fonts-dejavu-core fonts-dejavu-mono && \
    fc-cache -f >/dev/null 2>&1 && \
    rm -rf /var/lib/apt/lists/*
COPY --from=jdk /opt/java/openjdk /opt/java
ENV JAVA_HOME=/opt/java
ENV PATH=$JAVA_HOME/bin:$PATH

ADD target/utmstack.war ./
COPY filters /utmstack/filters
COPY rules /utmstack/rules

HEALTHCHECK --start-period=60s --interval=60s --timeout=60s --retries=3 \
  CMD curl -f http://localhost:8080/api/healthcheck || exit 1

EXPOSE 8080
ENTRYPOINT ["java", "-jar", "utmstack.war"]

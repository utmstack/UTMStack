# Base migration: eclipse-temurin:17-alpine (Alpine 3.24) -> Debian 13 (trixie) slim.
# Debian 13 ships no JDK 17 in its repos (only 21/25), so the SAME Temurin 17 JRE we
# run today is layered onto a trixie-slim base. Java version is unchanged (17).
# liblcms2-2 / libfontconfig1 / fonts-dejavu* are the JRE runtime natives + fonts the
# app needs (matching the fonts present in the previous base).
FROM eclipse-temurin:17-jre AS jdk

FROM debian:trixie-slim
RUN apt-get update -qq && \
    apt-get install -y --no-install-recommends \
        ca-certificates curl liblcms2-2 libfontconfig1 fontconfig fonts-dejavu-core fonts-dejavu-mono && \
    fc-cache -f >/dev/null 2>&1 && \
    rm -rf /var/lib/apt/lists/*
COPY --from=jdk /opt/java/openjdk /opt/java
ENV JAVA_HOME=/opt/java
ENV PATH=$JAVA_HOME/bin:$PATH

ADD target/user-auditor-1.0.0.jar ./

EXPOSE 8080

ENTRYPOINT ["java","-jar","user-auditor-1.0.0.jar"]

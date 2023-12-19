FROM eclipse-temurin:11

COPY ./utmstack.war /home/utmstack/
WORKDIR /home/utmstack/

HEALTHCHECK --start-period=60s --interval=60s --timeout=60s --retries=3 \
  CMD curl -f http://localhost:8080/api/healthcheck || exit 1

EXPOSE 8080
ENTRYPOINT ["java", "-jar", "./utmstack.war"]

FROM ubuntu:24.04

COPY backend/backend /app/backend
COPY backend/migrations /app/migrations

RUN apt-get update && \
    apt-get install -y ca-certificates curl && \
    update-ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

COPY definitions/filters /utmstack/filters
COPY definitions/rules /utmstack/rules
COPY definitions/compliance /utmstack/compliance

HEALTHCHECK --start-period=60s --interval=60s --timeout=60s --retries=3 \
  CMD curl -f http://localhost:8080/api/healthcheck || exit 1

EXPOSE 8080
CMD ["/app/backend"]

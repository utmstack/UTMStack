FROM golang:1.23-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/backend .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    adduser -D -u 1001 appuser

WORKDIR /app

COPY --from=build /out/backend /app/backend
COPY migrations /app/migrations

RUN chown -R appuser:appuser /app
USER appuser

EXPOSE 8080
CMD ["/app/backend"]

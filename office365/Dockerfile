FROM debian:bookworm-slim

RUN apt-get update
RUN apt-get install -y ca-certificates
RUN update-ca-certificates

COPY office365 .

RUN chmod +x office365
ENTRYPOINT ./office365
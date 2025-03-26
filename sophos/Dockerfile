FROM ubuntu:24.04

RUN apt-get update
RUN apt-get install -y ca-certificates
RUN update-ca-certificates

COPY sophos .

RUN chmod +x sophos
ENTRYPOINT ./sophos


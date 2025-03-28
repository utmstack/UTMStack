FROM ubuntu:24.04

RUN apt-get update
RUN apt-get install -y ca-certificates
RUN update-ca-certificates

COPY aws .

RUN chmod +x aws
ENTRYPOINT ./aws

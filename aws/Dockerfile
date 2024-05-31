FROM ubuntu:latest

RUN apt update

COPY aws .

RUN chmod +x aws
ENTRYPOINT ./aws


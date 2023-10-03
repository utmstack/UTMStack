FROM ubuntu:latest

RUN apt update

COPY o365_integration .

RUN chmod +x o365_integration
ENTRYPOINT ./o365_integration
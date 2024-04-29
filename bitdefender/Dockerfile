FROM ubuntu:latest

RUN apt update

COPY bitdefender .

RUN chmod +x bitdefender
ENTRYPOINT ./bitdefender


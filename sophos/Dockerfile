FROM ubuntu:latest

RUN apt update

COPY sophos .

RUN chmod +x sophos
ENTRYPOINT ./sophos


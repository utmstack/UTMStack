FROM ubuntu:latest

RUN apt update

COPY bdgz_integration .

RUN chmod +x bdgz_integration
ENTRYPOINT ./bdgz_integration


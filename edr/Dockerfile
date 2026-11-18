FROM ubuntu:24.04

RUN apt-get update && \
    apt-get install -y ca-certificates python3 python3-pip && \
    pip3 install --break-system-packages cvdupdate && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

COPY edr /app/edr

EXPOSE 9002

CMD ["/app/edr"]

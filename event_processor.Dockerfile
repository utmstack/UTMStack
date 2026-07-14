ARG BASE_IMAGE

FROM ${BASE_IMAGE}

COPY ./geolocation/ /workdir/geolocation/
COPY ./plugin-binaries/ /workdir/plugins/utmstack/
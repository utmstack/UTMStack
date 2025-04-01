ARG BASE_IMAGE

FROM ${BASE_IMAGE}

COPY ./pipelines/ /build-data/pipeline/
COPY ./geolocation/ /build-data/geolocation/
COPY ./plugins/alerts/com.utmstack.alerts.plugin /build-data/plugins/
COPY ./plugins/aws/com.utmstack.aws.plugin /build-data/plugins/
COPY ./plugins/azure/com.utmstack.azure.plugin /build-data/plugins/
COPY ./plugins/bitdefender/com.utmstack.bitdefender.plugin /build-data/plugins/
COPY ./plugins/config/com.utmstack.config.plugin /build-data/plugins/
COPY ./plugins/events/com.utmstack.events.plugin /build-data/plugins/
COPY ./plugins/gcp/com.utmstack.gcp.plugin /build-data/plugins/
COPY ./plugins/geolocation/com.utmstack.geolocation.plugin /build-data/plugins/
COPY ./plugins/inputs/com.utmstack.inputs.plugin /build-data/plugins/
COPY ./plugins/o365/com.utmstack.o365.plugin /build-data/plugins/
COPY ./plugins/sophos/com.utmstack.sophos.plugin /build-data/plugins/
COPY ./plugins/stats/com.utmstack.stats.plugin /build-data/plugins/

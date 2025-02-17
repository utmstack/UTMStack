ARG BASE_IMAGE

FROM ${BASE_IMAGE}

COPY ./pipelines/ /workdir/pipeline/
COPY ./geolocation/ /workdir/geolocation/
COPY ./plugins/alerts/com.utmstack.alerts.plugin /workdir/plugins/utmstack/
COPY ./plugins/aws/com.utmstack.aws.plugin /workdir/plugins/utmstack/
COPY ./plugins/azure/com.utmstack.azure.plugin /workdir/plugins/utmstack/
COPY ./plugins/bitdefender/com.utmstack.bitdefender.plugin /workdir/plugins/utmstack/
COPY ./plugins/config/com.utmstack.config.plugin /workdir/plugins/utmstack/
COPY ./plugins/events/com.utmstack.events.plugin /workdir/plugins/utmstack/
COPY ./plugins/gcp/com.utmstack.gcp.plugin /workdir/plugins/utmstack/
COPY ./plugins/geolocation/com.utmstack.geolocation.plugin /workdir/plugins/utmstack/
COPY ./plugins/inputs/com.utmstack.inputs.plugin /workdir/plugins/utmstack/
COPY ./plugins/o365/com.utmstack.o365.plugin /workdir/plugins/utmstack/
COPY ./plugins/sophos/com.utmstack.sophos.plugin /workdir/plugins/utmstack/
COPY ./plugins/stats/com.utmstack.stats.plugin /workdir/plugins/utmstack/

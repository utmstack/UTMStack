package com.park.utmstack.config;

import org.springframework.boot.context.properties.ConfigurationProperties;

/**
 * Properties specific to Utmstack.
 * <p>
 * Properties are configured in the application.yml file. See {@link tech.jhipster.config.JHipsterProperties} for a good
 * example.
 */
@ConfigurationProperties(prefix = "application", ignoreUnknownFields = false)
public class ApplicationProperties {
    private final ApplicationProperties.ChartBuilder chartBuilder = new ApplicationProperties.ChartBuilder();
    private final ApplicationProperties.IncidentResponse incidentResponse = new ApplicationProperties.IncidentResponse();

    public ChartBuilder getChartBuilder() {
        return chartBuilder;
    }

    public IncidentResponse getIncidentResponse() {
        return incidentResponse;
    }

    public static class ChartBuilder {
        private String ipInfoIndexName;

        public ChartBuilder() {
        }

        public String getIpInfoIndexName() {
            return ipInfoIndexName;
        }

        public void setIpInfoIndexName(String ipInfoIndexName) {
            this.ipInfoIndexName = ipInfoIndexName;
        }
    }

    public static class IncidentResponse {
        private Long assetVerificationInterval;

        public Long getAssetVerificationInterval() {
            return assetVerificationInterval;
        }

        public void setAssetVerificationInterval(Long assetVerificationInterval) {
            this.assetVerificationInterval = assetVerificationInterval;
        }
    }
}

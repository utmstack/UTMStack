package com.park.utmstack.service.logstash_pipeline.response.pipeline;

import com.fasterxml.jackson.annotation.JsonGetter;
import com.fasterxml.jackson.annotation.JsonSetter;

public class PipelineReloads {
    String lastFailureTimestamp;
    Long successes;
    Long failures;
    PipelineLastError lastError;
    String lastSuccessTimestamp;

    public PipelineReloads(){}

    public PipelineReloads(String lastFailureTimestamp,
                           Long successes,
                           Long failures,
                           PipelineLastError lastError,
                           String lastSuccessTimestamp) {
        this.lastFailureTimestamp = lastFailureTimestamp;
        this.successes = successes;
        this.failures = failures;
        this.lastError = lastError;
        this.lastSuccessTimestamp = lastSuccessTimestamp;
    }

    @JsonGetter("lastFailureTimestamp")
    public String getLastFailureTimestamp() {
        return lastFailureTimestamp;
    }

    @JsonSetter("last_failure_timestamp")
    public void setLastFailureTimestamp(String lastFailureTimestamp) {
        this.lastFailureTimestamp = lastFailureTimestamp;
    }

    public Long getSuccesses() {
        return successes;
    }

    public void setSuccesses(Long successes) {
        this.successes = successes;
    }

    public Long getFailures() {
        return failures;
    }

    public void setFailures(Long failures) {
        this.failures = failures;
    }

    @JsonGetter("lastError")
    public PipelineLastError getLastError() {
        return lastError;
    }

    @JsonSetter("last_error")
    public void setLastError(PipelineLastError lastError) {
        this.lastError = lastError;
    }

    @JsonGetter("lastSuccessTimestamp")
    public String getLastSuccessTimestamp() {
        return lastSuccessTimestamp;
    }

    @JsonSetter("last_success_timestamp")
    public void setLastSuccessTimestamp(String lastSuccessTimestamp) {
        this.lastSuccessTimestamp = lastSuccessTimestamp;
    }
}

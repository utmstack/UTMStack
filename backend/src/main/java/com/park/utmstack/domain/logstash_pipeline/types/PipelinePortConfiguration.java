package com.park.utmstack.domain.logstash_pipeline.types;

public class PipelinePortConfiguration {
    String inputId;
    String port;

    public PipelinePortConfiguration(){}
    public PipelinePortConfiguration(String inputId, String port) {
        this.inputId = inputId;
        this.port = port;
    }

    public String getInputId() {
        return inputId;
    }

    public void setInputId(String inputId) {
        this.inputId = inputId;
    }

    public String getPort() {
        return port;
    }

    public void setPort(String port) {
        this.port = port;
    }
}

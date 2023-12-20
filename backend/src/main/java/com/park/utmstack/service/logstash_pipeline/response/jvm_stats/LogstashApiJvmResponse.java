package com.park.utmstack.service.logstash_pipeline.response.jvm_stats;

import com.fasterxml.jackson.annotation.JsonGetter;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonSetter;
import com.park.utmstack.service.logstash_pipeline.response.pipeline.PipelineConfig;

public class LogstashApiJvmResponse {
    String host;
    String version;
    @JsonIgnore
    String httpAddress;
    @JsonIgnore
    String id;
    @JsonIgnore
    String name;
    @JsonIgnore
    String ephemeralId;
    String status;
    @JsonIgnore
    Boolean snapshot;
    PipelineConfig pipeline;
    JvmConfig jvm;

    public LogstashApiJvmResponse(){
        this.host = "Unknown";
        this.version = "Unknown";
        this.status = "down";
        this.pipeline = new PipelineConfig();
        this.jvm = new JvmConfig();
    }
    public LogstashApiJvmResponse(String host, String version, String httpAddress,
                                  String id, String name, String ephemeralId,
                                  String status, Boolean snapshot, PipelineConfig pipeline,
                                  JvmConfig jvm) {
        this.host = host;
        this.version = version;
        this.httpAddress = httpAddress;
        this.id = id;
        this.name = name;
        this.ephemeralId = ephemeralId;
        this.status = status;
        this.snapshot = snapshot;
        this.pipeline = pipeline;
        this.jvm = jvm;
    }

    public String getHost() {
        return host;
    }

    public void setHost(String host) {
        this.host = host;
    }

    public String getVersion() {
        return version;
    }

    public void setVersion(String version) {
        this.version = version;
    }

    @JsonIgnore
    @JsonGetter("httpAddress")
    public String getHttpAddress() {
        return httpAddress;
    }

    @JsonSetter("http_address")
    public void setHttpAddress(String httpAddress) {
        this.httpAddress = httpAddress;
    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    @JsonIgnore
    @JsonGetter("ephemeralId")
    public String getEphemeralId() {
        return ephemeralId;
    }

    @JsonSetter("ephemeral_id")
    public void setEphemeralId(String ephemeralId) {
        this.ephemeralId = ephemeralId;
    }

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }

    public Boolean getSnapshot() {
        return snapshot;
    }

    public void setSnapshot(Boolean snapshot) {
        this.snapshot = snapshot;
    }

    public PipelineConfig getPipeline() {
        return pipeline;
    }

    public void setPipeline(PipelineConfig pipeline) {
        this.pipeline = pipeline;
    }

    public JvmConfig getJvm() {
        return jvm;
    }

    public void setJvm(JvmConfig jvm) {
        this.jvm = jvm;
    }
}

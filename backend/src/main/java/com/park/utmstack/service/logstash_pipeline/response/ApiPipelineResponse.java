package com.park.utmstack.service.logstash_pipeline.response;

import com.park.utmstack.service.logstash_pipeline.response.pipeline.PipelineData;

import java.util.LinkedHashMap;
import java.util.Map;

public class ApiPipelineResponse {
    String host;
    String version;
    Map<String, PipelineData> pipelines;

    public ApiPipelineResponse(){
        this.host = "Unknown";
        this.version = "Unknown";
        this.pipelines = new LinkedHashMap<>();
    }

    public ApiPipelineResponse(String host, String version, Map<String, PipelineData> pipelines) {
        this.host = host;
        this.version = version;
        this.pipelines = pipelines;
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

    public Map<String, PipelineData> getPipelines() {
        return pipelines;
    }

    public void setPipelines(Map<String, PipelineData> pipelines) {
        this.pipelines = pipelines;
    }
}

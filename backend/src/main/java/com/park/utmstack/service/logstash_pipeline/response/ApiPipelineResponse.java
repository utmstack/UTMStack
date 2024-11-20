package com.park.utmstack.service.logstash_pipeline.response;

import com.park.utmstack.service.logstash_pipeline.response.pipeline.PipelineData;
import lombok.Getter;
import lombok.Setter;

import java.util.LinkedHashMap;
import java.util.Map;

@Setter
@Getter
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

}

package com.park.utmstack.service.logstash_pipeline.response;

import com.park.utmstack.service.logstash_pipeline.response.engine.ApiEngineResponse;
import com.park.utmstack.service.logstash_pipeline.response.pipeline.PipelineStats;

import java.util.List;

public class ApiStatsResponse {
    ApiEngineResponse general;
    List<PipelineStats> pipelines;

    public ApiStatsResponse(){}
    public ApiStatsResponse(ApiEngineResponse general, List<PipelineStats> pipelines) {
        this.general = general;
        this.pipelines = pipelines;
    }

    public ApiEngineResponse getGeneral() {
        return general;
    }

    public void setGeneral(ApiEngineResponse general) {
        this.general = general;
    }

    public List<PipelineStats> getPipelines() {
        return pipelines;
    }

    public void setPipelines(List<PipelineStats> pipelines) {
        this.pipelines = pipelines;
    }
}

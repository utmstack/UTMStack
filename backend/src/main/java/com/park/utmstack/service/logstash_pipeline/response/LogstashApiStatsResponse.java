package com.park.utmstack.service.logstash_pipeline.response;

import com.park.utmstack.service.logstash_pipeline.response.jvm_stats.LogstashApiJvmResponse;
import com.park.utmstack.service.logstash_pipeline.response.pipeline.PipelineStats;

import java.util.List;

public class LogstashApiStatsResponse {
    LogstashApiJvmResponse general;
    List<PipelineStats> pipelines;

    public LogstashApiStatsResponse(){}
    public LogstashApiStatsResponse(LogstashApiJvmResponse general, List<PipelineStats> pipelines) {
        this.general = general;
        this.pipelines = pipelines;
    }

    public LogstashApiJvmResponse getGeneral() {
        return general;
    }

    public void setGeneral(LogstashApiJvmResponse general) {
        this.general = general;
    }

    public List<PipelineStats> getPipelines() {
        return pipelines;
    }

    public void setPipelines(List<PipelineStats> pipelines) {
        this.pipelines = pipelines;
    }
}

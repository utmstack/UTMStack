package com.park.utmstack.service.logstash_pipeline.response.pipeline;

public class PipelineEvents {
    Long out;

    public PipelineEvents(){}

    public PipelineEvents(Long out) {
        this.out = out;
    }

    public Long getOut() {
        return out;
    }

    public void setOut(Long out) {
        this.out = out;
    }
}

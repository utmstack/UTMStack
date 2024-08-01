package com.park.utmstack.service.logstash_pipeline.response.pipeline;

public class PipelineEvents {
    Long in;
    Long filtered;
    Long out;

    public PipelineEvents(){}

    public PipelineEvents(Long in,
                          Long filtered,
                          Long out) {
        this.in = in;
        this.filtered = filtered;
        this.out = out;
    }

    public Long getIn() {
        return in;
    }

    public void setIn(Long in) {
        this.in = in;
    }

    public Long getFiltered() {
        return filtered;
    }

    public void setFiltered(Long filtered) {
        this.filtered = filtered;
    }

    public Long getOut() {
        return out;
    }

    public void setOut(Long out) {
        this.out = out;
    }
}

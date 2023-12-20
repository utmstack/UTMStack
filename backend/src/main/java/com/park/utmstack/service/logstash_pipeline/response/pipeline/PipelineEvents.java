package com.park.utmstack.service.logstash_pipeline.response.pipeline;

import com.fasterxml.jackson.annotation.JsonGetter;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonSetter;

public class PipelineEvents {
    @JsonIgnore
    Long queuePushDurationInMillis;
    @JsonIgnore
    Long durationInMillis;
    Long in;
    Long filtered;
    Long out;

    public PipelineEvents(){}

    public PipelineEvents(Long queuePushDurationInMillis,
                          Long durationInMillis,
                          Long in,
                          Long filtered,
                          Long out) {
        this.queuePushDurationInMillis = queuePushDurationInMillis;
        this.durationInMillis = durationInMillis;
        this.in = in;
        this.filtered = filtered;
        this.out = out;
    }

    @JsonIgnore
    @JsonGetter("queuePushDurationInMillis")
    public Long getQueuePushDurationInMillis() {
        return queuePushDurationInMillis;
    }

    @JsonSetter("queue_push_duration_in_millis")
    public void setQueuePushDurationInMillis(Long queuePushDurationInMillis) {
        this.queuePushDurationInMillis = queuePushDurationInMillis;
    }

    @JsonIgnore
    @JsonGetter("durationInMillis")
    public Long getDurationInMillis() {
        return durationInMillis;
    }

    @JsonSetter("duration_in_millis")
    public void setDurationInMillis(Long durationInMillis) {
        this.durationInMillis = durationInMillis;
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

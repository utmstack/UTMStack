package com.park.utmstack.service.logstash_pipeline.response.pipeline;


public class PipelineData {
    PipelineEvents events;
    Long errors = 0L;

    public PipelineData(){}

    public PipelineData(PipelineEvents events, Long errors) {
        this.events = events;
        this.errors = errors;
    }

    public PipelineEvents getEvents() {
        return events;
    }

    public void setEvents(PipelineEvents events) {
        this.events = events;
    }

    public Long getErrors() {
        return errors;
    }

    public void setErrors(Long errors) {
        this.errors = errors;
    }
}

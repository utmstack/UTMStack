package com.park.utmstack.service.logstash_pipeline.response.pipeline;

import com.park.utmstack.domain.logstash_pipeline.UtmLogstashPipeline;

public class PipelineStats {
    Long id;
    String pipelineId;
    String pipelineName;
    String pipelineStatus;
    String moduleName;
    Boolean systemOwner;
    String pipelineDescription;
    PipelineEvents events;
    Long errors;

    public PipelineStats() {
    }

    public PipelineStats(UtmLogstashPipeline info, PipelineEvents events, Long errors) {
        this.id = info.getId();
        this.pipelineId = info.getPipelineId();
        this.pipelineName = info.getPipelineName();
        this.pipelineStatus = info.getPipelineStatus();
        this.moduleName = info.getModuleName();
        this.systemOwner = info.getSystemOwner();
        this.events = events;
        this.errors = errors;
    }
    public static PipelineStats getPipelineStats(UtmLogstashPipeline info) {
        PipelineStats stats = new PipelineStats();
        stats.setId(info.getId());
        stats.setPipelineId(info.getPipelineId());
        stats.setPipelineName(info.getPipelineName());
        stats.setPipelineStatus(info.getPipelineStatus());
        stats.setModuleName(info.getModuleName());
        stats.setSystemOwner(info.getSystemOwner());
        stats.setPipelineDescription(info.getPipelineDescription());
        PipelineEvents events = new PipelineEvents();
        events.setOut(info.getEventsOut());
        stats.setEvents(events);
        stats.setErrors(0L);
        return stats;
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getPipelineId() {
        return pipelineId;
    }

    public void setPipelineId(String pipelineId) {
        this.pipelineId = pipelineId;
    }

    public String getPipelineName() {
        return pipelineName;
    }

    public void setPipelineName(String pipelineName) {
        this.pipelineName = pipelineName;
    }

    public String getPipelineStatus() {
        return pipelineStatus;
    }

    public void setPipelineStatus(String pipelineStatus) {
        this.pipelineStatus = pipelineStatus;
    }

    public String getModuleName() {
        return moduleName;
    }

    public void setModuleName(String moduleName) {
        this.moduleName = moduleName;
    }

    public Boolean getSystemOwner() {
        return systemOwner;
    }

    public void setSystemOwner(Boolean systemOwner) {
        this.systemOwner = systemOwner;
    }

    public String getPipelineDescription() {
        return pipelineDescription;
    }

    public void setPipelineDescription(String pipelineDescription) {
        this.pipelineDescription = pipelineDescription;
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



package com.park.utmstack.service.logstash_pipeline.response.pipeline;

import com.fasterxml.jackson.annotation.JsonGetter;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonSetter;

public class PipelineData {
    PipelineEvents events;
    @JsonIgnore
    Object plugins;
    PipelineReloads reloads;
    @JsonIgnore
    Object queue;
    @JsonIgnore
    String hash;
    @JsonIgnore
    String ephemeralId;

    public PipelineData(){}

    public PipelineData(PipelineEvents events,
                        Object plugins, PipelineReloads reloads,
                        Object queue, String hash, String ephemeralId) {
        this.events = events;
        this.plugins = plugins;
        this.reloads = reloads;
        this.queue = queue;
        this.hash = hash;
        this.ephemeralId = ephemeralId;
    }

    public PipelineEvents getEvents() {
        return events;
    }

    public void setEvents(PipelineEvents events) {
        this.events = events;
    }

    public Object getPlugins() {
        return plugins;
    }

    public void setPlugins(Object plugins) {
        this.plugins = plugins;
    }

    public PipelineReloads getReloads() {
        return reloads;
    }

    public void setReloads(PipelineReloads reloads) {
        this.reloads = reloads;
    }

    public Object getQueue() {
        return queue;
    }

    public void setQueue(Object queue) {
        this.queue = queue;
    }

    public String getHash() {
        return hash;
    }

    public void setHash(String hash) {
        this.hash = hash;
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
}

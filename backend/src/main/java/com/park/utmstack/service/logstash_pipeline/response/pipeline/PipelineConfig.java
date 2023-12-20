package com.park.utmstack.service.logstash_pipeline.response.pipeline;

import com.fasterxml.jackson.annotation.JsonGetter;
import com.fasterxml.jackson.annotation.JsonSetter;

public class PipelineConfig {
    Integer workers = 0;
    Integer batchSize = 0;
    Integer batchDelay = 0;

    public PipelineConfig(){}

    public PipelineConfig(Integer workers, Integer batchSize, Integer batchDelay) {
        this.workers = workers;
        this.batchSize = batchSize;
        this.batchDelay = batchDelay;
    }

    public Integer getWorkers() {
        return workers;
    }

    public void setWorkers(Integer workers) {
        this.workers = workers;
    }

    @JsonGetter("batchSize")
    public Integer getBatchSize() {
        return batchSize;
    }

    @JsonSetter("batch_size")
    public void setBatchSize(Integer batchSize) {
        this.batchSize = batchSize;
    }

    @JsonGetter("batchDelay")
    public Integer getBatchDelay() {
        return batchDelay;
    }

    @JsonSetter("batch_delay")
    public void setBatchDelay(Integer batchDelay) {
        this.batchDelay = batchDelay;
    }
}

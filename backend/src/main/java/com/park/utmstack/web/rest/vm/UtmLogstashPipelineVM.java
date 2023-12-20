package com.park.utmstack.web.rest.vm;

import com.fasterxml.jackson.annotation.JsonGetter;
import com.fasterxml.jackson.annotation.JsonSetter;
import com.park.utmstack.domain.logstash_pipeline.UtmGroupLogstashPipelineFilters;
import com.park.utmstack.domain.logstash_pipeline.types.InputConfiguration;
import com.park.utmstack.service.dto.logstash_pipeline.UtmLogstashPipelineDTO;

import java.util.List;

public class UtmLogstashPipelineVM {
    UtmLogstashPipelineDTO pipeline;
    List<UtmGroupLogstashPipelineFilters> filters;
    List<InputConfiguration> inputs;

    public UtmLogstashPipelineVM() {
    }

    public UtmLogstashPipelineVM(UtmLogstashPipelineDTO pipeline, List<UtmGroupLogstashPipelineFilters> filters, List<InputConfiguration> inputs) {
        this.pipeline = pipeline;
        this.filters = filters;
        this.inputs = inputs;
    }

    @JsonGetter("pipeline")
    public UtmLogstashPipelineDTO getPipelineDTO() {
        return pipeline;
    }

    @JsonSetter("pipeline")
    public void setPipelineDTO(UtmLogstashPipelineDTO pipeline) {
        this.pipeline = pipeline;
    }

    public List<UtmGroupLogstashPipelineFilters> getFilters() {
        return filters;
    }

    public void setFilters(List<UtmGroupLogstashPipelineFilters> filters) {
        this.filters = filters;
    }

    public List<InputConfiguration> getInputs() {
        return inputs;
    }

    public void setInputs(List<InputConfiguration> inputs) {
        this.inputs = inputs;
    }
}

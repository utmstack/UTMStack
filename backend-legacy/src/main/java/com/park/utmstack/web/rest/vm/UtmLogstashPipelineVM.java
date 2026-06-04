package com.park.utmstack.web.rest.vm;

import com.fasterxml.jackson.annotation.JsonGetter;
import com.fasterxml.jackson.annotation.JsonSetter;
import com.park.utmstack.domain.logstash_pipeline.UtmGroupLogstashPipelineFilters;
import com.park.utmstack.service.dto.logstash_pipeline.UtmLogstashPipelineDTO;

import java.util.List;

public class UtmLogstashPipelineVM {
    UtmLogstashPipelineDTO pipeline;
    List<UtmGroupLogstashPipelineFilters> filters;

    public UtmLogstashPipelineVM() {
    }

    public UtmLogstashPipelineVM(UtmLogstashPipelineDTO pipeline, List<UtmGroupLogstashPipelineFilters> filters) {
        this.pipeline = pipeline;
        this.filters = filters;
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
}

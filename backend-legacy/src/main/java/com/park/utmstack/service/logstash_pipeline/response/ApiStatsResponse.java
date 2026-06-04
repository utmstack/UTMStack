package com.park.utmstack.service.logstash_pipeline.response;

import com.park.utmstack.service.logstash_pipeline.response.engine.ApiEngineResponse;
import com.park.utmstack.service.logstash_pipeline.response.pipeline.PipelineStats;
import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.Setter;

import java.util.List;

@Setter
@Getter
@RequiredArgsConstructor
public class ApiStatsResponse {
    ApiEngineResponse general;
    List<PipelineStats> pipelines;
}

package com.park.utmstack.service.logstash_pipeline.response.pipeline;


import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.Setter;

@Setter
@Getter
@RequiredArgsConstructor
public class PipelineData {
    PipelineEvents events;
    Long errors = 0L;
}

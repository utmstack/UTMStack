package com.park.utmstack.service.logstash_pipeline.response.statistic;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Getter;
import lombok.Setter;

@Setter
@Getter
public class StatisticDocument {
    private String dataSource;
    private String dataType;
    private Long count;
    private String type;

    @JsonProperty("@timestamp")
    private String timestamp;
}

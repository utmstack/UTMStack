package com.park.utmstack.service.elasticsearch.processor;

import org.springframework.stereotype.Component;

@Component
public class SearchProcessorRegistry {
    public SearchResultProcessor resolve(String groupByField) {
        return (groupByField != null && !groupByField.isBlank())
                ? new GroupByFieldProcessor(groupByField)
                : new NoOpProcessor();
    }
}


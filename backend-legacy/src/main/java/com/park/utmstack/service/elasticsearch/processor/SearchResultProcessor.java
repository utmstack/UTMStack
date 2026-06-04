package com.park.utmstack.service.elasticsearch.processor;

import java.util.List;
import java.util.Map;

public interface SearchResultProcessor {
    List<Map<String, Object>> process(List<Map<String, Object>> rawResults);
}


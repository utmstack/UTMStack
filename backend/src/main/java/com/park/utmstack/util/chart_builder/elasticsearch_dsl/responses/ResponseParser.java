package com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses;

import com.fasterxml.jackson.databind.node.ObjectNode;
import com.park.utmstack.domain.chart_builder.UtmVisualization;
import org.opensearch.client.opensearch.core.SearchResponse;

import java.util.List;

public interface ResponseParser<T> {
    List<T> parse(UtmVisualization visualization, SearchResponse<ObjectNode> result);
}

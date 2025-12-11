package com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses;

import com.fasterxml.jackson.databind.node.ObjectNode;
import com.park.utmstack.domain.chart_builder.UtmVisualization;
import com.utmstack.opensearch_connector.types.SearchSqlResponse;
import org.opensearch.client.opensearch.core.SearchResponse;

import java.util.List;
import java.util.Map;

public interface ResponseParser<T> {
    List<T> parse(UtmVisualization visualization, SearchResponse<ObjectNode> result);

    default List<T> parse(UtmVisualization visualization, SearchSqlResponse<Map> result) {
        return null;
    }
}

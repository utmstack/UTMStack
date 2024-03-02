package com.utmstack.userauditor.model;

import org.opensearch.client.opensearch._types.SortOrder;
import org.opensearch.client.opensearch._types.aggregations.Aggregation;

import java.util.List;
import java.util.Map;

public class TermsAggregationBuilder {

    public static Aggregation buildTermsAggregation(String field) {
        return Aggregation.of(agg -> agg
                .terms(t -> t.field(field)));
    }
}

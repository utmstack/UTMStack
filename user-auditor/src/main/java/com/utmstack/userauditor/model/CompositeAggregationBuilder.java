package com.utmstack.userauditor.model;

import org.opensearch.client.opensearch._types.SortOrder;
import org.opensearch.client.opensearch._types.aggregations.Aggregation;
import org.opensearch.client.opensearch._types.aggregations.CompositeAggregationSource;

import java.util.List;
import java.util.Map;

public class CompositeAggregationBuilder {

    public static Aggregation buildCompositeTermAggregation(Aggregation termsAggregation, String name, int size, String after) {
        return Aggregation.of(agg -> agg
                .composite(c -> c.size(size)
                        .sources(List.of(Map.of(name,
                                CompositeAggregationSource.of(composite ->
                                        composite.terms(termsAggregation.terms())))))
                        .after("users", after))
                .aggregations("top_events", subAgg -> subAgg.topHits(th -> th
                        .size(1)
                        .sort(sort -> sort.field(f -> f.field("@timestamp").order(SortOrder.Desc))))
                )
        );
    }

}
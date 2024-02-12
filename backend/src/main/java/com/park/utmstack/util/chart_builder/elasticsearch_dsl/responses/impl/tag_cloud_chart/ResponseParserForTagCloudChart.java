package com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.tag_cloud_chart;

import com.fasterxml.jackson.databind.node.ObjectNode;
import com.park.utmstack.domain.chart_builder.UtmVisualization;
import com.park.utmstack.domain.chart_builder.types.aggregation.AggregationType;
import com.park.utmstack.domain.chart_builder.types.aggregation.Bucket;
import com.park.utmstack.domain.chart_builder.types.aggregation.Metric;
import com.park.utmstack.domain.chart_builder.types.aggregation.enums.BucketAggregationEnum;
import com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.ResponseParser;
import com.utmstack.opensearch_connector.parsers.TermAggregateParser;
import com.utmstack.opensearch_connector.types.BucketAggregation;
import org.opensearch.client.opensearch._types.aggregations.*;
import org.opensearch.client.opensearch.core.SearchResponse;
import org.springframework.util.Assert;
import org.springframework.util.CollectionUtils;

import java.util.*;

public class ResponseParserForTagCloudChart implements ResponseParser<TagCloudChartResult> {
    private static final String CLASSNAME = "ResponseParserForTagCloudChart";

    @Override
    public List<TagCloudChartResult> parse(UtmVisualization visualization, SearchResponse<ObjectNode> result) {
        final String ctx = CLASSNAME + ".parse";
        try {
            Assert.notNull(visualization, "Param visualization must not be null");
            AggregationType aggs = visualization.getAggregationType();
            List<Metric> metrics = aggs.getMetrics();

            if (CollectionUtils.isEmpty(metrics))
                throw new RuntimeException("A metric is needed");

            Bucket bucket = aggs.getBucket();

            if (bucket != null) {
                if (Objects.requireNonNull(bucket.getAggregation()) == BucketAggregationEnum.TERMS)
                    return parseTermAggregation(result.aggregations(), bucket, metrics);
            }
            return Collections.singletonList(new TagCloudChartResult("all", result.hits().total().value(), null, null));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    private List<TagCloudChartResult> parseTermAggregation(Map<String, Aggregate> aggregations, Bucket bucket,
                                                           List<Metric> metrics) {
        final String ctx = CLASSNAME + ".parseTermAggregation";
        try {
            List<BucketAggregation> _buckets = TermAggregateParser.parse(aggregations.get(bucket.toString()));
            List<TagCloudChartResult> _return = new ArrayList<>();
            for (Metric metric : metrics) {
                for (BucketAggregation _bucket : _buckets) {
                    switch (metric.getAggregation()) {
                        case AVERAGE:
                            AvgAggregate avg = _bucket.getSubAggregations().get(metric.getId()).avg();
                            _return.add(new TagCloudChartResult(_bucket.getKey(), avg.value(), metric.getId(), bucket.getId()));
                            break;
                        case COUNT:
                            _return.add(new TagCloudChartResult(_bucket.getKey(), _bucket.getDocCount(), metric.getId(), bucket.getId()));
                            break;
                        case MAX:
                            MaxAggregate max = _bucket.getSubAggregations().get(metric.getId()).max();
                            _return.add(new TagCloudChartResult(_bucket.getKey(), max.value(), metric.getId(), bucket.getId()));
                            break;
                        case MIN:
                            MinAggregate min = _bucket.getSubAggregations().get(metric.getId()).min();
                            _return.add(new TagCloudChartResult(_bucket.getKey(), min.value(), metric.getId(), bucket.getId()));
                            break;
                        case SUM:
                            SumAggregate sum = _bucket.getSubAggregations().get(metric.getId()).sum();
                            _return.add(new TagCloudChartResult(_bucket.getKey(), sum.value(), metric.getId(), bucket.getId()));
                            break;
                    }
                }
            }
            return _return;
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }
}

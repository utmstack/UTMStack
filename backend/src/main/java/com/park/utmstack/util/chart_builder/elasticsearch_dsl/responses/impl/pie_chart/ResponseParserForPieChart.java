package com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.pie_chart;

import com.fasterxml.jackson.databind.node.ObjectNode;
import com.park.utmstack.domain.chart_builder.UtmVisualization;
import com.park.utmstack.domain.chart_builder.types.aggregation.AggregationType;
import com.park.utmstack.domain.chart_builder.types.aggregation.Bucket;
import com.park.utmstack.domain.chart_builder.types.aggregation.Metric;
import com.park.utmstack.domain.chart_builder.types.aggregation.enums.BucketAggregationEnum;
import com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.ResponseParser;
import com.park.utmstack.util.exceptions.UtmChartBuilderException;
import com.utmstack.opensearch_connector.parsers.TermAggregateParser;
import com.utmstack.opensearch_connector.types.BucketAggregation;
import org.opensearch.client.opensearch._types.aggregations.*;
import org.opensearch.client.opensearch.core.SearchResponse;
import org.springframework.util.Assert;
import org.springframework.util.CollectionUtils;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Objects;

public class ResponseParserForPieChart implements ResponseParser<PieChartResult> {
    private static final String CLASSNAME = "ResponseParserForPieChart";

    @Override
    public List<PieChartResult> parse(UtmVisualization visualization, SearchResponse<ObjectNode> result) {
        final String ctx = CLASSNAME + ".parse";
        try {
            Assert.notNull(visualization, "Param visualization must not be null");
            AggregationType aggregationType = visualization.getAggregationType();
            Assert.notNull(aggregationType, "Param visualization.aggregationType must not be null");

            Bucket bucket = aggregationType.getBucket();

            // In pie chart only one metric is allowed
            if (CollectionUtils.isEmpty(aggregationType.getMetrics()) || aggregationType.getMetrics().size() > 1)
                throw new UtmChartBuilderException("In pie charts it is required one and only one metric type");

            Metric metric = aggregationType.getMetrics().get(0);

            if (bucket != null) {
                if (Objects.requireNonNull(bucket.getAggregation()) == BucketAggregationEnum.TERMS) {
                    return parseTermAggregation(result.aggregations(), bucket, metric);
                }
            } else {
                List<PieChartResult> _return = new ArrayList<>();
                Map<String, Aggregate> agg = result.aggregations();

                switch (metric.getAggregation()) {
                    case AVERAGE:
                        AvgAggregate avg = agg.get(metric.getId()).avg();
                        _return.add(new PieChartResult(metric.getId(), avg.value(), null, null));
                        break;
                    case COUNT:
                        _return.add(new PieChartResult(metric.getId(), result.hits().total().value(), null, null));
                        break;
                    case MAX:
                        MaxAggregate max = agg.get(metric.getId()).max();
                        _return.add(new PieChartResult(metric.getId(), max.value(), null, null));
                        break;
                    case MIN:
                        MinAggregate min = agg.get(metric.getId()).min();
                        _return.add(new PieChartResult(metric.getId(), min.value(), null, null));
                        break;
                    case SUM:
                        SumAggregate sum = agg.get(metric.getId()).sum();
                        _return.add(new PieChartResult(metric.getId(), sum.value(), null, null));
                        break;
                }
                return _return;
            }
            return new ArrayList<>();
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    private List<PieChartResult> parseTermAggregation(Map<String, Aggregate> result, Bucket bucket, Metric metric) {
        final String ctx = CLASSNAME + ".parseTermAggregation";
        try {
            List<BucketAggregation> _buckets = TermAggregateParser.parse(result.get(bucket.toString()));
            List<PieChartResult> _return = new ArrayList<>();

            for (BucketAggregation _bucket : _buckets) {
                if(_bucket.getKey().isEmpty()){
                    _bucket.setKey("UNKNOWN");
                }
                switch (metric.getAggregation()) {
                    case AVERAGE:
                        AvgAggregate avg = _bucket.getSubAggregations().get(metric.getId()).avg();
                        _return.add(new PieChartResult(metric.getId(), avg.value(), _bucket.getKey(), bucket.getId()));
                        break;
                    case COUNT:
                        _return.add(new PieChartResult(metric.getId(), _bucket.getDocCount(), _bucket.getKey(), bucket.getId()));
                        break;
                    case MAX:
                        MaxAggregate max = _bucket.getSubAggregations().get(metric.getId()).max();
                        _return.add(new PieChartResult(metric.getId(), max.value(), _bucket.getKey(), bucket.getId()));
                        break;
                    case MIN:
                        MinAggregate min = _bucket.getSubAggregations().get(metric.getId()).min();
                        _return.add(new PieChartResult(metric.getId(), min.value(), _bucket.getKey(), bucket.getId()));
                        break;
                    case SUM:
                        SumAggregate sum = _bucket.getSubAggregations().get(metric.getId()).sum();
                        _return.add(new PieChartResult(metric.getId(), sum.value(), _bucket.getKey(), bucket.getId()));
                        break;
                }
            }
            return _return;
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }
}

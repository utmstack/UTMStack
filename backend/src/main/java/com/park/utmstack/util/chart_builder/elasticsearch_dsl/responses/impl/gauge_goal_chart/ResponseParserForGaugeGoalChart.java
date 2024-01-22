package com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.gauge_goal_chart;

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

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Objects;

public class ResponseParserForGaugeGoalChart implements ResponseParser<GaugeGoalChartResult> {
    private static final String CLASSNAME = "ResponseParserForGaugeGoalChart";

    @Override
    public List<GaugeGoalChartResult> parse(UtmVisualization visualization, SearchResponse<ObjectNode> result) {
        final String ctx = CLASSNAME + ".parse";
        try {
            Assert.notNull(visualization, "Param visualization must not be null");
            AggregationType aggregationType = visualization.getAggregationType();
            Assert.notNull(aggregationType, "Param visualization.aggregationType must not be null");

            Bucket bucket = aggregationType.getBucket();
            List<Metric> metrics = aggregationType.getMetrics();

            if (bucket != null) {
                if (Objects.requireNonNull(bucket.getAggregation()) == BucketAggregationEnum.TERMS)
                    return parseTermAggregation(result.aggregations(), bucket, metrics);
            } else {
                GaugeGoalChartResult obj;
                List<GaugeGoalChartResult> rturn = new ArrayList<>();
                Map<String, Aggregate> ma = result.aggregations();
                for (Metric metric : metrics) {
                    switch (metric.getAggregation()) {
                        case AVERAGE:
                            AvgAggregate avg = ma.get(metric.getId()).avg();
                            obj = new GaugeGoalChartResult(metric.getId(), avg.value(), null, null);
                            rturn.add(obj);
                            break;
                        case COUNT:
                            obj = new GaugeGoalChartResult(metric.getId(), result.hits().total().value(), null, null);
                            rturn.add(obj);
                            break;
                        case MAX:
                            MaxAggregate max = ma.get(metric.getId()).max();
                            obj = new GaugeGoalChartResult(metric.getId(), max.value(), null, null);
                            rturn.add(obj);
                            break;
                        case MIN:
                            MinAggregate min = ma.get(metric.getId()).min();
                            obj = new GaugeGoalChartResult(metric.getId(), min.value(), null, null);
                            rturn.add(obj);
                            break;
                        case SUM:
                            SumAggregate sum = ma.get(metric.getId()).sum();
                            obj = new GaugeGoalChartResult(metric.getId(), sum.value(), null, null);
                            rturn.add(obj);
                            break;
                    }
                }
                return rturn;
            }
            return new ArrayList<>();
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    private List<GaugeGoalChartResult> parseTermAggregation(Map<String, Aggregate> result, Bucket bucket, List<Metric> metrics) {
        List<BucketAggregation> resultBuckets = TermAggregateParser.parse(result.get(bucket.toString()));
        List<GaugeGoalChartResult> rturn = new ArrayList<>();
        GaugeGoalChartResult obj;
        for (Metric metric : metrics) {
            for (BucketAggregation entry : resultBuckets) {
                switch (metric.getAggregation()) {
                    case AVERAGE:
                        AvgAggregate avg = entry.getSubAggregations().get(metric.getId()).avg();
                        obj = new GaugeGoalChartResult(metric.getId(), avg.value(), entry.getKey(), bucket.getId());
                        rturn.add(obj);
                        break;
                    case COUNT:
                        obj = new GaugeGoalChartResult(metric.getId(), entry.getDocCount(), entry.getKey(), bucket.getId());
                        rturn.add(obj);
                        break;
                    case MAX:
                        MaxAggregate max = entry.getSubAggregations().get(metric.getId()).max();
                        obj = new GaugeGoalChartResult(metric.getId(), max.value(), entry.getKey(), bucket.getId());
                        rturn.add(obj);
                        break;
                    case MIN:
                        MinAggregate min = entry.getSubAggregations().get(metric.getId()).min();
                        obj = new GaugeGoalChartResult(metric.getId(), min.value(), entry.getKey(), bucket.getId());
                        rturn.add(obj);
                        break;
                    case SUM:
                        SumAggregate sum = entry.getSubAggregations().get(metric.getId()).sum();
                        obj = new GaugeGoalChartResult(metric.getId(), sum.value(), entry.getKey(), bucket.getId());
                        rturn.add(obj);
                        break;
                }
            }
        }
        return rturn;
    }
}

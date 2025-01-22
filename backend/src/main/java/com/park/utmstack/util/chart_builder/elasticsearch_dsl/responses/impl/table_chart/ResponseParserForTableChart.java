package com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.table_chart;

import com.fasterxml.jackson.databind.node.ObjectNode;
import com.park.utmstack.domain.chart_builder.UtmVisualization;
import com.park.utmstack.domain.chart_builder.types.aggregation.AggregationType;
import com.park.utmstack.domain.chart_builder.types.aggregation.Bucket;
import com.park.utmstack.domain.chart_builder.types.aggregation.Metric;
import com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.ResponseParser;
import com.utmstack.opensearch_connector.parsers.DateHistogramAggregateParser;
import com.utmstack.opensearch_connector.parsers.TermAggregateParser;
import com.utmstack.opensearch_connector.types.BucketAggregation;
import org.opensearch.client.opensearch._types.aggregations.*;
import org.opensearch.client.opensearch.core.SearchResponse;
import org.springframework.util.Assert;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Map;

public class ResponseParserForTableChart implements ResponseParser<TableChartResult> {

    private static final String CLASSNAME = "ResponseParserForTableChart";
    private final TableChartResult table = new TableChartResult();
    private final List<TableChartResult.Cell<?>> rowCells = new ArrayList<>();

    @Override
    public List<TableChartResult> parse(UtmVisualization visualization, SearchResponse<ObjectNode> result) {
        final String ctx = CLASSNAME + ".parse";

        Assert.notNull(visualization, ctx + ": Param visualization must not be null");

        try {
            AggregationType aggs = visualization.getAggregationType();

            List<Metric> metrics = aggs.getMetrics();

            if (CollectionUtils.isEmpty(metrics))
                throw new RuntimeException("It's needed at least one metric aggregation to show");

            Bucket bucket = aggs.getBucket();

            if (bucket != null) {
                Bucket temp = bucket;
                while (temp != null) {
                    String col = !StringUtils.hasText(temp.getCustomLabel()) ? temp.getField() + "->" + temp
                        .getField() : temp.getField() + "->" + temp.getCustomLabel();
                    table.addColumn(col);
                    temp = temp.getSubBucket();
                }
            }

            metrics.forEach(m -> {
                String customLabel = !StringUtils.hasText(m.getCustomLabel()) ? m.getField() + "->" + m.getDefaultLabel() : m
                    .getField() + "->" + m.getCustomLabel();
                table.addColumn(StringUtils.hasText(customLabel) ? customLabel : m.getDefaultLabel());
            });

            if (bucket != null)
                parseBuckets(bucket, result.aggregations(), metrics, null);
            else
                parseMetric(result, metrics);

            return Collections.singletonList(table);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    private void parseBuckets(Bucket bucket, Map<String, Aggregate> aggregations, List<Metric> metrics, BucketAggregation parent) {
        final String ctx = CLASSNAME + ".parseBuckets";
        try {
            if (bucket != null) {
                List<BucketAggregation> _buckets = new ArrayList<>();
                switch (bucket.getAggregation()) {
                    case TERMS:
                        _buckets = TermAggregateParser.parse(aggregations.get(bucket.getId()));
                        break;
                    case DATE_HISTOGRAM:
                        _buckets = DateHistogramAggregateParser.parse(aggregations.get(bucket.getId()));
                        break;
                }
                if (CollectionUtils.isEmpty(_buckets))
                    return;
                for (BucketAggregation _bucket : _buckets) {
                    rowCells.add(new TableChartResult.Cell<>(!_bucket.getKey().isEmpty() ? _bucket.getKey() : "UNKNOWN", false));
                    parseBuckets(bucket.getSubBucket(), _bucket.getSubAggregations(), metrics, _bucket);
                }
            } else {
                for (Metric metric : metrics) {
                    switch (metric.getAggregation()) {
                        case MAX:
                            MaxAggregate max = aggregations.get(metric.getId()).max();
                            rowCells.add(new TableChartResult.Cell<>(max.value(), true));
                            break;
                        case MIN:
                            MinAggregate min = aggregations.get(metric.getId()).min();
                            rowCells.add(new TableChartResult.Cell<>(min.value(), true));
                            break;
                        case SUM:
                            SumAggregate sum = aggregations.get(metric.getId()).sum();
                            rowCells.add(new TableChartResult.Cell<>(sum.value(), true));
                            break;
                        case AVERAGE:
                            AvgAggregate avg = aggregations.get(metric.getId()).avg();
                            rowCells.add(new TableChartResult.Cell<>(avg.value(), true));
                            break;
                        case COUNT:
                            rowCells.add(new TableChartResult.Cell<>(parent.getDocCount(), true));
                            break;
                    }
                }
                table.addRow(new ArrayList<>(rowCells));
                rowCells.clear();
            }
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    private void parseMetric(SearchResponse<ObjectNode> result, List<Metric> metrics) {
        final String ctx = CLASSNAME + ".parseMetric";
        try {
            Map<String, Aggregate> _metrics = result.aggregations();
            for (Metric metric : metrics) {
                switch (metric.getAggregation()) {
                    case MAX:
                        MaxAggregate max = _metrics.get(metric.getId()).max();
                        rowCells.add(new TableChartResult.Cell<>(max.value(), true));
                        break;
                    case MIN:
                        MinAggregate min = _metrics.get(metric.getId()).min();
                        rowCells.add(new TableChartResult.Cell<>(min.value(), true));
                        break;
                    case SUM:
                        SumAggregate sum = _metrics.get(metric.getId()).sum();
                        rowCells.add(new TableChartResult.Cell<>(sum.value(), true));
                        break;
                    case AVERAGE:
                        AvgAggregate avg = _metrics.get(metric.getId()).avg();
                        rowCells.add(new TableChartResult.Cell<>(avg.value(), true));
                        break;
                    case COUNT:
                        rowCells.add(new TableChartResult.Cell<>(result.hits().total().value(), true));
                        break;
                }
            }
            table.addRow(new ArrayList<>(rowCells));
            rowCells.clear();
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }
}

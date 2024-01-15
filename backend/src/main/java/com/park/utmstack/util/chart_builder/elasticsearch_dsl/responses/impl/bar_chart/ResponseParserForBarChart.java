package com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.bar_chart;

import com.fasterxml.jackson.databind.node.ObjectNode;
import com.park.utmstack.domain.chart_builder.UtmVisualization;
import com.park.utmstack.domain.chart_builder.types.aggregation.AggregationType;
import com.park.utmstack.domain.chart_builder.types.aggregation.Bucket;
import com.park.utmstack.domain.chart_builder.types.aggregation.Metric;
import com.park.utmstack.domain.chart_builder.types.aggregation.enums.BucketType;
import com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.ResponseParser;
import com.utmstack.opensearch_connector.parsers.DateHistogramAggregateParser;
import com.utmstack.opensearch_connector.parsers.TermAggregateParser;
import com.utmstack.opensearch_connector.types.BucketAggregation;
import org.opensearch.client.opensearch._types.aggregations.*;
import org.opensearch.client.opensearch.core.SearchResponse;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import java.util.*;

public class ResponseParserForBarChart implements ResponseParser<BarChartResult> {

    private static final String CLASSNAME = "ResponseParserForBarChart";
    private final BarChartResult retValue = new BarChartResult();
    private final List<String> keys = new ArrayList<>();
    private final Map<String, Double> info = new LinkedHashMap<>();

    @Override
    public List<BarChartResult> parse(UtmVisualization visualization, SearchResponse<ObjectNode> result) {
        final String ctx = CLASSNAME + ".parse";
        try {
            AggregationType aggs = visualization.getAggregationType();
            List<Metric> metrics = aggs.getMetrics();
            Bucket bucket = aggs.getBucket();

            if (bucket != null) {
                parseBuckets(bucket, result.aggregations(), metrics, null);

                if (retValue.getCategories().isEmpty())
                    retValue.addCategory("_all");

                List<String> keys = new ArrayList<>(info.keySet());

                for (int i = 0; i < keys.size(); i++) {
                    BarChartResult.Serie serie = new BarChartResult.Serie();
                    String[] key = keys.get(i).split("->");

                    String id = key[key.length - 1];
                    String serieName = keys.get(i).split("->", 2)[1];
                    List<String> serieNameClean = new ArrayList<>(Arrays.asList(serieName.split("->")));
                    serieNameClean.remove(serieNameClean.size() - 1);

                    serie.setMetricId(id);

                    int pos = 0;
                    StringBuilder ser = new StringBuilder();
                    Bucket temporalBucket = bucket.getSubBucket();
                    while (temporalBucket != null) {
                        ser.append(temporalBucket.getField()).append(":=").append(serieNameClean.get(pos)).append(
                            pos != serieNameClean.size() - 1 ? "->" : "");
                        pos++;
                        temporalBucket = temporalBucket.getSubBucket();
                    }

                    serie.setName(ser.toString());

                    boolean keysChange = false;
                    for (String category : retValue.getCategories()) {
                        Double value = info.get(category + "->" + serieName);
                        if (value == null) {
                            serie.addData(0.0);
                        } else {
                            serie.addData(value);
                            info.remove(category + "->" + serieName);
                            keys.remove(category + "->" + serieName);
                            keysChange = true;
                        }
                    }
                    retValue.addSerie(serie);
                    if (keysChange)
                        i--;
                }
            } else {
                parseMetric(result, metrics);
            }

            return Collections.singletonList(retValue);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    private void parseBuckets(Bucket bucket, Map<String, Aggregate> aggregations, List<Metric> metrics,
                              BucketAggregation currentBucket) {
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
                    if (bucket.getType().equals(BucketType.AXIS))
                        retValue.addCategory(_bucket.getKey());
                    keys.add(_bucket.getKey());
                    parseBuckets(bucket.getSubBucket(), _bucket.getSubAggregations(), metrics, _bucket);
                    keys.remove(_bucket.getKey());
                }
            } else {
                StringBuilder k = new StringBuilder();
                if (retValue.getCategories().isEmpty())
                    k.append("_all->");
                for (String key : keys)
                    k.append(key).append("->");

                for (Metric metric : metrics) {
                    switch (metric.getAggregation()) {
                        case MAX:
                            MaxAggregate max = aggregations.get(metric.getId()).max();
                            info.put(k + metric.getId(), max != null ? max.value() : 0.0);
                            break;
                        case MIN:
                            MinAggregate min = aggregations.get(metric.getId()).min();
                            info.put(k + metric.getId(), min != null ? min.value() : 0.0);
                            break;
                        case SUM:
                            SumAggregate sum = aggregations.get(metric.getId()).sum();
                            info.put(k + metric.getId(), sum != null ? sum.value() : 0.0);
                            break;
                        case AVERAGE:
                            AvgAggregate avg = aggregations.get(metric.getId()).avg();
                            info.put(k + metric.getId(), avg != null ? avg.value() : 0.0);
                            break;
                        case COUNT:
                            info.put(k + metric.getId(), currentBucket.getDocCount().doubleValue());
                            break;
                    }
                }
            }
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private void parseMetric(SearchResponse<?> result, List<Metric> metrics) {
        final String ctx = CLASSNAME + ".parseMetric";
        try {
            retValue.addCategory("_all");
            Map<String, Aggregate> aggregations = result.aggregations();
            for (Metric metric : metrics) {
                BarChartResult.Serie serie = new BarChartResult.Serie();
                String seriesName = !StringUtils.hasText(metric.getCustomLabel()) ?
                    metric.getDefaultLabel() : metric.getCustomLabel();
                switch (metric.getAggregation()) {
                    case MAX:
                        MaxAggregate max = aggregations.get(metric.getId()).max();
                        serie.setMetricId(metric.getId());
                        serie.setName(seriesName);
                        serie.addData(max.value());
                        retValue.addSerie(serie);
                        break;
                    case MIN:
                        MinAggregate min = aggregations.get(metric.getId()).min();
                        serie.setMetricId(metric.getId());
                        serie.setName(seriesName);
                        serie.addData(min.value());
                        retValue.addSerie(serie);
                        break;
                    case SUM:
                        SumAggregate sum = aggregations.get(metric.getId()).sum();
                        serie.setMetricId(metric.getId());
                        serie.setName(seriesName);
                        serie.addData(sum.value());
                        retValue.addSerie(serie);
                        break;
                    case AVERAGE:
                        AvgAggregate avg = aggregations.get(metric.getId()).avg();
                        serie.setMetricId(metric.getId());
                        serie.setName(seriesName);
                        serie.addData(avg.value());
                        retValue.addSerie(serie);
                        break;
                    case COUNT:
                        serie.setMetricId(metric.getId());
                        serie.setName(seriesName);
                        serie.addData(Long.valueOf(result.hits().total().value()).doubleValue());
                        retValue.addSerie(serie);
                        break;
                }
            }
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }
}

package com.park.utmstack.util.chart_builder.elasticsearch_dsl.requests;

import com.park.utmstack.domain.chart_builder.UtmVisualization;
import com.park.utmstack.domain.chart_builder.types.ChartType;
import com.park.utmstack.domain.chart_builder.types.aggregation.AggregationType;
import com.park.utmstack.domain.chart_builder.types.aggregation.Bucket;
import com.park.utmstack.domain.chart_builder.types.aggregation.Metric;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.service.elasticsearch.SearchUtil;
import com.park.utmstack.util.TimezoneUtil;
import com.park.utmstack.util.exceptions.UtmElasticsearchException;
import org.opensearch.client.opensearch._types.SortOrder;
import org.opensearch.client.opensearch._types.aggregations.*;
import org.opensearch.client.opensearch.core.SearchRequest;
import org.opensearch.client.opensearch.core.search.TrackHits;
import org.springframework.util.Assert;
import org.springframework.util.CollectionUtils;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

public class RequestDsl {
    private static final String CLASSNAME = "RequestDsl";
    private final SearchRequest.Builder searchRequestBuilder;
    private final UtmVisualization visualization;

    public RequestDsl(UtmVisualization visualization) {
        Assert.notNull(visualization, "Constructor parameter: visualization must not be null");
        searchRequestBuilder = new SearchRequest.Builder().index(visualization.getPattern().getPattern());
        this.visualization = visualization;
    }

    public SearchRequest.Builder getSearchSourceBuilder() throws UtmElasticsearchException {
        final String ctx = CLASSNAME + ".getSearchSourceBuilder";
        try {
            List<FilterType> filters = visualization.getFilterType();

            if (CollectionUtils.isEmpty(filters))
                filters = new ArrayList<>();

            searchRequestBuilder.query(SearchUtil.toQuery(filters));

            if (visualization.getChartType() != ChartType.LIST_CHART &&
                visualization.getChartType() != ChartType.TEXT_CHART) {
                buildAggregation();
                searchRequestBuilder.size(0);
                searchRequestBuilder.trackTotalHits(TrackHits.of(t -> t.enabled(true)));
            } else {
                searchRequestBuilder.size(10000);
            }
            return searchRequestBuilder;
        } catch (Exception e) {
            throw new UtmElasticsearchException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Build an aggregation section for an elasticsearch dsl request
     *
     * @throws UtmElasticsearchException In case of any error
     */
    private void buildAggregation() throws UtmElasticsearchException {
        final String ctx = CLASSNAME + ".buildAggregation";
        try {
            AggregationType aggregationType = visualization.getAggregationType();
            Assert.notNull(aggregationType, "Object UtmVisualization.AggregationType must not be null");

            Bucket bucket = aggregationType.getBucket();
            List<Metric> metrics = aggregationType.getMetrics();

            Map<String, Aggregation.Builder.ContainerBuilder> bucketAggregations = buildBucketAggregation(bucket);
            Map<String, Aggregation> metricAggregations = buildMetricAggregation(metrics);

            if (!CollectionUtils.isEmpty(bucketAggregations)) {
                Map<String, Aggregation> root = new LinkedHashMap<>();
                if (bucketAggregations.size() > 1) {
                    ArrayList<Map.Entry<String, Aggregation.Builder.ContainerBuilder>> buckets = new ArrayList<>(bucketAggregations.entrySet());

                    for (int i = buckets.size() - 1; i > 0; i--) {
                        Map.Entry<String, Aggregation.Builder.ContainerBuilder> child = buckets.get(i);
                        Map.Entry<String, Aggregation.Builder.ContainerBuilder> parent = buckets.get(i - 1);

                        parent.getValue().aggregations(child.getKey(), child.getValue().aggregations(metricAggregations).build());

                        if (i == 1)
                            root.put(parent.getKey(), parent.getValue().aggregations(metricAggregations).build());
                    }
                } else {
                    Map.Entry<String, Aggregation.Builder.ContainerBuilder> agg = bucketAggregations.entrySet().iterator().next();
                    root.put(agg.getKey(), agg.getValue().aggregations(metricAggregations).build());
                }
                searchRequestBuilder.aggregations(root);
            } else {
                metricAggregations.forEach(searchRequestBuilder::aggregations);
            }
        } catch (Exception e) {
            throw new UtmElasticsearchException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Define aggregation buckets
     *
     * @param bucket: Object with needed information to build the buckets
     * @return A list of AggregationBuilder object with bucket defined
     */
    private Map<String, Aggregation.Builder.ContainerBuilder> buildBucketAggregation(Bucket bucket) {
        final String ctx = CLASSNAME + ".buildBucketAggregation";
        Map<String, Aggregation.Builder.ContainerBuilder> bucketAggregations = new LinkedHashMap<>();
        try {
            while (bucket != null) {
                switch (bucket.getAggregation()) {
                    case TERMS:
                        TermsAggregation term = new TermsAggregation.Builder().field(bucket.getField())
                            .size(bucket.getTerms().getSize()).order(List.of(Map.of(bucket.getTerms().getSortBy(),
                                bucket.getTerms().getAsc() ? SortOrder.Asc : SortOrder.Desc))).build();
                        bucketAggregations.put(bucket.toString(), new Aggregation.Builder().terms(term));
                        break;
                    case RANGE:
                        RangeAggregation range = new RangeAggregation.Builder().field(bucket.getField())
                            .ranges(bucket.getRanges().stream().map(r -> AggregationRange.of(a -> a.from(String.valueOf(r.getFrom()))
                                .to(String.valueOf(r.getTo())))).collect(Collectors.toList())).build();
                        bucketAggregations.put(bucket.toString(), new Aggregation.Builder().range(range));
                        break;
                    case DATE_HISTOGRAM:
                        DateHistogramAggregation.Builder histogram = new DateHistogramAggregation.Builder().field(bucket.getField())
                            .timeZone(TimezoneUtil.getAppTimezone().toString());
                        String interval = bucket.getDateHistogram().getInterval();
                        if (bucket.getDateHistogram().isFixedInterval())
                            histogram.fixedInterval(i -> i.time(interval));
                        else if (bucket.getDateHistogram().isCalendarInterval())
                            histogram.calendarInterval(CalendarInterval.valueOf(interval));
                        else
                            throw new Exception(String.format("Interval %1$s is not allowed", interval));
                        bucketAggregations.put(bucket.toString(), new Aggregation.Builder().dateHistogram(histogram.build()));
                        break;
                    default:
                        throw new RuntimeException(String.format("Bucket aggregation %1$s is not implemented", bucket.getAggregation().name()));
                }
                bucket = bucket.getSubBucket();
            }
            return bucketAggregations;
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Define aggregation metrics
     *
     * @param metrics: Object with needed information to build the metrics
     * @return A list of AggregationBuilder object with metrics defined
     */
    private Map<String, Aggregation> buildMetricAggregation(List<Metric> metrics) {
        final String ctx = CLASSNAME + ".buildMetricAggregation";
        Map<String, Aggregation> metricAggregations = new LinkedHashMap<>();
        try {
            for (Metric metric : metrics) {
                switch (metric.getAggregation()) {
                    case AVERAGE:
                        metricAggregations.put(metric.getId(), Aggregation.of(agg ->
                            agg.avg(avg -> avg.field(metric.getField()))));
                        break;
                    case MAX:
                        metricAggregations.put(metric.getId(), Aggregation.of(agg ->
                            agg.max(max -> max.field(metric.getField()))));
                        break;
                    case MIN:
                        metricAggregations.put(metric.getId(), Aggregation.of(agg ->
                            agg.min(min -> min.field(metric.getField()))));
                        break;
                    case SUM:
                        metricAggregations.put(metric.getId(), Aggregation.of(agg ->
                            agg.sum(sum -> sum.field(metric.getField()))));
                        break;
                    case MEDIAN:
                        metricAggregations.put(metric.getId(), Aggregation.of(agg ->
                            agg.percentiles(percentile -> percentile.field(metric.getField())
                                .keyed(false).percents(50D))));
                        break;
                }
            }
            return metricAggregations;
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }
}

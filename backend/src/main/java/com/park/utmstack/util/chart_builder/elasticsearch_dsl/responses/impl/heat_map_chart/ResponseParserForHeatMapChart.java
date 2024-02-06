package com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.heat_map_chart;

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
import org.opensearch.client.opensearch._types.aggregations.Aggregate;
import org.opensearch.client.opensearch.core.SearchResponse;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import java.util.*;


public class ResponseParserForHeatMapChart implements ResponseParser<HeatMapChartResult> {
    private static final String CLASSNAME = "ResponseParserForHeatMapChart";
    private final HeatMapChartResult retValue = new HeatMapChartResult();
    private final List<List<String>> tempData = new ArrayList<>();
    private final String[] coord = new String[3];

    @Override
    public List<HeatMapChartResult> parse(UtmVisualization visualization, SearchResponse<ObjectNode> result) {
        final String ctx = CLASSNAME + ".parse";
        try {
            AggregationType aggs = visualization.getAggregationType();
            Metric metric = aggs.getMetrics().get(0);
            Bucket bucket = aggs.getBucket();

            if (bucket != null) {
                if (!Objects.isNull(bucket.getSubBucket()) && !Objects.isNull(bucket.getSubBucket().getSubBucket()))
                    throw new Exception("Can't apply a third bucket over Heat Map visualizations");

                parseBuckets(bucket, result.aggregations(), metric, null);

                if (CollectionUtils.isEmpty(retValue.getxAxis())) {
                    retValue.addXAxis("_all");
                } else if (CollectionUtils.isEmpty(retValue.getyAxis())) {
                    retValue.addYAxis(StringUtils.hasText(metric.getCustomLabel()) ? metric.getCustomLabel()
                        : metric.getDefaultLabel());
                }

                for (List<String> tempDatum : tempData) {
                    Double[] data = new Double[3];
                    int x = retValue.getxAxis().indexOf(tempDatum.get(0));
                    int y = retValue.getyAxis().indexOf(tempDatum.get(1));
                    data[0] = x == -1 ? 0 : (double) x;
                    data[1] = y == -1 ? 0 : (double) y;
                    data[2] = Double.valueOf(tempDatum.get(2));
                    retValue.addData(data);
                }
            } else {
                retValue.addYAxis(StringUtils.hasText(metric.getCustomLabel()) ? metric.getCustomLabel() : metric
                    .getDefaultLabel());
                retValue.addXAxis("_all");
                Double[] value = new Double[3];
                value[0] = 0.0;
                value[1] = 0.0;
                switch (metric.getAggregation()) {
                    case COUNT:
                        value[2] = (double) result.hits().total().value();
                        break;
                    case AVERAGE:
                        value[2] = result.aggregations().get(metric.getId()).avg().value();
                        break;
                    case SUM:
                        value[2] = result.aggregations().get(metric.getId()).sum().value();
                        break;
                    case MAX:
                        value[2] = result.aggregations().get(metric.getId()).max().value();
                        break;
                    case MIN:
                        value[2] = result.aggregations().get(metric.getId()).min().value();
                        break;
                }
                retValue.addData(value);
            }
            return Collections.singletonList(retValue);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    private void parseBuckets(Bucket bucket, Map<String, Aggregate> aggregation, Metric metric, BucketAggregation currentBucket) {
        if (bucket != null) {
            Aggregate agg = aggregation.get(bucket.getId());
            List<BucketAggregation> entries = new ArrayList<>();

            switch (bucket.getAggregation()) {
                case TERMS:
                    entries = TermAggregateParser.parse(agg);
                    break;
                case DATE_HISTOGRAM:
                    entries = DateHistogramAggregateParser.parse(agg);
                    break;
            }

            if (CollectionUtils.isEmpty(entries))
                return;

            for (BucketAggregation entry : entries) {
                if (bucket.getType().equals(BucketType.AXIS)) {
                    if (!bucket.getTerms().getAsc())
                        retValue.addXAxis(bucket.getField() + ":=" + entry.getKey());
                    else
                        retValue.addXAxis(0, bucket.getField() + ":=" + entry.getKey());
                    coord[0] = bucket.getField() + ":=" + entry.getKey();
                } else {
                    if (!bucket.getTerms().getAsc())
                        retValue.addYAxis(0, bucket.getField() + ":=" + entry.getKey());
                    else
                        retValue.addYAxis(bucket.getField() + ":=" + entry.getKey());
                    coord[1] = bucket.getField() + ":=" + entry.getKey();
                }
                parseBuckets(bucket.getSubBucket(), entry.getSubAggregations(), metric, entry);
                coord[Arrays.asList(coord).indexOf(bucket.getField() + ":=" + entry.getKey())] = "";
            }
        } else {
            switch (metric.getAggregation()) {
                case COUNT:
                    coord[2] = String.valueOf(currentBucket.getDocCount());
                    tempData.add(new ArrayList<>(Arrays.asList(coord)));
                    coord[2] = "";
                    break;
                case AVERAGE:
                    coord[2] = aggregation.get(metric.getId()).avg().toString();
                    tempData.add(new ArrayList<>(Arrays.asList(coord)));
                    coord[2] = "";
                    break;
                case SUM:
                    coord[2] = aggregation.get(metric.getId()).sum().toString();
                    tempData.add(new ArrayList<>(Arrays.asList(coord)));
                    coord[2] = "";
                    break;
                case MAX:
                    coord[2] = aggregation.get(metric.getId()).max().toString();
                    tempData.add(new ArrayList<>(Arrays.asList(coord)));
                    coord[2] = "";
                    break;
                case MIN:
                    coord[2] = aggregation.get(metric.getId()).min().toString();
                    tempData.add(new ArrayList<>(Arrays.asList(coord)));
                    coord[2] = "";
                    break;
            }
        }
    }
}

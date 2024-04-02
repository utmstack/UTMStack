package com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.coordinate_map;

import com.fasterxml.jackson.databind.node.ObjectNode;
import com.park.utmstack.domain.chart_builder.UtmVisualization;
import com.park.utmstack.domain.chart_builder.types.aggregation.AggregationType;
import com.park.utmstack.domain.chart_builder.types.aggregation.Bucket;
import com.park.utmstack.domain.chart_builder.types.aggregation.Metric;
import com.park.utmstack.domain.ip_info.GeoIp;
import com.park.utmstack.service.ip_info.IpInfoService;
import com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.ResponseParser;
import com.park.utmstack.util.exceptions.UtmIpInfoException;
import com.utmstack.opensearch_connector.parsers.TermAggregateParser;
import com.utmstack.opensearch_connector.types.BucketAggregation;
import org.opensearch.client.opensearch.core.SearchResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;
import org.springframework.util.StringUtils;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

@Component
public class ResponseParserForCoordinateMapChart implements ResponseParser<CoordinateMapChartResult> {
    private static final String CLASSNAME = "ResponseParserForCoordinateMapChart";

    private final Logger log = LoggerFactory.getLogger(ResponseParserForCoordinateMapChart.class);

    private final IpInfoService ipInfoService;

    public ResponseParserForCoordinateMapChart(IpInfoService ipInfoService) {
        this.ipInfoService = ipInfoService;
    }

    @Override
    public List<CoordinateMapChartResult> parse(UtmVisualization visualization, SearchResponse<ObjectNode> result) {
        final String ctx = CLASSNAME + ".parse";
        List<CoordinateMapChartResult> retValue = new ArrayList<>();

        try {
            AggregationType aggs = visualization.getAggregationType();
            Metric metric = aggs.getMetrics().get(0);
            Bucket bucket = aggs.getBucket();

            if (bucket != null) {
                List<BucketAggregation> entries = TermAggregateParser.parse(result.aggregations().get(bucket.getId()));
                entries = entries.stream().filter(e -> StringUtils.hasText(e.getKey())).collect(Collectors.toList());

                for (BucketAggregation entry : entries) {
                    GeoIp ipV4Info;
                    try {
                        ipV4Info = ipInfoService.getIpInfo(entry.getKey());

                        if (ipV4Info == null)
                            continue;
                    } catch (UtmIpInfoException e) {
                        log.error(e.getMessage());
                        continue;
                    }

                    CoordinateMapChartResult value = new CoordinateMapChartResult();
                    value.setName(entry.getKey());
                    value.addLatitude(ipV4Info.getLatitude()).addLongitude(ipV4Info.getLongitude());

                    switch (metric.getAggregation()) {
                        case COUNT:
                            value.addMetricValue(entry.getDocCount().doubleValue());
                            break;
                        case MAX:
                            value.addMetricValue(entry.getSubAggregations().get(metric.getId()).max().value());
                            break;
                        case MIN:
                            value.addMetricValue(entry.getSubAggregations().get(metric.getId()).min().value());
                            break;
                        case SUM:
                            value.addMetricValue(entry.getSubAggregations().get(metric.getId()).sum().value());
                            break;
                        case AVERAGE:
                            value.addMetricValue(entry.getSubAggregations().get(metric.getId()).avg().value());
                            break;
                    }
                    retValue.add(value);
                }
            }
            return retValue;
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }
}

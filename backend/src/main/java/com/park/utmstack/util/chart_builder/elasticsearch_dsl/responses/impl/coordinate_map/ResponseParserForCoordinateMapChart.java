package com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.coordinate_map;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.park.utmstack.domain.chart_builder.UtmVisualization;
import com.park.utmstack.domain.chart_builder.types.aggregation.AggregationType;
import com.park.utmstack.domain.chart_builder.types.aggregation.Bucket;
import com.park.utmstack.domain.chart_builder.types.aggregation.Metric;
import com.park.utmstack.util.chart_builder.elasticsearch_dsl.requests.RequestDsl;
import com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.ResponseParser;
import com.utmstack.opensearch_connector.parsers.TermAggregateParser;
import com.utmstack.opensearch_connector.types.BucketAggregation;
import com.utmstack.opensearch_connector.types.SearchSqlResponse;
import org.opensearch.client.opensearch._types.aggregations.Aggregate;
import org.opensearch.client.opensearch._types.aggregations.TopHitsAggregate;
import org.opensearch.client.opensearch.core.SearchResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;
import org.springframework.util.Assert;
import org.springframework.util.StringUtils;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@Component
public class ResponseParserForCoordinateMapChart implements ResponseParser<CoordinateMapChartResult> {
    private static final String CLASSNAME = "ResponseParserForCoordinateMapChart";
    private static final List<String> GEO_PREFIXES = List.of("origin", "source", "destination");

    private final Logger log = LoggerFactory.getLogger(ResponseParserForCoordinateMapChart.class);

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
                entries = entries.stream().filter(e -> isValidIP(e.getKey()))
                        .collect(Collectors.toList());


                for (BucketAggregation entry : entries) {
                    Double[] latLon = extractLatLongFromTopHits(entry.getSubAggregations());
                    if (latLon == null)
                        continue;

                    CoordinateMapChartResult value = new CoordinateMapChartResult();
                    value.setName(entry.getKey());
                    value.addLatitude(latLon[0]).addLongitude(latLon[1]);

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

    public static boolean isValidIP(String ip) {
        return isValidIPv4(ip) || isValidIPv6(ip);
    }


    public static boolean isValidIPv4(String ip) {
        if (ip == null || ip.isEmpty()) return false;
        String regex =
                "^((25[0-5]|2[0-4]\\d|1\\d{2}|[1-9]?\\d)(\\.|$)){4}$";
        return ip.matches(regex);
    }

    public static boolean isValidIPv6(String ip) {
        if (ip == null || ip.isEmpty()) return false;
        String regex =
                "^(?:[\\da-fA-F]{1,4}:){7}[\\da-fA-F]{1,4}$";
        return ip.matches(regex);
    }

    @Override
    public List<CoordinateMapChartResult> parse(UtmVisualization visualization, SearchSqlResponse<Map> result) {
        final String ctx = CLASSNAME + ".parse(SearchSqlResponse)";
        List<CoordinateMapChartResult> retValue = new ArrayList<>();

        try {
            Assert.notNull(visualization, "Param visualization must not be null");
            List<?> data = result.getData();

            for (int i = 0; i < data.size(); i++) {
                Object rowObj = data.get(i);
                if (!(rowObj instanceof Map)) continue;
                Map<String, Object> row = (Map<String, Object>) rowObj;

                String ip = null;

                for (Map.Entry<String, Object> entry : row.entrySet()) {
                    Object val = entry.getValue();
                    if (val == null) continue;

                    String strVal = val.toString();
                    if (ip == null && isValidIP(strVal)) {
                        ip = strVal;
                    }
                }
                if (!StringUtils.hasText(ip)) continue;

                Double[] latLon = extractLatLongFromRow(row);
                if (latLon == null) continue;

                CoordinateMapChartResult chartResult = new CoordinateMapChartResult();
                chartResult.setName(ip);
                chartResult.setValue(new Double[] {
                        latLon[0],
                        latLon[1],
                        (double) i
                });

                retValue.add(chartResult);
            }

            return retValue;
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage(), e);
        }
    }

    private Double[] extractLatLongFromTopHits(Map<String, Aggregate> subAggs) {
        if (subAggs == null) return null;
        Aggregate geoAgg = subAggs.get(RequestDsl.GEO_HIT_AGG);
        if (geoAgg == null || !geoAgg.isTopHits()) return null;
        TopHitsAggregate topHits = geoAgg.topHits();
        if (topHits.hits() == null || topHits.hits().hits().isEmpty()) return null;
        var hit = topHits.hits().hits().get(0);
        if (hit.source() == null) return null;
        try {
            ObjectNode source = hit.source().to(ObjectNode.class);
            return extractLatLongFromJson(source);
        } catch (Exception e) {
            log.warn("Failed to deserialize top_hits source for geolocation: {}", e.getMessage());
            return null;
        }
    }

    private Double[] extractLatLongFromJson(ObjectNode source) {
        if (source == null) return null;
        for (String prefix : GEO_PREFIXES) {
            JsonNode geo = source.path(prefix).path("geolocation");
            JsonNode lat = geo.path("latitude");
            JsonNode lon = geo.path("longitude");
            if (lat.isNumber() && lon.isNumber()) {
                return new Double[]{lat.asDouble(), lon.asDouble()};
            }
        }
        return null;
    }

    private Double[] extractLatLongFromRow(Map<String, Object> row) {
        Double lat = null;
        Double lon = null;
        for (Map.Entry<String, Object> e : row.entrySet()) {
            if (e.getValue() == null) continue;
            String key = e.getKey();
            if (lat == null && (key.endsWith(".geolocation.latitude") || key.equals("latitude"))) {
                lat = toDouble(e.getValue());
            } else if (lon == null && (key.endsWith(".geolocation.longitude") || key.equals("longitude"))) {
                lon = toDouble(e.getValue());
            }
        }
        if (lat == null || lon == null) return null;
        return new Double[]{lat, lon};
    }

    private Double toDouble(Object val) {
        if (val instanceof Number) return ((Number) val).doubleValue();
        try {
            return Double.parseDouble(val.toString());
        } catch (NumberFormatException e) {
            return null;
        }
    }
}

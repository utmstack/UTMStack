package com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.list_chart;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.park.utmstack.domain.chart_builder.UtmVisualization;
import com.park.utmstack.domain.chart_builder.types.aggregation.Bucket;
import com.park.utmstack.util.MapUtil;
import com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.ResponseParser;
import com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.table_chart.TableChartResult;
import org.opensearch.client.opensearch.core.SearchResponse;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

public class ResponseParserForListChart implements ResponseParser<TableChartResult> {
    private static final String CLASSNAME = "ResponseParserForListChart";

    private final TableChartResult retValue = new TableChartResult();
    private final List<String> columns = new ArrayList<>();

    @Override
    public List<TableChartResult> parse(UtmVisualization visualization, SearchResponse<ObjectNode> result) {
        final String ctx = CLASSNAME + ".parse";
        try {
            if (visualization.getAggregationType() == null || visualization.getAggregationType().getBucket() == null)
                return Collections.singletonList(retValue);

            Bucket bucket = visualization.getAggregationType().getBucket();

            if (bucket == null)
                return Collections.singletonList(retValue);

            extractColumns(bucket);

            if (result == null || result.hits().total().value() == 0)
                return Collections.singletonList(retValue);

            List<Map<String, ?>> sources = result.hits().hits().stream().map(hit -> {
                try {
                    return new ObjectMapper().convertValue(hit.source(), new TypeReference<Map<String, ?>>() {
                    });
                } catch (Exception e) {
                    throw new RuntimeException(e);
                }
            }).collect(Collectors.toList());

            if (CollectionUtils.isEmpty(sources))
                return Collections.singletonList(retValue);

            for (Map<String, ?> src : sources) {
                List<TableChartResult.Cell<?>> cells = new ArrayList<>();
                Map<String, String> map = MapUtil.flattenToStringMap(src, true);
                for (String column : columns) {
                    TableChartResult.Cell<String> cell = new TableChartResult.Cell<>();
                    cell.setValue(map.get(column));
                    cells.add(cell);
                }
                retValue.addRow(cells);
            }
            return Collections.singletonList(retValue);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    private void extractColumns(Bucket bucket) {
        final String ctx = CLASSNAME + ".extractColumns";
        try {
            while (bucket != null) {
                String col = !StringUtils.hasText(bucket.getCustomLabel()) ? bucket.getField() + "->" + bucket
                    .getField() : bucket.getField() + "->" + bucket.getCustomLabel();
                retValue.addColumn(col);
                columns.add(bucket.getField().split(".keyword")[0]);
                bucket = bucket.getSubBucket();
            }
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }
}

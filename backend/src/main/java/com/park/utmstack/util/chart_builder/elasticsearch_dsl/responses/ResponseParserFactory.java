package com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses;

import com.park.utmstack.domain.chart_builder.types.ChartType;
import com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.bar_chart.ResponseParserForBarChart;
import com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.coordinate_map.ResponseParserForCoordinateMapChart;
import com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.gauge_goal_chart.ResponseParserForGaugeGoalChart;
import com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.heat_map_chart.ResponseParserForHeatMapChart;
import com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.list_chart.ResponseParserForListChart;
import com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.metric_chart.ResponseParserForMetricChart;
import com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.pie_chart.ResponseParserForPieChart;
import com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.table_chart.ResponseParserForTableChart;
import com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.tag_cloud_chart.ResponseParserForTagCloudChart;
import com.park.utmstack.util.exceptions.UtmChartBuilderException;
import org.springframework.stereotype.Component;

@Component
public class ResponseParserFactory {

    private static final String CLASS_NAME = "ResponseParserFactory";

    private final ResponseParserForCoordinateMapChart responseParserForCoordinateMapChart;

    public ResponseParserFactory(ResponseParserForCoordinateMapChart responseParserForCoordinateMapChart) {
        this.responseParserForCoordinateMapChart = responseParserForCoordinateMapChart;
    }

    public ResponseParser instance(ChartType chartType) throws UtmChartBuilderException {
        final String ctx = CLASS_NAME + ".instance";

        if (chartType.equals(ChartType.METRIC_CHART))
            return new ResponseParserForMetricChart();
        else if (chartType.equals(ChartType.PIE_CHART))
            return new ResponseParserForPieChart();
        else if (chartType.equals(ChartType.GAUGE_CHART) || chartType.equals(ChartType.GOAL_CHART))
            return new ResponseParserForGaugeGoalChart();
        else if (chartType.equals(ChartType.TAG_CLOUD_CHART))
            return new ResponseParserForTagCloudChart();
        else if (chartType.equals(ChartType.TABLE_CHART))
            return new ResponseParserForTableChart();
        else if (chartType.equals(ChartType.LINE_CHART) || chartType.equals(ChartType.AREA_CHART) || chartType.equals(
            ChartType.VERTICAL_BAR_CHART) || chartType.equals(ChartType.HORIZONTAL_BAR_CHART))
            return new ResponseParserForBarChart();
        else if (chartType.equals(ChartType.HEATMAP_CHART))
            return new ResponseParserForHeatMapChart();
        else if (chartType.equals(ChartType.COORDINATE_MAP_CHART))
            return responseParserForCoordinateMapChart;
        else if (chartType.equals(ChartType.LIST_CHART))
            return new ResponseParserForListChart();

        throw new UtmChartBuilderException(ctx + ": No implementation founded for chart type: " + chartType.name());
    }
}

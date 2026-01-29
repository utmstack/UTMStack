package com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.pie_chart;

import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
public class PieChartResult {
    private String metricId;
    private String bucketKey;
    private Double value;
    private String bucketId;

    public PieChartResult(String metricId, double value, String bucketKey, String bucketId) {
        this.metricId = metricId;
        this.bucketKey = bucketKey;
        this.value = value;
        this.bucketId = bucketId;
    }


}

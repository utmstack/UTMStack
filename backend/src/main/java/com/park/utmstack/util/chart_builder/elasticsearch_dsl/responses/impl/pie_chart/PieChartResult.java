package com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.pie_chart;

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

    public String getBucketKey() {
        return bucketKey;
    }

    public void setBucketKey(String bucketKey) {
        this.bucketKey = bucketKey;
    }

    public Double getValue() {
        return value;
    }

    public void setValue(Double value) {
        this.value = value;
    }

    public String getMetricId() {
        return metricId;
    }

    public void setMetricId(String metricId) {
        this.metricId = metricId;
    }

    public String getBucketId() {
        return bucketId;
    }

    public void setBucketId(String bucketId) {
        this.bucketId = bucketId;
    }
}

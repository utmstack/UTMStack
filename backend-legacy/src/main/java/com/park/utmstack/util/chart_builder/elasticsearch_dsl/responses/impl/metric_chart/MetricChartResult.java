package com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.metric_chart;

public class MetricChartResult {
    private String metricId;
    private double value;
    private String bucketKey;
    private String bucketId;

    public MetricChartResult(String metricId, double value, String bucketKey, String bucketId) {
        this.metricId = metricId;
        this.value = value;
        this.bucketKey = bucketKey;
        this.bucketId = bucketId;
    }

    public String getMetricId() {
        return metricId;
    }

    public void setMetricId(String metricId) {
        this.metricId = metricId;
    }

    public double getValue() {
        return value;
    }

    public void setValue(double value) {
        this.value = value;
    }

    public String getBucketKey() {
        return bucketKey;
    }

    public void setBucketKey(String bucketKey) {
        this.bucketKey = bucketKey;
    }

    public String getBucketId() {
        return bucketId;
    }

    public void setBucketId(String bucketId) {
        this.bucketId = bucketId;
    }
}

package com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.tag_cloud_chart;

public class TagCloudChartResult {
    private String bucketKey;
    private double value;
    private String metricId;
    private String bucketId;

    public TagCloudChartResult() {
    }

    public TagCloudChartResult(String bucketKey, double value, String metricId, String bucketId) {
        this.bucketKey = bucketKey;
        this.value = value;
        this.metricId = metricId;
        this.bucketId = bucketId;
    }

    public String getBucketKey() {
        return bucketKey;
    }

    public void setBucketKey(String bucketKey) {
        this.bucketKey = bucketKey;
    }

    public double getValue() {
        return value;
    }

    public void setValue(double value) {
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

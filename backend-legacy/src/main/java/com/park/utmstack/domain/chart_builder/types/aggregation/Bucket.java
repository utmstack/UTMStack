package com.park.utmstack.domain.chart_builder.types.aggregation;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.park.utmstack.domain.chart_builder.types.aggregation.enums.BucketAggregationEnum;
import com.park.utmstack.domain.chart_builder.types.aggregation.enums.BucketType;
import org.springframework.util.Assert;

import java.io.Serializable;
import java.util.List;

@JsonInclude(JsonInclude.Include.NON_NULL)
public class Bucket implements Serializable {
    private String id;
    private BucketAggregationEnum aggregation;
    private BucketType type;
    private String field;
    private String customLabel;
    private Bucket subBucket;
    private TermsBucket terms;
    private DateHistogramBucket dateHistogram;
    private List<RangeBucket> ranges;

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public BucketAggregationEnum getAggregation() {
        return aggregation;
    }

    public void setAggregation(BucketAggregationEnum aggregation) {
        this.aggregation = aggregation;
    }

    public BucketType getType() {
        return type;
    }

    public void setType(BucketType type) {
        this.type = type;
    }

    public String getField() {
        return field;
    }

    public void setField(String field) {
        this.field = field;
    }

    public String getCustomLabel() {
        return customLabel;
    }

    public void setCustomLabel(String customLabel) {
        this.customLabel = customLabel;
    }

    public Bucket getSubBucket() {
        return subBucket;
    }

    public void setSubBucket(Bucket subBucket) {
        this.subBucket = subBucket;
    }

    public TermsBucket getTerms() {
        return terms;
    }

    public void setTerms(TermsBucket terms) {
        this.terms = terms;
    }

    public List<RangeBucket> getRanges() {
        return ranges;
    }

    public void setRanges(List<RangeBucket> ranges) {
        this.ranges = ranges;
    }

    public DateHistogramBucket getDateHistogram() {
        return dateHistogram;
    }

    public void setDateHistogram(DateHistogramBucket dateHistogram) {
        this.dateHistogram = dateHistogram;
    }

    @Override
    public String toString() {
        Assert.hasText(id, "Bucket.id must not be null");
        return id;
    }
}

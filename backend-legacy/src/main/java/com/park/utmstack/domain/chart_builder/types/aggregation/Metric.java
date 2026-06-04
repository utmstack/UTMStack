package com.park.utmstack.domain.chart_builder.types.aggregation;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonInclude;
import com.park.utmstack.domain.chart_builder.types.aggregation.enums.MetricAggregationEnum;

import java.io.Serializable;

@JsonInclude(JsonInclude.Include.NON_NULL)
public class Metric implements Serializable {
    private String id;
    private MetricAggregationEnum aggregation;
    private String field;
    private String customLabel;

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public MetricAggregationEnum getAggregation() {
        return aggregation;
    }

    public void setAggregation(MetricAggregationEnum aggregation) {
        this.aggregation = aggregation;
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

    /**
     * If customLabel is null this method can provide a default label for each metric type
     *
     * @return A string with the default label provided
     */
    @JsonIgnore
    public String getDefaultLabel() {
        switch (aggregation) {
            case AVERAGE:
                return "Average " + field;
            case COUNT:
                return "Count";
            case MAX:
                return "Max " + field;
            case MEDIAN:
                return "50th percentile of " + field;
            case MIN:
                return "Min " + field;
            case SUM:
                return "Sum of " + field;
        }
        return null;
    }
}

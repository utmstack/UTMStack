package com.park.utmstack.domain.chart_builder.types.aggregation;

import com.fasterxml.jackson.annotation.JsonIgnore;
import org.opensearch.client.opensearch._types.aggregations.CalendarInterval;

public class DateHistogramBucket {
    private static final String CLASS_NAME = "DateHistogramBucket";
    private String interval;

    public String getInterval() {
        return interval;
    }

    public void setInterval(String interval) {
        this.interval = interval;
    }

    @JsonIgnore
    public boolean isFixedInterval() {
        return interval.matches("^\\d+(?:ms|m|s|h|d)$");
    }

    @JsonIgnore
    public boolean isCalendarInterval() {
        try {
            CalendarInterval.valueOf(interval);
            return true;
        } catch (Exception e) {
            return false;
        }
    }
}

package com.park.utmstack.service.logstash_pipeline.response.jvm_stats;

import com.fasterxml.jackson.annotation.JsonGetter;
import com.fasterxml.jackson.annotation.JsonSetter;

public class LogstashJVMThreads {
    Integer count = 0;
    Integer peakCount = 0;

    public LogstashJVMThreads(){}
    public LogstashJVMThreads(Integer count, Integer peakCount) {
        this.count = count;
        this.peakCount = peakCount;
    }

    public Integer getCount() {
        return count;
    }

    public void setCount(Integer count) {
        this.count = count;
    }

    @JsonGetter("peakCount")
    public Integer getPeakCount() {
        return peakCount;
    }

    @JsonSetter("peak_count")
    public void setPeakCount(Integer peakCount) {
        this.peakCount = peakCount;
    }
}

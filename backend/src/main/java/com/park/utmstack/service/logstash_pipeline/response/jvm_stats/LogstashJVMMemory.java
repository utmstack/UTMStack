package com.park.utmstack.service.logstash_pipeline.response.jvm_stats;

import com.fasterxml.jackson.annotation.JsonGetter;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonSetter;

public class LogstashJVMMemory {
    Integer heapUsedPercent = 0;
    Long heapCommittedInBytes = 0L;
    Long heapMaxInBytes = 0L;
    Long heapUsedInBytes = 0L;
    Long nonHeapUsedInBytes = 0L;
    Long nonHeapCommittedInBytes = 0L;
    @JsonIgnore
    Object pools;

    public LogstashJVMMemory(){}
    public LogstashJVMMemory(Integer heapUsedPercent,
                             Long heapCommittedInBytes,
                             Long heapMaxInBytes,
                             Long heapUsedInBytes,
                             Long nonHeapUsedInBytes,
                             Long nonHeapCommittedInBytes,
                             Object pools) {
        this.heapUsedPercent = heapUsedPercent;
        this.heapCommittedInBytes = heapCommittedInBytes;
        this.heapMaxInBytes = heapMaxInBytes;
        this.heapUsedInBytes = heapUsedInBytes;
        this.nonHeapUsedInBytes = nonHeapUsedInBytes;
        this.nonHeapCommittedInBytes = nonHeapCommittedInBytes;
        this.pools = pools;
    }

    @JsonGetter("heapUsedPercent")
    public Integer getHeapUsedPercent() {
        return heapUsedPercent;
    }

    @JsonSetter("heap_used_percent")
    public void setHeapUsedPercent(Integer heapUsedPercent) {
        this.heapUsedPercent = heapUsedPercent;
    }

    @JsonGetter("heapCommittedInBytes")
    public Long getHeapCommittedInBytes() {
        return heapCommittedInBytes;
    }

    @JsonSetter("heap_committed_in_bytes")
    public void setHeapCommittedInBytes(Long heapCommittedInBytes) {
        this.heapCommittedInBytes = heapCommittedInBytes;
    }

    @JsonGetter("heapMaxInBytes")
    public Long getHeapMaxInBytes() {
        return heapMaxInBytes;
    }

    @JsonSetter("heap_max_in_bytes")
    public void setHeapMaxInBytes(Long heapMaxInBytes) {
        this.heapMaxInBytes = heapMaxInBytes;
    }

    @JsonGetter("heapUsedInBytes")
    public Long getHeapUsedInBytes() {
        return heapUsedInBytes;
    }

    @JsonSetter("heap_used_in_bytes")
    public void setHeapUsedInBytes(Long heapUsedInBytes) {
        this.heapUsedInBytes = heapUsedInBytes;
    }

    @JsonGetter("nonHeapUsedInBytes")
    public Long getNonHeapUsedInBytes() {
        return nonHeapUsedInBytes;
    }

    @JsonSetter("non_heap_used_in_bytes")
    public void setNonHeapUsedInBytes(Long nonHeapUsedInBytes) {
        this.nonHeapUsedInBytes = nonHeapUsedInBytes;
    }

    @JsonGetter("nonHeapCommittedInBytes")
    public Long getNonHeapCommittedInBytes() {
        return nonHeapCommittedInBytes;
    }

    @JsonSetter("non_heap_committed_in_bytes")
    public void setNonHeapCommittedInBytes(Long nonHeapCommittedInBytes) {
        this.nonHeapCommittedInBytes = nonHeapCommittedInBytes;
    }

    public Object getPools() {
        return pools;
    }

    public void setPools(Object pools) {
        this.pools = pools;
    }
}

package com.park.utmstack.service.logstash_pipeline.response.jvm_stats;

import com.fasterxml.jackson.annotation.JsonGetter;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonSetter;

public class JvmConfig {
    LogstashJVMThreads threads;
    LogstashJVMMemory mem;
    @JsonIgnore
    Object gc;
    Long uptimeInMillis = 0L;

    public JvmConfig(){
        this.threads = new LogstashJVMThreads();
        this.mem = new LogstashJVMMemory();
    }
    public JvmConfig(LogstashJVMThreads threads, LogstashJVMMemory mem, Object gc, Long uptimeInMillis) {
        this.threads = threads;
        this.mem = mem;
        this.gc = gc;
        this.uptimeInMillis = uptimeInMillis;
    }

    public LogstashJVMThreads getThreads() {
        return threads;
    }

    public void setThreads(LogstashJVMThreads threads) {
        this.threads = threads;
    }

    public LogstashJVMMemory getMem() {
        return mem;
    }

    public void setMem(LogstashJVMMemory mem) {
        this.mem = mem;
    }

    public Object getGc() {
        return gc;
    }

    public void setGc(Object gc) {
        this.gc = gc;
    }

    @JsonGetter("uptimeInMillis")
    public Long getUptimeInMillis() {
        return uptimeInMillis;
    }

    @JsonSetter("uptime_in_millis")
    public void setUptimeInMillis(Long uptimeInMillis) {
        this.uptimeInMillis = uptimeInMillis;
    }
}

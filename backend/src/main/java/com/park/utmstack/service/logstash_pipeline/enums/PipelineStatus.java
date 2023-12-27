package com.park.utmstack.service.logstash_pipeline.enums;

public enum PipelineStatus {
    PIPELINE_STATUS_UP("up"),
    PIPELINE_STATUS_DOWN("down"),
    LOGSTASH_STATUS_DOWN("down"),
    LOGSTASH_STATUS_RED("red"),
    LOGSTASH_STATUS_YELLOW("yellow"),
    LOGSTASH_STATUS_GREEN("green");

    private String status;
    PipelineStatus(String status){
        this.status = status;
    }
    public String get(){
        return this.status;
    }
}

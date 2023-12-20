package com.park.utmstack.service.logstash_pipeline.enums;

public enum PipelineRelation {
    PIPELINE_FILTER("PIPELINE FILTER"),
    USER_CUSTOM_FILTER("USER_CUSTOM_FILTER");

    private String relation;
    PipelineRelation(String relation){
        this.relation = relation;
    }
    public String getRelation(){
        return this.relation;
    }
}

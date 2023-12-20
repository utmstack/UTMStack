package com.park.utmstack.domain.logstash_pipeline.enums;

/**
 * Supported input configuration types
 */
public enum InputConfigTypes {
    PORT("port"),
    STRING("String"),
    PATH("path");
    private String value;
    InputConfigTypes(String value){
        this.value = value;
    }
    public String getValue(){
        return this.value;
    }
}

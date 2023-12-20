package com.park.utmstack.domain.logstash_pipeline.enums;

public enum ValidationRegexp {
    PORT("^((6553[0-5])|(655[0-2][0-9])|(65[0-4][0-9]{2})|(6[0-4][0-9]{3})|([1-5][0-9]{4})|([0-5]{0,5})|([0-9]{1,4}))$"),
    ONLY_STRING("[a-zA-Z_0-9]");

    private String regexp;
    ValidationRegexp(String regexp){
        this.regexp = regexp;
    }
    public String getRegexp(){
        return this.regexp;
    }
}

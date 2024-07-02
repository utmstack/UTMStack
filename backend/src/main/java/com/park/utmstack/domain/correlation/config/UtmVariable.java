package com.park.utmstack.domain.correlation.config;

import com.fasterxml.jackson.annotation.JsonProperty;

public class UtmVariable {
    @JsonProperty("get")
    private String field;

    @JsonProperty("as")
    private String name;

    @JsonProperty("of_type")
    private String type;

    public UtmVariable() {
    }

    public String getField() {
        return field;
    }

    public void setField(String field) {
        this.field = field;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getType() {
        return type;
    }

    public void setType(String type) {
        this.type = type;
    }
}

package com.park.utmstack.domain.correlation.rules;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Getter;
import lombok.Setter;

@Setter
@Getter
public class RuleVariable {
    @JsonProperty("get")
    private String field;

    @JsonProperty("as")
    private String name;

    @JsonProperty("ofType")
    private String type;

    public RuleVariable() {
    }

}

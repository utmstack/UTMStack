package com.park.utmstack.domain.correlation.rules;

import lombok.Data;

@Data
public class Expression {
    private String field;
    private Operator operator;
    private Object value;
}

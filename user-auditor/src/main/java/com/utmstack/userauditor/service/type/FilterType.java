package com.utmstack.userauditor.service.type;

import com.fasterxml.jackson.annotation.JsonInclude;
import org.springframework.util.Assert;

import java.io.Serializable;
import java.util.List;

@JsonInclude(JsonInclude.Include.NON_NULL)
public class FilterType implements Serializable {
    private OperatorType operator;
    private String field;
    private Object value;

    public FilterType() {
    }

    public FilterType(String field, OperatorType operator, Object value) {
        this.operator = operator;
        this.field = field;
        this.value = value;
    }

    public OperatorType getOperator() {
        return operator;
    }

    public void setOperator(OperatorType operator) {
        this.operator = operator;
    }

    public String getField() {
        return field;
    }

    public void setField(String field) {
        this.field = field;
    }

    public Object getValue() {
        return value;
    }

    public void setValue(Object value) {
        this.value = value;
    }

    public void validate() {
        switch (operator) {
            case IS_IN_FIELDS:
            case IS_NOT_IN_FIELDS:
                Assert.notNull(value, "Filter value is missing");
                break;
            case IS_BETWEEN:
            case IS_NOT_BETWEEN:
                Assert.hasText(field, "Filter field is missing");
                Assert.isInstanceOf(List.class, value, "Value of filter has to be an instance of list");
                break;
            case IS_ONE_OF_TERMS:
            case IS_ONE_OF:
            case IS_NOT_ONE_OF:
                Assert.isInstanceOf(List.class, value, "Value of filter has to be an instance of list");
                break;
            case EXIST:
            case DOES_NOT_EXIST:
                Assert.hasText(field, "Filter field is missing");
                break;
            default:
                Assert.hasText(field, "Filter field is missing");
                Assert.notNull(value, "Filter value is missing");
                break;
        }
    }
}

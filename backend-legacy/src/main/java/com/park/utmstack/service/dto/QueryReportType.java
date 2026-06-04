package com.park.utmstack.service.dto;

import java.io.Serializable;

public class QueryReportType implements Serializable {
    private static final long serialVersionUID = 1L;

    private String field;
    private String fieldLabel;
    private String operator;
    private String value;

    public String getField() {
        return field;
    }

    public void setField(String field) {
        this.field = field;
    }

    public String getFieldLabel() {
        return fieldLabel;
    }

    public void setFieldLabel(String fieldLabel) {
        this.fieldLabel = fieldLabel;
    }

    public String getOperator() {
        return operator;
    }

    public void setOperator(String operator) {
        this.operator = operator;
    }

    public String getValue() {
        return value;
    }

    public void setValue(String value) {
        this.value = value;
    }

    @Override
    public String toString() {
        return field + " " + operator + " " + value;
    }
}

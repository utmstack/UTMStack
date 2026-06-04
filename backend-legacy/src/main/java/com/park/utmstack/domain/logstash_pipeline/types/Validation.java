package com.park.utmstack.domain.logstash_pipeline.types;

public class Validation {
    String entity;
    String field;
    String msg;

    public Validation(String entity, String field, String msg) {
        this.entity = entity;
        this.field = field;
        this.msg = msg;
    }

    public String getEntity() {
        return entity;
    }

    public void setEntity(String entity) {
        this.entity = entity;
    }

    public String getField() {
        return field;
    }

    public void setField(String field) {
        this.field = field;
    }

    public String getMsg() {
        return msg;
    }

    public void setMsg(String msg) {
        this.msg = msg;
    }
}

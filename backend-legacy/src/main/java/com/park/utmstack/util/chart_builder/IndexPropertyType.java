package com.park.utmstack.util.chart_builder;

import java.io.Serializable;

/**
 * Simulate a field of a document of a index in elasticsearch
 */
public class IndexPropertyType implements Serializable {
    private String name;
    private String type;

    public IndexPropertyType() {
    }

    public IndexPropertyType(String name, String type) {
        this.name = name;
        this.type = type;
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

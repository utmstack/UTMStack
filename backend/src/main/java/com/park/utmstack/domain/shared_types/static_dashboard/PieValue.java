package com.park.utmstack.domain.shared_types.static_dashboard;

public class PieValue {
    private Integer value;
    private String name;

    public PieValue() {
    }

    public PieValue(Integer value, String name) {
        this.value = value;
        this.name = name;
    }

    public Integer getValue() {
        return value;
    }

    public void setValue(Integer value) {
        this.value = value;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }
}

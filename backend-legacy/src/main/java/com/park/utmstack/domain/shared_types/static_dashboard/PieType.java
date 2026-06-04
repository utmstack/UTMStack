package com.park.utmstack.domain.shared_types.static_dashboard;

import java.util.ArrayList;
import java.util.List;

public class PieType {
    private List<String> data = new ArrayList<>();
    private List<PieValue> value = new ArrayList<>();

    public List<String> getData() {
        return data;
    }

    public void setData(List<String> data) {
        this.data = data;
    }

    public List<PieValue> getValue() {
        return value;
    }

    public void setValue(List<PieValue> value) {
        this.value = value;
    }

    public PieType addData(String data) {
        this.data.add(data);
        return this;
    }

    public PieType addValue(PieValue value) {
        this.value.add(value);
        return this;
    }


}




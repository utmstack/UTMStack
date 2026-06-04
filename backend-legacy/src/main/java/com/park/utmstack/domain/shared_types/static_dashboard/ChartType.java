package com.park.utmstack.domain.shared_types.static_dashboard;

import java.util.List;

public class ChartType {
    private List<String> series;
    private List<Integer> values;

    public List<String> getSeries() {
        return series;
    }

    public void setSeries(List<String> series) {
        this.series = series;
    }

    public List<Integer> getValues() {
        return values;
    }

    public void setValues(List<Integer> values) {
        this.values = values;
    }
}

package com.park.utmstack.domain.shared_types.static_dashboard;

import java.util.ArrayList;
import java.util.List;

public class BarType {
    private List<String> categories = new ArrayList<>();
    private List<Long> series = new ArrayList<>();

    public List<String> getCategories() {
        return categories;
    }

    public void setCategories(List<String> categories) {
        this.categories = categories;
    }

    public List<Long> getSeries() {
        return series;
    }

    public void setSeries(List<Long> series) {
        this.series = series;
    }

    public BarType addCategory(String category) {
        categories.add(category);
        return this;
    }

    public BarType addSerie(Long serie) {
        series.add(serie);
        return this;
    }
}

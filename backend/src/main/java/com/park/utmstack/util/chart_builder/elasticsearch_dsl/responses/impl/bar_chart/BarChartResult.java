package com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.bar_chart;

import java.util.ArrayList;
import java.util.List;

public class BarChartResult {

    private List<String> categories;
    private List<Serie> series;

    BarChartResult() {
        categories = new ArrayList<>();
        series = new ArrayList<>();
    }

    public List<String> getCategories() {
        return categories;
    }

    public void setCategories(List<String> categories) {
        this.categories = categories;
    }

    public BarChartResult addCategory(String category) {
        if (!categories.contains(category))
            categories.add(category);
        return this;
    }

    public List<Serie> getSeries() {
        return series;
    }

    public void setSeries(List<Serie> series) {
        this.series = series;
    }

    public BarChartResult addSerie(Serie serie) {
        series.add(serie);
        return this;
    }

    public static class Serie {
        private String metricId;
        private String name;
        private List<Double> data = new ArrayList<>();

        public String getMetricId() {
            return metricId;
        }

        public void setMetricId(String metricId) {
            this.metricId = metricId;
        }

        public String getName() {
            return name;
        }

        public void setName(String name) {
            this.name = name;
        }

        public List<Double> getData() {
            return data;
        }

        public void setData(List<Double> data) {
            this.data = data;
        }

        public Serie addData(Double value) {
            data.add(value);
            return this;
        }
    }
}

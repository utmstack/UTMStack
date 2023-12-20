package com.park.utmstack.domain.shared_types;

import java.util.ArrayList;
import java.util.List;

public class EventsByObjectsInTimeType {
    private List<String> categories = new ArrayList<>();
    private List<Metadata> series = new ArrayList<>();

    public List<String> getCategories() {
        return categories;
    }

    public void setCategories(List<String> categories) {
        this.categories = categories;
    }

    public List<Metadata> getSeries() {
        return series;
    }

    public void setSeries(List<Metadata> series) {
        this.series = series;
    }

    public void addCategory(String category) {
        categories.add(category);
    }

    public void addSerie(String serie, Long value) {
        boolean existMetadata = false;
        for (Metadata metadata : series) {
            if (metadata.serie.equals(serie)) {
                metadata.addValue(value);
                existMetadata = true;
                break;
            }
        }
        if (!existMetadata) {
            Metadata meta = new Metadata(serie);
            meta.addValue(value);
            series.add(meta);
        }
    }

    public static class Metadata {
        private String serie;
        private List<Long> values = new ArrayList<>();

        public Metadata(String serie) {
            this.serie = serie;
        }

        public String getSerie() {
            return serie;
        }

        public void setSerie(String serie) {
            this.serie = serie;
        }

        public List<Long> getValues() {
            return values;
        }

        public void setValues(List<Long> values) {
            this.values = values;
        }

        public void addValue(Long val) {
            values.add(val);
        }
    }
}

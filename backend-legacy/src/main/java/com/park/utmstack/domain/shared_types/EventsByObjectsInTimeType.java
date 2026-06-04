package com.park.utmstack.domain.shared_types;

import lombok.Getter;
import lombok.Setter;

import java.util.ArrayList;
import java.util.List;

@Setter
@Getter
public class EventsByObjectsInTimeType {
    private List<String> categories = new ArrayList<>();
    private List<Metadata> series = new ArrayList<>();

    public void addCategory(String category) {
        categories.add(category);
    }

    public void addSeries(String serie, Long value) {
        boolean existMetadata = false;
        for (Metadata metadata : series) {
            if (metadata.series.equals(serie)) {
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
        private final String series;
        @Setter
        @Getter
        private List<Long> values = new ArrayList<>();

        public Metadata(String series) {
            this.series = series;
        }

        public void addValue(Long val) {
            values.add(val);
        }
    }
}

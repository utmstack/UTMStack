package com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.impl.table_chart;

import java.util.ArrayList;
import java.util.List;

public class TableChartResult {
    private List<String> columns = new ArrayList<>();
    private List<List<Cell<?>>> rows = new ArrayList<>();

    public List<String> getColumns() {
        return columns;
    }

    public void setColumns(List<String> columns) {
        this.columns = columns;
    }

    public List<List<Cell<?>>> getRows() {
        return rows;
    }

    public void setRows(List<List<Cell<?>>> rows) {
        this.rows = rows;
    }

    public TableChartResult addRow(List<Cell<?>> cells) {
        rows.add(cells);
        return this;
    }

    public TableChartResult addColumn(String column) {
        if (!columns.contains(column))
            columns.add(column);
        return this;
    }

    public static class Cell<T> {
        private T value;
        private boolean isMetric;

        public Cell() {
        }

        public Cell(T value, boolean isMetric) {
            this.value = value;
            this.isMetric = isMetric;
        }

        public T getValue() {
            return value;
        }

        public void setValue(T value) {
            this.value = value;
        }

        public Cell<T> value(T value) {
            this.value = value;
            return this;
        }

        public boolean isMetric() {
            return isMetric;
        }

        public void setMetric(boolean metric) {
            isMetric = metric;
        }

        public Cell<T> metric(boolean metric) {
            this.isMetric = metric;
            return this;
        }
    }
}

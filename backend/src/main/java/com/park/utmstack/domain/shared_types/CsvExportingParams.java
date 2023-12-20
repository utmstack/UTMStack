package com.park.utmstack.domain.shared_types;

import com.park.utmstack.domain.chart_builder.types.query.FilterType;

import javax.validation.constraints.NotNull;
import java.util.List;

public class CsvExportingParams {
    @NotNull
    private DataColumn[] columns;

    private List<FilterType> filters;

    @NotNull
    private Integer top;

    @NotNull
    private String indexPattern;

    public DataColumn[] getColumns() {
        return columns;
    }

    public void setColumns(DataColumn[] columns) {
        this.columns = columns;
    }

    public List<FilterType> getFilters() {
        return filters;
    }

    public void setFilters(List<FilterType> filters) {
        this.filters = filters;
    }

    public Integer getTop() {
        return top;
    }

    public void setTop(Integer top) {
        this.top = top;
    }

    public String getIndexPattern() {
        return indexPattern;
    }

    public void setIndexPattern(String indexPattern) {
        this.indexPattern = indexPattern;
    }
}

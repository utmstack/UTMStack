package com.park.utmstack.domain.compliance.types;

import com.park.utmstack.domain.chart_builder.types.query.FilterType;

import java.util.List;

public class IndexPatternWithFilterType {
    private String indexPattern;
    private List<FilterType> filterType;

    public IndexPatternWithFilterType(){}
    public IndexPatternWithFilterType(String indexPattern, List<FilterType> filterType) {
        this.indexPattern = indexPattern;
        this.filterType = filterType;
    }

    public String getIndexPattern() {
        return indexPattern;
    }

    public void setIndexPattern(String indexPattern) {
        this.indexPattern = indexPattern;
    }

    public List<FilterType> getFilterType() {
        return filterType;
    }

    public void setFilterType(List<FilterType> filterType) {
        this.filterType = filterType;
    }
}

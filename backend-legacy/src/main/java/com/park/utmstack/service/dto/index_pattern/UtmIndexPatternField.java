package com.park.utmstack.service.dto.index_pattern;

import com.park.utmstack.util.chart_builder.IndexPropertyType;

import java.util.List;

public class UtmIndexPatternField {
    String indexPattern;
    List<IndexPropertyType> fields;

    public UtmIndexPatternField(){};

    public UtmIndexPatternField(String indexPatter, List<IndexPropertyType> fields) {
        this.indexPattern = indexPatter;
        this.fields = fields;
    }

    public String getIndexPattern() {
        return indexPattern;
    }

    public void setIndexPattern(String indexPattern) {
        this.indexPattern = indexPattern;
    }

    public List<IndexPropertyType> getFields() {
        return fields;
    }

    public void setFields(List<IndexPropertyType> fields) {
        this.fields = fields;
    }
}

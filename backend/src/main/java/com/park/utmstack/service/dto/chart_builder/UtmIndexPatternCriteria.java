package com.park.utmstack.service.dto.chart_builder;

import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;

public class UtmIndexPatternCriteria implements Serializable {
    private LongFilter id;
    private StringFilter pattern;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public StringFilter getPattern() {
        return pattern;
    }

    public void setPattern(StringFilter pattern) {
        this.pattern = pattern;
    }
}

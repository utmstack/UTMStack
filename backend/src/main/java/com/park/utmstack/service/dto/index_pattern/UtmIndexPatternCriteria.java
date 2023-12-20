package com.park.utmstack.service.dto.index_pattern;

import tech.jhipster.service.filter.BooleanFilter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;

public class UtmIndexPatternCriteria implements Serializable {
    private LongFilter id;
    private StringFilter pattern;
    private StringFilter patternModule;
    private BooleanFilter patternSystem;
    private BooleanFilter isActive;

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

    public StringFilter getPatternModule() {
        return patternModule;
    }

    public void setPatternModule(StringFilter patternModule) {
        this.patternModule = patternModule;
    }

    public BooleanFilter getPatternSystem() {
        return patternSystem;
    }

    public void setPatternSystem(BooleanFilter patternSystem) {
        this.patternSystem = patternSystem;
    }

    public BooleanFilter getIsActive() {
        return isActive;
    }

    public void setIsActive(BooleanFilter isActive) {
        this.isActive = isActive;
    }
}

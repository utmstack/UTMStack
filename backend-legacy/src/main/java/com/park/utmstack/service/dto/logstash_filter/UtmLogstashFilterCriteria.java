package com.park.utmstack.service.dto.logstash_filter;

import tech.jhipster.service.filter.BooleanFilter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;

public class UtmLogstashFilterCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;
    private StringFilter filterName;
    private LongFilter filterGroupId;
    private BooleanFilter isActive;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public StringFilter getFilterName() {
        return filterName;
    }

    public void setFilterName(StringFilter filterName) {
        this.filterName = filterName;
    }

    public LongFilter getFilterGroupId() {
        return filterGroupId;
    }

    public void setFilterGroupId(LongFilter filterGroupId) {
        this.filterGroupId = filterGroupId;
    }

    public BooleanFilter getIsActive() {
        return isActive;
    }

    public void setIsActive(BooleanFilter isActive) {
        this.isActive = isActive;
    }
}

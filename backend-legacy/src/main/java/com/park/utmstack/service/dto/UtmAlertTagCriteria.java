package com.park.utmstack.service.dto;

import tech.jhipster.service.filter.BooleanFilter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;

public class UtmAlertTagCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;
    private StringFilter tagName;
    private BooleanFilter systemOwner;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public StringFilter getTagName() {
        return tagName;
    }

    public void setTagName(StringFilter tagName) {
        this.tagName = tagName;
    }

    public BooleanFilter getSystemOwner() {
        return systemOwner;
    }

    public void setSystemOwner(BooleanFilter systemOwner) {
        this.systemOwner = systemOwner;
    }
}

package com.park.utmstack.service.dto.chart_builder;

import tech.jhipster.service.filter.InstantFilter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;

public class UtmVisualizationCriteria implements Serializable {
    private LongFilter id;
    private StringFilter name;
    private InstantFilter createdDate;
    private InstantFilter modifiedDate;
    private StringFilter userCreated;
    private StringFilter userModified;
    private StringFilter chartType;
    private LongFilter idPattern;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public StringFilter getName() {
        return name;
    }

    public void setName(StringFilter name) {
        this.name = name;
    }

    public InstantFilter getCreatedDate() {
        return createdDate;
    }

    public void setCreatedDate(InstantFilter createdDate) {
        this.createdDate = createdDate;
    }

    public InstantFilter getModifiedDate() {
        return modifiedDate;
    }

    public void setModifiedDate(InstantFilter modifiedDate) {
        this.modifiedDate = modifiedDate;
    }

    public StringFilter getUserCreated() {
        return userCreated;
    }

    public void setUserCreated(StringFilter userCreated) {
        this.userCreated = userCreated;
    }

    public StringFilter getUserModified() {
        return userModified;
    }

    public void setUserModified(StringFilter userModified) {
        this.userModified = userModified;
    }

    public StringFilter getChartType() {
        return chartType;
    }

    public void setChartType(StringFilter chartType) {
        this.chartType = chartType;
    }

    public LongFilter getIdPattern() {
        return idPattern;
    }

    public void setIdPattern(LongFilter idPattern) {
        this.idPattern = idPattern;
    }
}

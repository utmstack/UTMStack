package com.park.utmstack.service.dto;

import tech.jhipster.service.filter.*;

import java.io.Serializable;

public class UtmRuleCriteria implements Serializable {
    private static final long serialVersionUID = 1L;

    private LongFilter id;
    private StringFilter whenGeneratedBy;
    private InstantFilter createdDate;
    private StringFilter whenRule;
    private IntegerFilter whenSeverity;
    private StringFilter whenSensor;
    private StringFilter whenOrigin;
    private StringFilter whenDestination;
    private IntegerFilter thenStatus;
    private BooleanFilter isTags;
    private StringFilter tags;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public StringFilter getWhenGeneratedBy() {
        return whenGeneratedBy;
    }

    public void setWhenGeneratedBy(StringFilter whenGeneratedBy) {
        this.whenGeneratedBy = whenGeneratedBy;
    }

    public InstantFilter getCreatedDate() {
        return createdDate;
    }

    public void setCreatedDate(InstantFilter createdDate) {
        this.createdDate = createdDate;
    }

    public StringFilter getWhenRule() {
        return whenRule;
    }

    public void setWhenRule(StringFilter whenRule) {
        this.whenRule = whenRule;
    }

    public IntegerFilter getWhenSeverity() {
        return whenSeverity;
    }

    public void setWhenSeverity(IntegerFilter whenSeverity) {
        this.whenSeverity = whenSeverity;
    }

    public StringFilter getWhenSensor() {
        return whenSensor;
    }

    public void setWhenSensor(StringFilter whenSensor) {
        this.whenSensor = whenSensor;
    }

    public StringFilter getWhenOrigin() {
        return whenOrigin;
    }

    public void setWhenOrigin(StringFilter whenOrigin) {
        this.whenOrigin = whenOrigin;
    }

    public StringFilter getWhenDestination() {
        return whenDestination;
    }

    public void setWhenDestination(StringFilter whenDestination) {
        this.whenDestination = whenDestination;
    }

    public IntegerFilter getThenStatus() {
        return thenStatus;
    }

    public void setThenStatus(IntegerFilter thenStatus) {
        this.thenStatus = thenStatus;
    }

    public BooleanFilter getIsTags() {
        return isTags;
    }

    public void setIsTags(BooleanFilter isTags) {
        this.isTags = isTags;
    }

    public StringFilter getTags() {
        return tags;
    }

    public void setTags(StringFilter tags) {
        this.tags = tags;
    }
}

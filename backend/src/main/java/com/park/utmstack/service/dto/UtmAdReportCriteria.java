package com.park.utmstack.service.dto;

import tech.jhipster.service.filter.Filter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;

/**
 * Criteria class for the UtmAdReport entity. This class is used in UtmAdReportResource to receive all the possible filtering
 * options from the Http GET request parameters. For example the following could be a valid requests:
 * <code> /utm-ad-reports?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use fix type
 * specific filters.
 */
public class UtmAdReportCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;

    private StringFilter name;

    private StringFilter description;

    private LongFilter schedule;

    private LongFilter limit;

    private StringFilter user;

    private StringFilter objects;

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

    public StringFilter getDescription() {
        return description;
    }

    public void setDescription(StringFilter description) {
        this.description = description;
    }

    public LongFilter getSchedule() {
        return schedule;
    }

    public void setSchedule(LongFilter schedule) {
        this.schedule = schedule;
    }

    public LongFilter getLimit() {
        return limit;
    }

    public void setLimit(LongFilter limit) {
        this.limit = limit;
    }

    public StringFilter getUser() {
        return user;
    }

    public void setUser(StringFilter user) {
        this.user = user;
    }

    public StringFilter getObjects() {
        return objects;
    }

    public void setObjects(StringFilter objects) {
        this.objects = objects;
    }
}

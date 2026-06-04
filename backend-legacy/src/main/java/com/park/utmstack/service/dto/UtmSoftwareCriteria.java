package com.park.utmstack.service.dto;

import tech.jhipster.service.filter.Filter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;

/**
 * Criteria class for the UtmSoftware entity. This class is used in UtmSoftwareResource to receive all the possible filtering
 * options from the Http GET request parameters. For example the following could be a valid requests:
 * <code> /utm-softwares?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use fix type
 * specific filters.
 */
public class UtmSoftwareCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;
    private StringFilter softName;
    private StringFilter softDeveloper;
    private StringFilter softVersion;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public StringFilter getSoftName() {
        return softName;
    }

    public void setSoftName(StringFilter softName) {
        this.softName = softName;
    }

    public StringFilter getSoftDeveloper() {
        return softDeveloper;
    }

    public void setSoftDeveloper(StringFilter softDeveloper) {
        this.softDeveloper = softDeveloper;
    }

    public StringFilter getSoftVersion() {
        return softVersion;
    }

    public void setSoftVersion(StringFilter softVersion) {
        this.softVersion = softVersion;
    }
}

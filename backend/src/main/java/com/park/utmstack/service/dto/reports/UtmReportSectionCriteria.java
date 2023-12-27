package com.park.utmstack.service.dto.reports;

import tech.jhipster.service.filter.*;

import java.io.Serializable;

/**
 * Criteria class for the UtmReportSection entity. This class is used in UtmReportSectionResource to
 * receive all the possible filtering options from the Http GET request parameters.
 * For example the following could be a valid requests:
 * <code> /utm-report-sections?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use
 * fix type specific filters.
 */
public class UtmReportSectionCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;

    private StringFilter repSecName;

    private StringFilter repSecDescription;

    private StringFilter creationUser;

    private InstantFilter creationDate;

    private StringFilter modificationUser;

    private InstantFilter modificationDate;

    private BooleanFilter repSecSystem;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public StringFilter getRepSecName() {
        return repSecName;
    }

    public void setRepSecName(StringFilter repSecName) {
        this.repSecName = repSecName;
    }

    public StringFilter getRepSecDescription() {
        return repSecDescription;
    }

    public void setRepSecDescription(StringFilter repSecDescription) {
        this.repSecDescription = repSecDescription;
    }

    public StringFilter getCreationUser() {
        return creationUser;
    }

    public void setCreationUser(StringFilter creationUser) {
        this.creationUser = creationUser;
    }

    public InstantFilter getCreationDate() {
        return creationDate;
    }

    public void setCreationDate(InstantFilter creationDate) {
        this.creationDate = creationDate;
    }

    public StringFilter getModificationUser() {
        return modificationUser;
    }

    public void setModificationUser(StringFilter modificationUser) {
        this.modificationUser = modificationUser;
    }

    public InstantFilter getModificationDate() {
        return modificationDate;
    }

    public void setModificationDate(InstantFilter modificationDate) {
        this.modificationDate = modificationDate;
    }

    public BooleanFilter getRepSecSystem() {
        return repSecSystem;
    }

    public void setRepSecSystem(BooleanFilter repSecSystem) {
        this.repSecSystem = repSecSystem;
    }
}

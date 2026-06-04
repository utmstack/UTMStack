package com.park.utmstack.service.dto;

import tech.jhipster.service.filter.*;

import java.io.Serializable;

/**
 * Criteria class for the UtmConfigurationParameter entity. This class is used in UtmConfigurationParameterResource to
 * receive all the possible filtering options from the Http GET request parameters.
 * For example the following could be a valid requests:
 * <code> /utm-configuration-parameters?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use
 * fix type specific filters.
 */
public class UtmConfigurationParameterCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;

    private LongFilter sectionId;

    private StringFilter confParamShort;

    private StringFilter confParamLarge;

    private StringFilter confParamDescription;

    private StringFilter confParamValue;

    private StringFilter confParamDatatype;

    private BooleanFilter confParamRequired;

    private InstantFilter modificationTime;

    private StringFilter modificationUser;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public LongFilter getSectionId() {
        return sectionId;
    }

    public void setSectionId(LongFilter sectionId) {
        this.sectionId = sectionId;
    }

    public StringFilter getConfParamShort() {
        return confParamShort;
    }

    public void setConfParamShort(StringFilter confParamShort) {
        this.confParamShort = confParamShort;
    }

    public StringFilter getConfParamLarge() {
        return confParamLarge;
    }

    public void setConfParamLarge(StringFilter confParamLarge) {
        this.confParamLarge = confParamLarge;
    }

    public StringFilter getConfParamDescription() {
        return confParamDescription;
    }

    public void setConfParamDescription(StringFilter confParamDescription) {
        this.confParamDescription = confParamDescription;
    }

    public StringFilter getConfParamValue() {
        return confParamValue;
    }

    public void setConfParamValue(StringFilter confParamValue) {
        this.confParamValue = confParamValue;
    }

    public BooleanFilter getConfParamRequired() {
        return confParamRequired;
    }

    public void setConfParamRequired(BooleanFilter confParamRequired) {
        this.confParamRequired = confParamRequired;
    }

    public InstantFilter getModificationTime() {
        return modificationTime;
    }

    public void setModificationTime(InstantFilter modificationTime) {
        this.modificationTime = modificationTime;
    }

    public StringFilter getModificationUser() {
        return modificationUser;
    }

    public void setModificationUser(StringFilter modificationUser) {
        this.modificationUser = modificationUser;
    }

    public StringFilter getConfParamDatatype() {
        return confParamDatatype;
    }

    public void setConfParamDatatype(StringFilter confParamDatatype) {
        this.confParamDatatype = confParamDatatype;
    }
}

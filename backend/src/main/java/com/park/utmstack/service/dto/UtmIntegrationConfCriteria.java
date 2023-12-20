package com.park.utmstack.service.dto;

import tech.jhipster.service.filter.Filter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;
import java.util.Objects;

/**
 * Criteria class for the UtmIntegrationConf entity. This class is used in UtmIntegrationConfResource to
 * receive all the possible filtering options from the Http GET request parameters.
 * For example the following could be a valid requests:
 * <code> /utm-integration-confs?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use
 * fix type specific filters.
 */
public class UtmIntegrationConfCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;

    private LongFilter integrationId;

    private StringFilter confShort;

    private StringFilter confLarge;

    private StringFilter confDescription;

    private StringFilter confValue;

    private StringFilter confDatatype;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public LongFilter getIntegrationId() {
        return integrationId;
    }

    public void setIntegrationId(LongFilter integrationId) {
        this.integrationId = integrationId;
    }

    public StringFilter getConfShort() {
        return confShort;
    }

    public void setConfShort(StringFilter confShort) {
        this.confShort = confShort;
    }

    public StringFilter getConfLarge() {
        return confLarge;
    }

    public void setConfLarge(StringFilter confLarge) {
        this.confLarge = confLarge;
    }

    public StringFilter getConfDescription() {
        return confDescription;
    }

    public void setConfDescription(StringFilter confDescription) {
        this.confDescription = confDescription;
    }

    public StringFilter getConfValue() {
        return confValue;
    }

    public void setConfValue(StringFilter confValue) {
        this.confValue = confValue;
    }

    public StringFilter getConfDatatype() {
        return confDatatype;
    }

    public void setConfDatatype(StringFilter confDatatype) {
        this.confDatatype = confDatatype;
    }


    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        final UtmIntegrationConfCriteria that = (UtmIntegrationConfCriteria) o;
        return
            Objects.equals(id, that.id) &&
            Objects.equals(integrationId, that.integrationId) &&
            Objects.equals(confShort, that.confShort) &&
            Objects.equals(confLarge, that.confLarge) &&
            Objects.equals(confDescription, that.confDescription) &&
            Objects.equals(confValue, that.confValue) &&
            Objects.equals(confDatatype, that.confDatatype);
    }

    @Override
    public int hashCode() {
        return Objects.hash(
        id,
        integrationId,
        confShort,
        confLarge,
        confDescription,
        confValue,
        confDatatype
        );
    }

    @Override
    public String toString() {
        return "UtmIntegrationConfCriteria{" +
                (id != null ? "id=" + id + ", " : "") +
                (integrationId != null ? "integrationId=" + integrationId + ", " : "") +
                (confShort != null ? "confShort=" + confShort + ", " : "") +
                (confLarge != null ? "confLarge=" + confLarge + ", " : "") +
                (confDescription != null ? "confDescription=" + confDescription + ", " : "") +
                (confValue != null ? "confValue=" + confValue + ", " : "") +
                (confDatatype != null ? "confDatatype=" + confDatatype + ", " : "") +
            "}";
    }

}

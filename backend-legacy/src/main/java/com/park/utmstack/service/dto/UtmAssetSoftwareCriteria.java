package com.park.utmstack.service.dto;

import tech.jhipster.service.filter.Filter;
import tech.jhipster.service.filter.LongFilter;

import java.io.Serializable;
import java.util.Objects;

/**
 * Criteria class for the UtmAssetSoftware entity. This class is used in UtmAssetSoftwareResource to
 * receive all the possible filtering options from the Http GET request parameters.
 * For example the following could be a valid requests:
 * <code> /utm-asset-softwares?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use
 * fix type specific filters.
 */
public class UtmAssetSoftwareCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;

    private LongFilter softwareId;

    private LongFilter assetId;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public LongFilter getSoftwareId() {
        return softwareId;
    }

    public void setSoftwareId(LongFilter softwareId) {
        this.softwareId = softwareId;
    }

    public LongFilter getAssetId() {
        return assetId;
    }

    public void setAssetId(LongFilter assetId) {
        this.assetId = assetId;
    }


    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        final UtmAssetSoftwareCriteria that = (UtmAssetSoftwareCriteria) o;
        return
            Objects.equals(id, that.id) &&
            Objects.equals(softwareId, that.softwareId) &&
            Objects.equals(assetId, that.assetId);
    }

    @Override
    public int hashCode() {
        return Objects.hash(
        id,
        softwareId,
        assetId
        );
    }

    @Override
    public String toString() {
        return "UtmAssetSoftwareCriteria{" +
                (id != null ? "id=" + id + ", " : "") +
                (softwareId != null ? "softwareId=" + softwareId + ", " : "") +
                (assetId != null ? "assetId=" + assetId + ", " : "") +
            "}";
    }

}

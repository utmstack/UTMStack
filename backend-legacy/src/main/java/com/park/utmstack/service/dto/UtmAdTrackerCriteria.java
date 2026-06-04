package com.park.utmstack.service.dto;

import tech.jhipster.service.filter.Filter;
import tech.jhipster.service.filter.InstantFilter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;
import java.util.Objects;

/**
 * Criteria class for the UtmAdTracker entity. This class is used in UtmAdTrackerResource to
 * receive all the possible filtering options from the Http GET request parameters.
 * For example the following could be a valid requests:
 * <code> /utm-ad-trackers?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use
 * fix type specific filters.
 */
public class UtmAdTrackerCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;

    private StringFilter objectId;

    private StringFilter objectName;

    private StringFilter objectType;

    private InstantFilter whenTracked;

    private StringFilter userTracker;

    private InstantFilter lastEventDate;

    private LongFilter changesAmount;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public StringFilter getObjectId() {
        return objectId;
    }

    public void setObjectId(StringFilter objectId) {
        this.objectId = objectId;
    }

    public StringFilter getObjectName() {
        return objectName;
    }

    public void setObjectName(StringFilter objectName) {
        this.objectName = objectName;
    }

    public StringFilter getObjectType() {
        return objectType;
    }

    public void setObjectType(StringFilter objectType) {
        this.objectType = objectType;
    }

    public InstantFilter getWhenTracked() {
        return whenTracked;
    }

    public void setWhenTracked(InstantFilter whenTracked) {
        this.whenTracked = whenTracked;
    }

    public StringFilter getUserTracker() {
        return userTracker;
    }

    public void setUserTracker(StringFilter userTracker) {
        this.userTracker = userTracker;
    }

    public InstantFilter getLastEventDate() {
        return lastEventDate;
    }

    public void setLastEventDate(InstantFilter lastEventDate) {
        this.lastEventDate = lastEventDate;
    }

    public LongFilter getChangesAmount() {
        return changesAmount;
    }

    public void setChangesAmount(LongFilter changesAmount) {
        this.changesAmount = changesAmount;
    }


    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        final UtmAdTrackerCriteria that = (UtmAdTrackerCriteria) o;
        return
            Objects.equals(id, that.id) &&
            Objects.equals(objectId, that.objectId) &&
            Objects.equals(objectName, that.objectName) &&
            Objects.equals(objectType, that.objectType) &&
            Objects.equals(whenTracked, that.whenTracked) &&
            Objects.equals(userTracker, that.userTracker) &&
            Objects.equals(lastEventDate, that.lastEventDate) &&
            Objects.equals(changesAmount, that.changesAmount);
    }

    @Override
    public int hashCode() {
        return Objects.hash(
        id,
        objectId,
        objectName,
        objectType,
        whenTracked,
        userTracker,
        lastEventDate,
        changesAmount
        );
    }

    @Override
    public String toString() {
        return "UtmAdTrackerCriteria{" +
                (id != null ? "id=" + id + ", " : "") +
                (objectId != null ? "objectId=" + objectId + ", " : "") +
                (objectName != null ? "objectName=" + objectName + ", " : "") +
                (objectType != null ? "objectType=" + objectType + ", " : "") +
                (whenTracked != null ? "whenTracked=" + whenTracked + ", " : "") +
                (userTracker != null ? "userTracker=" + userTracker + ", " : "") +
                (lastEventDate != null ? "lastEventDate=" + lastEventDate + ", " : "") +
                (changesAmount != null ? "changesAmount=" + changesAmount + ", " : "") +
            "}";
    }

}

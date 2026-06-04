package com.park.utmstack.service.dto.incident;

import com.park.utmstack.domain.incident.enums.IncidentStatusEnum;
import tech.jhipster.service.filter.*;

import java.io.Serializable;
import java.util.Objects;

/**
 * Criteria class for the UtmIncident entity. This class is used in UtmIncidentResource to
 * receive all the possible filtering options from the Http GET request parameters.
 * For example the following could be a valid requests:
 * <code> /utm-incidents?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use
 * fix type specific filters.
 */
public class UtmIncidentCriteria implements Serializable {
    /**
     * Class for filtering IncidentStatusEnum
     */
    public static class IncidentStatusEnumFilter extends Filter<IncidentStatusEnum> {
    }

    private static final long serialVersionUID = 1L;

    private LongFilter id;

    private StringFilter incidentName;

    private StringFilter incidentDescription;

    private IncidentStatusEnumFilter incidentStatus;

    private StringFilter incidentAssignedTo;

    private InstantFilter incidentCreatedDate;

    private IntegerFilter incidentSeverity;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public StringFilter getIncidentName() {
        return incidentName;
    }

    public void setIncidentName(StringFilter incidentName) {
        this.incidentName = incidentName;
    }

    public StringFilter getIncidentDescription() {
        return incidentDescription;
    }

    public void setIncidentDescription(StringFilter incidentDescription) {
        this.incidentDescription = incidentDescription;
    }

    public IncidentStatusEnumFilter getIncidentStatus() {
        return incidentStatus;
    }

    public void setIncidentStatus(IncidentStatusEnumFilter incidentStatus) {
        this.incidentStatus = incidentStatus;
    }

    public StringFilter getIncidentAssignedTo() {
        return incidentAssignedTo;
    }

    public void setIncidentAssignedTo(StringFilter incidentAssignedTo) {
        this.incidentAssignedTo = incidentAssignedTo;
    }

    public InstantFilter getIncidentCreatedDate() {
        return incidentCreatedDate;
    }

    public void setIncidentCreatedDate(InstantFilter incidentCreatedDate) {
        this.incidentCreatedDate = incidentCreatedDate;
    }

    public IntegerFilter getIncidentSeverity() {
        return incidentSeverity;
    }

    public void setIncidentSeverity(IntegerFilter incidentSeverity) {
        this.incidentSeverity = incidentSeverity;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        final UtmIncidentCriteria that = (UtmIncidentCriteria) o;
        return
            Objects.equals(id, that.id) &&
                Objects.equals(incidentName, that.incidentName) &&
                Objects.equals(incidentDescription, that.incidentDescription) &&
                Objects.equals(incidentStatus, that.incidentStatus) &&
                Objects.equals(incidentAssignedTo, that.incidentAssignedTo) &&
                Objects.equals(incidentSeverity, that.incidentSeverity) &&
                Objects.equals(incidentCreatedDate, that.incidentCreatedDate);
    }

    @Override
    public int hashCode() {
        return Objects.hash(
            id,
            incidentName,
            incidentDescription,
            incidentStatus,
            incidentAssignedTo,
            incidentCreatedDate,
            incidentSeverity
        );
    }

    @Override
    public String toString() {
        return "UtmIncidentCriteria{" +
            (id != null ? "id=" + id + ", " : "") +
            (incidentName != null ? "incidentName=" + incidentName + ", " : "") +
            (incidentDescription != null ? "incidentDescription=" + incidentDescription + ", " : "") +
            (incidentStatus != null ? "incidentStatus=" + incidentStatus + ", " : "") +
            (incidentAssignedTo != null ? "incidentAssignedTo=" + incidentAssignedTo + ", " : "") +
            (incidentCreatedDate != null ? "incidentCreatedDate=" + incidentCreatedDate + ", " : "") +
            "}";
    }

}

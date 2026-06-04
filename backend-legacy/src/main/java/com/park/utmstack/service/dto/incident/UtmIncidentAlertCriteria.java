package com.park.utmstack.service.dto.incident;

import tech.jhipster.service.filter.Filter;
import tech.jhipster.service.filter.IntegerFilter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;
import java.util.Objects;

/**
 * Criteria class for the UtmIncidentAlert entity. This class is used in UtmIncidentAlertResource to
 * receive all the possible filtering options from the Http GET request parameters.
 * For example the following could be a valid requests:
 * <code> /utm-incident-alerts?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use
 * fix type specific filters.
 */
public class UtmIncidentAlertCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;

    private LongFilter incidentId;

    private StringFilter alertId;

    private StringFilter alertName;

    private IntegerFilter alertStatus;

    private IntegerFilter alertSeverity;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public LongFilter getIncidentId() {
        return incidentId;
    }

    public void setIncidentId(LongFilter incidentId) {
        this.incidentId = incidentId;
    }

    public StringFilter getAlertId() {
        return alertId;
    }

    public void setAlertId(StringFilter alertId) {
        this.alertId = alertId;
    }

    public StringFilter getAlertName() {
        return alertName;
    }

    public void setAlertName(StringFilter alertName) {
        this.alertName = alertName;
    }

    public IntegerFilter getAlertStatus() {
        return alertStatus;
    }

    public void setAlertStatus(IntegerFilter alertStatus) {
        this.alertStatus = alertStatus;
    }

    public IntegerFilter getAlertSeverity() {
        return alertSeverity;
    }

    public void setAlertSeverity(IntegerFilter alertSeverity) {
        this.alertSeverity = alertSeverity;
    }


    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        final UtmIncidentAlertCriteria that = (UtmIncidentAlertCriteria) o;
        return
            Objects.equals(id, that.id) &&
            Objects.equals(incidentId, that.incidentId) &&
            Objects.equals(alertId, that.alertId) &&
            Objects.equals(alertName, that.alertName) &&
            Objects.equals(alertStatus, that.alertStatus) &&
            Objects.equals(alertSeverity, that.alertSeverity);
    }

    @Override
    public int hashCode() {
        return Objects.hash(
        id,
        incidentId,
        alertId,
        alertName,
        alertStatus,
        alertSeverity
        );
    }

    @Override
    public String toString() {
        return "UtmIncidentAlertCriteria{" +
                (id != null ? "id=" + id + ", " : "") +
                (incidentId != null ? "incidentId=" + incidentId + ", " : "") +
                (alertId != null ? "alertId=" + alertId + ", " : "") +
                (alertName != null ? "alertName=" + alertName + ", " : "") +
                (alertStatus != null ? "alertStatus=" + alertStatus + ", " : "") +
                (alertSeverity != null ? "alertSeverity=" + alertSeverity + ", " : "") +
            "}";
    }

}

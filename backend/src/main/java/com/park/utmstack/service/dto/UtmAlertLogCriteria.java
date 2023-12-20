package com.park.utmstack.service.dto;

import tech.jhipster.service.filter.Filter;
import tech.jhipster.service.filter.InstantFilter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;
import java.util.Objects;

/**
 * Criteria class for the UtmAlertLog entity. This class is used in UtmAlertLogResource to
 * receive all the possible filtering options from the Http GET request parameters.
 * For example the following could be a valid requests:
 * <code> /utm-alert-logs?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use
 * fix type specific filters.
 */
public class UtmAlertLogCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;

    private StringFilter alertId;

    private StringFilter logUser;

    private InstantFilter logDate;

    private StringFilter logAction;

    private StringFilter logOldValue;

    private StringFilter logNewValue;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public StringFilter getAlertId() {
        return alertId;
    }

    public void setAlertId(StringFilter alertId) {
        this.alertId = alertId;
    }

    public StringFilter getLogUser() {
        return logUser;
    }

    public void setLogUser(StringFilter logUser) {
        this.logUser = logUser;
    }

    public InstantFilter getLogDate() {
        return logDate;
    }

    public void setLogDate(InstantFilter logDate) {
        this.logDate = logDate;
    }

    public StringFilter getLogAction() {
        return logAction;
    }

    public void setLogAction(StringFilter logAction) {
        this.logAction = logAction;
    }

    public StringFilter getLogOldValue() {
        return logOldValue;
    }

    public void setLogOldValue(StringFilter logOldValue) {
        this.logOldValue = logOldValue;
    }

    public StringFilter getLogNewValue() {
        return logNewValue;
    }

    public void setLogNewValue(StringFilter logNewValue) {
        this.logNewValue = logNewValue;
    }


    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        final UtmAlertLogCriteria that = (UtmAlertLogCriteria) o;
        return
            Objects.equals(id, that.id) &&
            Objects.equals(alertId, that.alertId) &&
            Objects.equals(logUser, that.logUser) &&
            Objects.equals(logDate, that.logDate) &&
            Objects.equals(logAction, that.logAction) &&
            Objects.equals(logOldValue, that.logOldValue) &&
            Objects.equals(logNewValue, that.logNewValue);
    }

    @Override
    public int hashCode() {
        return Objects.hash(
        id,
        alertId,
        logUser,
        logDate,
        logAction,
        logOldValue,
        logNewValue
        );
    }

    @Override
    public String toString() {
        return "UtmAlertLogCriteria{" +
                (id != null ? "id=" + id + ", " : "") +
                (alertId != null ? "alertId=" + alertId + ", " : "") +
                (logUser != null ? "logUser=" + logUser + ", " : "") +
                (logDate != null ? "logDate=" + logDate + ", " : "") +
                (logAction != null ? "logAction=" + logAction + ", " : "") +
                (logOldValue != null ? "logOldValue=" + logOldValue + ", " : "") +
                (logNewValue != null ? "logNewValue=" + logNewValue + ", " : "") +
            "}";
    }

}

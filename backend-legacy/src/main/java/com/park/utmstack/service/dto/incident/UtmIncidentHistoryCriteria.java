package com.park.utmstack.service.dto.incident;

import com.park.utmstack.domain.incident.enums.IncidentHistoryActionEnum;
import tech.jhipster.service.filter.Filter;
import tech.jhipster.service.filter.InstantFilter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;
import java.util.Objects;

/**
 * Criteria class for the UtmIncidentHistory entity. This class is used in UtmIncidentHistoryResource to
 * receive all the possible filtering options from the Http GET request parameters.
 * For example the following could be a valid requests:
 * <code> /utm-incident-histories?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use
 * fix type specific filters.
 */
public class UtmIncidentHistoryCriteria implements Serializable {
    /**
     * Class for filtering INCIDENT_CREATED
     */
    public static class IncidentHistoryActionFilter extends Filter<IncidentHistoryActionEnum> {
    }

    private static final long serialVersionUID = 1L;

    private LongFilter id;

    private LongFilter incidentId;

    private InstantFilter actionDate;

    private IncidentHistoryActionFilter actionType;

    private StringFilter actionCreatedBy;

    private StringFilter action;

    private StringFilter actionDetail;

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

    public InstantFilter getActionDate() {
        return actionDate;
    }

    public void setActionDate(InstantFilter actionDate) {
        this.actionDate = actionDate;
    }

    public IncidentHistoryActionFilter getActionType() {
        return actionType;
    }

    public void setActionType(IncidentHistoryActionFilter actionType) {
        this.actionType = actionType;
    }

    public StringFilter getActionCreatedBy() {
        return actionCreatedBy;
    }

    public void setActionCreatedBy(StringFilter actionCreatedBy) {
        this.actionCreatedBy = actionCreatedBy;
    }

    public StringFilter getAction() {
        return action;
    }

    public void setAction(StringFilter action) {
        this.action = action;
    }

    public StringFilter getActionDetail() {
        return actionDetail;
    }

    public void setActionDetail(StringFilter actionDetail) {
        this.actionDetail = actionDetail;
    }


    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        final UtmIncidentHistoryCriteria that = (UtmIncidentHistoryCriteria) o;
        return
            Objects.equals(id, that.id) &&
                Objects.equals(incidentId, that.incidentId) &&
                Objects.equals(actionDate, that.actionDate) &&
                Objects.equals(actionType, that.actionType) &&
                Objects.equals(actionCreatedBy, that.actionCreatedBy) &&
                Objects.equals(action, that.action) &&
                Objects.equals(actionDetail, that.actionDetail);
    }

    @Override
    public int hashCode() {
        return Objects.hash(
            id,
            incidentId,
            actionDate,
            actionType,
            actionCreatedBy,
            action,
            actionDetail
        );
    }

    @Override
    public String toString() {
        return "UtmIncidentHistoryCriteria{" +
            (id != null ? "id=" + id + ", " : "") +
            (incidentId != null ? "incidentId=" + incidentId + ", " : "") +
            (actionDate != null ? "actionDate=" + actionDate + ", " : "") +
            (actionType != null ? "actionType=" + actionType + ", " : "") +
            (actionCreatedBy != null ? "actionCreatedBy=" + actionCreatedBy + ", " : "") +
            (action != null ? "action=" + action + ", " : "") +
            (actionDetail != null ? "actionDetail=" + actionDetail + ", " : "") +
            "}";
    }

}

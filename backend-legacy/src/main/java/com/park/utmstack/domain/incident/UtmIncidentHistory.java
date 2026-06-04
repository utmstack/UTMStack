package com.park.utmstack.domain.incident;


import com.park.utmstack.domain.incident.enums.IncidentHistoryActionEnum;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import java.io.Serializable;
import java.time.Instant;
import java.util.Objects;

/**
 * A UtmIncidentHistory.
 */
@Entity
@Table(name = "utm_incident_history")
public class UtmIncidentHistory implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @NotNull
    @Column(name = "incident_id", nullable = false)
    private Long incidentId;

    @NotNull
    @Column(name = "action_date", nullable = false)
    private Instant actionDate;

    @NotNull
    @Enumerated(EnumType.STRING)
    @Column(name = "action_type", nullable = false)
    private IncidentHistoryActionEnum actionType;

    @NotNull
    @Column(name = "action_created_by", nullable = false)
    private String actionCreatedBy;

    @NotNull
    @Column(name = "action", nullable = false)
    private String action;

    @Column(name = "action_detail")
    private String actionDetail;

    // jhipster-needle-entity-add-field - JHipster will add fields here, do not remove
    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getIncidentId() {
        return incidentId;
    }

    public UtmIncidentHistory incidentId(Long incidentId) {
        this.incidentId = incidentId;
        return this;
    }

    public void setIncidentId(Long incidentId) {
        this.incidentId = incidentId;
    }

    public Instant getActionDate() {
        return actionDate;
    }

    public UtmIncidentHistory actionDate(Instant actionDate) {
        this.actionDate = actionDate;
        return this;
    }

    public void setActionDate(Instant actionDate) {
        this.actionDate = actionDate;
    }

    public IncidentHistoryActionEnum getActionType() {
        return actionType;
    }

    public UtmIncidentHistory actionType(IncidentHistoryActionEnum actionType) {
        this.actionType = actionType;
        return this;
    }

    public void setActionType(IncidentHistoryActionEnum actionType) {
        this.actionType = actionType;
    }

    public String getActionCreatedBy() {
        return actionCreatedBy;
    }

    public UtmIncidentHistory actionCreatedBy(String actionCreatedBy) {
        this.actionCreatedBy = actionCreatedBy;
        return this;
    }

    public void setActionCreatedBy(String actionCreatedBy) {
        this.actionCreatedBy = actionCreatedBy;
    }

    public String getAction() {
        return action;
    }

    public UtmIncidentHistory action(String action) {
        this.action = action;
        return this;
    }

    public void setAction(String action) {
        this.action = action;
    }

    public String getActionDetail() {
        return actionDetail;
    }

    public UtmIncidentHistory actionDetail(String actionDetail) {
        this.actionDetail = actionDetail;
        return this;
    }

    public void setActionDetail(String actionDetail) {
        this.actionDetail = actionDetail;
    }
    // jhipster-needle-entity-add-getters-setters - JHipster will add getters and setters here, do not remove

    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        UtmIncidentHistory utmIncidentHistory = (UtmIncidentHistory) o;
        if (utmIncidentHistory.getId() == null || getId() == null) {
            return false;
        }
        return Objects.equals(getId(), utmIncidentHistory.getId());
    }

    @Override
    public int hashCode() {
        return Objects.hashCode(getId());
    }

    @Override
    public String toString() {
        return "UtmIncidentHistory{" +
            "id=" + getId() +
            ", incidentId=" + getIncidentId() +
            ", actionDate='" + getActionDate() + "'" +
            ", actionType='" + getActionType() + "'" +
            ", actionCreatedBy='" + getActionCreatedBy() + "'" +
            ", action='" + getAction() + "'" +
            ", actionDetail='" + getActionDetail() + "'" +
            "}";
    }
}

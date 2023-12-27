package com.park.utmstack.domain.incident;


import javax.persistence.*;
import javax.validation.constraints.NotNull;
import java.io.Serializable;
import java.util.Objects;

/**
 * A UtmIncidentAlert.
 */
@Entity
@Table(name = "utm_incident_alert")
public class UtmIncidentAlert implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @NotNull
    @Column(name = "incident_id", nullable = false)
    private Long incidentId;

    @NotNull
    @Column(name = "alert_id", nullable = false, unique = true)
    private String alertId;

    @NotNull
    @Column(name = "alert_name", nullable = false)
    private String alertName;

    @NotNull
    @Column(name = "alert_status", nullable = false)
    private Integer alertStatus;

    @NotNull
    @Column(name = "alert_severity", nullable = false)
    private Integer alertSeverity;

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

    public UtmIncidentAlert incidentId(Long incidentId) {
        this.incidentId = incidentId;
        return this;
    }

    public void setIncidentId(Long incidentId) {
        this.incidentId = incidentId;
    }

    public String getAlertId() {
        return alertId;
    }

    public UtmIncidentAlert alertId(String alertId) {
        this.alertId = alertId;
        return this;
    }

    public void setAlertId(String alertId) {
        this.alertId = alertId;
    }

    public String getAlertName() {
        return alertName;
    }

    public UtmIncidentAlert alertName(String alertName) {
        this.alertName = alertName;
        return this;
    }

    public void setAlertName(String alertName) {
        this.alertName = alertName;
    }

    public Integer getAlertStatus() {
        return alertStatus;
    }

    public UtmIncidentAlert alertStatus(Integer alertStatus) {
        this.alertStatus = alertStatus;
        return this;
    }

    public void setAlertStatus(Integer alertStatus) {
        this.alertStatus = alertStatus;
    }

    public Integer getAlertSeverity() {
        return alertSeverity;
    }

    public UtmIncidentAlert alertSeverity(Integer alertSeverity) {
        this.alertSeverity = alertSeverity;
        return this;
    }

    public void setAlertSeverity(Integer alertSeverity) {
        this.alertSeverity = alertSeverity;
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
        UtmIncidentAlert utmIncidentAlert = (UtmIncidentAlert) o;
        if (utmIncidentAlert.getId() == null || getId() == null) {
            return false;
        }
        return Objects.equals(getId(), utmIncidentAlert.getId());
    }

    @Override
    public int hashCode() {
        return Objects.hashCode(getId());
    }

    @Override
    public String toString() {
        return "UtmIncidentAlert{" +
            "id=" + getId() +
            ", incidentId=" + getIncidentId() +
            ", alertId='" + getAlertId() + "'" +
            ", alertName='" + getAlertName() + "'" +
            ", alertStatus=" + getAlertStatus() +
            ", alertSeverity=" + getAlertSeverity() +
            "}";
    }
}

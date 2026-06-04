package com.park.utmstack.domain.incident;


import com.park.utmstack.domain.incident.enums.IncidentStatusEnum;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.time.Instant;
import java.util.Objects;

/**
 * A UtmIncident.
 */
@Entity
@Table(name = "utm_incident")
public class UtmIncident implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @NotNull
    @Size(max = 250)
    @Column(name = "incident_name", length = 250, nullable = false, unique = true)
    private String incidentName;

    @NotNull
    @Size(max = 2000)
    @Column(name = "incident_description", length = 2000, nullable = false)
    private String incidentDescription;

    @NotNull
    @Enumerated(EnumType.STRING)
    @Column(name = "incident_status", nullable = false)
    private IncidentStatusEnum incidentStatus;

    @Column(name = "incident_assigned_to")
    private String incidentAssignedTo;

    @NotNull
    @Column(name = "incident_severity", nullable = false)
    private Integer incidentSeverity;
    @NotNull
    @Column(name = "incident_created_date", nullable = false)
    private Instant incidentCreatedDate;

    @Column(name = "incident_solution", nullable = true)
    private String incidentSolution;

    // jhipster-needle-entity-add-field - JHipster will add fields here, do not remove
    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getIncidentName() {
        return incidentName;
    }

    public UtmIncident incidentName(String incidentName) {
        this.incidentName = incidentName;
        return this;
    }

    public void setIncidentName(String incidentName) {
        this.incidentName = incidentName;
    }

    public String getIncidentDescription() {
        return incidentDescription;
    }

    public UtmIncident incidentDescription(String incidentDescription) {
        this.incidentDescription = incidentDescription;
        return this;
    }

    public void setIncidentDescription(String incidentDescription) {
        this.incidentDescription = incidentDescription;
    }

    public IncidentStatusEnum getIncidentStatus() {
        return incidentStatus;
    }

    public UtmIncident incidentStatus(IncidentStatusEnum incidentStatus) {
        this.incidentStatus = incidentStatus;
        return this;
    }

    public void setIncidentStatus(IncidentStatusEnum incidentStatus) {
        this.incidentStatus = incidentStatus;
    }

    public String getIncidentAssignedTo() {
        return incidentAssignedTo;
    }

    public UtmIncident incidentAssignedTo(String incidentAssignedTo) {
        this.incidentAssignedTo = incidentAssignedTo;
        return this;
    }

    public void setIncidentAssignedTo(String incidentAssignedTo) {
        this.incidentAssignedTo = incidentAssignedTo;
    }

    public Instant getIncidentCreatedDate() {
        return incidentCreatedDate;
    }

    public UtmIncident incidentCreatedDate(Instant incidentCreatedDate) {
        this.incidentCreatedDate = incidentCreatedDate;
        return this;
    }

    public void setIncidentCreatedDate(Instant incidentCreatedDate) {
        this.incidentCreatedDate = incidentCreatedDate;
    }

    public Integer getIncidentSeverity() {
        return incidentSeverity;
    }

    public void setIncidentSeverity(Integer incidentSeverity) {
        this.incidentSeverity = incidentSeverity;
    }

    public String getIncidentSolution() {
        return incidentSolution;
    }

    public void setIncidentSolution(String incidentSolution) {
        this.incidentSolution = incidentSolution;
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
        UtmIncident utmIncident = (UtmIncident) o;
        if (utmIncident.getId() == null || getId() == null) {
            return false;
        }
        return Objects.equals(getId(), utmIncident.getId());
    }

    @Override
    public int hashCode() {
        return Objects.hashCode(getId());
    }

    @Override
    public String toString() {
        return "UtmIncident{" +
            "id=" + getId() +
            ", incidentName='" + getIncidentName() + "'" +
            ", incidentDescription='" + getIncidentDescription() + "'" +
            ", incidentStatus='" + getIncidentStatus() + "'" +
            ", incidentAssignedTo='" + getIncidentAssignedTo() + "'" +
            ", incidentCreatedDate='" + getIncidentCreatedDate() + "'" +
            "}";
    }
}

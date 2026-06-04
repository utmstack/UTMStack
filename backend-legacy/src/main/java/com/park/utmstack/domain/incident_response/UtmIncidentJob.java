package com.park.utmstack.domain.incident_response;


import com.park.utmstack.domain.incident_response.enums.IncidentJobStatus;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import java.io.Serializable;
import java.time.Instant;

/**
 * A UtmIncidentJob.
 */
@Entity
@Table(name = "utm_incident_jobs")
public class UtmIncidentJob implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "action_id")
    private Long actionId;

    @Column(name = "params")
    private String params;

    @Column(name = "agent")
    private String agent;

    @Column(name = "status")
    private Integer status;

    @Column(name = "job_result")
    private String jobResult;

    @NotNull
    @Column(name = "origin_id")
    private Integer originId;

    @NotNull
    @Column(name = "origin_type")
    private String originType;

    @Column(name = "created_date", nullable = false)
    private Instant createdDate;

    @Column(name = "modified_date")
    private Instant modifiedDate;

    @Column(name = "created_user", nullable = false, length = 50)
    private String createdUser;

    @Column(name = "modified_user", length = 50)
    private String modifiedUser;

    @ManyToOne
    @JoinColumn(name = "action_id", referencedColumnName = "id", insertable = false, updatable = false)
    private UtmIncidentAction action;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getActionId() {
        return actionId;
    }

    public void setActionId(Long actionId) {
        this.actionId = actionId;
    }

    public String getParams() {
        return params;
    }

    public void setParams(String params) {
        this.params = params;
    }

    public String getAgent() {
        return agent;
    }

    public void setAgent(String agent) {
        this.agent = agent;
    }

    public Integer getStatus() {
        return status;
    }

    public void setStatus(Integer status) {
        this.status = status;
    }

    public String getJobResult() {
        return jobResult;
    }

    public void setJobResult(String jobResult) {
        this.jobResult = jobResult;
    }

    public Integer getOriginId() {
        return originId;
    }

    public void setOriginId(Integer originId) {
        this.originId = originId;
    }

    public String getOriginType() {
        return originType;
    }

    public void setOriginType(String originType) {
        this.originType = originType;
    }

    public Instant getCreatedDate() {
        return createdDate;
    }

    public void setCreatedDate(Instant createdDate) {
        this.createdDate = createdDate;
    }

    public Instant getModifiedDate() {
        return modifiedDate;
    }

    public void setModifiedDate(Instant modifiedDate) {
        this.modifiedDate = modifiedDate;
    }

    public String getCreatedUser() {
        return createdUser;
    }

    public void setCreatedUser(String createdUser) {
        this.createdUser = createdUser;
    }

    public String getModifiedUser() {
        return modifiedUser;
    }

    public void setModifiedUser(String modifiedUser) {
        this.modifiedUser = modifiedUser;
    }

    public UtmIncidentAction getAction() {
        return action;
    }

    public void setAction(UtmIncidentAction action) {
        this.action = action;
    }

    public IncidentJobStatus getStatusLabel() {
        switch (status) {
            case 0:
                return IncidentJobStatus.PENDING;
            case 1:
                return IncidentJobStatus.RUNNING;
            case 2:
                return IncidentJobStatus.EXECUTED;
            default:
                return IncidentJobStatus.ERROR;
        }
    }
}

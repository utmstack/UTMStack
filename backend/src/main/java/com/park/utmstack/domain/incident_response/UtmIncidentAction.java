package com.park.utmstack.domain.incident_response;


import javax.persistence.*;
import java.io.Serializable;
import java.time.Instant;

/**
 * A UtmIncidentAction.
 */
@Entity
@Table(name = "utm_incident_actions")
public class UtmIncidentAction implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "action_command")
    private String actionCommand;

    @Column(name = "action_description")
    private String actionDescription;

    @Column(name = "action_params")
    private String actionParams;

    @Column(name = "action_type")
    private Integer actionType;

    @Column(name = "action_editable", nullable = false)
    private Boolean actionEditable;

    @Column(name = "created_date", nullable = false)
    private Instant createdDate;

    @Column(name = "modified_date")
    private Instant modifiedDate;

    @Column(name = "created_user", nullable = false, length = 50)
    private String createdUser;

    @Column(name = "modified_user", length = 50)
    private String modifiedUser;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getActionCommand() {
        return actionCommand;
    }

    public void setActionCommand(String actionCommand) {
        this.actionCommand = actionCommand;
    }

    public String getActionDescription() {
        return actionDescription;
    }

    public void setActionDescription(String actionDescription) {
        this.actionDescription = actionDescription;
    }

    public String getActionParams() {
        return actionParams;
    }

    public void setActionParams(String actionParams) {
        this.actionParams = actionParams;
    }

    public Integer getActionType() {
        return actionType;
    }

    public void setActionType(Integer actionType) {
        this.actionType = actionType;
    }

    public Boolean getActionEditable() {
        return actionEditable;
    }

    public void setActionEditable(Boolean actionEditable) {
        this.actionEditable = actionEditable;
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
}

package com.park.utmstack.domain;


import javax.persistence.*;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.time.Instant;

/**
 * A UtmAlertLog.
 */
@Entity
@Table(name = "utm_alert_log")
public class UtmAlertLog implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @NotNull
    @Size(max = 100)
    @Column(name = "alert_id", length = 100, nullable = false)
    private String alertId;

    @NotNull
    @Size(max = 50)
    @Column(name = "log_user", length = 50, nullable = false)
    private String logUser;

    @NotNull
    @Column(name = "log_date", nullable = false)
    private Instant logDate;

    @NotNull
    @Size(max = 50)
    @Column(name = "log_action", length = 50, nullable = false)
    private String logAction;

    @Column(name = "log_old_value")
    private String logOldValue;

    @Column(name = "log_new_value")
    private String logNewValue;

    @Column(name = "log_message")
    private String logMessage;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getAlertId() {
        return alertId;
    }

    public UtmAlertLog alertId(String alertId) {
        this.alertId = alertId;
        return this;
    }

    public void setAlertId(String alertId) {
        this.alertId = alertId;
    }

    public String getLogUser() {
        return logUser;
    }

    public UtmAlertLog logUser(String logUser) {
        this.logUser = logUser;
        return this;
    }

    public void setLogUser(String logUser) {
        this.logUser = logUser;
    }

    public Instant getLogDate() {
        return logDate;
    }

    public UtmAlertLog logDate(Instant logDate) {
        this.logDate = logDate;
        return this;
    }

    public void setLogDate(Instant logDate) {
        this.logDate = logDate;
    }

    public String getLogAction() {
        return logAction;
    }

    public UtmAlertLog logAction(String logAction) {
        this.logAction = logAction;
        return this;
    }

    public void setLogAction(String logAction) {
        this.logAction = logAction;
    }

    public String getLogOldValue() {
        return logOldValue;
    }

    public UtmAlertLog logOldValue(String logOldValue) {
        this.logOldValue = logOldValue;
        return this;
    }

    public void setLogOldValue(String logOldValue) {
        this.logOldValue = logOldValue;
    }

    public String getLogNewValue() {
        return logNewValue;
    }

    public UtmAlertLog logNewValue(String logNewValue) {
        this.logNewValue = logNewValue;
        return this;
    }

    public void setLogNewValue(String logNewValue) {
        this.logNewValue = logNewValue;
    }

    public String getLogMessage() {
        return logMessage;
    }

    public void setLogMessage(String logMessage) {
        this.logMessage = logMessage;
    }
}

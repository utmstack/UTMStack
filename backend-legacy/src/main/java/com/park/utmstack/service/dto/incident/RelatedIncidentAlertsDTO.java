package com.park.utmstack.service.dto.incident;

import javax.validation.constraints.NotNull;

public class RelatedIncidentAlertsDTO {
    @NotNull
    private String alertId;

    @NotNull
    private String alertName;

    @NotNull
    private Integer alertStatus;

    @NotNull
    private Integer alertSeverity;

    public RelatedIncidentAlertsDTO() {
    }

    public String getAlertId() {
        return alertId;
    }

    public void setAlertId(String alertId) {
        this.alertId = alertId;
    }

    public String getAlertName() {
        return alertName;
    }

    public void setAlertName(String alertName) {
        this.alertName = alertName;
    }

    public Integer getAlertStatus() {
        return alertStatus;
    }

    public void setAlertStatus(Integer alertStatus) {
        this.alertStatus = alertStatus;
    }

    public Integer getAlertSeverity() {
        return alertSeverity;
    }

    public void setAlertSeverity(Integer alertSeverity) {
        this.alertSeverity = alertSeverity;
    }
}

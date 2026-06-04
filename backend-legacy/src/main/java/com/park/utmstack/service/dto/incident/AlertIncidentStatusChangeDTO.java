package com.park.utmstack.service.dto.incident;

import javax.validation.constraints.NotNull;
import java.util.List;

public class AlertIncidentStatusChangeDTO {
    @NotNull
    private Long incidentId;
    @NotNull
    private List<String> alertIds;
    @NotNull
    private Integer status;

    public AlertIncidentStatusChangeDTO() {
    }

    public Long getIncidentId() {
        return incidentId;
    }

    public void setIncidentId(Long incidentId) {
        this.incidentId = incidentId;
    }

    public List<String> getAlertIds() {
        return alertIds;
    }

    public void setAlertIds(List<String> alertId) {
        this.alertIds = alertId;
    }

    public Integer getStatus() {
        return status;
    }

    public void setStatus(Integer status) {
        this.status = status;
    }
}

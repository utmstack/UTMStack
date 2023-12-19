package com.park.utmstack.service.dto.incident;

import javax.validation.constraints.NotNull;
import java.util.List;

public class AddToIncidentDTO {
    @NotNull
    public Long incidentId;
    @NotNull
    public List<RelatedIncidentAlertsDTO> alertList;

    public AddToIncidentDTO() {
    }

    public Long getIncidentId() {
        return incidentId;
    }

    public void setIncidentId(Long incidentId) {
        this.incidentId = incidentId;
    }

    public List<RelatedIncidentAlertsDTO> getAlertList() {
        return alertList;
    }

    public void setAlertList(List<RelatedIncidentAlertsDTO> alertList) {
        this.alertList = alertList;
    }
}

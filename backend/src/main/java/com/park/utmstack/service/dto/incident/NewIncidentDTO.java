package com.park.utmstack.service.dto.incident;

import javax.validation.constraints.NotNull;
import javax.validation.constraints.Pattern;
import java.util.List;

public class NewIncidentDTO {
    @NotNull
    @Pattern(regexp = "^[^\"]*$", message = "Double quotes are not allowed")
    public String incidentName;
    public String incidentDescription;
    public String incidentAssignedTo;
    @NotNull
    public List<RelatedIncidentAlertsDTO> alertList;

    public NewIncidentDTO() {
    }

    public String getIncidentName() {
        return incidentName;
    }

    public void setIncidentName(String incidentName) {
        this.incidentName = incidentName;
    }

    public String getIncidentDescription() {
        return incidentDescription;
    }

    public void setIncidentDescription(String incidentDescription) {
        this.incidentDescription = incidentDescription;
    }

    public String getIncidentAssignedTo() {
        return incidentAssignedTo;
    }

    public void setIncidentAssignedTo(String incidentAssignedTo) {
        this.incidentAssignedTo = incidentAssignedTo;
    }

    public List<RelatedIncidentAlertsDTO> getAlertList() {
        return alertList;
    }

    public void setAlertList(List<RelatedIncidentAlertsDTO> alertList) {
        this.alertList = alertList;
    }

    @Deprecated
    public String toString() {
        return "{" +
            "incidentName='" + incidentName + '\'' +
            ", incidentDescription='" + incidentDescription + '\'' +
            ", incidentAssignedTo='" + incidentAssignedTo + '\'' +
            ", alertList=" + alertList +
            '}';
    }
}

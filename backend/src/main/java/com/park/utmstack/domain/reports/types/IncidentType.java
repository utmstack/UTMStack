package com.park.utmstack.domain.reports.types;

import com.park.utmstack.domain.incident_response.UtmIncidentJob;
import com.park.utmstack.domain.shared_types.AlertType;

import java.util.List;

public class IncidentType {
    private AlertType incident;
    private List<UtmIncidentJob> srcResponses;
    private List<UtmIncidentJob> destResponses;

    public AlertType getIncident() {
        return incident;
    }

    public void setIncident(AlertType incident) {
        this.incident = incident;
    }

    public List<UtmIncidentJob> getSrcResponses() {
        return srcResponses;
    }

    public void setSrcResponses(List<UtmIncidentJob> srcResponses) {
        this.srcResponses = srcResponses;
    }

    public List<UtmIncidentJob> getDestResponses() {
        return destResponses;
    }

    public void setDestResponses(List<UtmIncidentJob> destResponses) {
        this.destResponses = destResponses;
    }
}

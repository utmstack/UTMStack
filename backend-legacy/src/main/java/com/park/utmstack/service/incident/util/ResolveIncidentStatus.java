package com.park.utmstack.service.incident.util;

import com.park.utmstack.domain.incident.UtmIncident;

public class ResolveIncidentStatus {

    ResolveIncidentStatus() {
    }

    public static String incidentLabel(UtmIncident utmIncident) {
        switch (utmIncident.getIncidentStatus()) {
            case OPEN:
                return "Open";
            case IN_REVIEW:
                return "In review";
            case COMPLETED:
                return "Completed";
            default:
                return "none";
        }
    }

    public static Integer getAlertStatus(UtmIncident utmIncident) {
        switch (utmIncident.getIncidentStatus()) {
            case OPEN:
                return 2;
            case IN_REVIEW:
                return 3;
            case COMPLETED:
                return 5;
            default:
                return 0;
        }
    }

    public static String incidentLabelByInteger(Integer incidentStatus) {
        switch (incidentStatus) {
            case 2:
                return "Open";
            case 3:
                return "In review";
            case 5:
                return "Completed";
            default:
                return "none";
        }
    }
}

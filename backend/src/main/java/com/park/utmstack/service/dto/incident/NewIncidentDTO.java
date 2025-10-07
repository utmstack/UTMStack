package com.park.utmstack.service.dto.incident;

import com.park.utmstack.service.dto.auditable.AuditableDTO;
import lombok.Getter;
import lombok.Setter;

import javax.validation.constraints.NotNull;
import javax.validation.constraints.Pattern;
import java.util.List;
import java.util.Map;

@Setter
@Getter
public class NewIncidentDTO implements AuditableDTO {
    @NotNull
    @Pattern(regexp = "^[^\"]*$", message = "Double quotes are not allowed")
    public String incidentName;
    public String incidentDescription;
    public String incidentAssignedTo;
    @NotNull
    public List<RelatedIncidentAlertsDTO> alertList;

    public NewIncidentDTO() {
    }

    @Override
    public Map<String, Object> toAuditMap() {
        List<String> alertIds = alertList.stream()
                .map(RelatedIncidentAlertsDTO::getAlertId)
                .toList();

        return Map.of(
                "incidentName", incidentName,
                "alertIds", alertIds
        );
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

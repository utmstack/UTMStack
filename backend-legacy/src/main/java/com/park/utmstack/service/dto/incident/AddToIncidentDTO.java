package com.park.utmstack.service.dto.incident;

import com.park.utmstack.service.dto.auditable.AuditableDTO;
import lombok.Getter;
import lombok.Setter;

import javax.validation.constraints.NotNull;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@Setter
@Getter
public class AddToIncidentDTO implements AuditableDTO {
    @NotNull
    public Long incidentId;
    @NotNull
    public List<RelatedIncidentAlertsDTO> alertList;

    public AddToIncidentDTO() {
    }

    @Override
    public Map<String, Object> toAuditMap() {
        List<String> alertIds = alertList.stream()
                .map(RelatedIncidentAlertsDTO::getAlertId)
                .toList();

        return Map.of(
                "incidentId", incidentId,
                "alertIds", alertIds
        );
    }
}

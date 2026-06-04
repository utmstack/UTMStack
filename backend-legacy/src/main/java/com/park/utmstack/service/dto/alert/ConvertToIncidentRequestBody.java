package com.park.utmstack.service.dto.alert;

import com.park.utmstack.service.dto.auditable.AuditableDTO;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

import javax.validation.constraints.NotNull;
import javax.validation.constraints.Pattern;
import java.util.List;
import java.util.Map;

@AllArgsConstructor
@NoArgsConstructor
@Getter
@Setter
public class ConvertToIncidentRequestBody implements AuditableDTO {
    @NotNull
    private List<String> eventIds;
    @NotNull
    @Pattern(regexp = "^[^\"]*$", message = "Double quotes are not allowed")
    private String incidentName;
    @NotNull
    private Integer incidentId;
    @NotNull
    private String incidentSource;

    @Override
    public Map<String, Object> toAuditMap() {
        return Map.of(
                "eventIds", eventIds,
                "incidentName", incidentName,
                "incidentId", incidentId,
                "incidentSource", incidentSource
        );
    }
}

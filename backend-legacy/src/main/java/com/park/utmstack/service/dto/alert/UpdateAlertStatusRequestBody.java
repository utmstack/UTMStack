package com.park.utmstack.service.dto.alert;

import com.park.utmstack.service.dto.auditable.AuditableDTO;
import lombok.*;

import javax.validation.constraints.NotNull;
import java.util.List;
import java.util.Map;

@AllArgsConstructor
@NoArgsConstructor
@Getter
@Setter
public class UpdateAlertStatusRequestBody implements AuditableDTO {
    @NotNull
    private List<String> alertIds;
    private String statusObservation;
    @NotNull
    private int status;
    boolean addFalsePositiveTag;


    @Override
    public Map<String, Object> toAuditMap() {
        return Map.of(
                "alertIds", alertIds,
                "statusObservation", statusObservation,
                "status", status,
                "addFalsePositiveTag", addFalsePositiveTag
        );
    }
}

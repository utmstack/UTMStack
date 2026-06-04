package com.park.utmstack.service.dto.alert;

import com.park.utmstack.service.dto.auditable.AuditableDTO;
import lombok.*;

import javax.validation.constraints.NotNull;
import java.util.List;
import java.util.Map;

@Getter
@Setter
@AllArgsConstructor
@NoArgsConstructor
public class UpdateAlertTagsRequestBody implements AuditableDTO {

    @NotNull
    private List<String> alertIds;

    private List<String> tags;
    @NotNull
    private Boolean createRule;

    @Override
    public Map<String, Object> toAuditMap() {
        return Map.of(
                "alertIds", alertIds,
                "tags", tags,
                "createRule", createRule
        );
    }
}

package com.park.utmstack.service.dto.application_modules;

import com.park.utmstack.domain.application_modules.enums.ModuleName;
import com.park.utmstack.service.dto.auditable.AuditableDTO;
import lombok.*;

import java.util.Map;

@Getter
@Setter
@Builder
@AllArgsConstructor
@NoArgsConstructor
public class ModuleActivationDTO implements AuditableDTO {
    private Long serverId;
    private ModuleName moduleName;
    private Boolean activationStatus;

    @Override
    public Map<String, Object> toAuditMap() {
        return Map.of(
                "serverId", serverId,
                "moduleName", moduleName != null ? moduleName.name() : null,
                "activationStatus", activationStatus
        );
    }
}

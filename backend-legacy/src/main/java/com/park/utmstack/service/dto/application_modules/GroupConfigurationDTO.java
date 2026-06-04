package com.park.utmstack.service.dto.application_modules;

import com.park.utmstack.domain.application_modules.UtmModuleGroupConfiguration;
import com.park.utmstack.domain.application_modules.validators.ValidModuleConfiguration;
import com.park.utmstack.service.dto.auditable.AuditableDTO;
import lombok.Data;

import javax.validation.constraints.NotEmpty;
import javax.validation.constraints.NotNull;
import java.util.List;
import java.util.Map;

@Data
@ValidModuleConfiguration
public class GroupConfigurationDTO implements AuditableDTO {
    @NotNull
    private Long moduleId;
    @NotEmpty
    private List<UtmModuleGroupConfiguration> keys;

    @Override
    public Map<String, Object> toAuditMap() {
        return Map.of("moduleId", moduleId);
    }
}

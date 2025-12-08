package com.park.utmstack.service.dto.collectors.dto;

import com.park.utmstack.domain.application_modules.UtmModuleGroupConfiguration;
import lombok.Getter;
import lombok.Setter;

import javax.validation.constraints.NotNull;
import java.util.List;

@Setter
@Getter
public class CollectorConfigDTO {
        @NotNull
        CollectorDTO collector;
        @NotNull
        private Long moduleId;
        @NotNull
        private List<UtmModuleGroupConfiguration> keys;

}

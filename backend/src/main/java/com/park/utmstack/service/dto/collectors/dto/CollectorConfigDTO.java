package com.park.utmstack.service.dto.collectors.dto;

import com.park.utmstack.domain.application_modules.UtmModuleGroupConfiguration;
import com.park.utmstack.domain.collector.validators.UniqueServerName;
import lombok.Getter;
import lombok.Setter;

import jakarta.validation.constraints.NotEmpty;
import jakarta.validation.constraints.NotNull;
import java.util.List;

@Setter
@Getter
public class CollectorConfigDTO {

    @NotNull
    CollectorDTO collector;

    @NotNull
    private Long moduleId;

    @NotNull
    @UniqueServerName
    private List<UtmModuleGroupConfiguration> keys;

}

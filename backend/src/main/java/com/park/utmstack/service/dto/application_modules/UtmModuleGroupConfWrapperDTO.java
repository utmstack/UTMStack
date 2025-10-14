package com.park.utmstack.service.dto.application_modules;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class UtmModuleGroupConfWrapperDTO {
    private List<UtmModuleGroupConfDTO> moduleGroupConfigurations;
}


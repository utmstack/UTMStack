package com.park.utmstack.service.dto.application_modules;

import com.park.utmstack.domain.application_modules.enums.ModuleRequirementStatus;
import com.park.utmstack.domain.application_modules.types.ModuleRequirement;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

import java.util.List;

@Setter
@Getter
@AllArgsConstructor
@NoArgsConstructor
public class CheckRequirementsResponse {
    private ModuleRequirementStatus status;
    private List<ModuleRequirement> checks;

}

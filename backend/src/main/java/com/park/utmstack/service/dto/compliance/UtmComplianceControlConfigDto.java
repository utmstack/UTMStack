package com.park.utmstack.service.dto.compliance;

import com.park.utmstack.domain.compliance.enums.ComplianceStrategy;
import lombok.Data;

import javax.validation.constraints.NotNull;
import javax.validation.constraints.Size;
import java.util.List;

@Data
public class UtmComplianceControlConfigDto {

    private Long id;

    @NotNull
    private Long standardSectionId;

    private UtmComplianceStandardSectionDto section;

    @NotNull
    @Size(min = 10, max = 200)
    private String controlName;

    @Size(max = 2000)
    private String controlSolution;

    @Size(max = 2000)
    private String controlRemediation;

    @NotNull
    private ComplianceStrategy controlStrategy;

    private List<UtmComplianceQueryConfigDto> queriesConfigs;
}

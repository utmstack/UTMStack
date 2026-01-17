package com.park.utmstack.service.dto.compliance;

import com.park.utmstack.domain.compliance.enums.ComplianceStrategy;
import lombok.Data;

import java.util.List;

@Data
public class UtmComplianceControlConfigDto {
    private Long id;
    private Long standardSectionId;
    private String controlName;
    private String controlSolution;
    private String controlRemediation;
    private ComplianceStrategy controlStrategy;
    private List<UtmComplianceQueryConfigDto> queriesConfigs;
}
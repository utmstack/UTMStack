package com.park.utmstack.service.dto.compliance;

import com.park.utmstack.domain.compliance.enums.ComplianceStrategy;
import lombok.Data;

import java.util.List;

@Data
public class UtmComplianceControlEvaluationDto {
    private Long id;
    private Long standardSectionId;
    private UtmComplianceStandardSectionDto section;

    private String controlName;
    private String controlSolution;
    private String controlRemediation;
    private ComplianceStrategy controlStrategy;

    private List<UtmComplianceQueryConfigDto> queriesConfigs;

    private String lastEvaluationStatus;
    private String lastEvaluationTimestamp;
}

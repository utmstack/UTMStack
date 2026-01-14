package com.park.utmstack.service.dto.compliance;

import lombok.Data;

@Data
public class UtmComplianceControlConfigRequestDto {
    private Long standardSectionId;
    private String controlName;
    private String controlSolution;
    private String controlRemediation;
    private String controlStrategy;
}
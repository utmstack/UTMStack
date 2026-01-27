package com.park.utmstack.service.dto.compliance;

import lombok.Data;

@Data
public class UtmComplianceStandardSectionDto {
    private Long id;
    private String standardSectionName;
    private String standardSectionDescription;
    private Long standardId;
}
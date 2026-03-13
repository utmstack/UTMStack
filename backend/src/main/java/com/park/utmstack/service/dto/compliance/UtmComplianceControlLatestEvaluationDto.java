package com.park.utmstack.service.dto.compliance;

import lombok.Data;
import lombok.EqualsAndHashCode;

@Data
@EqualsAndHashCode(callSuper = true)
public class UtmComplianceControlLatestEvaluationDto extends UtmComplianceControlConfigDto {
    private String lastEvaluationStatus;
    private String lastEvaluationTimestamp;
}

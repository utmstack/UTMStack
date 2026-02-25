package com.park.utmstack.service.dto.compliance;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;

import java.time.Instant;
import java.util.List;

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
public class UtmComplianceControlEvaluationsDto {
    private Long controlId;
    private String controlName;
    private String status;
    private Instant timestamp;
    private List<UtmComplianceQueryEvaluationDto> queryEvaluations;
}


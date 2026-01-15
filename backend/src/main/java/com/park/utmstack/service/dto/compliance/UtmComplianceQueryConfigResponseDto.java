package com.park.utmstack.service.dto.compliance;

import lombok.Data;

@Data
public class UtmComplianceQueryConfigResponseDto {
    private String queryDescription;
    private String sqlQuery;
    private String evaluationRule;
    private Long indexPatternId;
    private Long reportConfigId;
}

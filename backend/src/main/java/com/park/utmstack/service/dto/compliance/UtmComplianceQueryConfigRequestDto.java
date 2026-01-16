package com.park.utmstack.service.dto.compliance;

import com.park.utmstack.domain.compliance.enums.EvaluationRule;
import lombok.Data;

@Data
public class UtmComplianceQueryConfigRequestDto {
    private Long id;
    private String queryDescription;
    private String sqlQuery;
    private EvaluationRule evaluationRule;
    private Long indexPatternId;
    private Long controlConfigId;
}

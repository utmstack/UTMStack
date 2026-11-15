package com.park.utmstack.service.dto.compliance;

import com.park.utmstack.domain.compliance.enums.EvaluationRule;
import lombok.Data;

import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Size;

@Data
public class UtmComplianceQueryConfigDto {

    private Long id;

    @NotNull
    @Size(min = 10, max = 200)
    private String queryName;

    @NotNull
    @Size(max = 2000)
    private String queryDescription;

    @NotNull
    @Size(max = 2000)
    private String sqlQuery;

    @NotNull
    private EvaluationRule evaluationRule;

    private Integer ruleValue;

    @NotNull
    private Long indexPatternId;

    @NotNull
    private Long controlConfigId;
}

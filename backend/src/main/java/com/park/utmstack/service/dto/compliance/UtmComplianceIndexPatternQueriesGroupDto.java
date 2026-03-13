package com.park.utmstack.service.dto.compliance;

import lombok.Data;

import java.util.List;

@Data
public class UtmComplianceIndexPatternQueriesGroupDto {
    private Long indexPatternId;
    private String indexPatternName;
    private List<UtmComplianceQueryEvaluationDto> queries;
}
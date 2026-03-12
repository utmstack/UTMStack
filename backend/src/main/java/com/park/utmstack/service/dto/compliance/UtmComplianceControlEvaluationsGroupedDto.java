package com.park.utmstack.service.dto.compliance;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;
import java.util.List;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class UtmComplianceControlEvaluationsGroupedDto {
    private Long controlId;
    private String controlName;
    private String status;
    private Instant timestamp;
    private List<IndexPatternQueriesGroupDto> queryEvaluations;
}
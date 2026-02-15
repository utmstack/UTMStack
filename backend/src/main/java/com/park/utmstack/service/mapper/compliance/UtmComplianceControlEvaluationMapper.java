package com.park.utmstack.service.mapper.compliance;

import com.park.utmstack.service.dto.compliance.UtmComplianceControlEvaluationDto;
import com.park.utmstack.service.dto.compliance.UtmComplianceQueryEvaluationDto;

import java.time.Instant;
import java.util.List;
import java.util.Map;

public class UtmComplianceControlEvaluationMapper {

    private UtmComplianceControlEvaluationMapper() {

    }

    public static UtmComplianceControlEvaluationDto mapToEvaluationDto(Map<String, Object> src) {
        UtmComplianceControlEvaluationDto dto = new UtmComplianceControlEvaluationDto();

        dto.setControlId(((Number) src.get("control_id")).longValue());
        dto.setControlName((String) src.get("control_name"));
        dto.setStatus((String) src.get("status"));
        dto.setTimestamp(Instant.parse((String) src.get("timestamp")));

        List<Map<String, Object>> q = (List<Map<String, Object>>) src.get("query_evaluations");
        if (q != null) {
            dto.setQueryEvaluations(q.stream().map(UtmComplianceControlEvaluationMapper::mapQueryEval).toList());
        }

        return dto;
    }

    private static UtmComplianceQueryEvaluationDto mapQueryEval(Map<String, Object> src) {
        UtmComplianceQueryEvaluationDto dto = new UtmComplianceQueryEvaluationDto();

        dto.setQueryConfigId(((Number) src.get("queryConfigId")).longValue());
        dto.setQueryName((String) src.get("queryName"));
        dto.setHits(((Number) src.get("hits")).intValue());
        dto.setStatus((String) src.get("status"));
        dto.setEvidence((List<Map<String, Object>>) src.get("evidence"));

        return dto;
    }
}

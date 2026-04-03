package com.park.utmstack.service.mapper.compliance;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;

import com.park.utmstack.service.dto.compliance.UtmComplianceControlEvaluationHistoryDto;
import com.park.utmstack.service.dto.compliance.UtmComplianceQueryEvaluationDto;

import java.time.Instant;
import java.util.List;
import java.util.Map;
public class UtmComplianceControlEvaluationHistoryMapper {

    private UtmComplianceControlEvaluationHistoryMapper() {

    }

    public static UtmComplianceControlEvaluationHistoryDto mapToEvaluationDto(Map<String, Object> src) {
        UtmComplianceControlEvaluationHistoryDto dto = new UtmComplianceControlEvaluationHistoryDto();

        dto.setControlId(((Number) src.get("control_id")).longValue());
        dto.setControlName((String) src.get("control_name"));
        dto.setStatus((String) src.get("status"));
        dto.setTimestamp(Instant.parse((String) src.get("timestamp")));


        ObjectMapper mapper = new ObjectMapper();
        List<Map<String, Object>> q = mapper.convertValue(src.get("query_evaluations"), new TypeReference<>() {});
        if (q != null) {
            dto.setQueryEvaluations(q.stream().map(UtmComplianceControlEvaluationHistoryMapper::mapQueryEval).toList());
        }

        return dto;
    }

    private static UtmComplianceQueryEvaluationDto mapQueryEval(Map<String, Object> src) {
        UtmComplianceQueryEvaluationDto dto = new UtmComplianceQueryEvaluationDto();

        dto.setQueryConfigId(((Number) src.get("queryConfigId")).longValue());
        dto.setQueryName((String) src.get("queryName"));
        dto.setEvaluationRule((String) src.get("evaluationRule"));

        Object raw = src.get("ruleValue");
        dto.setRuleValue(raw instanceof Number ? ((Number) raw).intValue() : null);

        dto.setHits(((Number) src.get("hits")).intValue());
        dto.setStatus((String) src.get("status"));

        ObjectMapper mapper = new ObjectMapper();
        List<Map<String, Object>> evidence = mapper.convertValue(src.get("evidence"), new TypeReference<>() {});
        dto.setEvidence(evidence);

        return dto;
    }
}

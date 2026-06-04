package com.park.utmstack.service.mapper.compliance;

import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigDto;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlLatestEvaluationDto;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlEvaluationHistoryDto;

import java.time.Instant;
import java.util.Map;

public class UtmComplianceControlLatestEvaluationMapper {

    public static UtmComplianceControlLatestEvaluationDto toDto(
            UtmComplianceControlConfigDto control,
            UtmComplianceControlEvaluationHistoryDto controlEvaluationHistory
    ) {
        UtmComplianceControlLatestEvaluationDto dto = new UtmComplianceControlLatestEvaluationDto();

        dto.setId(control.getId());
        dto.setStandardSectionId(control.getStandardSectionId());
        dto.setSection(control.getSection());
        dto.setControlName(control.getControlName());
        dto.setControlSolution(control.getControlSolution());
        dto.setControlRemediation(control.getControlRemediation());
        dto.setControlStrategy(control.getControlStrategy());
        dto.setQueriesConfigs(control.getQueriesConfigs());

        if (controlEvaluationHistory != null) {
            dto.setLastEvaluationStatus(controlEvaluationHistory.getStatus());
            dto.setLastEvaluationTimestamp(
                    controlEvaluationHistory.getTimestamp() != null ? controlEvaluationHistory.getTimestamp().toString() : null
            );
        }

        return dto;
    }

    public static UtmComplianceControlEvaluationHistoryDto mapToEvaluationDto(Map<String, Object> source) {
        if (source == null) {
            return null;
        }

        UtmComplianceControlEvaluationHistoryDto dto = new UtmComplianceControlEvaluationHistoryDto();
        dto.setControlId(getLong(source.get("control_id")));
        dto.setControlName(getString(source.get("control_name")));
        dto.setStatus(getString(source.get("status")));

        Object ts = source.get("timestamp");
        if (ts != null) {
            dto.setTimestamp(Instant.parse(ts.toString()));
        }

        return dto;
    }

    private static String getString(Object o) {
        return o != null ? o.toString() : null;
    }

    private static Long getLong(Object o) {
        if (o == null) return null;
        if (o instanceof Number n) return n.longValue();
        return Long.parseLong(o.toString());
    }
}

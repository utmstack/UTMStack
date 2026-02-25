package com.park.utmstack.service.mapper.compliance;

import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigDto;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlEvaluationDto;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlEvaluationsDto;

import java.time.Instant;
import java.util.Map;

public class UtmComplianceControlEvaluationMapper {

    public static UtmComplianceControlEvaluationDto toDto(
            UtmComplianceControlConfigDto control,
            UtmComplianceControlEvaluationsDto controlEvaluations
    ) {
        UtmComplianceControlEvaluationDto dto = new UtmComplianceControlEvaluationDto();

        dto.setId(control.getId());
        dto.setStandardSectionId(control.getStandardSectionId());
        dto.setSection(control.getSection());
        dto.setControlName(control.getControlName());
        dto.setControlSolution(control.getControlSolution());
        dto.setControlRemediation(control.getControlRemediation());
        dto.setControlStrategy(control.getControlStrategy());
        dto.setQueriesConfigs(control.getQueriesConfigs());

        //TODO: ELENA - this is a temporary solution, we need to decide how to handle multiple evaluations for the same control
        if (controlEvaluations != null) {
            dto.setLastEvaluationStatus(controlEvaluations.getStatus());
            dto.setLastEvaluationTimestamp(
                    controlEvaluations.getTimestamp() != null ? controlEvaluations.getTimestamp().toString() : null
            );
        }

        return dto;
    }

    public static UtmComplianceControlEvaluationsDto mapToEvaluationDto(Map<String, Object> source) {
        if (source == null) return null;

        UtmComplianceControlEvaluationsDto dto = new UtmComplianceControlEvaluationsDto();

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

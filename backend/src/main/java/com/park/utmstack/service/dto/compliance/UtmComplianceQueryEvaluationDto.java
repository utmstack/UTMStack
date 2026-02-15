package com.park.utmstack.service.dto.compliance;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import lombok.Data;

import java.util.List;
import java.util.Map;

@Data
@JsonIgnoreProperties(ignoreUnknown = true)
public class UtmComplianceQueryEvaluationDto {

    private Long queryConfigId;
    private String queryName;
    private Integer hits;
    private String status;
    private String errorMessage;
    private List<Map<String, Object>> evidence;
}

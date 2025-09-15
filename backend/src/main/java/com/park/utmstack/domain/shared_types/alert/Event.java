package com.park.utmstack.domain.shared_types.alert;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;

import java.util.List;
import java.util.Map;

@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
@JsonIgnoreProperties(ignoreUnknown = true)
public class Event {

    private String id;

    @JsonProperty("@timestamp")
    private String timestamp;

    private String deviceTime;
    private String dataType;
    private String dataSource;
    private String tenantId;
    private String tenantName;
    private String raw;

    private Map<String, Object> log;

    private Side target;
    private Side origin;

    private String protocol;
    private String connectionStatus;
    private Integer statusCode;
    private String actionResult;
    private String action;
    private String severity;

    private List<String> errors;
    private Map<String, ComplianceValues> compliance;
}


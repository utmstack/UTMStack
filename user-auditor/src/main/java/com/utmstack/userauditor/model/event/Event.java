package com.utmstack.userauditor.model.event;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;
import lombok.Getter;

import java.util.Map;


@Data
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
    private int statusCode;
    private String actionResult;
    private String action;
    private String command;
    private String severity;
}
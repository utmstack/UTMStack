package com.park.utmstack.service.dto;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.google.gson.Gson;
import com.google.gson.reflect.TypeToken;
import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseRule;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import org.springframework.util.StringUtils;

import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotEmpty;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Size;
import java.time.Instant;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

public class UtmAlertResponseRuleDTO {
    private Long id;
    @NotBlank
    @Size(max = 150)
    private String name;
    @Size(max = 512)
    private String description;
    @NotEmpty
    private List<FilterType> conditions;
    @NotBlank
    private String command;
    @NotNull
    private Boolean active;
    @NotBlank
    private String agentPlatform;
    @Size(max = 500)
    private String defaultAgent;
    private List<String> excludedAgents = new ArrayList<>();
    @JsonProperty(access = JsonProperty.Access.READ_ONLY)
    private String createdBy;
    @JsonProperty(access = JsonProperty.Access.READ_ONLY)
    private Instant createdDate;
    @JsonProperty(access = JsonProperty.Access.READ_ONLY)
    private String lastModifiedBy;
    @JsonProperty(access = JsonProperty.Access.READ_ONLY)
    private Instant lastModifiedDate;

    public UtmAlertResponseRuleDTO() {
    }

    public UtmAlertResponseRuleDTO(UtmAlertResponseRule rule) {
        this.id = rule.getId();
        this.name = rule.getRuleName();
        this.description = rule.getRuleDescription();
        this.conditions = new Gson().fromJson(rule.getRuleConditions(), TypeToken.getParameterized(List.class, FilterType.class).getType());
        this.command = rule.getRuleCmd();
        this.active = rule.getRuleActive();
        this.agentPlatform = rule.getAgentPlatform();
        this.defaultAgent = rule.getDefaultAgent();
        if (StringUtils.hasText(rule.getExcludedAgents()))
            this.excludedAgents.addAll(Arrays.asList(rule.getExcludedAgents().split(",")));
        this.createdBy = rule.getCreatedBy();
        this.createdDate = rule.getCreatedDate();
        this.lastModifiedBy = rule.getLastModifiedBy();
        this.lastModifiedDate = rule.getLastModifiedDate();
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public List<FilterType> getConditions() {
        return conditions;
    }

    public void setConditions(List<FilterType> conditions) {
        this.conditions = conditions;
    }

    public String getCommand() {
        return command;
    }

    public void setCommand(String command) {
        this.command = command;
    }

    public Boolean getActive() {
        return active;
    }

    public void setActive(Boolean active) {
        this.active = active;
    }

    public String getAgentPlatform() {
        return agentPlatform;
    }

    public String getDefaultAgent() {
        return defaultAgent;
    }

    public void setDefaultAgent(String defaultAgent) {
        this.defaultAgent = defaultAgent;
    }

    public void setAgentPlatform(String agentPlatform) {
        this.agentPlatform = agentPlatform;
    }

    public List<String> getExcludedAgents() {
        return excludedAgents;
    }

    public void setExcludedAgents(List<String> excludedAgents) {
        this.excludedAgents = excludedAgents;
    }
}

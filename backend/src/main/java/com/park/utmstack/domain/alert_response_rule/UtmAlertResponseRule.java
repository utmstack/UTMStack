package com.park.utmstack.domain.alert_response_rule;


import com.google.gson.Gson;
import com.park.utmstack.service.dto.UtmAlertResponseRuleDTO;
import org.springframework.data.annotation.CreatedBy;
import org.springframework.data.annotation.CreatedDate;
import org.springframework.data.annotation.LastModifiedBy;
import org.springframework.data.annotation.LastModifiedDate;
import org.springframework.data.jpa.domain.support.AuditingEntityListener;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import javax.persistence.*;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.time.Instant;

@Entity
@Table(name = "utm_alert_response_rule")
@EntityListeners(AuditingEntityListener.class)
public class UtmAlertResponseRule implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    @Column(name = "rule_name", length = 150, nullable = false)
    private String ruleName;
    @Column(name = "rule_description", length = 512)
    private String ruleDescription;
    @Column(name = "rule_conditions", nullable = false)
    private String ruleConditions;
    @Column(name = "rule_cmd", nullable = false)
    private String ruleCmd;
    @Column(name = "rule_active", nullable = false)
    private Boolean ruleActive;
    @Column(name = "agent_platform")
    private String agentPlatform;
    @Column(name = "excluded_agents")
    private String excludedAgents;
    @Size(max = 500)
    @Column(name = "default_agent" , length = 500)
    private String defaultAgent;
    @CreatedBy
    @Column(name = "created_by", nullable = false, length = 50, updatable = false)
    private String createdBy;
    @CreatedDate
    @Column(name = "created_date", updatable = false)
    private Instant createdDate;
    @LastModifiedBy
    @Column(name = "last_modified_by", length = 50)
    private String lastModifiedBy;
    @LastModifiedDate
    @Column(name = "last_modified_date")
    private Instant lastModifiedDate;

    public UtmAlertResponseRule() {
    }

    public UtmAlertResponseRule(UtmAlertResponseRuleDTO dto) {
        this.id = dto.getId();
        this.ruleName = dto.getName();
        this.ruleDescription = dto.getDescription();
        this.ruleConditions = new Gson().toJson(dto.getConditions());
        this.ruleCmd = dto.getCommand();
        this.ruleActive = dto.getActive();
        this.agentPlatform = dto.getAgentPlatform();
        this.defaultAgent = dto.getDefaultAgent();
        if (!CollectionUtils.isEmpty(dto.getExcludedAgents()))
            this.excludedAgents = String.join(",", dto.getExcludedAgents());
        else
            this.excludedAgents = null;
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getRuleName() {
        return ruleName;
    }

    public void setRuleName(String ruleName) {
        this.ruleName = ruleName;
    }

    public String getRuleDescription() {
        return ruleDescription;
    }

    public void setRuleDescription(String ruleDescription) {
        this.ruleDescription = ruleDescription;
    }

    public String getRuleConditions() {
        return ruleConditions;
    }

    public void setRuleConditions(String ruleConditions) {
        this.ruleConditions = ruleConditions;
    }

    public String getRuleCmd() {
        return ruleCmd;
    }

    public void setRuleCmd(String ruleCmd) {
        this.ruleCmd = ruleCmd;
    }

    public Boolean getRuleActive() {
        return ruleActive;
    }

    public void setRuleActive(Boolean ruleActive) {
        this.ruleActive = ruleActive;
    }

    public String getAgentPlatform() {
        return agentPlatform;
    }

    public void setAgentPlatform(String agentPlatform) {
        this.agentPlatform = agentPlatform;
    }

    public String getExcludedAgents() {
        return excludedAgents;
    }

    public void setExcludedAgents(String excludedAgents) {
        this.excludedAgents = excludedAgents;
    }

    public String getDefaultAgent() {
        return defaultAgent;
    }

    public void setDefaultAgent(String defaultAgent) {
        this.defaultAgent = defaultAgent;
    }

    public String getCreatedBy() {
        return createdBy;
    }

    public void setCreatedBy(String createdBy) {
        this.createdBy = createdBy;
    }

    public Instant getCreatedDate() {
        return createdDate;
    }

    public void setCreatedDate(Instant createdDate) {
        this.createdDate = createdDate;
    }

    public String getLastModifiedBy() {
        return lastModifiedBy;
    }

    public void setLastModifiedBy(String lastModifiedBy) {
        this.lastModifiedBy = lastModifiedBy;
    }

    public Instant getLastModifiedDate() {
        return lastModifiedDate;
    }

    public void setLastModifiedDate(Instant lastModifiedDate) {
        this.lastModifiedDate = lastModifiedDate;
    }
}

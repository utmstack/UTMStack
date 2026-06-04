package com.park.utmstack.service.dto;

import com.park.utmstack.domain.alert_response_rule.enums.RuleExecutionStatus;
import com.park.utmstack.domain.alert_response_rule.enums.RuleNonExecutionCause;
import tech.jhipster.service.filter.Filter;
import tech.jhipster.service.filter.InstantFilter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;

public class UtmAlertResponseRuleExecutionCriteria implements Serializable {

    private static final long serialVersionUID = 1L;
    private LongFilter id;
    private LongFilter ruleId;
    private StringFilter alertId;
    private StringFilter agent;
    private InstantFilter executionDate;
    private RuleExecutionStatusFilter executionStatus;
    private RuleNonExecutionCauseFilter nonExecutionCause;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public LongFilter getRuleId() {
        return ruleId;
    }

    public void setRuleId(LongFilter ruleId) {
        this.ruleId = ruleId;
    }

    public StringFilter getAlertId() {
        return alertId;
    }

    public void setAlertId(StringFilter alertId) {
        this.alertId = alertId;
    }

    public StringFilter getAgent() {
        return agent;
    }

    public void setAgent(StringFilter agent) {
        this.agent = agent;
    }

    public InstantFilter getExecutionDate() {
        return executionDate;
    }

    public void setExecutionDate(InstantFilter executionDate) {
        this.executionDate = executionDate;
    }

    public RuleExecutionStatusFilter getExecutionStatus() {
        return executionStatus;
    }

    public void setExecutionStatus(RuleExecutionStatusFilter executionStatus) {
        this.executionStatus = executionStatus;
    }

    public RuleNonExecutionCauseFilter getNonExecutionCause() {
        return nonExecutionCause;
    }

    public void setNonExecutionCause(RuleNonExecutionCauseFilter nonExecutionCause) {
        this.nonExecutionCause = nonExecutionCause;
    }

    public static class RuleExecutionStatusFilter extends Filter<RuleExecutionStatus> {
    }

    public static class RuleNonExecutionCauseFilter extends Filter<RuleNonExecutionCause> {
    }
}

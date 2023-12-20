package com.park.utmstack.domain.alert_response_rule;


import com.park.utmstack.domain.alert_response_rule.enums.RuleExecutionStatus;
import com.park.utmstack.domain.alert_response_rule.enums.RuleNonExecutionCause;
import org.springframework.data.annotation.CreatedDate;
import org.springframework.data.jpa.domain.support.AuditingEntityListener;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.time.Instant;

@Entity
@Table(name = "utm_alert_response_rule_execution")
@EntityListeners(AuditingEntityListener.class)
public class UtmAlertResponseRuleExecution implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @NotNull
    @Column(name = "rule_id", nullable = false)
    private Long ruleId;

    @Size(max = 150)
    @NotNull
    @Column(name = "alert_id", nullable = false, length = 150)
    private String alertId;

    @NotNull
    @Column(name = "command", nullable = false)
    private String command;

    @Column(name = "command_result")
    private String commandResult;

    @Size(max = 150)
    @NotNull
    @Column(name = "agent", nullable = false, length = 150)
    private String agent;

    @NotNull
    @CreatedDate
    @Column(name = "execution_date", nullable = false, updatable = false)
    private Instant executionDate;

    @NotNull
    @Enumerated(EnumType.STRING)
    @Column(name = "execution_status", nullable = false, length = 100)
    private RuleExecutionStatus executionStatus;

    @Enumerated(EnumType.STRING)
    @Column(name = "non_execution_cause", length = 100)
    private RuleNonExecutionCause nonExecutionCause;

    @Column(name = "execution_retries")
    private Integer executionRetries = 0;


    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getRuleId() {
        return ruleId;
    }

    public void setRuleId(Long ruleId) {
        this.ruleId = ruleId;
    }

    public String getAlertId() {
        return alertId;
    }

    public void setAlertId(String alertId) {
        this.alertId = alertId;
    }

    public String getCommand() {
        return command;
    }

    public void setCommand(String command) {
        this.command = command;
    }

    public String getCommandResult() {
        return commandResult;
    }

    public void setCommandResult(String commandResult) {
        this.commandResult = commandResult;
    }

    public String getAgent() {
        return agent;
    }

    public void setAgent(String agent) {
        this.agent = agent;
    }

    public Instant getExecutionDate() {
        return executionDate;
    }

    public void setExecutionDate(Instant executionDate) {
        this.executionDate = executionDate;
    }

    public RuleExecutionStatus getExecutionStatus() {
        return executionStatus;
    }

    public void setExecutionStatus(RuleExecutionStatus executionStatus) {
        this.executionStatus = executionStatus;
    }

    public RuleNonExecutionCause getNonExecutionCause() {
        return nonExecutionCause;
    }

    public void setNonExecutionCause(RuleNonExecutionCause nonExecutionCause) {
        this.nonExecutionCause = nonExecutionCause;
    }

    public Integer getExecutionRetries() {
        return executionRetries;
    }

    public void setExecutionRetries(Integer executionRetries) {
        this.executionRetries = executionRetries;
    }
}

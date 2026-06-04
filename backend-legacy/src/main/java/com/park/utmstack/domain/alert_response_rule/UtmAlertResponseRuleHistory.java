package com.park.utmstack.domain.alert_response_rule;


import com.google.gson.GsonBuilder;
import com.google.gson.JsonSerializer;
import com.park.utmstack.service.dto.UtmAlertResponseRuleDTO;
import org.springframework.data.annotation.CreatedDate;
import org.springframework.data.jpa.domain.support.AuditingEntityListener;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import java.io.Serializable;
import java.time.Instant;

@Entity
@Table(name = "utm_alert_response_rule_history")
@EntityListeners(AuditingEntityListener.class)
public class UtmAlertResponseRuleHistory implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @NotNull
    @Column(name = "rule_id", nullable = false)
    private Long ruleId;

    @NotNull
    @CreatedDate
    @Column(name = "created_date", nullable = false)
    private Instant createdDate;

    @NotNull
    @Column(name = "previous_state", nullable = false)
    private String previousState;

    public UtmAlertResponseRuleHistory() {
    }

    public UtmAlertResponseRuleHistory(UtmAlertResponseRuleDTO rule) {
        this.ruleId = rule.getId();
        this.previousState = new GsonBuilder()
            .registerTypeAdapter(Instant.class, (JsonSerializer<Instant>) (src, typeOfSrc, context) -> context.serialize(src.toString()))
            .create().toJson(rule);
    }

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

    public Instant getCreatedDate() {
        return createdDate;
    }

    public void setCreatedDate(Instant createdDate) {
        this.createdDate = createdDate;
    }

    public String getPreviousState() {
        return previousState;
    }

    public void setPreviousState(String previousState) {
        this.previousState = previousState;
    }
}

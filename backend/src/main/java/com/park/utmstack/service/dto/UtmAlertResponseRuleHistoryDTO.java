package com.park.utmstack.service.dto;

import com.google.gson.GsonBuilder;
import com.google.gson.JsonDeserializer;
import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseRuleHistory;

import java.time.Instant;

public class UtmAlertResponseRuleHistoryDTO {
    private Long id;
    private Long ruleId;
    private Instant createdDate;
    private UtmAlertResponseRuleDTO previousState;

    public UtmAlertResponseRuleHistoryDTO() {
    }

    public UtmAlertResponseRuleHistoryDTO(UtmAlertResponseRuleHistory history) {
        this.id = history.getId();
        this.ruleId = history.getRuleId();
        this.createdDate = history.getCreatedDate();
        this.previousState = new GsonBuilder()
            .registerTypeAdapter(Instant.class, (JsonDeserializer<Instant>) (jsonElement, type, jsonDeserializationContext) ->
                Instant.parse(jsonElement.getAsString())).create().fromJson(history.getPreviousState(), UtmAlertResponseRuleDTO.class);
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

    public UtmAlertResponseRuleDTO getPreviousState() {
        return previousState;
    }

    public void setPreviousState(UtmAlertResponseRuleDTO previousState) {
        this.previousState = previousState;
    }
}

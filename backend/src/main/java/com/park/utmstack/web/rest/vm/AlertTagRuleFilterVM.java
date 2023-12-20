package com.park.utmstack.web.rest.vm;

import java.util.List;

public class AlertTagRuleFilterVM {
    private Long id;
    private String name;
    private String conditionField;
    private String conditionValue;
    private Boolean ruleActive;
    private Boolean ruleDeleted;
    private List<Long> tagIds;

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

    public String getConditionField() {
        return conditionField;
    }

    public void setConditionField(String conditionField) {
        this.conditionField = conditionField;
    }

    public String getConditionValue() {
        return conditionValue;
    }

    public void setConditionValue(String conditionValue) {
        this.conditionValue = conditionValue;
    }

    public List<Long> getTagIds() {
        return tagIds;
    }

    public void setTagIds(List<Long> tagIds) {
        this.tagIds = tagIds;
    }

    public Boolean getRuleActive() {
        return ruleActive;
    }

    public void setRuleActive(Boolean ruleActive) {
        this.ruleActive = ruleActive;
    }

    public Boolean getRuleDeleted() {
        return ruleDeleted;
    }

    public void setRuleDeleted(Boolean ruleDeleted) {
        this.ruleDeleted = ruleDeleted;
    }
}

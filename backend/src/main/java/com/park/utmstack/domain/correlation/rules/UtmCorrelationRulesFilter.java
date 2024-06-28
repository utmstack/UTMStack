package com.park.utmstack.domain.correlation.rules;

import java.time.Instant;
import java.util.List;


public class UtmCorrelationRulesFilter {

    private String ruleName;
    private List<Integer> ruleConfidentiality;
    private List<Integer> ruleIntegrity;
    private List<Integer> ruleAvailability;
    private List<String> ruleCategory;
    private List<String> ruleTechnique;
    private Instant ruleInitDate;
    private Instant ruleEndDate;
    private List<Boolean> ruleActive;
    private List<Boolean> systemOwner;
    private List<Long> dataTypes;

    public String getRuleName() {
        return ruleName;
    }

    public void setRuleName(String ruleName) {
        this.ruleName = ruleName;
    }

    public List<Integer> getRuleConfidentiality() {
        return ruleConfidentiality;
    }

    public void setRuleConfidentiality(List<Integer> ruleConfidentiality) {
        this.ruleConfidentiality = ruleConfidentiality;
    }

    public List<Integer> getRuleIntegrity() {
        return ruleIntegrity;
    }

    public void setRuleIntegrity(List<Integer> ruleIntegrity) {
        this.ruleIntegrity = ruleIntegrity;
    }

    public List<Integer> getRuleAvailability() {
        return ruleAvailability;
    }

    public void setRuleAvailability(List<Integer> ruleAvailability) {
        this.ruleAvailability = ruleAvailability;
    }

    public List<String> getRuleCategory() {
        return ruleCategory;
    }

    public void setRuleCategory(List<String> ruleCategory) {
        this.ruleCategory = ruleCategory;
    }

    public List<String> getRuleTechnique() {
        return ruleTechnique;
    }

    public void setRuleTechnique(List<String> ruleTechnique) {
        this.ruleTechnique = ruleTechnique;
    }

    public Instant getRuleInitDate() {
        return ruleInitDate;
    }

    public void setRuleInitDate(Instant ruleInitDate) {
        this.ruleInitDate = ruleInitDate;
    }

    public Instant getRuleEndDate() {
        return ruleEndDate;
    }

    public void setRuleEndDate(Instant ruleEndDate) {
        this.ruleEndDate = ruleEndDate;
    }

    public List<Boolean> getRuleActive() {
        return ruleActive;
    }

    public void setRuleActive(List<Boolean> ruleActive) {
        this.ruleActive = ruleActive;
    }

    public List<Boolean> getSystemOwner() {
        return systemOwner;
    }

    public void setSystemOwner(List<Boolean> systemOwner) {
        this.systemOwner = systemOwner;
    }

    public List<Long> getDataTypes() {
        return dataTypes;
    }

    public void setDataTypes(List<Long> dataTypes) {
        this.dataTypes = dataTypes;
    }
}

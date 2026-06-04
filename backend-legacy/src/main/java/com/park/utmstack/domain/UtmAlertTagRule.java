package com.park.utmstack.domain;


import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.web.rest.vm.AlertTagRuleVM;
import org.springframework.util.StringUtils;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import java.io.Serializable;
import java.util.Arrays;
import java.util.List;
import java.util.stream.Collectors;

/**
 * A UtmTagRule.
 */
@Entity
@Table(name = "utm_alert_tag_rule")
public class UtmAlertTagRule extends AbstractAuditingEntity implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @NotNull
    @Column(name = "rule_name", nullable = false, unique = true)
    private String ruleName;

    @NotNull
    @Column(name = "rule_description", nullable = false)
    private String ruleDescription;

    @NotNull
    @Column(name = "rule_conditions", nullable = false)
    private String ruleConditions;

    @NotNull
    @Column(name = "rule_applied_tags", nullable = false)
    private String ruleAppliedTags;

    @Transient
    private List<UtmAlertTag> tags;

    @NotNull
    @Column(name = "rule_active", nullable = false)
    private Boolean ruleActive;

    @NotNull
    @Column(name = "rule_deleted", nullable = false)
    private Boolean ruleDeleted;

    @Transient
    @JsonIgnore
    private List<FilterType> conditions;

    @Transient
    @JsonIgnore
    private String appliedTagsAsStrip;

    @Transient
    @JsonIgnore
    private List<String> appliedTagsAsList;

    public UtmAlertTagRule() {
    }

    public UtmAlertTagRule(AlertTagRuleVM ruleVM) throws JsonProcessingException {
        this.id = ruleVM.getId();
        this.ruleName = ruleVM.getName();
        this.ruleDescription = ruleVM.getDescription();
        this.ruleAppliedTags = ruleVM.getTags().stream().map(tag -> String.valueOf(tag.getId()))
            .collect(Collectors.joining(","));
        this.ruleActive = ruleVM.getActive();
        this.ruleDeleted = ruleVM.getDeleted();

        ObjectMapper om = new ObjectMapper();
        this.ruleConditions = om.writeValueAsString(ruleVM.getConditions());
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

    public UtmAlertTagRule ruleName(String ruleName) {
        this.ruleName = ruleName;
        return this;
    }

    public void setRuleName(String ruleName) {
        this.ruleName = ruleName;
    }

    public String getRuleDescription() {
        return ruleDescription;
    }

    public UtmAlertTagRule ruleDescription(String ruleDescription) {
        this.ruleDescription = ruleDescription;
        return this;
    }

    public void setRuleDescription(String ruleDescription) {
        this.ruleDescription = ruleDescription;
    }

    public String getRuleConditions() {
        return ruleConditions;
    }

    public UtmAlertTagRule ruleConditions(String ruleConditions) {
        this.ruleConditions = ruleConditions;
        return this;
    }

    public void setRuleConditions(String ruleConditions) {
        this.ruleConditions = ruleConditions;
    }

    public String getRuleAppliedTags() {
        return ruleAppliedTags;
    }

    public UtmAlertTagRule ruleAppliedTags(String ruleAppliedTags) {
        this.ruleAppliedTags = ruleAppliedTags;
        return this;
    }

    public void setRuleAppliedTags(String ruleAppliedTags) {
        this.ruleAppliedTags = ruleAppliedTags;
    }

    public List<FilterType> getConditions() throws JsonProcessingException {
        if (StringUtils.hasText(ruleConditions)) {
            ObjectMapper om = new ObjectMapper();
            conditions = om.readValue(ruleConditions, new TypeReference<>() {
            });
        }
        return conditions;
    }

    public List<Long> getAppliedTagsAsListOfLong() {
        return StringUtils.hasText(ruleAppliedTags) ? Arrays.stream(ruleAppliedTags.split(","))
            .map(Long::parseLong).collect(Collectors.toList()) : null;
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

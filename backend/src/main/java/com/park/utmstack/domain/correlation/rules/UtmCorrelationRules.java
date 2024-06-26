package com.park.utmstack.domain.correlation.rules;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.fasterxml.jackson.databind.annotation.JsonSerialize;
import com.park.utmstack.util.UtilSerializer;
import com.park.utmstack.util.exceptions.UtmSerializationException;
import org.hibernate.annotations.GenericGenerator;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import javax.persistence.*;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.time.Clock;
import java.time.Instant;
import java.util.List;

/**
 * Correlation rules entity template.
 */
@Entity
@Table(name = "utm_correlation_rules")
public class UtmCorrelationRules implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GenericGenerator(name = "CustomIdentityGenerator", strategy = "com.park.utmstack.util.CustomIdentityGenerator")
    @GeneratedValue(generator = "CustomIdentityGenerator")
    @Column(name = "id", updatable = false)
    private Long id;

    @Size(max = 250)
    @Column(name = "rule_name", length = 250, nullable = false)
    private String ruleName;

    @Column(name = "rule_confidentiality", nullable = false)
    private Integer ruleConfidentiality;

    @Column(name = "rule_integrity", nullable = false)
    private Integer ruleIntegrity;

    @Column(name = "rule_availability", nullable = false)
    private Integer ruleAvailability;

    @Size(max = 250)
    @Column(name = "rule_category", length = 250, nullable = false)
    private String ruleCategory;

    @Size(max = 500)
    @Column(name = "rule_technique", length = 500, nullable = false)
    private String ruleTechnique;

    @Column(name = "rule_description")
    private String ruleDescription;

    @JsonIgnore
    @Column(name = "rule_references_def")
    private String ruleReferencesDef;

    @Transient
    @JsonSerialize
    @JsonDeserialize
    private List<String> ruleReferences;

    @Column(name = "rule_definition", nullable = false)
    private String ruleDefinition;

    @Column(name = "rule_last_update")
    private Instant ruleLastUpdate;

    @Column(name = "rule_active", nullable = false)
    private Boolean ruleActive;

    @Column(name = "system_owner", nullable = false)
    private Boolean systemOwner;

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

    public Integer getRuleConfidentiality() {
        return ruleConfidentiality;
    }

    public void setRuleConfidentiality(Integer ruleConfidentiality) {
        this.ruleConfidentiality = ruleConfidentiality;
    }

    public Integer getRuleIntegrity() {
        return ruleIntegrity;
    }

    public void setRuleIntegrity(Integer ruleIntegrity) {
        this.ruleIntegrity = ruleIntegrity;
    }

    public Integer getRuleAvailability() {
        return ruleAvailability;
    }

    public void setRuleAvailability(Integer ruleAvailability) {
        this.ruleAvailability = ruleAvailability;
    }

    public String getRuleCategory() {
        return ruleCategory;
    }

    public void setRuleCategory(String ruleCategory) {
        this.ruleCategory = ruleCategory;
    }

    public String getRuleTechnique() {
        return ruleTechnique;
    }

    public void setRuleTechnique(String ruleTechnique) {
        this.ruleTechnique = ruleTechnique;
    }

    public String getRuleDescription() {
        return ruleDescription;
    }

    public void setRuleDescription(String ruleDescription) {
        this.ruleDescription = ruleDescription;
    }

    public String getRuleReferencesDef() {
        return ruleReferencesDef;
    }

    public void setRuleReferencesDef(String ruleReferencesDef) {
        this.ruleReferencesDef = ruleReferencesDef;
    }

    public List<String> getRuleReferences() throws UtmSerializationException {
        if (StringUtils.hasText(ruleReferencesDef))
            ruleReferences = UtilSerializer.jsonDeserializeList(String.class, ruleReferencesDef);
        return ruleReferences;
    }

    public void setRuleReferences(List<String> ruleReferences) throws UtmSerializationException {
        if (CollectionUtils.isEmpty(ruleReferences))
            this.ruleReferencesDef = null;
        else
            this.ruleReferencesDef = UtilSerializer.jsonSerialize(ruleReferences);
        this.ruleReferences = ruleReferences;
    }

    public String getRuleDefinition() {
        return ruleDefinition;
    }

    public void setRuleDefinition(String ruleDefinition) {
        this.ruleDefinition = ruleDefinition;
    }

    public Instant getRuleLastUpdate() {
        return ruleLastUpdate;
    }

    public void setRuleLastUpdate() {
        this.ruleLastUpdate = Instant.now(Clock.systemUTC());
    }

    public Boolean getSystemOwner() {
        return systemOwner;
    }

    public Boolean getRuleActive() {
        return ruleActive;
    }

    public void setRuleActive(Boolean ruleActive) {
        this.ruleActive = ruleActive;
    }

    public void setSystemOwner(Boolean systemOwner) {
        this.systemOwner = systemOwner == null ? Boolean.FALSE : systemOwner;
    }
}

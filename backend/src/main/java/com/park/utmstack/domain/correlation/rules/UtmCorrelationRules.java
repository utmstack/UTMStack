package com.park.utmstack.domain.correlation.rules;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.fasterxml.jackson.databind.annotation.JsonSerialize;
import com.park.utmstack.domain.correlation.config.UtmDataTypes;
import com.park.utmstack.service.dto.correlation.AdversaryType;
import com.park.utmstack.util.UtilSerializer;
import com.park.utmstack.util.exceptions.UtmSerializationException;
import lombok.AccessLevel;
import lombok.Getter;
import lombok.Setter;
import org.hibernate.annotations.GenericGenerator;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import javax.persistence.*;
import javax.validation.constraints.Max;
import javax.validation.constraints.Min;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.time.Instant;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

/**
 * Correlation rules entity template.
 */
@Entity
@Getter
@Setter
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

    @Enumerated(EnumType.STRING)
    @Column(name = "rule_adversary", length = 25, nullable = false)
    private AdversaryType ruleAdversary;

    @Min(value = 0)
    @Max(value = 3)
    @Column(name = "rule_confidentiality", nullable = false)
    private Integer ruleConfidentiality;

    @Min(value = 0)
    @Max(value = 3)
    @Column(name = "rule_integrity", nullable = false)
    private Integer ruleIntegrity;

    @Min(value = 0)
    @Max(value = 3)
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
    @Getter(AccessLevel.NONE)
    @Setter(AccessLevel.NONE)
    private List<String> ruleReferences;

    @JsonIgnore
    @Column(name = "rule_definition_def", nullable = false)
    private String ruleDefinitionDef;

    @Transient
    @JsonSerialize
    @JsonDeserialize
    @Getter(AccessLevel.NONE)
    @Setter(AccessLevel.NONE)
    private RuleDefinition ruleDefinition;

    @Column(name = "rule_last_update")
    private Instant ruleLastUpdate;

    @Column(name = "rule_active", nullable = false)
    private Boolean ruleActive;

    @Column(name = "system_owner", nullable = false)
    private Boolean systemOwner;

    @ManyToMany(fetch = FetchType.EAGER)
    @JoinTable(name = "utm_group_rules_data_type",
            joinColumns = @JoinColumn(name = "rule_id"),
            inverseJoinColumns = @JoinColumn(name = "data_type_id"))
    private Set<UtmDataTypes> dataTypes = new HashSet<>();

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


    public RuleDefinition getRuleDefinition() throws UtmSerializationException {
        if (StringUtils.hasText(ruleDefinitionDef))
            ruleDefinition = UtilSerializer.jsonDeserialize(RuleDefinition.class, ruleDefinitionDef);
        return ruleDefinition;
    }

    public void setRuleDefinition(RuleDefinition ruleDefinition) throws UtmSerializationException {
        if (ruleDefinition == null)
            this.ruleDefinitionDef = null;
        else
            this.ruleDefinitionDef = UtilSerializer.jsonSerialize(ruleDefinition);

        this.ruleDefinition = ruleDefinition;
    }
}

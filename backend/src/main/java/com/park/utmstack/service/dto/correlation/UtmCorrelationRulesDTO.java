package com.park.utmstack.service.dto.correlation;

import com.park.utmstack.domain.correlation.config.UtmDataTypes;
import com.park.utmstack.domain.correlation.rules.RuleDefinition;
import org.hibernate.validator.constraints.URL;

import javax.validation.constraints.Max;
import javax.validation.constraints.Min;
import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotEmpty;
import java.io.Serializable;
import java.util.List;
import java.util.Set;

public class UtmCorrelationRulesDTO implements Serializable {

    private static final long serialVersionUID = 1L;

    private Long id;
    @NotBlank
    private String name;
    @Min(value = 0)
    @Max(value = 3)
    private Integer confidentiality;
    @Min(value = 0)
    @Max(value = 3)
    private Integer integrity;
    @Min(value = 0)
    @Max(value = 3)
    private Integer availability;
    @NotBlank
    private String category;
    @NotBlank
    private String technique;
    private String description;
    private List<@URL(message = "Reference must be a valid URL ") String>references;
    @NotEmpty
    private Set<UtmDataTypes> dataTypes;
    private RuleDefinition definition;
    private Boolean systemOwner;

    private Boolean ruleActive;

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

    public Integer getConfidentiality() {
        return confidentiality;
    }

    public void setConfidentiality(Integer confidentiality) {
        this.confidentiality = confidentiality;
    }

    public Integer getIntegrity() {
        return integrity;
    }

    public void setIntegrity(Integer integrity) {
        this.integrity = integrity;
    }

    public Integer getAvailability() {
        return availability;
    }

    public void setAvailability(Integer availability) {
        this.availability = availability;
    }

    public String getCategory() {
        return category;
    }

    public void setCategory(String category) {
        this.category = category;
    }

    public String getTechnique() {
        return technique;
    }

    public void setTechnique(String technique) {
        this.technique = technique;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public List<String> getReferences() {
        return references;
    }

    public void setReferences(List<String> rReferences) {
        this.references = rReferences;
    }

    public Set<UtmDataTypes> getDataTypes() {
        return dataTypes;
    }

    public void setDataTypes(Set<UtmDataTypes> dataTypes) {
        this.dataTypes = dataTypes;
    }

    public RuleDefinition getDefinition() {
        return definition;
    }

    public void setDefinition(RuleDefinition definition) {
        this.definition = definition;
    }

    public Boolean getSystemOwner() {
        return systemOwner;
    }

    public void setSystemOwner(Boolean systemOwner) {
        this.systemOwner = systemOwner;
    }

    public Boolean getRuleActive() {
        return ruleActive;
    }

    public void setRuleActive(Boolean ruleActive) {
        this.ruleActive = ruleActive;
    }
}


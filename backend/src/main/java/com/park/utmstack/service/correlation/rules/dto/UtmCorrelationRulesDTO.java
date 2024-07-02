package com.park.utmstack.service.correlation.rules.dto;

import com.park.utmstack.domain.correlation.config.UtmDataTypes;
import com.park.utmstack.domain.correlation.config.UtmVariable;

import java.io.Serializable;
import java.util.List;
import java.util.Set;

public class UtmCorrelationRulesDTO implements Serializable {

    private static final long serialVersionUID = 1L;

    private Long id;
    private String name;
    private Integer confidentiality;
    private Integer integrity;
    private Integer availability;
    private String category;
    private String technique;
    private String description;
    private List<String>references;
    private String expression;
    private Set<UtmDataTypes> dataTypes;
    private List<UtmVariable> variables;

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

    public String getExpression() {
        return expression;
    }

    public void setExpression(String expression) {
        this.expression = expression;
    }

    public Set<UtmDataTypes> getDataTypes() {
        return dataTypes;
    }

    public void setDataTypes(Set<UtmDataTypes> dataTypes) {
        this.dataTypes = dataTypes;
    }

    public List<UtmVariable> getVariables() {
        return variables;
    }

    public void setVariables(List<UtmVariable> variables) {
        this.variables = variables;
    }
}


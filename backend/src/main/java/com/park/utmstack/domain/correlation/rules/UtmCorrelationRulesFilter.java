package com.park.utmstack.domain.correlation.rules;

import java.time.Instant;
import java.util.List;


public class UtmCorrelationRulesFilter {

    private String name;
    private List<Integer> confidentiality;
    private List<Integer> integrity;
    private List<Integer> availability;
    private List<String> category;
    private List<String> technique;
    private Instant initDate;
    private Instant endDate;
    private List<Boolean> active;
    private List<Boolean> systemOwner;
    private List<String> dataTypes;

    private String search;

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public List<Integer> getConfidentiality() {
        return confidentiality;
    }

    public void setConfidentiality(List<Integer> confidentiality) {
        this.confidentiality = confidentiality;
    }

    public List<Integer> getIntegrity() {
        return integrity;
    }

    public void setIntegrity(List<Integer> integrity) {
        this.integrity = integrity;
    }

    public List<Integer> getAvailability() {
        return availability;
    }

    public void setAvailability(List<Integer> availability) {
        this.availability = availability;
    }

    public List<String> getCategory() {
        return category;
    }

    public void setCategory(List<String> category) {
        this.category = category;
    }

    public List<String> getTechnique() {
        return technique;
    }

    public void setTechnique(List<String> technique) {
        this.technique = technique;
    }

    public Instant getInitDate() {
        return initDate;
    }

    public void setInitDate(Instant initDate) {
        this.initDate = initDate;
    }

    public Instant getEndDate() {
        return endDate;
    }

    public void setEndDate(Instant endDate) {
        this.endDate = endDate;
    }

    public List<Boolean> getActive() {
        return active;
    }

    public void setActive(List<Boolean> active) {
        this.active = active;
    }

    public List<Boolean> getSystemOwner() {
        return systemOwner;
    }

    public void setSystemOwner(List<Boolean> systemOwner) {
        this.systemOwner = systemOwner;
    }

    public List<String> getDataTypes() {
        return dataTypes;
    }

    public void setDataTypes(List<String> dataTypes) {
        this.dataTypes = dataTypes;
    }

    public String getSearch() {
        return search;
    }

    public void setSearch(String search) {
        this.search = search;
    }
}

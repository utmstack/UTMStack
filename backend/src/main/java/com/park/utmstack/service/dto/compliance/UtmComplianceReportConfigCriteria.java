package com.park.utmstack.service.dto.compliance;

import tech.jhipster.service.filter.BooleanFilter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

public class UtmComplianceReportConfigCriteria {
    private LongFilter id;
    private StringFilter configSolution;
    private LongFilter standardSectionId;
    private BooleanFilter configReportEditable;
    private BooleanFilter expandDashboard;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public StringFilter getConfigSolution() {
        return configSolution;
    }

    public void setConfigSolution(StringFilter configSolution) {
        this.configSolution = configSolution;
    }

    public LongFilter getStandardSectionId() {
        return standardSectionId;
    }

    public void setStandardSectionId(LongFilter standardSectionId) {
        this.standardSectionId = standardSectionId;
    }

    public BooleanFilter getConfigReportEditable() {
        return configReportEditable;
    }

    public void setConfigReportEditable(BooleanFilter configReportEditable) {
        this.configReportEditable = configReportEditable;
    }

    public BooleanFilter getExpandDashboard() {
        return expandDashboard;
    }

    public void setExpandDashboard(BooleanFilter expandDashboard) {
        this.expandDashboard = expandDashboard;
    }
}

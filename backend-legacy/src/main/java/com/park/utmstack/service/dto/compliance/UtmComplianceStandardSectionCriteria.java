package com.park.utmstack.service.dto.compliance;

import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

public class UtmComplianceStandardSectionCriteria {
    private LongFilter id;
    private StringFilter standardSectionName;
    private LongFilter standardId;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public StringFilter getStandardSectionName() {
        return standardSectionName;
    }

    public void setStandardSectionName(StringFilter standardSectionName) {
        this.standardSectionName = standardSectionName;
    }

    public LongFilter getStandardId() {
        return standardId;
    }

    public void setStandardId(LongFilter standardId) {
        this.standardId = standardId;
    }
}

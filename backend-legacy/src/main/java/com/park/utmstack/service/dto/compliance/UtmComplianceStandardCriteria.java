package com.park.utmstack.service.dto.compliance;

import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

public class UtmComplianceStandardCriteria {

    private LongFilter id;
    private StringFilter standardName;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public StringFilter getStandardName() {
        return standardName;
    }

    public void setStandardName(StringFilter standardName) {
        this.standardName = standardName;
    }
}

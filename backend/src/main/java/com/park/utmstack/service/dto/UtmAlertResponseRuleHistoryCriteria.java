package com.park.utmstack.service.dto;

import tech.jhipster.service.filter.InstantFilter;
import tech.jhipster.service.filter.LongFilter;

import java.io.Serializable;

public class UtmAlertResponseRuleHistoryCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;

    private LongFilter ruleId;

    private InstantFilter createdDate;


    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public LongFilter getRuleId() {
        return ruleId;
    }

    public void setRuleId(LongFilter ruleId) {
        this.ruleId = ruleId;
    }

    public InstantFilter getCreatedDate() {
        return createdDate;
    }

    public void setCreatedDate(InstantFilter createdDate) {
        this.createdDate = createdDate;
    }
}

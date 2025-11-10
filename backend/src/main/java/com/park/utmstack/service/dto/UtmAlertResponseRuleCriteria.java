package com.park.utmstack.service.dto;

import lombok.Getter;
import lombok.Setter;
import tech.jhipster.service.filter.BooleanFilter;
import tech.jhipster.service.filter.InstantFilter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;

@Getter
@Setter
public class UtmAlertResponseRuleCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;
    private StringFilter name;
    private BooleanFilter active;
    private StringFilter agentPlatform;
    private StringFilter createdBy;
    private StringFilter lastModifiedBy;
    private InstantFilter createdDate;
    private InstantFilter lastModifiedDate;
    private BooleanFilter systemOwner;
}

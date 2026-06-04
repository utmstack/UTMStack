package com.park.utmstack.service.dto;

import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.ToString;
import tech.jhipster.service.filter.BooleanFilter;
import tech.jhipster.service.filter.InstantFilter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;

@Data
@EqualsAndHashCode
@ToString
public class UtmAlertResponseActionTemplateCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;
    private StringFilter label;
    private StringFilter description;
    private StringFilter command;
    private BooleanFilter systemOwner;

}

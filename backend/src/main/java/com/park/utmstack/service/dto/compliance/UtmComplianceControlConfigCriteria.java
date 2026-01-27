package com.park.utmstack.service.dto.compliance;

import lombok.Getter;
import lombok.Setter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

@Getter
@Setter
public class UtmComplianceControlConfigCriteria {
    private LongFilter id;
    private LongFilter standardSectionId;
    private StringFilter controlName;
    private StringFilter controlSolution;
    private StringFilter controlRemediation;
    private ComplianceStrategyFilter controlStrategy;
}
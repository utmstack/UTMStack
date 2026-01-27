package com.park.utmstack.service.compliance.config;

import com.park.utmstack.domain.compliance.*;
import com.park.utmstack.repository.compliance.UtmComplianceControlConfigRepository;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigCriteria;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigDto;
import com.park.utmstack.service.mapper.compliance.UtmComplianceControlConfigMapper;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.domain.Specification;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import tech.jhipster.service.QueryService;

@Service
@Transactional(readOnly = true)
public class UtmComplianceControlConfigQueryService extends QueryService<UtmComplianceControlConfig> {

    private final UtmComplianceControlConfigRepository complianceControlConfigRepository;
    private final UtmComplianceControlConfigMapper mapper;

    public UtmComplianceControlConfigQueryService(UtmComplianceControlConfigRepository complianceControlConfigRepository,
                                                  UtmComplianceControlConfigMapper mapper) {
        this.complianceControlConfigRepository = complianceControlConfigRepository;
        this.mapper = mapper;
    }

    @Transactional(readOnly = true)
    public Page<UtmComplianceControlConfigDto> findByCriteria(UtmComplianceControlConfigCriteria criteria, Pageable page) {
        final Specification<UtmComplianceControlConfig> specification = createSpecification(criteria);
        return complianceControlConfigRepository.findAll(specification, page).map(mapper::toDto);
    }

    private Specification<UtmComplianceControlConfig> createSpecification(UtmComplianceControlConfigCriteria criteria) {
        Specification<UtmComplianceControlConfig> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmComplianceControlConfig_.id));
            }
            if (criteria.getStandardSectionId() != null) {
                specification = specification.and(
                        buildRangeSpecification(criteria.getStandardSectionId(), UtmComplianceControlConfig_.standardSectionId));
            }
            if (criteria.getControlName() != null) {
                specification = specification.and(
                        buildStringSpecification(criteria.getControlName(), UtmComplianceControlConfig_.controlName));
            }
            if (criteria.getControlSolution() != null ) {
                specification = specification.and(
                        buildStringSpecification(criteria.getControlSolution(), UtmComplianceControlConfig_.controlSolution));
            }
            if (criteria.getControlRemediation() != null) {
                specification = specification.and(
                        buildStringSpecification(criteria.getControlRemediation(), UtmComplianceControlConfig_.controlRemediation));
            }
            if (criteria.getControlStrategy() != null) {
                specification = specification.and(
                        buildSpecification(criteria.getControlStrategy(), UtmComplianceControlConfig_.controlStrategy)
                );
            }

        }
        return specification;
    }
}
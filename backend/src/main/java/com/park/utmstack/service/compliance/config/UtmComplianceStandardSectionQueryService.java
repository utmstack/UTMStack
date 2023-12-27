package com.park.utmstack.service.compliance.config;

import com.park.utmstack.domain.compliance.UtmComplianceStandardSection;
import com.park.utmstack.domain.compliance.UtmComplianceStandardSection_;
import com.park.utmstack.repository.compliance.UtmComplianceStandardSectionRepository;
import com.park.utmstack.service.dto.compliance.UtmComplianceStandardSectionCriteria;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.domain.Specification;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import tech.jhipster.service.QueryService;

import java.util.List;

@Service
@Transactional(readOnly = true)
public class UtmComplianceStandardSectionQueryService extends QueryService<UtmComplianceStandardSection> {
    private final Logger log = LoggerFactory.getLogger(UtmComplianceStandardSectionQueryService.class);

    private final UtmComplianceStandardSectionRepository complianceStandardSectionRepository;

    public UtmComplianceStandardSectionQueryService(
        UtmComplianceStandardSectionRepository complianceStandardSectionRepository) {
        this.complianceStandardSectionRepository = complianceStandardSectionRepository;
    }


    @Transactional(readOnly = true)
    public List<UtmComplianceStandardSection> findByCriteria(UtmComplianceStandardSectionCriteria criteria) {
        final Specification<UtmComplianceStandardSection> specification = createSpecification(criteria);
        return complianceStandardSectionRepository.findAll(specification);
    }

    @Transactional(readOnly = true)
    public Page<UtmComplianceStandardSection> findByCriteria(UtmComplianceStandardSectionCriteria criteria, Pageable page) {
        final Specification<UtmComplianceStandardSection> specification = createSpecification(criteria);
        return complianceStandardSectionRepository.findAll(specification, page);
    }

    private Specification<UtmComplianceStandardSection> createSpecification(UtmComplianceStandardSectionCriteria criteria) {
        Specification<UtmComplianceStandardSection> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmComplianceStandardSection_.id));
            }
            if (criteria.getStandardId() != null) {
                specification = specification.and(
                    buildRangeSpecification(criteria.getStandardId(), UtmComplianceStandardSection_.standardId));
            }
            if (criteria.getStandardSectionName() != null) {
                specification = specification.and(buildStringSpecification(criteria.getStandardSectionName(),
                                                                           UtmComplianceStandardSection_.standardSectionName));
            }
        }
        return specification;
    }
}

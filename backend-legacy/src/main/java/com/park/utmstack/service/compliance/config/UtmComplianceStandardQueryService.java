package com.park.utmstack.service.compliance.config;

import com.park.utmstack.domain.compliance.UtmComplianceStandard;
import com.park.utmstack.domain.compliance.UtmComplianceStandard_;
import com.park.utmstack.repository.compliance.UtmComplianceStandardRepository;
import com.park.utmstack.service.dto.compliance.UtmComplianceStandardCriteria;
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
public class UtmComplianceStandardQueryService extends QueryService<UtmComplianceStandard> {
    private final Logger log = LoggerFactory.getLogger(UtmComplianceStandardQueryService.class);

    private final UtmComplianceStandardRepository complianceStandardRepository;

    public UtmComplianceStandardQueryService(UtmComplianceStandardRepository complianceStandardRepository) {
        this.complianceStandardRepository = complianceStandardRepository;
    }


    @Transactional(readOnly = true)
    public List<UtmComplianceStandard> findByCriteria(UtmComplianceStandardCriteria criteria) {
        final Specification<UtmComplianceStandard> specification = createSpecification(criteria);
        return complianceStandardRepository.findAll(specification);
    }

    @Transactional(readOnly = true)
    public Page<UtmComplianceStandard> findByCriteria(UtmComplianceStandardCriteria criteria, Pageable page) {
        final Specification<UtmComplianceStandard> specification = createSpecification(criteria);
        return complianceStandardRepository.findAll(specification, page);
    }

    private Specification<UtmComplianceStandard> createSpecification(UtmComplianceStandardCriteria criteria) {
        Specification<UtmComplianceStandard> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmComplianceStandard_.id));
            }
            if (criteria.getStandardName() != null) {
                specification = specification.and(
                    buildStringSpecification(criteria.getStandardName(), UtmComplianceStandard_.standardName));
            }
        }
        return specification;
    }
}

package com.park.utmstack.service.compliance.config;

import com.park.utmstack.domain.compliance.UtmComplianceReportConfig;
import com.park.utmstack.domain.compliance.UtmComplianceReportConfig_;
import com.park.utmstack.repository.compliance.UtmComplianceReportConfigRepository;
import com.park.utmstack.service.dto.compliance.UtmComplianceReportConfigCriteria;
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
public class UtmComplianceReportConfigQueryService extends QueryService<UtmComplianceReportConfig> {
    private final Logger log = LoggerFactory.getLogger(UtmComplianceReportConfigQueryService.class);

    private final UtmComplianceReportConfigRepository complianceReportConfigRepository;

    public UtmComplianceReportConfigQueryService(UtmComplianceReportConfigRepository complianceReportConfigRepository) {
        this.complianceReportConfigRepository = complianceReportConfigRepository;
    }

    @Transactional(readOnly = true)
    public List<UtmComplianceReportConfig> findByCriteria(UtmComplianceReportConfigCriteria criteria) {
        final Specification<UtmComplianceReportConfig> specification = createSpecification(criteria);
        return complianceReportConfigRepository.findAll(specification);
    }

    @Transactional(readOnly = true)
    public Page<UtmComplianceReportConfig> findByCriteria(UtmComplianceReportConfigCriteria criteria, Pageable page) {
        final Specification<UtmComplianceReportConfig> specification = createSpecification(criteria);
        return complianceReportConfigRepository.findAll(specification, page);
    }

    private Specification<UtmComplianceReportConfig> createSpecification(UtmComplianceReportConfigCriteria criteria) {
        Specification<UtmComplianceReportConfig> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmComplianceReportConfig_.id));
            }
            if (criteria.getConfigReportEditable() != null) {
                specification = specification.and(
                    buildSpecification(criteria.getConfigReportEditable(), UtmComplianceReportConfig_.configReportEditable));
            }
            if (criteria.getConfigSolution() != null) {
                specification = specification.and(
                    buildStringSpecification(criteria.getConfigSolution(), UtmComplianceReportConfig_.configSolution));
            }
            if (criteria.getStandardSectionId() != null) {
                specification = specification.and(
                    buildRangeSpecification(criteria.getStandardSectionId(), UtmComplianceReportConfig_.standardSectionId));
            }
        }
        return specification;
    }
}

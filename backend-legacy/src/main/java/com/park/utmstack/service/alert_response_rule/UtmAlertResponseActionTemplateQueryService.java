package com.park.utmstack.service.alert_response_rule;

import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseActionTemplate;
import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseActionTemplate_;
import com.park.utmstack.repository.alert_response_rule.UtmAlertResponseActionTemplateRepository;
import com.park.utmstack.service.dto.UtmAlertResponseActionTemplateCriteria;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.domain.Specification;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import tech.jhipster.service.QueryService;

@Service
@Transactional(readOnly = true)
public class UtmAlertResponseActionTemplateQueryService extends QueryService<UtmAlertResponseActionTemplate> {
    private static final String CLASSNAME = "UtmAlertResponseActionTemplateQueryService";

    private final UtmAlertResponseActionTemplateRepository utmAlertResponseActionTemplateRepository;

    public UtmAlertResponseActionTemplateQueryService(UtmAlertResponseActionTemplateRepository utmAlertResponseActionTemplateRepository) {
        this.utmAlertResponseActionTemplateRepository = utmAlertResponseActionTemplateRepository;

    }

    @Transactional(readOnly = true)
    public Page<UtmAlertResponseActionTemplate> findByCriteria(UtmAlertResponseActionTemplateCriteria criteria, Pageable page) {
        final String ctx = CLASSNAME + ".findByCriteria";
        try {
            final Specification<UtmAlertResponseActionTemplate> specification = createSpecification(criteria);
            return utmAlertResponseActionTemplateRepository.findAll(specification, page);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private Specification<UtmAlertResponseActionTemplate> createSpecification(UtmAlertResponseActionTemplateCriteria criteria) {
        Specification<UtmAlertResponseActionTemplate> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmAlertResponseActionTemplate_.id));
            }
            if (criteria.getLabel() != null) {
                specification = specification.and(buildStringSpecification(criteria.getLabel(), UtmAlertResponseActionTemplate_.title));
            }
            if (criteria.getDescription() != null) {
                specification = specification.and(buildStringSpecification(criteria.getDescription(), UtmAlertResponseActionTemplate_.description));
            }
            if (criteria.getCommand() != null) {
                specification = specification.and(buildStringSpecification(criteria.getCommand(), UtmAlertResponseActionTemplate_.command));
            }
            if (criteria.getSystemOwner() != null) {
                specification = specification.and(buildSpecification(criteria.getSystemOwner(), UtmAlertResponseActionTemplate_.systemOwner));
            }
        }
        return specification;
    }

}

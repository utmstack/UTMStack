package com.park.utmstack.service.alert_response_rule;

import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseRule;
import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseRule_;
import com.park.utmstack.repository.alert_response_rule.UtmAlertResponseRuleRepository;
import com.park.utmstack.service.dto.UtmAlertResponseRuleCriteria;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.domain.Specification;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import tech.jhipster.service.QueryService;

@Service
@Transactional(readOnly = true)
public class UtmAlertResponseRuleQueryService extends QueryService<UtmAlertResponseRule> {
    private static final String CLASSNAME = "UtmAlertSoarRuleQueryService";

    private final UtmAlertResponseRuleRepository utmAlertResponseRuleRepository;

    public UtmAlertResponseRuleQueryService(UtmAlertResponseRuleRepository utmAlertResponseRuleRepository) {
        this.utmAlertResponseRuleRepository = utmAlertResponseRuleRepository;
    }

    @Transactional(readOnly = true)
    public Page<UtmAlertResponseRule> findByCriteria(UtmAlertResponseRuleCriteria criteria, Pageable page) {
        final String ctx = CLASSNAME + ".findByCriteria";
        try {
            final Specification<UtmAlertResponseRule> specification = createSpecification(criteria);
            return utmAlertResponseRuleRepository.findAll(specification, page);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private Specification<UtmAlertResponseRule> createSpecification(UtmAlertResponseRuleCriteria criteria) {
        Specification<UtmAlertResponseRule> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmAlertResponseRule_.id));
            }
            if (criteria.getName() != null) {
                specification = specification.and(buildStringSpecification(criteria.getName(), UtmAlertResponseRule_.ruleName));
            }
            if (criteria.getActive() != null) {
                specification = specification.and(buildSpecification(criteria.getActive(), UtmAlertResponseRule_.ruleActive));
            }
            if (criteria.getAgentPlatform() != null) {
                specification = specification.and(buildStringSpecification(criteria.getAgentPlatform(), UtmAlertResponseRule_.agentPlatform));
            }
            if (criteria.getCreatedBy() != null) {
                specification = specification.and(buildStringSpecification(criteria.getCreatedBy(), UtmAlertResponseRule_.createdBy));
            }
            if (criteria.getLastModifiedBy() != null) {
                specification = specification.and(buildStringSpecification(criteria.getLastModifiedBy(), UtmAlertResponseRule_.lastModifiedBy));
            }
            if (criteria.getCreatedDate() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getCreatedDate(), UtmAlertResponseRule_.createdDate));
            }
            if (criteria.getLastModifiedDate() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getLastModifiedDate(), UtmAlertResponseRule_.lastModifiedDate));
            }
        }
        return specification;
    }
}

package com.park.utmstack.service.alert_response_rule;

import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseRuleExecution;
import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseRuleExecution_;
import com.park.utmstack.repository.alert_response_rule.UtmAlertResponseRuleExecutionRepository;
import com.park.utmstack.service.dto.UtmAlertResponseRuleExecutionCriteria;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.domain.Specification;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import tech.jhipster.service.QueryService;

import java.util.List;

@Service
@Transactional(readOnly = true)
public class UtmAlertResponseRuleExecutionQueryService extends QueryService<UtmAlertResponseRuleExecution> {

    private static final String CLASSNAME = "UtmAlertResponseRuleExecutionQueryService";

    private final UtmAlertResponseRuleExecutionRepository ruleExecutionRepository;

    public UtmAlertResponseRuleExecutionQueryService(UtmAlertResponseRuleExecutionRepository ruleExecutionRepository) {
        this.ruleExecutionRepository = ruleExecutionRepository;
    }

    @Transactional(readOnly = true)
    public List<UtmAlertResponseRuleExecution> findByCriteria(UtmAlertResponseRuleExecutionCriteria criteria) {
        final String ctx = CLASSNAME + ".findByCriteria";
        try {
            return ruleExecutionRepository.findAll(createSpecification(criteria));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    @Transactional(readOnly = true)
    public Page<UtmAlertResponseRuleExecution> findByCriteria(UtmAlertResponseRuleExecutionCriteria criteria, Pageable page) {
        final String ctx = CLASSNAME + ".findByCriteria";
        try {
            return ruleExecutionRepository.findAll(createSpecification(criteria), page);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private Specification<UtmAlertResponseRuleExecution> createSpecification(UtmAlertResponseRuleExecutionCriteria criteria) {
        Specification<UtmAlertResponseRuleExecution> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmAlertResponseRuleExecution_.id));
            }
            if (criteria.getRuleId() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getRuleId(), UtmAlertResponseRuleExecution_.ruleId));
            }
            if (criteria.getAlertId() != null) {
                specification = specification.and(buildStringSpecification(criteria.getAlertId(), UtmAlertResponseRuleExecution_.alertId));
            }
            if (criteria.getAgent() != null) {
                specification = specification.and(buildStringSpecification(criteria.getAgent(), UtmAlertResponseRuleExecution_.agent));
            }
            if (criteria.getExecutionDate() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getExecutionDate(), UtmAlertResponseRuleExecution_.executionDate));
            }
            if (criteria.getExecutionStatus() != null) {
                specification = specification.and(buildSpecification(criteria.getExecutionStatus(), UtmAlertResponseRuleExecution_.executionStatus));
            }
            if (criteria.getNonExecutionCause() != null) {
                specification = specification.and(buildSpecification(criteria.getNonExecutionCause(), UtmAlertResponseRuleExecution_.nonExecutionCause));
            }
        }
        return specification;
    }
}

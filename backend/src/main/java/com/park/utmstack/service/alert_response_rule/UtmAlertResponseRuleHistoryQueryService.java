package com.park.utmstack.service.alert_response_rule;

import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseRuleHistory;
import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseRuleHistory_;
import com.park.utmstack.repository.alert_response_rule.UtmAlertResponseRuleHistoryRepository;
import com.park.utmstack.service.dto.UtmAlertResponseRuleHistoryCriteria;
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
public class UtmAlertResponseRuleHistoryQueryService extends QueryService<UtmAlertResponseRuleHistory> {

    private final Logger log = LoggerFactory.getLogger(UtmAlertResponseRuleHistoryQueryService.class);

    private final UtmAlertResponseRuleHistoryRepository alertResponseRuleHistoryRepository;

    public UtmAlertResponseRuleHistoryQueryService(UtmAlertResponseRuleHistoryRepository alertResponseRuleHistoryRepository) {
        this.alertResponseRuleHistoryRepository = alertResponseRuleHistoryRepository;
    }

    @Transactional(readOnly = true)
    public List<UtmAlertResponseRuleHistory> findByCriteria(UtmAlertResponseRuleHistoryCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmAlertResponseRuleHistory> specification = createSpecification(criteria);
        return alertResponseRuleHistoryRepository.findAll(specification);
    }

    @Transactional(readOnly = true)
    public Page<UtmAlertResponseRuleHistory> findByCriteria(UtmAlertResponseRuleHistoryCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmAlertResponseRuleHistory> specification = createSpecification(criteria);
        return alertResponseRuleHistoryRepository.findAll(specification, page);
    }

    private Specification<UtmAlertResponseRuleHistory> createSpecification(UtmAlertResponseRuleHistoryCriteria criteria) {
        Specification<UtmAlertResponseRuleHistory> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmAlertResponseRuleHistory_.id));
            }
            if (criteria.getRuleId() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getRuleId(), UtmAlertResponseRuleHistory_.ruleId));
            }
            if (criteria.getCreatedDate() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getCreatedDate(), UtmAlertResponseRuleHistory_.createdDate));
            }
        }
        return specification;
    }
}

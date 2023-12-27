package com.park.utmstack.service;

import com.park.utmstack.domain.UtmAlertLog;
import com.park.utmstack.domain.UtmAlertLog_;
import com.park.utmstack.repository.UtmAlertLogRepository;
import com.park.utmstack.service.dto.UtmAlertLogCriteria;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.domain.Specification;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import tech.jhipster.service.QueryService;

import java.util.List;

/**
 * Service for executing complex queries for UtmAlertLog entities in the database.
 * The main input is a {@link UtmAlertLogCriteria} which gets converted to {@link Specification},
 * in a way that all the filters must apply.
 * It returns a {@link List} of {@link UtmAlertLog} or a {@link Page} of {@link UtmAlertLog} which fulfills the criteria.
 */
@Service
@Transactional(readOnly = true)
public class UtmAlertLogQueryService extends QueryService<UtmAlertLog> {

    private final Logger log = LoggerFactory.getLogger(UtmAlertLogQueryService.class);

    private final UtmAlertLogRepository utmAlertLogRepository;

    public UtmAlertLogQueryService(UtmAlertLogRepository utmAlertLogRepository) {
        this.utmAlertLogRepository = utmAlertLogRepository;
    }

    /**
     * Return a {@link List} of {@link UtmAlertLog} which matches the criteria from the database
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmAlertLog> findByCriteria(UtmAlertLogCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmAlertLog> specification = createSpecification(criteria);
        return utmAlertLogRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmAlertLog} which matches the criteria from the database
     * @param criteria The object which holds all the filters, which the entities should match.
     * @param page The page, which should be returned.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmAlertLog> findByCriteria(UtmAlertLogCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmAlertLog> specification = createSpecification(criteria);
        return utmAlertLogRepository.findAll(specification, page);
    }

    /**
     * Return the number of matching entities in the database
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the number of matching entities.
     */
    @Transactional(readOnly = true)
    public long countByCriteria(UtmAlertLogCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<UtmAlertLog> specification = createSpecification(criteria);
        return utmAlertLogRepository.count(specification);
    }

    /**
     * Function to convert UtmAlertLogCriteria to a {@link Specification}
     */
    private Specification<UtmAlertLog> createSpecification(UtmAlertLogCriteria criteria) {
        Specification<UtmAlertLog> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmAlertLog_.id));
            }
            if (criteria.getAlertId() != null) {
                specification = specification.and(buildStringSpecification(criteria.getAlertId(), UtmAlertLog_.alertId));
            }
            if (criteria.getLogUser() != null) {
                specification = specification.and(buildStringSpecification(criteria.getLogUser(), UtmAlertLog_.logUser));
            }
            if (criteria.getLogDate() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getLogDate(), UtmAlertLog_.logDate));
            }
            if (criteria.getLogAction() != null) {
                specification = specification.and(buildStringSpecification(criteria.getLogAction(), UtmAlertLog_.logAction));
            }
            if (criteria.getLogOldValue() != null) {
                specification = specification.and(buildStringSpecification(criteria.getLogOldValue(), UtmAlertLog_.logOldValue));
            }
            if (criteria.getLogNewValue() != null) {
                specification = specification.and(buildStringSpecification(criteria.getLogNewValue(), UtmAlertLog_.logNewValue));
            }
        }
        return specification;
    }
}

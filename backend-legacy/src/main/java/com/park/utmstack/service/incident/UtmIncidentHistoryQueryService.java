package com.park.utmstack.service.incident;

import com.park.utmstack.domain.incident.UtmIncidentHistory;
import com.park.utmstack.domain.incident.UtmIncidentHistory_;
import com.park.utmstack.repository.incident.UtmIncidentHistoryRepository;
import com.park.utmstack.service.dto.incident.UtmIncidentHistoryCriteria;
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
 * Service for executing complex queries for UtmIncidentHistory entities in the database.
 * The main input is a {@link UtmIncidentHistoryCriteria} which gets converted to {@link Specification},
 * in a way that all the filters must apply.
 * It returns a {@link List} of {@link UtmIncidentHistory} or a {@link Page} of {@link UtmIncidentHistory} which fulfills the criteria.
 */
@Service
@Transactional(readOnly = true)
public class UtmIncidentHistoryQueryService extends QueryService<UtmIncidentHistory> {

    private final Logger log = LoggerFactory.getLogger(UtmIncidentHistoryQueryService.class);

    private final UtmIncidentHistoryRepository utmIncidentHistoryRepository;

    public UtmIncidentHistoryQueryService(UtmIncidentHistoryRepository utmIncidentHistoryRepository) {
        this.utmIncidentHistoryRepository = utmIncidentHistoryRepository;
    }

    /**
     * Return a {@link List} of {@link UtmIncidentHistory} which matches the criteria from the database
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmIncidentHistory> findByCriteria(UtmIncidentHistoryCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmIncidentHistory> specification = createSpecification(criteria);
        return utmIncidentHistoryRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmIncidentHistory} which matches the criteria from the database
     * @param criteria The object which holds all the filters, which the entities should match.
     * @param page The page, which should be returned.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmIncidentHistory> findByCriteria(UtmIncidentHistoryCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmIncidentHistory> specification = createSpecification(criteria);
        return utmIncidentHistoryRepository.findAll(specification, page);
    }

    /**
     * Return the number of matching entities in the database
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the number of matching entities.
     */
    @Transactional(readOnly = true)
    public long countByCriteria(UtmIncidentHistoryCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<UtmIncidentHistory> specification = createSpecification(criteria);
        return utmIncidentHistoryRepository.count(specification);
    }

    /**
     * Function to convert UtmIncidentHistoryCriteria to a {@link Specification}
     */
    private Specification<UtmIncidentHistory> createSpecification(UtmIncidentHistoryCriteria criteria) {
        Specification<UtmIncidentHistory> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmIncidentHistory_.id));
            }
            if (criteria.getIncidentId() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getIncidentId(), UtmIncidentHistory_.incidentId));
            }
            if (criteria.getActionDate() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getActionDate(), UtmIncidentHistory_.actionDate));
            }
            if (criteria.getActionType() != null) {
                specification = specification.and(buildSpecification(criteria.getActionType(), UtmIncidentHistory_.actionType));
            }
            if (criteria.getActionCreatedBy() != null) {
                specification = specification.and(buildStringSpecification(criteria.getActionCreatedBy(), UtmIncidentHistory_.actionCreatedBy));
            }
            if (criteria.getAction() != null) {
                specification = specification.and(buildStringSpecification(criteria.getAction(), UtmIncidentHistory_.action));
            }
            if (criteria.getActionDetail() != null) {
                specification = specification.and(buildStringSpecification(criteria.getActionDetail(), UtmIncidentHistory_.actionDetail));
            }
        }
        return specification;
    }
}

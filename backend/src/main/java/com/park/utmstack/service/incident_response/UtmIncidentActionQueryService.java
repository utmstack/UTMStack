package com.park.utmstack.service.incident_response;

import com.park.utmstack.domain.incident_response.UtmIncidentAction;
import com.park.utmstack.domain.incident_response.UtmIncidentAction_;
import com.park.utmstack.repository.incident_response.UtmIncidentActionRepository;
import com.park.utmstack.service.dto.incident_response.UtmIncidentActionCriteria;
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
 * Service for executing complex queries for UtmIncidentAction entities in the database. The main input is a {@link
 * UtmIncidentActionCriteria} which gets converted to {@link Specification}, in a way that all the filters must apply. It
 * returns a {@link List} of {@link UtmIncidentAction} or a {@link Page} of {@link UtmIncidentAction} which fulfills the
 * criteria.
 */
@Service
@Transactional(readOnly = true)
public class UtmIncidentActionQueryService extends QueryService<UtmIncidentAction> {

    private final Logger log = LoggerFactory.getLogger(UtmIncidentActionQueryService.class);

    private final UtmIncidentActionRepository utmIncidentActionRepository;

    public UtmIncidentActionQueryService(UtmIncidentActionRepository utmIncidentActionRepository) {
        this.utmIncidentActionRepository = utmIncidentActionRepository;
    }

    /**
     * Return a {@link List} of {@link UtmIncidentAction} which matches the criteria from the database
     *
     * @param criteria
     *     The object which holds all the filters, which the entities should match.
     *
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmIncidentAction> findByCriteria(UtmIncidentActionCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmIncidentAction> specification = createSpecification(criteria);
        return utmIncidentActionRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmIncidentAction} which matches the criteria from the database
     *
     * @param criteria
     *     The object which holds all the filters, which the entities should match.
     * @param page
     *     The page, which should be returned.
     *
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmIncidentAction> findByCriteria(UtmIncidentActionCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmIncidentAction> specification = createSpecification(criteria);
        return utmIncidentActionRepository.findAll(specification, page);
    }

    /**
     * Return the number of matching entities in the database
     *
     * @param criteria
     *     The object which holds all the filters, which the entities should match.
     *
     * @return the number of matching entities.
     */
    @Transactional(readOnly = true)
    public long countByCriteria(UtmIncidentActionCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<UtmIncidentAction> specification = createSpecification(criteria);
        return utmIncidentActionRepository.count(specification);
    }

    /**
     * Function to convert UtmIncidentActionCriteria to a {@link Specification}
     */
    private Specification<UtmIncidentAction> createSpecification(UtmIncidentActionCriteria criteria) {
        Specification<UtmIncidentAction> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmIncidentAction_.id));
            }
            if (criteria.getActionCommand() != null) {
                specification = specification.and(
                    buildStringSpecification(criteria.getActionCommand(), UtmIncidentAction_.actionCommand));
            }
            if (criteria.getActionDescription() != null) {
                specification = specification.and(
                    buildStringSpecification(criteria.getActionDescription(), UtmIncidentAction_.actionDescription));
            }
            if (criteria.getActionParams() != null) {
                specification = specification.and(
                    buildStringSpecification(criteria.getActionParams(), UtmIncidentAction_.actionParams));
            }
            if (criteria.getActionEditable() != null) {
                specification = specification.and(
                    buildSpecification(criteria.getActionEditable(), UtmIncidentAction_.actionEditable));
            }
            if (criteria.getActionType() != null) {
                specification = specification.and(
                    buildRangeSpecification(criteria.getActionType(), UtmIncidentAction_.actionType));
            }
        }
        return specification;
    }
}

package com.park.utmstack.service.incident_response;

import com.park.utmstack.domain.incident_response.UtmIncidentJob;
import com.park.utmstack.domain.incident_response.UtmIncidentJob_;
import com.park.utmstack.repository.incident_response.UtmIncidentJobRepository;
import com.park.utmstack.service.dto.incident_response.UtmIncidentJobCriteria;
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
 * Service for executing complex queries for UtmIncidentJob entities in the database. The main input is a {@link
 * UtmIncidentJobCriteria} which gets converted to {@link Specification}, in a way that all the filters must apply. It
 * returns a {@link List} of {@link UtmIncidentJob} or a {@link Page} of {@link UtmIncidentJob} which fulfills the criteria.
 */
@Service
@Transactional(readOnly = true)
public class UtmIncidentJobQueryService extends QueryService<UtmIncidentJob> {

    private final Logger log = LoggerFactory.getLogger(UtmIncidentJobQueryService.class);

    private final UtmIncidentJobRepository utmIncidentJobRepository;

    public UtmIncidentJobQueryService(UtmIncidentJobRepository utmIncidentJobRepository) {
        this.utmIncidentJobRepository = utmIncidentJobRepository;
    }

    /**
     * Return a {@link List} of {@link UtmIncidentJob} which matches the criteria from the database
     *
     * @param criteria
     *     The object which holds all the filters, which the entities should match.
     *
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmIncidentJob> findByCriteria(UtmIncidentJobCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmIncidentJob> specification = createSpecification(criteria);
        return utmIncidentJobRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmIncidentJob} which matches the criteria from the database
     *
     * @param criteria
     *     The object which holds all the filters, which the entities should match.
     * @param page
     *     The page, which should be returned.
     *
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmIncidentJob> findByCriteria(UtmIncidentJobCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmIncidentJob> specification = createSpecification(criteria);
        return utmIncidentJobRepository.findAll(specification, page);
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
    public long countByCriteria(UtmIncidentJobCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<UtmIncidentJob> specification = createSpecification(criteria);
        return utmIncidentJobRepository.count(specification);
    }

    /**
     * Function to convert UtmIncidentJobCriteria to a {@link Specification}
     */
    private Specification<UtmIncidentJob> createSpecification(UtmIncidentJobCriteria criteria) {
        Specification<UtmIncidentJob> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmIncidentJob_.id));
            }
            if (criteria.getActionId() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getActionId(), UtmIncidentJob_.actionId));
            }
            if (criteria.getParams() != null) {
                specification = specification.and(buildStringSpecification(criteria.getParams(), UtmIncidentJob_.params));
            }
            if (criteria.getAgent() != null) {
                specification = specification.and(buildStringSpecification(criteria.getAgent(), UtmIncidentJob_.agent));
            }
            if (criteria.getStatus() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getStatus(), UtmIncidentJob_.status));
            }
            if (criteria.getJobResult() != null) {
                specification = specification.and(
                    buildStringSpecification(criteria.getJobResult(), UtmIncidentJob_.jobResult));
            }
            if (criteria.getOriginId() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getOriginId(), UtmIncidentJob_.originId));
            }
            if (criteria.getOriginType() != null) {
                specification = specification.and(
                    buildStringSpecification(criteria.getOriginType(), UtmIncidentJob_.originType));
            }
        }
        return specification;
    }
}

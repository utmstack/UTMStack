package com.park.utmstack.service.incident;

import com.park.utmstack.domain.incident.UtmIncidentAlert;
import com.park.utmstack.domain.incident.UtmIncidentAlert_;
import com.park.utmstack.repository.incident.UtmIncidentAlertRepository;
import com.park.utmstack.service.dto.incident.UtmIncidentAlertCriteria;
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
 * Service for executing complex queries for UtmIncidentAlert entities in the database.
 * The main input is a {@link UtmIncidentAlertCriteria} which gets converted to {@link Specification},
 * in a way that all the filters must apply.
 * It returns a {@link List} of {@link UtmIncidentAlert} or a {@link Page} of {@link UtmIncidentAlert} which fulfills the criteria.
 */
@Service
@Transactional(readOnly = true)
public class UtmIncidentAlertQueryService extends QueryService<UtmIncidentAlert> {

    private final Logger log = LoggerFactory.getLogger(UtmIncidentAlertQueryService.class);

    private final UtmIncidentAlertRepository utmIncidentAlertRepository;

    public UtmIncidentAlertQueryService(UtmIncidentAlertRepository utmIncidentAlertRepository) {
        this.utmIncidentAlertRepository = utmIncidentAlertRepository;
    }

    /**
     * Return a {@link List} of {@link UtmIncidentAlert} which matches the criteria from the database
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmIncidentAlert> findByCriteria(UtmIncidentAlertCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmIncidentAlert> specification = createSpecification(criteria);
        return utmIncidentAlertRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmIncidentAlert} which matches the criteria from the database
     * @param criteria The object which holds all the filters, which the entities should match.
     * @param page The page, which should be returned.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmIncidentAlert> findByCriteria(UtmIncidentAlertCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmIncidentAlert> specification = createSpecification(criteria);
        return utmIncidentAlertRepository.findAll(specification, page);
    }

    /**
     * Return the number of matching entities in the database
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the number of matching entities.
     */
    @Transactional(readOnly = true)
    public long countByCriteria(UtmIncidentAlertCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<UtmIncidentAlert> specification = createSpecification(criteria);
        return utmIncidentAlertRepository.count(specification);
    }

    /**
     * Function to convert UtmIncidentAlertCriteria to a {@link Specification}
     */
    private Specification<UtmIncidentAlert> createSpecification(UtmIncidentAlertCriteria criteria) {
        Specification<UtmIncidentAlert> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmIncidentAlert_.id));
            }
            if (criteria.getIncidentId() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getIncidentId(), UtmIncidentAlert_.incidentId));
            }
            if (criteria.getAlertId() != null) {
                specification = specification.and(buildStringSpecification(criteria.getAlertId(), UtmIncidentAlert_.alertId));
            }
            if (criteria.getAlertName() != null) {
                specification = specification.and(buildStringSpecification(criteria.getAlertName(), UtmIncidentAlert_.alertName));
            }
            if (criteria.getAlertStatus() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getAlertStatus(), UtmIncidentAlert_.alertStatus));
            }
            if (criteria.getAlertSeverity() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getAlertSeverity(), UtmIncidentAlert_.alertSeverity));
            }
        }
        return specification;
    }
}

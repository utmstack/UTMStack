package com.park.utmstack.service.incident_response;

import com.park.utmstack.domain.incident_response.UtmIncidentVariable;
import com.park.utmstack.domain.incident_response.UtmIncidentVariable_;
import com.park.utmstack.repository.incident_response.UtmIncidentVariableRepository;
import com.park.utmstack.service.dto.incident_response.UtmIncidentVariableCriteria;
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
 * Service for executing complex queries for UtmIncidentVariable entities in the database. The main input is a {@link
 * UtmIncidentVariableCriteria} which gets converted to {@link Specification}, in a way that all the filters must apply. It
 * returns a {@link List} of {@link UtmIncidentVariable} or a {@link Page} of {@link UtmIncidentVariable} which fulfills the
 * criteria.
 */
@Service
@Transactional(readOnly = true)
public class UtmIncidentVariableQueryService extends QueryService<UtmIncidentVariable> {

    private final Logger log = LoggerFactory.getLogger(UtmIncidentVariableQueryService.class);

    private final UtmIncidentVariableRepository utmIncidentActionRepository;

    public UtmIncidentVariableQueryService(UtmIncidentVariableRepository utmIncidentActionRepository) {
        this.utmIncidentActionRepository = utmIncidentActionRepository;
    }

    /**
     * Return a {@link List} of {@link UtmIncidentVariable} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmIncidentVariable> findByCriteria(UtmIncidentVariableCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmIncidentVariable> specification = createSpecification(criteria);
        return utmIncidentActionRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmIncidentVariable} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @param page     The page, which should be returned.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmIncidentVariable> findByCriteria(UtmIncidentVariableCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmIncidentVariable> specification = createSpecification(criteria);
        return utmIncidentActionRepository.findAll(specification, page);
    }

    /**
     * Return the number of matching entities in the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the number of matching entities.
     */
    @Transactional(readOnly = true)
    public long countByCriteria(UtmIncidentVariableCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<UtmIncidentVariable> specification = createSpecification(criteria);
        return utmIncidentActionRepository.count(specification);
    }

    /**
     * Function to convert UtmIncidentVariableCriteria to a {@link Specification}
     */
    private Specification<UtmIncidentVariable> createSpecification(UtmIncidentVariableCriteria criteria) {
        Specification<UtmIncidentVariable> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmIncidentVariable_.id));
            }
            if (criteria.getVariableName() != null) {
                specification = specification.and(
                        buildStringSpecification(criteria.getVariableName(), UtmIncidentVariable_.variableName));
            }
        }
        return specification;
    }
}

package com.park.utmstack.service.incident_response;

import com.park.utmstack.domain.incident_response.UtmIncidentActionCommand;
import com.park.utmstack.domain.incident_response.UtmIncidentActionCommand_;
import com.park.utmstack.repository.incident_response.UtmIncidentActionCommandRepository;
import com.park.utmstack.service.dto.incident_response.UtmIncidentActionCommandCriteria;
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
 * Service for executing complex queries for UtmIncidentActionCommand entities in the database.
 * The main input is a {@link UtmIncidentActionCommandCriteria} which gets converted to {@link Specification},
 * in a way that all the filters must apply.
 * It returns a {@link List} of {@link UtmIncidentActionCommand} or a {@link Page} of {@link UtmIncidentActionCommand} which fulfills the criteria.
 */
@Service
@Transactional(readOnly = true)
public class UtmIncidentActionCommandQueryService extends QueryService<UtmIncidentActionCommand> {

    private final Logger log = LoggerFactory.getLogger(UtmIncidentActionCommandQueryService.class);

    private final UtmIncidentActionCommandRepository utmIncidentActionCommandRepository;

    public UtmIncidentActionCommandQueryService(UtmIncidentActionCommandRepository utmIncidentActionCommandRepository) {
        this.utmIncidentActionCommandRepository = utmIncidentActionCommandRepository;
    }

    /**
     * Return a {@link List} of {@link UtmIncidentActionCommand} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmIncidentActionCommand> findByCriteria(UtmIncidentActionCommandCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmIncidentActionCommand> specification = createSpecification(criteria);
        return utmIncidentActionCommandRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmIncidentActionCommand} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @param page     The page, which should be returned.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmIncidentActionCommand> findByCriteria(UtmIncidentActionCommandCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmIncidentActionCommand> specification = createSpecification(criteria);
        return utmIncidentActionCommandRepository.findAll(specification, page);
    }

    /**
     * Return the number of matching entities in the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the number of matching entities.
     */
    @Transactional(readOnly = true)
    public long countByCriteria(UtmIncidentActionCommandCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<UtmIncidentActionCommand> specification = createSpecification(criteria);
        return utmIncidentActionCommandRepository.count(specification);
    }

    /**
     * Function to convert UtmIncidentActionCommandCriteria to a {@link Specification}
     */
    private Specification<UtmIncidentActionCommand> createSpecification(UtmIncidentActionCommandCriteria criteria) {
        Specification<UtmIncidentActionCommand> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmIncidentActionCommand_.id));
            }
            if (criteria.getActionId() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getActionId(), UtmIncidentActionCommand_.actionId));
            }
            if (criteria.getOsDistribution() != null) {
                specification = specification.and(buildStringSpecification(criteria.getOsDistribution(), UtmIncidentActionCommand_.osPlatform));
            }
            if (criteria.getCommand() != null) {
                specification = specification.and(buildStringSpecification(criteria.getCommand(), UtmIncidentActionCommand_.command));
            }
        }
        return specification;
    }
}

package com.park.utmstack.service;

import com.park.utmstack.domain.UtmIntegration;
import com.park.utmstack.domain.UtmIntegration_;
import com.park.utmstack.repository.UtmIntegrationRepository;
import com.park.utmstack.service.dto.UtmIntegrationCriteria;
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
 * Service for executing complex queries for UtmIntegration entities in the database.
 * The main input is a {@link UtmIntegrationCriteria} which gets converted to {@link Specification},
 * in a way that all the filters must apply.
 * It returns a {@link List} of {@link UtmIntegration} or a {@link Page} of {@link UtmIntegration} which fulfills the criteria.
 */
@Service
@Transactional(readOnly = true)
public class UtmIntegrationQueryService extends QueryService<UtmIntegration> {

    private final Logger log = LoggerFactory.getLogger(UtmIntegrationQueryService.class);

    private final UtmIntegrationRepository utmIntegrationRepository;

    public UtmIntegrationQueryService(UtmIntegrationRepository utmIntegrationRepository) {
        this.utmIntegrationRepository = utmIntegrationRepository;
    }

    /**
     * Return a {@link List} of {@link UtmIntegration} which matches the criteria from the database
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmIntegration> findByCriteria(UtmIntegrationCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmIntegration> specification = createSpecification(criteria);
        return utmIntegrationRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmIntegration} which matches the criteria from the database
     * @param criteria The object which holds all the filters, which the entities should match.
     * @param page The page, which should be returned.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmIntegration> findByCriteria(UtmIntegrationCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmIntegration> specification = createSpecification(criteria);
        return utmIntegrationRepository.findAll(specification, page);
    }

    /**
     * Return the number of matching entities in the database
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the number of matching entities.
     */
    @Transactional(readOnly = true)
    public long countByCriteria(UtmIntegrationCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<UtmIntegration> specification = createSpecification(criteria);
        return utmIntegrationRepository.count(specification);
    }

    /**
     * Function to convert UtmIntegrationCriteria to a {@link Specification}
     */
    private Specification<UtmIntegration> createSpecification(UtmIntegrationCriteria criteria) {
        Specification<UtmIntegration> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmIntegration_.id));
            }
            if (criteria.getModuleId() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getModuleId(), UtmIntegration_.moduleId));
            }
            if (criteria.getIntegrationName() != null) {
                specification = specification.and(buildStringSpecification(criteria.getIntegrationName(), UtmIntegration_.integrationName));
            }
            if (criteria.getIntegrationDescription() != null) {
                specification = specification.and(buildStringSpecification(criteria.getIntegrationDescription(), UtmIntegration_.integrationDescription));
            }
            if (criteria.getUrl() != null) {
                specification = specification.and(buildStringSpecification(criteria.getUrl(), UtmIntegration_.url));
            }
        }
        return specification;
    }
}

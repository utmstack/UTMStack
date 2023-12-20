package com.park.utmstack.service;

import com.park.utmstack.domain.UtmIntegrationConf;
import com.park.utmstack.domain.UtmIntegrationConf_;
import com.park.utmstack.repository.UtmIntegrationConfRepository;
import com.park.utmstack.service.dto.UtmIntegrationConfCriteria;
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
 * Service for executing complex queries for UtmIntegrationConf entities in the database.
 * The main input is a {@link UtmIntegrationConfCriteria} which gets converted to {@link Specification},
 * in a way that all the filters must apply.
 * It returns a {@link List} of {@link UtmIntegrationConf} or a {@link Page} of {@link UtmIntegrationConf} which fulfills the criteria.
 */
@Service
@Transactional(readOnly = true)
public class UtmIntegrationConfQueryService extends QueryService<UtmIntegrationConf> {

    private final Logger log = LoggerFactory.getLogger(UtmIntegrationConfQueryService.class);

    private final UtmIntegrationConfRepository utmIntegrationConfRepository;

    public UtmIntegrationConfQueryService(UtmIntegrationConfRepository utmIntegrationConfRepository) {
        this.utmIntegrationConfRepository = utmIntegrationConfRepository;
    }

    /**
     * Return a {@link List} of {@link UtmIntegrationConf} which matches the criteria from the database
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmIntegrationConf> findByCriteria(UtmIntegrationConfCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmIntegrationConf> specification = createSpecification(criteria);
        return utmIntegrationConfRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmIntegrationConf} which matches the criteria from the database
     * @param criteria The object which holds all the filters, which the entities should match.
     * @param page The page, which should be returned.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmIntegrationConf> findByCriteria(UtmIntegrationConfCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmIntegrationConf> specification = createSpecification(criteria);
        return utmIntegrationConfRepository.findAll(specification, page);
    }

    /**
     * Return the number of matching entities in the database
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the number of matching entities.
     */
    @Transactional(readOnly = true)
    public long countByCriteria(UtmIntegrationConfCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<UtmIntegrationConf> specification = createSpecification(criteria);
        return utmIntegrationConfRepository.count(specification);
    }

    /**
     * Function to convert UtmIntegrationConfCriteria to a {@link Specification}
     */
    private Specification<UtmIntegrationConf> createSpecification(UtmIntegrationConfCriteria criteria) {
        Specification<UtmIntegrationConf> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmIntegrationConf_.id));
            }
            if (criteria.getIntegrationId() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getIntegrationId(), UtmIntegrationConf_.integrationId));
            }
            if (criteria.getConfShort() != null) {
                specification = specification.and(buildStringSpecification(criteria.getConfShort(), UtmIntegrationConf_.confShort));
            }
            if (criteria.getConfLarge() != null) {
                specification = specification.and(buildStringSpecification(criteria.getConfLarge(), UtmIntegrationConf_.confLarge));
            }
            if (criteria.getConfDescription() != null) {
                specification = specification.and(buildStringSpecification(criteria.getConfDescription(), UtmIntegrationConf_.confDescription));
            }
            if (criteria.getConfValue() != null) {
                specification = specification.and(buildStringSpecification(criteria.getConfValue(), UtmIntegrationConf_.confValue));
            }
            if (criteria.getConfDatatype() != null) {
                specification = specification.and(buildStringSpecification(criteria.getConfDatatype(), UtmIntegrationConf_.confDatatype));
            }
        }
        return specification;
    }
}

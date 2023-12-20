package com.park.utmstack.service;

import com.park.utmstack.domain.UtmConfigurationSection;
import com.park.utmstack.domain.UtmConfigurationSection_;
import com.park.utmstack.repository.UtmConfigurationSectionRepository;
import com.park.utmstack.service.dto.UtmConfigurationSectionCriteria;
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
 * Service for executing complex queries for UtmConfigurationSection entities in the database.
 * The main input is a {@link UtmConfigurationSectionCriteria} which gets converted to {@link Specification},
 * in a way that all the filters must apply.
 * It returns a {@link List} of {@link UtmConfigurationSection} or a {@link Page} of {@link UtmConfigurationSection} which fulfills the criteria.
 */
@Service
@Transactional(readOnly = true)
public class UtmConfigurationSectionQueryService extends QueryService<UtmConfigurationSection> {

    private final Logger log = LoggerFactory.getLogger(UtmConfigurationSectionQueryService.class);

    private final UtmConfigurationSectionRepository utmConfigurationSectionRepository;

    public UtmConfigurationSectionQueryService(UtmConfigurationSectionRepository utmConfigurationSectionRepository) {
        this.utmConfigurationSectionRepository = utmConfigurationSectionRepository;
    }

    /**
     * Return a {@link List} of {@link UtmConfigurationSection} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmConfigurationSection> findByCriteria(UtmConfigurationSectionCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmConfigurationSection> specification = createSpecification(criteria);
        return utmConfigurationSectionRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmConfigurationSection} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @param page     The page, which should be returned.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmConfigurationSection> findByCriteria(UtmConfigurationSectionCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmConfigurationSection> specification = createSpecification(criteria);
        return utmConfigurationSectionRepository.findAll(specification, page);
    }

    /**
     * Return the number of matching entities in the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the number of matching entities.
     */
    @Transactional(readOnly = true)
    public long countByCriteria(UtmConfigurationSectionCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<UtmConfigurationSection> specification = createSpecification(criteria);
        return utmConfigurationSectionRepository.count(specification);
    }

    /**
     * Function to convert UtmConfigurationSectionCriteria to a {@link Specification}
     */
    private Specification<UtmConfigurationSection> createSpecification(UtmConfigurationSectionCriteria criteria) {
        Specification<UtmConfigurationSection> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmConfigurationSection_.id));
            }
            if (criteria.getSection() != null) {
                specification = specification.and(buildStringSpecification(criteria.getSection(), UtmConfigurationSection_.section));
            }
            if (criteria.getDescription() != null) {
                specification = specification.and(buildStringSpecification(criteria.getDescription(), UtmConfigurationSection_.description));
            }
            if (criteria.getModuleNameShort() != null) {
                specification = specification.and(buildSpecification(criteria.getModuleNameShort(), UtmConfigurationSection_.moduleNameShort));
            }
        }
        return specification;
    }
}

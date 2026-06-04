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
import tech.jhipster.service.filter.BooleanFilter;

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

    private final UtmConfigurationSectionRepository configurationSectionRepository;

    public UtmConfigurationSectionQueryService(UtmConfigurationSectionRepository configurationSectionRepository) {
        this.configurationSectionRepository = configurationSectionRepository;
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
        final Specification<UtmConfigurationSection> specification = createSpecification(criteria);
        return configurationSectionRepository.findAll(specification, page);
    }

    /**
     * Function to convert UtmConfigurationSectionCriteria to a {@link Specification}
     */
    private Specification<UtmConfigurationSection> createSpecification(UtmConfigurationSectionCriteria criteria) {
        Specification<UtmConfigurationSection> specification = Specification.where(null);

        BooleanFilter sectionActiveFilter = new BooleanFilter();
        sectionActiveFilter.setEquals(true);
        specification = specification.and(buildSpecification(sectionActiveFilter, UtmConfigurationSection_.sectionActive));

        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmConfigurationSection_.id));
            }
            if (criteria.getSection() != null) {
                specification = specification.and(buildStringSpecification(criteria.getSection(), UtmConfigurationSection_.section));
            }
            if (criteria.getShortName() != null) {
                specification = specification.and(buildSpecification(criteria.getShortName(), UtmConfigurationSection_.shortName));
            }
        }

        return specification;
    }
}

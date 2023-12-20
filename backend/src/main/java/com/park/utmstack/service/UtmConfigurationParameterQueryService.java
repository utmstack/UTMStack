package com.park.utmstack.service;

import com.park.utmstack.domain.UtmConfigurationParameter;
import com.park.utmstack.domain.UtmConfigurationParameter_;
import com.park.utmstack.repository.UtmConfigurationParameterRepository;
import com.park.utmstack.service.dto.UtmConfigurationParameterCriteria;
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
 * Service for executing complex queries for UtmConfigurationParameter entities in the database.
 * The main input is a {@link UtmConfigurationParameterCriteria} which gets converted to {@link Specification},
 * in a way that all the filters must apply.
 * It returns a {@link List} of {@link UtmConfigurationParameter} or a {@link Page} of {@link UtmConfigurationParameter} which fulfills the criteria.
 */
@Service
@Transactional(readOnly = true)
public class UtmConfigurationParameterQueryService extends QueryService<UtmConfigurationParameter> {

    private final Logger log = LoggerFactory.getLogger(UtmConfigurationParameterQueryService.class);

    private final UtmConfigurationParameterRepository utmConfigurationParameterRepository;

    public UtmConfigurationParameterQueryService(UtmConfigurationParameterRepository utmConfigurationParameterRepository) {
        this.utmConfigurationParameterRepository = utmConfigurationParameterRepository;
    }

    /**
     * Return a {@link List} of {@link UtmConfigurationParameter} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmConfigurationParameter> findByCriteria(UtmConfigurationParameterCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmConfigurationParameter> specification = createSpecification(criteria);
        return utmConfigurationParameterRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmConfigurationParameter} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @param page     The page, which should be returned.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmConfigurationParameter> findByCriteria(UtmConfigurationParameterCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmConfigurationParameter> specification = createSpecification(criteria);
        return utmConfigurationParameterRepository.findAll(specification, page);
    }

    /**
     * Return the number of matching entities in the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the number of matching entities.
     */
    @Transactional(readOnly = true)
    public long countByCriteria(UtmConfigurationParameterCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<UtmConfigurationParameter> specification = createSpecification(criteria);
        return utmConfigurationParameterRepository.count(specification);
    }

    /**
     * Function to convert UtmConfigurationParameterCriteria to a {@link Specification}
     */
    private Specification<UtmConfigurationParameter> createSpecification(UtmConfigurationParameterCriteria criteria) {
        Specification<UtmConfigurationParameter> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmConfigurationParameter_.id));
            }
            if (criteria.getSectionId() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getSectionId(), UtmConfigurationParameter_.sectionId));
            }
            if (criteria.getConfParamShort() != null) {
                specification = specification.and(buildStringSpecification(criteria.getConfParamShort(), UtmConfigurationParameter_.confParamShort));
            }
            if (criteria.getConfParamLarge() != null) {
                specification = specification.and(buildStringSpecification(criteria.getConfParamLarge(), UtmConfigurationParameter_.confParamLarge));
            }
            if (criteria.getConfParamDescription() != null) {
                specification = specification.and(buildStringSpecification(criteria.getConfParamDescription(), UtmConfigurationParameter_.confParamDescription));
            }
            if (criteria.getConfParamValue() != null) {
                specification = specification.and(buildStringSpecification(criteria.getConfParamValue(), UtmConfigurationParameter_.confParamValue));
            }
            if (criteria.getConfParamDatatype() != null) {
                specification = specification.and(buildStringSpecification(criteria.getConfParamDatatype(), UtmConfigurationParameter_.confParamDatatype));
            }
            if (criteria.getConfParamRequired() != null) {
                specification = specification.and(buildSpecification(criteria.getConfParamRequired(), UtmConfigurationParameter_.confParamRequired));
            }
            if (criteria.getModificationTime() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getModificationTime(), UtmConfigurationParameter_.modificationTime));
            }
            if (criteria.getModificationUser() != null) {
                specification = specification.and(buildStringSpecification(criteria.getModificationUser(), UtmConfigurationParameter_.modificationUser));
            }
        }
        return specification;
    }
}

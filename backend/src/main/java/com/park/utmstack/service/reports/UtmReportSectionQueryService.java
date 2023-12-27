package com.park.utmstack.service.reports;

import com.park.utmstack.domain.reports.UtmReportSection;
import com.park.utmstack.domain.reports.UtmReportSection_;
import com.park.utmstack.repository.reports.UtmReportSectionRepository;
import com.park.utmstack.service.dto.reports.UtmReportSectionCriteria;
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
 * Service for executing complex queries for UtmReportSection entities in the database.
 * The main input is a {@link UtmReportSectionCriteria} which gets converted to {@link Specification},
 * in a way that all the filters must apply.
 * It returns a {@link List} of {@link UtmReportSection} or a {@link Page} of {@link UtmReportSection} which fulfills the criteria.
 */
@Service
@Transactional(readOnly = true)
public class UtmReportSectionQueryService extends QueryService<UtmReportSection> {

    private final Logger log = LoggerFactory.getLogger(UtmReportSectionQueryService.class);

    private final UtmReportSectionRepository utmReportSectionRepository;

    public UtmReportSectionQueryService(UtmReportSectionRepository utmReportSectionRepository) {
        this.utmReportSectionRepository = utmReportSectionRepository;
    }

    /**
     * Return a {@link List} of {@link UtmReportSection} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmReportSection> findByCriteria(UtmReportSectionCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmReportSection> specification = createSpecification(criteria);
        return utmReportSectionRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmReportSection} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @param page     The page, which should be returned.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmReportSection> findByCriteria(UtmReportSectionCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmReportSection> specification = createSpecification(criteria);
        return utmReportSectionRepository.findAll(specification, page);
    }

    /**
     * Return the number of matching entities in the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the number of matching entities.
     */
    @Transactional(readOnly = true)
    public long countByCriteria(UtmReportSectionCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<UtmReportSection> specification = createSpecification(criteria);
        return utmReportSectionRepository.count(specification);
    }

    /**
     * Function to convert UtmReportSectionCriteria to a {@link Specification}
     */
    private Specification<UtmReportSection> createSpecification(UtmReportSectionCriteria criteria) {
        Specification<UtmReportSection> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmReportSection_.id));
            }
            if (criteria.getRepSecName() != null) {
                specification = specification.and(buildStringSpecification(criteria.getRepSecName(), UtmReportSection_.repSecName));
            }
            if (criteria.getRepSecDescription() != null) {
                specification = specification.and(buildStringSpecification(criteria.getRepSecDescription(), UtmReportSection_.repSecDescription));
            }
            if (criteria.getCreationUser() != null) {
                specification = specification.and(buildStringSpecification(criteria.getCreationUser(), UtmReportSection_.creationUser));
            }
            if (criteria.getCreationDate() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getCreationDate(), UtmReportSection_.creationDate));
            }
            if (criteria.getModificationUser() != null) {
                specification = specification.and(buildStringSpecification(criteria.getModificationUser(), UtmReportSection_.modificationUser));
            }
            if (criteria.getModificationDate() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getModificationDate(), UtmReportSection_.modificationDate));
            }
            if (criteria.getRepSecSystem() != null) {
                specification = specification.and(buildSpecification(criteria.getRepSecSystem(), UtmReportSection_.repSecSystem));
            }
        }
        return specification;
    }
}

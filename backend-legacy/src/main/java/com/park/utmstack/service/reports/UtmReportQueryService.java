package com.park.utmstack.service.reports;

import com.park.utmstack.domain.reports.UtmReport;
import com.park.utmstack.domain.reports.UtmReport_;
import com.park.utmstack.repository.reports.UtmReportRepository;
import com.park.utmstack.service.dto.reports.UtmReportCriteria;
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
 * Service for executing complex queries for UtmReports entities in the database.
 * The main input is a {@link UtmReportCriteria} which gets converted to {@link Specification},
 * in a way that all the filters must apply.
 * It returns a {@link List} of {@link UtmReport} or a {@link Page} of {@link UtmReport} which fulfills the criteria.
 */
@Service
@Transactional(readOnly = true)
public class UtmReportQueryService extends QueryService<UtmReport> {

    private final Logger log = LoggerFactory.getLogger(UtmReportQueryService.class);

    private final UtmReportRepository utmReportsRepository;

    public UtmReportQueryService(UtmReportRepository utmReportsRepository) {
        this.utmReportsRepository = utmReportsRepository;
    }

    /**
     * Return a {@link List} of {@link UtmReport} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmReport> findByCriteria(UtmReportCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmReport> specification = createSpecification(criteria);
        return utmReportsRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmReport} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @param page     The page, which should be returned.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmReport> findByCriteria(UtmReportCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmReport> specification = createSpecification(criteria);
        return utmReportsRepository.findAll(specification, page);
    }

    /**
     * Return the number of matching entities in the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the number of matching entities.
     */
    @Transactional(readOnly = true)
    public long countByCriteria(UtmReportCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<UtmReport> specification = createSpecification(criteria);
        return utmReportsRepository.count(specification);
    }

    /**
     * Function to convert UtmReportsCriteria to a {@link Specification}
     */
    private Specification<UtmReport> createSpecification(UtmReportCriteria criteria) {
        Specification<UtmReport> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmReport_.id));
            }
            if (criteria.getRepName() != null) {
                specification = specification.and(buildStringSpecification(criteria.getRepName(), UtmReport_.repName));
            }
            if (criteria.getRepDescription() != null) {
                specification = specification.and(buildStringSpecification(criteria.getRepDescription(), UtmReport_.repDescription));
            }
            if (criteria.getReportSectionId() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getReportSectionId(), UtmReport_.reportSectionId));
            }
            if (criteria.getDashboardId() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getDashboardId(), UtmReport_.dashboardId));
            }
            if (criteria.getCreationUser() != null) {
                specification = specification.and(buildStringSpecification(criteria.getCreationUser(), UtmReport_.creationUser));
            }
            if (criteria.getCreationDate() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getCreationDate(), UtmReport_.creationDate));
            }
            if (criteria.getModificationUser() != null) {
                specification = specification.and(buildStringSpecification(criteria.getModificationUser(), UtmReport_.modificationUser));
            }
            if (criteria.getModificationDate() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getModificationDate(), UtmReport_.modificationDate));
            }
            if (criteria.getRepType() != null) {
                specification = specification.and(buildSpecification(criteria.getRepType(), UtmReport_.repType));
            }
            if (criteria.getRepModule() != null) {
                specification = specification.and(buildStringSpecification(criteria.getRepModule(), UtmReport_.repModule));
            }
            if (criteria.getRepShortName() != null) {
                specification = specification.and(buildStringSpecification(criteria.getRepShortName(), UtmReport_.repShortName));
            }
            if (criteria.getRepHttpMethod() != null) {
                specification = specification.and(buildStringSpecification(criteria.getRepHttpMethod(), UtmReport_.repHttpMethod));
            }
        }
        return specification;
    }
}

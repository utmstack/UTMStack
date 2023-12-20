package com.park.utmstack.service.log_analyzer;

import com.park.utmstack.domain.log_analyzer.LogAnalyzerQuery;
import com.park.utmstack.domain.log_analyzer.LogAnalyzerQuery_;
import com.park.utmstack.repository.log_analyzer.LogAnalyzerQueryRepository;
import com.park.utmstack.service.dto.log_analyzer.LogAnalyzerQueryCriteria;
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
 * Service for executing complex queries for LogAnalyzerQuery entities in the database. The main input is a {@link
 * LogAnalyzerQueryCriteria} which gets converted to {@link Specification}, in a way that all the filters must apply. It
 * returns a {@link List} of {@link LogAnalyzerQuery} or a {@link Page} of {@link LogAnalyzerQuery} which fulfills the
 * criteria.
 */
@Service
@Transactional(readOnly = true)
public class LogAnalyzerQueryService extends QueryService<LogAnalyzerQuery> {

    private final Logger log = LoggerFactory.getLogger(LogAnalyzerQueryService.class);

    private final LogAnalyzerQueryRepository logAnalyzerQueryRepository;

    public LogAnalyzerQueryService(LogAnalyzerQueryRepository logAnalyzerQueryRepository) {
        this.logAnalyzerQueryRepository = logAnalyzerQueryRepository;
    }

    /**
     * Return a {@link List} of {@link LogAnalyzerQuery} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<LogAnalyzerQuery> findByCriteria(LogAnalyzerQueryCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<LogAnalyzerQuery> specification = createSpecification(criteria);
        return logAnalyzerQueryRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link LogAnalyzerQuery} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @param page     The page, which should be returned.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<LogAnalyzerQuery> findByCriteria(LogAnalyzerQueryCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<LogAnalyzerQuery> specification = createSpecification(criteria);
        return logAnalyzerQueryRepository.findAll(specification, page);
    }

    /**
     * Return the number of matching entities in the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the number of matching entities.
     */
    @Transactional(readOnly = true)
    public long countByCriteria(LogAnalyzerQueryCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<LogAnalyzerQuery> specification = createSpecification(criteria);
        return logAnalyzerQueryRepository.count(specification);
    }

    /**
     * Function to convert LogAnalyzerQueryCriteria to a {@link Specification}
     */
    private Specification<LogAnalyzerQuery> createSpecification(LogAnalyzerQueryCriteria criteria) {
        Specification<LogAnalyzerQuery> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), LogAnalyzerQuery_.id));
            }
            if (criteria.getName() != null) {
                specification = specification.and(buildStringSpecification(criteria.getName(), LogAnalyzerQuery_.name));
            }
            if (criteria.getDescription() != null) {
                specification = specification.and(
                    buildStringSpecification(criteria.getDescription(), LogAnalyzerQuery_.description));
            }
            if (criteria.getOwner() != null) {
                specification = specification.and(buildStringSpecification(criteria.getOwner(), LogAnalyzerQuery_.owner));
            }
            if (criteria.getCreationDate() != null) {
                specification = specification.and(
                    buildRangeSpecification(criteria.getCreationDate(), LogAnalyzerQuery_.creationDate));
            }
            if (criteria.getModificationDate() != null) {
                specification = specification.and(
                    buildRangeSpecification(criteria.getModificationDate(), LogAnalyzerQuery_.modificationDate));
            }
            if (criteria.getIdPattern() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getIdPattern(), LogAnalyzerQuery_.idPattern));
            }

        }
        return specification;
    }
}

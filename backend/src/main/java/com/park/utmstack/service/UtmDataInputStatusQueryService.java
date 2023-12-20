package com.park.utmstack.service;

import com.park.utmstack.domain.UtmDataInputStatus;
import com.park.utmstack.domain.UtmDataInputStatus_;
import com.park.utmstack.repository.UtmDataInputStatusRepository;
import com.park.utmstack.service.dto.UtmDataInputStatusCriteria;
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
 * Service for executing complex queries for UtmDataInputStatus entities in the database.
 * The main input is a {@link UtmDataInputStatusCriteria} which gets converted to {@link Specification},
 * in a way that all the filters must apply.
 * It returns a {@link List} of {@link UtmDataInputStatus} or a {@link Page} of {@link UtmDataInputStatus} which fulfills the criteria.
 */
@Service
@Transactional(readOnly = true)
public class UtmDataInputStatusQueryService extends QueryService<UtmDataInputStatus> {

    private final Logger log = LoggerFactory.getLogger(UtmDataInputStatusQueryService.class);

    private final UtmDataInputStatusRepository utmDataInputStatusRepository;

    public UtmDataInputStatusQueryService(UtmDataInputStatusRepository utmDataInputStatusRepository) {
        this.utmDataInputStatusRepository = utmDataInputStatusRepository;
    }

    /**
     * Return a {@link List} of {@link UtmDataInputStatus} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmDataInputStatus> findByCriteria(UtmDataInputStatusCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmDataInputStatus> specification = createSpecification(criteria);
        return utmDataInputStatusRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmDataInputStatus} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @param page     The page, which should be returned.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmDataInputStatus> findByCriteria(UtmDataInputStatusCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmDataInputStatus> specification = createSpecification(criteria);
        return utmDataInputStatusRepository.findAll(specification, page);
    }

    /**
     * Return the number of matching entities in the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the number of matching entities.
     */
    @Transactional(readOnly = true)
    public long countByCriteria(UtmDataInputStatusCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<UtmDataInputStatus> specification = createSpecification(criteria);
        return utmDataInputStatusRepository.count(specification);
    }

    /**
     * Function to convert UtmDataInputStatusCriteria to a {@link Specification}
     */
    private Specification<UtmDataInputStatus> createSpecification(UtmDataInputStatusCriteria criteria) {
        Specification<UtmDataInputStatus> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildStringSpecification(criteria.getId(), UtmDataInputStatus_.id));
            }
            if (criteria.getSource() != null) {
                specification = specification.and(buildStringSpecification(criteria.getSource(), UtmDataInputStatus_.source));
            }
            if (criteria.getDataType() != null) {
                specification = specification.and(buildStringSpecification(criteria.getDataType(), UtmDataInputStatus_.dataType));
            }
            if (criteria.getTimestamp() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getTimestamp(), UtmDataInputStatus_.timestamp));
            }
            if (criteria.getMedian() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getMedian(), UtmDataInputStatus_.median));
            }
        }
        return specification;
    }
}

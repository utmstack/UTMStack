package com.park.utmstack.service.logstash_filter;

import com.park.utmstack.domain.logstash_filter.UtmLogstashFilterGroup;
import com.park.utmstack.domain.logstash_filter.UtmLogstashFilterGroup_;
import com.park.utmstack.repository.logstash_filter.UtmLogstashFilterGroupRepository;
import com.park.utmstack.service.dto.logstash_filter.UtmLogstashFilterGroupCriteria;
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
 * Service for executing complex queries for UtmLogstashFilterGroup entities in the database.
 * The main input is a {@link UtmLogstashFilterGroupCriteria} which gets converted to {@link Specification},
 * in a way that all the filters must apply.
 * It returns a {@link List} of {@link UtmLogstashFilterGroup} or a {@link Page} of {@link UtmLogstashFilterGroup} which fulfills the criteria.
 */
@Service
@Transactional(readOnly = true)
public class UtmLogstashFilterGroupQueryService extends QueryService<UtmLogstashFilterGroup> {

    private final Logger log = LoggerFactory.getLogger(UtmLogstashFilterGroupQueryService.class);

    private final UtmLogstashFilterGroupRepository utmLogstashFilterGroupRepository;

    public UtmLogstashFilterGroupQueryService(UtmLogstashFilterGroupRepository utmLogstashFilterGroupRepository) {
        this.utmLogstashFilterGroupRepository = utmLogstashFilterGroupRepository;
    }

    /**
     * Return a {@link List} of {@link UtmLogstashFilterGroup} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmLogstashFilterGroup> findByCriteria(UtmLogstashFilterGroupCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmLogstashFilterGroup> specification = createSpecification(criteria);
        return utmLogstashFilterGroupRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmLogstashFilterGroup} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @param page     The page, which should be returned.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmLogstashFilterGroup> findByCriteria(UtmLogstashFilterGroupCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmLogstashFilterGroup> specification = createSpecification(criteria);
        return utmLogstashFilterGroupRepository.findAll(specification, page);
    }

    /**
     * Return the number of matching entities in the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the number of matching entities.
     */
    @Transactional(readOnly = true)
    public long countByCriteria(UtmLogstashFilterGroupCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<UtmLogstashFilterGroup> specification = createSpecification(criteria);
        return utmLogstashFilterGroupRepository.count(specification);
    }

    /**
     * Function to convert UtmLogstashFilterGroupCriteria to a {@link Specification}
     */
    private Specification<UtmLogstashFilterGroup> createSpecification(UtmLogstashFilterGroupCriteria criteria) {
        Specification<UtmLogstashFilterGroup> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmLogstashFilterGroup_.id));
            }
            if (criteria.getGroupName() != null) {
                specification = specification.and(buildStringSpecification(criteria.getGroupName(), UtmLogstashFilterGroup_.groupName));
            }
            if (criteria.getGroupDescription() != null) {
                specification = specification.and(buildStringSpecification(criteria.getGroupDescription(), UtmLogstashFilterGroup_.groupDescription));
            }
        }
        return specification;
    }
}

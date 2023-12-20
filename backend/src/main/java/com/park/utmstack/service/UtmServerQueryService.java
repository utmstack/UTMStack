package com.park.utmstack.service;

import com.park.utmstack.domain.UtmServer;
import com.park.utmstack.domain.UtmServer_;
import com.park.utmstack.repository.UtmServerRepository;
import com.park.utmstack.service.dto.UtmServerCriteria;
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
 * Service for executing complex queries for UtmServer entities in the database.
 * The main input is a {@link UtmServerCriteria} which gets converted to {@link Specification},
 * in a way that all the filters must apply.
 * It returns a {@link List} of {@link UtmServer} or a {@link Page} of {@link UtmServer} which fulfills the criteria.
 */
@Service
@Transactional(readOnly = true)
public class UtmServerQueryService extends QueryService<UtmServer> {

    private final Logger log = LoggerFactory.getLogger(UtmServerQueryService.class);

    private final UtmServerRepository utmServerRepository;

    public UtmServerQueryService(UtmServerRepository utmServerRepository) {
        this.utmServerRepository = utmServerRepository;
    }

    /**
     * Return a {@link List} of {@link UtmServer} which matches the criteria from the database
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmServer> findByCriteria(UtmServerCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmServer> specification = createSpecification(criteria);
        return utmServerRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmServer} which matches the criteria from the database
     * @param criteria The object which holds all the filters, which the entities should match.
     * @param page The page, which should be returned.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmServer> findByCriteria(UtmServerCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmServer> specification = createSpecification(criteria);
        return utmServerRepository.findAll(specification, page);
    }

    /**
     * Return the number of matching entities in the database
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the number of matching entities.
     */
    @Transactional(readOnly = true)
    public long countByCriteria(UtmServerCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<UtmServer> specification = createSpecification(criteria);
        return utmServerRepository.count(specification);
    }

    /**
     * Function to convert UtmServerCriteria to a {@link Specification}
     */
    private Specification<UtmServer> createSpecification(UtmServerCriteria criteria) {
        Specification<UtmServer> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmServer_.id));
            }
            if (criteria.getServerName() != null) {
                specification = specification.and(buildStringSpecification(criteria.getServerName(), UtmServer_.serverName));
            }
            if (criteria.getServerType() != null) {
                specification = specification.and(buildStringSpecification(criteria.getServerType(), UtmServer_.serverType));
            }
        }
        return specification;
    }
}

package com.park.utmstack.service;

import com.park.utmstack.domain.UtmServerModule;
import com.park.utmstack.domain.UtmServerModule_;
import com.park.utmstack.repository.UtmServerModuleRepository;
import com.park.utmstack.service.dto.UtmServerModuleCriteria;
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
 * Service for executing complex queries for UtmServerModule entities in the database.
 * The main input is a {@link UtmServerModuleCriteria} which gets converted to {@link Specification},
 * in a way that all the filters must apply.
 * It returns a {@link List} of {@link UtmServerModule} or a {@link Page} of {@link UtmServerModule} which fulfills the criteria.
 */
@Service
@Transactional(readOnly = true)
public class UtmServerModuleQueryService extends QueryService<UtmServerModule> {

    private final Logger log = LoggerFactory.getLogger(UtmServerModuleQueryService.class);

    private final UtmServerModuleRepository utmServerModuleRepository;

    public UtmServerModuleQueryService(UtmServerModuleRepository utmServerModuleRepository) {
        this.utmServerModuleRepository = utmServerModuleRepository;
    }

    /**
     * Return a {@link List} of {@link UtmServerModule} which matches the criteria from the database
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmServerModule> findByCriteria(UtmServerModuleCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmServerModule> specification = createSpecification(criteria);
        return utmServerModuleRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmServerModule} which matches the criteria from the database
     * @param criteria The object which holds all the filters, which the entities should match.
     * @param page The page, which should be returned.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmServerModule> findByCriteria(UtmServerModuleCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmServerModule> specification = createSpecification(criteria);
        return utmServerModuleRepository.findAll(specification, page);
    }

    /**
     * Return the number of matching entities in the database
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the number of matching entities.
     */
    @Transactional(readOnly = true)
    public long countByCriteria(UtmServerModuleCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<UtmServerModule> specification = createSpecification(criteria);
        return utmServerModuleRepository.count(specification);
    }

    /**
     * Function to convert UtmServerModuleCriteria to a {@link Specification}
     */
    private Specification<UtmServerModule> createSpecification(UtmServerModuleCriteria criteria) {
        Specification<UtmServerModule> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmServerModule_.id));
            }
            if (criteria.getServerId() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getServerId(), UtmServerModule_.serverId));
            }
            if (criteria.getModuleName() != null) {
                specification = specification.and(buildStringSpecification(criteria.getModuleName(), UtmServerModule_.moduleName));
            }
            if (criteria.getNeedsRestarts() != null) {
                specification = specification.and(buildSpecification(criteria.getNeedsRestarts(), UtmServerModule_.needsRestart));
            }
            if (criteria.getPrettyName() != null) {
                specification = specification.and(buildStringSpecification(criteria.getPrettyName(), UtmServerModule_.prettyName));
            }
        }
        return specification;
    }
}

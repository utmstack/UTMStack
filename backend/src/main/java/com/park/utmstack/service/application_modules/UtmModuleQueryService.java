package com.park.utmstack.service.application_modules;

import com.park.utmstack.domain.application_modules.UtmModule;
import com.park.utmstack.domain.application_modules.UtmModule_;
import com.park.utmstack.repository.application_modules.UtmModuleRepository;
import com.park.utmstack.service.dto.application_modules.UtmModuleCriteria;
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
 * Service for executing complex queries for UtmModule entities in the database.
 * The main input is a {@link UtmModuleCriteria} which gets converted to {@link Specification},
 * in a way that all the filters must apply.
 * It returns a {@link List} of {@link UtmModule} or a {@link Page} of {@link UtmModule} which fulfills the criteria.
 */
@Service
@Transactional(readOnly = true)
public class UtmModuleQueryService extends QueryService<UtmModule> {

    private final Logger log = LoggerFactory.getLogger(UtmModuleQueryService.class);

    private final UtmModuleRepository utmModuleRepository;

    public UtmModuleQueryService(UtmModuleRepository utmModuleRepository) {
        this.utmModuleRepository = utmModuleRepository;
    }

    /**
     * Return a {@link List} of {@link UtmModule} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmModule> findByCriteria(UtmModuleCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmModule> specification = createSpecification(criteria);
        return utmModuleRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmModule} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @param page     The page, which should be returned.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmModule> findByCriteria(UtmModuleCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmModule> specification = createSpecification(criteria);
        return utmModuleRepository.findAll(specification, page);
    }

    /**
     * Return the number of matching entities in the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the number of matching entities.
     */
    @Transactional(readOnly = true)
    public long countByCriteria(UtmModuleCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<UtmModule> specification = createSpecification(criteria);
        return utmModuleRepository.count(specification);
    }

    /**
     * Function to convert UtmModuleCriteria to a {@link Specification}
     */
    private Specification<UtmModule> createSpecification(UtmModuleCriteria criteria) {
        Specification<UtmModule> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmModule_.id));
            }
            if (criteria.getServerId() != null) {
                specification = specification.and(buildSpecification(criteria.getServerId(), UtmModule_.serverId));
            }
            if (criteria.getModuleName() != null) {
                specification = specification.and(buildSpecification(criteria.getModuleName(), UtmModule_.moduleName));
            }
            if (criteria.getModuleCategory() != null) {
                specification = specification.and(buildStringSpecification(criteria.getModuleCategory(), UtmModule_.moduleCategory));
            }
            if (criteria.getModuleActive() != null) {
                specification = specification.and(buildSpecification(criteria.getModuleActive(), UtmModule_.moduleActive));
            }
            if (criteria.getNeedsRestart() != null) {
                specification = specification.and(buildSpecification(criteria.getNeedsRestart(), UtmModule_.needsRestart));
            }
            if (criteria.getLiteVersion() != null) {
                specification = specification.and(buildSpecification(criteria.getLiteVersion(), UtmModule_.liteVersion));
            }
            if (criteria.getIsActivatable() != null) {
                specification = specification.and(buildSpecification(criteria.getIsActivatable(), UtmModule_.isActivatable));
            }
            if (criteria.getPrettyName() != null) {
                specification = specification.and(buildStringSpecification(criteria.getPrettyName(), UtmModule_.prettyName));
            }
        }
        return specification;
    }
}

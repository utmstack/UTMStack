package com.park.utmstack.service;

import com.park.utmstack.domain.UtmMenu;
import com.park.utmstack.domain.UtmMenu_;
import com.park.utmstack.repository.UtmMenuRepository;
import com.park.utmstack.service.dto.UtmMenuCriteria;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.domain.Specification;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import tech.jhipster.service.QueryService;

import java.util.List;

@Service
@Transactional(readOnly = true)
public class UtmMenuQueryService extends QueryService<UtmMenu> {
    private final Logger log = LoggerFactory.getLogger(UtmMenuQueryService.class);

    private final UtmMenuRepository menuRepository;

    public UtmMenuQueryService(UtmMenuRepository menuRepository) {
        this.menuRepository = menuRepository;
    }

    /**
     * Return a {@link List} of {@link UtmMenu} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmMenu> findByCriteria(UtmMenuCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmMenu> specification = createSpecification(criteria);
        return menuRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmMenu} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @param page     The page, which should be returned.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmMenu> findByCriteria(UtmMenuCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmMenu> specification = createSpecification(criteria);
        return menuRepository.findAll(specification, page);
    }

    /**
     * Function to convert UtmMenuCriteria to a {@link Specification}
     */
    private Specification<UtmMenu> createSpecification(UtmMenuCriteria criteria) {
        Specification<UtmMenu> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmMenu_.id));
            }
            if (criteria.getName() != null) {
                specification = specification.and(buildStringSpecification(criteria.getName(), UtmMenu_.name));
            }
            if (criteria.getUrl() != null) {
                specification = specification.and(buildStringSpecification(criteria.getUrl(), UtmMenu_.url));
            }
            if (criteria.getParentId() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getParentId(), UtmMenu_.parentId));
            }
            if (criteria.getType() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getType(), UtmMenu_.type));
            }
            if (criteria.getDashboardId() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getDashboardId(), UtmMenu_.dashboardId));
            }
            if (criteria.getModulesNameShort() != null) {
                specification = specification.and(buildStringSpecification(criteria.getModulesNameShort(), UtmMenu_.moduleNameShort));
            }
        }
        return specification;
    }


}

package com.park.utmstack.service.chart_builder;

import com.park.utmstack.domain.chart_builder.UtmDashboard;
import com.park.utmstack.domain.chart_builder.UtmDashboard_;
import com.park.utmstack.repository.chart_builder.UtmDashboardRepository;
import com.park.utmstack.service.dto.chart_builder.UtmDashboardCriteria;
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
public class UtmDashboardQueryService extends QueryService<UtmDashboard> {
    private final Logger log = LoggerFactory.getLogger(UtmDashboardQueryService.class);

    private final UtmDashboardRepository dashboardRepository;

    public UtmDashboardQueryService(UtmDashboardRepository dashboardRepository) {
        this.dashboardRepository = dashboardRepository;
    }


    @Transactional(readOnly = true)
    public List<UtmDashboard> findByCriteria(UtmDashboardCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmDashboard> specification = createSpecification(criteria);
        return dashboardRepository.findAll(specification);
    }

    @Transactional(readOnly = true)
    public Page<UtmDashboard> findByCriteria(UtmDashboardCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmDashboard> specification = createSpecification(criteria);
        return dashboardRepository.findAll(specification, page);
    }

    private Specification<UtmDashboard> createSpecification(UtmDashboardCriteria criteria) {
        Specification<UtmDashboard> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmDashboard_.id));
            }
            if (criteria.getName() != null) {
                specification = specification.and(buildStringSpecification(criteria.getName(), UtmDashboard_.name));
            }
            if (criteria.getCreatedDate() != null) {
                specification = specification.and(
                    buildRangeSpecification(criteria.getCreatedDate(), UtmDashboard_.createdDate));
            }
            if (criteria.getModifiedDate() != null) {
                specification = specification.and(
                    buildRangeSpecification(criteria.getModifiedDate(), UtmDashboard_.modifiedDate));
            }
            if (criteria.getUserCreated() != null) {
                specification = specification.and(
                    buildStringSpecification(criteria.getUserCreated(), UtmDashboard_.userCreated));
            }
            if (criteria.getUserModified() != null) {
                specification = specification.and(
                    buildStringSpecification(criteria.getUserModified(), UtmDashboard_.userModified));
            }
        }
        return specification;
    }
}

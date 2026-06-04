package com.park.utmstack.service.chart_builder;

import com.park.utmstack.domain.chart_builder.UtmDashboardVisualization;
import com.park.utmstack.domain.chart_builder.UtmDashboardVisualization_;
import com.park.utmstack.repository.chart_builder.UtmDashboardVisualizationRepository;
import com.park.utmstack.service.dto.chart_builder.UtmDashboardVisualizationCriteria;
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
public class UtmDashboardVisualizationQueryService extends QueryService<UtmDashboardVisualization> {
    private final Logger log = LoggerFactory.getLogger(UtmDashboardVisualizationQueryService.class);

    private final UtmDashboardVisualizationRepository dashboardVisualizationRepository;

    public UtmDashboardVisualizationQueryService(UtmDashboardVisualizationRepository dashboardVisualizationRepository) {
        this.dashboardVisualizationRepository = dashboardVisualizationRepository;
    }


    @Transactional(readOnly = true)
    public List<UtmDashboardVisualization> findByCriteria(UtmDashboardVisualizationCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmDashboardVisualization> specification = createSpecification(criteria);
        return dashboardVisualizationRepository.findAll(specification);
    }

    @Transactional(readOnly = true)
    public Page<UtmDashboardVisualization> findByCriteria(UtmDashboardVisualizationCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmDashboardVisualization> specification = createSpecification(criteria);
        return dashboardVisualizationRepository.findAll(specification, page);
    }

    private Specification<UtmDashboardVisualization> createSpecification(UtmDashboardVisualizationCriteria criteria) {
        Specification<UtmDashboardVisualization> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmDashboardVisualization_.id));
            }
            if (criteria.getIdVisualization() != null) {
                specification = specification.and(
                    buildRangeSpecification(criteria.getIdVisualization(), UtmDashboardVisualization_.idVisualization));
            }
            if (criteria.getIdDashboard() != null) {
                specification = specification.and(
                    buildRangeSpecification(criteria.getIdDashboard(), UtmDashboardVisualization_.idDashboard));
            }
            if (criteria.getOrder() != null) {
                specification = specification.and(
                    buildRangeSpecification(criteria.getOrder(), UtmDashboardVisualization_.order));
            }
        }
        return specification;
    }
}

package com.park.utmstack.service.chart_builder;

import com.park.utmstack.domain.chart_builder.UtmVisualization;
import com.park.utmstack.domain.chart_builder.UtmVisualization_;
import com.park.utmstack.repository.chart_builder.UtmVisualizationRepository;
import com.park.utmstack.service.dto.chart_builder.UtmVisualizationCriteria;
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
public class UtmVisualizationQueryService extends QueryService<UtmVisualization> {
    private final Logger log = LoggerFactory.getLogger(UtmVisualizationQueryService.class);

    private final UtmVisualizationRepository visualizationRepository;

    public UtmVisualizationQueryService(UtmVisualizationRepository visualizationRepository) {
        this.visualizationRepository = visualizationRepository;
    }

    @Transactional(readOnly = true)
    public List<UtmVisualization> findByCriteria(UtmVisualizationCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmVisualization> specification = createSpecification(criteria);
        return visualizationRepository.findAll(specification);
    }

    @Transactional(readOnly = true)
    public Page<UtmVisualization> findByCriteria(UtmVisualizationCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmVisualization> specification = createSpecification(criteria);
        return visualizationRepository.findAll(specification, page);
    }

    private Specification<UtmVisualization> createSpecification(UtmVisualizationCriteria criteria) {
        Specification<UtmVisualization> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmVisualization_.id));
            }
            if (criteria.getName() != null) {
                specification = specification.and(buildStringSpecification(criteria.getName(), UtmVisualization_.name));
            }
            if (criteria.getCreatedDate() != null) {
                specification = specification.and(
                    buildRangeSpecification(criteria.getCreatedDate(), UtmVisualization_.createdDate));
            }
            if (criteria.getModifiedDate() != null) {
                specification = specification.and(
                    buildRangeSpecification(criteria.getModifiedDate(), UtmVisualization_.modifiedDate));
            }
            if (criteria.getUserCreated() != null) {
                specification = specification.and(
                    buildStringSpecification(criteria.getUserCreated(), UtmVisualization_.userCreated));
            }
            if (criteria.getUserModified() != null) {
                specification = specification.and(
                    buildStringSpecification(criteria.getUserModified(), UtmVisualization_.userModified));
            }
            if (criteria.getChartType() != null) {
                specification = specification.and(
                    buildStringSpecification(criteria.getChartType(), UtmVisualization_.chType));
            }
            if (criteria.getIdPattern() != null) {
                specification = specification.and(
                    buildRangeSpecification(criteria.getIdPattern(), UtmVisualization_.idPattern));
            }
        }
        return specification;
    }
}

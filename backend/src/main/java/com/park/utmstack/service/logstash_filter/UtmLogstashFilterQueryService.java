package com.park.utmstack.service.logstash_filter;

import com.park.utmstack.domain.logstash_filter.UtmLogstashFilter;
import com.park.utmstack.domain.logstash_filter.UtmLogstashFilter_;
import com.park.utmstack.repository.logstash_filter.UtmLogstashFilterRepository;
import com.park.utmstack.service.dto.logstash_filter.UtmLogstashFilterCriteria;
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
public class UtmLogstashFilterQueryService extends QueryService<UtmLogstashFilter> {

    private final Logger log = LoggerFactory.getLogger(UtmLogstashFilterQueryService.class);

    private final UtmLogstashFilterRepository logstashFilterRepository;

    public UtmLogstashFilterQueryService(UtmLogstashFilterRepository logstashFilterRepository) {
        this.logstashFilterRepository = logstashFilterRepository;
    }

    @Transactional(readOnly = true)
    public List<UtmLogstashFilter> findByCriteria(UtmLogstashFilterCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmLogstashFilter> specification = createSpecification(criteria);
        return logstashFilterRepository.findAll(specification);
    }

    @Transactional(readOnly = true)
    public Page<UtmLogstashFilter> findByCriteria(UtmLogstashFilterCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmLogstashFilter> specification = createSpecification(criteria);
        return logstashFilterRepository.findAll(specification, page);
    }

    @Transactional(readOnly = true)
    public long countByCriteria(UtmLogstashFilterCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<UtmLogstashFilter> specification = createSpecification(criteria);
        return logstashFilterRepository.count(specification);
    }

    private Specification<UtmLogstashFilter> createSpecification(UtmLogstashFilterCriteria criteria) {
        Specification<UtmLogstashFilter> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmLogstashFilter_.id));
            }
            if (criteria.getFilterName() != null) {
                specification = specification.and(buildStringSpecification(criteria.getFilterName(), UtmLogstashFilter_.filterName));
            }
            if (criteria.getFilterGroupId() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getFilterGroupId(), UtmLogstashFilter_.filterGroupId));
            }
            if (criteria.getIsActive() != null) {
                specification = specification.and(buildSpecification(criteria.getIsActive(), UtmLogstashFilter_.isActive));
            }
        }
        return specification;
    }
}

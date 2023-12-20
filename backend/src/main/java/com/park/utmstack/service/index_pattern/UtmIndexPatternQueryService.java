package com.park.utmstack.service.index_pattern;

import com.park.utmstack.domain.index_pattern.UtmIndexPattern;
import com.park.utmstack.domain.index_pattern.UtmIndexPattern_;
import com.park.utmstack.repository.index_pattern.UtmIndexPatternRepository;
import com.park.utmstack.service.dto.index_pattern.UtmIndexPatternCriteria;
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
public class UtmIndexPatternQueryService extends QueryService<UtmIndexPattern> {
    private final Logger log = LoggerFactory.getLogger(UtmIndexPatternQueryService.class);

    private final UtmIndexPatternRepository indexPatternRepository;

    public UtmIndexPatternQueryService(UtmIndexPatternRepository indexPatternRepository) {
        this.indexPatternRepository = indexPatternRepository;
    }

    @Transactional(readOnly = true)
    public List<UtmIndexPattern> findByCriteria(UtmIndexPatternCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmIndexPattern> specification = createSpecification(criteria);
        return indexPatternRepository.findAll(specification);
    }

    @Transactional(readOnly = true)
    public Page<UtmIndexPattern> findByCriteria(UtmIndexPatternCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmIndexPattern> specification = createSpecification(criteria);
        return indexPatternRepository.findAll(specification, page);
    }

    private Specification<UtmIndexPattern> createSpecification(UtmIndexPatternCriteria criteria) {
        Specification<UtmIndexPattern> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmIndexPattern_.id));
            }
            if (criteria.getPattern() != null) {
                specification = specification.and(buildStringSpecification(criteria.getPattern(), UtmIndexPattern_.pattern));
            }
            if (criteria.getPatternModule() != null) {
                specification = specification.and(buildStringSpecification(criteria.getPatternModule(), UtmIndexPattern_.patternModule));
            }
            if (criteria.getPatternSystem() != null) {
                specification = specification.and(buildSpecification(criteria.getPatternSystem(), UtmIndexPattern_.patternSystem));
            }
            if (criteria.getIsActive() != null) {
                specification = specification.and(buildSpecification(criteria.getIsActive(), UtmIndexPattern_.isActive));
            }
        }
        return specification;
    }
}

package com.park.utmstack.service.index_pattern;

import com.park.utmstack.domain.index_pattern.UtmIndexPattern;
import com.park.utmstack.domain.index_pattern.UtmIndexPattern_;
import com.park.utmstack.repository.index_pattern.UtmIndexPatternRepository;
import com.park.utmstack.service.dto.index_pattern.UtmIndexPatternCriteria;
import com.park.utmstack.service.dto.index_pattern.UtmIndexPatternField;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageImpl;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.domain.Specification;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import tech.jhipster.service.QueryService;

import java.util.List;
import java.util.stream.Collectors;

@Service
@Transactional(readOnly = true)
public class UtmIndexPatternQueryService extends QueryService<UtmIndexPattern> {
    private final Logger log = LoggerFactory.getLogger(UtmIndexPatternQueryService.class);

    private final UtmIndexPatternRepository indexPatternRepository;

    private final ElasticsearchService elasticsearchService;

    public UtmIndexPatternQueryService(UtmIndexPatternRepository indexPatternRepository, ElasticsearchService elasticsearchService) {
        this.indexPatternRepository = indexPatternRepository;
        this.elasticsearchService = elasticsearchService;
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

    @Transactional(readOnly = true)
    public Page<UtmIndexPatternField> findWithFieldsByCriteria(UtmIndexPatternCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmIndexPattern> specification = createSpecification(criteria);

        Page<UtmIndexPattern> indexPatterns = indexPatternRepository.findAll(specification, page);

        List<UtmIndexPatternField> utmIndexPatternFields = indexPatterns.getContent().stream()
                .map(index -> {
                    try {
                        return new UtmIndexPatternField(index.getPattern(), elasticsearchService.getIndexProperties(index.getPattern()));
                    } catch (Exception e) {
                        log.error("Error fetching index properties for pattern: {}", index.getPattern(), e);
                        return new UtmIndexPatternField(index.getPattern(), null);
                    }
                }).collect(Collectors.toList());

        return new PageImpl<>(utmIndexPatternFields, page, indexPatterns.getTotalElements());
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

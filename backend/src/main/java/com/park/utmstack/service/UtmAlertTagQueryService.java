package com.park.utmstack.service;

import com.park.utmstack.domain.UtmAlertTag;
import com.park.utmstack.domain.UtmAlertTag_;
import com.park.utmstack.repository.UtmAlertTagRepository;
import com.park.utmstack.service.dto.UtmAlertTagCriteria;
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
public class UtmAlertTagQueryService extends QueryService<UtmAlertTag> {
    private final Logger log = LoggerFactory.getLogger(UtmAlertTagQueryService.class);

    private final UtmAlertTagRepository alertTagRepository;

    public UtmAlertTagQueryService(UtmAlertTagRepository alertTagRepository) {
        this.alertTagRepository = alertTagRepository;
    }

    /**
     * Return a {@link List} of {@link UtmAlertTag} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmAlertTag> findByCriteria(UtmAlertTagCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmAlertTag> specification = createSpecification(criteria);
        return alertTagRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmAlertTag} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @param page     The page, which should be returned.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmAlertTag> findByCriteria(UtmAlertTagCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmAlertTag> specification = createSpecification(criteria);
        return alertTagRepository.findAll(specification, page);
    }

    /**
     * Function to convert UtmAlertTagCriteria to a {@link Specification}
     */
    private Specification<UtmAlertTag> createSpecification(UtmAlertTagCriteria criteria) {
        Specification<UtmAlertTag> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmAlertTag_.id));
            }
            if (criteria.getTagName() != null) {
                specification = specification.and(buildStringSpecification(criteria.getTagName(), UtmAlertTag_.tagName));
            }
            if (criteria.getSystemOwner() != null) {
                specification = specification.and(buildSpecification(criteria.getSystemOwner(), UtmAlertTag_.systemOwner));
            }
        }
        return specification;
    }
}

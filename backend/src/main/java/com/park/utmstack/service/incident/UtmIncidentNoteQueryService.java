package com.park.utmstack.service.incident;

import com.park.utmstack.domain.incident.UtmIncidentNote;
import com.park.utmstack.domain.incident.UtmIncidentNote_;
import com.park.utmstack.repository.incident.UtmIncidentNoteRepository;
import com.park.utmstack.service.dto.incident.UtmIncidentNoteCriteria;
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
 * Service for executing complex queries for UtmIncidentNote entities in the database.
 * The main input is a {@link UtmIncidentNoteCriteria} which gets converted to {@link Specification},
 * in a way that all the filters must apply.
 * It returns a {@link List} of {@link UtmIncidentNote} or a {@link Page} of {@link UtmIncidentNote} which fulfills the criteria.
 */
@Service
@Transactional(readOnly = true)
public class UtmIncidentNoteQueryService extends QueryService<UtmIncidentNote> {

    private final Logger log = LoggerFactory.getLogger(UtmIncidentNoteQueryService.class);

    private final UtmIncidentNoteRepository utmIncidentNoteRepository;

    public UtmIncidentNoteQueryService(UtmIncidentNoteRepository utmIncidentNoteRepository) {
        this.utmIncidentNoteRepository = utmIncidentNoteRepository;
    }

    /**
     * Return a {@link List} of {@link UtmIncidentNote} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmIncidentNote> findByCriteria(UtmIncidentNoteCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmIncidentNote> specification = createSpecification(criteria);
        return utmIncidentNoteRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmIncidentNote} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @param page     The page, which should be returned.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmIncidentNote> findByCriteria(UtmIncidentNoteCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmIncidentNote> specification = createSpecification(criteria);
        return utmIncidentNoteRepository.findAll(specification, page);
    }

    /**
     * Return the number of matching entities in the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the number of matching entities.
     */
    @Transactional(readOnly = true)
    public long countByCriteria(UtmIncidentNoteCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<UtmIncidentNote> specification = createSpecification(criteria);
        return utmIncidentNoteRepository.count(specification);
    }

    /**
     * Function to convert UtmIncidentNoteCriteria to a {@link Specification}
     */
    private Specification<UtmIncidentNote> createSpecification(UtmIncidentNoteCriteria criteria) {
        Specification<UtmIncidentNote> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmIncidentNote_.id));
            }
            if (criteria.getIncidentId() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getIncidentId(), UtmIncidentNote_.incidentId));
            }
            if (criteria.getNoteText() != null) {
                specification = specification.and(buildStringSpecification(criteria.getNoteText(), UtmIncidentNote_.noteText));
            }
            if (criteria.getNoteSendDate() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getNoteSendDate(), UtmIncidentNote_.noteSendDate));
            }
            if (criteria.getNoteSendBy() != null) {
                specification = specification.and(buildStringSpecification(criteria.getNoteSendBy(), UtmIncidentNote_.noteSendBy));
            }
        }
        return specification;
    }
}

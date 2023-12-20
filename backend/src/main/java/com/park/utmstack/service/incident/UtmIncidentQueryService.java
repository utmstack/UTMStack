package com.park.utmstack.service.incident;

import com.park.utmstack.domain.incident.UtmIncident;
import com.park.utmstack.domain.incident.UtmIncident_;
import com.park.utmstack.repository.incident.UtmIncidentRepository;
import com.park.utmstack.service.UserService;
import com.park.utmstack.service.dto.incident.IncidentUserAssignedDTO;
import com.park.utmstack.service.dto.incident.UtmIncidentCriteria;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.domain.Specification;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import tech.jhipster.service.QueryService;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.stream.Collectors;

/**
 * Service for executing complex queries for UtmIncident entities in the database.
 * The main input is a {@link UtmIncidentCriteria} which gets converted to {@link Specification},
 * in a way that all the filters must apply.
 * It returns a {@link List} of {@link UtmIncident} or a {@link Page} of {@link UtmIncident} which fulfills the criteria.
 */
@Service
@Transactional(readOnly = true)
public class UtmIncidentQueryService extends QueryService<UtmIncident> {

    private final Logger log = LoggerFactory.getLogger(UtmIncidentQueryService.class);

    private final UtmIncidentRepository utmIncidentRepository;

    private final UserService userService;

    public UtmIncidentQueryService(UtmIncidentRepository utmIncidentRepository, UserService userService) {
        this.utmIncidentRepository = utmIncidentRepository;
        this.userService = userService;
    }

    /**
     * Return a {@link List} of {@link UtmIncident} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmIncident> findByCriteria(UtmIncidentCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmIncident> specification = createSpecification(criteria);
        return utmIncidentRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmIncident} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @param page     The page, which should be returned.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmIncident> findByCriteria(UtmIncidentCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmIncident> specification = createSpecification(criteria);
        return utmIncidentRepository.findAll(specification, page);
    }

    /**
     * Return a {@link Page} of {@link UtmIncident} which matches the criteria from the database
     *
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<IncidentUserAssignedDTO> getAllUsersAssigned() {
        List<UtmIncident> incidents = utmIncidentRepository.findAll();
        List<Long> userIds = new ArrayList<>();
        for (UtmIncident incident : incidents) {
            if (incident.getIncidentAssignedTo() != null) {
                List<Long> ids = Arrays.stream(incident.getIncidentAssignedTo().split(",")).mapToLong(Long::parseLong).boxed().collect(Collectors.toList());
                userIds.addAll(ids);
            }
        }
        return userService.getAllUsersIn(userIds.stream().distinct().collect(Collectors.toList())).stream().map(user -> {
            IncidentUserAssignedDTO dto = new IncidentUserAssignedDTO();
            dto.setId(user.getId());
            dto.setLogin(user.getLogin());
            return dto;
        }).collect(Collectors.toList());

    }

    /**
     * Return the number of matching entities in the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the number of matching entities.
     */
    @Transactional(readOnly = true)
    public long countByCriteria(UtmIncidentCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<UtmIncident> specification = createSpecification(criteria);
        return utmIncidentRepository.count(specification);
    }

    /**
     * Function to convert UtmIncidentCriteria to a {@link Specification}
     */
    private Specification<UtmIncident> createSpecification(UtmIncidentCriteria criteria) {
        Specification<UtmIncident> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmIncident_.id));
            }
            if (criteria.getIncidentName() != null) {
                specification = specification.and(buildStringSpecification(criteria.getIncidentName(), UtmIncident_.incidentName));
            }
            if (criteria.getIncidentDescription() != null) {
                specification = specification.and(buildStringSpecification(criteria.getIncidentDescription(), UtmIncident_.incidentDescription));
            }
            if (criteria.getIncidentStatus() != null) {
                specification = specification.and(buildSpecification(criteria.getIncidentStatus(), UtmIncident_.incidentStatus));
            }
            if (criteria.getIncidentAssignedTo() != null) {
                specification = specification.and(buildStringSpecification(criteria.getIncidentAssignedTo(), UtmIncident_.incidentAssignedTo));
            }
            if (criteria.getIncidentCreatedDate() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getIncidentCreatedDate(), UtmIncident_.incidentCreatedDate));
            }
            if (criteria.getIncidentSeverity() != null) {
                specification = specification.and(buildSpecification(criteria.getIncidentSeverity(), UtmIncident_.incidentSeverity));
            }
        }
        return specification;
    }
}

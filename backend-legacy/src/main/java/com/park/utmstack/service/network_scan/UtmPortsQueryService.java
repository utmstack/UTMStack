package com.park.utmstack.service.network_scan;

import com.park.utmstack.domain.network_scan.UtmPorts;
import com.park.utmstack.domain.network_scan.UtmPorts_;
import com.park.utmstack.repository.network_scan.UtmPortsRepository;
import com.park.utmstack.service.dto.network_scan.UtmPortsCriteria;
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
 * Service for executing complex queries for UtmOpenPort entities in the database.
 * The main input is a {@link UtmPortsCriteria} which gets converted to {@link Specification},
 * in a way that all the filters must apply.
 * It returns a {@link List} of {@link UtmPorts}.
 */
@Service
@Transactional(readOnly = true)
public class UtmPortsQueryService extends QueryService<UtmPorts> {

    private final Logger log = LoggerFactory.getLogger(UtmPortsQueryService.class);

    private final UtmPortsRepository utmOpenPortRepository;

    public UtmPortsQueryService(UtmPortsRepository utmOpenPortRepository) {
        this.utmOpenPortRepository = utmOpenPortRepository;
    }

    /**
     * Return a {@link List} of {@link UtmPorts} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmPorts> findByCriteria(UtmPortsCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmPorts> specification = createSpecification(criteria);
        return utmOpenPortRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmPorts} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @param page     The page, which should be returned.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmPorts> findByCriteria(UtmPortsCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmPorts> specification = createSpecification(criteria);
        return utmOpenPortRepository.findAll(specification, page);
    }

    /**
     * Return the number of matching entities in the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the number of matching entities.
     */
    @Transactional(readOnly = true)
    public long countByCriteria(UtmPortsCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<UtmPorts> specification = createSpecification(criteria);
        return utmOpenPortRepository.count(specification);
    }

    /**
     * Function to convert UtmOpenPortCriteria to a {@link Specification}
     */
    private Specification<UtmPorts> createSpecification(UtmPortsCriteria criteria) {
        Specification<UtmPorts> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmPorts_.id));
            }
            if (criteria.getPort() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getScanId(), UtmPorts_.scanId));
            }
            if (criteria.getPort() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getPort(), UtmPorts_.port));
            }
            if (criteria.getTcp() != null) {
                specification = specification.and(buildStringSpecification(criteria.getTcp(), UtmPorts_.tcp));
            }
            if (criteria.getUdp() != null) {
                specification = specification.and(buildStringSpecification(criteria.getUdp(), UtmPorts_.udp));
            }
        }
        return specification;
    }
}

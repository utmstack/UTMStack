package com.park.utmstack.service.network_scan;

import com.park.utmstack.domain.network_scan.UtmNetworkScan;
import com.park.utmstack.domain.network_scan.UtmNetworkScan_;
import com.park.utmstack.repository.network_scan.UtmNetworkScanRepository;
import com.park.utmstack.service.dto.network_scan.UtmNetworkScanCriteria;
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
 * Service for executing complex queries for UtmNetworkScan entities in the database.
 * The main input is a {@link UtmNetworkScanCriteria} which gets converted to {@link Specification},
 * in a way that all the filters must apply.
 * It returns a {@link List} of {@link UtmNetworkScan} or a {@link Page} of {@link UtmNetworkScan} which fulfills the criteria.
 */
@Service
@Transactional(readOnly = true)
public class UtmNetworkScanQueryService extends QueryService<UtmNetworkScan> {

    private final Logger log = LoggerFactory.getLogger(UtmNetworkScanQueryService.class);

    private final UtmNetworkScanRepository utmNetworkScanRepository;

    public UtmNetworkScanQueryService(UtmNetworkScanRepository utmNetworkScanRepository) {
        this.utmNetworkScanRepository = utmNetworkScanRepository;
    }

    /**
     * Return a {@link List} of {@link UtmNetworkScan} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public List<UtmNetworkScan> findByCriteria(UtmNetworkScanCriteria criteria) {
        log.debug("find by criteria : {}", criteria);
        final Specification<UtmNetworkScan> specification = createSpecification(criteria);
        return utmNetworkScanRepository.findAll(specification);
    }

    /**
     * Return a {@link Page} of {@link UtmNetworkScan} which matches the criteria from the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @param page     The page, which should be returned.
     * @return the matching entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmNetworkScan> findByCriteria(UtmNetworkScanCriteria criteria, Pageable page) {
        log.debug("find by criteria : {}, page: {}", criteria, page);
        final Specification<UtmNetworkScan> specification = createSpecification(criteria);
        return utmNetworkScanRepository.findAll(specification, page);
    }

    /**
     * Return the number of matching entities in the database
     *
     * @param criteria The object which holds all the filters, which the entities should match.
     * @return the number of matching entities.
     */
    @Transactional(readOnly = true)
    public long countByCriteria(UtmNetworkScanCriteria criteria) {
        log.debug("count by criteria : {}", criteria);
        final Specification<UtmNetworkScan> specification = createSpecification(criteria);
        return utmNetworkScanRepository.count(specification);
    }

    /**
     * Function to convert UtmNetworkScanCriteria to a {@link Specification}
     */
    private Specification<UtmNetworkScan> createSpecification(UtmNetworkScanCriteria criteria) {
        Specification<UtmNetworkScan> specification = Specification.where(null);
        if (criteria != null) {
            if (criteria.getId() != null) {
                specification = specification.and(buildSpecification(criteria.getId(), UtmNetworkScan_.id));
            }
            if (criteria.getIp() != null) {
                specification = specification.and(buildStringSpecification(criteria.getIp(), UtmNetworkScan_.assetIp));
            }
            if (criteria.getMac() != null) {
                specification = specification.and(buildStringSpecification(criteria.getMac(), UtmNetworkScan_.assetMac));
            }
            if (criteria.getOs() != null) {
                specification = specification.and(buildStringSpecification(criteria.getOs(), UtmNetworkScan_.assetOs));
            }
            if (criteria.getName() != null) {
                specification = specification.and(buildStringSpecification(criteria.getName(), UtmNetworkScan_.assetName));
            }
            if (criteria.getAlive() != null) {
                specification = specification.and(buildSpecification(criteria.getAlive(), UtmNetworkScan_.assetAlive));
            }
            if (criteria.getStatus() != null) {
                specification = specification.and(buildSpecification(criteria.getStatus(), UtmNetworkScan_.assetStatus));
            }
            if (criteria.getDiscoveredAt() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getDiscoveredAt(), UtmNetworkScan_.discoveredAt));
            }
            if (criteria.getModifiedAt() != null) {
                specification = specification.and(buildRangeSpecification(criteria.getModifiedAt(), UtmNetworkScan_.modifiedAt));
            }
        }
        return specification;
    }
}

package com.park.utmstack.service;

import com.park.utmstack.domain.UtmAssetMetrics;
import com.park.utmstack.repository.UtmAssetMetricsRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;

/**
 * Service Implementation for managing UtmAssetMetrics.
 */
@Service
@Transactional
public class UtmAssetMetricsService {

    private final Logger log = LoggerFactory.getLogger(UtmAssetMetricsService.class);

    private final UtmAssetMetricsRepository utmAssetMetricsRepository;

    public UtmAssetMetricsService(UtmAssetMetricsRepository utmAssetMetricsRepository) {
        this.utmAssetMetricsRepository = utmAssetMetricsRepository;
    }

    /**
     * Save a utmAssetMetrics.
     *
     * @param utmAssetMetrics the entity to save
     * @return the persisted entity
     */
    public UtmAssetMetrics save(UtmAssetMetrics utmAssetMetrics) {
        log.debug("Request to save UtmAssetMetrics : {}", utmAssetMetrics);
        return utmAssetMetricsRepository.save(utmAssetMetrics);
    }

    /**
     * Get all the utmAssetMetrics.
     *
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public List<UtmAssetMetrics> findAll() {
        log.debug("Request to get all UtmAssetMetrics");
        return utmAssetMetricsRepository.findAll();
    }


    /**
     * Get one utmAssetMetrics by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmAssetMetrics> findOne(String id) {
        log.debug("Request to get UtmAssetMetrics : {}", id);
        return utmAssetMetricsRepository.findById(id);
    }

    /**
     * Delete the utmAssetMetrics by id.
     *
     * @param id the id of the entity
     */
    public void delete(String id) {
        log.debug("Request to delete UtmAssetMetrics : {}", id);
        utmAssetMetricsRepository.deleteById(id);
    }
}

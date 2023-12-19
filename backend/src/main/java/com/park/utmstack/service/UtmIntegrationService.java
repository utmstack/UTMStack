package com.park.utmstack.service;

import com.park.utmstack.domain.UtmIntegration;
import com.park.utmstack.repository.UtmIntegrationRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.Optional;

/**
 * Service Implementation for managing UtmIntegration.
 */
@Service
@Transactional
public class UtmIntegrationService {

    private final Logger log = LoggerFactory.getLogger(UtmIntegrationService.class);

    private final UtmIntegrationRepository utmIntegrationRepository;

    public UtmIntegrationService(UtmIntegrationRepository utmIntegrationRepository) {
        this.utmIntegrationRepository = utmIntegrationRepository;
    }

    /**
     * Save a utmIntegration.
     *
     * @param utmIntegration the entity to save
     * @return the persisted entity
     */
    public UtmIntegration save(UtmIntegration utmIntegration) {
        log.debug("Request to save UtmIntegration : {}", utmIntegration);
        return utmIntegrationRepository.save(utmIntegration);
    }

    /**
     * Get all the utmIntegrations.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmIntegration> findAll(Pageable pageable) {
        log.debug("Request to get all UtmIntegrations");
        return utmIntegrationRepository.findAll(pageable);
    }


    /**
     * Get one utmIntegration by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmIntegration> findOne(Long id) {
        log.debug("Request to get UtmIntegration : {}", id);
        return utmIntegrationRepository.findById(id);
    }

    /**
     * Delete the utmIntegration by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmIntegration : {}", id);
        utmIntegrationRepository.deleteById(id);
    }
}

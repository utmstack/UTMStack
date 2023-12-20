package com.park.utmstack.service;

import com.park.utmstack.domain.UtmClient;
import com.park.utmstack.repository.UtmClientRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;

/**
 * Service Implementation for managing UtmClient.
 */
@Service
@Transactional
public class UtmClientService {

    private final Logger log = LoggerFactory.getLogger(UtmClientService.class);

    private final UtmClientRepository utmClientRepository;

    public UtmClientService(UtmClientRepository utmClientRepository) {
        this.utmClientRepository = utmClientRepository;
    }

    /**
     * Save a utmClient.
     *
     * @param utmClient the entity to save
     * @return the persisted entity
     */
    public UtmClient save(UtmClient utmClient) {
        log.debug("Request to save UtmClient : {}", utmClient);
        return utmClientRepository.save(utmClient);
    }

    /**
     * Get all the utmClients.
     *
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public List<UtmClient> findAll() {
        log.debug("Request to get all UtmClients");
        return utmClientRepository.findAll();
    }


    /**
     * Get one utmClient by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmClient> findOne(Long id) {
        log.debug("Request to get UtmClient : {}", id);
        return utmClientRepository.findById(id);
    }

    /**
     * Delete the utmClient by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmClient : {}", id);
        utmClientRepository.deleteById(id);
    }


}

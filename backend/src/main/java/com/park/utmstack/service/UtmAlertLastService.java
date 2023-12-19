package com.park.utmstack.service;

import com.park.utmstack.domain.UtmAlertLast;
import com.park.utmstack.repository.UtmAlertLastRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;

/**
 * Service Implementation for managing UtmAlertLast.
 */
@Service
@Transactional
public class UtmAlertLastService {

    private final Logger log = LoggerFactory.getLogger(UtmAlertLastService.class);

    private final UtmAlertLastRepository utmAlertLastRepository;

    public UtmAlertLastService(UtmAlertLastRepository utmAlertLastRepository) {
        this.utmAlertLastRepository = utmAlertLastRepository;
    }

    /**
     * Save a utmAlertLast.
     *
     * @param utmAlertLast the entity to save
     * @return the persisted entity
     */
    public UtmAlertLast save(UtmAlertLast utmAlertLast) {
        log.debug("Request to save UtmAlertLast : {}", utmAlertLast);
        return utmAlertLastRepository.save(utmAlertLast);
    }

    /**
     * Get all the utmAlertLasts.
     *
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public List<UtmAlertLast> findAll() {
        log.debug("Request to get all UtmAlertLasts");
        return utmAlertLastRepository.findAll();
    }


    /**
     * Get one utmAlertLast by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmAlertLast> findOne(Long id) {
        log.debug("Request to get UtmAlertLast : {}", id);
        return utmAlertLastRepository.findById(id);
    }

    /**
     * Delete the utmAlertLast by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmAlertLast : {}", id);
        utmAlertLastRepository.deleteById(id);
    }
}

package com.park.utmstack.service;

import com.park.utmstack.domain.UtmServer;
import com.park.utmstack.repository.UtmServerRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;

/**
 * Service Implementation for managing UtmServer.
 */
@Service
@Transactional
public class UtmServerService {

    private final Logger log = LoggerFactory.getLogger(UtmServerService.class);

    private final UtmServerRepository utmServerRepository;

    public UtmServerService(UtmServerRepository utmServerRepository) {
        this.utmServerRepository = utmServerRepository;
    }

    /**
     * Save a utmServer.
     *
     * @param utmServer the entity to save
     * @return the persisted entity
     */
    public UtmServer save(UtmServer utmServer) {
        log.debug("Request to save UtmServer : {}", utmServer);
        return utmServerRepository.save(utmServer);
    }

    /**
     * Get all the utmServers.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmServer> findAll(Pageable pageable) {
        log.debug("Request to get all UtmServers");
        return utmServerRepository.findAll(pageable);
    }

    /**
     * Get all the utmServers but pagination mode is not included.
     *
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public List<UtmServer> findAll() {
        return utmServerRepository.findAll();
    }


    /**
     * Get one utmServer by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmServer> findOne(Long id) {
        log.debug("Request to get UtmServer : {}", id);
        return utmServerRepository.findById(id);
    }

    /**
     * Delete the utmServer by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmServer : {}", id);
        utmServerRepository.deleteById(id);
    }
}

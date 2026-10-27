package com.park.utmstack.service.incident_response;

import com.park.utmstack.domain.incident_response.UtmIncidentAction;
import com.park.utmstack.repository.incident_response.UtmIncidentActionRepository;
import com.park.utmstack.util.exceptions.ApiException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.Optional;

/**
 * Service Implementation for managing UtmIncidentAction.
 */
@Service
@Transactional
public class UtmIncidentActionService {

    private final Logger log = LoggerFactory.getLogger(UtmIncidentActionService.class);

    private final UtmIncidentActionRepository utmIncidentActionRepository;

    public UtmIncidentActionService(UtmIncidentActionRepository utmIncidentActionRepository) {
        this.utmIncidentActionRepository = utmIncidentActionRepository;
    }

    /**
     * Save a utmIncidentAction.
     *
     * <p>Seeded system presets (actionEditable=false) are immutable via API for
     * every role — analysts author their own actions, but the shared presets
     * (SHUTDOWN_SERVER, RUN_CMD, ...) cannot be rewritten or removed through
     * the service. Repairing a preset is a reseed/liquibase change.
     *
     * @param utmIncidentAction the entity to save
     * @return the persisted entity
     */
    public UtmIncidentAction save(UtmIncidentAction utmIncidentAction) {
        log.debug("Request to save UtmIncidentAction : {}", utmIncidentAction);
        if (utmIncidentAction.getId() != null) {
            requireEditable(utmIncidentAction.getId());
        }
        return utmIncidentActionRepository.save(utmIncidentAction);
    }

    private void requireEditable(Long id) {
        UtmIncidentAction existing = utmIncidentActionRepository.findById(id)
                .orElseThrow(() -> new ApiException("Incident action " + id + " not found", HttpStatus.NOT_FOUND));
        if (!Boolean.TRUE.equals(existing.getActionEditable())) {
            throw new ApiException("System incident action " + id + " is not editable", HttpStatus.FORBIDDEN);
        }
    }

    /**
     * Get all the utmIncidentActions.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmIncidentAction> findAll(Pageable pageable) {
        log.debug("Request to get all UtmIncidentActions");
        return utmIncidentActionRepository.findAll(pageable);
    }


    /**
     * Get one utmIncidentAction by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmIncidentAction> findOne(Long id) {
        log.debug("Request to get UtmIncidentAction : {}", id);
        return utmIncidentActionRepository.findById(id);
    }

    /**
     * Delete the utmIncidentAction by id.
     *
     * <p>System presets (actionEditable=false) cannot be removed via the API —
     * see {@link #save(UtmIncidentAction)}.
     *
     * @param id the id of the utmIncidentAction
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmIncidentAction : {}", id);
        requireEditable(id);
        utmIncidentActionRepository.deleteById(id);
    }
}

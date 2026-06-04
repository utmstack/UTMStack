package com.park.utmstack.service.incident_response;

import com.park.utmstack.domain.incident_response.UtmIncidentActionCommand;
import com.park.utmstack.repository.incident_response.UtmIncidentActionCommandRepository;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.Optional;

/**
 * Service Implementation for managing UtmIncidentActionCommand.
 */
@Service
@Transactional
public class UtmIncidentActionCommandService {

    private static final String CLASSNAME = "UtmIncidentActionCommandService";

    private final UtmIncidentActionCommandRepository incidentActionCommandRepository;

    public UtmIncidentActionCommandService(UtmIncidentActionCommandRepository incidentActionCommandRepository) {
        this.incidentActionCommandRepository = incidentActionCommandRepository;
    }

    /**
     * Save a utmIncidentActionCommand.
     *
     * @param incidentActionCommand the entity to save
     * @return the persisted entity
     */
    public UtmIncidentActionCommand save(UtmIncidentActionCommand incidentActionCommand) throws Exception {
        final String ctx = CLASSNAME + ".save";
        try {
            return incidentActionCommandRepository.save(incidentActionCommand);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Get all the utmIncidentActionCommands.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmIncidentActionCommand> findAll(Pageable pageable) throws Exception {
        final String ctx = CLASSNAME + ".findAll";
        try {
            return incidentActionCommandRepository.findAll(pageable);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Delete the utmIncidentActionCommand by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) throws Exception {
        final String ctx = CLASSNAME + ".findAll";
        try {
            incidentActionCommandRepository.deleteById(id);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public Optional<String> getCommand(String osPlatform, Long actionId) throws Exception {
        final String ctx = CLASSNAME + ".getSpecificCommand";
        try {
            osPlatform = osPlatform.equalsIgnoreCase("windows") ? osPlatform : "linux";
            return incidentActionCommandRepository.findCommand(osPlatform, actionId);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }
}

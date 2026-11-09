package com.park.utmstack.service.incident_response;

import com.park.utmstack.domain.incident_response.UtmIncidentAction;
import com.park.utmstack.domain.incident_response.UtmIncidentActionCommand;
import com.park.utmstack.repository.incident_response.UtmIncidentActionCommandRepository;
import com.park.utmstack.repository.incident_response.UtmIncidentActionRepository;
import com.park.utmstack.util.exceptions.ApiException;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpStatus;
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
    private final UtmIncidentActionRepository utmIncidentActionRepository;

    public UtmIncidentActionCommandService(UtmIncidentActionCommandRepository incidentActionCommandRepository,
                                           UtmIncidentActionRepository utmIncidentActionRepository) {
        this.incidentActionCommandRepository = incidentActionCommandRepository;
        this.utmIncidentActionRepository = utmIncidentActionRepository;
    }

    /**
     * Save a utmIncidentActionCommand.
     *
     * <p>Command rows inherit immutability from their parent action: the shell
     * strings under seeded system presets (SHUTDOWN_SERVER, RUN_CMD, ...) cannot
     * be rewritten or added through the API by any role.
     *
     * @param incidentActionCommand the entity to save
     * @return the persisted entity
     */
    public UtmIncidentActionCommand save(UtmIncidentActionCommand incidentActionCommand) throws Exception {
        final String ctx = CLASSNAME + ".save";
        try {
            if (incidentActionCommand.getId() != null) {
                UtmIncidentActionCommand stored = incidentActionCommandRepository.findById(incidentActionCommand.getId())
                        .orElseThrow(() -> new ApiException("Incident action command " + incidentActionCommand.getId() + " not found", HttpStatus.NOT_FOUND));
                requireParentEditable(stored.getActionId());
            }
            requireParentEditable(incidentActionCommand.getActionId());
            return incidentActionCommandRepository.save(incidentActionCommand);
        } catch (ApiException e) {
            throw e;
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    private void requireParentEditable(Long actionId) {
        if (actionId == null) {
            throw new ApiException("Incident action command requires an actionId", HttpStatus.BAD_REQUEST);
        }
        UtmIncidentAction parent = utmIncidentActionRepository.findById(actionId)
                .orElseThrow(() -> new ApiException("Incident action " + actionId + " not found", HttpStatus.NOT_FOUND));
        if (!Boolean.TRUE.equals(parent.getActionEditable())) {
            throw new ApiException("Incident action command belongs to system action " + actionId + " and is not editable", HttpStatus.FORBIDDEN);
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
     * Delete a utmIncidentActionCommand by id.
     *
     * <p>System-preset commands cannot be removed via the API — see {@link #save(UtmIncidentActionCommand)}.
     *
     * @param id the id of the utmIncidentActionCommand
     */
    public void delete(Long id) throws Exception {
        final String ctx = CLASSNAME + ".delete";
        try {
            UtmIncidentActionCommand cmd = incidentActionCommandRepository.findById(id)
                    .orElseThrow(() -> new ApiException("Incident action command " + id + " not found", HttpStatus.NOT_FOUND));
            requireParentEditable(cmd.getActionId());
            incidentActionCommandRepository.deleteById(id);
        } catch (ApiException e) {
            throw e;
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

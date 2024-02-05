package com.park.utmstack.service.incident_response;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.incident_response.UtmIncidentVariable;
import com.park.utmstack.repository.incident_response.UtmIncidentVariableRepository;
import com.park.utmstack.util.CipherUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.Optional;

/**
 * Service Implementation for managing UtmIncidentVariable.
 */
@Service
@Transactional
public class UtmIncidentVariableService {

    private final Logger log = LoggerFactory.getLogger(UtmIncidentVariableService.class);

    private final UtmIncidentVariableRepository utmIncidentVariableRepository;

    public UtmIncidentVariableService(UtmIncidentVariableRepository utmIncidentVariableRepository) {

        this.utmIncidentVariableRepository = utmIncidentVariableRepository;
    }

    /**
     * Save a utmIncidentVariable.
     *
     * @param utmIncidentVariable the entity to save
     * @return the persisted entity
     */
    public UtmIncidentVariable save(UtmIncidentVariable utmIncidentVariable) {
        log.debug("Request to save UtmIncidentVariable : {}", utmIncidentVariable);
        if (utmIncidentVariable.isSecret()) {
            String currentValue = utmIncidentVariable.getVariableValue();
            utmIncidentVariable.setVariableValue(CipherUtil.encrypt(currentValue, System.getenv(Constants.ENV_ENCRYPTION_KEY)));
        }
        return utmIncidentVariableRepository.save(utmIncidentVariable);
    }

    /**
     * Get all the utmIncidentVariables.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmIncidentVariable> findAll(Pageable pageable) {
        log.debug("Request to get all UtmIncidentVariables");
        return utmIncidentVariableRepository.findAll(pageable);
    }


    /**
     * Get one utmIncidentVariable by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmIncidentVariable> findOne(Long id) {
        log.debug("Request to get UtmIncidentVariable : {}", id);
        return utmIncidentVariableRepository.findById(id);
    }

    /**
     * Delete the utmIncidentVariable by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmIncidentVariable : {}", id);
        utmIncidentVariableRepository.deleteById(id);
    }
}

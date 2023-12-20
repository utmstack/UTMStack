package com.park.utmstack.service.incident;

import com.park.utmstack.domain.incident.UtmIncidentAlert;
import com.park.utmstack.domain.incident.enums.IncidentHistoryActionEnum;
import com.park.utmstack.repository.incident.UtmIncidentAlertRepository;
import com.park.utmstack.service.dto.incident.AlertIncidentStatusChangeDTO;
import com.park.utmstack.service.incident.util.ResolveIncidentStatus;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;

/**
 * Service Implementation for managing UtmIncidentAlert.
 */
@Service
@Transactional
public class UtmIncidentAlertService {
    private final String CLASSNAME = "UtmIncidentAlertService";
    private final Logger log = LoggerFactory.getLogger(UtmIncidentAlertService.class);

    private final UtmIncidentAlertRepository utmIncidentAlertRepository;

    private final UtmIncidentHistoryService utmIncidentHistoryService;

    public UtmIncidentAlertService(UtmIncidentAlertRepository utmIncidentAlertRepository,
                                   UtmIncidentHistoryService utmIncidentHistoryService) {
        this.utmIncidentAlertRepository = utmIncidentAlertRepository;
        this.utmIncidentHistoryService = utmIncidentHistoryService;
    }

    /**
     * Save a utmIncidentAlert.
     *
     * @param utmIncidentAlert the entity to save
     * @return the persisted entity
     */
    public UtmIncidentAlert save(UtmIncidentAlert utmIncidentAlert) {
        log.debug("Request to save UtmIncidentAlert : {}", utmIncidentAlert);
        return utmIncidentAlertRepository.save(utmIncidentAlert);
    }

    public void updateAlertStatusByAlertId(AlertIncidentStatusChangeDTO alertIncidentStatusChangeDTO) {
        final String ctx = "UtmIncidentAlertService.updateAlertStatusByAlertId";
        try {
            utmIncidentAlertRepository.updateAlertStatusByAlertIdIn(alertIncidentStatusChangeDTO.getAlertIds(), alertIncidentStatusChangeDTO.getStatus());
            utmIncidentHistoryService.createHistory(IncidentHistoryActionEnum.INCIDENT_ALERT_STATUS_CHANGED, alertIncidentStatusChangeDTO.getIncidentId(),
                "Alert status changed", "Alerts changed to " + ResolveIncidentStatus.incidentLabelByInteger(alertIncidentStatusChangeDTO.getStatus()));
        } catch (Exception e) {
            throw new RuntimeException(ctx + " - " + e.getMessage());
        }
    }

    /**
     * Get all the utmIncidentAlerts.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmIncidentAlert> findAll(Pageable pageable) {
        log.debug("Request to get all UtmIncidentAlerts");
        return utmIncidentAlertRepository.findAll(pageable);
    }


    /**
     * Get one utmIncidentAlert by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmIncidentAlert> findOne(Long id) {
        log.debug("Request to get UtmIncidentAlert : {}", id);
        return utmIncidentAlertRepository.findById(id);
    }

    /**
     * Delete the utmIncidentAlert by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        log.debug("Request to delete UtmIncidentAlert : {}", id);
        utmIncidentAlertRepository.deleteById(id);
    }

    public void saveAll(List<UtmIncidentAlert> utmIncidentAlerts) {
        utmIncidentAlertRepository.saveAll(utmIncidentAlerts);
    }

    public List<UtmIncidentAlert> findAllByIncidentId(Long incidentId) {
        return utmIncidentAlertRepository.findAllByIncidentId(incidentId);
    }
}

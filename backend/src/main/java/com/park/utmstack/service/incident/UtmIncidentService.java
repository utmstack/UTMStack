package com.park.utmstack.service.incident;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.incident.UtmIncident;
import com.park.utmstack.domain.incident.UtmIncidentAlert;
import com.park.utmstack.domain.incident.enums.IncidentHistoryActionEnum;
import com.park.utmstack.domain.incident.enums.IncidentStatusEnum;
import com.park.utmstack.domain.shared_types.AlertType;
import com.park.utmstack.repository.incident.UtmIncidentRepository;
import com.park.utmstack.service.MailService;
import com.park.utmstack.service.UserService;
import com.park.utmstack.service.UtmAlertService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.incident.AddToIncidentDTO;
import com.park.utmstack.service.dto.incident.AlertIncidentStatusChangeDTO;
import com.park.utmstack.service.dto.incident.NewIncidentDTO;
import com.park.utmstack.service.dto.incident.RelatedIncidentAlertsDTO;
import com.park.utmstack.service.incident.util.ResolveIncidentStatus;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.CollectionUtils;

import javax.validation.Valid;
import java.util.Arrays;
import java.util.Date;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

/**
 * Service Implementation for managing UtmIncident.
 */
@Service
@Transactional
public class UtmIncidentService {


    private final String CLASSNAME = "UtmIncidentService";
    private final Logger log = LoggerFactory.getLogger(UtmIncidentService.class);

    private final UtmIncidentRepository utmIncidentRepository;
    private final UtmIncidentAlertService utmIncidentAlertService;

    private final UtmIncidentHistoryService utmIncidentHistoryService;

    private final UtmAlertService utmAlertService;

    private final ApplicationEventService eventService;

    private final MailService mailService;

    public UtmIncidentService(UtmIncidentRepository utmIncidentRepository,
                              UtmIncidentAlertService utmIncidentAlertService,
                              UtmIncidentHistoryService utmIncidentHistoryService,
                              UserService userService, UtmAlertService utmAlertService,
                              ApplicationEventService eventService, MailService mailService) {
        this.utmIncidentRepository = utmIncidentRepository;
        this.utmIncidentAlertService = utmIncidentAlertService;
        this.utmIncidentHistoryService = utmIncidentHistoryService;
        this.utmAlertService = utmAlertService;
        this.eventService = eventService;
        this.mailService = mailService;
    }

    /**
     * Save a utmIncident.
     *
     * @param utmIncident the entity to save
     * @return the persisted entity
     */
    public UtmIncident save(UtmIncident utmIncident) {
        final String ctx = ".save";
        log.debug("Request to save UtmIncident : {}", utmIncident);
        try {
            return utmIncidentRepository.save(utmIncident);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            throw new RuntimeException(msg);
        }
    }

    /**
     * Save a utmIncident.
     *
     * @param utmIncident the entity to save
     * @return the persisted entity
     */
    public UtmIncident changeStatus(UtmIncident utmIncident) {
        final String ctx = ".changeStatus";
        try {
            log.debug("Request to save UtmIncident : {}", utmIncident);
            String oldIncident = ResolveIncidentStatus.incidentLabel(utmIncidentRepository.
                findById(utmIncident.getId())
                .orElseThrow(() -> new RuntimeException("Incident not found")));

            UtmIncident incident = utmIncidentRepository.save(utmIncident);
            List<UtmIncidentAlert> alerts = utmIncidentAlertService.findAllByIncidentId(incident.getId());

            if (!CollectionUtils.isEmpty(alerts)) {
                List<String> alertIds = alerts.stream().map(UtmIncidentAlert::getAlertId).collect(Collectors.toList());

                AlertIncidentStatusChangeDTO alertIncidentStatusChangeDTO = new AlertIncidentStatusChangeDTO();
                alertIncidentStatusChangeDTO.setAlertIds(alertIds);
                alertIncidentStatusChangeDTO.setIncidentId(incident.getId());
                alertIncidentStatusChangeDTO.setStatus(utmIncident.getIncidentStatus().getValue());
                utmIncidentAlertService.updateAlertStatusByAlertId(alertIncidentStatusChangeDTO);

                utmAlertService.updateStatus(alertIds, ResolveIncidentStatus.getAlertStatus(utmIncident), utmIncident.getIncidentSolution());

            }

            String msg = String.format("Incident status changed from %s to %s ", oldIncident, ResolveIncidentStatus.incidentLabel(utmIncident));
            if (incident.getIncidentStatus().equals(IncidentStatusEnum.COMPLETED)) {
                msg += (" with solution: " + incident.getIncidentSolution());
            }

            utmIncidentHistoryService.createHistory(IncidentHistoryActionEnum.INCIDENT_STATUS_CHANGE, incident.getId(), "Alert change status", msg);

            return incident;
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            throw new RuntimeException(msg);
        }
    }

    /**
     * Save a utmIncident.
     *
     * @param newIncidentDTO the entity to save
     * @return the persisted entity
     */
    public UtmIncident createIncident(NewIncidentDTO newIncidentDTO) {
        final String ctx = CLASSNAME + ".createIncident";
        try {
            UtmIncident utmIncident = new UtmIncident();
            utmIncident.setIncidentName(newIncidentDTO.getIncidentName());
            utmIncident.setIncidentDescription(newIncidentDTO.getIncidentDescription());
            utmIncident.setIncidentStatus(IncidentStatusEnum.OPEN);
            Integer severity = newIncidentDTO.getAlertList().stream()
                .mapToInt(RelatedIncidentAlertsDTO::getAlertSeverity).max().orElse(0);
            utmIncident.setIncidentSeverity(severity);
            utmIncident.setIncidentAssignedTo(newIncidentDTO.getIncidentAssignedTo());
            utmIncident.setIncidentCreatedDate(new Date().toInstant());

            UtmIncident savedIncident = utmIncidentRepository.save(utmIncident);

            saveRelatedAlerts(newIncidentDTO.getAlertList(), savedIncident.getId());

            sendIncidentsEmail(newIncidentDTO.getAlertList().stream().map(RelatedIncidentAlertsDTO::getAlertId).collect(Collectors.toList()), savedIncident);

            String historyMessage = String.format("Incident created with %d alerts", newIncidentDTO.getAlertList().size());
            utmIncidentHistoryService.createHistory(IncidentHistoryActionEnum.INCIDENT_CREATED, savedIncident.getId(), "Incident Created", historyMessage);

            return savedIncident;
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            throw new RuntimeException(msg);
        }
    }

    /**
     * Save a utmIncident.
     *
     * @param addToIncidentDTO the entity to add alerts
     * @return the persisted entity
     */
    public UtmIncident addAlertsIncident(@Valid AddToIncidentDTO addToIncidentDTO) {
        final String ctx = CLASSNAME + ".addAlertsIncident";
        try {
            log.debug("Request to add alert to UtmIncident : {}", addToIncidentDTO);
            UtmIncident utmIncident = utmIncidentRepository.findById(addToIncidentDTO.getIncidentId()).orElseThrow(() -> new RuntimeException(ctx + ": Incident not found"));
            saveRelatedAlerts(addToIncidentDTO.getAlertList(), utmIncident.getId());
            String historyMessage = String.format("New %d alerts added to incident", addToIncidentDTO.getAlertList().size());
            utmIncidentHistoryService.createHistory(IncidentHistoryActionEnum.INCIDENT_ALERT_ADD, utmIncident.getId(), "New alerts added to incident", historyMessage);
            return utmIncident;
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            throw new RuntimeException(msg);
        }
    }

    @Async
    public void saveRelatedAlerts(List<RelatedIncidentAlertsDTO> alertsDTO, Long incidentId) {
        final String ctx = CLASSNAME + ".saveRelatedAlerts";
        try {
            List<UtmIncidentAlert> utmIncidentAlerts = alertsDTO.stream().map(alert -> {
                UtmIncidentAlert utmIncidentAlert = new UtmIncidentAlert();
                utmIncidentAlert.setIncidentId(incidentId);
                utmIncidentAlert.setAlertName(alert.getAlertName());
                utmIncidentAlert.setAlertSeverity(alert.getAlertSeverity());
                utmIncidentAlert.setAlertStatus(alert.getAlertStatus());
                utmIncidentAlert.setAlertId(alert.getAlertId());
                return utmIncidentAlert;
            }).collect(Collectors.toList());

            utmIncidentAlertService.saveAll(utmIncidentAlerts);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            throw new RuntimeException(msg);
        }
    }

    /**
     * Get all the utmIncidents.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmIncident> findAll(Pageable pageable) {
        final String ctx = CLASSNAME + ".findAll";
        try {
            log.debug("Request to get all UtmIncidents");
            return utmIncidentRepository.findAll(pageable);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            throw new RuntimeException(msg);
        }
    }


    /**
     * Get one utmIncident by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmIncident> findOne(Long id) {
        final String ctx = CLASSNAME + ".findAll";
        try {
            return utmIncidentRepository.findById(id);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            throw new RuntimeException(msg);
        }
    }

    /**
     * Delete the utmIncident by id.
     *
     * @param id the id of the entity
     */
    public void delete(Long id) {
        final String ctx = CLASSNAME + ".delete";
        try {
            log.debug("Request to delete UtmIncident : {}", id);
            utmIncidentRepository.deleteById(id);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            throw new RuntimeException(msg);
        }
    }

    private void sendIncidentsEmail(List<String> alertIds, UtmIncident utmIncident) {
        final String ctx = CLASSNAME + ".sendIncidentsEmail";
        try {
            List<AlertType> alerts = utmAlertService.getAlertsByIds(alertIds);

            if (CollectionUtils.isEmpty(alerts))
                return;

            String[] addressToNotify = Constants.CFG.get(Constants.PROP_ALERT_ADDRESS_TO_NOTIFY_INCIDENTS)
                .replace(" ", "").split(",");

            mailService.sendIncidentEmail(Arrays.asList(addressToNotify), alerts, utmIncident);

        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            eventService.createEvent(msg, ApplicationEventType.ERROR);
        }
    }
}

package com.park.utmstack.web.rest.incident;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.incident.UtmIncident;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.incident.*;
import com.park.utmstack.service.incident.UtmIncidentAlertService;
import com.park.utmstack.service.incident.UtmIncidentQueryService;
import com.park.utmstack.service.incident.UtmIncidentService;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.util.CollectionUtils;
import org.springframework.web.bind.annotation.*;
import tech.jhipster.web.util.ResponseUtil;

import javax.validation.Valid;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

/**
 * REST controller for managing UtmIncident.
 */
@RestController
@RequestMapping("/api")
public class UtmIncidentResource {
    private final String CLASS_NAME = "UtmIncidentResource";
    private final Logger log = LoggerFactory.getLogger(UtmIncidentResource.class);

    private static final String ENTITY_NAME = "utmIncident";

    private final UtmIncidentService utmIncidentService;

    private final UtmIncidentAlertService utmIncidentAlertService;

    private final UtmIncidentQueryService utmIncidentQueryService;
    private final ApplicationEventService applicationEventService;

    public UtmIncidentResource(UtmIncidentService utmIncidentService,
                               UtmIncidentAlertService utmIncidentAlertService,
                               UtmIncidentQueryService utmIncidentQueryService,
                               ApplicationEventService applicationEventService) {
        this.utmIncidentService = utmIncidentService;
        this.utmIncidentAlertService = utmIncidentAlertService;
        this.utmIncidentQueryService = utmIncidentQueryService;
        this.applicationEventService = applicationEventService;
    }

    /**
     * Creates a new incident based on the provided details.
     *
     * This endpoint accepts a {@link NewIncidentDTO} object, validates the data,
     * and attempts to create a new incident. The process includes:
     * - Verifying that the alert list is not empty.
     * - Checking if any of the provided alerts are already associated with another incident.
     * - Creating the incident if all validations pass.
     *
     * @param newIncidentDTO the DTO containing the details of the incident to create, including associated alerts.
     * @return a {@link ResponseEntity} containing:
     *         - HTTP 201 (Created) if the incident is successfully created.
     *         - HTTP 400 (Bad Request) if the alert list is empty.
     *         - HTTP 409 (Conflict) if one or more alerts are already associated with another incident.
     *         - HTTP 500 (Internal Server Error) if an unexpected error occurs during processing.
     * @throws IllegalArgumentException if the input data is invalid.
     */
    @PostMapping("/utm-incidents")
    public ResponseEntity<UtmIncident> createUtmIncident(@Valid @RequestBody NewIncidentDTO newIncidentDTO) {
        final String ctx = ".createUtmIncident";
        try {
            if (CollectionUtils.isEmpty(newIncidentDTO.getAlertList())) {
                String msg = ctx + ": A new incident has to have at least one alert related";
                log.error(msg);
                applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
                return UtilResponse.buildErrorResponse(HttpStatus.BAD_REQUEST, msg);
            }

            List<String> alertIds = newIncidentDTO.getAlertList().stream()
                    .map(RelatedIncidentAlertsDTO::getAlertId)
                    .collect(Collectors.toList());

            List<String> alertsFound = utmIncidentAlertService.existsAnyAlert(alertIds);

            if (!alertsFound.isEmpty()) {
                String alertIdsList = String.join(", ", alertIds);
                String msg = "Some alerts are already linked to another incident. Alert IDs: " + alertIdsList + ". Check the related incidents for more details.";
                log.error(msg);
                applicationEventService.createEvent(ctx + ": " + msg , ApplicationEventType.ERROR);
                return UtilResponse.buildErrorResponse(HttpStatus.CONFLICT, utmIncidentAlertService.formatAlertMessage(alertsFound));
            }


            return ResponseEntity.ok(utmIncidentService.createIncident(newIncidentDTO));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * POST /utm-incidents/add-alerts : Add alerts to an existing utmIncident.
     *
     * This endpoint allows users to associate a list of alerts with an existing utmIncident.
     * If any of the provided alerts are already linked to another incident, a conflict response is returned.
     *
     * @param addToIncidentDTO the DTO containing the details of the utmIncident and the list of alerts to add
     * @return the ResponseEntity with:
     *         - status 201 (Created) and the updated utmIncident if successful,
     *         - status 400 (Bad Request) if the alert list is empty,
     *         - status 409 (Conflict) if some alerts are already associated with another incident,
     *         - status 500 (Internal Server Error) if an unexpected error occurs.
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PostMapping("/utm-incidents/add-alerts")
    public ResponseEntity<UtmIncident> addAlertsToUtmIncident(@Valid @RequestBody AddToIncidentDTO addToIncidentDTO) throws URISyntaxException {
        final String ctx = ".addAlertsToUtmIncident";
        try {
            log.debug("REST request to save UtmIncident : {}", addToIncidentDTO);
            if (CollectionUtils.isEmpty(addToIncidentDTO.getAlertList())) {
                throw new BadRequestAlertException("Add utmIncident cannot already have an empty related alerts", ENTITY_NAME, "alertList");
            }
            List<String> alertIds = addToIncidentDTO.getAlertList().stream()
                    .map(RelatedIncidentAlertsDTO::getAlertId)
                    .collect(Collectors.toList());

            List<String> alertsFound = utmIncidentAlertService.existsAnyAlert(alertIds);

            if (!alertsFound.isEmpty()) {
                String alertIdsList = String.join(", ", alertIds);
                String msg = "Some alerts are already linked to another incident. Alert IDs: " + alertIdsList + ". Check the related incidents for more details.";
                log.error(msg);
                applicationEventService.createEvent(ctx + ": " + msg , ApplicationEventType.ERROR);
                return UtilResponse.buildErrorResponse(HttpStatus.CONFLICT, utmIncidentAlertService.formatAlertMessage(alertsFound));
            }
            UtmIncident result = utmIncidentService.addAlertsIncident(addToIncidentDTO);
            return ResponseEntity.created(new URI("/api/utm-incidents/add-alerts" + result.getId()))
                .headers(HeaderUtil.createEntityCreationAlert(ENTITY_NAME, result.getId().toString()))
                .body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(HeaderUtil.createFailureAlert(CLASS_NAME, null, msg)).body(null);
        }
    }

    /**
     * PUT  /utm-incidents : Updates an existing utmIncident.
     *
     * @param utmIncident the utmIncident to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmIncident,
     * or with status 400 (Bad Request) if the utmIncident is not valid,
     * or with status 500 (Internal Server Error) if the utmIncident couldn't be updated
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PutMapping("/utm-incidents/change-status")
    public ResponseEntity<UtmIncident> updateUtmIncident(@Valid @RequestBody UtmIncident utmIncident) throws URISyntaxException {
        final String ctx = ".updateUtmIncident";
        try {
            log.debug("REST request to update UtmIncident : {}", utmIncident);
            if (utmIncident.getId() == null) {
                throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "idnull");
            }
            UtmIncident result = utmIncidentService.changeStatus(utmIncident);
            return ResponseEntity.ok()
                .headers(HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, utmIncident.getId().toString()))
                .body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(HeaderUtil.createFailureAlert(CLASS_NAME, null, msg)).body(null);
        }
    }

    /**
     * GET  /utm-incidents : get all the utmIncidents.
     *
     * @param pageable the pagination information
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the list of utmIncidents in body
     */
    @GetMapping("/utm-incidents")
    public ResponseEntity<List<UtmIncident>> getAllUtmIncidents(UtmIncidentCriteria criteria, Pageable pageable) {
        final String ctx = ".getAllUtmIncidents";
        try {
            log.debug("REST request to get UtmIncidents by criteria: {}", criteria);
            Page<UtmIncident> page = utmIncidentQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-incidents");
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(HeaderUtil.createFailureAlert(CLASS_NAME, null, msg)).body(null);
        }
    }

    /**
     * GET  /utm-incidents/users-assigned : get all users assigned to incidents.
     *
     * @return the ResponseEntity with status 200 (OK) and the list of IncidentUserAssignedDTO in body
     */
    @GetMapping("/utm-incidents/users-assigned")
    public ResponseEntity<List<IncidentUserAssignedDTO>> getAllUserAssigned() {
        final String ctx = ".getAllUserAssigned";
        try {
            return ResponseEntity.ok().body(utmIncidentQueryService.getAllUsersAssigned());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(HeaderUtil.createFailureAlert(CLASS_NAME, null, msg)).body(null);
        }
    }

    /**
     * GET  /utm-incidents/:id : get the "id" utmIncident.
     *
     * @param id the id of the utmIncident to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmIncident, or with status 404 (Not Found)
     */
    @GetMapping("/utm-incidents/{id}")
    public ResponseEntity<UtmIncident> getUtmIncident(@PathVariable Long id) {
        final String ctx = ".getUtmIncident";
        try {
            log.debug("REST request to get UtmIncident : {}", id);
            Optional<UtmIncident> utmIncident = utmIncidentService.findOne(id);
            return ResponseUtil.wrapOrNotFound(utmIncident);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(HeaderUtil.createFailureAlert(CLASS_NAME, null, msg)).body(null);
        }
    }
}

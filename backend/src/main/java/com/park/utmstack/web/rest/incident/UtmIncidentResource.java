package com.park.utmstack.web.rest.incident;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.incident.UtmIncident;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.incident.AddToIncidentDTO;
import com.park.utmstack.service.dto.incident.IncidentUserAssignedDTO;
import com.park.utmstack.service.dto.incident.NewIncidentDTO;
import com.park.utmstack.service.dto.incident.UtmIncidentCriteria;
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

    private final UtmIncidentQueryService utmIncidentQueryService;
    private final ApplicationEventService applicationEventService;

    public UtmIncidentResource(UtmIncidentService utmIncidentService, UtmIncidentQueryService utmIncidentQueryService, ApplicationEventService applicationEventService) {
        this.utmIncidentService = utmIncidentService;
        this.utmIncidentQueryService = utmIncidentQueryService;
        this.applicationEventService = applicationEventService;
    }

    /**
     * POST  /utm-incidents : Create a new utmIncident.
     *
     * @param newIncidentDTO the utmIncident to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmIncident, or with status 400 (Bad Request) if the utmIncident has already an ID
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
            return ResponseEntity.ok(utmIncidentService.createIncident(newIncidentDTO));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * POST  /utm-incidents : Create a new utmIncident.
     *
     * @param addToIncidentDTO the utmIncident to add alerts to
     * @return the ResponseEntity with status 201 (Created) and with body the new utmIncident, or with status 400 (Bad Request) if the utmIncident has already an ID
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

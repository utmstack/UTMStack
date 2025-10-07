package com.park.utmstack.web.rest.incident;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.incident.UtmIncident;
import com.park.utmstack.service.dto.incident.*;
import com.park.utmstack.service.incident.UtmIncidentQueryService;
import com.park.utmstack.service.incident.UtmIncidentService;
import com.park.utmstack.util.exceptions.NoAlertsProvidedException;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.ResponseEntity;
import org.springframework.util.CollectionUtils;
import org.springframework.web.bind.annotation.*;
import com.park.utmstack.aop.logging.AuditEvent;

import javax.validation.Valid;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing UtmIncident.
 */
@RestController
@RequiredArgsConstructor
@Slf4j
@RequestMapping("/api")
public class UtmIncidentResource {

    private static final String ENTITY_NAME = "utmIncident";

    private final UtmIncidentService utmIncidentService;

    private final UtmIncidentQueryService utmIncidentQueryService;


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
    @AuditEvent(
        attemptType = ApplicationEventType.INCIDENT_CREATION_ATTEMPT,
        attemptMessage = "Attempt to create a new incident initiated",
        successType = ApplicationEventType.INCIDENT_CREATION_SUCCESS,
        successMessage = "Incident created successfully"
    )
    public ResponseEntity<UtmIncident> createUtmIncident(@Valid @RequestBody NewIncidentDTO newIncidentDTO) {
        final String ctx = ".createUtmIncident";

        if (CollectionUtils.isEmpty(newIncidentDTO.getAlertList())) {
            String msg = ctx + ": A new incident has to have at least one alert related";
            throw new NoAlertsProvidedException(ctx + ": " + msg);
        }

        return ResponseEntity.ok(utmIncidentService.createIncident(newIncidentDTO));
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
    @AuditEvent(
        attemptType = ApplicationEventType.INCIDENT_ALERT_ADD_ATTEMPT,
        attemptMessage = "Attempt to add alerts to incident initiated",
        successType = ApplicationEventType.INCIDENT_ALERT_ADD_SUCCESS,
        successMessage = "Alerts added to incident successfully"
    )
    public ResponseEntity<UtmIncident> addAlertsToUtmIncident(@Valid @RequestBody AddToIncidentDTO addToIncidentDTO) throws URISyntaxException {

        if (CollectionUtils.isEmpty(addToIncidentDTO.getAlertList())) {
            throw new NoAlertsProvidedException("Add utmIncident cannot already have an empty related alerts");
        }

        UtmIncident result = utmIncidentService.addAlertsIncident(addToIncidentDTO);
           return ResponseEntity.created(new URI("/api/utm-incidents/add-alerts" + result.getId()))
                .headers(HeaderUtil.createEntityCreationAlert(ENTITY_NAME, result.getId().toString()))
                .body(result);
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
    @AuditEvent(
        attemptType = ApplicationEventType.INCIDENT_UPDATE_ATTEMPT,
        attemptMessage = "Attempt to update incident status initiated",
        successType = ApplicationEventType.INCIDENT_UPDATE_SUCCESS,
        successMessage = "Incident status updated successfully"
    )
    public ResponseEntity<UtmIncident> updateUtmIncident(@Valid @RequestBody UtmIncident utmIncident) {

        if (utmIncident.getId() == null) {
            throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "idnull");
        }
        UtmIncident result = utmIncidentService.changeStatus(utmIncident);

        return ResponseEntity.ok()
                .headers(HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, utmIncident.getId().toString()))
                .body(result);
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

        Page<UtmIncident> page = utmIncidentQueryService.findByCriteria(criteria, pageable);
        HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-incidents");
        return ResponseEntity.ok().headers(headers).body(page.getContent());
    }

    /**
     * GET  /utm-incidents/users-assigned : get all users assigned to incidents.
     *
     * @return the ResponseEntity with status 200 (OK) and the list of IncidentUserAssignedDTO in body
     */
    @GetMapping("/utm-incidents/users-assigned")
    public ResponseEntity<List<IncidentUserAssignedDTO>> getAllUserAssigned() {
        return ResponseEntity.ok().body(utmIncidentQueryService.getAllUsersAssigned());
    }

    /**
     * GET  /utm-incidents/:id : get the "id" utmIncident.
     *
     * @param id the id of the utmIncident to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmIncident, or with status 404 (Not Found)
     */
    @GetMapping("/utm-incidents/{id}")
    public ResponseEntity<UtmIncident> getUtmIncident(@PathVariable Long id) {

        Optional<UtmIncident> utmIncident = utmIncidentService.findOne(id);
        return tech.jhipster.web.util.ResponseUtil.wrapOrNotFound(utmIncident);

    }
}

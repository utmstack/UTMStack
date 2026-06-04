package com.park.utmstack.web.rest.incident;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.incident.UtmIncidentAlert;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.incident.AlertIncidentStatusChangeDTO;
import com.park.utmstack.service.dto.incident.UtmIncidentAlertCriteria;
import com.park.utmstack.service.incident.UtmIncidentAlertQueryService;
import com.park.utmstack.service.incident.UtmIncidentAlertService;
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
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.List;

/**
 * REST controller for managing UtmIncidentAlert.
 */
@RestController
@RequestMapping("/api")
public class UtmIncidentAlertResource {
    private final String CLASS_NAME = "UtmIncidentAlertResource";

    private final Logger log = LoggerFactory.getLogger(UtmIncidentAlertResource.class);

    private static final String ENTITY_NAME = "utmIncidentAlert";

    private final UtmIncidentAlertService utmIncidentAlertService;

    private final UtmIncidentAlertQueryService utmIncidentAlertQueryService;

    private final ApplicationEventService applicationEventService;

    public UtmIncidentAlertResource(UtmIncidentAlertService utmIncidentAlertService, UtmIncidentAlertQueryService utmIncidentAlertQueryService,
                                    ApplicationEventService applicationEventService) {
        this.utmIncidentAlertService = utmIncidentAlertService;
        this.utmIncidentAlertQueryService = utmIncidentAlertQueryService;
        this.applicationEventService = applicationEventService;
    }

    /**
     * POST  /utm-incident-alerts : Create a new utmIncidentAlert.
     *
     * @param utmIncidentAlert the utmIncidentAlert to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmIncidentAlert, or with status 400 (Bad Request) if the utmIncidentAlert has already an ID
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PostMapping("/utm-incident-alerts")
    public ResponseEntity<UtmIncidentAlert> createUtmIncidentAlert(@Valid @RequestBody UtmIncidentAlert utmIncidentAlert) throws URISyntaxException {
        final String ctx = ".updateIncidentAlertStatus";
        try {
            log.debug("REST request to save UtmIncidentAlert : {}", utmIncidentAlert);
            if (utmIncidentAlert.getId() != null) {
                throw new BadRequestAlertException("A new utmIncidentAlert cannot already have an ID", ENTITY_NAME, "idexists");
            }
            UtmIncidentAlert result = utmIncidentAlertService.save(utmIncidentAlert);
            return ResponseEntity.created(new URI("/api/utm-incident-alerts/" + result.getId()))
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
     * POST  /utm-incident-alerts : Create a new utmIncidentAlert.
     *
     * @param alertIncidentStatusChangeDTO the AlertIncidentStatusChangeDTO to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmIncidentAlert, or with status 400 (Bad Request) if the utmIncidentAlert has already an ID
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PostMapping("/utm-incident-alerts/update-status")
    public ResponseEntity<Void> updateIncidentAlertStatus(@Valid @RequestBody AlertIncidentStatusChangeDTO alertIncidentStatusChangeDTO) throws URISyntaxException {
        final String ctx = ".updateIncidentAlertStatus";
        try {
            log.debug("REST request to save UtmIncidentAlert : {}", alertIncidentStatusChangeDTO);
            if (alertIncidentStatusChangeDTO.getAlertIds().isEmpty()) {
                throw new BadRequestAlertException("A new utmIncidentAlert cannot already have an ID", ENTITY_NAME, "idexists");
            }
            utmIncidentAlertService.updateAlertStatusByAlertId(alertIncidentStatusChangeDTO);
            return ResponseEntity.ok().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(HeaderUtil.createFailureAlert(CLASS_NAME, null, msg)).body(null);
        }
    }

    /**
     * PUT  /utm-incident-alerts : Updates an existing utmIncidentAlert.
     *
     * @param utmIncidentAlert the utmIncidentAlert to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmIncidentAlert,
     * or with status 400 (Bad Request) if the utmIncidentAlert is not valid,
     * or with status 500 (Internal Server Error) if the utmIncidentAlert couldn't be updated
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PutMapping("/utm-incident-alerts")
    public ResponseEntity<UtmIncidentAlert> updateUtmIncidentAlert(@Valid @RequestBody UtmIncidentAlert utmIncidentAlert) throws URISyntaxException {
        final String ctx = ".updateUtmIncidentAlert";
        try {
            if (utmIncidentAlert.getId() == null) {
                throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "idnull");
            }
            UtmIncidentAlert result = utmIncidentAlertService.save(utmIncidentAlert);
            return ResponseEntity.ok()
                .headers(HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, utmIncidentAlert.getId().toString()))
                .body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(HeaderUtil.createFailureAlert(CLASS_NAME, null, msg)).body(null);
        }
    }

    /**
     * GET  /utm-incident-alerts : get all the utmIncidentAlerts.
     *
     * @param pageable the pagination information
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the list of utmIncidentAlerts in body
     */
    @GetMapping("/utm-incident-alerts")
    public ResponseEntity<List<UtmIncidentAlert>> getAllUtmIncidentAlerts(UtmIncidentAlertCriteria criteria, Pageable pageable) {
        final String ctx = ".getAllUtmIncidentAlerts";
        try {
            log.debug("REST request to get UtmIncidentAlerts by criteria: {}", criteria);
            Page<UtmIncidentAlert> page = utmIncidentAlertQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-incident-alerts");
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(HeaderUtil.createFailureAlert(CLASS_NAME, null, msg)).body(null);
        }
    }

    /**
     * DELETE  /utm-incident-alerts/:id : delete the "id" utmIncidentAlert.
     *
     * @param id the id of the utmIncidentAlert to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/utm-incident-alerts/{id}")
    public ResponseEntity<Void> deleteUtmIncidentAlert(@PathVariable Long id) {
        final String ctx = ".getAllUtmIncidentAlerts";
        try {
            log.debug("REST request to delete UtmIncidentAlert : {}", id);
            utmIncidentAlertService.delete(id);
            return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, id.toString())).build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(HeaderUtil.createFailureAlert(CLASS_NAME, null, msg)).body(null);
        }
    }
}

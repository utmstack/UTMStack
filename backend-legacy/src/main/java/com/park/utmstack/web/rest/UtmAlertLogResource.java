package com.park.utmstack.web.rest;

import com.park.utmstack.domain.UtmAlertLog;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.UtmAlertLogQueryService;
import com.park.utmstack.service.UtmAlertLogService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.UtmAlertLogCriteria;
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
import tech.jhipster.web.util.ResponseUtil;

import javax.validation.Valid;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing UtmAlertLog.
 */
@RestController
@RequestMapping("/api")
public class UtmAlertLogResource {

    private final Logger log = LoggerFactory.getLogger(UtmAlertLogResource.class);
    private static final String CLASSNAME = "UtmAlertLogResource";

    private static final String ENTITY_NAME = "utmAlertLog";
    private final UtmAlertLogService utmAlertLogService;
    private final UtmAlertLogQueryService utmAlertLogQueryService;
    private final ApplicationEventService applicationEventService;

    public UtmAlertLogResource(UtmAlertLogService utmAlertLogService,
                               UtmAlertLogQueryService utmAlertLogQueryService,
                               ApplicationEventService applicationEventService) {
        this.utmAlertLogService = utmAlertLogService;
        this.utmAlertLogQueryService = utmAlertLogQueryService;
        this.applicationEventService = applicationEventService;
    }

    /**
     * POST  /utm-alert-logs : Create a new utmAlertLog.
     *
     * @param utmAlertLog the utmAlertLog to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmAlertLog, or with status 400 (Bad Request) if the utmAlertLog has already an ID
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PostMapping("/utm-alert-logs")
    public ResponseEntity<UtmAlertLog> createUtmAlertLog(@Valid @RequestBody UtmAlertLog utmAlertLog) throws URISyntaxException {
        final String ctx = CLASSNAME + ".createUtmAlertLog";
        try {
            if (utmAlertLog.getId() != null)
                throw new BadRequestAlertException("A new utmAlertLog cannot already have an ID", ENTITY_NAME, "idexists");

            UtmAlertLog result = utmAlertLogService.save(utmAlertLog);
            return ResponseEntity.created(new URI("/api/utm-alert-logs/" + result.getId()))
                .headers(HeaderUtil.createEntityCreationAlert(ENTITY_NAME, result.getId().toString()))
                .body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * PUT  /utm-alert-logs : Updates an existing utmAlertLog.
     *
     * @param utmAlertLog the utmAlertLog to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmAlertLog,
     * or with status 400 (Bad Request) if the utmAlertLog is not valid,
     * or with status 500 (Internal Server Error) if the utmAlertLog couldn't be updated
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PutMapping("/utm-alert-logs")
    public ResponseEntity<UtmAlertLog> updateUtmAlertLog(@Valid @RequestBody UtmAlertLog utmAlertLog) {
        final String ctx = CLASSNAME + ".updateUtmAlertLog";
        try {
            if (utmAlertLog.getId() == null)
                throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "idnull");

            UtmAlertLog result = utmAlertLogService.save(utmAlertLog);
            return ResponseEntity.ok()
                .headers(HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, utmAlertLog.getId().toString()))
                .body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * GET  /utm-alert-logs : get all the utmAlertLogs.
     *
     * @param pageable the pagination information
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the list of utmAlertLogs in body
     */
    @GetMapping("/utm-alert-logs")
    public ResponseEntity<List<UtmAlertLog>> getAllUtmAlertLogs(UtmAlertLogCriteria criteria, Pageable pageable) {
        final String ctx = CLASSNAME + ".getAllUtmAlertLogs";
        try {
            Page<UtmAlertLog> page = utmAlertLogQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-alert-logs");
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * GET  /utm-alert-logs/count : count all the utmAlertLogs.
     *
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the count in body
     */
    @GetMapping("/utm-alert-logs/count")
    public ResponseEntity<Long> countUtmAlertLogs(UtmAlertLogCriteria criteria) {
        final String ctx = CLASSNAME + ".countUtmAlertLogs";
        try {
            return ResponseEntity.ok().body(utmAlertLogQueryService.countByCriteria(criteria));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * GET  /utm-alert-logs/:id : get the "id" utmAlertLog.
     *
     * @param id the id of the utmAlertLog to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmAlertLog, or with status 404 (Not Found)
     */
    @GetMapping("/utm-alert-logs/{id}")
    public ResponseEntity<UtmAlertLog> getUtmAlertLog(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".getUtmAlertLog";
        try {
            Optional<UtmAlertLog> utmAlertLog = utmAlertLogService.findOne(id);
            return ResponseUtil.wrapOrNotFound(utmAlertLog);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * DELETE  /utm-alert-logs/:id : delete the "id" utmAlertLog.
     *
     * @param id the id of the utmAlertLog to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/utm-alert-logs/{id}")
    public ResponseEntity<Void> deleteUtmAlertLog(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".deleteUtmAlertLog";
        try {
            utmAlertLogService.delete(id);
            return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, id.toString())).build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }
}

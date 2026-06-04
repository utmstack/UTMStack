package com.park.utmstack.web.rest;

import com.park.utmstack.domain.UtmSchedule;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.UtmScheduleService;
import com.park.utmstack.service.application_events.ApplicationEventService;
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
 * REST controller for managing UtmSchedule.
 */
@RestController
@RequestMapping("/api")
public class UtmScheduleResource {

    private final Logger log = LoggerFactory.getLogger(UtmScheduleResource.class);

    private static final String ENTITY_NAME = "utmSchedule";
    private static final String CLASSNAME = "UtmScheduleResource";

    private final UtmScheduleService utmScheduleService;
    private final ApplicationEventService applicationEventService;

    public UtmScheduleResource(UtmScheduleService utmScheduleService,
                               ApplicationEventService applicationEventService) {
        this.utmScheduleService = utmScheduleService;
        this.applicationEventService = applicationEventService;
    }

    /**
     * POST  /utm-schedules : Create a new utmSchedule.
     *
     * @param utmSchedule the utmSchedule to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmSchedule, or with status 400 (Bad Request) if the utmSchedule has already an ID
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PostMapping("/utm-schedules")
    public ResponseEntity<UtmSchedule> createUtmSchedule(@Valid @RequestBody UtmSchedule utmSchedule) throws URISyntaxException {
        final String ctx = CLASSNAME + ".createUtmSchedule";
        try {
            if (utmSchedule.getId() != null)
                throw new BadRequestAlertException("A new utmSchedule cannot already have an ID", ENTITY_NAME, "idexists");

            UtmSchedule result = utmScheduleService.save(utmSchedule);
            return ResponseEntity.created(new URI("/api/utm-schedules/" + result.getId()))
                .headers(HeaderUtil.createEntityCreationAlert(ENTITY_NAME, result.getId().toString()))
                .body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(null);
        }
    }

    /**
     * PUT  /utm-schedules : Updates an existing utmSchedule.
     *
     * @param utmSchedule the utmSchedule to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmSchedule,
     * or with status 400 (Bad Request) if the utmSchedule is not valid,
     * or with status 500 (Internal Server Error) if the utmSchedule couldn't be updated
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PutMapping("/utm-schedules")
    public ResponseEntity<UtmSchedule> updateUtmSchedule(@Valid @RequestBody UtmSchedule utmSchedule) throws URISyntaxException {
        final String ctx = CLASSNAME + ".updateUtmSchedule";
        try {
            if (utmSchedule.getId() == null)
                throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "idnull");

            UtmSchedule result = utmScheduleService.save(utmSchedule);
            return ResponseEntity.ok()
                .headers(HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, utmSchedule.getId().toString()))
                .body(result);
        } catch (BadRequestAlertException e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(null);
        }
    }

    /**
     * GET  /utm-schedules : get all the utmSchedules.
     *
     * @param pageable the pagination information
     * @return the ResponseEntity with status 200 (OK) and the list of utmSchedules in body
     */
    @GetMapping("/utm-schedules")
    public ResponseEntity<List<UtmSchedule>> getAllUtmSchedules(Pageable pageable) {
        final String ctx = CLASSNAME + ".getAllUtmSchedules";
        try {
            Page<UtmSchedule> page = utmScheduleService.findAll(pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-schedules");
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(null);
        }
    }

    /**
     * GET  /utm-schedules/:id : get the "id" utmSchedule.
     *
     * @param id the id of the utmSchedule to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmSchedule, or with status 404 (Not Found)
     */
    @GetMapping("/utm-schedules/{id}")
    public ResponseEntity<UtmSchedule> getUtmSchedule(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".getUtmSchedule";
        try {
            Optional<UtmSchedule> utmSchedule = utmScheduleService.findOne(id);
            return ResponseUtil.wrapOrNotFound(utmSchedule);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(null);
        }
    }

    /**
     * DELETE  /utm-schedules/:id : delete the "id" utmSchedule.
     *
     * @param id the id of the utmSchedule to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/utm-schedules/{id}")
    public ResponseEntity<Void> deleteUtmSchedule(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".deleteUtmSchedule";
        try {
            utmScheduleService.delete(id);
            return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, id.toString())).build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(null);
        }
    }
}

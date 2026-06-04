package com.park.utmstack.web.rest.incident_response;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.incident_response.UtmIncidentAction;
import com.park.utmstack.security.SecurityUtils;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.incident_response.UtmIncidentActionCriteria;
import com.park.utmstack.service.incident_response.UtmIncidentActionQueryService;
import com.park.utmstack.service.incident_response.UtmIncidentActionService;
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

import java.time.Instant;
import java.util.Arrays;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

/**
 * REST controller for managing UtmIncidentAction.
 */
@RestController
@RequestMapping("/api")
public class UtmIncidentActionResource {

    private final Logger log = LoggerFactory.getLogger(UtmIncidentActionResource.class);

    private static final String ENTITY_NAME = "utmIncidentAction";
    private static final String CLASS_NAME = "UtmIncidentActionResource";

    private final UtmIncidentActionService utmIncidentActionService;
    private final UtmIncidentActionQueryService utmIncidentActionQueryService;
    private final ApplicationEventService applicationEventService;

    public UtmIncidentActionResource(UtmIncidentActionService utmIncidentActionService,
                                     UtmIncidentActionQueryService utmIncidentActionQueryService,
                                     ApplicationEventService applicationEventService) {
        this.utmIncidentActionService = utmIncidentActionService;
        this.utmIncidentActionQueryService = utmIncidentActionQueryService;
        this.applicationEventService = applicationEventService;
    }

    /**
     * POST  /utm-incident-actions : Create a new utmIncidentAction.
     *
     * @param utmIncidentAction the utmIncidentAction to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmIncidentAction, or with status 400 (Bad
     * Request) if the utmIncidentAction has already an ID
     */
    @PostMapping("/utm-incident-actions")
    public ResponseEntity<UtmIncidentAction> createUtmIncidentAction(@RequestBody UtmIncidentAction utmIncidentAction) {
        final String ctx = CLASS_NAME + ".createUtmIncidentAction";
        try {
            if (utmIncidentAction.getId() != null)
                throw new Exception("A new utmIncidentAction cannot already have an ID");

            utmIncidentAction.setCreatedDate(Instant.now());
            utmIncidentAction.setCreatedUser(SecurityUtils.getCurrentUserLogin().orElse("system"));

            UtmIncidentAction result = utmIncidentActionService.save(utmIncidentAction);
            return ResponseEntity.ok(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * PUT  /utm-incident-actions : Updates an existing utmIncidentAction.
     *
     * @param utmIncidentAction the utmIncidentAction to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmIncidentAction, or with status 400 (Bad
     * Request) if the utmIncidentAction is not valid, or with status 500 (Internal Server Error) if the utmIncidentAction
     * couldn't be updated
     */
    @PutMapping("/utm-incident-actions")
    public ResponseEntity<UtmIncidentAction> updateUtmIncidentAction(@RequestBody UtmIncidentAction utmIncidentAction) {
        final String ctx = CLASS_NAME + ".updateUtmIncidentAction";

        try {
            if (utmIncidentAction.getId() == null)
                throw new Exception("Id can't be null");

            utmIncidentAction.setModifiedDate(Instant.now());
            utmIncidentAction.setModifiedUser(SecurityUtils.getCurrentUserLogin().orElse("system"));

            UtmIncidentAction result = utmIncidentActionService.save(utmIncidentAction);
            return ResponseEntity.ok(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * GET  /utm-incident-actions : get all the utmIncidentActions.
     *
     * @param pageable the pagination information
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the list of utmIncidentActions in body
     */
    @GetMapping("/utm-incident-actions")
    public ResponseEntity<List<UtmIncidentAction>> getAllUtmIncidentActions(UtmIncidentActionCriteria criteria,
                                                                            Pageable pageable) {
        final String ctx = CLASS_NAME + ".getAllUtmIncidentActions";
        try {
            Page<UtmIncidentAction> page = utmIncidentActionQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-incident-actions");
            List<UtmIncidentAction> content = page.getContent();

            if (!CollectionUtils.isEmpty(content))
                content = content.stream().filter(a -> Arrays.asList("SHUTDOWN_SERVER", "RUN_CMD").contains(a.getActionCommand()))
                    .collect(Collectors.toList());

            return ResponseEntity.ok().headers(headers).body(content);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * GET  /utm-incident-actions/count : count all the utmIncidentActions.
     *
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the count in body
     */
    @GetMapping("/utm-incident-actions/count")
    public ResponseEntity<Long> countUtmIncidentActions(UtmIncidentActionCriteria criteria) {
        final String ctx = CLASS_NAME + ".countUtmIncidentActions";
        try {
            return ResponseEntity.ok().body(utmIncidentActionQueryService.countByCriteria(criteria));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * GET  /utm-incident-actions/:id : get the "id" utmIncidentAction.
     *
     * @param id the id of the utmIncidentAction to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmIncidentAction, or with status 404 (Not Found)
     */
    @GetMapping("/utm-incident-actions/{id}")
    public ResponseEntity<UtmIncidentAction> getUtmIncidentAction(@PathVariable Long id) {
        final String ctx = CLASS_NAME + ".getUtmIncidentAction";
        try {
            Optional<UtmIncidentAction> utmIncidentAction = utmIncidentActionService.findOne(id);
            return ResponseUtil.wrapOrNotFound(utmIncidentAction);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * DELETE  /utm-incident-actions/:id : delete the "id" utmIncidentAction.
     *
     * @param id the id of the utmIncidentAction to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/utm-incident-actions/{id}")
    public ResponseEntity<Void> deleteUtmIncidentAction(@PathVariable Long id) {
        final String ctx = CLASS_NAME + ".deleteUtmIncidentAction";
        try {
            utmIncidentActionService.delete(id);
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

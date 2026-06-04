package com.park.utmstack.web.rest;

import com.park.utmstack.domain.UtmAlertTag;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.UtmAlertTagQueryService;
import com.park.utmstack.service.UtmAlertTagService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.UtmAlertTagCriteria;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.dao.DataIntegrityViolationException;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import tech.jhipster.web.util.ResponseUtil;

import javax.validation.Valid;
import java.net.URI;
import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing UtmAlertTag.
 */
@RestController
@RequestMapping("/api")
public class UtmAlertTagResource {

    private final Logger log = LoggerFactory.getLogger(UtmAlertTagResource.class);

    private static final String ENTITY_NAME = "utmAlertTag";
    private static final String CLASS_NAME = "UtmAlertTagResource";

    private final UtmAlertTagService alertTagService;
    private final UtmAlertTagQueryService alertTagQueryService;
    private final ApplicationEventService applicationEventService;

    public UtmAlertTagResource(UtmAlertTagService alertTagService,
                               UtmAlertTagQueryService alertTagQueryService,
                               ApplicationEventService applicationEventService) {
        this.alertTagService = alertTagService;
        this.alertTagQueryService = alertTagQueryService;
        this.applicationEventService = applicationEventService;
    }

    /**
     * POST  /utm-alert-tags : Create a new utmAlertTag.
     *
     * @param alertTag the utmAlertTag to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmAlertTag, or with status 400 (Bad
     * Request) if the utmAlertTag has already an ID
     */
    @PostMapping("/utm-alert-tags")
    public ResponseEntity<UtmAlertTag> createUtmAlertTag(@Valid @RequestBody UtmAlertTag alertTag) {
        final String ctx = CLASS_NAME + ".createUtmAlertTag";
        UtmAlertTag result = null;
        try {
            if (alertTag.getId() != null)
                throw new BadRequestAlertException("A new utmAlertTag cannot already have an ID", ENTITY_NAME, "idexists");

            alertTag.setSystemOwner(false);
            result = alertTagService.save(alertTag);
            return ResponseEntity.created(new URI("/api/utm-alert-tags/" + result.getId())).headers(
                HeaderUtil.createEntityCreationAlert(ENTITY_NAME, result.getId().toString())).body(result);
        } catch (DataIntegrityViolationException e) {
            String msg = ctx + ": " + e.getMostSpecificCause().getMessage().replaceAll("\n", "");
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.CONFLICT).headers(HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg))
                .body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(result);
        }
    }

    /**
     * PUT  /utm-alert-tags : Updates an existing utmAlertTag.
     *
     * @param utmAlertTag the utmAlertTag to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmAlertTag, or with status 400 (Bad
     * Request) if the utmAlertTag is not valid, or with status 500 (Internal Server Error) if the utmAlertTag couldn't be
     * updated
     */
    @PutMapping("/utm-alert-tags")
    public ResponseEntity<UtmAlertTag> updateUtmAlertTag(@Valid @RequestBody UtmAlertTag utmAlertTag) {
        final String ctx = CLASS_NAME + ".updateUtmAlertTag";
        try {
            if (utmAlertTag.getId() == null)
                throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "idnull");

            UtmAlertTag result = alertTagService.save(utmAlertTag);
            return ResponseEntity.ok().headers(HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, utmAlertTag.getId().toString()))
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
     * GET  /utm-alert-tags : get all the utmAlertCategories.
     *
     * @param pageable the pagination information
     * @return the ResponseEntity with status 200 (OK) and the list of utmAlertCategories in body
     */
    @GetMapping("/utm-alert-tags")
    public ResponseEntity<List<UtmAlertTag>> getAllUtmAlertCategories(UtmAlertTagCriteria criteria, Pageable pageable) {
        final String ctx = CLASS_NAME + ".getAllUtmAlertCategories";
        try {
            Page<UtmAlertTag> page = alertTagQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-alert-tags");
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
     * GET  /utm-alert-tags/:id : get the "id" utmAlertTag.
     *
     * @param id the id of the utmAlertTag to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmAlertTag, or with status 404 (Not Found)
     */
    @GetMapping("/utm-alert-tags/{id}")
    public ResponseEntity<UtmAlertTag> getUtmAlertTag(@PathVariable Long id) {
        final String ctx = CLASS_NAME + ".getUtmAlertTag";
        try {
            Optional<UtmAlertTag> utmAlertTag = alertTagService.findOne(id);
            return ResponseUtil.wrapOrNotFound(utmAlertTag);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(null);
        }
    }

    /**
     * DELETE  /utm-alert-tags/:id : delete the "id" utmAlertTag.
     *
     * @param id the id of the utmAlertTag to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/utm-alert-tags/{id}")
    public ResponseEntity<Void> deleteUtmAlertTag(@PathVariable Long id) {
        final String ctx = CLASS_NAME + ".deleteUtmAlertTag";
        try {
            alertTagService.delete(id);
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

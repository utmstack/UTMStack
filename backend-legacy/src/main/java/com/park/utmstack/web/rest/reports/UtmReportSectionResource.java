package com.park.utmstack.web.rest.reports;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.reports.UtmReportSection;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.reports.UtmReportSectionCriteria;
import com.park.utmstack.service.reports.UtmReportSectionQueryService;
import com.park.utmstack.service.reports.UtmReportSectionService;
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
import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing UtmReportSection.
 */
@RestController
@RequestMapping("/api")
public class UtmReportSectionResource {

    private final Logger log = LoggerFactory.getLogger(UtmReportSectionResource.class);

    private static final String ENTITY_NAME = "utmReportSection";
    private static final String CLASSNAME = "UtmReportSectionResource";

    private final UtmReportSectionService utmReportSectionService;
    private final UtmReportSectionQueryService utmReportSectionQueryService;
    private final ApplicationEventService applicationEventService;

    public UtmReportSectionResource(UtmReportSectionService utmReportSectionService, UtmReportSectionQueryService utmReportSectionQueryService, ApplicationEventService applicationEventService) {
        this.utmReportSectionService = utmReportSectionService;
        this.utmReportSectionQueryService = utmReportSectionQueryService;
        this.applicationEventService = applicationEventService;
    }

    /**
     * POST  /utm-report-sections : Create a new utmReportSection.
     *
     * @param utmReportSection the utmReportSection to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmReportSection, or with status 400 (Bad Request) if the utmReportSection has already an ID
     */
    @PostMapping("/utm-report-sections")
    public ResponseEntity<UtmReportSection> createUtmReportSection(@Valid @RequestBody UtmReportSection utmReportSection) {
        final String ctx = CLASSNAME + ".createUtmReportSection";
        try {
            if (utmReportSection.getId() != null)
                throw new BadRequestAlertException("A new utmReportSection cannot already have an ID", ENTITY_NAME, "idexists");

            utmReportSection.setRepSecSystem(false);
            UtmReportSection result = utmReportSectionService.save(utmReportSection);
            return ResponseEntity.created(new URI("/api/utm-report-sections/" + result.getId()))
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
     * PUT  /utm-report-sections : Updates an existing utmReportSection.
     *
     * @param utmReportSection the utmReportSection to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmReportSection,
     * or with status 400 (Bad Request) if the utmReportSection is not valid,
     * or with status 500 (Internal Server Error) if the utmReportSection couldn't be updated
     */
    @PutMapping("/utm-report-sections")
    public ResponseEntity<UtmReportSection> updateUtmReportSection(@Valid @RequestBody UtmReportSection utmReportSection) {
        final String ctx = CLASSNAME + ".updateUtmReportSection";
        try {
            if (utmReportSection.getId() == null)
                throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "idnull");

            utmReportSection.setRepSecSystem(false);
            UtmReportSection result = utmReportSectionService.save(utmReportSection);
            return ResponseEntity.ok()
                .headers(HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, utmReportSection.getId().toString()))
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
     * GET  /utm-report-sections : get all the utmReportSections.
     *
     * @param pageable the pagination information
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the list of utmReportSections in body
     */
    @GetMapping("/utm-report-sections")
    public ResponseEntity<List<UtmReportSection>> getAllUtmReportSections(UtmReportSectionCriteria criteria, Pageable pageable) {
        final String ctx = CLASSNAME + ".getAllUtmReportSections";
        try {
            Page<UtmReportSection> page = utmReportSectionQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-report-sections");
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
     * GET  /utm-report-sections/count : count all the utmReportSections.
     *
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the count in body
     */
    @GetMapping("/utm-report-sections/count")
    public ResponseEntity<Long> countUtmReportSections(UtmReportSectionCriteria criteria) {
        final String ctx = CLASSNAME + ".countUtmReportSections";
        try {
            return ResponseEntity.ok().body(utmReportSectionQueryService.countByCriteria(criteria));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(null);
        }
    }

    /**
     * GET  /utm-report-sections/:id : get the "id" utmReportSection.
     *
     * @param id the id of the utmReportSection to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmReportSection, or with status 404 (Not Found)
     */
    @GetMapping("/utm-report-sections/{id}")
    public ResponseEntity<UtmReportSection> getUtmReportSection(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".getUtmReportSection";
        try {
            Optional<UtmReportSection> utmReportSection = utmReportSectionService.findOne(id);
            return ResponseUtil.wrapOrNotFound(utmReportSection);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(null);
        }
    }

    /**
     * DELETE  /utm-report-sections/:id : delete the "id" utmReportSection.
     *
     * @param id the id of the utmReportSection to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/utm-report-sections/{id}")
    public ResponseEntity<Void> deleteUtmReportSection(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".deleteUtmReportSection";
        try {
            utmReportSectionService.delete(id);
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

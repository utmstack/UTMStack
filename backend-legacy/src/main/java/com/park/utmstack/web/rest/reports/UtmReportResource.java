package com.park.utmstack.web.rest.reports;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.reports.UtmReport;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.reports.UtmReportCriteria;
import com.park.utmstack.service.reports.UtmReportQueryService;
import com.park.utmstack.service.reports.UtmReportService;
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
 * REST controller for managing UtmReports.
 */
@RestController
@RequestMapping("/api")
public class UtmReportResource {

    private static final String CLASS_NAME = "UtmReportResource";
    private final Logger log = LoggerFactory.getLogger(UtmReportResource.class);

    private static final String ENTITY_NAME = "utmReports";

    private final UtmReportService utmReportsService;

    private final ApplicationEventService applicationEventService;
    private final UtmReportQueryService utmReportsQueryService;

    public UtmReportResource(UtmReportService utmReportsService,
                             ApplicationEventService applicationEventService,
                             UtmReportQueryService utmReportsQueryService) {
        this.utmReportsService = utmReportsService;
        this.applicationEventService = applicationEventService;
        this.utmReportsQueryService = utmReportsQueryService;
    }

    /**
     * POST  /utm-reports : Create a new utmReports.
     *
     * @param utmReports the utmReports to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmReports, or with status 400 (Bad Request) if the utmReports has already an ID
     */
    @PostMapping("/utm-reports")
    public ResponseEntity<UtmReport> createUtmReports(@Valid @RequestBody UtmReport utmReports) {
        final String ctx = CLASS_NAME + ".createUtmReports";
        try {
            if (utmReports.getId() != null)
                throw new BadRequestAlertException("A new utmReports cannot already have an ID", ENTITY_NAME, "idexists");

            UtmReport result = utmReportsService.save(utmReports);
            return ResponseEntity.created(new URI("/api/utm-reports/" + result.getId()))
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
     * PUT  /utm-reports : Updates an existing utmReports.
     *
     * @param utmReports the utmReports to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmReports,
     * or with status 400 (Bad Request) if the utmReports is not valid,
     * or with status 500 (Internal Server Error) if the utmReports couldn't be updated
     */
    @PutMapping("/utm-reports")
    public ResponseEntity<UtmReport> updateUtmReports(@Valid @RequestBody UtmReport utmReports) {
        final String ctx = CLASS_NAME + ".updateUtmReports";
        try {
            if (utmReports.getId() == null)
                throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "idnull");

            UtmReport result = utmReportsService.save(utmReports);
            return ResponseEntity.ok()
                .headers(HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, utmReports.getId().toString()))
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
     * GET  /utm-reports : get all the utmReports.
     *
     * @param pageable the pagination information
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the list of utmReports in body
     */
    @GetMapping("/utm-reports")
    public ResponseEntity<List<UtmReport>> getAllUtmReports(UtmReportCriteria criteria, Pageable pageable) {
        final String ctx = CLASS_NAME + ".getAllUtmReports";
        try {
            Page<UtmReport> page = utmReportsQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-reports");
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
     * GET  /utm-reports/count : count all the utmReports.
     *
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the count in body
     */
    @GetMapping("/utm-reports/count")
    public ResponseEntity<Long> countUtmReports(UtmReportCriteria criteria) {
        final String ctx = CLASS_NAME + ".countUtmReports";
        try {
            return ResponseEntity.ok().body(utmReportsQueryService.countByCriteria(criteria));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(null);
        }
    }

    /**
     * GET  /utm-reports/:id : get the "id" utmReports.
     *
     * @param id the id of the utmReports to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmReports, or with status 404 (Not Found)
     */
    @GetMapping("/utm-reports/{id}")
    public ResponseEntity<UtmReport> getUtmReports(@PathVariable Long id) {
        final String ctx = CLASS_NAME + ".getUtmReports";
        try {
            Optional<UtmReport> utmReports = utmReportsService.findOne(id);
            return ResponseUtil.wrapOrNotFound(utmReports);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(null);
        }
    }

    /**
     * DELETE  /utm-reports/:id : delete the "id" utmReports.
     *
     * @param id the id of the utmReports to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/utm-reports/{id}")
    public ResponseEntity<Void> deleteUtmReports(@PathVariable Long id) {
        final String ctx = CLASS_NAME + ".deleteUtmReports";
        try {
            utmReportsService.delete(id);
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

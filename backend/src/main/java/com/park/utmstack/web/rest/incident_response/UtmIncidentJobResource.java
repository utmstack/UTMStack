package com.park.utmstack.web.rest.incident_response;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.incident_response.UtmIncidentJob;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.incident_response.UtmIncidentJobCriteria;
import com.park.utmstack.service.incident_response.UtmIncidentJobQueryService;
import com.park.utmstack.service.incident_response.UtmIncidentJobService;
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

import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing UtmIncidentJob.
 */
@RestController
@RequestMapping("/api")
public class UtmIncidentJobResource {

    private final Logger log = LoggerFactory.getLogger(UtmIncidentJobResource.class);

    private static final String ENTITY_NAME = "utmIncidentJob";
    private static final String CLASS_NAME = "UtmIncidentJobResource";

    private final UtmIncidentJobService incidentJobService;
    private final UtmIncidentJobQueryService incidentJobQueryService;
    private final ApplicationEventService applicationEventService;

    public UtmIncidentJobResource(UtmIncidentJobService incidentJobService,
                                  UtmIncidentJobQueryService incidentJobQueryService,
                                  ApplicationEventService applicationEventService) {
        this.incidentJobService = incidentJobService;
        this.incidentJobQueryService = incidentJobQueryService;
        this.applicationEventService = applicationEventService;
    }

    /**
     * POST  /utm-incident-jobs : Create a new incidentJob.
     *
     * @param incidentJob the incidentJob to create
     * @return the ResponseEntity with status 201 (Created) and with body the new incidentJob, or with status 400 (Bad
     * Request) if the incidentJob has already an ID
     */
    @PostMapping("/utm-incident-jobs")
    public ResponseEntity<UtmIncidentJob> createIncidentJob(@RequestBody UtmIncidentJob incidentJob) {
        final String ctx = CLASS_NAME + ".createIncidentJob";
        try {
            if (incidentJob.getId() != null)
                throw new Exception("A new IncidentJob cannot already have an ID");
            return ResponseEntity.ok(incidentJobService.save(incidentJob));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * GET  /utm-incident-jobs : get all the utmIncidentJobs.
     *
     * @param pageable the pagination information
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the list of utmIncidentJobs in body
     */
    @GetMapping("/utm-incident-jobs")
    public ResponseEntity<List<UtmIncidentJob>> getAllUtmIncidentJobs(UtmIncidentJobCriteria criteria, Pageable pageable) {
        final String ctx = CLASS_NAME + ".getAllUtmIncidentJobs";
        try {
            Page<UtmIncidentJob> page = incidentJobQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-incident-jobs");
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
     * GET  /utm-incident-jobs/count : count all the utmIncidentJobs.
     *
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the count in body
     */
    @GetMapping("/utm-incident-jobs/count")
    public ResponseEntity<Long> countUtmIncidentJobs(UtmIncidentJobCriteria criteria) {
        final String ctx = CLASS_NAME + ".countUtmIncidentJobs";
        try {
            return ResponseEntity.ok().body(incidentJobQueryService.countByCriteria(criteria));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * GET  /utm-incident-jobs/:id : get the "id" utmIncidentJob.
     *
     * @param id the id of the utmIncidentJob to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmIncidentJob, or with status 404 (Not Found)
     */
    @GetMapping("/utm-incident-jobs/{id}")
    public ResponseEntity<UtmIncidentJob> getUtmIncidentJob(@PathVariable Long id) {
        final String ctx = CLASS_NAME + ".getUtmIncidentJob";
        try {
            Optional<UtmIncidentJob> utmIncidentJob = incidentJobService.findOne(id);
            return ResponseUtil.wrapOrNotFound(utmIncidentJob);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * DELETE  /utm-incident-jobs/:id : delete the "id" utmIncidentJob.
     *
     * @param id the id of the utmIncidentJob to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/utm-incident-jobs/{id}")
    public ResponseEntity<Void> deleteUtmIncidentJob(@PathVariable Long id) {
        final String ctx = CLASS_NAME + ".getUtmIncidentJob";
        try {
            incidentJobService.delete(id);
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

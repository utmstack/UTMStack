package com.park.utmstack.web.rest.compliance;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.compliance.UtmComplianceReportSchedule;
import com.park.utmstack.repository.compliance.UtmComplianceReportScheduleRepository;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.compliance.UtmComplianceReportScheduleService;
import com.park.utmstack.service.dto.compliance.UtmComplianceReportScheduleCriteria;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;

import java.net.URISyntaxException;
import java.util.List;
import java.util.Optional;

import com.park.utmstack.web.rest.util.PaginationUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.util.StringUtils;
import org.springframework.web.bind.annotation.*;
import tech.jhipster.web.util.ResponseUtil;

import javax.validation.Valid;

/**
 * REST controller for managing {@link UtmComplianceReportSchedule}.
 */
@RestController
@RequestMapping("/api")
public class UtmComplianceReportScheduleResource {

    private static final String CLASSNAME = "UtmComplianceReportScheduleResource";
    private final Logger log = LoggerFactory.getLogger(UtmComplianceReportScheduleResource.class);

    private static final String ENTITY_NAME = "utmComplianceReportSchedule";
    private final ApplicationEventService applicationEventService;


    private final UtmComplianceReportScheduleService utmComplianceReportScheduleService;

    private final UtmComplianceReportScheduleRepository utmComplianceReportScheduleRepository;

    public UtmComplianceReportScheduleResource(
        ApplicationEventService applicationEventService,
        UtmComplianceReportScheduleService utmComplianceReportScheduleService,
        UtmComplianceReportScheduleRepository utmComplianceReportScheduleRepository
    ) {
        this.applicationEventService = applicationEventService;
        this.utmComplianceReportScheduleService = utmComplianceReportScheduleService;
        this.utmComplianceReportScheduleRepository = utmComplianceReportScheduleRepository;
    }

    /**
     * {@code POST  /compliance-report-schedules} : Create a new utmComplianceReportSchedule.
     *
     * @param utmComplianceReportSchedule the utmComplianceReportSchedule to create.
     * @return the {@link ResponseEntity} with status {@code 201 (Created)} and with body the new utmComplianceReportSchedule, or with status {@code 400 (Bad Request)} if the utmComplianceReportSchedule has already an ID.
     * @throws URISyntaxException if the Location URI syntax is incorrect.
     */
    @PostMapping("/compliance-report-schedules")
    public ResponseEntity<UtmComplianceReportSchedule> createUtmComplianceReportSchedule(
        @Valid @RequestBody UtmComplianceReportSchedule utmComplianceReportSchedule
    ) throws URISyntaxException {
        final String ctx = CLASSNAME + ".createUtmComplianceReportSchedule";
        try {
            log.debug("REST request to save UtmComplianceReportSchedule : {}", utmComplianceReportSchedule);
            if (utmComplianceReportSchedule.getId() != null) {
                throw new BadRequestAlertException("A new utmComplianceReportSchedule cannot already have an ID", ENTITY_NAME, "idexists");
            }
            Optional<UtmComplianceReportSchedule> toCheck = utmComplianceReportScheduleService.findByComplianceReportValues(utmComplianceReportSchedule);
            if (toCheck.isPresent() && utmComplianceReportSchedule.equals(toCheck.get())) {
                throw new BadRequestAlertException(ctx + ": " + "This entity is already inserted", "utmComplianceReportSchedule", "alreadyinserted");
            }
            UtmComplianceReportSchedule result = utmComplianceReportScheduleService.save(utmComplianceReportSchedule);
            return ResponseEntity.ok().body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * {@code PUT  /compliance-report-schedules} : Updates an existing utmComplianceReportSchedule.
     *
     * @param utmComplianceReportSchedule the utmComplianceReportSchedule to update.
     * @return the {@link ResponseEntity} with status {@code 200 (OK)} and with body the updated utmComplianceReportSchedule,
     * or with status {@code 400 (Bad Request)} if the utmComplianceReportSchedule is not valid,
     * or with status {@code 500 (Internal Server Error)} if the utmComplianceReportSchedule couldn't be updated.
     * @throws URISyntaxException if the Location URI syntax is incorrect.
     */
    @PutMapping("/compliance-report-schedules")
    public ResponseEntity<UtmComplianceReportSchedule> updateUtmComplianceReportSchedule(
        @Valid @RequestBody UtmComplianceReportSchedule utmComplianceReportSchedule
    ) throws URISyntaxException {
        final String ctx = CLASSNAME + ".updateUtmComplianceReportSchedule";
        try {
            log.debug("REST request to update UtmComplianceReportSchedule : {}", utmComplianceReportSchedule);
            if (utmComplianceReportSchedule.getId() == null) {
                throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "idnull");
            }

            if (!utmComplianceReportScheduleRepository.existsById(utmComplianceReportSchedule.getId())) {
                throw new BadRequestAlertException("Entity not found", ENTITY_NAME, "idnotfound");
            }
            if (!StringUtils.hasText(utmComplianceReportSchedule.getScheduleString())) {
                throw new BadRequestAlertException("Invalid or null schedule string", ENTITY_NAME, "cronnotfound");
            }

            UtmComplianceReportSchedule result = utmComplianceReportScheduleService.save(utmComplianceReportSchedule);
            return ResponseEntity.ok().body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * {@code GET  /compliance-report-schedules-by-user} : get all the utmComplianceReportSchedules.
     *
     * @return the {@link ResponseEntity} with status {@code 200 (OK)} and the list of utmComplianceReportSchedules in body.
     */
    @GetMapping("/compliance-report-schedules-by-user")
    public ResponseEntity<List<UtmComplianceReportSchedule>> getAllUtmComplianceReportSchedules(UtmComplianceReportScheduleCriteria criteria, Pageable pageable) {
        final String ctx = CLASSNAME + ".getAllUtmComplianceReportSchedules";
        try {
            log.debug("REST request to get all UtmComplianceReportSchedules");
            Page<UtmComplianceReportSchedule> page = utmComplianceReportScheduleService.findAllOfCurrentUser(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/compliance-report-schedules-by-user");
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * {@code GET  /compliance-report-schedules-by-id/id} : get a utmComplianceReportSchedule.
     *
     * @return the {@link ResponseEntity} with status {@code 200 (OK)} and the utmComplianceReportSchedule matching the id.
     * or with status {@code 404 (Not Found)} if the utmComplianceReportSchedule not exists
     */
    @GetMapping("/compliance-report-schedules-by-id/{id}")
    public ResponseEntity<UtmComplianceReportSchedule> getUtmComplianceReportScheduleById(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".getUtmComplianceReportScheduleById";
        try {
            log.debug("REST request to get a UtmComplianceReportSchedule by id");
            return ResponseUtil.wrapOrNotFound(utmComplianceReportScheduleService.findOne(id));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * {@code DELETE  /compliance-report-schedules/id} : delete the utmComplianceReportSchedule by id.
     *
     * @param id the id of the utmComplianceReportSchedule to delete.
     * @return the {@link ResponseEntity} with status {@code 204 (NO_CONTENT)}.
     */
    @DeleteMapping("/compliance-report-schedules/{id}")
    public ResponseEntity<Void> deleteUtmComplianceReportSchedule(@PathVariable Long id) {
        log.debug("REST request to delete UtmComplianceReportSchedule : {}", id);
        final String ctx = CLASSNAME + ".deleteUtmComplianceReportSchedule";
        try {
            utmComplianceReportScheduleService.delete(id);
            return ResponseEntity.ok().headers(com.park.utmstack.web.rest.util.HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, id.toString())).build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }
}

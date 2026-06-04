package com.park.utmstack.web.rest.chart_builder;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.chart_builder.UtmDashboard;
import com.park.utmstack.domain.chart_builder.UtmDashboardVisualization;
import com.park.utmstack.security.SecurityUtils;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.chart_builder.UtmDashboardQueryService;
import com.park.utmstack.service.chart_builder.UtmDashboardService;
import com.park.utmstack.service.dto.chart_builder.UtmDashboardCriteria;
import com.park.utmstack.util.exceptions.UtmEntityCreationException;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.core.env.Environment;
import org.springframework.dao.DataIntegrityViolationException;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import tech.jhipster.config.JHipsterConstants;
import tech.jhipster.web.util.ResponseUtil;

import javax.validation.Valid;
import javax.validation.constraints.NotNull;
import java.net.URI;
import java.time.Instant;
import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.util.Arrays;
import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing UtmDashboard.
 */
@RestController
@RequestMapping("/api")
public class UtmDashboardResource {

    private static final String CLASSNAME = "UtmDashboardResource";
    private final Logger log = LoggerFactory.getLogger(UtmDashboardResource.class);

    private static final String ENTITY_NAME = "utmDashboard";

    private final UtmDashboardService utmDashboardService;
    private final UtmDashboardQueryService dashboardQueryService;
    private final ApplicationEventService applicationEventService;
    private final Environment env;

    public UtmDashboardResource(UtmDashboardService utmDashboardService,
                                UtmDashboardQueryService dashboardQueryService,
                                ApplicationEventService applicationEventService, Environment env) {
        this.utmDashboardService = utmDashboardService;
        this.dashboardQueryService = dashboardQueryService;
        this.applicationEventService = applicationEventService;
        this.env = env;
    }

    /**
     * POST  /utm-dashboards : Create a new utmDashboard.
     *
     * @param utmDashboard the utmDashboard to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmDashboard, or with status 400 (Bad
     * Request) if the utmDashboard has already an ID
     */
    @PostMapping("/utm-dashboards")
    public ResponseEntity<UtmDashboard> createUtmDashboard(@Valid @RequestBody UtmDashboard utmDashboard) {
        final String ctx = CLASSNAME + ".createUtmDashboard";
        if (utmDashboard.getId() != null) {
            throw new BadRequestAlertException("A new utmDashboard cannot already have an ID", ENTITY_NAME, "idexists");
        }

        UtmDashboard result = null;
        try {
            utmDashboard.setUserCreated(
                SecurityUtils.getCurrentUserLogin().orElseThrow(() -> new UtmEntityCreationException("Missing user login")));
            utmDashboard.setCreatedDate(LocalDateTime.now().toInstant(ZoneOffset.UTC));

            // Set the next system sequence value only if the environment is dev
            if (Arrays.asList(env.getActiveProfiles()).contains(JHipsterConstants.SPRING_PROFILE_DEVELOPMENT)) {
                utmDashboard.setId(utmDashboardService.getSystemSequenceNextValue());
                utmDashboard.setSystemOwner(true);
            }

            result = utmDashboardService.save(utmDashboard);
            return ResponseEntity.created(new URI("/api/utm-dashboards/" + result.getId())).headers(
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
     * PUT  /utm-dashboards : Updates an existing utmDashboard.
     *
     * @param utmDashboard the utmDashboard to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmDashboard, or with status 400 (Bad
     * Request) if the utmDashboard is not valid, or with status 500 (Internal Server Error) if the utmDashboard couldn't be
     * updated
     */
    @PutMapping("/utm-dashboards")
    public ResponseEntity<UtmDashboard> updateUtmDashboard(@Valid @RequestBody UtmDashboard utmDashboard) {
        final String ctx = CLASSNAME + ".updateUtmDashboard";
        if (utmDashboard.getId() == null) {
            throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "idnull");
        }

        UtmDashboard result = null;
        try {
            utmDashboard.setUserModified(
                SecurityUtils.getCurrentUserLogin().orElseThrow(() -> new UtmEntityCreationException("Missing user login")));
            utmDashboard.setModifiedDate(Instant.now());
            utmDashboard.setSystemOwner(utmDashboard.getSystemOwner() == null ? utmDashboard.getId() < 1000000 : utmDashboard.getSystemOwner());

            result = utmDashboardService.save(utmDashboard);
            return ResponseEntity.ok().headers(
                HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, utmDashboard.getId().toString())).body(result);
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
     * GET  /utm-dashboards : get all the utmDashboards.
     *
     * @param pageable the pagination information
     * @return the ResponseEntity with status 200 (OK) and the list of utmDashboards in body
     */
    @GetMapping("/utm-dashboards")
    public ResponseEntity<List<UtmDashboard>> getAllUtmDashboards(UtmDashboardCriteria criteria, Pageable pageable) {
        final String ctx = CLASSNAME + ".getAllUtmDashboards";
        try {
            Page<UtmDashboard> page = dashboardQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-dashboards");
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    /**
     * GET  /utm-dashboards/:id : get the "id" utmDashboard.
     *
     * @param id the id of the utmDashboard to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmDashboard, or with status 404 (Not Found)
     */
    @GetMapping("/utm-dashboards/{id}")
    public ResponseEntity<UtmDashboard> getUtmDashboard(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".getUtmDashboard";
        try {
            Optional<UtmDashboard> utmDashboard = utmDashboardService.findOne(id);
            return ResponseUtil.wrapOrNotFound(utmDashboard);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    /**
     * DELETE  /utm-dashboards/:id : delete the "id" utmDashboard.
     *
     * @param id the id of the utmDashboard to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/utm-dashboards/{id}")
    public ResponseEntity<Void> deleteUtmDashboard(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".deleteUtmDashboard";
        try {
            utmDashboardService.delete(id);
            return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, id.toString())).build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    @PostMapping("/utm-dashboards/import")
    public ResponseEntity<Void> importDashboards(@Valid @RequestBody ImportDashboardsBody body) {
        final String ctx = CLASSNAME + ".importDashboards";

        try {
            utmDashboardService.importDashboards(body.getDashboards(), body.getOverride());
            return ResponseEntity.ok().build();
        } catch (Exception e) {
            String msg;
            if (e instanceof DataIntegrityViolationException) {
                msg = ((DataIntegrityViolationException) e).getMostSpecificCause().getMessage().replaceAll("\n", "");
                log.error(msg);
                applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
                return ResponseEntity.status(HttpStatus.CONFLICT).headers(HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg))
                    .build();
            }
            msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).build();
        }
    }

    public static class ImportDashboardsBody {
        @NotNull
        private List<UtmDashboardVisualization> dashboards;

        @NotNull
        private Boolean override;

        public List<UtmDashboardVisualization> getDashboards() {
            return dashboards;
        }

        public void setDashboards(List<UtmDashboardVisualization> dashboards) {
            this.dashboards = dashboards;
        }

        public Boolean getOverride() {
            return override;
        }

        public void setOverride(Boolean override) {
            this.override = override;
        }
    }
}

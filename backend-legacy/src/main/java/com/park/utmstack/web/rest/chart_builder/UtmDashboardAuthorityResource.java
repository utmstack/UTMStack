package com.park.utmstack.web.rest.chart_builder;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.chart_builder.UtmDashboardAuthority;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.chart_builder.UtmDashboardAuthorityService;
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
 * REST controller for managing UtmDashboardAuthority.
 */
@RestController
@RequestMapping("/api")
public class UtmDashboardAuthorityResource {

    private final Logger log = LoggerFactory.getLogger(UtmDashboardAuthorityResource.class);

    private static final String ENTITY_NAME = "utmDashboardAuthority";
    private static final String CLASSNAME = "UtmDashboardAuthorityResource";

    private final UtmDashboardAuthorityService utmDashboardAuthorityService;
    private final ApplicationEventService applicationEventService;

    public UtmDashboardAuthorityResource(UtmDashboardAuthorityService utmDashboardAuthorityService,
                                         ApplicationEventService applicationEventService) {
        this.utmDashboardAuthorityService = utmDashboardAuthorityService;
        this.applicationEventService = applicationEventService;
    }

    /**
     * POST  /utm-dashboard-authorities : Create a new utmDashboardAuthority.
     *
     * @param utmDashboardAuthority the utmDashboardAuthority to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmDashboardAuthority, or with status 400 (Bad Request) if the utmDashboardAuthority has already an ID
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PostMapping("/utm-dashboard-authorities")
    public ResponseEntity<UtmDashboardAuthority> createUtmDashboardAuthority(@Valid @RequestBody UtmDashboardAuthority utmDashboardAuthority) throws URISyntaxException {
        final String ctx = CLASSNAME + ".createUtmDashboardAuthority";
        try {
            if (utmDashboardAuthority.getId() != null)
                throw new BadRequestAlertException("A new utmDashboardAuthority cannot already have an ID", ENTITY_NAME, "idexists");

            UtmDashboardAuthority result = utmDashboardAuthorityService.save(utmDashboardAuthority);
            return ResponseEntity.created(new URI("/api/utm-dashboard-authorities/" + result.getId()))
                .headers(HeaderUtil.createEntityCreationAlert(ENTITY_NAME, result.getId().toString()))
                .body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    /**
     * PUT  /utm-dashboard-authorities : Updates an existing utmDashboardAuthority.
     *
     * @param utmDashboardAuthority the utmDashboardAuthority to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmDashboardAuthority,
     * or with status 400 (Bad Request) if the utmDashboardAuthority is not valid,
     * or with status 500 (Internal Server Error) if the utmDashboardAuthority couldn't be updated
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PutMapping("/utm-dashboard-authorities")
    public ResponseEntity<UtmDashboardAuthority> updateUtmDashboardAuthority(@Valid @RequestBody UtmDashboardAuthority utmDashboardAuthority) throws URISyntaxException {
        final String ctx = CLASSNAME + ".updateUtmDashboardAuthority";
        try {
            if (utmDashboardAuthority.getId() == null)
                throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "idnull");

            UtmDashboardAuthority result = utmDashboardAuthorityService.save(utmDashboardAuthority);
            return ResponseEntity.ok()
                .headers(HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, utmDashboardAuthority.getId().toString()))
                .body(result);
        } catch (BadRequestAlertException e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    /**
     * GET  /utm-dashboard-authorities : get all the utmDashboardAuthorities.
     *
     * @param pageable the pagination information
     * @return the ResponseEntity with status 200 (OK) and the list of utmDashboardAuthorities in body
     */
    @GetMapping("/utm-dashboard-authorities")
    public ResponseEntity<List<UtmDashboardAuthority>> getAllUtmDashboardAuthorities(Pageable pageable) {
        final String ctx = CLASSNAME + ".getAllUtmDashboardAuthorities";
        try {
            Page<UtmDashboardAuthority> page = utmDashboardAuthorityService.findAll(pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-dashboard-authorities");
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
     * GET  /utm-dashboard-authorities/:id : get the "id" utmDashboardAuthority.
     *
     * @param id the id of the utmDashboardAuthority to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmDashboardAuthority, or with status 404 (Not Found)
     */
    @GetMapping("/utm-dashboard-authorities/{id}")
    public ResponseEntity<UtmDashboardAuthority> getUtmDashboardAuthority(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".getUtmDashboardAuthority";
        try {
            Optional<UtmDashboardAuthority> utmDashboardAuthority = utmDashboardAuthorityService.findOne(id);
            return ResponseUtil.wrapOrNotFound(utmDashboardAuthority);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    /**
     * DELETE  /utm-dashboard-authorities/:id : delete the "id" utmDashboardAuthority.
     *
     * @param id the id of the utmDashboardAuthority to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/utm-dashboard-authorities/{id}")
    public ResponseEntity<Void> deleteUtmDashboardAuthority(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".deleteUtmDashboardAuthority";
        try {
            utmDashboardAuthorityService.delete(id);
            return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, id.toString())).build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }
}

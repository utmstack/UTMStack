package com.park.utmstack.web.rest;

import com.park.utmstack.domain.UtmIntegrationConf;
import com.park.utmstack.domain.UtmServerModule;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.UtmIntegrationConfQueryService;
import com.park.utmstack.service.UtmIntegrationConfService;
import com.park.utmstack.service.UtmServerModuleService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.UtmIntegrationConfCriteria;
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

import java.net.URI;
import java.net.URISyntaxException;
import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing UtmIntegrationConf.
 */
@RestController
@RequestMapping("/api")
public class UtmIntegrationConfResource {
    private static final String CLASSNAME = "UtmIntegrationConfResource";

    private final Logger log = LoggerFactory.getLogger(UtmIntegrationConfResource.class);

    private static final String ENTITY_NAME = "utmIntegrationConf";

    private final UtmIntegrationConfService utmIntegrationConfService;
    private final UtmIntegrationConfQueryService utmIntegrationConfQueryService;
    private final UtmServerModuleService serverModuleService;
    private final ApplicationEventService eventService;

    public UtmIntegrationConfResource(UtmIntegrationConfService utmIntegrationConfService,
                                      UtmIntegrationConfQueryService utmIntegrationConfQueryService,
                                      UtmServerModuleService serverModuleService,
                                      ApplicationEventService eventService) {
        this.utmIntegrationConfService = utmIntegrationConfService;
        this.utmIntegrationConfQueryService = utmIntegrationConfQueryService;
        this.serverModuleService = serverModuleService;
        this.eventService = eventService;
    }

    /**
     * POST  /utm-integration-confs : Create a new utmIntegrationConf.
     *
     * @param utmIntegrationConf the utmIntegrationConf to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmIntegrationConf, or with status 400 (Bad Request) if the utmIntegrationConf has already an ID
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PostMapping("/utm-integration-confs")
    public ResponseEntity<UtmIntegrationConf> createUtmIntegrationConf(@RequestBody UtmIntegrationConf utmIntegrationConf) throws URISyntaxException {
        final String ctx = CLASSNAME + ".createUtmIntegrationConf";
        try {
            if (utmIntegrationConf.getId() != null) {
                String msg = ctx + ": A new integration configuration param cannot already have an ID";
                log.error(msg);
                eventService.createEvent(msg, ApplicationEventType.ERROR);
                return ResponseEntity.status(HttpStatus.BAD_REQUEST).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
            }
            UtmIntegrationConf result = utmIntegrationConfService.save(utmIntegrationConf);
            return ResponseEntity.created(new URI("/api/utm-integration-confs/" + result.getId()))
                .headers(HeaderUtil.createEntityCreationAlert(ENTITY_NAME, result.getId().toString()))
                .body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * PUT  /utm-integration-confs : Updates an existing utmIntegrationConf.
     *
     * @param utmIntegrationConf the utmIntegrationConf to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmIntegrationConf,
     * or with status 400 (Bad Request) if the utmIntegrationConf is not valid,
     * or with status 500 (Internal Server Error) if the utmIntegrationConf couldn't be updated
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PutMapping("/utm-integration-confs")
    public ResponseEntity<UtmIntegrationConf> updateUtmIntegrationConf(@RequestBody UtmIntegrationConf utmIntegrationConf) throws URISyntaxException {
        final String ctx = CLASSNAME + ".updateUtmIntegrationConf";
        try {
            if (utmIntegrationConf.getId() == null) {
                String msg = ctx + ": Integration ID can't be null";
                log.error(msg);
                eventService.createEvent(msg, ApplicationEventType.ERROR);
                return ResponseEntity.status(HttpStatus.BAD_REQUEST).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
            }

            UtmIntegrationConf result = utmIntegrationConfService.save(utmIntegrationConf);

            UtmServerModule module = result.getIntegration().getModule();
            module.setNeedsRestart(true);
            serverModuleService.save(module);

            return ResponseEntity.ok()
                .headers(HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, utmIntegrationConf.getId().toString()))
                .body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * GET  /utm-integration-confs : get all the utmIntegrationConfs.
     *
     * @param pageable the pagination information
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the list of utmIntegrationConfs in body
     */
    @GetMapping("/utm-integration-confs")
    public ResponseEntity<List<UtmIntegrationConf>> getAllUtmIntegrationConfs(UtmIntegrationConfCriteria criteria, Pageable pageable) {
        log.debug("REST request to get UtmIntegrationConfs by criteria: {}", criteria);
        Page<UtmIntegrationConf> page = utmIntegrationConfQueryService.findByCriteria(criteria, pageable);
        HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-integration-confs");
        return ResponseEntity.ok().headers(headers).body(page.getContent());
    }

    /**
     * GET  /utm-integration-confs/count : count all the utmIntegrationConfs.
     *
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the count in body
     */
    @GetMapping("/utm-integration-confs/count")
    public ResponseEntity<Long> countUtmIntegrationConfs(UtmIntegrationConfCriteria criteria) {
        log.debug("REST request to count UtmIntegrationConfs by criteria: {}", criteria);
        return ResponseEntity.ok().body(utmIntegrationConfQueryService.countByCriteria(criteria));
    }

    /**
     * GET  /utm-integration-confs/:id : get the "id" utmIntegrationConf.
     *
     * @param id the id of the utmIntegrationConf to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmIntegrationConf, or with status 404 (Not Found)
     */
    @GetMapping("/utm-integration-confs/{id}")
    public ResponseEntity<UtmIntegrationConf> getUtmIntegrationConf(@PathVariable Long id) {
        log.debug("REST request to get UtmIntegrationConf : {}", id);
        Optional<UtmIntegrationConf> utmIntegrationConf = utmIntegrationConfService.findOne(id);
        return ResponseUtil.wrapOrNotFound(utmIntegrationConf);
    }

    /**
     * DELETE  /utm-integration-confs/:id : delete the "id" utmIntegrationConf.
     *
     * @param id the id of the utmIntegrationConf to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/utm-integration-confs/{id}")
    public ResponseEntity<Void> deleteUtmIntegrationConf(@PathVariable Long id) {
        log.debug("REST request to delete UtmIntegrationConf : {}", id);
        utmIntegrationConfService.delete(id);
        return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, id.toString())).build();
    }
}

package com.park.utmstack.web.rest;

import com.park.utmstack.domain.UtmServerModule;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.UtmServerModuleQueryService;
import com.park.utmstack.service.UtmServerModuleService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.UtmServerModuleCriteria;
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

import java.net.URI;
import java.net.URISyntaxException;
import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing UtmServerModule.
 */
@RestController
@RequestMapping("/api")
public class UtmServerModuleResource {

    private final Logger log = LoggerFactory.getLogger(UtmServerModuleResource.class);

    private static final String ENTITY_NAME = "utmServerModule";
    private static final String CLASSNAME = "UtmServerModuleResource";

    private final UtmServerModuleService utmServerModuleService;
    private final UtmServerModuleQueryService utmServerModuleQueryService;
    private final ApplicationEventService eventService;

    public UtmServerModuleResource(UtmServerModuleService utmServerModuleService,
                                   UtmServerModuleQueryService utmServerModuleQueryService,
                                   ApplicationEventService eventService) {
        this.utmServerModuleService = utmServerModuleService;
        this.utmServerModuleQueryService = utmServerModuleQueryService;
        this.eventService = eventService;
    }

    /**
     * POST  /utm-server-modules : Create a new utmServerModule.
     *
     * @param utmServerModule the utmServerModule to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmServerModule, or with status 400 (Bad Request) if the utmServerModule has already an ID
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PostMapping("/utm-server-modules")
    public ResponseEntity<UtmServerModule> createUtmServerModule(@RequestBody UtmServerModule utmServerModule) throws URISyntaxException {
        log.debug("REST request to save UtmServerModule : {}", utmServerModule);
        if (utmServerModule.getId() != null) {
            throw new BadRequestAlertException("A new utmServerModule cannot already have an ID", ENTITY_NAME, "idexists");
        }
        UtmServerModule result = utmServerModuleService.save(utmServerModule);
        return ResponseEntity.created(new URI("/api/utm-server-modules/" + result.getId()))
            .headers(HeaderUtil.createEntityCreationAlert(ENTITY_NAME, result.getId().toString()))
            .body(result);
    }

    /**
     * PUT  /utm-server-modules : Updates an existing utmServerModule.
     *
     * @param utmServerModule the utmServerModule to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmServerModule,
     * or with status 400 (Bad Request) if the utmServerModule is not valid,
     * or with status 500 (Internal Server Error) if the utmServerModule couldn't be updated
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PutMapping("/utm-server-modules")
    public ResponseEntity<UtmServerModule> updateUtmServerModule(@RequestBody UtmServerModule utmServerModule) throws URISyntaxException {
        log.debug("REST request to update UtmServerModule : {}", utmServerModule);
        if (utmServerModule.getId() == null) {
            throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "idnull");
        }
        UtmServerModule result = utmServerModuleService.save(utmServerModule);
        return ResponseEntity.ok()
            .headers(HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, utmServerModule.getId().toString()))
            .body(result);
    }

    /**
     * GET  /utm-server-modules : get all the utmServerModules.
     *
     * @param pageable the pagination information
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the list of utmServerModules in body
     */
    @GetMapping("/utm-server-modules")
    public ResponseEntity<List<UtmServerModule>> getAllUtmServerModules(UtmServerModuleCriteria criteria, Pageable pageable) {
        log.debug("REST request to get UtmServerModules by criteria: {}", criteria);
        Page<UtmServerModule> page = utmServerModuleQueryService.findByCriteria(criteria, pageable);
        HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-server-modules");
        return ResponseEntity.ok().headers(headers).body(page.getContent());
    }

    /**
     * GET  /utm-server-modules/count : count all the utmServerModules.
     *
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the count in body
     */
    @GetMapping("/utm-server-modules/count")
    public ResponseEntity<Long> countUtmServerModules(UtmServerModuleCriteria criteria) {
        log.debug("REST request to count UtmServerModules by criteria: {}", criteria);
        return ResponseEntity.ok().body(utmServerModuleQueryService.countByCriteria(criteria));
    }

    /**
     * GET  /utm-server-modules/:id : get the "id" utmServerModule.
     *
     * @param id the id of the utmServerModule to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmServerModule, or with status 404 (Not Found)
     */
    @GetMapping("/utm-server-modules/{id}")
    public ResponseEntity<UtmServerModule> getUtmServerModule(@PathVariable Long id) {
        log.debug("REST request to get UtmServerModule : {}", id);
        Optional<UtmServerModule> utmServerModule = utmServerModuleService.findOne(id);
        return ResponseUtil.wrapOrNotFound(utmServerModule);
    }

    /**
     * DELETE  /utm-server-modules/:id : delete the "id" utmServerModule.
     *
     * @param id the id of the utmServerModule to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/utm-server-modules/{id}")
    public ResponseEntity<Void> deleteUtmServerModule(@PathVariable Long id) {
        log.debug("REST request to delete UtmServerModule : {}", id);
        utmServerModuleService.delete(id);
        return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, id.toString())).build();
    }

    @GetMapping("/utm-server-modules/modules-with-integrations")
    public ResponseEntity<List<UtmServerModule>> getModulesWithIntegrations(@RequestParam(required = false) Long serverId, @RequestParam(required = false) String prettyName) {
        final String ctx = CLASSNAME + ".getModulesWithIntegrations";
        try {
            return ResponseEntity.ok(utmServerModuleService.getModulesWithIntegrations(serverId, prettyName));
        } catch (Exception e) {
            final String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(null);
        }
    }
}

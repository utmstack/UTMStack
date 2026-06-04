package com.park.utmstack.web.rest;

import com.park.utmstack.domain.UtmIntegration;
import com.park.utmstack.service.UtmIntegrationQueryService;
import com.park.utmstack.service.UtmIntegrationService;
import com.park.utmstack.service.dto.UtmIntegrationCriteria;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import tech.jhipster.web.util.ResponseUtil;

import java.net.URI;
import java.net.URISyntaxException;
import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing UtmIntegration.
 */
@RestController
@RequestMapping("/api")
public class UtmIntegrationResource {

    private final Logger log = LoggerFactory.getLogger(UtmIntegrationResource.class);

    private static final String ENTITY_NAME = "utmIntegration";

    private final UtmIntegrationService utmIntegrationService;

    private final UtmIntegrationQueryService utmIntegrationQueryService;

    public UtmIntegrationResource(UtmIntegrationService utmIntegrationService, UtmIntegrationQueryService utmIntegrationQueryService) {
        this.utmIntegrationService = utmIntegrationService;
        this.utmIntegrationQueryService = utmIntegrationQueryService;
    }

    /**
     * POST  /utm-integrations : Create a new utmIntegration.
     *
     * @param utmIntegration the utmIntegration to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmIntegration, or with status 400 (Bad Request) if the utmIntegration has already an ID
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PostMapping("/utm-integrations")
    public ResponseEntity<UtmIntegration> createUtmIntegration(@RequestBody UtmIntegration utmIntegration) throws URISyntaxException {
        log.debug("REST request to save UtmIntegration : {}", utmIntegration);
        if (utmIntegration.getId() != null) {
            throw new BadRequestAlertException("A new utmIntegration cannot already have an ID", ENTITY_NAME, "idexists");
        }
        UtmIntegration result = utmIntegrationService.save(utmIntegration);
        return ResponseEntity.created(new URI("/api/utm-integrations/" + result.getId()))
            .headers(HeaderUtil.createEntityCreationAlert(ENTITY_NAME, result.getId().toString()))
            .body(result);
    }

    /**
     * PUT  /utm-integrations : Updates an existing utmIntegration.
     *
     * @param utmIntegration the utmIntegration to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmIntegration,
     * or with status 400 (Bad Request) if the utmIntegration is not valid,
     * or with status 500 (Internal Server Error) if the utmIntegration couldn't be updated
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PutMapping("/utm-integrations")
    public ResponseEntity<UtmIntegration> updateUtmIntegration(@RequestBody UtmIntegration utmIntegration) throws URISyntaxException {
        log.debug("REST request to update UtmIntegration : {}", utmIntegration);
        if (utmIntegration.getId() == null) {
            throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "idnull");
        }
        UtmIntegration result = utmIntegrationService.save(utmIntegration);
        return ResponseEntity.ok()
            .headers(HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, utmIntegration.getId().toString()))
            .body(result);
    }

    /**
     * GET  /utm-integrations : get all the utmIntegrations.
     *
     * @param pageable the pagination information
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the list of utmIntegrations in body
     */
    @GetMapping("/utm-integrations")
    public ResponseEntity<List<UtmIntegration>> getAllUtmIntegrations(UtmIntegrationCriteria criteria, Pageable pageable) {
        log.debug("REST request to get UtmIntegrations by criteria: {}", criteria);
        Page<UtmIntegration> page = utmIntegrationQueryService.findByCriteria(criteria, pageable);
        HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-integrations");
        return ResponseEntity.ok().headers(headers).body(page.getContent());
    }

    /**
    * GET  /utm-integrations/count : count all the utmIntegrations.
    *
    * @param criteria the criterias which the requested entities should match
    * @return the ResponseEntity with status 200 (OK) and the count in body
    */
    @GetMapping("/utm-integrations/count")
    public ResponseEntity<Long> countUtmIntegrations(UtmIntegrationCriteria criteria) {
        log.debug("REST request to count UtmIntegrations by criteria: {}", criteria);
        return ResponseEntity.ok().body(utmIntegrationQueryService.countByCriteria(criteria));
    }

    /**
     * GET  /utm-integrations/:id : get the "id" utmIntegration.
     *
     * @param id the id of the utmIntegration to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmIntegration, or with status 404 (Not Found)
     */
    @GetMapping("/utm-integrations/{id}")
    public ResponseEntity<UtmIntegration> getUtmIntegration(@PathVariable Long id) {
        log.debug("REST request to get UtmIntegration : {}", id);
        Optional<UtmIntegration> utmIntegration = utmIntegrationService.findOne(id);
        return ResponseUtil.wrapOrNotFound(utmIntegration);
    }

    /**
     * DELETE  /utm-integrations/:id : delete the "id" utmIntegration.
     *
     * @param id the id of the utmIntegration to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/utm-integrations/{id}")
    public ResponseEntity<Void> deleteUtmIntegration(@PathVariable Long id) {
        log.debug("REST request to delete UtmIntegration : {}", id);
        utmIntegrationService.delete(id);
        return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, id.toString())).build();
    }
}

package com.park.utmstack.web.rest;

import com.park.utmstack.domain.UtmServer;
import com.park.utmstack.service.UtmServerQueryService;
import com.park.utmstack.service.UtmServerService;
import com.park.utmstack.service.dto.UtmServerCriteria;
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
 * REST controller for managing UtmServer.
 */
@RestController
@RequestMapping("/api")
public class UtmServerResource {

    private final Logger log = LoggerFactory.getLogger(UtmServerResource.class);

    private static final String ENTITY_NAME = "utmServer";

    private final UtmServerService utmServerService;

    private final UtmServerQueryService utmServerQueryService;

    public UtmServerResource(UtmServerService utmServerService, UtmServerQueryService utmServerQueryService) {
        this.utmServerService = utmServerService;
        this.utmServerQueryService = utmServerQueryService;
    }

    /**
     * POST  /utm-servers : Create a new utmServer.
     *
     * @param utmServer the utmServer to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmServer, or with status 400 (Bad Request) if the utmServer has already an ID
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PostMapping("/utm-servers")
    public ResponseEntity<UtmServer> createUtmServer(@RequestBody UtmServer utmServer) throws URISyntaxException {
        log.debug("REST request to save UtmServer : {}", utmServer);
        if (utmServer.getId() != null) {
            throw new BadRequestAlertException("A new utmServer cannot already have an ID", ENTITY_NAME, "idexists");
        }
        UtmServer result = utmServerService.save(utmServer);
        return ResponseEntity.created(new URI("/api/utm-servers/" + result.getId()))
            .headers(HeaderUtil.createEntityCreationAlert(ENTITY_NAME, result.getId().toString()))
            .body(result);
    }

    /**
     * PUT  /utm-servers : Updates an existing utmServer.
     *
     * @param utmServer the utmServer to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmServer,
     * or with status 400 (Bad Request) if the utmServer is not valid,
     * or with status 500 (Internal Server Error) if the utmServer couldn't be updated
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PutMapping("/utm-servers")
    public ResponseEntity<UtmServer> updateUtmServer(@RequestBody UtmServer utmServer) throws URISyntaxException {
        log.debug("REST request to update UtmServer : {}", utmServer);
        if (utmServer.getId() == null) {
            throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "idnull");
        }
        UtmServer result = utmServerService.save(utmServer);
        return ResponseEntity.ok()
            .headers(HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, utmServer.getId().toString()))
            .body(result);
    }

    /**
     * GET  /utm-servers : get all the utmServers.
     *
     * @param pageable the pagination information
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the list of utmServers in body
     */
    @GetMapping("/utm-servers")
    public ResponseEntity<List<UtmServer>> getAllUtmServers(UtmServerCriteria criteria, Pageable pageable) {
        log.debug("REST request to get UtmServers by criteria: {}", criteria);
        Page<UtmServer> page = utmServerQueryService.findByCriteria(criteria, pageable);
        HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-servers");
        return ResponseEntity.ok().headers(headers).body(page.getContent());
    }

    /**
    * GET  /utm-servers/count : count all the utmServers.
    *
    * @param criteria the criterias which the requested entities should match
    * @return the ResponseEntity with status 200 (OK) and the count in body
    */
    @GetMapping("/utm-servers/count")
    public ResponseEntity<Long> countUtmServers(UtmServerCriteria criteria) {
        log.debug("REST request to count UtmServers by criteria: {}", criteria);
        return ResponseEntity.ok().body(utmServerQueryService.countByCriteria(criteria));
    }

    /**
     * GET  /utm-servers/:id : get the "id" utmServer.
     *
     * @param id the id of the utmServer to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmServer, or with status 404 (Not Found)
     */
    @GetMapping("/utm-servers/{id}")
    public ResponseEntity<UtmServer> getUtmServer(@PathVariable Long id) {
        log.debug("REST request to get UtmServer : {}", id);
        Optional<UtmServer> utmServer = utmServerService.findOne(id);
        return ResponseUtil.wrapOrNotFound(utmServer);
    }

    /**
     * DELETE  /utm-servers/:id : delete the "id" utmServer.
     *
     * @param id the id of the utmServer to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/utm-servers/{id}")
    public ResponseEntity<Void> deleteUtmServer(@PathVariable Long id) {
        log.debug("REST request to delete UtmServer : {}", id);
        utmServerService.delete(id);
        return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, id.toString())).build();
    }
}

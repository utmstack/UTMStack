package com.park.utmstack.web.rest.network_scan;

import com.park.utmstack.domain.network_scan.UtmPorts;
import com.park.utmstack.service.dto.network_scan.UtmPortsCriteria;
import com.park.utmstack.service.network_scan.UtmPortsQueryService;
import com.park.utmstack.service.network_scan.UtmPortsService;
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
 * REST controller for managing UtmOpenPort.
 */
@RestController
@RequestMapping("/api")
public class UtmPortsResource {

    private final Logger log = LoggerFactory.getLogger(UtmPortsResource.class);

    private static final String ENTITY_NAME = "utmOpenPort";

    private final UtmPortsService utmOpenPortService;

    private final UtmPortsQueryService utmOpenPortQueryService;

    public UtmPortsResource(UtmPortsService utmOpenPortService, UtmPortsQueryService utmOpenPortQueryService) {
        this.utmOpenPortService = utmOpenPortService;
        this.utmOpenPortQueryService = utmOpenPortQueryService;
    }

    /**
     * POST  /utm-open-ports : Create a new utmOpenPort.
     *
     * @param utmPort the utmOpenPortDTO to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmOpenPortDTO, or with status 400 (Bad Request) if the utmOpenPort has already an ID
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PostMapping("/utm-open-ports")
    public ResponseEntity<UtmPorts> createUtmOpenPort(@RequestBody UtmPorts utmPort) throws URISyntaxException {
        log.debug("REST request to save UtmPorts : {}", utmPort);
        if (utmPort.getId() != null) {
            throw new BadRequestAlertException("A new utmOpenPort cannot already have an ID", ENTITY_NAME, "idexists");
        }
        UtmPorts result = utmOpenPortService.save(utmPort);
        return ResponseEntity.created(new URI("/api/utm-open-ports/" + result.getId()))
            .headers(HeaderUtil.createEntityCreationAlert(ENTITY_NAME, result.getId().toString()))
            .body(result);
    }

    /**
     * PUT  /utm-open-ports : Updates an existing utmOpenPort.
     *
     * @param utmPort the utmOpenPortDTO to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmOpenPortDTO,
     * or with status 400 (Bad Request) if the utmOpenPortDTO is not valid,
     * or with status 500 (Internal Server Error) if the utmOpenPortDTO couldn't be updated
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PutMapping("/utm-open-ports")
    public ResponseEntity<UtmPorts> updateUtmOpenPort(@RequestBody UtmPorts utmPort) throws URISyntaxException {
        log.debug("REST request to update UtmOpenPort : {}", utmPort);
        if (utmPort.getId() == null) {
            throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "idnull");
        }
        UtmPorts result = utmOpenPortService.save(utmPort);
        return ResponseEntity.ok()
            .headers(HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, utmPort.getId().toString()))
            .body(result);
    }

    /**
     * GET  /utm-open-ports : get all the utmOpenPorts.
     *
     * @param pageable the pagination information
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the list of utmOpenPorts in body
     */
    @GetMapping("/utm-open-ports")
    public ResponseEntity<List<UtmPorts>> getAllUtmOpenPorts(UtmPortsCriteria criteria, Pageable pageable) {
        log.debug("REST request to get UtmOpenPorts by criteria: {}", criteria);
        Page<UtmPorts> page = utmOpenPortQueryService.findByCriteria(criteria, pageable);
        HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-open-ports");
        return ResponseEntity.ok().headers(headers).body(page.getContent());
    }

    /**
     * GET  /utm-open-ports/count : count all the utmOpenPorts.
     *
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the count in body
     */
    @GetMapping("/utm-open-ports/count")
    public ResponseEntity<Long> countUtmOpenPorts(UtmPortsCriteria criteria) {
        log.debug("REST request to count UtmOpenPorts by criteria: {}", criteria);
        return ResponseEntity.ok().body(utmOpenPortQueryService.countByCriteria(criteria));
    }

    /**
     * GET  /utm-open-ports/:id : get the "id" utmOpenPort.
     *
     * @param id the id of the utmOpenPortDTO to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmOpenPortDTO, or with status 404 (Not Found)
     */
    @GetMapping("/utm-open-ports/{id}")
    public ResponseEntity<UtmPorts> getUtmOpenPort(@PathVariable Long id) {
        log.debug("REST request to get UtmOpenPort : {}", id);
        Optional<UtmPorts> utmPort = utmOpenPortService.findOne(id);
        return ResponseUtil.wrapOrNotFound(utmPort);
    }

    /**
     * DELETE  /utm-open-ports/:id : delete the "id" utmOpenPort.
     *
     * @param id the id of the utmOpenPortDTO to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/utm-open-ports/{id}")
    public ResponseEntity<Void> deleteUtmOpenPort(@PathVariable Long id) {
        log.debug("REST request to delete UtmOpenPort : {}", id);
        utmOpenPortService.delete(id);
        return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, id.toString())).build();
    }
}

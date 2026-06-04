package com.park.utmstack.web.rest;

import com.park.utmstack.domain.UtmAssetMetrics;
import com.park.utmstack.service.UtmAssetMetricsService;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.park.utmstack.web.rest.util.HeaderUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import tech.jhipster.web.util.ResponseUtil;

import javax.validation.Valid;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing UtmAssetMetrics.
 */
@RestController
@RequestMapping("/api")
public class UtmAssetMetricsResource {

    private final Logger log = LoggerFactory.getLogger(UtmAssetMetricsResource.class);

    private static final String ENTITY_NAME = "utmAssetMetrics";

    private final UtmAssetMetricsService utmAssetMetricsService;

    public UtmAssetMetricsResource(UtmAssetMetricsService utmAssetMetricsService) {
        this.utmAssetMetricsService = utmAssetMetricsService;
    }

    /**
     * POST  /utm-asset-metrics : Create a new utmAssetMetrics.
     *
     * @param utmAssetMetrics the utmAssetMetrics to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmAssetMetrics, or with status 400 (Bad Request) if the utmAssetMetrics has already an ID
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PostMapping("/utm-asset-metrics")
    public ResponseEntity<UtmAssetMetrics> createUtmAssetMetrics(@Valid @RequestBody UtmAssetMetrics utmAssetMetrics) throws URISyntaxException {
        log.debug("REST request to save UtmAssetMetrics : {}", utmAssetMetrics);
        if (utmAssetMetrics.getId() != null) {
            throw new BadRequestAlertException("A new utmAssetMetrics cannot already have an ID", ENTITY_NAME, "idexists");
        }
        UtmAssetMetrics result = utmAssetMetricsService.save(utmAssetMetrics);
        return ResponseEntity.created(new URI("/api/utm-asset-metrics/" + result.getId()))
            .headers(HeaderUtil.createEntityCreationAlert(ENTITY_NAME, result.getId().toString()))
            .body(result);
    }

    /**
     * PUT  /utm-asset-metrics : Updates an existing utmAssetMetrics.
     *
     * @param utmAssetMetrics the utmAssetMetrics to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmAssetMetrics,
     * or with status 400 (Bad Request) if the utmAssetMetrics is not valid,
     * or with status 500 (Internal Server Error) if the utmAssetMetrics couldn't be updated
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PutMapping("/utm-asset-metrics")
    public ResponseEntity<UtmAssetMetrics> updateUtmAssetMetrics(@Valid @RequestBody UtmAssetMetrics utmAssetMetrics) throws URISyntaxException {
        log.debug("REST request to update UtmAssetMetrics : {}", utmAssetMetrics);
        if (utmAssetMetrics.getId() == null) {
            throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "idnull");
        }
        UtmAssetMetrics result = utmAssetMetricsService.save(utmAssetMetrics);
        return ResponseEntity.ok()
            .headers(HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, utmAssetMetrics.getId().toString()))
            .body(result);
    }

    /**
     * GET  /utm-asset-metrics : get all the utmAssetMetrics.
     *
     * @return the ResponseEntity with status 200 (OK) and the list of utmAssetMetrics in body
     */
    @GetMapping("/utm-asset-metrics")
    public List<UtmAssetMetrics> getAllUtmAssetMetrics() {
        log.debug("REST request to get all UtmAssetMetrics");
        return utmAssetMetricsService.findAll();
    }

    /**
     * GET  /utm-asset-metrics/:id : get the "id" utmAssetMetrics.
     *
     * @param id the id of the utmAssetMetrics to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmAssetMetrics, or with status 404 (Not Found)
     */
    @GetMapping("/utm-asset-metrics/{id}")
    public ResponseEntity<UtmAssetMetrics> getUtmAssetMetrics(@PathVariable String id) {
        log.debug("REST request to get UtmAssetMetrics : {}", id);
        Optional<UtmAssetMetrics> utmAssetMetrics = utmAssetMetricsService.findOne(id);
        return ResponseUtil.wrapOrNotFound(utmAssetMetrics);
    }

    /**
     * DELETE  /utm-asset-metrics/:id : delete the "id" utmAssetMetrics.
     *
     * @param id the id of the utmAssetMetrics to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/utm-asset-metrics/{id}")
    public ResponseEntity<Void> deleteUtmAssetMetrics(@PathVariable String id) {
        log.debug("REST request to delete UtmAssetMetrics : {}", id);
        utmAssetMetricsService.delete(id);
        return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, id.toString())).build();
    }
}

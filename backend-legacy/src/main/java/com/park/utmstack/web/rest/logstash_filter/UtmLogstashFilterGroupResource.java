package com.park.utmstack.web.rest.logstash_filter;

import com.park.utmstack.domain.logstash_filter.UtmLogstashFilterGroup;
import com.park.utmstack.service.dto.logstash_filter.UtmLogstashFilterGroupCriteria;
import com.park.utmstack.service.logstash_filter.UtmLogstashFilterGroupQueryService;
import com.park.utmstack.service.logstash_filter.UtmLogstashFilterGroupService;
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

import javax.validation.Valid;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing UtmLogstashFilterGroup.
 */
@RestController
@RequestMapping("/api")
public class UtmLogstashFilterGroupResource {

    private final Logger log = LoggerFactory.getLogger(UtmLogstashFilterGroupResource.class);

    private static final String ENTITY_NAME = "utmLogstashFilterGroup";

    private final UtmLogstashFilterGroupService utmLogstashFilterGroupService;

    private final UtmLogstashFilterGroupQueryService utmLogstashFilterGroupQueryService;

    public UtmLogstashFilterGroupResource(UtmLogstashFilterGroupService utmLogstashFilterGroupService, UtmLogstashFilterGroupQueryService utmLogstashFilterGroupQueryService) {
        this.utmLogstashFilterGroupService = utmLogstashFilterGroupService;
        this.utmLogstashFilterGroupQueryService = utmLogstashFilterGroupQueryService;
    }

    /**
     * POST  /utm-logstash-filter-groups : Create a new utmLogstashFilterGroup.
     *
     * @param utmLogstashFilterGroup the utmLogstashFilterGroup to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmLogstashFilterGroup, or with status 400 (Bad Request) if the utmLogstashFilterGroup has already an ID
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PostMapping("/utm-logstash-filter-groups")
    public ResponseEntity<UtmLogstashFilterGroup> createUtmLogstashFilterGroup(@Valid @RequestBody UtmLogstashFilterGroup utmLogstashFilterGroup) throws URISyntaxException {
        log.debug("REST request to save UtmLogstashFilterGroup : {}", utmLogstashFilterGroup);
        if (utmLogstashFilterGroup.getId() != null) {
            throw new BadRequestAlertException("A new utmLogstashFilterGroup cannot already have an ID", ENTITY_NAME, "idexists");
        }
        UtmLogstashFilterGroup result = utmLogstashFilterGroupService.save(utmLogstashFilterGroup);
        return ResponseEntity.created(new URI("/api/utm-logstash-filter-groups/" + result.getId()))
            .headers(HeaderUtil.createEntityCreationAlert(ENTITY_NAME, result.getId().toString()))
            .body(result);
    }

    /**
     * PUT  /utm-logstash-filter-groups : Updates an existing utmLogstashFilterGroup.
     *
     * @param utmLogstashFilterGroup the utmLogstashFilterGroup to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmLogstashFilterGroup,
     * or with status 400 (Bad Request) if the utmLogstashFilterGroup is not valid,
     * or with status 500 (Internal Server Error) if the utmLogstashFilterGroup couldn't be updated
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PutMapping("/utm-logstash-filter-groups")
    public ResponseEntity<UtmLogstashFilterGroup> updateUtmLogstashFilterGroup(@Valid @RequestBody UtmLogstashFilterGroup utmLogstashFilterGroup) throws URISyntaxException {
        log.debug("REST request to update UtmLogstashFilterGroup : {}", utmLogstashFilterGroup);
        if (utmLogstashFilterGroup.getId() == null) {
            throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "idnull");
        }
        UtmLogstashFilterGroup result = utmLogstashFilterGroupService.save(utmLogstashFilterGroup);
        return ResponseEntity.ok()
            .headers(HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, utmLogstashFilterGroup.getId().toString()))
            .body(result);
    }

    /**
     * GET  /utm-logstash-filter-groups : get all the utmLogstashFilterGroups.
     *
     * @param pageable the pagination information
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the list of utmLogstashFilterGroups in body
     */
    @GetMapping("/utm-logstash-filter-groups")
    public ResponseEntity<List<UtmLogstashFilterGroup>> getAllUtmLogstashFilterGroups(UtmLogstashFilterGroupCriteria criteria, Pageable pageable) {
        log.debug("REST request to get UtmLogstashFilterGroups by criteria: {}", criteria);
        Page<UtmLogstashFilterGroup> page = utmLogstashFilterGroupQueryService.findByCriteria(criteria, pageable);
        HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-logstash-filter-groups");
        return ResponseEntity.ok().headers(headers).body(page.getContent());
    }

    /**
    * GET  /utm-logstash-filter-groups/count : count all the utmLogstashFilterGroups.
    *
    * @param criteria the criterias which the requested entities should match
    * @return the ResponseEntity with status 200 (OK) and the count in body
    */
    @GetMapping("/utm-logstash-filter-groups/count")
    public ResponseEntity<Long> countUtmLogstashFilterGroups(UtmLogstashFilterGroupCriteria criteria) {
        log.debug("REST request to count UtmLogstashFilterGroups by criteria: {}", criteria);
        return ResponseEntity.ok().body(utmLogstashFilterGroupQueryService.countByCriteria(criteria));
    }

    /**
     * GET  /utm-logstash-filter-groups/:id : get the "id" utmLogstashFilterGroup.
     *
     * @param id the id of the utmLogstashFilterGroup to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmLogstashFilterGroup, or with status 404 (Not Found)
     */
    @GetMapping("/utm-logstash-filter-groups/{id}")
    public ResponseEntity<UtmLogstashFilterGroup> getUtmLogstashFilterGroup(@PathVariable Long id) {
        log.debug("REST request to get UtmLogstashFilterGroup : {}", id);
        Optional<UtmLogstashFilterGroup> utmLogstashFilterGroup = utmLogstashFilterGroupService.findOne(id);
        return ResponseUtil.wrapOrNotFound(utmLogstashFilterGroup);
    }

    /**
     * DELETE  /utm-logstash-filter-groups/:id : delete the "id" utmLogstashFilterGroup.
     *
     * @param id the id of the utmLogstashFilterGroup to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/utm-logstash-filter-groups/{id}")
    public ResponseEntity<Void> deleteUtmLogstashFilterGroup(@PathVariable Long id) {
        log.debug("REST request to delete UtmLogstashFilterGroup : {}", id);
        utmLogstashFilterGroupService.delete(id);
        return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, id.toString())).build();
    }
}

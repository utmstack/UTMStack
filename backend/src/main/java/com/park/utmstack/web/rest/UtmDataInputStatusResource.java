package com.park.utmstack.web.rest;

import com.park.utmstack.domain.UtmDataInputStatus;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.UtmDataInputStatusQueryService;
import com.park.utmstack.service.UtmDataInputStatusService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.UtmDataInputStatusCriteria;
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
 * REST controller for managing UtmDataInputStatus.
 */
@RestController
@RequestMapping("/api")
public class UtmDataInputStatusResource {

    private static final String CLASSNAME = "UtmDataInputStatusResource";
    private final Logger log = LoggerFactory.getLogger(UtmDataInputStatusResource.class);

    private static final String ENTITY_NAME = "utmDataInputStatus";

    private final UtmDataInputStatusService dataInputStatusService;
    private final UtmDataInputStatusQueryService utmDataInputStatusQueryService;
    private final ApplicationEventService applicationEventService;

    public UtmDataInputStatusResource(UtmDataInputStatusService dataInputStatusService,
                                      UtmDataInputStatusQueryService utmDataInputStatusQueryService,
                                      ApplicationEventService applicationEventService) {
        this.dataInputStatusService = dataInputStatusService;
        this.utmDataInputStatusQueryService = utmDataInputStatusQueryService;
        this.applicationEventService = applicationEventService;
    }

    /**
     * POST  /utm-data-input-statuses : Create a new utmDataInputStatus.
     *
     * @param utmDataInputStatus the utmDataInputStatus to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmDataInputStatus, or with status 400 (Bad Request) if the utmDataInputStatus has already an ID
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PostMapping("/utm-data-input-statuses")
    public ResponseEntity<UtmDataInputStatus> createUtmDataInputStatus(@Valid @RequestBody UtmDataInputStatus utmDataInputStatus) throws URISyntaxException {
        log.debug("REST request to save UtmDataInputStatus : {}", utmDataInputStatus);
        if (utmDataInputStatus.getId() != null) {
            throw new BadRequestAlertException("A new utmDataInputStatus cannot already have an ID", ENTITY_NAME, "idexists");
        }
        UtmDataInputStatus result = dataInputStatusService.save(utmDataInputStatus);
        return ResponseEntity.created(new URI("/api/utm-data-input-statuses/" + result.getId()))
            .headers(HeaderUtil.createEntityCreationAlert(ENTITY_NAME, result.getId()))
            .body(result);
    }

    /**
     * PUT  /utm-data-input-statuses : Updates an existing utmDataInputStatus.
     *
     * @param utmDataInputStatus the utmDataInputStatus to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmDataInputStatus,
     * or with status 400 (Bad Request) if the utmDataInputStatus is not valid,
     * or with status 500 (Internal Server Error) if the utmDataInputStatus couldn't be updated
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PutMapping("/utm-data-input-statuses")
    public ResponseEntity<UtmDataInputStatus> updateUtmDataInputStatus(@Valid @RequestBody UtmDataInputStatus utmDataInputStatus) throws URISyntaxException {
        log.debug("REST request to update UtmDataInputStatus : {}", utmDataInputStatus);
        if (utmDataInputStatus.getId() == null) {
            throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "idnull");
        }
        UtmDataInputStatus result = dataInputStatusService.save(utmDataInputStatus);
        return ResponseEntity.ok()
            .headers(HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, utmDataInputStatus.getId()))
            .body(result);
    }

    /**
     * GET  /utm-data-input-statuses : get all the utmDataInputStatuses.
     *
     * @param pageable the pagination information
     * @return the ResponseEntity with status 200 (OK) and the list of utmDataInputStatuses in body
     */
    @GetMapping("/utm-data-input-statuses")
    public ResponseEntity<List<UtmDataInputStatus>> getAllUtmDataInputStatuses(Pageable pageable) {
        final String ctx = CLASSNAME + ".getAllUtmDataInputStatuses";
        try {
            Page<UtmDataInputStatus> page = dataInputStatusService.findImportantDatasource(pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-data-input-statuses");
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.ok().headers(HeaderUtil.createFailureAlert(null, null, msg)).body(null);

        }
    }

    /**
     * GET  /utm-data-input-statuses/count : count all the utmDataInputStatuses.
     *
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the count in body
     */
    @GetMapping("/utm-data-input-statuses/count")
    public ResponseEntity<Long> countUtmDataInputStatuses(UtmDataInputStatusCriteria criteria) {
        log.debug("REST request to count UtmDataInputStatuses by criteria: {}", criteria);
        return ResponseEntity.ok().body(utmDataInputStatusQueryService.countByCriteria(criteria));
    }

    /**
     * GET  /utm-data-input-statuses/:id : get the "id" utmDataInputStatus.
     *
     * @param id the id of the utmDataInputStatus to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmDataInputStatus, or with status 404 (Not Found)
     */
    @GetMapping("/utm-data-input-statuses/{id}")
    public ResponseEntity<UtmDataInputStatus> getUtmDataInputStatus(@PathVariable String id) {
        log.debug("REST request to get UtmDataInputStatus : {}", id);
        Optional<UtmDataInputStatus> utmDataInputStatus = dataInputStatusService.findOne(id);
        return ResponseUtil.wrapOrNotFound(utmDataInputStatus);
    }

    /**
     * DELETE  /utm-data-input-statuses/:id : delete the "id" utmDataInputStatus.
     *
     * @param id the id of the utmDataInputStatus to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/utm-data-input-statuses/{id}")
    public ResponseEntity<Void> deleteUtmDataInputStatus(@PathVariable String id) {
        log.debug("REST request to delete UtmDataInputStatus : {}", id);
        dataInputStatusService.delete(id);
        return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, id)).build();
    }
}

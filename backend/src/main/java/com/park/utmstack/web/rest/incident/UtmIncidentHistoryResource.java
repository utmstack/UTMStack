package com.park.utmstack.web.rest.incident;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.incident.UtmIncidentHistory;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.incident.UtmIncidentHistoryCriteria;
import com.park.utmstack.service.incident.UtmIncidentHistoryQueryService;
import com.park.utmstack.service.incident.UtmIncidentHistoryService;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import tech.jhipster.web.util.ResponseUtil;

import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing UtmIncidentHistory.
 */
@RestController
@RequestMapping("/api")
public class UtmIncidentHistoryResource {
    private final String CLASS_NAME = "UtmIncidentHistoryResource";
    private final Logger log = LoggerFactory.getLogger(UtmIncidentHistoryResource.class);

    private static final String ENTITY_NAME = "utmIncidentHistory";

    private final UtmIncidentHistoryService utmIncidentHistoryService;

    private final UtmIncidentHistoryQueryService utmIncidentHistoryQueryService;

    private final ApplicationEventService applicationEventService;

    public UtmIncidentHistoryResource(UtmIncidentHistoryService utmIncidentHistoryService,
                                      UtmIncidentHistoryQueryService utmIncidentHistoryQueryService,
                                      ApplicationEventService applicationEventService) {
        this.utmIncidentHistoryService = utmIncidentHistoryService;
        this.utmIncidentHistoryQueryService = utmIncidentHistoryQueryService;
        this.applicationEventService = applicationEventService;
    }


    /**
     * GET  /utm-incident-histories : get all the utmIncidentHistories.
     *
     * @param pageable the pagination information
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the list of utmIncidentHistories in body
     */
    @GetMapping("/utm-incident-histories")
    public ResponseEntity<List<UtmIncidentHistory>> getAllUtmIncidentHistories(UtmIncidentHistoryCriteria criteria, Pageable pageable) {
        final String ctx = ".getAllUtmIncidentHistories";
        try {
            log.debug("REST request to get UtmIncidentHistories by criteria: {}", criteria);
            Page<UtmIncidentHistory> page = utmIncidentHistoryQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-incident-histories");
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(HeaderUtil.createFailureAlert(CLASS_NAME, null, msg)).body(null);
        }

    }

    /**
     * GET  /utm-incident-histories/count : count all the utmIncidentHistories.
     *
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the count in body
     */
    @GetMapping("/utm-incident-histories/count")
    public ResponseEntity<Long> countUtmIncidentHistories(UtmIncidentHistoryCriteria criteria) {
        final String ctx = ".countUtmIncidentHistories";
        try {
            log.debug("REST request to count UtmIncidentHistories by criteria: {}", criteria);
            return ResponseEntity.ok().body(utmIncidentHistoryQueryService.countByCriteria(criteria));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(HeaderUtil.createFailureAlert(CLASS_NAME, null, msg)).body(null);
        }
    }

    /**
     * GET  /utm-incident-histories/:id : get the "id" utmIncidentHistory.
     *
     * @param id the id of the utmIncidentHistory to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmIncidentHistory, or with status 404 (Not Found)
     */
    @GetMapping("/utm-incident-histories/{id}")
    public ResponseEntity<UtmIncidentHistory> getUtmIncidentHistory(@PathVariable Long id) {
        final String ctx = ".getUtmIncidentHistory";
        try {
            log.debug("REST request to get UtmIncidentHistory : {}", id);
            Optional<UtmIncidentHistory> utmIncidentHistory = utmIncidentHistoryService.findOne(id);
            return ResponseUtil.wrapOrNotFound(utmIncidentHistory);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(HeaderUtil.createFailureAlert(CLASS_NAME, null, msg)).body(null);
        }
    }

//    /**
//     * POST  /utm-incident-histories : Create a new utmIncidentHistory.
//     *
//     * @param utmIncidentHistory the utmIncidentHistory to create
//     * @return the ResponseEntity with status 201 (Created) and with body the new utmIncidentHistory, or with status 400 (Bad Request) if the utmIncidentHistory has already an ID
//     * @throws URISyntaxException if the Location URI syntax is incorrect
//     */
//    @PostMapping("/utm-incident-histories")
//    public ResponseEntity<UtmIncidentHistory> createUtmIncidentHistory(@Valid @RequestBody UtmIncidentHistory utmIncidentHistory) throws URISyntaxException {
//        log.debug("REST request to save UtmIncidentHistory : {}", utmIncidentHistory);
//        final String ctx = ".createUtmIncidentHistory";
//        try {
//            if (utmIncidentHistory.getId() != null) {
//                throw new BadRequestAlertException("A new utmIncidentHistory cannot already have an ID", ENTITY_NAME, "idexists");
//            }
//            UtmIncidentHistory result = utmIncidentHistoryService.save(utmIncidentHistory);
//            return ResponseEntity.created(new URI("/api/utm-incident-histories/" + result.getId()))
//                .headers(HeaderUtil.createEntityCreationAlert(ENTITY_NAME, result.getId().toString()))
//                .body(result);
//        } catch (Exception e) {
//            String msg = ctx + ": " + e.getMessage();
//            log.error(msg);
//            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
//            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(HeaderUtil.createFailureAlert(CLASS_NAME, null, msg)).body(null);
//        }
//    }

}

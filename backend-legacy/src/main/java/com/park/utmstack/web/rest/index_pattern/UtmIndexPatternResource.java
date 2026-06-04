package com.park.utmstack.web.rest.index_pattern;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.index_pattern.UtmIndexPattern;
import com.park.utmstack.service.UtmStackService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.index_pattern.UtmIndexPatternCriteria;
import com.park.utmstack.service.dto.index_pattern.UtmIndexPatternField;
import com.park.utmstack.service.index_pattern.UtmIndexPatternQueryService;
import com.park.utmstack.service.index_pattern.UtmIndexPatternService;
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
import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing UtmIndexPattern.
 */
@RestController
@RequestMapping("/api")
public class UtmIndexPatternResource {

    private final Logger log = LoggerFactory.getLogger(UtmIndexPatternResource.class);

    private static final String ENTITY_NAME = "utmIndexPattern";
    private static final String CLASSNAME = "UtmIndexPatternResource";

    private final UtmIndexPatternService indexPatternService;
    private final UtmIndexPatternQueryService indexPatternQueryService;
    private final ApplicationEventService applicationEventService;
    private final UtmStackService utmStackService;

    public UtmIndexPatternResource(UtmIndexPatternService indexPatternService,
                                   UtmIndexPatternQueryService indexPatternQueryService,
                                   ApplicationEventService applicationEventService,
                                   UtmStackService utmStackService) {
        this.indexPatternService = indexPatternService;
        this.indexPatternQueryService = indexPatternQueryService;
        this.applicationEventService = applicationEventService;
        this.utmStackService = utmStackService;
    }

    /**
     * POST  /utm-index-patterns : Create a new utmIndexPattern.
     *
     * @param indexPattern the utmIndexPattern to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmIndexPattern, or with status 400 (Bad Request) if the utmIndexPattern has already an ID
     */
    @PostMapping("/utm-index-patterns")
    public ResponseEntity<UtmIndexPattern> createUtmIndexPattern(@Valid @RequestBody UtmIndexPattern indexPattern) {
        final String ctx = CLASSNAME + ".createUtmIndexPattern";

        try {
            if (indexPattern.getId() != null)
                throw new Exception("A new pattern can't have an id");
            boolean inDevelop = utmStackService.isInDevelop();

            indexPattern.setId(inDevelop ? indexPatternService.getSystemSequenceNextValue() : null);
            indexPattern.setPatternSystem(inDevelop);
            UtmIndexPattern result = indexPatternService.save(indexPattern);

            return ResponseEntity.ok(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    /**
     * PUT  /utm-index-patterns : Updates an existing utmIndexPattern.
     *
     * @param utmIndexPattern the utmIndexPattern to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmIndexPattern,
     * or with status 400 (Bad Request) if the utmIndexPattern is not valid,
     * or with status 500 (Internal Server Error) if the utmIndexPattern couldn't be updated
     */
    @PutMapping("/utm-index-patterns")
    public ResponseEntity<UtmIndexPattern> updateUtmIndexPattern(@Valid @RequestBody UtmIndexPattern utmIndexPattern) {
        final String ctx = CLASSNAME + ".updateUtmIndexPattern";
        try {
            if (utmIndexPattern.getId() == null)
                throw new Exception("Invalid id: " + utmIndexPattern.getId());

            UtmIndexPattern result = indexPatternService.save(utmIndexPattern);
            return ResponseEntity.ok(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    /**
     * GET  /utm-index-patterns : get all the utmIndexPatterns.
     *
     * @param pageable the pagination information
     * @return the ResponseEntity with status 200 (OK) and the list of utmIndexPatterns in body
     */
    @GetMapping("/utm-index-patterns")
    public ResponseEntity<List<UtmIndexPattern>> getAllUtmIndexPatterns(UtmIndexPatternCriteria criteria,
                                                                        Pageable pageable) {
        final String ctx = CLASSNAME + ".getAllUtmIndexPatterns";
        try {
            Page<UtmIndexPattern> page = indexPatternQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-index-patterns");
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    @GetMapping("/utm-index-patterns/fields")
    public ResponseEntity<List<UtmIndexPatternField>> getAllUtmIndexPatternsWithFields(UtmIndexPatternCriteria criteria,
                                                                                       Pageable pageable) {
        final String ctx = CLASSNAME + ".getAllUtmIndexPatternsWithFields";
        try {
            Page<UtmIndexPatternField> page = indexPatternQueryService.findWithFieldsByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-index-patterns-fields");
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
     * GET  /utm-index-patterns/:id : get the "id" utmIndexPattern.
     *
     * @param id the id of the utmIndexPattern to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmIndexPattern, or with status 404 (Not Found)
     */
    @GetMapping("/utm-index-patterns/{id}")
    public ResponseEntity<UtmIndexPattern> getUtmIndexPattern(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".getUtmIndexPattern";
        try {
            Optional<UtmIndexPattern> utmIndexPattern = indexPatternService.findOne(id);
            return ResponseUtil.wrapOrNotFound(utmIndexPattern);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    /**
     * DELETE  /utm-index-patterns/:id : delete the "id" utmIndexPattern.
     *
     * @param id the id of the utmIndexPattern to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/utm-index-patterns/{id}")
    public ResponseEntity<Void> deleteUtmIndexPattern(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".deleteUtmIndexPattern";
        try {
            indexPatternService.delete(id);
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

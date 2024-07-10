package com.park.utmstack.web.rest.correlation.config;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.correlation.config.UtmRegexPattern;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.correlation.config.UtmRegexPatternService;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import io.undertow.util.BadRequestException;
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
 * REST controller for managing {@link UtmRegexPatternResource}.
 */
@RestController
@RequestMapping("/api")
public class UtmRegexPatternResource {
    private static final String CLASSNAME = "UtmRegexPatternResource";
    private final Logger log = LoggerFactory.getLogger(UtmRegexPatternResource.class);

    private final ApplicationEventService applicationEventService;
    private final UtmRegexPatternService regexPatternService;

    public UtmRegexPatternResource(ApplicationEventService applicationEventService, UtmRegexPatternService regexPatternService) {
        this.applicationEventService = applicationEventService;
        this.regexPatternService = regexPatternService;
    }
    /**
     * {@code POST  /regex-pattern} : Add a new regex pattern.
     *
     * @param pattern the regex pattern to insert.
     * @return the {@link ResponseEntity} with status {@code 204 (No Content)}, with status {@code 400 (Bad Request)}, or with status {@code 500 (Internal)} if errors occurred.
     */
    @PostMapping("/regex-pattern")
    public ResponseEntity<Void> addRegexPattern(@Valid @RequestBody UtmRegexPattern pattern) {
        final String ctx = CLASSNAME + ".addRegexPattern";
        try {
            regexPatternService.addRegexPattern(pattern);
            return ResponseEntity.noContent().build();
        } catch (BadRequestException e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.BAD_REQUEST, msg);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * {@code PUT  /regex-pattern} : Update a regex pattern.
     *
     * @param pattern the regex pattern to update.
     * @return the {@link ResponseEntity} with status {@code 204 (No Content)}, with status {@code 400 (Bad Request)}, or with status {@code 500 (Internal)} if errors occurred.
     */
    @PutMapping("/regex-pattern")
    public ResponseEntity<Void> updateRegexPattern(@Valid @RequestBody UtmRegexPattern pattern) {
        final String ctx = CLASSNAME + ".updateRegexPattern";
        try {
            regexPatternService.updateRegexPattern(pattern);
            return ResponseEntity.noContent().build();
        } catch (BadRequestException e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.BAD_REQUEST, msg);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * {@code DELETE  /regex-pattern/:id} : Remove a regex pattern.
     *
     * @param id the id of the regex pattern to remove.
     * @return the {@link ResponseEntity} with status {@code 204 (No Content)}, with status {@code 400 (Bad Request)}, or with status {@code 500 (Internal)} if errors occurred.
     */
    @DeleteMapping("/regex-pattern/{id}")
    public ResponseEntity<Void> removeRegexPattern(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".removeRegexPattern";
        try {
            regexPatternService.delete(id);
            return ResponseEntity.noContent().build();
        } catch (BadRequestException e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.BAD_REQUEST, msg);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * GET  /regex-pattern : get all the regex patterns.
     *
     * @param pageable the pagination information
     * @return the {@link List} of {@link UtmRegexPattern} in body with status {@code 200 (OK)}, or with status {@code 500 (Internal)} if errors occurred.
     */
    @GetMapping("/regex-pattern")
    public ResponseEntity<List<UtmRegexPattern>> getAllRegexPatterns(@RequestParam(required = false) String search, Pageable pageable) {
        final String ctx = CLASSNAME + ".getAllRegexPatterns";
        try {
            Page<UtmRegexPattern> page = regexPatternService.findAll(search, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/regex-pattern");
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * GET  /regex-pattern/:id : The id of the regex pattern.
     *
     * @param id the id of the regex pattern to retrieve
     * @return the {@link ResponseEntity} of {@link UtmRegexPattern} with status 200 (OK) and with body the regex pattern, or with status 404 (Not Found)
     */
    @GetMapping("/regex-pattern/{id}")
    public ResponseEntity<UtmRegexPattern> getRegexPattern(@PathVariable Long id) {
        log.debug("REST request to get UtmRegexPattern : {}", id);
        Optional<UtmRegexPattern> pattern = regexPatternService.findOne(id);
        return ResponseUtil.wrapOrNotFound(pattern);
    }
}

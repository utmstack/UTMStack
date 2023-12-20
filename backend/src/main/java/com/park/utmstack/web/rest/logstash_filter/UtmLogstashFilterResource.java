package com.park.utmstack.web.rest.logstash_filter;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.logstash_filter.UtmLogstashFilter;
import com.park.utmstack.domain.logstash_pipeline.UtmGroupLogstashPipelineFilters;
import com.park.utmstack.domain.logstash_pipeline.UtmLogstashPipeline;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.logstash_filter.UtmLogstashFilterCriteria;
import com.park.utmstack.service.logstash_filter.UtmLogstashFilterQueryService;
import com.park.utmstack.service.logstash_filter.UtmLogstashFilterService;
import com.park.utmstack.service.logstash_pipeline.UtmGroupLogstashPipelineFiltersService;
import com.park.utmstack.service.logstash_pipeline.UtmLogstashPipelineService;
import com.park.utmstack.service.logstash_pipeline.enums.PipelineRelation;
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
 * REST controller for managing UtmLogstashFilter.
 */
@RestController
@RequestMapping("/api")
public class UtmLogstashFilterResource {

    private final Logger log = LoggerFactory.getLogger(UtmLogstashFilterResource.class);

    private static final String CLASSNAME = "UtmLogstashFilterResource";

    private final UtmLogstashFilterService logstashFilterService;
    private final UtmLogstashFilterQueryService logstashFilterQueryService;
    private final UtmGroupLogstashPipelineFiltersService groupLogstashPipelineFiltersService;
    private final UtmLogstashPipelineService pipelineService;
    private final ApplicationEventService applicationEventService;

    public UtmLogstashFilterResource(UtmLogstashFilterService utmLogstashFilterService,
                                     UtmLogstashFilterQueryService logstashFilterQueryService,
                                     UtmGroupLogstashPipelineFiltersService groupLogstashPipelineFiltersService, UtmLogstashPipelineService pipelineService, ApplicationEventService applicationEventService) {
        this.logstashFilterService = utmLogstashFilterService;
        this.logstashFilterQueryService = logstashFilterQueryService;
        this.groupLogstashPipelineFiltersService = groupLogstashPipelineFiltersService;
        this.pipelineService = pipelineService;
        this.applicationEventService = applicationEventService;
    }

    /**
     * POST  /utm-logstash-filters : Create a new utmLogstashFilter.
     *
     * @param logstashFilter the utmLogstashFilter to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmLogstashFilter, or with status 400 (Bad Request) if the utmLogstashFilter has already an ID
     */
    @PostMapping("/logstash-filters")
    public ResponseEntity<UtmLogstashFilter> createLogstashFilter(@Valid @RequestBody UtmLogstashFilter logstashFilter,
                                                                  @RequestParam Long pipelineId) {
        final String ctx = CLASSNAME + ".createLogstashFilter";
        try {
            if (logstashFilter.getId() != null)
                throw new Exception("A new logstash filter cannot already have an ID");

            // If you provide a pipelineId we create relation, otherwise only create the filter
            if (pipelineId!=null) {
                Optional<UtmLogstashPipeline> pipeline = pipelineService.findOne(pipelineId);
                if (!pipeline.isPresent()) {
                    throw new Exception("The pipeline with ID (" + pipelineId + ") not exists");
                }
                UtmLogstashFilter filter = logstashFilterService.save(logstashFilter);
                Long filterId = filter.getId();
                UtmGroupLogstashPipelineFilters relation = new UtmGroupLogstashPipelineFilters();
                relation.setFilterId(filterId.intValue());
                relation.setPipelineId(pipelineId.intValue());
                relation.setRelation(PipelineRelation.USER_CUSTOM_FILTER.getRelation());
                groupLogstashPipelineFiltersService.save(relation);
                return ResponseEntity.ok(filter);
            } else {
                return ResponseEntity.ok(logstashFilterService.save(logstashFilter));
            }

        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * PUT  /utm-logstash-filters : Updates an existing utmLogstashFilter.
     *
     * @param logstashFilter the utmLogstashFilter to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmLogstashFilter,
     * or with status 400 (Bad Request) if the utmLogstashFilter is not valid,
     * or with status 500 (Internal Server Error) if the utmLogstashFilter couldn't be updated
     */
    @PutMapping("/logstash-filters")
    public ResponseEntity<UtmLogstashFilter> updateLogstashFilter(@Valid @RequestBody UtmLogstashFilter logstashFilter) {
        final String ctx = CLASSNAME + ".updateLogstashFilter";
        try {
            if (logstashFilter.getId() == null)
                throw new Exception("Logstash filter id is null");

            return ResponseEntity.ok(logstashFilterService.save(logstashFilter));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * GET  /utm-logstash-filters : get all the utmLogstashFilters.
     *
     * @param pageable the pagination information
     * @return the ResponseEntity with status 200 (OK) and the list of utmLogstashFilters in body
     */
    @GetMapping("/logstash-filters")
    public ResponseEntity<List<UtmLogstashFilter>> getAllLogstashFilters(UtmLogstashFilterCriteria criteria, Pageable pageable) {
        final String ctx = CLASSNAME + ".getAllLogstashFilters";
        try {
            Page<UtmLogstashFilter> page = logstashFilterQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/logstash-filters");
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
     * GET  /logstash-filters/by-pipelineid : get all the utmLogstashFilters associated to a pipeline.
     *
     * @param pipelineId the pipeline Id to search for
     * @return the ResponseEntity with status 200 (OK) and the list of {@link UtmLogstashFilter} in body
     */
    @GetMapping("/logstash-filters/by-pipelineid")
    public ResponseEntity<List<UtmLogstashFilter>> filtersByPipelineId(@RequestParam Long pipelineId) {
        final String ctx = CLASSNAME + ".filtersByPipelineId";
        try {
            List<UtmLogstashFilter> filtersByPipeline = logstashFilterService.filtersByPipelineId(pipelineId);
            return ResponseEntity.ok().body(filtersByPipeline);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * GET  /utm-logstash-filters/:id : get the "id" utmLogstashFilter.
     *
     * @param id the id of the utmLogstashFilter to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmLogstashFilter, or with status 404 (Not Found)
     */
    @GetMapping("/logstash-filters/{id}")
    public ResponseEntity<UtmLogstashFilter> getLogstashFilter(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".getLogstashFilter";
        try {
            Optional<UtmLogstashFilter> utmLogstashFilter = logstashFilterService.findOne(id);
            return ResponseUtil.wrapOrNotFound(utmLogstashFilter);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * DELETE  /utm-logstash-filters/:id : delete the "id" utmLogstashFilter.
     *
     * @param id the id of the utmLogstashFilter to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/logstash-filters/{id}")
    public ResponseEntity<Void> deleteLogstashFilter(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".deleteLogstashFilter";
        try {
            logstashFilterService.delete(id);
            return ResponseEntity.ok().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }
}

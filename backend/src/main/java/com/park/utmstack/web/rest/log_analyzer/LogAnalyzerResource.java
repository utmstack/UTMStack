package com.park.utmstack.web.rest.log_analyzer;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.log_analyzer.LogAnalyzerQuery;
import com.park.utmstack.security.SecurityUtils;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.log_analyzer.LogAnalyzerQueryCriteria;
import com.park.utmstack.service.log_analyzer.LogAnalyzerQueryService;
import com.park.utmstack.service.log_analyzer.LogAnalyzerService;
import com.park.utmstack.util.exceptions.UtmLogAnalyzerException;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import org.opensearch.client.opensearch._types.aggregations.CalendarInterval;
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
import java.time.Instant;
import java.util.List;
import java.util.Optional;

@RestController
@RequestMapping("/api/log-analyzer")
public class LogAnalyzerResource {

    private final Logger log = LoggerFactory.getLogger(LogAnalyzerResource.class);
    private static final String CLASS_NAME = "LogAnalyzerResource";
    private final LogAnalyzerService logAnalyzerService;
    private final LogAnalyzerQueryService logAnalyzerQueryService;
    private final ApplicationEventService applicationEventService;
    private static final String ENTITY_NAME = "logAnalyzerQuery";

    public LogAnalyzerResource(LogAnalyzerService logAnalyzerService,
                               LogAnalyzerQueryService logAnalyzerQueryService,
                               ApplicationEventService applicationEventService) {
        this.logAnalyzerService = logAnalyzerService;
        this.logAnalyzerQueryService = logAnalyzerQueryService;
        this.applicationEventService = applicationEventService;
    }

    /**
     * @param filters
     * @param field
     * @param top
     * @param pageable
     * @return
     */
    @PostMapping("/top-x-values/{indexPattern}/{field}/{top}")
    public ResponseEntity<LogAnalyzerService.TopValuesResult> getTopXValues(@RequestBody(required = false) List<FilterType> filters,
                                                                            @PathVariable String indexPattern,
                                                                            @PathVariable String field,
                                                                            @PathVariable int top, Pageable pageable) {
        final String ctx = CLASS_NAME + ".getTopXValues";

        try {
            return ResponseEntity.ok(logAnalyzerService.topXValues(filters, top, field, indexPattern, pageable));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * POST  /log-analyzer-queries : Create a new logAnalyzerQuery.
     *
     * @param logAnalyzerQuery the logAnalyzerQuery to create
     * @return the ResponseEntity with status 201 (Created) and with body the new logAnalyzerQuery, or with status 400 (Bad
     * Request) if the logAnalyzerQuery has already an ID
     */
    @PostMapping("/queries")
    public ResponseEntity<LogAnalyzerQuery> createLogAnalyzerQuery(@Valid @RequestBody LogAnalyzerQuery logAnalyzerQuery) {
        final String ctx = CLASS_NAME + ".createLogAnalyzerQuery";
        try {
            if (logAnalyzerQuery.getId() != null)
                throw new UtmLogAnalyzerException("A new LogAnalyzerQuery cannot already have an ID");

            SecurityUtils.getCurrentUserLogin().ifPresent(logAnalyzerQuery::setOwner);
            logAnalyzerQuery.setCreationDate(Instant.now());
            LogAnalyzerQuery result = logAnalyzerService.save(logAnalyzerQuery);
            return ResponseEntity.ok(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * PUT  /log-analyzer-queries : Updates an existing logAnalyzerQuery.
     *
     * @param logAnalyzerQuery the logAnalyzerQuery to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated logAnalyzerQuery, or with status 400 (Bad
     * Request) if the logAnalyzerQuery is not valid, or with status 500 (Internal Server Error) if the logAnalyzerQuery
     * couldn't be updated
     */
    @PutMapping("/queries")
    public ResponseEntity<LogAnalyzerQuery> updateLogAnalyzerQuery(@Valid @RequestBody LogAnalyzerQuery logAnalyzerQuery) {
        final String ctx = CLASS_NAME + ".updateLogAnalyzerQuery";
        try {
            if (logAnalyzerQuery.getId() == null)
                throw new UtmLogAnalyzerException("Invalid id");

            logAnalyzerQuery.setModificationDate(Instant.now());
            LogAnalyzerQuery result = logAnalyzerService.save(logAnalyzerQuery);
            return ResponseEntity.ok().headers(
                HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, logAnalyzerQuery.getId().toString())).body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * GET  /log-analyzer-queries : get all the logAnalyzerQueries.
     *
     * @param pageable the pagination information
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the list of logAnalyzerQueries in body
     */
    @GetMapping("/queries")
    public ResponseEntity<List<LogAnalyzerQuery>> getAllLogAnalyzerQueries(LogAnalyzerQueryCriteria criteria,
                                                                           Pageable pageable) {
        final String ctx = CLASS_NAME + ".getAllLogAnalyzerQueries";
        try {
            Page<LogAnalyzerQuery> page = logAnalyzerQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/log-analyzer/queries");
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
     * GET  /log-analyzer-queries/:id : get the "id" logAnalyzerQuery.
     *
     * @param id the id of the logAnalyzerQuery to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the logAnalyzerQuery, or with status 404 (Not Found)
     */
    @GetMapping("/queries/{id}")
    public ResponseEntity<LogAnalyzerQuery> getLogAnalyzerQuery(@PathVariable Long id) {
        final String ctx = CLASS_NAME + ".getLogAnalyzerQuery";
        try {
            Optional<LogAnalyzerQuery> logAnalyzerQuery = logAnalyzerService.findOne(id);
            return ResponseUtil.wrapOrNotFound(logAnalyzerQuery);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * DELETE  /log-analyzer-queries/:id : delete the "id" logAnalyzerQuery.
     *
     * @param id the id of the logAnalyzerQuery to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/queries/{id}")
    public ResponseEntity<Void> deleteLogAnalyzerQuery(@PathVariable Long id) {
        final String ctx = CLASS_NAME + ".deleteLogAnalyzerQuery";
        try {
            logAnalyzerService.delete(id);
            return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, id.toString())).build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @PostMapping("/chart-view")
    public ResponseEntity<LogAnalyzerService.ChartView> getChartViewData(@RequestBody(required = false) ChartViewRequest request) {
        final String ctx = CLASS_NAME + ".dateHistogramView";
        try {
            LogAnalyzerService.ChartView result = logAnalyzerService.getChartViewData(request.getIndexPattern(),
                request.getFilters(),
                request.getInterval(),
                request.getTop(), request.getField(),
                request.getFieldDataType());
            return ResponseEntity.ok().body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            if (msg.contains("too_many_buckets_exception"))
                return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlertWithCode(msg, "", "1000")).body(null);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    public static class ChartViewRequest {
        private String indexPattern;
        private List<FilterType> filters;
        private CalendarInterval interval;
        private Integer top;
        private String field;
        private String fieldDataType;

        public String getIndexPattern() {
            return indexPattern;
        }

        public void setIndexPattern(String indexPattern) {
            this.indexPattern = indexPattern;
        }

        public List<FilterType> getFilters() {
            return filters;
        }

        public void setFilters(List<FilterType> filters) {
            this.filters = filters;
        }

        public CalendarInterval getInterval() {
            return interval;
        }

        public void setInterval(CalendarInterval interval) {
            this.interval = interval;
        }

        public Integer getTop() {
            return top;
        }

        public void setTop(Integer top) {
            this.top = top;
        }

        public String getField() {
            return field;
        }

        public void setField(String field) {
            this.field = field;
        }

        public String getFieldDataType() {
            return fieldDataType;
        }

        public void setFieldDataType(String fieldDataType) {
            this.fieldDataType = fieldDataType;
        }
    }
}

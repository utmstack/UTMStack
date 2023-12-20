package com.park.utmstack.web.rest.chart_builder;

import com.fasterxml.jackson.databind.node.ObjectNode;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.chart_builder.UtmVisualization;
import com.park.utmstack.security.SecurityUtils;
import com.park.utmstack.service.UtmStackService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.chart_builder.UtmVisualizationQueryService;
import com.park.utmstack.service.chart_builder.UtmVisualizationService;
import com.park.utmstack.service.dto.chart_builder.UtmVisualizationCriteria;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.util.chart_builder.elasticsearch_dsl.requests.RequestDsl;
import com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.ResponseParser;
import com.park.utmstack.util.chart_builder.elasticsearch_dsl.responses.ResponseParserFactory;
import com.park.utmstack.util.exceptions.UtmChartBuilderException;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import org.opensearch.client.opensearch.core.SearchResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.dao.DataIntegrityViolationException;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.util.Assert;
import org.springframework.util.CollectionUtils;
import org.springframework.web.bind.annotation.*;
import tech.jhipster.web.util.ResponseUtil;

import javax.validation.Valid;
import java.net.URI;
import java.time.Instant;
import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.util.Collections;
import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing UtmVisualization.
 */
@RestController
@RequestMapping("/api")
public class UtmVisualizationResource {

    private final Logger log = LoggerFactory.getLogger(UtmVisualizationResource.class);

    private static final String ENTITY_NAME = "utmVisualization";
    private static final String CLASSNAME = "UtmVisualizationResource";

    private final UtmVisualizationService visualizationService;
    private final UtmVisualizationQueryService visualizationQueryService;
    private final ResponseParserFactory responseParserFactory;
    private final ApplicationEventService applicationEventService;
    private final UtmStackService utmStackService;
    private final ElasticsearchService elasticsearchService;

    public UtmVisualizationResource(UtmVisualizationService visualizationService,
                                    UtmVisualizationQueryService visualizationQueryService,
                                    ResponseParserFactory responseParserFactory,
                                    ApplicationEventService applicationEventService,
                                    UtmStackService utmStackService,
                                    ElasticsearchService elasticsearchService) {
        this.visualizationService = visualizationService;
        this.visualizationQueryService = visualizationQueryService;
        this.responseParserFactory = responseParserFactory;
        this.applicationEventService = applicationEventService;
        this.elasticsearchService = elasticsearchService;
        this.utmStackService = utmStackService;
    }

    /**
     * POST  /utm-visualizations : Create a new utmVisualization.
     *
     * @param utmVisualization the utmVisualization to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmVisualization, or with status 400 (Bad
     * Request) if the utmVisualization has already an ID
     */
    @PostMapping("/utm-visualizations")
    public ResponseEntity<UtmVisualization> createUtmVisualization(@Valid @RequestBody UtmVisualization utmVisualization) {
        final String ctx = CLASSNAME + ".createUtmVisualization";

        UtmVisualization result = null;
        try {
            if (utmVisualization.getId() != null)
                throw new BadRequestAlertException("A new utmVisualization cannot already have an ID", ENTITY_NAME, "idexists");

            RequestDsl requestQuery = new RequestDsl(utmVisualization);
            utmVisualization.setQuery(requestQuery.getSearchSourceBuilder().toString());

            SecurityUtils.getCurrentUserLogin().ifPresent(utmVisualization::setUserCreated);
            utmVisualization.setCreatedDate(LocalDateTime.now().toInstant(ZoneOffset.UTC));

            // Set the next system sequence value only if the environment is dev
            // All visualizations created under the development environment are considered as from the system
            if (utmStackService.isInDevelop()) {
                utmVisualization.setId(visualizationService.getSystemSequenceNextValue());
                utmVisualization.setSystemOwner(true);
            }

            result = visualizationService.save(utmVisualization);
            return ResponseEntity.created(new URI("/api/utm-visualizations/" + result.getId()))
                .headers(HeaderUtil.createEntityCreationAlert(ENTITY_NAME, result.getId().toString())).body(result);
        } catch (DataIntegrityViolationException e) {
            String msg = ctx + ": " + e.getMostSpecificCause().getMessage().replaceAll("\n", "");
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.CONFLICT).headers(HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(result);
        }
    }

    @PostMapping("/utm-visualizations/batch")
    public ResponseEntity<Void> createBatchUtmVisualization(@Valid @RequestBody CreateBatchVisualizationBody body) {
        final String ctx = CLASSNAME + ".createBatchUtmVisualization";

        try {
            List<UtmVisualization> visualizations = body.getVisualizations();
            boolean inDevelop = utmStackService.isInDevelop();

            for (int i = 0; i < visualizations.size(); i++) {
                UtmVisualization visualization = visualizations.get(i);
                Optional<UtmVisualization> utmVisualization = visualizationService.findByName(visualization.getName());

                if (utmVisualization.isPresent()) {
                    UtmVisualization eVisualization = utmVisualization.get();
                    if (body.getOverride()) {
                        visualization.setId(eVisualization.getId());
                        visualization.setModifiedDate(LocalDateTime.now().toInstant(ZoneOffset.UTC));
                        visualization.setSystemOwner(inDevelop);
                    } else {
                        visualizations.remove(i);
                        i--;
                    }
                } else {
                    visualization.setId(inDevelop ? visualizationService.getSystemSequenceNextValue() : null);
                    visualization.setSystemOwner(inDevelop);
                    visualization.setCreatedDate(LocalDateTime.now().toInstant(ZoneOffset.UTC));
                    visualization.setUserCreated(SecurityUtils.getCurrentUserLogin().orElse("system"));
                }
            }

            if (!CollectionUtils.isEmpty(visualizations))
                visualizationService.saveAll(visualizations);

            return ResponseEntity.ok().build();
        } catch (DataIntegrityViolationException e) {
            String msg = ctx + ": " + e.getMostSpecificCause().getMessage().replaceAll("\n", "");
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.CONFLICT).headers(HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).build();
        }
    }

    /**
     * PUT  /utm-visualizations : Updates an existing utmVisualization.
     *
     * @param utmVisualization the utmVisualization to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmVisualization, or with status 400 (Bad
     * Request) if the utmVisualization is not valid, or with status 500 (Internal Server Error) if the utmVisualization
     * couldn't be updated
     */
    @PutMapping("/utm-visualizations")
    public ResponseEntity<UtmVisualization> updateUtmVisualization(@Valid @RequestBody UtmVisualization utmVisualization) {
        final String ctx = CLASSNAME + ".updateUtmVisualization";
        if (utmVisualization.getId() == null) {
            throw new BadRequestAlertException("Invalid id", ENTITY_NAME, "idnull");
        }
        UtmVisualization result = null;
        try {
            RequestDsl requestQuery = new RequestDsl(utmVisualization);
            utmVisualization.setQuery(requestQuery.getSearchSourceBuilder().toString());

            SecurityUtils.getCurrentUserLogin().ifPresent(utmVisualization::setUserModified);
            utmVisualization.setModifiedDate(Instant.now());
            utmVisualization.setSystemOwner(utmVisualization.getSystemOwner() == null ? utmVisualization.getId() < 1000000 : utmVisualization.getSystemOwner());

            result = visualizationService.save(utmVisualization);
            return ResponseEntity.ok().headers(HeaderUtil.createEntityUpdateAlert(ENTITY_NAME, utmVisualization.getId().toString())).body(result);
        } catch (DataIntegrityViolationException e) {
            String msg = ctx + ": " + e.getMostSpecificCause().getMessage().replaceAll("\n", "");
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.CONFLICT).headers(HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(HeaderUtil.createFailureAlert(ENTITY_NAME, null, msg)).body(result);
        }
    }

    /**
     * GET  /utm-visualizations : get all the utmVisualizations.
     *
     * @param pageable the pagination information
     * @return the ResponseEntity with status 200 (OK) and the list of utmVisualizations in body
     */
    @GetMapping("/utm-visualizations")
    public ResponseEntity<List<UtmVisualization>> getAllUtmVisualizations(UtmVisualizationCriteria criteria, Pageable pageable) {
        final String ctx = CLASSNAME + ".getAllUtmVisualizations";
        try {
            Page<UtmVisualization> page = visualizationQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-visualizations");
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    /**
     * GET  /utm-visualizations/:id : get the "id" utmVisualization.
     *
     * @param id the id of the utmVisualization to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmVisualization, or with status 404 (Not Found)
     */
    @GetMapping("/utm-visualizations/{id}")
    public ResponseEntity<UtmVisualization> getUtmVisualization(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".getUtmVisualization";
        try {
            Optional<UtmVisualization> utmVisualization = visualizationService.findOne(id);
            return ResponseUtil.wrapOrNotFound(utmVisualization);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    /**
     * DELETE  /utm-visualizations/:id : delete the "id" utmVisualization.
     *
     * @param id the id of the utmVisualization to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/utm-visualizations/{id}")
    public ResponseEntity<Void> deleteUtmVisualization(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".deleteUtmVisualization";
        try {
            visualizationService.delete(id);
            return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, id.toString())).build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    @DeleteMapping("/utm-visualizations/bulk-delete")
    public ResponseEntity<Void> bulkDelete(@RequestParam List<Long> ids) {
        final String ctx = CLASSNAME + ".bulkDelete";
        try {
            visualizationService.deleteByIdIn(ids);
            return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, ids.toString())).build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    @PostMapping("/utm-visualizations/run")
    public ResponseEntity<List<?>> run(@RequestBody UtmVisualization visualization) throws UtmChartBuilderException {
        final String ctx = CLASSNAME + ".run";
        try {
            Assert.notNull(visualization, "Param utmVisualization must not be null");

            if (!elasticsearchService.indexExist(visualization.getPattern().getPattern()))
                return ResponseEntity.ok(Collections.emptyList());

            RequestDsl requestQuery = new RequestDsl(visualization);
            SearchResponse<ObjectNode> result = elasticsearchService.search(requestQuery.getSearchSourceBuilder().build(), ObjectNode.class);
            ResponseParser<?> responseParser = responseParserFactory.instance(visualization.getChartType());
            return ResponseEntity.ok().body(responseParser.parse(visualization, result));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    public static class CreateBatchVisualizationBody {
        private List<UtmVisualization> visualizations;
        private Boolean override;

        public List<UtmVisualization> getVisualizations() {
            return visualizations;
        }

        public void setVisualizations(List<UtmVisualization> visualizations) {
            this.visualizations = visualizations;
        }

        public Boolean getOverride() {
            return override;
        }

        public void setOverride(Boolean override) {
            this.override = override;
        }
    }
}

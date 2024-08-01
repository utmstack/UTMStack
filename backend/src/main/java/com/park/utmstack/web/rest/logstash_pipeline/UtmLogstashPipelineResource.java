package com.park.utmstack.web.rest.logstash_pipeline;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.logstash_pipeline.UtmLogstashPipeline;
import com.park.utmstack.domain.logstash_pipeline.types.PipelineErrors;
import com.park.utmstack.repository.logstash_pipeline.UtmGroupLogstashPipelineFiltersRepository;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.logstash_pipeline.UtmLogstashPipelineDTO;
import com.park.utmstack.service.logstash_pipeline.UtmLogstashPipelineService;
import com.park.utmstack.service.logstash_pipeline.response.ApiStatsResponse;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.vm.UtmLogstashPipelineVM;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springdoc.api.annotations.ParameterObject;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.util.StringUtils;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.support.ServletUriComponentsBuilder;
import tech.jhipster.web.util.PaginationUtil;

import java.net.URISyntaxException;
import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing {@link UtmLogstashPipeline}.
 */
@RestController
@RequestMapping("/api")
public class UtmLogstashPipelineResource {

    private static final String CLASSNAME = "UtmLogstashPipelineResource";
    private final Logger log = LoggerFactory.getLogger(UtmLogstashPipelineResource.class);

    private static final String ENTITY_NAME = "utmLogstashPipeline";
    private final ApplicationEventService applicationEventService;

    private final UtmLogstashPipelineService utmLogstashPipelineService;

    private final UtmGroupLogstashPipelineFiltersRepository utmGroupLogstashPipelineFiltersRepository;

    public UtmLogstashPipelineResource(
        UtmLogstashPipelineService utmLogstashPipelineService,
        ApplicationEventService applicationEventService,
        UtmGroupLogstashPipelineFiltersRepository utmGroupLogstashPipelineFiltersRepository) {
        this.utmLogstashPipelineService = utmLogstashPipelineService;
        this.applicationEventService = applicationEventService;
        this.utmGroupLogstashPipelineFiltersRepository = utmGroupLogstashPipelineFiltersRepository;
    }


    /**
     * {@code GET  /logstash-pipelines} : Get all active pipelines general information.
     *
     * @param pageable the pagination parameters.
     * @return the {@link ResponseEntity} with a list of {@link UtmLogstashPipelineDTO} representing the active pipelines and its status, with status {@code 200 (Ok)}, or with status {@code 500 (Internal)} if errors occurred.
     * @throws URISyntaxException if the Location URI syntax is incorrect.
     */
    @GetMapping("/logstash-pipelines")
    public ResponseEntity<List<UtmLogstashPipelineDTO>> getAllActivePipelines(
        @ParameterObject Pageable pageable
    ) {
        final String ctx = CLASSNAME + ".getAllActivePipelines";
        try {
            Page<UtmLogstashPipelineDTO> page = utmLogstashPipelineService.activePipelines(pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(ServletUriComponentsBuilder.fromCurrentRequest(), page);
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * {@code GET  /logstash-pipelines/stats} : Get general logstash stats and pipeline stats.
     *
     * @return the {@link ResponseEntity} with a {@link ApiStatsResponse} object representing logstash statistics and the active pipelines and its stats, with status {@code 200 (Ok)}, or with status {@code 500 (Internal)} if errors occurred.
     * @throws URISyntaxException if the Location URI syntax is incorrect.
     */
    @GetMapping("/logstash-pipelines/stats")
    public ResponseEntity<ApiStatsResponse> getLogstashStats() {
        final String ctx = CLASSNAME + ".getLogstashStats";
        try {
            ApiStatsResponse statsResponse = utmLogstashPipelineService.getLogstashStats();
            return ResponseEntity.ok().body(statsResponse);
        } catch (org.springframework.web.client.ResourceAccessException ex) {
            String msg = ctx + ": Logstash server can't be reached, may be it's down, check the message -> " + ex.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * {@code GET  /logstash-pipelines/:id} : Get a pipeline with all of its components (filters, inputs, ports).
     *
     * @return the {@link ResponseEntity} with a {@link UtmLogstashPipelineVM} object representing a pipelines and its components, with status {@code 200 (Ok)}, or with status {@code 500 (Internal)} if errors occurred.
     * @throws URISyntaxException if the Location URI syntax is incorrect.
     */
    @GetMapping("/logstash-pipelines/{id}")
    public ResponseEntity<UtmLogstashPipelineVM> getUtmLogstashPipeline(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".getUtmLogstashPipeline";
        try {
            Optional<UtmLogstashPipeline> utmLogstashPipeline = utmLogstashPipelineService.findOne(id);
            UtmLogstashPipelineVM utmLogstashPipelineVM = new UtmLogstashPipelineVM();
            if (utmLogstashPipeline != null && utmLogstashPipeline.isPresent()) {
                Integer pipelineId = utmLogstashPipeline.get().getId().intValue();
                utmLogstashPipelineVM.setPipelineDTO(new UtmLogstashPipelineDTO(utmLogstashPipeline.get()));
                utmLogstashPipelineVM.setFilters(utmGroupLogstashPipelineFiltersRepository.getFilters(pipelineId));
            }
            if (utmLogstashPipelineVM.getPipelineDTO() == null) {
                throw new RuntimeException("Pipeline not found");
            } else
                return ResponseEntity.ok().body(utmLogstashPipelineVM);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * {@code GET  /logstash-pipelines/validate} : Validates a pipeline information (UtmLogstashPipelineVM).
     *
     * @param pipeline the {@link UtmLogstashPipelineVM} to validate.
     * @param mode     the mode of validation, can be (INSERT or UPDATE)
     * @return Status {@code 204 (No Content)} if no validation errors or {@link PipelineErrors} representing all validation errors detected with status {@code 500 (Internal)}.
     * @throws URISyntaxException if the Location URI syntax is incorrect.
     */
    @PostMapping("/logstash-pipelines/validate")
    public ResponseEntity<PipelineErrors> validatePipelines(@RequestBody UtmLogstashPipelineVM pipeline, @ParameterObject String mode) {
        final String ctx = CLASSNAME + ".validatePipelines";
        try {
            if (!StringUtils.hasText(mode)) {
                throw new Exception("The value of mode that you provide is missing, only INSERT or UPDATE are allowed");
            }
            PipelineErrors errors = utmLogstashPipelineService.validatePipeline(pipeline, mode);
            return errors == null || errors.getValidationErrors().isEmpty() ? ResponseEntity.noContent().build() : ResponseEntity.internalServerError().body(errors);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * {@code DELETE  /logstash-pipelines/:id} : Removes a pipeline with given id.
     *
     * @param id the id of the {@link UtmLogstashPipeline} to delete.
     * @return the {@link ResponseEntity} with status {@code 204 (NO_CONTENT)}.
     */
    @DeleteMapping("/logstash-pipelines/{id}")
    public ResponseEntity<Void> deleteUtmLogstashPipeline(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".deleteUtmLogstashPipeline";
        try {
            Optional<UtmLogstashPipeline> toDelete = utmLogstashPipelineService.findOne(id);
            if (toDelete.isPresent() && toDelete.get().getSystemOwner()) {
                throw new BadRequestAlertException("You can't delete system pipelines", "pipelineManagement", "system pipeline");
            }
            if (toDelete == null || !toDelete.isPresent()) {
                throw new BadRequestAlertException("The pipeline you're trying to delete don't exist", "pipelineManagement", "system pipeline");
            }
            utmLogstashPipelineService.delete(id);
            return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, id.toString())).build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }
}

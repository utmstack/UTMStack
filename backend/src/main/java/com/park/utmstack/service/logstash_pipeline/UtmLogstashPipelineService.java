package com.park.utmstack.service.logstash_pipeline;

import com.park.utmstack.domain.logstash_filter.UtmLogstashFilter;
import com.park.utmstack.domain.logstash_pipeline.UtmGroupLogstashPipelineFilters;
import com.park.utmstack.domain.logstash_pipeline.UtmLogstashPipeline;
import com.park.utmstack.domain.logstash_pipeline.enums.PipelineValidationMode;
import com.park.utmstack.domain.logstash_pipeline.types.*;
import com.park.utmstack.repository.logstash_filter.UtmLogstashFilterRepository;
import com.park.utmstack.repository.logstash_pipeline.UtmGroupLogstashPipelineFiltersRepository;
import com.park.utmstack.repository.logstash_pipeline.UtmLogstashPipelineRepository;
import com.park.utmstack.service.dto.logstash_pipeline.UtmLogstashPipelineDTO;
import com.park.utmstack.service.logstash_pipeline.enums.PipelineStatus;
import com.park.utmstack.service.logstash_pipeline.response.ApiPipelineResponse;
import com.park.utmstack.service.logstash_pipeline.response.ApiStatsResponse;
import com.park.utmstack.service.logstash_pipeline.response.engine.ApiEngineResponse;
import com.park.utmstack.service.logstash_pipeline.response.pipeline.PipelineData;
import com.park.utmstack.service.logstash_pipeline.response.pipeline.PipelineStats;
import com.park.utmstack.service.web_clients.rest_template.RestTemplateService;
import com.park.utmstack.web.rest.vm.UtmLogstashPipelineVM;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.support.PagedListHolder;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.support.PageableExecutionUtils;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.StringUtils;

import java.util.*;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.stream.Collectors;

/**
 * Service Implementation for managing {@link UtmLogstashPipeline}.
 */
@Service
@Transactional
public class UtmLogstashPipelineService {

    private final Logger log = LoggerFactory.getLogger(UtmLogstashPipelineService.class);

    private final UtmLogstashPipelineRepository utmLogstashPipelineRepository;
    private final RestTemplateService restTemplateService;

    private static final String CLASSNAME = "UtmLogstashPipelineService";

    private final UtmGroupLogstashPipelineFiltersRepository utmGroupLogstashPipelineFiltersRepository;
    private final UtmLogstashFilterRepository utmLogstashFilterRepository;

    private final List<String> logshtashPipelines = Arrays.asList("AZURE","GCP");

    public UtmLogstashPipelineService(UtmLogstashPipelineRepository utmLogstashPipelineRepository, RestTemplateService restTemplateService, UtmGroupLogstashPipelineFiltersRepository utmGroupLogstashPipelineFiltersRepository, UtmLogstashFilterRepository utmLogstashFilterRepository) {
        this.utmLogstashPipelineRepository = utmLogstashPipelineRepository;
        this.restTemplateService = restTemplateService;
        this.utmGroupLogstashPipelineFiltersRepository = utmGroupLogstashPipelineFiltersRepository;
        this.utmLogstashFilterRepository = utmLogstashFilterRepository;
    }

    /**
     * Save a utmLogstashPipeline.
     *
     * @param utmLogstashPipeline the entity to save.
     * @return the persisted entity.
     */
    public UtmLogstashPipeline save(UtmLogstashPipeline utmLogstashPipeline) {
        log.debug("Request to save UtmLogstashPipeline : {}", utmLogstashPipeline);
        if (utmLogstashPipeline.getId() == null) {
            utmLogstashPipeline.setId(utmLogstashPipelineRepository.getNextId());
        }
        return utmLogstashPipelineRepository.save(utmLogstashPipeline);
    }

    /**
     * Save a utmLogstashPipeline.
     *
     * @param utmLogstashPipelineVM the entity to save.
     * @return the persisted entity.
     */
    public void createPipeline(UtmLogstashPipelineVM utmLogstashPipelineVM) throws Exception {
        final String ctx = CLASSNAME + ".createPipeline";
        List<Validation> validationList = validate(utmLogstashPipelineVM, PipelineValidationMode.INSERT);
        if (!validationList.isEmpty()) {
            throw new Exception("Pipeline is not valid, please validate it first before insert");
        } else {
            try {
                // This is a template method, to implement if needed in the future
            } catch (Exception e) {
                throw new Exception(ctx + ": " + e.getMessage());
            }
        }
    }

    /**
     * Update a utmLogstashPipeline with all its dependencies.
     *
     * @param utmLogstashPipelineVM the entity to save.
     * @return the persisted entity.
     */
    public void update(UtmLogstashPipelineVM utmLogstashPipelineVM) throws Exception {
        log.debug("Request to save UtmLogstashPipeline : {}", utmLogstashPipelineVM);
        final String ctx = CLASSNAME + ".update";
        List<Validation> validationList = validate(utmLogstashPipelineVM, PipelineValidationMode.UPDATE);
        if (!validationList.isEmpty()) {
            throw new Exception("Pipeline is not valid, please validate it first before update");
        } else {
            try {
                // This is a template method, to implement if needed in the future
            } catch (Exception e) {
                throw new Exception(ctx + ": " + e.getMessage());
            }
        }
    }

    /**
     * Get all the utmLogstashPipelines.
     *
     * @param pageable the pagination information.
     * @return the list of entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmLogstashPipeline> findAll(Pageable pageable) {
        log.debug("Request to get all UtmLogstashPipelines");
        return utmLogstashPipelineRepository.findAll(pageable);
    }

    /**
     * Get all active utmLogstashPipelines.
     *
     * @param pageable the pagination information.
     * @return the list of entities.
     */
    @Transactional(readOnly = true)
    public Page<UtmLogstashPipelineDTO> activePipelines(Pageable pageable) {
        log.debug("Request to get active UtmLogstashPipelines paginated");
        List<UtmLogstashPipeline> pipelineList = new ArrayList<>(activePipelinesList());
        List<UtmLogstashPipelineDTO> resultList = pipelineList.stream()
                .map(UtmLogstashPipelineDTO::new).collect(Collectors.toList());

        PagedListHolder<UtmLogstashPipelineDTO> pageDefinition = new PagedListHolder<>();
        pageDefinition.setSource(resultList);
        pageDefinition.setPageSize(pageable.getPageSize());
        pageDefinition.setPage(pageable.getPageNumber());
        return PageableExecutionUtils.getPage(pageDefinition.getPageList(), pageable, pipelineList::size);
    }

    /**
     * Get a list of active UtmLogstashPipeline.
     * .
     *
     * @return the list of entities.
     */
    public List<UtmLogstashPipeline> activePipelinesList() {
        log.debug("Request to get active UtmLogstashPipelines");
        List<UtmLogstashPipeline> activePipelines = utmLogstashPipelineRepository.allActivePipelinesByServer();

        return activePipelines;
    }

    /**
     * Get one utmLogstashPipeline by id.
     *
     * @param id the id of the entity.
     * @return the entity.
     */
    @Transactional(readOnly = true)
    public Optional<UtmLogstashPipeline> findOne(Long id) {
        log.debug("Request to get UtmLogstashPipeline : {}", id);
        return utmLogstashPipelineRepository.findById(id);
    }

    /**
     * Delete the utmLogstashPipeline by id.
     *
     * @param id the id of the entity.
     */
    public void delete(Long id) throws Exception {
        final String ctx = CLASSNAME + ".delete";
        log.debug("Request to delete UtmLogstashPipeline : {}", id);
        try {
            if (id == null) {
                throw new Exception("Unable to delete, pipeline id is null");
            }
            Integer pipelineId = id.intValue();
            // First, perform delete on filter group
            List<UtmGroupLogstashPipelineFilters> filterGpList = utmGroupLogstashPipelineFiltersRepository
                    .getFilters(pipelineId);
            utmGroupLogstashPipelineFiltersRepository.deleteAll(filterGpList);
            // Second, delete non system filters associated to this pipeline
            List<UtmLogstashFilter> filters = utmLogstashFilterRepository.findAllByListOfId(filterGpList.stream().map(
                    gpl -> gpl.getFilterId().longValue()
            ).collect(Collectors.toList()));
            if (!filters.isEmpty()) {
                utmLogstashFilterRepository.deleteAll(filters);
            }
            utmLogstashPipelineRepository.deleteById(id);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Validate an utmLogstashPipelineVM.
     *
     * @param utmLogstashPipelineVM the pipeline to validate.
     * @return Null if no errors, else return a PipelineErrors object.
     */
    public PipelineErrors validatePipeline(UtmLogstashPipelineVM utmLogstashPipelineVM, String mode) throws Exception {
        final String ctx = CLASSNAME + ".validatePipeline";
        try {
            List<Validation> validationList = validate(utmLogstashPipelineVM, PipelineValidationMode.valueOf(mode));
            if (!validationList.isEmpty()) {
                return new PipelineErrors(validationList);
            } else {
                return null;
            }
        } catch (IllegalArgumentException e) {
            throw new Exception("The value of mode that you provide is wrong, only INSERT or UPDATE are allowed");
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Getting logstash pipelines information
     */
    private ApiPipelineResponse pipelineApiResponse() {
        final String ctx = CLASSNAME + ".pipelineApiResponse";
        return new ApiPipelineResponse();
    }

    /**
     * Getting logstash jvm information
     */
    public ApiEngineResponse logstashJvmApiResponse() {
        final String ctx = CLASSNAME + ".logstashJvmApiResponse";
        return new ApiEngineResponse();

    }

    /**
     * Getting active pipelines stats from DB, general jvm stats from logstash
     */
    public ApiStatsResponse getLogstashStats() throws Exception {
        final String ctx = CLASSNAME + ".getLogstashStats";
        ApiStatsResponse statsResponse = new ApiStatsResponse();

        // Variables used to set the general pipeline's status
        AtomicInteger activePipelinesCount = new AtomicInteger();
        AtomicInteger upPipelinesCount = new AtomicInteger();
        boolean isCorrelationUp = isEngineUp();

        try {
            // Getting Jvm information (not used)
            ApiEngineResponse jvmData = logstashJvmApiResponse();
            if (jvmData != null) {
                statsResponse.setGeneral(jvmData);
            }
            // List to store stats mapped from DB
            List<PipelineStats> infoStats;

            // Getting the active pipelines statistics
            infoStats = activePipelinesList().stream().map(activePip -> {

                // Calculating stats for pipelines
                // Setting stats for non-logstash pipelines (correlation engine)
                    if (isCorrelationUp) {
                        activePipelinesCount.getAndIncrement(); // Total pipelines that have to be active
                        if (activePip.getPipelineStatus().equals(PipelineStatus.PIPELINE_STATUS_UP.get())) {
                            upPipelinesCount.getAndIncrement();
                        }
                    } else {
                        activePip.setPipelineStatus(PipelineStatus.PIPELINE_STATUS_DOWN.get());
                    }
               // }
                // Mapping stats from DB pipeline
                return PipelineStats.getPipelineStats(activePip);
            }).collect(Collectors.toList());

            // Setting the final global status of pipelines
            if (isCorrelationUp) {
                if (upPipelinesCount.get() == 0) {
                    jvmData.setStatus(PipelineStatus.ENGINE_STATUS_RED.get());
                } else if (upPipelinesCount.get() == activePipelinesCount.get()) {
                    jvmData.setStatus(PipelineStatus.ENGINE_STATUS_GREEN.get());
                } else {
                    jvmData.setStatus(PipelineStatus.ENGINE_STATUS_YELLOW.get());
                }
            }
            statsResponse.setPipelines(infoStats);
        } catch (Exception ex) {
            throw new Exception(ctx + ": " + ex.getMessage());
        }
        return statsResponse;
    }

    /**
     * Method to set the DB pipelines status
     */
    @Scheduled(fixedDelay = 20000, initialDelay = 30000)
    public List<UtmLogstashPipeline> pipelineStatus() {
        final String ctx = CLASSNAME + ".pipelineStatus";
        // Only logstash pipelines get updated for the moment
        // We will add correlation pipelines status update when we know how to get the status or metrics
        List<UtmLogstashPipeline> activeByServer = utmLogstashPipelineRepository.allActivePipelinesByServer();

            try {

                ApiPipelineResponse response = pipelineApiResponse();
                Map<String, Long> mapInit = activeByServer.stream()
                        .collect(Collectors.toMap(UtmLogstashPipeline::getPipelineId, myId -> (
                                getFailures(myId, response)
                        )));

                Thread.sleep(2000);

                ApiPipelineResponse responseLast = pipelineApiResponse();
                Map<String, PipelineData> pipelineInfo = responseLast.getPipelines();

                activeByServer.stream().forEach((p) -> {
                    // Getting stats and updating DB pipeline
                    PipelineData data = pipelineInfo.get(p.getPipelineId());
                    if (data != null) {
                        p.setEventsOut(data.getEvents().getOut());
                    }
                    Long firstFailuresCount = mapInit.get(p.getPipelineId());
                    Long lastFailuresCount = getFailures(p, responseLast);
                    if ((firstFailuresCount != -1 && lastFailuresCount != -1) && lastFailuresCount <= firstFailuresCount) {
                        p.setPipelineStatus(PipelineStatus.PIPELINE_STATUS_UP.get());
                    } else {
                        p.setPipelineStatus(PipelineStatus.PIPELINE_STATUS_DOWN.get());
                    }
                });
                utmLogstashPipelineRepository.saveAll(activeByServer);

            } catch (Exception connectException) {
                String msg = ctx + ": " + connectException.getMessage();
                log.error(msg);
                activeByServer.stream().forEach((p) -> {
                    p.setPipelineStatus(PipelineStatus.PIPELINE_STATUS_DOWN.get());
                });
                utmLogstashPipelineRepository.saveAll(activeByServer);
            }
        //}
        return activeByServer;
    }

    // Method to count the failures by pipeline
    // Will be modified when engine stats is finished
    private Long getFailures(UtmLogstashPipeline e, ApiPipelineResponse response) {
        Map<String, PipelineData> pipelines = response.getPipelines();
        PipelineData pipData = pipelines.get(e.getPipelineId());
        if (pipData != null) {
            return 0L;
            // return pipData.getReloads().getFailures();
        }
        return 0L;
    }

    // Method used to generate unique name based pipelineId
    private String getFormattedPipelineName(String baseName) {
        return baseName.replaceAll("\\W", "_").replaceAll("(_+)", "_").toLowerCase(Locale.ROOT);
    }

    // Method used to validate UtmLogstashPipelineDTO
    private List<Validation> validate(UtmLogstashPipelineVM utmLogstashPipelineVM, PipelineValidationMode MODE) throws Exception {
        final String ctx = CLASSNAME + ".validate";
        List<Validation> validationList = new ArrayList<>();
        try {
            UtmLogstashPipeline utmLogstashPipeline = utmLogstashPipelineVM.getPipelineDTO().getPipeline(null);
            List<UtmGroupLogstashPipelineFilters> filters = utmLogstashPipelineVM.getFilters();

            // Common validations
            if (MODE.equals(PipelineValidationMode.INSERT) || MODE.equals(PipelineValidationMode.UPDATE)) {
                // Pipeline validations
                if (utmLogstashPipeline == null) {
                    Validation val = new Validation("Pipeline", "Pipeline name", "Entire pipeline definition is null");
                    validationList.add(val);
                }
                if (utmLogstashPipeline.getPipelineName() == null || !StringUtils.hasText(utmLogstashPipeline.getPipelineName().trim())) {
                    Validation val = new Validation("Pipeline", "Pipeline name", "Value is null or empty");
                    validationList.add(val);
                }

                // Filters validations
                if (filters == null || filters.isEmpty()) {
                    Validation val = new Validation("Filter relation", "Filter id", "There is no filter associated to the pipeline: " + (utmLogstashPipeline != null ? utmLogstashPipeline.getPipelineName() : "Undefined pipeline"));
                    validationList.add(val);
                } else {
                    filters.forEach(f -> {
                        if (f.getFilterId() == null) {
                            Validation val = new Validation("Filter", "Filter id", "Value is null");
                            validationList.add(val);
                        } else if (f.getFilterId() != null && !utmLogstashFilterRepository.existsById(f.getFilterId().longValue())) {
                            Validation val = new Validation("Filter", "Filter id", "Value " + f.getFilterId() + " not exist");
                            validationList.add(val);
                        }
                        if (MODE.equals(PipelineValidationMode.INSERT)) {
                            if (f.getId() != null) {
                                Validation val = new Validation("Filter relation", "Relation id", "Value must be null when inserting");
                                validationList.add(val);
                            }
                        }
                    });
                }
            }
            // Begining only insert MODE validations
            if (MODE.equals(PipelineValidationMode.INSERT)) {
                // Pipeline validations
                if (utmLogstashPipeline.getId() != null) {
                    Validation val = new Validation("Pipeline", "Pipeline id", "Value must be null when inserting");
                    validationList.add(val);
                }
            }
            // Begining only update MODE validations
            if (MODE.equals(PipelineValidationMode.UPDATE)) {
                // Pipeline validations
                if (utmLogstashPipeline.getId() == null) {
                    Validation val = new Validation("Pipeline", "Pipeline id", "Value can't be null when updating");
                    validationList.add(val);
                }
            }
            return validationList;
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Method to implement later to know if correlation engine is up, if up all the integrations are up.
     */
    private boolean isEngineUp() {
        return true;
    }
}

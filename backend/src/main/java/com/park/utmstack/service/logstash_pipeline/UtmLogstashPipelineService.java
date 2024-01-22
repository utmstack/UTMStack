package com.park.utmstack.service.logstash_pipeline;

import com.park.utmstack.domain.logstash_filter.UtmLogstashFilter;
import com.park.utmstack.domain.logstash_pipeline.UtmGroupLogstashPipelineFilters;
import com.park.utmstack.domain.logstash_pipeline.UtmLogstashInput;
import com.park.utmstack.domain.logstash_pipeline.UtmLogstashInputConfiguration;
import com.park.utmstack.domain.logstash_pipeline.UtmLogstashPipeline;
import com.park.utmstack.domain.logstash_pipeline.enums.InputConfigTypes;
import com.park.utmstack.domain.logstash_pipeline.enums.PipelineValidationMode;
import com.park.utmstack.domain.logstash_pipeline.types.*;
import com.park.utmstack.repository.logstash_filter.UtmLogstashFilterRepository;
import com.park.utmstack.repository.logstash_pipeline.UtmGroupLogstashPipelineFiltersRepository;
import com.park.utmstack.repository.logstash_pipeline.UtmLogstashInputConfigurationRepository;
import com.park.utmstack.repository.logstash_pipeline.UtmLogstashPipelineRepository;
import com.park.utmstack.service.dto.logstash_pipeline.UtmLogstashPipelineDTO;
import com.park.utmstack.service.logstash_pipeline.enums.PipelineRelation;
import com.park.utmstack.service.logstash_pipeline.enums.PipelineStatus;
import com.park.utmstack.service.logstash_pipeline.response.LogstashApiPipelineResponse;
import com.park.utmstack.service.logstash_pipeline.response.LogstashApiStatsResponse;
import com.park.utmstack.service.logstash_pipeline.response.jvm_stats.LogstashApiJvmResponse;
import com.park.utmstack.service.logstash_pipeline.response.pipeline.PipelineData;
import com.park.utmstack.service.logstash_pipeline.response.pipeline.PipelineReloads;
import com.park.utmstack.service.logstash_pipeline.response.pipeline.PipelineStats;
import com.park.utmstack.service.web_clients.rest_template.RestTemplateService;
import com.park.utmstack.web.rest.vm.UtmLogstashPipelineVM;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.support.PagedListHolder;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.support.PageableExecutionUtils;
import org.springframework.http.ResponseEntity;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.StringUtils;

import java.util.*;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.atomic.AtomicReference;
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

    private final String LOGSTASH_URL = System.getenv("LOGSTASH_URL");

    private static final String CLASSNAME = "UtmLogstashPipelineService";

    private final UtmGroupLogstashPipelineFiltersRepository utmGroupLogstashPipelineFiltersRepository;
    private final UtmLogstashInputService utmLogstashInputService;
    private final UtmLogstashInputConfigurationRepository utmLogstashInputConfigurationRepository;
    private final UtmLogstashFilterRepository utmLogstashFilterRepository;
    private final UtmLogstashInputConfigurationService utmLogstashInputConfigurationService;

    public UtmLogstashPipelineService(UtmLogstashPipelineRepository utmLogstashPipelineRepository, RestTemplateService restTemplateService, UtmGroupLogstashPipelineFiltersRepository utmGroupLogstashPipelineFiltersRepository, UtmLogstashInputService utmLogstashInputService, UtmLogstashInputConfigurationRepository utmLogstashInputConfigurationRepository, UtmLogstashFilterRepository utmLogstashFilterRepository, UtmLogstashInputConfigurationService utmLogstashInputConfigurationService) {
        this.utmLogstashPipelineRepository = utmLogstashPipelineRepository;
        this.restTemplateService = restTemplateService;
        this.utmGroupLogstashPipelineFiltersRepository = utmGroupLogstashPipelineFiltersRepository;
        this.utmLogstashInputService = utmLogstashInputService;
        this.utmLogstashInputConfigurationRepository = utmLogstashInputConfigurationRepository;
        this.utmLogstashFilterRepository = utmLogstashFilterRepository;
        this.utmLogstashInputConfigurationService = utmLogstashInputConfigurationService;
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
                UtmLogstashPipeline utmLogstashPipeline = utmLogstashPipelineVM.getPipelineDTO().getPipeline(null);
                List<UtmGroupLogstashPipelineFilters> filters = utmLogstashPipelineVM.getFilters();
                List<InputConfiguration> inputs = utmLogstashPipelineVM.getInputs();

                // Setting default values
                utmLogstashPipeline.setDefaults();
                // Removing all non word characters from pipelineId and make it unique
                Integer pipelineId = utmLogstashPipelineRepository.getNextId().intValue();
                String pipID = getFormattedPipelineName(utmLogstashPipeline.getPipelineName());
                utmLogstashPipeline.setPipelineId(pipID + pipelineId);
                utmLogstashPipeline.setId(pipelineId.longValue());
                this.save(utmLogstashPipeline);

                // Saving filter-pipeline relations
                List<UtmGroupLogstashPipelineFilters> pipFilterGroup = filters.stream().map(group -> {
                    group.setPipelineId(pipelineId);
                    group.setRelation(PipelineRelation.PIPELINE_FILTER.getRelation());
                    return group;
                }).collect(Collectors.toList());
                utmGroupLogstashPipelineFiltersRepository.saveAll(pipFilterGroup);

                // Creating inputs with configuration
                AtomicReference<Integer> configsIterator = new AtomicReference<>(0); // used to generate input configuration ids after first getNextId() call
                inputs.forEach((input) -> {
                    UtmLogstashInput newInput = input.getUtmLogstashInput();
                    newInput.setPipelineId(pipelineId);
                    Long inputId = utmLogstashInputService.save(newInput).getId();
                    List<UtmLogstashInputConfiguration> inputConfigs = input.getConfigs().stream().map(keys -> {
                        UtmLogstashInputConfiguration inputConf = keys.getInputConfiguration(inputId);
                        inputConf.setId(utmLogstashInputConfigurationRepository.getNextId() + (configsIterator.get()));
                        configsIterator.getAndSet(configsIterator.get() + 1);
                        return inputConf;
                    }).collect(Collectors.toList());
                    utmLogstashInputConfigurationRepository.saveAll(inputConfigs);
                });
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
                UtmLogstashPipeline realPipeline = findOne(utmLogstashPipelineVM.getPipelineDTO().getId()).get();
                UtmLogstashPipeline currentPip = utmLogstashPipelineVM.getPipelineDTO().getPipeline(realPipeline);
                List<UtmGroupLogstashPipelineFilters> updateFilters = utmLogstashPipelineVM.getFilters();
                List<InputConfiguration> updateInputs = utmLogstashPipelineVM.getInputs();

                Integer pipelineId = currentPip.getId().intValue();
                // Removing all non word characters from pipelineId and make it unique
                String pipID = getFormattedPipelineName(currentPip.getPipelineName());
                currentPip.setPipelineId(pipID + pipelineId);
                // Search for deleted filters and remove from relation filter-pipeline
                List<UtmGroupLogstashPipelineFilters> currentFilters = utmGroupLogstashPipelineFiltersRepository
                    .getFilters(pipelineId);
                List<UtmGroupLogstashPipelineFilters> deletedFilters = new ArrayList<>();
                currentFilters.forEach(filter -> {
                    Optional<UtmGroupLogstashPipelineFilters> f = updateFilters.stream().filter(update -> update.getId() == filter.getId()).findFirst();
                    if (f == null || !f.isPresent()) {
                        deletedFilters.add(filter);
                    }
                });
                utmGroupLogstashPipelineFiltersRepository.deleteAll(deletedFilters);
                // Save new changes relation filter-pipeline
                utmGroupLogstashPipelineFiltersRepository.saveAll(updateFilters);

                // ------------------ Then work with input configuration changes ------------------//
                List<UtmLogstashInput> currentInputs = utmLogstashInputService.getUtmLogstashInputsByPipelineId(pipelineId);
                List<UtmLogstashInput> deletedInputs = new ArrayList<>();
                List<UtmLogstashInputConfiguration> deletedInputConfigs = new ArrayList<>();
                List<UtmLogstashInput> updateInputList = updateInputs.stream().map(InputConfiguration::getUtmLogstashInput).collect(Collectors.toList());
                currentInputs.forEach(input -> {
                        Optional<UtmLogstashInput> optInput = updateInputList.stream().filter(update -> update.getId() != null && update.getId() == input.getId()).findFirst();
                        if (optInput == null || !optInput.isPresent()) {
                            deletedInputs.add(input);
                        }
                    }
                );
                // Once we have inputs to delete, proceed to get and delete configurations of that inputs
                if (!deletedInputs.isEmpty()) {
                    deletedInputs.forEach(delImp -> {
                        deletedInputConfigs.addAll(utmLogstashInputConfigurationService.getUtmLogstashInputConfigurationsByInputId(delImp.getId().intValue()));
                    });
                }
                // Remove deleted input configs
                if (!deletedInputConfigs.isEmpty()) {
                    utmLogstashInputConfigurationRepository.deleteAll(deletedInputConfigs);
                }
                // Remove deleted inputs
                if (!deletedInputs.isEmpty()) {
                    utmLogstashInputService.deleteAllInputs(deletedInputs);
                }
                // Creating/updating inputs with configuration
                AtomicReference<Integer> configsIterator = new AtomicReference<>(0); // used to generate input configuration ids after first getNextId() call
                updateInputs.forEach((input) -> {
                    UtmLogstashInput newOrUpdateInput = input.getUtmLogstashInput();
                    newOrUpdateInput.setPipelineId(pipelineId);
                    Long inputId = utmLogstashInputService.save(newOrUpdateInput).getId();

                    List<UtmLogstashInputConfiguration> inputConfigs = input.getConfigs().stream().map(keys -> {
                        UtmLogstashInputConfiguration inputConf = keys.getInputConfiguration(inputId);
                        if (inputConf.getId() == null) {
                            inputConf.setId(utmLogstashInputConfigurationRepository.getNextId() + (configsIterator.get()));
                            configsIterator.getAndSet(configsIterator.get() + 1);
                        } else {
                            inputConf.setId(keys.getId());
                        }
                        return inputConf;
                    }).collect(Collectors.toList());
                    utmLogstashInputConfigurationRepository.saveAll(inputConfigs);
                });
                // ------------------ End work with input configuration changes ------------------//
                // Finally, update pipeline
                save(currentPip);

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
     * Get all the pipeline id and ports of active pipelines.
     *
     * @param pageable   the pagination information.
     * @param isInternal to return internal (value=true), external (value=false) or all (value is null) pipelines ports.
     * @return the list of entities (inputId = value of db field pipeline_id, port).
     */
    @Transactional(readOnly = true)
    public Page<PipelinePortConfiguration> activePipelinePorts(Pageable pageable, Boolean isInternal) {
        log.debug("Request to get all pipelines ports");
        List<UtmLogstashPipeline> activePipelines = activePipelinesList();
        List<PipelinePortConfiguration> pipelinePortConfigurationList = new ArrayList<>();
        activePipelines.forEach((pipeline) -> {
            List<InputConfiguration> inputList = utmLogstashInputService.getUtmLogstashInputsByPipelineId(pipeline.getId().intValue())
                .stream().map(InputConfiguration::new).collect(Collectors.toList());
            inputList.forEach((input) -> {
                    List<InputConfigurationKey> inputConfKeys = utmLogstashInputConfigurationService
                        .getUtmLogstashInputConfigurationsByInputId(input.getId().intValue())
                        .stream().map(InputConfigurationKey::new).collect(Collectors.toList());
                    inputConfKeys.forEach((confKey) -> {
                        if (confKey.getConfType().equals(InputConfigTypes.PORT.getValue())) {
                            if (isInternal == null || (isInternal != null && isInternal == pipeline.getPipelineInternal())) {
                                PipelinePortConfiguration portConf = new PipelinePortConfiguration(pipeline.getPipelineId(), confKey.getConfValue());
                                pipelinePortConfigurationList.add(portConf);
                            }
                        }
                    });
                }
            );
        });

        PagedListHolder<PipelinePortConfiguration> pageDefinition = new PagedListHolder<>();
        pageDefinition.setSource(pipelinePortConfigurationList);
        pageDefinition.setPageSize(pageable.getPageSize());
        pageDefinition.setPage(pageable.getPageNumber());
        return PageableExecutionUtils.getPage(pageDefinition.getPageList(), pageable, pipelinePortConfigurationList::size);
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
        List<UtmLogstashPipeline> pipelineList = new ArrayList<>(activePipelines);
        List<Long> parentPipelines = utmLogstashPipelineRepository.getParents();
        if (parentPipelines != null && !parentPipelines.isEmpty()) {
            activePipelines.forEach(activePip -> {
                if (isParent(parentPipelines, activePip.getId())) {
                    Optional<UtmLogstashPipeline> parentPipWithSonActive = activePipelines.stream().filter(
                        v -> v.getParentPipeline() != null && v.getParentPipeline().longValue() == activePip.getId()).findFirst();
                    if (parentPipWithSonActive == null || !parentPipWithSonActive.isPresent()) {
                        pipelineList.remove(activePip);
                    }
                }
            });

        }
        return pipelineList;
    }

    /**
     * To check if a pipeline id is a parent pipeline
     */
    public boolean isParent(List<Long> parents, Long searchValue) {
        Optional<Long> v = parents.stream().filter(val -> val == searchValue).findFirst();
        return (v.isPresent() && v != null);
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

            // Then, perform delete on InputConfigurations
            List<UtmLogstashInput> inputList = utmLogstashInputService.getUtmLogstashInputsByPipelineId(pipelineId);
            inputList.forEach((input) -> {
                List<UtmLogstashInputConfiguration> configs = utmLogstashInputConfigurationRepository
                    .getUtmLogstashInputConfigurationsByInputId(input.getId().intValue());
                if (!configs.isEmpty()) {
                    utmLogstashInputConfigurationRepository.deleteAll(configs);
                }
            });
            // Then, delete Inputs
            if (!inputList.isEmpty()) {
                utmLogstashInputService.deleteAllInputs(inputList);
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
    private LogstashApiPipelineResponse pipelineApiResponse() {
        final String ctx = CLASSNAME + ".pipelineApiResponse";

        LogstashApiPipelineResponse result = new LogstashApiPipelineResponse();

        try {
            ResponseEntity<LogstashApiPipelineResponse> rs = restTemplateService.get(LOGSTASH_URL + "/_node/stats/pipelines?pretty",
                LogstashApiPipelineResponse.class);
            if (!rs.getStatusCode().is2xxSuccessful())
                log.error(ctx + ": " + restTemplateService.extractErrorMessage(rs));

            result = rs.getBody();
        } catch (Exception ex) {
            log.error(ctx + ": " + ex.getMessage());
            return result;
        }
        return result;

    }

    /**
     * Getting logstash jvm information
     */
    public LogstashApiJvmResponse logstashJvmApiResponse() {
        final String ctx = CLASSNAME + ".logstashJvmApiResponse";
        LogstashApiJvmResponse result = new LogstashApiJvmResponse();

        try {
            ResponseEntity<LogstashApiJvmResponse> rs = restTemplateService.get(LOGSTASH_URL + "/_node/stats/jvm",
                LogstashApiJvmResponse.class);
            if (!rs.getStatusCode().is2xxSuccessful())
                log.error(ctx + ": " + restTemplateService.extractErrorMessage(rs));

            result = rs.getBody();
        } catch (Exception ex) {
            log.error(ctx + ": " + ex.getMessage());
            return result;
        }
        return result;

    }

    /**
     * Getting active pipelines and stats from logstash
     */
    public LogstashApiStatsResponse getLogstashStats() throws Exception {
        final String ctx = CLASSNAME + ".getLogstashStats";
        LogstashApiStatsResponse statsResponse = new LogstashApiStatsResponse();

        // Variables used to set the general pipeline's status
        AtomicInteger activePipelinesCount = new AtomicInteger();
        AtomicInteger upPipelinesCount = new AtomicInteger();

        if (!StringUtils.hasText(LOGSTASH_URL)) {
            log.error(ctx + ": The pipeline's status cannot be processed because " +
                "the environment variable LOGSTASH_URL is not configured.");
        }
        try {
            // Getting Logstash Jvm information
            LogstashApiJvmResponse jvmData = logstashJvmApiResponse();
            if (jvmData != null) {
                statsResponse.setGeneral(jvmData);
            }
            // List to store stats mapped from DB
            List<PipelineStats> infoStats;

                // Getting the active pipelines statistics
                infoStats = activePipelinesList().stream().map(activePip -> {

                    if (!jvmData.getStatus().equals(PipelineStatus.LOGSTASH_STATUS_DOWN.get())) {
                        activePipelinesCount.getAndIncrement(); // Total pipelines that have to be active
                        if (activePip.getPipelineStatus().equals(PipelineStatus.PIPELINE_STATUS_UP.get())) {
                            upPipelinesCount.getAndIncrement();
                        }
                    } else {
                        activePip.setPipelineStatus(PipelineStatus.PIPELINE_STATUS_DOWN.get());
                    }
                    // Mapping stats from DB pipeline
                    return PipelineStats.getPipelineStats(activePip);
                }).collect(Collectors.toList());

            // Setting the final global status of pipelines
            if (!jvmData.getStatus().equals(PipelineStatus.LOGSTASH_STATUS_DOWN.get())) {
                if (upPipelinesCount.get() == 0) {
                    jvmData.setStatus(PipelineStatus.LOGSTASH_STATUS_RED.get());
                } else if (upPipelinesCount.get() == activePipelinesCount.get()) {
                    jvmData.setStatus(PipelineStatus.LOGSTASH_STATUS_GREEN.get());
                } else {
                    jvmData.setStatus(PipelineStatus.LOGSTASH_STATUS_YELLOW.get());
                }
            }
            statsResponse.setPipelines(infoStats);
        } catch (org.springframework.web.client.ResourceAccessException ex) {
            throw new Exception(ctx + ": Logstash server can't be reached, may be it's down, check the message -> " + ex.getMessage());
        } catch (Exception ex) {
            throw new Exception(ctx + ": " + ex.getMessage());
        }
        return statsResponse;
    }

    /**
     * Method to get the pipelines status
     */
    @Scheduled(fixedDelay = 20000, initialDelay = 30000)
    public List<UtmLogstashPipeline> pipelineStatus() {
        final String ctx = CLASSNAME + ".pipelineStatus";
        List<UtmLogstashPipeline> activeByServer = utmLogstashPipelineRepository.allActivePipelinesByServer();

        // Checking if LOGSTASH_URL is set, otherwise report an error
        if (!StringUtils.hasText(LOGSTASH_URL)) {
            log.error(ctx + ": The pipeline's status cannot be processed because the environment variable " +
                "LOGSTASH_URL is not configured.");
            activeByServer.stream().forEach((p) -> {
                p.setPipelineStatus(PipelineStatus.PIPELINE_STATUS_DOWN.get());
            });
            utmLogstashPipelineRepository.saveAll(activeByServer);
        } else {
            try {

                LogstashApiPipelineResponse response = pipelineApiResponse();
                Map<String, Long> mapInit = activeByServer.stream()
                    .collect(Collectors.toMap(UtmLogstashPipeline::getPipelineId, myId -> (
                        getFailures(myId, response)
                    )));

                Thread.sleep(2000);

                LogstashApiPipelineResponse responseLast = pipelineApiResponse();
                Map<String, PipelineData> pipelineInfo = responseLast.getPipelines();

                activeByServer.stream().forEach((p) -> {
                    // Getting stats from logstash and updating DB pipeline
                    PipelineData data = pipelineInfo.get(p.getPipelineId());
                    if (data != null) {
                        p.setEventsIn(data.getEvents().getIn());
                        p.setEventsFiltered(data.getEvents().getFiltered());
                        p.setEventsOut(data.getEvents().getOut());

                        PipelineReloads reloads = data.getReloads();
                        p.setReloadsFailures(reloads.getFailures());
                        p.setReloadsSuccesses(reloads.getSuccesses());
                        if (reloads.getLastError()!=null) p.setReloadsLastError(reloads.getLastError().getMessage());
                        p.setReloadsLastFailureTimestamp(reloads.getLastFailureTimestamp());
                        p.setReloadsLastSuccessTimestamp(reloads.getLastSuccessTimestamp());
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
        }
        return activeByServer;
    }

    // Method to count the failures by pipeline
    private Long getFailures(UtmLogstashPipeline e, LogstashApiPipelineResponse response) {
        Map<String, PipelineData> pipelines = response.getPipelines();
        PipelineData pipData = pipelines.get(e.getPipelineId());
        if (pipData != null) {
            return pipData.getReloads().getFailures();
        }
        return -1L;
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
            List<InputConfiguration> inputs = utmLogstashPipelineVM.getInputs();

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
                // Input validations
                if (inputs == null || inputs.isEmpty()) {
                    Validation val = new Validation("Input", "Input id", "There is no input associated to the pipeline: " + (utmLogstashPipeline != null ? utmLogstashPipeline.getPipelineName() : "Undefined pipeline"));
                    validationList.add(val);
                } else {
                    List<String> configs = new ArrayList<>();
                    inputs.forEach(i -> {
                        Long localId = i.getId();
                        Integer pipelineId = i.getPipelineId();
                        if (pipelineId != null && !utmLogstashPipelineRepository.findById(pipelineId.longValue()).isPresent() && MODE.equals(PipelineValidationMode.UPDATE)) {
                            Validation val = new Validation("Input", "Input pipeline id", "The pipeline with id: " + pipelineId + " not exists");
                            validationList.add(val);
                        }
                        if (pipelineId != null && utmLogstashPipeline.getId() != null && pipelineId != utmLogstashPipeline.getId().intValue() && MODE.equals(PipelineValidationMode.UPDATE)) {
                            Validation val = new Validation("Input", "Input pipeline id", "Value: " + pipelineId.intValue() + " from Input not match with value: " + utmLogstashPipeline.getId() + " from Pipeline");
                            validationList.add(val);
                        }
                        if (localId != null && !utmLogstashInputService.findOne(localId).isPresent()) {
                            Validation val = new Validation("Input", "Input id", "The input with id: " + localId + " not exists");
                            validationList.add(val);
                        }
                        if (localId != null && MODE.equals(PipelineValidationMode.INSERT)) {
                            Validation val = new Validation("Input", "Input id", "Value must be null when inserting");
                            validationList.add(val);
                        }
                        if (i.getConfigs() == null || i.getConfigs().isEmpty()) {
                            Validation val = new Validation("Input Configuration", "Input configuration id", "There is no input configuration associated to the input: " + i.getInputPrettyName());
                            validationList.add(val);
                        } else {
                            // Input configuration validations
                            i.getConfigs().forEach(c -> {
                                if ((c.getConfKey() != null && c.getConfKey().compareTo("") != 0)
                                    && (c.getConfType() != null && c.getConfType().compareTo("") != 0)
                                    && (c.getConfValue() != null && c.getConfValue().compareTo("") != 0)) {
                                    String configKey = c.getConfKey().split("_")[0] + c.getConfType() + c.getConfValue();
                                    if (configs.stream().filter(s -> s.compareToIgnoreCase(configKey) == 0).findFirst().isPresent()) {
                                        Validation val = new Validation("Input Configuration", "Input configuration value", "The value: " + c.getConfValue() + " is duplicated");
                                        validationList.add(val);
                                    } else {
                                        configs.add(configKey);
                                    }
                                } else {
                                    if (c.getConfKey() == null || c.getConfKey().compareTo("") == 0) {
                                        Validation val = new Validation("Input Configuration", "Input configuration key", "The value is not defined");
                                        validationList.add(val);
                                    }
                                    if (c.getConfType() == null || c.getConfType().compareTo("") == 0) {
                                        Validation val = new Validation("Input Configuration", "Input configuration type", "The value is not defined");
                                        validationList.add(val);
                                    }
                                    if (c.getConfValue() == null || c.getConfValue().compareTo("") == 0) {
                                        Validation val = new Validation("Input Configuration", "Input configuration value", "The value is not defined");
                                        validationList.add(val);
                                    }
                                }
                                if (c.getInputId() != null && MODE.equals(PipelineValidationMode.INSERT)) {
                                    Validation val = new Validation("Input Configuration", "Input configuration input id", "Value must be null when inserting");
                                    validationList.add(val);
                                }
                                if (c.getInputId() != null && localId.intValue() != c.getInputId()) {
                                    Validation val = new Validation("Input Configuration", "Input configuration input id", "Value: " + c.getInputId() + " in configuration not match with value: " + localId.intValue() + " from Input");
                                    validationList.add(val);
                                }
                            });
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
}

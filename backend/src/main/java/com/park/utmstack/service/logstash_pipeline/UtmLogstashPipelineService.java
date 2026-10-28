package com.park.utmstack.service.logstash_pipeline;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.logstash_filter.UtmLogstashFilter;
import com.park.utmstack.domain.logstash_pipeline.UtmGroupLogstashPipelineFilters;
import com.park.utmstack.domain.logstash_pipeline.UtmLogstashPipeline;
import com.park.utmstack.domain.logstash_pipeline.enums.PipelineValidationMode;
import com.park.utmstack.domain.logstash_pipeline.types.*;
import com.park.utmstack.repository.logstash_filter.UtmLogstashFilterRepository;
import com.park.utmstack.repository.logstash_pipeline.UtmGroupLogstashPipelineFiltersRepository;
import com.park.utmstack.repository.logstash_pipeline.UtmLogstashPipelineRepository;
import com.park.utmstack.service.dto.logstash_pipeline.UtmLogstashPipelineDTO;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import com.park.utmstack.service.elasticsearch.OpensearchClientBuilder;
import com.park.utmstack.service.logstash_pipeline.enums.PipelineStatus;
import com.park.utmstack.service.logstash_pipeline.response.ApiPipelineResponse;
import com.park.utmstack.service.logstash_pipeline.response.ApiStatsResponse;
import com.park.utmstack.service.logstash_pipeline.response.engine.ApiEngineResponse;
import com.park.utmstack.service.logstash_pipeline.response.pipeline.PipelineData;
import com.park.utmstack.service.logstash_pipeline.response.pipeline.PipelineStats;
import com.park.utmstack.service.logstash_pipeline.response.statistic.StatisticDocument;
import com.park.utmstack.service.web_clients.rest_template.RestTemplateService;
import com.park.utmstack.util.exceptions.ApiException;
import com.park.utmstack.web.rest.vm.UtmLogstashPipelineVM;
import com.utmstack.opensearch_connector.parsers.TermAggregateParser;
import com.utmstack.opensearch_connector.types.BucketAggregation;
import org.elasticsearch.search.aggregations.Aggregations;
import org.opensearch.client.json.JsonData;
import org.opensearch.client.opensearch._types.FieldValue;
import org.opensearch.client.opensearch._types.aggregations.Aggregate;
import org.opensearch.client.opensearch._types.aggregations.Aggregation;
import org.opensearch.client.opensearch._types.aggregations.TermsAggregation;
import org.opensearch.client.opensearch._types.aggregations.TopHitsAggregate;
import org.opensearch.client.opensearch.core.SearchRequest;
import org.opensearch.client.opensearch.core.SearchResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.support.PagedListHolder;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.support.PageableExecutionUtils;
import org.springframework.http.HttpStatus;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.StringUtils;

import javax.sql.DataSource;
import java.time.Duration;
import java.time.Instant;
import java.time.format.DateTimeFormatter;
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

    private final ElasticsearchService elasticsearchService;

    private final DateTimeFormatter formatter = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");
    private final DataSource dataSource;

    public UtmLogstashPipelineService(UtmLogstashPipelineRepository utmLogstashPipelineRepository, RestTemplateService restTemplateService, UtmGroupLogstashPipelineFiltersRepository utmGroupLogstashPipelineFiltersRepository, UtmLogstashFilterRepository utmLogstashFilterRepository, OpensearchClientBuilder client, ElasticsearchService elasticsearchService, DataSource dataSource) {
        this.utmLogstashPipelineRepository = utmLogstashPipelineRepository;
        this.restTemplateService = restTemplateService;
        this.utmGroupLogstashPipelineFiltersRepository = utmGroupLogstashPipelineFiltersRepository;
        this.utmLogstashFilterRepository = utmLogstashFilterRepository;
        this.elasticsearchService = elasticsearchService;
        this.dataSource = dataSource;
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

    try {
        ApiStatsResponse statsResponse = new ApiStatsResponse();
        boolean isCorrelationUp = isEngineUp();

        ApiEngineResponse jvmData = logstashJvmApiResponse();
        if (jvmData != null) {
            statsResponse.setGeneral(jvmData);
        }

        List<UtmLogstashPipeline> activePipelines = activePipelinesList();

        if (!isCorrelationUp) {
            activePipelines.forEach(p ->
                p.setPipelineStatus(PipelineStatus.PIPELINE_STATUS_DOWN.get())
            );
        }

        List<PipelineStats> pipelineStatsList = activePipelines.stream()
            .map(PipelineStats::getPipelineStats)
            .sorted( Comparator.comparing(PipelineStats::getPipelineStatus).reversed())
            .collect(Collectors.toList());

        statsResponse.setPipelines(pipelineStatsList);

        if (isCorrelationUp && jvmData != null) {
            long upCount = activePipelines.stream()
                .filter(p -> PipelineStatus.PIPELINE_STATUS_UP.get()
                    .equals(p.getPipelineStatus()))
                .count();

            int totalCount = activePipelines.size();

            if (upCount == 0) {
                jvmData.setStatus(PipelineStatus.ENGINE_STATUS_RED.get());
            } else if (upCount == totalCount) {
                jvmData.setStatus(PipelineStatus.ENGINE_STATUS_GREEN.get());
            } else {
                jvmData.setStatus(PipelineStatus.ENGINE_STATUS_YELLOW.get());
            }
        }

        return statsResponse;

    } catch (Exception ex) {
        log.error("{}: An error occurred while fetching logstash stats: {}", ctx, ex.getMessage(), ex);
        throw new ApiException(String.format("%s: An error occurred while fetching logstash stats", ctx), HttpStatus.INTERNAL_SERVER_ERROR);
    }
}

    /**
     * Method to set the DB pipelines status
     */
    @Scheduled(fixedDelay = 20000, initialDelay = 30000)
    public void pipelineStatus() {
        final String ctx = CLASSNAME + ".pipelineStatus";
        List<UtmLogstashPipeline> activeByServer = utmLogstashPipelineRepository.allActivePipelinesByServer();

        try {
            activeByServer.forEach((p) -> {
                // Evaluate ALL the module's data types. The previous code took
                // findFirst() on a HashSet, which is an arbitrary order, so
                // multi-filter modules were judged by an unrelated data type
                // (#2985).
                Set<UtmLogstashFilter> filters = p.getUtmModule() != null ? p.getUtmModule().getFilters() : null;
                List<String> dataTypes = Optional.ofNullable(filters)
                        .map(fs -> fs.stream()
                                .map(f -> f.getDatatype().getDataType())
                                .filter(StringUtils::hasText)
                                .distinct()
                                .collect(Collectors.toList()))
                        .filter(l -> !l.isEmpty())
                        .orElse(List.of(p.getPipelineName()));

                StatisticDocument stats = this.getStatisticsDataType(Constants.STATISTICS_INDEX_PATTERN, dataTypes);

                // "Processed events" is the sum of this pipeline's OWN data
                // over the last 24 h. When it has none, reset to 0 instead of
                // keeping the stale stored value (#2985).
                p.setEventsOut(stats != null && stats.getCount() != null ? stats.getCount() : 0L);

                if (stats == null || !StringUtils.hasText(stats.getTimestamp())) {
                    p.setPipelineStatus(PipelineStatus.PIPELINE_STATUS_DOWN.get());
                } else {
                    Instant timestampDate = Instant.parse(stats.getTimestamp());
                    long hoursDifference = Duration.between(timestampDate, Instant.now()).toHours();
                    if (hoursDifference > 6) {
                        p.setPipelineStatus(PipelineStatus.PIPELINE_STATUS_DOWN.get());
                    } else {
                        p.setPipelineStatus(PipelineStatus.PIPELINE_STATUS_UP.get());
                    }
                }
            });
            utmLogstashPipelineRepository.saveAll(activeByServer);

        } catch (Exception connectException) {
            String msg = ctx + ": " + connectException.getMessage();
            log.error(msg);
        }
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

    public StatisticDocument getStatisticsDataType(String indexName, List<String> dataTypeValues) {
        // Exact match on the keyword sub-field for every data type the
        // pipeline can produce, plus a 24 h window, and a sum of count. A
        // `match` on the text field split the value into tokens (e.g.
        // "antivirus-kaspersky" -> "antivirus" OR "kaspersky") and returned
        // another vendor's statistics; the sum replaces the "count of the
        // single newest document" (#2985).
        final String windowGte = DateTimeFormatter.ISO_INSTANT.format(Instant.now().minus(Duration.ofHours(24)));
        SearchRequest sr = SearchRequest.of(s -> s
                .index(indexName)
                .query(q -> q.bool(b -> b
                        .filter(f -> f.terms(t -> t.field(Constants.dataTypeKeyword)
                                .terms(tv -> tv.value(dataTypeValues.stream()
                                        .map(FieldValue::of)
                                        .collect(Collectors.toList())))))
                        .filter(f -> f.range(r -> r.field("@timestamp").gte(JsonData.of(windowGte))))))
                .aggregations("sum_count", a -> a.sum(sm -> sm.field("count")))
                .aggregations("max_ts", a -> a.max(mx -> mx.field("@timestamp")))
                .size(0)
        );

        SearchResponse<StatisticDocument> response = elasticsearchService.search(sr, StatisticDocument.class);

        Double sum = response.aggregations().get("sum_count").sum().value();
        Double maxTs = response.aggregations().get("max_ts").max().value();
        if (sum == null || maxTs == null || sum == 0.0) {
            return null;
        }
        StatisticDocument doc = new StatisticDocument();
        doc.setCount(Math.round(sum));
        doc.setTimestamp(DateTimeFormatter.ISO_INSTANT.format(Instant.ofEpochMilli(maxTs.longValue())));
        return doc;
    }
}

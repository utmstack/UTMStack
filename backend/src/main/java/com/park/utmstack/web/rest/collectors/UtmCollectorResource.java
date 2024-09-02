package com.park.utmstack.web.rest.collectors;

import agent.CollectorOuterClass.CollectorConfig;
import agent.Common.ListRequest;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.application_modules.UtmModuleGroup;
import com.park.utmstack.domain.network_scan.AssetGroupFilter;
import com.park.utmstack.domain.network_scan.NetworkScanFilter;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.application_modules.UtmModuleGroupConfigurationService;
import com.park.utmstack.service.application_modules.UtmModuleGroupService;
import com.park.utmstack.service.collectors.CollectorOpsService;
import com.park.utmstack.service.collectors.UtmCollectorService;
import com.park.utmstack.service.dto.collectors.CollectorActionEnum;
import com.park.utmstack.service.dto.collectors.CollectorHostnames;
import com.park.utmstack.service.dto.collectors.dto.CollectorConfigKeysDTO;
import com.park.utmstack.service.dto.collectors.dto.CollectorDTO;
import com.park.utmstack.service.dto.collectors.CollectorModuleEnum;
import com.park.utmstack.service.dto.collectors.dto.ErrorResponse;
import com.park.utmstack.service.dto.collectors.dto.ListCollectorsResponseDTO;
import com.park.utmstack.service.dto.network_scan.AssetGroupDTO;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.park.utmstack.web.rest.errors.InternalServerErrorException;
import com.park.utmstack.web.rest.network_scan.UtmNetworkScanResource;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import com.utmstack.grpc.exception.CollectorConfigurationGrpcException;
import com.utmstack.grpc.exception.CollectorServiceGrpcException;
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

import javax.validation.Valid;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;


/**
 * REST controller for managing {@link UtmCollectorResource}.
 */
@RestController
@RequestMapping("/api")
public class UtmCollectorResource {

    private static final String CLASSNAME = "UtmCollectorResource";
    private final CollectorOpsService collectorService;
    private final Logger log = LoggerFactory.getLogger(UtmCollectorResource.class);
    private final ApplicationEventService applicationEventService;
    private final UtmModuleGroupConfigurationService moduleGroupConfigurationService;

    private final UtmModuleGroupService moduleGroupService;

    private final ApplicationEventService eventService;

    private final UtmCollectorService utmCollectorService;

    public UtmCollectorResource(CollectorOpsService collectorService,
                                ApplicationEventService applicationEventService,
                                UtmModuleGroupConfigurationService moduleGroupConfigurationService,
                                UtmModuleGroupService moduleGroupService,
                                ApplicationEventService eventService,
                                UtmCollectorService utmCollectorService) {

        this.collectorService = collectorService;
        this.applicationEventService = applicationEventService;
        this.moduleGroupConfigurationService = moduleGroupConfigurationService;
        this.moduleGroupService = moduleGroupService;
        this.eventService = eventService;
        this.utmCollectorService = utmCollectorService;
    }

    /**
     * {@code POST  /collector-config} : Create or update the collector configs.
     *
     * @param collectorConfig the collector configs to be created/updated in the agent manager and updated in database.
     * @return the {@link ResponseEntity} with status {@code 204 (No Content)}, status {@code 400 (Bad request)} if the internal key is not set,
     * status {@code 502 (Bad Gateway)} if the agent manager returns an error, or with status {@code 500 (Internal Server Error)} if the database couldn't
     * persist the configurations.
     */
    @PostMapping("/collector-config")
    public ResponseEntity<Void> upsertCollectorConfig(
            @Valid @RequestBody CollectorConfigKeysDTO collectorConfig,
            @RequestParam(name = "action", defaultValue = "CREATE") CollectorActionEnum action) {

        final String ctx = CLASSNAME + ".upsertCollectorConfig";
        CollectorConfig cacheConfig = null;

        // Validate collector configuration
        String validationErrorMessage = this.collectorService.validateCollectorConfig(collectorConfig);
        if (validationErrorMessage != null) {
            return logAndResponse(new ErrorResponse(validationErrorMessage, HttpStatus.PRECONDITION_FAILED));
        }

        try {
            cacheConfig = this.collectorService.cacheCurrentCollectorConfig(collectorConfig.getCollector());
            this.upsert(collectorConfig);
            return ResponseEntity.noContent().build();

        } catch (Exception e) {
            return handleUpdateError(e, cacheConfig, collectorConfig.getCollector());
        }
    }

    /**
     * {@code GET  /collectors-list} : Get all collectors list by module.
     *
     * @param pageNumber the page number to show results from.
     * @param pageSize   the number of items to show in the page.
     * @param module     the module used to filter the collectors list. If no value is set, returns collectors by all modules
     * @param sortBy     the criteria to sort the results.
     * @return the {@link ResponseEntity} with status {@code 204 (No Content)}, status {@code 400 (Bad request)} if the internal key is not set,
     * or with status {@code 502 (Bad Gateway)} if the agent manager returns an error.
     */
    @GetMapping("/collectors-list")
    public ResponseEntity<ListCollectorsResponseDTO> listCollectorsByModule(@RequestParam(required = false) Integer pageNumber,
                                                                            @RequestParam(required = false) Integer pageSize,
                                                                            @RequestParam(required = false) CollectorModuleEnum module,
                                                                            @RequestParam(required = false) String sortBy) {
        final String ctx = CLASSNAME + ".listCollectorsByModule";
        try {
            ListRequest request = ListRequest.newBuilder()
                    .setPageNumber(pageNumber != null ? pageNumber : 0)
                    .setPageSize(pageSize != null ? pageSize : 1000000)
                    .setSearchQuery(module != null ? "module.Is=" + module.name() : "")
                    .setSortBy(sortBy != null ? sortBy : "")
                    .build();

            ListCollectorsResponseDTO response = collectorService.listCollector(request);
            HttpHeaders headers = new HttpHeaders();
            headers.add("X-Total-Count", Long.toString(response.getTotal()));
            return ResponseEntity.ok().headers(headers).body(response);
        } catch (BadRequestAlertException e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.BAD_REQUEST, msg);
        } catch (CollectorServiceGrpcException e) {
            String msg = ctx + ": UtmCollector manager is not available or was an error getting the collector list. " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.BAD_GATEWAY, msg);
        }
    }

    /**
     * {@code GET  /collector-hostnames} : Get all collector hostnames by module.
     *
     * @param pageNumber the page number to show results from.
     * @param pageSize   the number of items to show in the page.
     * @param module     the module used to filter the collectors list. If no value is set, returns collectors by all modules
     * @param sortBy     the criteria to sort the results.
     * @return the {@link ResponseEntity} with status {@code 200 (Ok)}, status {@code 400 (Bad request)} if the internal key is not set,
     * or with status {@code 502 (Bad Gateway)} if the agent manager returns an error.
     */
    @GetMapping("/collector-hostnames")
    public ResponseEntity<CollectorHostnames> listCollectorHostNames(@RequestParam(required = false) Integer pageNumber,
                                                                     @RequestParam(required = false) Integer pageSize,
                                                                     @RequestParam(required = false) CollectorModuleEnum module,
                                                                     @RequestParam(required = false) String sortBy) {
        final String ctx = CLASSNAME + ".listCollectorHostNames";
        try {
            ListRequest request = ListRequest.newBuilder()
                    .setPageNumber(pageNumber != null ? pageNumber : 0)
                    .setPageSize(pageSize != null ? pageSize : 1000000)
                    .setSearchQuery(module != null ? "module.Is=" + module : "")
                    .setSortBy(sortBy != null ? sortBy : "")
                    .build();
            return ResponseEntity.ok().body(collectorService.listCollectorHostnames(request));
        } catch (BadRequestAlertException e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.BAD_REQUEST, msg);
        } catch (CollectorServiceGrpcException e) {
            String msg = ctx + ": UtmCollector manager is not available or the parameters are wrong, please check." + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.BAD_GATEWAY, msg);
        }
    }

    /**
     * {@code GET  /collector-by-hostname-and-module} : Get collector's list according to the request params.
     *
     * @param hostname the host name to search for.
     * @param module   the collector module to search for
     * @return the {@link ResponseEntity} with status {@code 204 (No Content)}, status {@code 400 (Bad request)} if the internal key is not set,
     * or with status {@code 502 (Bad Gateway)} if the agent manager returns an error.
     */
    @GetMapping("/collector-by-hostname-and-module")
    public ResponseEntity<ListCollectorsResponseDTO> listCollectorByHostNameAndModule(@RequestParam String hostname,
                                                                                      @RequestParam CollectorModuleEnum module) {
        final String ctx = CLASSNAME + ".listCollectorByHostNameAndModule";
        try {
            return ResponseEntity.ok().body(collectorService.listCollector(
                    collectorService.getListRequestByHostnameAndModule(hostname, module)));
        } catch (BadRequestAlertException e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.BAD_REQUEST, msg);
        } catch (CollectorServiceGrpcException e) {
            String msg = ctx + ": UtmCollector manager is not available or was an error getting configuration. " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.BAD_GATEWAY, msg);
        }
    }

    @GetMapping("/groups-by-collectors/{collectorId}")
    public ResponseEntity<List<UtmModuleGroup>> getModuleGroups(@PathVariable String collectorId) {
        final String ctx = CLASSNAME + ".getModuleGroups";
        try {
            return ResponseEntity.ok(moduleGroupService.findAllByCollectorId(collectorId));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @PutMapping("/updateGroup")
    public ResponseEntity<Void> updateGroup(@Valid @RequestBody UtmNetworkScanResource.UpdateGroupRequestBody body) {
        final String ctx = CLASSNAME + ".updateGroup";
        try {
            collectorService.updateGroup(body.getAssetsIds(), body.getAssetGroupId());
            return ResponseEntity.ok().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/searchGroupsByFilter")
    public ResponseEntity<List<AssetGroupDTO>> searchGroupsByFilter(AssetGroupFilter filter, Pageable pageable) {
        final String ctx = CLASSNAME + ".searchGroupsByFilter";
        try {

            Page<AssetGroupDTO> page = collectorService.searchGroupsByFilter(filter, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/utm-asset-groups/searchGroupsByFilter");
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/search-by-filters")
    public ResponseEntity<List<CollectorDTO>> searchByFilters(@ParameterObject NetworkScanFilter filters,
                                                              @ParameterObject Pageable pageable) {
        final String ctx = CLASSNAME + ".searchByFilters";
        try {
            collectorService.listCollector(ListRequest.newBuilder()
                    .setPageNumber(0)
                    .setPageSize(1000000)
                    .setSearchQuery("module.Is=" + CollectorModuleEnum.AS_400)
                    .setSortBy("")
                    .build());
            Page<CollectorDTO> page = this.utmCollectorService.searchByFilters(filters, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/search-by-filters");
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @DeleteMapping("/collectors/{id}")
    public ResponseEntity<Void> deleteCollector(@PathVariable Long id) {

        try {
            log.debug("REST request to delete UtmCollector : {}", id);
            collectorService.deleteCollector(id);
            return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert("UtmCollector", id.toString())).build();
        } catch (Exception e) {
            applicationEventService.createEvent(e.getMessage(), ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert("UtmCollector", null, e.getMessage())).body(null);
        }
    }

    @PostMapping("/collectors-config")
    public ResponseEntity<Map<String, Object>> upsertCollectorsConfig(@RequestBody List<CollectorConfigKeysDTO> collectors) {
        Map<String, Object> results = new HashMap<>();
        final String ctx = CLASSNAME + ".upsertCollectorsConfig";
        CollectorConfig cacheConfig = null;

        List<Map<String, Object>> collectorsResults = new ArrayList<>();
        for (CollectorConfigKeysDTO collectorConfig : collectors) {
            Map<String, Object> collectorResult = new HashMap<>();
            collectorResult.put("collectorId", collectorConfig.getCollector().getId());
            try {
                cacheConfig = this.collectorService.cacheCurrentCollectorConfig(collectorConfig.getCollector());
                this.upsert(collectorConfig);
                collectorResult.put("status", "success");
            } catch (Exception e) {
                ErrorResponse error = this.getError(e, cacheConfig);
                collectorResult.put("status", "failure");
                collectorResult.put("errorMessage", error.getMessage());
            }
            collectorsResults.add(collectorResult);
        }

        results.put("results", collectorsResults);
        return ResponseEntity.status(HttpStatus.MULTI_STATUS).body(results);
    }

    private ResponseEntity<Void> handleUpdateError(Exception e, CollectorConfig cacheConfig, CollectorDTO collectorDTO) {
        return logAndResponse(this.getError(e, cacheConfig));
    }

    private ErrorResponse getError(Exception e, CollectorConfig cacheConfig) {
        String msg;
        HttpStatus status;

        try {
            if (e instanceof InternalServerErrorException) {
                /*collectorService.upsertCollectorConfig(cacheConfig);*/
                msg = "The collector configuration couldn't be persisted on database: " + e.getLocalizedMessage();
                status = HttpStatus.INTERNAL_SERVER_ERROR;

            } else if (e instanceof CollectorConfigurationGrpcException || e instanceof CollectorServiceGrpcException) {
                msg = "UtmCollector manager is not available or the configuration is wrong: " + e.getLocalizedMessage();
                status = HttpStatus.BAD_GATEWAY;

            } else if (e instanceof BadRequestAlertException) {
                msg = e.getLocalizedMessage();
                status = HttpStatus.BAD_REQUEST;

            } else {
                msg = "Unexpected error: " + e.getLocalizedMessage();
                status = HttpStatus.INTERNAL_SERVER_ERROR;
            }

        } catch (Exception rollbackException) {
            msg = "Failed to rollback the configuration: " + rollbackException.getLocalizedMessage();
            status = HttpStatus.INTERNAL_SERVER_ERROR;
        }

        return new ErrorResponse(msg, status);
    }

    private ResponseEntity<Void> logAndResponse(ErrorResponse error) {
        log.error(error.getMessage());
        applicationEventService.createEvent(error.getMessage(), ApplicationEventType.ERROR);
        return UtilResponse.buildErrorResponse(error.getStatus(), error.getMessage());
    }

    private void upsert(CollectorConfigKeysDTO collectorConfig) throws Exception {

        // Update local database with new configuration
        this.collectorService.updateCollectorConfigurationKeys(collectorConfig);

        // Attempt to update collector configuration via gRPC
        this.collectorService.updateCollectorConfigViaGrpc(collectorConfig, collectorConfig.getCollector());
    }
}

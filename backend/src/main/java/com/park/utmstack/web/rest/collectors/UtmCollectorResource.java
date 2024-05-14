package com.park.utmstack.web.rest.collectors;

import agent.CollectorOuterClass.CollectorModule;
import agent.CollectorOuterClass.FilterByHostAndModule;
import agent.CollectorOuterClass.CollectorHostnames;
import agent.CollectorOuterClass.ConfigRequest;
import agent.CollectorOuterClass.CollectorConfig;
import agent.Common.ListRequest;
import agent.Common.AuthResponse;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.application_modules.UtmModuleGroupConfigurationService;
import com.park.utmstack.service.collectors.CollectorOpsService;
import com.park.utmstack.service.dto.collectors.CollectorDTO;
import com.park.utmstack.service.dto.collectors.CollectorModuleEnum;
import com.park.utmstack.service.dto.collectors.ListCollectorsResponseDTO;
import com.park.utmstack.service.validators.collector.CollectorValidatorService;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.web.rest.application_modules.UtmModuleGroupConfigurationResource;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.park.utmstack.web.rest.errors.InternalServerErrorException;
import com.utmstack.grpc.exception.CollectorConfigurationGrpcException;
import com.utmstack.grpc.exception.CollectorServiceGrpcException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.validation.BeanPropertyBindingResult;
import org.springframework.validation.Errors;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;


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

    private final CollectorValidatorService collectorValidatorService;

    public UtmCollectorResource(CollectorOpsService collectorService, ApplicationEventService applicationEventService, UtmModuleGroupConfigurationService moduleGroupConfigurationService, CollectorValidatorService collectorValidatorService) {
        this.collectorService = collectorService;
        this.applicationEventService = applicationEventService;
        this.moduleGroupConfigurationService = moduleGroupConfigurationService;
        this.collectorValidatorService = collectorValidatorService;
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
            @Valid @RequestBody UtmModuleGroupConfigurationResource.UpdateConfigurationKeysBody collectorConfig,
            CollectorDTO collectorDTO) {
        final String ctx = CLASSNAME + ".upsertCollectorConfig";

        try {
            Errors errors = new BeanPropertyBindingResult(collectorConfig, "updateConfigurationKeysBody");
            collectorValidatorService.validate(collectorConfig, errors);

            if (errors.hasErrors()) {
                String msg =  "Validation failed: Hostname must be unique for this collector.";
                log.error(msg);
                applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
                return UtilResponse.buildPreconditionFailedResponse(msg);
            }

            // Get the actual configuration just in case of error when updating local db.
            CollectorConfig cacheConfig = collectorService.getCollectorConfig(
                    ConfigRequest.newBuilder().setModule(CollectorModule.valueOf(collectorDTO.getModule().toString())).build(),
                    AuthResponse.newBuilder().setId(collectorDTO.getId()).setKey(collectorDTO.getCollectorKey()).build());
            // Map the configurations to gRPC CollectorConfig and try to insert/update the collector config
            collectorService.upsertCollectorConfig(collectorService.mapToCollectorConfig(
                    collectorService.mapPasswordConfiguration(collectorConfig.getKeys()), collectorDTO));
            // If the update is fine via gRPC, then update the configurations in local db.
            try {
                moduleGroupConfigurationService.updateConfigurationKeys(collectorConfig.getModuleId(), collectorConfig.getKeys());
                return ResponseEntity.noContent().build();
            } catch (Exception e) {
                String msg = ctx + ": " + e.getMessage();
                // In case of db error try to reset the configuration from the cache
                collectorService.upsertCollectorConfig(cacheConfig);
                throw new InternalServerErrorException(msg);
            }
        } catch (BadRequestAlertException e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.BAD_REQUEST, msg);
        } catch (CollectorConfigurationGrpcException | CollectorServiceGrpcException e) {
            String msg = ctx + ": Collector manager is not available or the configuration is wrong, please check. " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.BAD_GATEWAY, msg);
        } catch (InternalServerErrorException e) {
            String msg = ctx + ": The collector configuration couldn't be persisted on database. " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * {@code GET  /collectors-list} : Get all collectors list by module.
     *
     * @param pageNumber the page number to show results from.
     * @param pageSize the number of items to show in the page.
     * @param module the module used to filter the collectors list. If no value is set, returns collectors by all modules
     * @param sortBy the criteria to sort the results.
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
                    .setPageSize(pageSize != null ? pageSize : 100)
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
            String msg = ctx + ": Collector manager is not available or was an error getting the collector list. " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.BAD_GATEWAY, msg);
        }
    }

    /**
     * {@code GET  /collector-hostnames} : Get all collector hostnames by module.
     *
     * @param pageNumber the page number to show results from.
     * @param pageSize the number of items to show in the page.
     * @param module the module used to filter the collectors list. If no value is set, returns collectors by all modules
     * @param sortBy the criteria to sort the results.
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
                    .setPageSize(pageSize != null ? pageSize : 0)
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
            String msg = ctx + ": Collector manager is not available or the parameters are wrong, please check." + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.BAD_GATEWAY, msg);
        }
    }

    /**
     * {@code GET  /collector-by-hostname-and-module} : Get collector's list according to the request params.
     *
     * @param hostname the host name to search for.
     * @param module the collector module to search for
     * @return the {@link ResponseEntity} with status {@code 204 (No Content)}, status {@code 400 (Bad request)} if the internal key is not set,
     * or with status {@code 502 (Bad Gateway)} if the agent manager returns an error.
     */
    @GetMapping("/collector-by-hostname-and-module")
    public ResponseEntity<ListCollectorsResponseDTO> listCollectorByHostNameAndModule(@RequestParam String hostname,
                                                                                  @RequestParam CollectorModuleEnum module) {
        final String ctx = CLASSNAME + ".listCollectorByHostNameAndModule";
        try {
            FilterByHostAndModule request = FilterByHostAndModule.newBuilder()
                    .setHostname(hostname)
                    .setModule(CollectorModule.valueOf(module.toString()))
                    .build();
            return ResponseEntity.ok().body(collectorService.getCollectorsByHostnameAndModule(request));
        } catch (BadRequestAlertException e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.BAD_REQUEST, msg);
        } catch (CollectorServiceGrpcException e) {
            String msg = ctx + ": Collector manager is not available or was an error getting configuration. " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.BAD_GATEWAY, msg);
        }
    }
}

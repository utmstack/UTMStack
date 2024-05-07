package com.park.utmstack.service.collectors;

import agent.CollectorOuterClass.CollectorGroupConfigurations;
import agent.CollectorOuterClass.CollectorDelete;
import agent.CollectorOuterClass.ListCollectorResponse;
import agent.CollectorOuterClass.ConfigKnowledge;
import agent.CollectorOuterClass.CollectorConfig;
import agent.CollectorOuterClass.CollectorHostnames;
import agent.CollectorOuterClass.FilterByHostAndModule;
import agent.CollectorOuterClass.Collector;
import agent.CollectorOuterClass.CollectorModule;
import agent.CollectorOuterClass.CollectorConfigGroup;
import agent.CollectorOuterClass.ConfigRequest;
import agent.Common;
import agent.Common.ListRequest;
import agent.Common.AuthResponse;
import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.application_modules.UtmModuleGroup;
import com.park.utmstack.domain.application_modules.UtmModuleGroupConfiguration;
import com.park.utmstack.security.SecurityUtils;
import com.park.utmstack.service.application_modules.UtmModuleGroupService;
import com.park.utmstack.service.dto.collectors.CollectorModuleEnum;
import com.park.utmstack.service.dto.collectors.ListCollectorsResponseDTO;
import com.park.utmstack.service.dto.collectors.CollectorDTO;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.utmstack.grpc.connection.GrpcConnection;
import com.utmstack.grpc.exception.CollectorConfigurationGrpcException;
import com.utmstack.grpc.exception.CollectorServiceGrpcException;
import com.utmstack.grpc.exception.GrpcConnectionException;
import com.utmstack.grpc.service.CollectorService;
import com.utmstack.grpc.service.PanelCollectorService;
import io.grpc.*;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.ArrayList;
import java.util.List;
import java.util.Objects;
import java.util.Optional;
import java.util.stream.Collectors;

@Service
public class CollectorOpsService {
    private final String CLASSNAME = "CollectorOpsService";
    private final Logger log = LoggerFactory.getLogger(CollectorOpsService.class);
    private final GrpcConnection grpcConnection;
    private final PanelCollectorService panelCollectorService;
    private final CollectorService collectorService;
    private final UtmModuleGroupService moduleGroupService;

    public CollectorOpsService(GrpcConnection grpcConnection, UtmModuleGroupService moduleGroupService) throws GrpcConnectionException {
        this.grpcConnection = grpcConnection;
        this.panelCollectorService = new PanelCollectorService(grpcConnection);
        this.collectorService = new CollectorService(grpcConnection);
        this.moduleGroupService = moduleGroupService;
    }

    /**
     * Method to update a collector's configuration.
     *
     * @param config is the configuration of the collectors to update.
     * @throws CollectorConfigurationGrpcException if the action can't be performed.
     */
    public ConfigKnowledge upsertCollectorConfig(CollectorConfig config) throws CollectorConfigurationGrpcException {
        final String ctx = CLASSNAME + ".upsertCollectorConfig";

        String internalKey = System.getenv(Constants.ENV_INTERNAL_KEY);

        if (!StringUtils.hasText(internalKey)) {
            throw new BadRequestAlertException(ctx + ": Internal key not configured.", ctx, CLASSNAME);
        }

        try {
            return panelCollectorService.insertCollectorConfig(config, internalKey);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            throw new CollectorConfigurationGrpcException(msg);
        }
    }

    /**
     * Method to get collectors list.
     *
     * @param request is the request with all the pagination and search params used to list collectors
     *                according to those params.
     * @throws CollectorServiceGrpcException if the action can't be performed or the request is malformed.
     */
    public ListCollectorsResponseDTO listCollector(ListRequest request) throws CollectorServiceGrpcException {
        final String ctx = CLASSNAME + ".listCollector";

        String internalKey = System.getenv(Constants.ENV_INTERNAL_KEY);

        if (!StringUtils.hasText(internalKey)) {
            throw new BadRequestAlertException(ctx + ": Internal key not configured.", ctx, CLASSNAME);
        }

        try {
            return mapToListCollectorsResponseDTO(collectorService.listCollector(request, internalKey));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            throw new CollectorServiceGrpcException(msg);
        }
    }

    /**
     * Method to List all Collector's hostnames.
     *
     * @param request is the request with all the pagination and search params used to list collectors.
     *                according to those params.
     * @throws CollectorServiceGrpcException if the action can't be performed or the request is malformed.
     */
    public CollectorHostnames listCollectorHostnames(ListRequest request) throws CollectorServiceGrpcException {
        final String ctx = CLASSNAME + ".ListCollectorHostnames";

        String internalKey = System.getenv(Constants.ENV_INTERNAL_KEY);

        if (!StringUtils.hasText(internalKey)) {
            throw new BadRequestAlertException(ctx + ": Internal key not configured.", ctx, CLASSNAME);
        }

        try {
            return collectorService.ListCollectorHostnames(request, internalKey);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            throw new CollectorServiceGrpcException(msg);
        }
    }

    /**
     * Method to get collectors by hostname and module.
     *
     * @param request contains the filter information used to search.
     * @throws CollectorServiceGrpcException if the action can't be performed or the request is malformed.
     */
    public ListCollectorsResponseDTO getCollectorsByHostnameAndModule(FilterByHostAndModule request) throws CollectorServiceGrpcException {
        final String ctx = CLASSNAME + ".GetCollectorsByHostnameAndModule";

        String internalKey = System.getenv(Constants.ENV_INTERNAL_KEY);

        if (!StringUtils.hasText(internalKey)) {
            throw new BadRequestAlertException(ctx + ": Internal key not configured.", ctx, CLASSNAME);
        }

        try {
            return mapToListCollectorsResponseDTO(collectorService.GetCollectorsByHostnameAndModule(request, internalKey));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            throw new CollectorServiceGrpcException(msg);
        }
    }

    /**
     * Method to get a collector config from agent manager via gRPC.
     *
     * @param request represents the CollectorModule to get the configurations from.
     * @param auth is the authentication parameters used to filter in order to get the collector configuration.
     * @throws CollectorServiceGrpcException if the action can't be performed or the request is malformed.
     */
    public CollectorConfig getCollectorConfig(ConfigRequest request, AuthResponse auth) throws CollectorServiceGrpcException {
        final String ctx = CLASSNAME + ".getCollectorConfig";


        try {
            return collectorService.requestCollectorConfig(request, auth);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            throw new CollectorServiceGrpcException(msg);
        }
    }

    /**
     * Method to transform a ListCollectorResponse to ListCollectorsResponseDTO
     */
    private ListCollectorsResponseDTO mapToListCollectorsResponseDTO(ListCollectorResponse response) throws Exception {
        final String ctx = CLASSNAME + ".mapToListCollectorsResponseDTO";
        try {
            ListCollectorsResponseDTO dto = new ListCollectorsResponseDTO();

            List<CollectorDTO> collectorDTOS = response.getRowsList().stream()
                    .map(this::protoToCollectorDto)
                    .collect(Collectors.toList());

            dto.setCollectors(collectorDTOS);
            dto.setTotal(response.getTotal());

            return dto;
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Method to map from List<UtmModuleGroupConfiguration> to CollectorConfig
     */

    public CollectorConfig mapToCollectorConfig(List<UtmModuleGroupConfiguration> keys, CollectorDTO collectorDTO) {
        final String ctx = CLASSNAME + ".mapToCollectorConfig";

        // Get distinct groups id
        List<Long> groups = keys.stream().mapToLong(UtmModuleGroupConfiguration::getGroupId).distinct().boxed().collect(Collectors.toList());
        CollectorConfig collectorConfig;
        List<CollectorConfigGroup> collectorGroups = new ArrayList<>();

        // Find all configurations for each group and convert them to CollectorGroupConfigurations
        groups.forEach(g -> {
            Optional<UtmModuleGroup> moduleGroup = moduleGroupService.findOne(g);
            moduleGroup.ifPresent(utmModuleGroup -> collectorGroups.add(CollectorConfigGroup.newBuilder()
                    .setGroupName(utmModuleGroup.getGroupName())
                    .setGroupDescription(utmModuleGroup.getGroupDescription())
                    .addAllConfigurations(
                            // Adding all the CollectorGroupConfigurations of the current group
                            keys.stream().filter(f -> Objects.equals(g, f.getGroupId()))
                                    .map(this::mapToCollectorGroupConfigurations).collect(Collectors.toList())
                    )
                    .setCollectorId(collectorDTO.getId())
                    .build()));
        });
        // Creating the final CollectorConfig object
        collectorConfig = CollectorConfig.newBuilder()
                .setCollectorKey(collectorDTO.getCollector_key())
                .setRequestId(String.valueOf(System.currentTimeMillis()))
                .addAllGroups(collectorGroups)
                .build();
        return collectorConfig;
    }

    /**
     * Method to transform a Collector to CollectorDTO
     */
    private CollectorDTO protoToCollectorDto(Collector collector) {
        return new CollectorDTO(collector);
    }

    /**
     * Method to map from UtmModuleGroupConfiguration to CollectorGroupConfigurations
     */
    private CollectorGroupConfigurations mapToCollectorGroupConfigurations(UtmModuleGroupConfiguration moduleConfig) {
        return CollectorGroupConfigurations.newBuilder()
                .setConfKey(moduleConfig.getConfKey())
                .setConfName(moduleConfig.getConfName())
                .setConfDescription(moduleConfig.getConfDescription())
                .setConfDataType(moduleConfig.getConfDataType())
                .setConfValue(moduleConfig.getConfValue())
                .setConfRequired(moduleConfig.getConfRequired()).build();
    }

    /**
     * Method to remove a collector, will be used to remove in the UtmNetworkScanService
     *
     * @param hostname the hostname of the collector to remove
     * @param module   the module of the collector to remove
     */
    public void deleteCollector(String hostname, CollectorModuleEnum module) {
        final String ctx = CLASSNAME + ".deleteCollector";
        try {
            String currentUser = SecurityUtils.getCurrentUserLogin().orElseThrow(() -> new RuntimeException("No current user login"));
            FilterByHostAndModule request = FilterByHostAndModule.newBuilder()
                    .setHostname(hostname)
                    .setModule(CollectorModule.valueOf(module.toString()))
                    .build();
            Optional<CollectorDTO> collectorToSearch = getCollectorsByHostnameAndModule(request).getCollectors()
                    .stream().findFirst();
            try {
                if (collectorToSearch.isEmpty()) {
                    log.error(String.format("%1$s: Collector %2$s could not be deleted because no information was obtained from collector manager", ctx, hostname));
                    return;
                }
            } catch (StatusRuntimeException e) {
                if (e.getStatus().getCode() == Status.Code.NOT_FOUND) {
                    log.error(String.format("%1$s: Collector %2$s could not be deleted because was not found", ctx, hostname));
                    return;
                }
            }

            CollectorDelete collectorDelete = CollectorDelete.newBuilder().setDeletedBy(currentUser).build();
            AuthResponse auth = Common.AuthResponse.newBuilder()
                    .setId(collectorToSearch.get().getId())
                    .setKey(collectorToSearch.get().getCollector_key())
                    .build();
            collectorService.deleteCollector(collectorDelete, auth);

        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            throw new RuntimeException(msg);
        }
    }
}

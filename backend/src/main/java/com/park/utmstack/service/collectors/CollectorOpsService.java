package com.park.utmstack.service.collectors;

import agent.CollectorOuterClass.CollectorGroupConfigurations;
import agent.CollectorOuterClass.ListCollectorResponse;
import agent.CollectorOuterClass.ConfigKnowledge;
import agent.CollectorOuterClass.CollectorConfig;
import agent.CollectorOuterClass.Collector;
import agent.CollectorOuterClass.CollectorModule;
import agent.CollectorOuterClass.CollectorConfigGroup;
import agent.CollectorOuterClass.ConfigRequest;
import agent.Common;
import agent.Common.ListRequest;
import agent.Common.AuthResponse;
import agent.Common.DeleteRequest;
import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.application_modules.UtmModule;
import com.park.utmstack.domain.application_modules.UtmModuleGroup;
import com.park.utmstack.domain.application_modules.UtmModuleGroupConfiguration;
import com.park.utmstack.domain.collector.UtmCollector;
import com.park.utmstack.domain.network_scan.AssetGroupFilter;
import com.park.utmstack.domain.network_scan.UtmAssetGroup;
import com.park.utmstack.repository.UtmModuleGroupConfigurationRepository;
import com.park.utmstack.repository.UtmModuleGroupRepository;
import com.park.utmstack.repository.application_modules.UtmModuleRepository;
import com.park.utmstack.repository.collector.UtmCollectorRepository;
import com.park.utmstack.security.SecurityUtils;
import com.park.utmstack.service.application_modules.UtmModuleGroupService;
import com.park.utmstack.service.application_modules.UtmModuleService;
import com.park.utmstack.service.dto.collectors.CollectorHostnames;
import com.park.utmstack.service.dto.collectors.CollectorModuleEnum;
import com.park.utmstack.service.dto.collectors.dto.CollectorConfigKeysDTO;
import com.park.utmstack.service.dto.collectors.dto.ListCollectorsResponseDTO;
import com.park.utmstack.service.dto.collectors.dto.CollectorDTO;
import com.park.utmstack.service.dto.network_scan.AssetGroupDTO;
import com.park.utmstack.service.validators.collector.CollectorValidatorService;
import com.park.utmstack.util.CipherUtil;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.utmstack.grpc.connection.GrpcConnection;
import com.utmstack.grpc.exception.CollectorConfigurationGrpcException;
import com.utmstack.grpc.exception.CollectorServiceGrpcException;
import com.utmstack.grpc.exception.GrpcConnectionException;
import com.utmstack.grpc.service.CollectorService;
import com.utmstack.grpc.service.PanelCollectorService;
import io.grpc.*;
import org.springframework.dao.InvalidDataAccessResourceUsageException;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageImpl;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.Assert;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.validation.BeanPropertyBindingResult;
import org.springframework.validation.Errors;

import javax.persistence.EntityManager;
import java.math.BigInteger;
import java.util.*;
import java.util.stream.Collectors;

@Service
public class CollectorOpsService {
    private final String CLASSNAME = "CollectorOpsService";
    private final Logger log = LoggerFactory.getLogger(CollectorOpsService.class);
    private final GrpcConnection grpcConnection;
    private final PanelCollectorService panelCollectorService;
    private final CollectorService collectorService;
    private final UtmModuleGroupService moduleGroupService;
    private final UtmModuleGroupConfigurationRepository utmModuleGroupConfigurationRepository;

    private final String ENCRYPTION_KEY = System.getenv(Constants.ENV_ENCRYPTION_KEY);

    private final UtmCollectorRepository utmCollectorRepository;

    private final UtmModuleGroupRepository utmModuleGroupRepository;

    private final EntityManager em;

    private final UtmCollectorService utmCollectorService;

    private final UtmModuleService utmModuleService;

    private final UtmModuleRepository utmModuleRepository;

    private final CollectorValidatorService collectorValidatorService;

    public CollectorOpsService(GrpcConnection grpcConnection,
                               UtmModuleGroupService moduleGroupService,
                               UtmModuleGroupConfigurationRepository utmModuleGroupConfigurationRepository,
                               UtmCollectorRepository utmCollectorRepository,
                               UtmModuleGroupRepository utmModuleGroupRepository,
                               EntityManager em,
                               UtmCollectorService utmCollectorService,
                               UtmModuleService utmModuleService,
                               UtmModuleRepository utmModuleRepository,
                               CollectorValidatorService collectorValidatorService) throws GrpcConnectionException {
        this.grpcConnection = grpcConnection;
        this.panelCollectorService = new PanelCollectorService(grpcConnection);
        this.collectorService = new CollectorService(grpcConnection);
        this.moduleGroupService = moduleGroupService;
        this.utmModuleGroupConfigurationRepository = utmModuleGroupConfigurationRepository;
        this.utmCollectorRepository = utmCollectorRepository;
        this.utmModuleGroupRepository = utmModuleGroupRepository;
        this.em = em;
        this.utmCollectorService = utmCollectorService;
        this.utmModuleService = utmModuleService;
        this.utmModuleRepository = utmModuleRepository;
        this.collectorValidatorService = collectorValidatorService;
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
     * Method to List all UtmCollector's hostnames.
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
            ListCollectorResponse response = collectorService.listCollector(request, internalKey);
            CollectorHostnames collectorHostnames = new CollectorHostnames();

            response.getRowsList().forEach(c->{
                collectorHostnames.getHostname().add(c.getHostname());
            });

            return collectorHostnames;
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
    public ListCollectorsResponseDTO getCollectorsByHostnameAndModule(ListRequest request) throws CollectorServiceGrpcException {
        final String ctx = CLASSNAME + ".GetCollectorsByHostnameAndModule";

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

            this.utmCollectorService.synchronize(collectorDTOS);

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
                .setCollectorId(String.valueOf(collectorDTO.getId()))
                .setRequestId(String.valueOf(System.currentTimeMillis()))
                .addAllGroups(collectorGroups)
                .build();
        return collectorConfig;
    }

    /**
     * Method to transform a UtmCollector to CollectorDTO
     */
    private CollectorDTO protoToCollectorDto(Collector collector) {
        return new CollectorDTO(this.utmCollectorService.saveCollector(collector));
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

            Optional<CollectorDTO> collectorToSearch = getCollectorsByHostnameAndModule(
                    getListRequestByHostnameAndModule(hostname, module)).getCollectors()
                    .stream().findFirst();
            try {
                if (collectorToSearch.isEmpty()) {
                    log.error(String.format("%1$s: UtmCollector %2$s could not be deleted because no information was obtained from collector manager", ctx, hostname));
                    return;
                }
            } catch (StatusRuntimeException e) {
                if (e.getStatus().getCode() == Status.Code.NOT_FOUND) {
                    log.error(String.format("%1$s: UtmCollector %2$s could not be deleted because was not found", ctx, hostname));
                    return;
                }
            }

            DeleteRequest collectorDelete = DeleteRequest.newBuilder().setDeletedBy(currentUser).build();
            AuthResponse auth = Common.AuthResponse.newBuilder()
                    .setId(collectorToSearch.get().getId())
                    .setKey(collectorToSearch.get().getCollectorKey())
                    .build();
            collectorService.deleteCollector(collectorDelete, auth);

        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            throw new RuntimeException(msg);
        }
    }

    public List<UtmModuleGroupConfiguration> mapPasswordConfiguration( List<UtmModuleGroupConfiguration> configs) {

       return configs.stream().peek(config -> {
            if (config.getConfDataType().equals("password")) {
                final UtmModuleGroupConfiguration utmModuleGroupConfiguration = utmModuleGroupConfigurationRepository.findById(config.getId())
                        .orElseThrow(() -> new RuntimeException(String.format("Configuration id %s not found", config.getId())));

                if (config.getConfValue().equals(utmModuleGroupConfiguration.getConfValue())){
                    config.setConfValue(CipherUtil.decrypt(utmModuleGroupConfiguration.getConfValue(), ENCRYPTION_KEY));
                }
            }
       }).collect(Collectors.toList());
    }

    @Transactional
    public void updateGroup(List<Long> collectorsIds, Long assetGroupId) throws Exception {
        final String ctx = CLASSNAME + ".updateGroup";
        Assert.notEmpty(collectorsIds, ctx + ": Missing parameter [collectorsIds]");
        try {
            utmCollectorRepository.updateGroup(collectorsIds, assetGroupId);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }
    @Transactional
    public Page<AssetGroupDTO> searchGroupsByFilter(AssetGroupFilter filter, Pageable pageable) throws Exception {
        final String ctx = CLASSNAME + ".searchGroupsByFilter";
        try {
            String query = searchQueryBuilder(filter);
            String queryWithPaginationAndSort = paginateAndSort(query, pageable);
            BigInteger count = (BigInteger) em.createNativeQuery(String.format("SELECT count(*) FROM (%1$s) AS total inner join utm_collectors on total.id = utm_collectors.group_id", query)).getSingleResult();
            List<UtmAssetGroup> results = new ArrayList<>(em.createNativeQuery(queryWithPaginationAndSort, UtmAssetGroup.class).getResultList());

            if (!CollectionUtils.isEmpty(results)) {
                results.forEach(g -> {
                    Optional<List<UtmCollector>> collectors = utmCollectorRepository.findAllByGroupId(g.getId());
                    collectors.ifPresent(g::setCollectors);
                });
            }
            return new PageImpl<>(results.stream().map(AssetGroupDTO::new).collect(Collectors.toList()), pageable, count.longValue());
        } catch (InvalidDataAccessResourceUsageException e) {
            String msg = ctx + ": " + e.getMostSpecificCause().getMessage().replaceAll("\n", "");
            throw new Exception(msg);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    private String searchQueryBuilder(AssetGroupFilter filters) {
        StringBuilder sb = new StringBuilder();
        sb.append(String.format("SELECT DISTINCT utm_asset_group.* FROM utm_asset_group LEFT JOIN utm_collectors ON utm_asset_group.id = utm_collectors.group_id where type =  '%s' \n", filters.getAssetType()));

        if (Objects.isNull(filters))
            return sb.toString();

        boolean where = false;

        // id
        if (Objects.nonNull(filters.getId())) {
            sb.append(String.format("WHERE utm_asset_group.id = %1$s\n", filters.getId()));
            where = false;
        }

        // groupName
        if (StringUtils.hasText(filters.getGroupName())) {
            sb.append(where ? "WHERE " : "AND ")
                    .append(String.format("lower(utm_asset_group.group_name) LIKE '%%%1$s%%'\n",
                            filters.getGroupName().toLowerCase()));
            where = false;
        }

        // createdDate
        if (Objects.nonNull(filters.getInitDate()) && Objects.nonNull(filters.getEndDate())) {
            sb.append(where ? "WHERE " : "AND ")
                    .append(String.format("(utm_asset_group.created_date BETWEEN '%1$s' AND '%2$s')\n",
                            filters.getInitDate(), filters.getEndDate()));
            where = false;
        }

        // assetIp
        if (!CollectionUtils.isEmpty(filters.getAssetIp())) {
            String ips = filters.getAssetIp().stream()
                    .map(ip -> String.format("'%1$s'", ip)).collect(Collectors.joining(","));
            sb.append(where ? "WHERE " : "AND ")
                    .append(String.format("utm_collectors.ip IN (%1$s)\n", ips));
            where = false;
        }

        // assetName
        if (!CollectionUtils.isEmpty(filters.getAssetName())) {
            String names = filters.getAssetName().stream()
                    .map(name -> String.format("'%1$s'", name)).collect(Collectors.joining(","));
            sb.append(where ? "WHERE " : "AND ")
                    .append(String.format("utm_collectors.hostname IN (%1$s)\n", names));
        }

        return sb.toString();
    }

    private String paginateAndSort(String query, Pageable pageable) {
        final String ctx = CLASSNAME + ".paginateAndSort";
        StringBuilder sb = new StringBuilder(query);

        try {
            Sort sort = pageable.getSort();

            if (sort.isSorted()) {
                sb.append("ORDER BY ");
                boolean firstProperty = true;

                List<Sort.Order> orders = sort.stream().collect(Collectors.toList());

                for (Sort.Order order : orders) {
                    sb.append(String.format(firstProperty ? "%1$s %2$s" : ", %1$s %2$s", order.getProperty(), order.getDirection().name()));
                    firstProperty = false;
                }
            }

            if (pageable.isPaged())
                sb.append(String.format(" OFFSET %1$s LIMIT %2$s", pageable.getOffset(), pageable.getPageSize()));

            return sb.toString();
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getMessage());
        }
    }

    @Transactional
    public void deleteCollector(Long id) throws Exception {
        Optional<UtmCollector> collector = utmCollectorRepository.findById(id);
        if (collector.isEmpty()) {
            log.error(String.format("Collector with %1$s not found", id));
            throw new RuntimeException(String.format("Collector with %1$s not found", id));
        } else if (collector.get().isActive()) {
            this.deleteCollector(collector.get().getHostname(), CollectorModuleEnum.AS_400);

            List<UtmModuleGroup> modules = this.utmModuleGroupRepository.findAllByCollector(id.toString());
            if(!modules.isEmpty()){
                UtmModule module = utmModuleRepository.findById(modules.get(0).getModuleId()).get();

                if(module.getModuleActive()){
                    modules = this.utmModuleGroupRepository.findAllByModuleId(module.getId())
                            .stream().filter( m -> !m.getCollector().equals(id.toString()))
                            .collect(Collectors.toList());

                    if(modules.isEmpty()){
                        this.utmModuleService.activateDeactivate(module.getServerId(), module.getModuleName(), false);
                    }
                }
            }
            this.utmModuleGroupRepository.deleteAllByCollector(id.toString());
        }

        utmCollectorRepository.deleteById(id);
    }


    public String validateCollectorConfig(CollectorConfigKeysDTO collectorConfig) {
        Errors errors = new BeanPropertyBindingResult(collectorConfig, "updateConfigurationKeysBody");
        collectorValidatorService.validate(collectorConfig, errors);

        if (errors.hasErrors()) {
            return "Validation failed: Hostname must be unique for this collector.";
        }
        return null;
    }

    public CollectorConfig cacheCurrentCollectorConfig(CollectorDTO collectorDTO) throws CollectorServiceGrpcException {
        return this.getCollectorConfig(
                ConfigRequest.newBuilder()
                        .setModule(CollectorModule.valueOf(collectorDTO.getModule().toString()))
                        .build(),
                AuthResponse.newBuilder()
                        .setId(collectorDTO.getId())
                        .setKey(collectorDTO.getCollectorKey())
                        .build());
    }

    public void updateCollectorConfigViaGrpc(
            CollectorConfigKeysDTO collectorConfig,
            CollectorDTO collectorDTO) throws CollectorConfigurationGrpcException {

        this.upsertCollectorConfig(
                this.mapToCollectorConfig(
                        this.mapPasswordConfiguration(collectorConfig.getKeys()), collectorDTO));
    }

    public void updateCollectorConfigurationKeys(CollectorConfigKeysDTO collectorConfig) throws Exception {
        final String ctx = CLASSNAME + ".updateCollectorConfigurationKeys";
        try {
            List<UtmModuleGroup> configs = utmModuleGroupRepository
                    .findAllByModuleIdAndCollector(collectorConfig.getModuleId(),
                        String.valueOf(collectorConfig.getCollector().getId()));
            List<UtmModuleGroupConfiguration> keys = collectorConfig.getKeys();

            if (CollectionUtils.isEmpty(collectorConfig.getKeys())){
                utmModuleGroupRepository.deleteAll(configs);
            } else {
                for (UtmModuleGroupConfiguration key : keys) {
                    if (key.getConfRequired() && !StringUtils.hasText(key.getConfValue()))
                        throw new Exception(String.format("No value was found for required configuration: %1$s (%2$s)", key.getConfName(), key.getConfKey()));
                    if (key.getConfDataType().equals("password"))
                        key.setConfValue(CipherUtil.encrypt(key.getConfValue(), System.getenv(Constants.ENV_ENCRYPTION_KEY)));
                }
                List<Long> keyGroupIds = keys.stream()
                        .map(UtmModuleGroupConfiguration::getGroupId)
                        .collect(Collectors.toList());

                List<UtmModuleGroup> groupsToDelete = configs.stream()
                        .filter(utmModuleGroup -> !keyGroupIds.contains(utmModuleGroup.getId()))
                        .collect(Collectors.toList());

                utmModuleGroupRepository.deleteAll(groupsToDelete);
                utmModuleGroupConfigurationRepository.saveAll(keys);
            }

        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public ListRequest getListRequestByHostnameAndModule(String hostname, CollectorModuleEnum module) {
        String query = "";
        if (module != null && StringUtils.hasText(hostname)) {
            query = "module.Is=" + module.name() + "&hostname.Is=" + hostname;
        } else if (StringUtils.hasText(hostname)) {
            query = "hostname.Is=" + hostname;
        } else if (module != null) {
            query = "module.Is=" + module.name();
        }
        ListRequest request = ListRequest.newBuilder()
                .setPageNumber(1)
                .setPageSize(1000000)
                .setSearchQuery(query)
                .setSortBy("id,desc")
                .build();
        return request;
    }
}

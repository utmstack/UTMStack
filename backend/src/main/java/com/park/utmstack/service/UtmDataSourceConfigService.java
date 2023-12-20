package com.park.utmstack.service;

import com.park.utmstack.domain.UtmDataSourceConfig;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.repository.UtmDataInputStatusRepository;
import com.park.utmstack.repository.UtmDataSourceConfigRepository;
import com.park.utmstack.repository.network_scan.UtmNetworkScanRepository;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.UtmDataSourceConfigDTO;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.CollectionUtils;

import java.util.List;
import java.util.stream.Collectors;

@Service
@Transactional
public class UtmDataSourceConfigService {
    private static final String CLASSNAME = "UtmDataSourceConfigService";
    private final Logger log = LoggerFactory.getLogger(UtmDataSourceConfigService.class);
    private final UtmDataSourceConfigRepository dataSourceConfigRepository;
    private final UtmDataInputStatusRepository dataInputStatusRepository;
    private final UtmDataInputStatusService dataInputStatusService;
    private final UtmNetworkScanRepository networkScanRepository;
    private final ApplicationEventService applicationEventService;

    public UtmDataSourceConfigService(UtmDataSourceConfigRepository dataSourceConfigRepository,
                                      UtmDataInputStatusRepository dataInputStatusRepository,
                                      UtmDataInputStatusService dataInputStatusService,
                                      UtmNetworkScanRepository networkScanRepository,
                                      ApplicationEventService applicationEventService) {
        this.dataSourceConfigRepository = dataSourceConfigRepository;
        this.dataInputStatusRepository = dataInputStatusRepository;
        this.dataInputStatusService = dataInputStatusService;
        this.networkScanRepository = networkScanRepository;
        this.applicationEventService = applicationEventService;
    }

    /**
     * Update data source configurations, skipped all configurations with an invalid ID
     *
     * @param configs Data source configuration information
     */
    public void update(List<UtmDataSourceConfigDTO> configs) {
        final String ctx = CLASSNAME + ".update";
        try {
            if (CollectionUtils.isEmpty(configs))
                return;
            configs = configs.stream().filter(cfg -> cfg.getId() != null)
                .collect(Collectors.toList());
            if (CollectionUtils.isEmpty(configs))
                return;
            dataSourceConfigRepository.saveAll(configs.stream().map(UtmDataSourceConfig::new).collect(Collectors.toList()));

            configs = configs.stream().filter(cfg-> !cfg.getIncluded()).collect(Collectors.toList());

            networkScanRepository.deleteAllAssetsByDataType(configs.stream().map(UtmDataSourceConfigDTO::getDataType)
                .collect(Collectors.toList()));

            dataInputStatusService.synchronizeSourcesToAssets();
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    /**
     * Retrieve all data source configurations, you can paginate the results
     *
     * @param pageable Paging information
     * @return A page with data source configuration information
     */
    @Transactional(readOnly = true)
    public Page<UtmDataSourceConfig> findAll(Pageable pageable) {
        final String ctx = CLASSNAME + ".findAll";
        try {
            return dataSourceConfigRepository.findAll(pageable);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    /**
     * Keeps synchronized the information between utm_data_input_status and utm_data_source_config
     */
    @Scheduled(fixedDelay = 60000, initialDelay = 10000)
    public void syncDataSourcesConfiguration() {
        final String ctx = CLASSNAME + ".syncDataSourcesConfiguration";
        try {
            synchronizeDatasourceConfigurations();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
        }
    }

    public void synchronizeDatasourceConfigurations() {
        final String ctx = CLASSNAME + ".synchronizeDatasourceConfigurations";
        try {
            // Getting data sources that are not configured
            List<String> newDataSources = dataInputStatusRepository.findDataSourcesToConfigure();

            // Getting all orphan data sources configuration
            List<UtmDataSourceConfig> orphanConfigurations = dataSourceConfigRepository.findOrphanDataSourceConfigurations();

            // Configuring new data sources
            if (!CollectionUtils.isEmpty(newDataSources))
                dataSourceConfigRepository.saveAll(newDataSources.stream().map(UtmDataSourceConfig::new)
                    .collect(Collectors.toList()));

            // Removing orphan data sources configurations
            if (!CollectionUtils.isEmpty(orphanConfigurations))
                dataSourceConfigRepository.deleteAllById(orphanConfigurations.stream().map(UtmDataSourceConfig::getId)
                    .collect(Collectors.toList()));
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }
}

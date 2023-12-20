package com.park.utmstack.service;

import com.park.utmstack.domain.UtmDataInputStatus;
import com.park.utmstack.domain.UtmDataSourceConfig;
import com.park.utmstack.domain.UtmServerModule;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.network_scan.UtmNetworkScan;
import com.park.utmstack.domain.network_scan.enums.AssetStatus;
import com.park.utmstack.domain.network_scan.enums.UpdateLevel;
import com.park.utmstack.domain.shared_types.AlertType;
import com.park.utmstack.repository.UtmDataInputStatusRepository;
import com.park.utmstack.repository.UtmDataSourceConfigRepository;
import com.park.utmstack.repository.network_scan.UtmNetworkScanRepository;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.elasticsearch.ElasticsearchService;
import com.park.utmstack.service.network_scan.UtmNetworkScanService;
import com.park.utmstack.util.enums.AlertSeverityEnum;
import com.park.utmstack.util.enums.AlertStatus;
import org.apache.http.conn.util.InetAddressUtils;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import java.time.LocalDateTime;
import java.time.ZoneId;
import java.time.ZoneOffset;
import java.time.format.DateTimeFormatter;
import java.util.*;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;

/**
 * Service Implementation for managing UtmDataInputStatus.
 */
@Service
@Transactional
public class UtmDataInputStatusService {

    private static final String CLASSNAME = "UtmDataInputStatusService";
    private final Logger log = LoggerFactory.getLogger(UtmDataInputStatusService.class);

    private final UtmDataInputStatusRepository dataInputStatusRepository;
    private final UtmServerModuleService serverModuleService;
    private final ApplicationEventService applicationEventService;
    private final UtmNetworkScanService networkScanService;
    private final ElasticsearchService elasticsearchService;
    private final UtmDataSourceConfigRepository dataSourceConfigRepository;
    private final UtmNetworkScanRepository networkScanRepository;

    public UtmDataInputStatusService(UtmDataInputStatusRepository dataInputStatusRepository,
                                     UtmServerModuleService serverModuleService,
                                     ApplicationEventService applicationEventService,
                                     UtmNetworkScanService networkScanService,
                                     ElasticsearchService elasticsearchService,
                                     UtmDataSourceConfigRepository dataSourceConfigRepository,
                                     UtmNetworkScanRepository networkScanRepository) {
        this.dataInputStatusRepository = dataInputStatusRepository;
        this.serverModuleService = serverModuleService;
        this.applicationEventService = applicationEventService;
        this.networkScanService = networkScanService;
        this.elasticsearchService = elasticsearchService;
        this.dataSourceConfigRepository = dataSourceConfigRepository;
        this.networkScanRepository = networkScanRepository;
    }

    /**
     * Save a utmDataInputStatus.
     *
     * @param utmDataInputStatus the entity to save
     * @return the persisted entity
     */
    public UtmDataInputStatus save(UtmDataInputStatus utmDataInputStatus) {
        log.debug("Request to save UtmDataInputStatus : {}", utmDataInputStatus);
        return dataInputStatusRepository.save(utmDataInputStatus);
    }

    /**
     * Get all the utmDataInputStatuses.
     *
     * @param pageable the pagination information
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public Page<UtmDataInputStatus> findImportantDatasource(Pageable pageable) throws Exception {
        final String ctx = CLASSNAME + ".findImportantDatasource";
        try {
            return dataInputStatusRepository.findImportantDatasource(pageable);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }


    /**
     * Get one utmDataInputStatus by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmDataInputStatus> findOne(String id) {
        log.debug("Request to get UtmDataInputStatus : {}", id);
        return dataInputStatusRepository.findById(id);
    }

    /**
     * Delete the utmDataInputStatus by id.
     *
     * @param id the id of the entity
     */
    public void delete(String id) {
        log.debug("Request to delete UtmDataInputStatus : {}", id);
        dataInputStatusRepository.deleteById(id);
    }

    @Scheduled(fixedDelay = 900000)
    public void checkDatasource() {
        final String ctx = CLASSNAME + ".checkDatasource";
        final List<String> types = Arrays.asList("aws", "o365", "hids");
        try {
            List<UtmDataInputStatus> rows = dataInputStatusRepository.findAllByDataTypeIn(types);

            if (CollectionUtils.isEmpty(rows))
                return;

            List<UtmDataInputStatus> aws = rows.stream().filter(row -> row.getDataType().equalsIgnoreCase("aws")).collect(Collectors.toList());
            List<UtmDataInputStatus> o365 = rows.stream().filter(row -> row.getDataType().equalsIgnoreCase("o365")).collect(Collectors.toList());
            List<UtmDataInputStatus> hids = rows.stream().filter(row -> row.getDataType().equalsIgnoreCase("hids")).collect(Collectors.toList());

            checkDataInputStatus(aws, "aws");
            checkDataInputStatus(o365, "office365");
            checkDataInputStatus(hids, "transporter");
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
        }
    }

    /**
     * Checks the data input status. If any source stop sending logs it is marked to needed restart
     *
     * @param inputs       Data input sources
     * @param serverModule Name of the server module
     * @throws Exception In case of any error
     */
    private void checkDataInputStatus(List<UtmDataInputStatus> inputs, String serverModule) throws Exception {
        final String ctx = CLASSNAME + ".checkDataInputStatus";
        try {
            if (CollectionUtils.isEmpty(inputs))
                return;

            long currentTimeInSeconds = TimeUnit.MILLISECONDS.toSeconds(System.currentTimeMillis());
            List<UtmDataInputStatus> inTime = inputs.stream().filter(row -> (currentTimeInSeconds - row.getTimestamp()) < 3600)
                .collect(Collectors.toList());
            if (!CollectionUtils.isEmpty(inTime))
                return;

            List<UtmServerModule> modules = serverModuleService.findAllByModuleName(serverModule);
            if (CollectionUtils.isEmpty(modules))
                return;

            modules.forEach(module -> module.setNeedsRestart(true));
            serverModuleService.saveAll(modules);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Gets the sources from utm_data_input_status that are not registered in utm_network_scan table
     * and create new assets with it. This method is a schedule with a delay of 1 hour
     */
    @Scheduled(fixedDelay = 15000, initialDelay = 30000)
    public void syncSourcesToAssets() {
        final String ctx = CLASSNAME + ".syncSourcesToAssets";
        try {
            synchronizeSourcesToAssets();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
        }
    }

    public void synchronizeSourcesToAssets() {
        final String ctx = CLASSNAME + ".syncSourcesToAssets";
        try {
            final List<String> excludeOfTypes = dataSourceConfigRepository.findAllByIncludedFalse().stream()
                .map(UtmDataSourceConfig::getDataType).collect(Collectors.toList());
            excludeOfTypes.addAll(Arrays.asList("utmstack", "UTMStack"));

            List<UtmDataInputStatus> sources = dataInputStatusRepository.extractSourcesToExport(excludeOfTypes);
            if (CollectionUtils.isEmpty(sources))
                return;

            Map<String, Boolean> sourcesWithStatus = extractSourcesWithUpDownStatus(sources);
            List<UtmNetworkScan> assets = networkScanService.findAll();

            List<UtmNetworkScan> saveOrUpdate = new ArrayList<>();
            sourcesWithStatus.forEach((key, value) -> {
                Optional<UtmNetworkScan> assetOpt = assets.stream()
                    .filter(asset -> ((StringUtils.hasText(asset.getAssetIp()) && asset.getAssetIp().equals(key))
                        || (StringUtils.hasText(asset.getAssetName()) && asset.getAssetName().equals(key))))
                    .findFirst();
                if (assetOpt.isPresent()) {
                    UtmNetworkScan utmAsset = assetOpt.get();
                    if (Objects.isNull(utmAsset.getUpdateLevel())
                        || utmAsset.getUpdateLevel().equals(UpdateLevel.DATASOURCE)) {
                        utmAsset.assetAlive(value)
                            .updateLevel(UpdateLevel.DATASOURCE)
                            .assetStatus(AssetStatus.CHECK)
                            .modifiedAt(LocalDateTime.now().toInstant(ZoneOffset.UTC));
                        saveOrUpdate.add(utmAsset);
                    }
                } else {
                    saveOrUpdate.add(new UtmNetworkScan(key, value));
                }
            });

            assets.forEach(asset -> {
                if (!sourcesWithStatus.containsKey(asset.getAssetIp()) && !sourcesWithStatus.containsKey(asset.getAssetName())
                    && !Objects.isNull(asset.getUpdateLevel()) && asset.getUpdateLevel().equals(UpdateLevel.DATASOURCE)) {
                    asset.assetStatus(AssetStatus.MISSING).updateLevel(null)
                        .modifiedAt(LocalDateTime.now().toInstant(ZoneOffset.UTC));
                    saveOrUpdate.add(asset);
                }
            });

            networkScanService.saveAll(saveOrUpdate);
            // Finally, delete excluded assets
            networkScanRepository.deleteAllAssetsByDataType(excludeOfTypes);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e.getLocalizedMessage());
        }
    }

    private Map<String, Boolean> extractSourcesWithUpDownStatus(List<UtmDataInputStatus> sources) {
        Map<String, Boolean> upDown = new HashMap<>();
        sources.forEach(src -> {
            Boolean status = upDown.get(src.getSource());
            if (!Objects.isNull(status)) {
                if (!status && !src.isDown())
                    upDown.put(src.getSource(), true);
            } else
                upDown.put(src.getSource(), !src.isDown());
        });
        return upDown;
    }

    /**
     * Check datasource from utm_data_input_status table that are not of type WORKSTATION and
     * if any of them are down then create a new alert. This method is a schedule with a delay
     * of 1 hour
     */
    @Scheduled(fixedDelay = 43200000, initialDelay = 20000)
    public void checkDatasourceDown() {
        final String ctx = CLASSNAME + ".checkDatasourceDown";
        try {
            List<UtmDataInputStatus> sources = dataInputStatusRepository.findDatasourceToCheckIfDown();
            if (CollectionUtils.isEmpty(sources))
                return;
            DateTimeFormatter f = DateTimeFormatter.ofPattern("yyyy.MM.dd").withZone(ZoneId.systemDefault());
            String index = String.format("alert-%1$s", f.format(LocalDateTime.now().toInstant(ZoneOffset.UTC)));
            for (UtmDataInputStatus src : sources) {
                if (src.isDown())
                    elasticsearchService.index(index, createAlertForDatasourceDown(src));
            }
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
        }
    }

    /**
     * Create the alert object for datasource down
     *
     * @param input: Datasource information
     * @return A ${@link AlertType} to index
     */
    private AlertType createAlertForDatasourceDown(UtmDataInputStatus input) {
        List<String> cloudTypes = Arrays.asList("aws", "o365", "office365", "azure", "gcp", "google", "nids", "netflow");
        AlertType alert = new AlertType();
        alert.setId(UUID.randomUUID().toString());
        alert.setTimestamp(LocalDateTime.now().toInstant(ZoneOffset.UTC).toString());

        if (cloudTypes.contains(input.getDataType()))
            alert.setName(String.format("The %1$s datasource is taking longer than usual to send logs", input.getDataType()));
        else
            alert.setName(String.format("The %1$s datasource installed in %2$s is taking longer than usual to send logs", input.getDataType(), input.getSource()));
        alert.setDescription("UTMStack launched this alert because the device exceeded the expected average time in which it can be without sending any log");
        alert.setTactic("Defense Evasion");
        alert.setReference(Collections.singletonList("https://attack.mitre.org/tactics/TA0005/"));
        alert.setStatus(AlertStatus.OPEN.getCode());
        alert.setStatusLabel(AlertStatus.OPEN);
        alert.setSeverity(AlertSeverityEnum.LOW.getCode());
        alert.setSeverityLabel(AlertSeverityEnum.LOW.getName());
        alert.setDataType(input.getDataType());
        alert.setDataSource(input.getSource());
        alert.setNotes("");
        alert.setTags(Collections.emptyList());
        alert.setLogs(Collections.emptyList());
        alert.setProtocol("");

        AlertType.Host src = new AlertType.Host();
        src.setCountry("");
        src.setAccuracyRadius(0);
        src.setCity("");
        src.setCoordinates(new Float[]{});
        src.setPort(0);
        src.setCountryCode("");
        src.setAnonymousProxy(false);
        src.setSatelliteProvider(false);
        src.setAso("");
        src.setAsn(0);
        if (InetAddressUtils.isIPv4Address(input.getSource()) || InetAddressUtils.isIPv6Address(input.getSource()))
            src.setIp(input.getSource());
        else
            src.setHost(input.getSource());

        alert.setSource(src);

        AlertType.Host dst = new AlertType.Host();
        dst.setCountry("");
        dst.setAccuracyRadius(0);
        dst.setCity("");
        dst.setCoordinates(new Float[]{});
        dst.setPort(0);
        dst.setCountryCode("");
        dst.setAnonymousProxy(false);
        dst.setSatelliteProvider(false);
        dst.setAso("");
        dst.setAsn(0);
        dst.setIp("");
        dst.setHost("");
        alert.setDestination(dst);

        alert.setIncident(false);

        AlertType.IncidentDetail incident = new AlertType.IncidentDetail();
        incident.setCreatedBy("");
        incident.setCreationDate("");
        incident.setIncidentId(0);
        incident.setIncidentName("");
        incident.setSource("");

        alert.setIncidentDetail(incident);

        alert.setSolution("Check the data source configuration, error logs, and if it is an agent; verify if it is installed and the service is running");
        alert.setStatusObservation("The system changed the alert status to be analyzed by rule engine");
        alert.setCategory("Data sources monitoring");
        return alert;
    }
}

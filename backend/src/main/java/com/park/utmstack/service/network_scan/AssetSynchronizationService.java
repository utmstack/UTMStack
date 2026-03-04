package com.park.utmstack.service.network_scan;

import com.park.utmstack.domain.correlation.config.UtmTenantConfig;
import com.park.utmstack.domain.datainput_ingestion.UtmDataInputStatus;
import com.park.utmstack.domain.network_scan.UtmNetworkScan;
import com.park.utmstack.domain.network_scan.enums.AssetStatus;
import com.park.utmstack.domain.network_scan.enums.UpdateLevel;
import com.park.utmstack.repository.datainput_ingestion.UtmDataInputStatusRepository;
import com.park.utmstack.repository.network_scan.UtmNetworkScanRepository;
import com.park.utmstack.service.UtmDataInputStatusService;
import com.park.utmstack.service.agent_manager.AgentService;
import com.park.utmstack.service.correlation.config.UtmTenantConfigService;
import com.park.utmstack.service.dto.agent_manager.AgentDTO;
import com.park.utmstack.service.logstash_pipeline.response.statistic.StatisticDocument;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import javax.transaction.Transactional;
import java.time.Instant;
import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.util.*;
import java.util.function.Function;
import java.util.stream.Collectors;

@Service
@Slf4j
@RequiredArgsConstructor
public class AssetSynchronizationService {

    private final AgentService agentService;
    private final SourceActivityProvider sourceActivityProvider;
    private final UtmDataInputStatusRepository dataInputStatusRepository;
    private final UtmNetworkScanRepository networkScanRepository;
    private final UtmDataInputStatusService dataInputStatusService;
    private final UtmTenantConfigService tenantConfigService;

    @Transactional
    @Scheduled(fixedDelay = 60000, initialDelay = 120000)
    public void syncDataInputsAndAssets() {

        String correlationId = UUID.randomUUID().toString().substring(0, 8);
        log.info("[{}] Starting unified asset synchronization cycle", correlationId);

        try {
            Map<String, AgentDTO> agentsMap = loadAgents();
            Map<String, StatisticDocument> statsMap = sourceActivityProvider.fetchLatestSourceActivity();

            if (statsMap.isEmpty()) {
                log.debug("[{}] No new activity detected in data sources", correlationId);
                return;
            }

            List<String> sourcesKeys = new ArrayList<>(statsMap.keySet());

            List<UtmDataInputStatus> dataInputStatus = dataInputStatusService.findDataInputStatus();

            Map<String, UtmDataInputStatus> currentDataInputStatusMap = dataInputStatus
                    .stream()
                    .collect(Collectors.toMap(UtmDataInputStatus::getId, f -> f));

            Map<String, UtmNetworkScan> currentAssetsMap =
                    networkScanRepository.findByAssetIpInOrAssetNameIn(sourcesKeys, sourcesKeys)
                            .stream()
                            .collect(Collectors.toMap(UtmNetworkScan::getAssetName, f -> f));

            List<UtmDataInputStatus> statusToSave = new ArrayList<>();
            List<UtmNetworkScan> assetsToSave = new ArrayList<>();

            Map<String, List<StatisticDocument>> statsBySource = statsMap.values()
                    .stream()
                    .collect(Collectors.groupingBy(StatisticDocument::getDataSource));

            statsBySource.forEach((sourceName, stats) -> {

                for (StatisticDocument stat : stats) {
                    processDataInputStatus(stat, currentDataInputStatusMap, statusToSave);
                }

                processNetworkAsset(sourceName, agentsMap, currentAssetsMap, currentDataInputStatusMap, assetsToSave);
            });

            if (!statusToSave.isEmpty()) dataInputStatusRepository.saveAll(statusToSave);
            if (!assetsToSave.isEmpty()) networkScanRepository.saveAll(assetsToSave);

            log.info("[{}] Asset synchronization cycle completed successfully - {} data input status updated, {} assets synced",
                    correlationId, statusToSave.size(), assetsToSave.size());

        } catch (Exception e) {
            log.error("[{}] Critical error during asset synchronization: {}", correlationId, e.getMessage(), e);
        }
    }

    private void processDataInputStatus(StatisticDocument stat,
                                        Map<String, UtmDataInputStatus> currentStatusMap,
                                        List<UtmDataInputStatus> statusToSave) {

        String statusId = stat.getDataType() + "-" + stat.getDataSource();
        long statTimestamp = Instant.parse(stat.getTimestamp()).getEpochSecond();

        UtmDataInputStatus status = currentStatusMap.getOrDefault(statusId,
                UtmDataInputStatus.builder()
                        .id(statusId)
                        .dataType(stat.getDataType())
                        .timestamp(statTimestamp)
                        .source(stat.getDataSource())
                        .median(86400L)
                        .build());

        boolean isExisting = status.getId() != null;

        if (isExisting && status.getTimestamp() != statTimestamp) {
            status.setTimestamp(statTimestamp);
        }

        statusToSave.add(status);
    }

    private void processNetworkAsset(String sourceName,
                                     Map<String, AgentDTO> agentsMap,
                                     Map<String, UtmNetworkScan> currentAssetsMap,
                                     Map<String, UtmDataInputStatus> currentDataInputStatusMap,
                                     List<UtmNetworkScan> assetsToSave) {

        boolean hasAlias = false;
        boolean isAlive = currentDataInputStatusMap.values().stream()
                .filter(status -> status.getSource().equalsIgnoreCase(sourceName))
                .anyMatch(s -> !s.isDown());

        UtmNetworkScan asset = currentAssetsMap.get(sourceName);
        String resolvedAssetName = null;

        if (asset == null) {
            asset = resolveAssetNameFromTenantConfig(sourceName);
             hasAlias = asset != null;
        }

        boolean isExisting = asset != null && asset.getId() != null;

        if (asset == null) {
            asset = new UtmNetworkScan(sourceName, isAlive);
        } else {
            if (hasAlias) {
                asset.assetName(sourceName);
            }
        }

        asset.assetAlive(isAlive)
                .updateLevel(UpdateLevel.DATASOURCE)
                .modifiedAt(LocalDateTime.now().toInstant(ZoneOffset.UTC));

        if (isExisting) {
            asset.assetStatus(AssetStatus.CHECK);
        }

        AgentDTO agentInfo = agentsMap.get(sourceName);
        if (agentInfo != null) {
            asset.setAssetIp(agentInfo.getIp());
            asset.setAssetOs(agentInfo.getOs());
            asset.setAssetOsPlatform(agentInfo.getPlatform());
            asset.setAssetOsVersion(agentInfo.getVersion());
            asset.setIsAgent(true);
        } else {
            asset.setIsAgent(false);
        }

        assetsToSave.add(asset);
    }

    private UtmNetworkScan resolveAssetNameFromTenantConfig(String sourceName) {

        if (!StringUtils.hasText(sourceName)) {
            return null;
        }

        try {

            Optional<UtmTenantConfig> configOpt = tenantConfigService.findByAssetName(sourceName);

            if (configOpt.isEmpty()) {
                return null;
            }

            UtmTenantConfig config = configOpt.get();
            List<String> hostnames = config.getAssetHostnameList();
            List<String> ips = config.getAssetIpList();

            List<UtmNetworkScan> networkScans =
                    networkScanRepository.findByAssetIpInOrAssetNameIn(ips, hostnames)
                            .stream()
                            .toList();

            if (networkScans.isEmpty()) {
                return null;
            }

            for (UtmNetworkScan networkScan : networkScans) {

                if (hostnames != null && hostnames.contains(networkScan.getAssetName()) ||
                    ips != null && ips.contains(networkScan.getAssetIp())) {

                    return networkScan;
                }

            }


        } catch (Exception e) {
            log.warn("Error resolving asset name from tenant config for source {}: {}", sourceName, e.getMessage());
        }

        return null;
    }

    private Map<String, AgentDTO> loadAgents() {
        return agentService.getInstalledAgents().stream()
                .collect(Collectors.toMap(AgentDTO::getHostname, Function.identity(), (a1, a2) -> a1));
    }

}

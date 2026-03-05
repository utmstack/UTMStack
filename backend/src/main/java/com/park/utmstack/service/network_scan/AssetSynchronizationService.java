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
            Map<String, StatisticDocument> statsMap = sourceActivityProvider.fetchLatestSourceActivity();
            if (statsMap.isEmpty()) {
                log.debug("[{}] No new activity detected in data sources", correlationId);
                return;
            }

            Map<String, AgentDTO> agentsMap = loadAgents();
            Map<String, UtmDataInputStatus> statusMap = buildDataInputStatusMap();
            Map<String, UtmNetworkScan> assetsMap = buildNetworkAssetsMap(
                    new ArrayList<>(statsMap.values()
                            .stream()
                            .map(StatisticDocument::getDataSource)
                            .collect(Collectors.toSet())));

            List<UtmDataInputStatus> statusToSave = new ArrayList<>();
            List<UtmNetworkScan> assetsToSave = new ArrayList<>();

            for (String sourceName : statsMap.keySet()) {

                StatisticDocument stat = statsMap.get(sourceName);
                UtmDataInputStatus status = processDataInputStatus(stat, statusMap);
                statusToSave.add(status);

                // Update network asset
                UtmNetworkScan asset = processNetworkAsset(stat.getDataSource(), agentsMap, assetsMap, statusMap);
                assetsToSave.add(asset);
            }

            if (!statusToSave.isEmpty()) {
                dataInputStatusRepository.saveAll(statusToSave);
            }

            if (!assetsToSave.isEmpty()) {
                networkScanRepository.saveAll(assetsToSave);
            }

            log.info("[{}] Asset synchronization cycle completed - {} status updated, {} assets synced",
                    correlationId, statusToSave.size(), assetsToSave.size());

        } catch (Exception e) {
            log.error("[{}] Critical error during asset synchronization: {}", correlationId, e.getMessage(), e);
        }
    }

    private Map<String, UtmDataInputStatus> buildDataInputStatusMap() {
        return dataInputStatusService.findDataInputStatus()
                .stream()
                .collect(Collectors.toMap(UtmDataInputStatus::getId, Function.identity()));
    }

    private Map<String, UtmNetworkScan> buildNetworkAssetsMap(List<String> sourcesKeys) {
        return networkScanRepository.findByAssetIpInOrAssetNameIn(sourcesKeys, sourcesKeys)
                .stream()
                .collect(Collectors.toMap(UtmNetworkScan::getAssetName, Function.identity(), (a1, a2) -> a1));
    }

    private UtmDataInputStatus processDataInputStatus(StatisticDocument stat,
                                                      Map<String, UtmDataInputStatus> statusMap) {
        String statusId = stat.getDataType() + "-" + stat.getDataSource();
        long statTimestamp = Instant.parse(stat.getTimestamp()).getEpochSecond();

        UtmDataInputStatus status = statusMap.getOrDefault(statusId, createNewDataInputStatus(statusId, stat, statTimestamp));

        if (status.getTimestamp() != statTimestamp) {
            status.setTimestamp(statTimestamp);
        }

        return status;
    }

    private UtmDataInputStatus createNewDataInputStatus(String id, StatisticDocument stat, long timestamp) {
        return UtmDataInputStatus.builder()
                .id(id)
                .dataType(stat.getDataType())
                .timestamp(timestamp)
                .source(stat.getDataSource())
                .median(86400L)
                .build();
    }

    private UtmNetworkScan processNetworkAsset(String sourceName,
                                               Map<String, AgentDTO> agentsMap,
                                               Map<String, UtmNetworkScan> assetsMap,
                                               Map<String, UtmDataInputStatus> statusMap) {
        boolean isAlive = isDataSourceAlive(sourceName, statusMap);
        UtmNetworkScan asset = resolveAsset(sourceName, assetsMap);
        boolean isExisting = asset != null && asset.getId() != null;

        if (asset == null) {
            asset = new UtmNetworkScan(sourceName, isAlive);
        }

        enrichAssetWithData(asset, sourceName, agentsMap, isAlive, isExisting);
        return asset;
    }

    private boolean isDataSourceAlive(String sourceName, Map<String, UtmDataInputStatus> statusMap) {
        return statusMap.values().stream()
                .filter(status -> status.getSource().equalsIgnoreCase(sourceName))
                .anyMatch(s -> !s.isDown());
    }

    private UtmNetworkScan resolveAsset(String sourceName, Map<String, UtmNetworkScan> assetsMap) {
        UtmNetworkScan asset = assetsMap.get(sourceName);

        if (asset == null) {
            asset = resolveAssetNameFromTenantConfig(sourceName);
        }

        return asset;
    }

    private void enrichAssetWithData(UtmNetworkScan asset,
                                     String sourceName,
                                     Map<String, AgentDTO> agentsMap,
                                     boolean isAlive,
                                     boolean isExisting) {
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

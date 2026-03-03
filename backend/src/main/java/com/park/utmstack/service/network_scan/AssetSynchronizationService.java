package com.park.utmstack.service.network_scan;

import com.park.utmstack.domain.datainput_ingestion.UtmDataInputStatus;
import com.park.utmstack.domain.network_scan.UtmNetworkScan;
import com.park.utmstack.domain.network_scan.enums.AssetStatus;
import com.park.utmstack.domain.network_scan.enums.UpdateLevel;
import com.park.utmstack.repository.datainput_ingestion.UtmDataInputStatusRepository;
import com.park.utmstack.repository.network_scan.UtmNetworkScanRepository;
import com.park.utmstack.service.UtmDataInputStatusService;
import com.park.utmstack.service.UtmServerModuleService;
import com.park.utmstack.service.agent_manager.AgentService;
import com.park.utmstack.service.dto.agent_manager.AgentDTO;
import com.park.utmstack.service.logstash_pipeline.response.statistic.StatisticDocument;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

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

    @Transactional
    @Scheduled(fixedDelay = 30000, initialDelay = 60000)
    public void syncEverything() {

        String correlationId = UUID.randomUUID().toString().substring(0, 8);
        log.info("[{}] Starting unified asset synchronization cycle...", correlationId);

        try {
            Map<String, AgentDTO> agentsMap = loadAgents();
            Map<String, StatisticDocument> statsMap = sourceActivityProvider.fetchLatestSourceActivity();

            if (statsMap.isEmpty()) {
                log.info("[{}] No new activity detected in data sources.", correlationId);
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

                processNetworkAsset(sourceName, stats, agentsMap, currentAssetsMap, currentDataInputStatusMap, assetsToSave);
            });

            if (!statusToSave.isEmpty()) dataInputStatusRepository.saveAll(statusToSave);
            if (!assetsToSave.isEmpty()) networkScanRepository.saveAll(assetsToSave);

            log.info("[{}] Cycle completed successfully. Status updated: {}, Assets synced: {}",
                    correlationId, statusToSave.size(), assetsToSave.size());

        } catch (Exception e) {
            log.error("[{}] Critical error during synchronization: {}", correlationId, e.getMessage(), e);
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
                                     List<StatisticDocument> stats,
                                     Map<String, AgentDTO> agentsMap,
                                     Map<String, UtmNetworkScan> currentAssetsMap,
                                     Map<String, UtmDataInputStatus> currentDataInputStatusMap,
                                     List<UtmNetworkScan> assetsToSave) {

        boolean isAlive = currentDataInputStatusMap.values().stream()
                .filter(status -> status.getSource().equalsIgnoreCase(sourceName))
                .anyMatch(s -> !s.isDown());

        UtmNetworkScan asset = currentAssetsMap.getOrDefault(sourceName,
                new UtmNetworkScan(sourceName, isAlive));

        boolean isExisting = asset.getId() != null;
        AgentDTO agentInfo = agentsMap.get(sourceName);

        asset.assetAlive(isAlive)
                .updateLevel(UpdateLevel.DATASOURCE)
                .modifiedAt(LocalDateTime.now().toInstant(ZoneOffset.UTC));

        if (isExisting) {
            asset.assetStatus(AssetStatus.CHECK);
        }

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

    private Map<String, AgentDTO> loadAgents() {
        return agentService.getInstalledAgents().stream()
                .collect(Collectors.toMap(AgentDTO::getHostname, Function.identity(), (a1, a2) -> a1));
    }

}

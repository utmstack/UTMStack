package com.park.utmstack.service.agent_manager;

import com.park.utmstack.repository.network_scan.UtmNetworkScanRepository;
import com.park.utmstack.service.grpc.ListRequest;
import com.park.utmstack.domain.agents_manager.Agent;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.network_scan.UtmNetworkScan;
import com.park.utmstack.domain.network_scan.enums.AssetStatus;
import com.park.utmstack.domain.network_scan.wrapper.NetworkScanWrapper;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.agent_manager.AgentDTO;
import com.park.utmstack.service.dto.agent_manager.ListAgentsResponseDTO;
import com.park.utmstack.service.network_scan.UtmNetworkScanService;
import com.park.utmstack.util.exceptions.ApiException;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.Assert;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.util.*;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
public class AgentService {
    private static final String CLASSNAME = "AgentService";
    private final Logger log = LoggerFactory.getLogger(AgentService.class);
    private final ApplicationEventService eventService;
    private final UtmNetworkScanService networkScanService;
    private final AgentGrpcService agentGrpcService;
    private final UtmNetworkScanRepository networkScanRepository;


    /**
     * It queries the agent API to gets the info of all installed agents
     *
     * @return A list of ${@link Agent}
     */
    public List<AgentDTO> getInstalledAgents() {
        final String ctx = CLASSNAME + ".getInstalledAgents";
        try {

            ListRequest request = ListRequest.newBuilder()
                    .setPageNumber(1)
                    .setPageSize(1000000)
                    .setSearchQuery("")
                    .setSortBy("")
                    .build();
            ListAgentsResponseDTO response = agentGrpcService.listAgents(request);

            if (CollectionUtils.isEmpty(response.getAgents()))
                return Collections.emptyList();

            return response.getAgents();
        } catch (Exception e) {
            log.error("{}: An error occurred while getting installed agents: {}", ctx, e.getMessage());
            throw new ApiException(String.format("%s: An error occurred while getting installed agents: %s", ctx, e.getMessage()), HttpStatus.INTERNAL_SERVER_ERROR);
        }
    }

    /**
     * Get an agent by his hostname
     *
     * @param hostname The hostname of the agent
     * @return An Optional of ${@link AgentDTO} object or empty if the agent was not found
     */
    public Optional<AgentDTO> getAgentByHostName(String hostname) {
        final String ctx = CLASSNAME + ".getAgentByHostName";
        try {
            Assert.hasText(hostname, "Parameter hostname is required");

            ListRequest rq = ListRequest.newBuilder()
                    .setPageNumber(1)
                    .setPageSize(100000)
                    .setSearchQuery("hostname.Is=" + hostname)
                    .setSortBy("")
                    .build();

            ListAgentsResponseDTO rs = agentGrpcService.listAgents(rq);

            if (CollectionUtils.isEmpty(rs.getAgents()))
                return Optional.empty();
            return Optional.of(rs.getAgents().get(0));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            throw new RuntimeException(msg);
        }
    }

    @Transactional
   /* @Scheduled(fixedDelay = 15000, initialDelay = 30000)*/
    public void synchronizeAgents() {
        final String ctx = CLASSNAME + ".synchronizeAgents";

        try {
            List<AgentDTO> agents = getInstalledAgents();

            if (CollectionUtils.isEmpty(agents)) {
                List<UtmNetworkScan> agentAssets = networkScanRepository.findByIsAgentTrue();

                for (UtmNetworkScan asset : agentAssets) {
                    asset.assetAlive(false)
                            .assetStatus(AssetStatus.MISSING)
                            .modifiedAt(LocalDateTime.now().toInstant(ZoneOffset.UTC));
                    networkScanService.save(asset);
                }
                return;
            }

            List<String> agentNames = agents.stream()
                    .map(AgentDTO::getHostname)
                    .filter(StringUtils::hasText)
                    .toList();

            List<UtmNetworkScan> existingAgentAssets = networkScanRepository.findByIsAgentTrue();

            this.insertUpdateOrMarkAsMissing(agents, existingAgentAssets, agentNames);

        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
        }
    }


    private void insertUpdateOrMarkAsMissing(List<AgentDTO> agents,
                                             List<UtmNetworkScan> existingAssets,
                                             List<String> agentNames) throws Exception {

        final String ctx = CLASSNAME + ".insertUpdateOrMarkAsMissing";

        try {
            Map<String, UtmNetworkScan> assetsByName = new HashMap<>();
            existingAssets.forEach(a -> assetsByName.put(a.getAssetName(), a));

            for (AgentDTO agent : agents) {

                String agentName = agent.getHostname();
                if (!StringUtils.hasText(agentName)) continue;

                UtmNetworkScan asset = assetsByName.get(agentName);

                if (asset != null) {
                    NetworkScanWrapper.mergeAgentIntoAsset(agent, asset, AssetStatus.CHECK);
                    asset.modifiedAt(LocalDateTime.now().toInstant(ZoneOffset.UTC));
                    networkScanService.save(asset);
                } else {
                    UtmNetworkScan newAsset = NetworkScanWrapper.agentToAsset(agent);
                    networkScanService.save(newAsset);
                }
            }

            for (UtmNetworkScan asset : existingAssets) {

                if (!asset.getIsAgent()) continue;

                if (!agentNames.contains(asset.getAssetName())) {
                    asset.assetAlive(false)
                            .assetStatus(AssetStatus.MISSING)
                            .modifiedAt(LocalDateTime.now().toInstant(ZoneOffset.UTC));
                    networkScanService.save(asset);
                }
            }

        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

}

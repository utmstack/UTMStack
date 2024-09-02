package com.park.utmstack.service.agent_manager;

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
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.util.Assert;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

@Service
public class AgentService {
    private static final String CLASSNAME = "AgentService";
    private final Logger log = LoggerFactory.getLogger(AgentService.class);
    private final ApplicationEventService eventService;
    private final UtmNetworkScanService networkScanService;
    private final AgentGrpcService agentGrpcService;

    public AgentService(ApplicationEventService eventService,
                        UtmNetworkScanService networkScanService,
                        AgentGrpcService agentGrpcService) {
        this.eventService = eventService;
        this.networkScanService = networkScanService;
        this.agentGrpcService = agentGrpcService;
    }

    /**
     * It queries the agent API to gets the info of all installed agents
     *
     * @return A list of ${@link Agent}
     * @throws Exception In case of any error
     */
    private List<AgentDTO> getInstalledAgents() throws Exception {
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
            throw new Exception(ctx + ": " + e.getMessage());
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

    /**
     *
     */
    @Scheduled(fixedDelay = 15000, initialDelay = 30000)
    public void synchronizeAgents() {
        final String ctx = CLASSNAME + ".synchronizeAgents";
        try {
            List<AgentDTO> agents = getInstalledAgents();
            List<UtmNetworkScan> assets = networkScanService.findAll();

            // If no remote agents installed founded, then it is searched if in the database are agents previously
            // registered, if so, then they are marked as missing.
            if (CollectionUtils.isEmpty(agents)) {
                if (!CollectionUtils.isEmpty(assets)) {
                    assets.forEach(asset -> {
                        if (asset.getIsAgent())
                            asset.assetAlive(false).modifiedAt(LocalDateTime.now().toInstant(ZoneOffset.UTC));
                    });
                    networkScanService.saveAll(assets);
                }
            } else if (CollectionUtils.isEmpty(assets)) {
                networkScanService.saveAll(agents.stream()
                    .map(UtmNetworkScan::new).collect(Collectors.toList()));
            } else {
                insertUpdateOrMarkAsMissing(agents, assets);
            }
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
        }
    }

    private void insertUpdateOrMarkAsMissing(List<AgentDTO> agents, List<UtmNetworkScan> assets) throws Exception {
        final String ctx = CLASSNAME + ".insertUpdateOrMarkAsMissing";
        try {
            List<UtmNetworkScan> result = new ArrayList<>();
            agents.forEach(agent -> {
                Optional<UtmNetworkScan> first = assets.stream()
                    .filter(asset -> (StringUtils.hasText(agent.getHostname()) && agent.getHostname().equals(asset.getAssetName())))
                    .findFirst();
                if (first.isPresent())
                    result.add(NetworkScanWrapper.mergeAgentIntoAsset(agent, first.get(), AssetStatus.CHECK));
                else
                    result.add(NetworkScanWrapper.agentToAsset(agent));
            });
            List<String> agentNames = agents.stream().map(AgentDTO::getHostname).collect(Collectors.toList());
            result.addAll(assets.stream()
                .filter(asset -> (asset.getIsAgent() && !agentNames.contains(asset.getAssetName())))
                .peek(asset -> asset.modifiedAt(LocalDateTime.now().toInstant(ZoneOffset.UTC))
                    .assetAlive(false)
                    .assetStatus(AssetStatus.MISSING)).collect(Collectors.toList()));
            networkScanService.saveAll(result);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }
}

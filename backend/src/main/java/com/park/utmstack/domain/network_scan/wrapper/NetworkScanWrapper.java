package com.park.utmstack.domain.network_scan.wrapper;

import com.park.utmstack.domain.network_scan.NetworkScan;
import com.park.utmstack.domain.network_scan.UtmNetworkScan;
import com.park.utmstack.domain.network_scan.enums.AssetRegisteredMode;
import com.park.utmstack.domain.network_scan.enums.AssetStatus;
import com.park.utmstack.domain.network_scan.enums.UpdateLevel;
import com.park.utmstack.service.dto.agent_manager.AgentDTO;
import com.park.utmstack.service.dto.agent_manager.AgentStatusEnum;
import org.springframework.util.CollectionUtils;

import java.time.LocalDateTime;
import java.time.ZoneOffset;

public class NetworkScanWrapper {
    private static final String CLASSNAME = "NetworkScanWrapper";

    /**
     * @param scan
     * @return
     */
    public static UtmNetworkScan build(NetworkScan scan) {
        final String ctx = CLASSNAME + ".build";
        try {
            UtmNetworkScan result = new UtmNetworkScan();
            return result.assetIp(scan.getIp())
                .assetAlive(scan.getAlive())
                .assetMac(scan.getMac())
                .assetName(scan.getName())
                .assetOs(scan.getOs())
                .registeredMode(AssetRegisteredMode.DISCOVERED)
                .assetAliases(!CollectionUtils.isEmpty(scan.getAliasList()) ? String.join(",", scan.getAliasList()) : null)
                .assetAddresses(!CollectionUtils.isEmpty(scan.getAddressList()) ? String.join(",", scan.getAddressList()) : null);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e);
        }
    }

    /**
     * @param networkScan
     * @param utmNetworkScan
     * @return
     */
    public static UtmNetworkScan merge(NetworkScan networkScan, UtmNetworkScan utmNetworkScan) {
        final String ctx = CLASSNAME + ".merge";
        try {
            return utmNetworkScan
                .assetIp(networkScan.getIp())
                .assetAlive(networkScan.getAlive())
                .assetName(networkScan.getName())
                .assetOs(networkScan.getOs())
                .assetAddresses(!CollectionUtils.isEmpty(networkScan.getAddressList()) ? String.join(",", networkScan.getAddressList()) : null)
                .assetAliases(!CollectionUtils.isEmpty(networkScan.getAliasList()) ? String.join(",", networkScan.getAliasList()) : null);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e);
        }
    }

    public static UtmNetworkScan agentToAsset(AgentDTO agent) {
        final String ctx = CLASSNAME + ".agentToAsset";
        try {
            return new UtmNetworkScan(agent);
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e);
        }
    }

    public static UtmNetworkScan mergeAgentIntoAsset(AgentDTO agent, UtmNetworkScan asset, AssetStatus status) {
        final String ctx = CLASSNAME + ".mergeAgentIntoAsset";
        try {
            asset.assetIp(agent.getIp())
                .updateLevel(UpdateLevel.AGENT)
                .isAgent(true)
                .registerIp(agent.getIp())
                .assetAlive(agent.getStatus() == AgentStatusEnum.ONLINE)
                .assetStatus(status)
                .modifiedAt(LocalDateTime.now().toInstant(ZoneOffset.UTC))
                .assetOs(agent.getOs())
                .assetOsPlatform(agent.getPlatform())
                .assetOsMajorVersion(agent.getOsMajorVersion())
                .assetOsMinorVersion(agent.getOsMinorVersion())
                .assetOsVersion(agent.getVersion());
            return asset;
        } catch (Exception e) {
            throw new RuntimeException(ctx + ": " + e);
        }
    }
}

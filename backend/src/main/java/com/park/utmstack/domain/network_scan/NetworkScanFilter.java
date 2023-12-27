package com.park.utmstack.domain.network_scan;

import com.park.utmstack.domain.network_scan.enums.AssetRegisteredMode;
import com.park.utmstack.domain.network_scan.enums.AssetStatus;

import java.time.Instant;
import java.util.List;

public class NetworkScanFilter {
    private String assetIpMacName;
    private List<String> os;
    private List<Boolean> alive;
    private List<AssetStatus> status;
    private List<String> type;
    private List<String> alias;
    private List<String> probe;
    private List<Integer> openPorts;
    private List<String> groups;
    private Instant discoveredInitDate;
    private Instant discoveredEndDate;
    private AssetRegisteredMode registeredMode;
    private List<Boolean> agent;

    private List<String> osPlatform;

    public String getAssetIpMacName() {
        return assetIpMacName;
    }

    public void setAssetIpMacName(String assetIpMacName) {
        this.assetIpMacName = assetIpMacName;
    }

    public List<String> getAlias() {
        return alias;
    }

    public void setAlias(List<String> alias) {
        this.alias = alias;
    }

    public List<String> getOs() {
        return os;
    }

    public void setOs(List<String> os) {
        this.os = os;
    }

    public List<Boolean> getAlive() {
        return alive;
    }

    public void setAlive(List<Boolean> alive) {
        this.alive = alive;
    }

    public List<AssetStatus> getStatus() {
        return status;
    }

    public void setStatus(List<AssetStatus> status) {
        this.status = status;
    }

    public List<String> getType() {
        return type;
    }

    public void setType(List<String> type) {
        this.type = type;
    }

    public List<String> getProbe() {
        return probe;
    }

    public void setProbe(List<String> probe) {
        this.probe = probe;
    }

    public List<Integer> getOpenPorts() {
        return openPorts;
    }

    public void setOpenPorts(List<Integer> openPorts) {
        this.openPorts = openPorts;
    }

    public Instant getDiscoveredInitDate() {
        return discoveredInitDate;
    }

    public void setDiscoveredInitDate(Instant discoveredInitDate) {
        this.discoveredInitDate = discoveredInitDate;
    }

    public Instant getDiscoveredEndDate() {
        return discoveredEndDate;
    }

    public void setDiscoveredEndDate(Instant discoveredEndDate) {
        this.discoveredEndDate = discoveredEndDate;
    }

    public List<String> getGroups() {
        return groups;
    }

    public void setGroups(List<String> groups) {
        this.groups = groups;
    }

    public AssetRegisteredMode getRegisteredMode() {
        return registeredMode;
    }

    public void setRegisteredMode(AssetRegisteredMode registeredMode) {
        this.registeredMode = registeredMode;
    }

    public List<Boolean> getAgent() {
        return agent;
    }

    public void setAgent(List<Boolean> agent) {
        this.agent = agent;
    }

    public List<String> getOsPlatform() {
        return osPlatform;
    }

    public void setOsPlatform(List<String> osPlatform) {
        this.osPlatform = osPlatform;
    }
}

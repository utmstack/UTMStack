package com.park.utmstack.service.dto.network_scan;

import com.park.utmstack.domain.UtmDataInputStatus;
import com.park.utmstack.domain.network_scan.UtmAssetGroup;
import com.park.utmstack.domain.network_scan.UtmAssetTypes;
import com.park.utmstack.domain.network_scan.UtmNetworkScan;
import com.park.utmstack.domain.network_scan.UtmPorts;
import com.park.utmstack.domain.network_scan.enums.AssetRegisteredMode;
import com.park.utmstack.domain.network_scan.enums.AssetStatus;
import org.springframework.util.CollectionUtils;

import java.time.Instant;
import java.util.*;
import java.util.stream.Collectors;

public class NetworkScanDTO {
    private Long id;
    private String assetIp;
    private String assetAddresses;
    private String assetMac;
    private String assetOs;
    private String assetOsArch;
    private String assetOsMajorVersion;
    private String assetOsMinorVersion;
    private String assetOsPlatform;
    private String assetOsVersion;
    private String assetName;
    private String assetAliases;
    private String assetAlias;
    private String serverName;
    private Boolean assetAlive;
    private AssetRegisteredMode registeredMode;
    private AssetStatus assetStatus;
    private Float assetSeverityMetric;
    private UtmAssetTypes assetType;
    private String assetNotes;
    private Instant discoveredAt;
    private Instant modifiedAt;
    private final Map<String, Long> metrics = new HashMap<>();
    private Set<Port> ports;
    private UtmAssetGroup group;
    private Boolean isAgent;

    private List<UtmDataInputStatus> dataInputList;

    public NetworkScanDTO() {
    }

    public NetworkScanDTO(UtmNetworkScan scan, boolean details) {
        this.id = scan.getId();
        this.assetIp = scan.getAssetIp();
        this.assetAddresses = scan.getAssetAddresses();
        this.assetMac = scan.getAssetMac();
        this.assetOs = scan.getAssetOs();
        this.assetOsArch = scan.getAssetOsArch();
        this.assetOsMajorVersion = scan.getAssetOsMajorVersion();
        this.assetOsMinorVersion = scan.getAssetOsMinorVersion();
        this.assetOsPlatform = scan.getAssetOsPlatform();
        this.assetOsVersion = scan.getAssetOsVersion();
        this.assetName = scan.getAssetName();
        this.assetAliases = scan.getAssetAliases();
        this.assetAlive = scan.getAssetAlive();
        this.assetStatus = scan.getAssetStatus();
        this.assetSeverityMetric = scan.getAssetSeverityMetric();
        this.assetType = scan.getAssetType();
        this.assetNotes = scan.getAssetNotes();
        this.discoveredAt = scan.getDiscoveredAt();
        this.modifiedAt = scan.getModifiedAt();
        this.serverName = scan.getServerName();
        this.group = scan.getAssetGroup();
        this.registeredMode = scan.getRegisteredMode();
        this.assetAlias = scan.getAssetAlias();
        this.isAgent = scan.getIsAgent();

        this.dataInputList = !CollectionUtils.isEmpty(scan.getDataInputList()) ? scan.getDataInputList() : new ArrayList<>();

        if (!CollectionUtils.isEmpty(scan.getMetrics()))
            scan.getMetrics().forEach(metric -> this.metrics.put(metric.getMetric(), metric.getAmount()));

        if (details) {
            if (!CollectionUtils.isEmpty(scan.getPorts()))
                this.ports = scan.getPorts().stream().map(Port::new).collect(Collectors.toSet());
        }
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getAssetIp() {
        return assetIp;
    }

    public void setAssetIp(String assetIp) {
        this.assetIp = assetIp;
    }

    public String getAssetAddresses() {
        return assetAddresses;
    }

    public void setAssetAddresses(String assetAddresses) {
        this.assetAddresses = assetAddresses;
    }

    public String getAssetMac() {
        return assetMac;
    }

    public void setAssetMac(String assetMac) {
        this.assetMac = assetMac;
    }

    public String getAssetOs() {
        return assetOs;
    }

    public void setAssetOs(String assetOs) {
        this.assetOs = assetOs;
    }

    public String getAssetOsArch() {
        return assetOsArch;
    }

    public void setAssetOsArch(String assetOsArch) {
        this.assetOsArch = assetOsArch;
    }

    public String getAssetOsMajorVersion() {
        return assetOsMajorVersion;
    }

    public void setAssetOsMajorVersion(String assetOsMajorVersion) {
        this.assetOsMajorVersion = assetOsMajorVersion;
    }

    public String getAssetOsMinorVersion() {
        return assetOsMinorVersion;
    }

    public void setAssetOsMinorVersion(String assetOsMinorVersion) {
        this.assetOsMinorVersion = assetOsMinorVersion;
    }

    public String getAssetOsPlatform() {
        return assetOsPlatform;
    }

    public void setAssetOsPlatform(String assetOsPlatform) {
        this.assetOsPlatform = assetOsPlatform;
    }

    public String getAssetOsVersion() {
        return assetOsVersion;
    }

    public void setAssetOsVersion(String assetOsVersion) {
        this.assetOsVersion = assetOsVersion;
    }

    public String getAssetName() {
        return assetName;
    }

    public void setAssetName(String assetName) {
        this.assetName = assetName;
    }

    public String getAssetAliases() {
        return assetAliases;
    }

    public void setAssetAliases(String assetAliases) {
        this.assetAliases = assetAliases;
    }

    public String getAssetAlias() {
        return assetAlias;
    }

    public void setAssetAlias(String assetAlias) {
        this.assetAlias = assetAlias;
    }


    public Boolean getAssetAlive() {
        return assetAlive;
    }

    public void setAssetAlive(Boolean assetAlive) {
        this.assetAlive = assetAlive;
    }

    public AssetStatus getAssetStatus() {
        return assetStatus;
    }

    public void setAssetStatus(AssetStatus assetStatus) {
        this.assetStatus = assetStatus;
    }

    public Float getAssetSeverityMetric() {
        return assetSeverityMetric;
    }

    public void setAssetSeverityMetric(Float assetSeverityMetric) {
        this.assetSeverityMetric = assetSeverityMetric;
    }

    public UtmAssetTypes getAssetType() {
        return assetType;
    }

    public void setAssetType(UtmAssetTypes assetType) {
        this.assetType = assetType;
    }

    public String getAssetNotes() {
        return assetNotes;
    }

    public void setAssetNotes(String assetNotes) {
        this.assetNotes = assetNotes;
    }

    public Instant getDiscoveredAt() {
        return discoveredAt;
    }

    public void setDiscoveredAt(Instant discoveredAt) {
        this.discoveredAt = discoveredAt;
    }

    public Instant getModifiedAt() {
        return modifiedAt;
    }

    public void setModifiedAt(Instant modifiedAt) {
        this.modifiedAt = modifiedAt;
    }

    public Set<Port> getPorts() {
        return ports;
    }

    public void setPorts(Set<Port> ports) {
        this.ports = ports;
    }

    public Map<String, Long> getMetrics() {
        return metrics;
    }

    public String getServerName() {
        return serverName;
    }

    public void setServerName(String serverName) {
        this.serverName = serverName;
    }

    public UtmAssetGroup getGroup() {
        return group;
    }

    public void setGroup(UtmAssetGroup group) {
        this.group = group;
    }

    public AssetRegisteredMode getRegisteredMode() {
        return registeredMode;
    }


    public void setRegisteredMode(AssetRegisteredMode registeredMode) {
        this.registeredMode = registeredMode;
    }

    public Boolean getAgent() {
        return isAgent;
    }

    public void setAgent(Boolean agent) {
        isAgent = agent;
    }

    public static class Port {
        private Integer port;
        private String tcp;
        private String udp;

        public Port() {
        }

        public Port(UtmPorts utmPort) {
            this.port = utmPort.getPort();
            this.tcp = utmPort.getTcp();
            this.udp = utmPort.getUdp();
        }

        public Integer getPort() {
            return port;
        }

        public void setPort(Integer port) {
            this.port = port;
        }

        public String getTcp() {
            return tcp;
        }

        public void setTcp(String tcp) {
            this.tcp = tcp;
        }

        public String getUdp() {
            return udp;
        }

        public void setUdp(String udp) {
            this.udp = udp;
        }
    }

    public List<UtmDataInputStatus> getDataInputList() {
        return dataInputList;
    }

    public void setDataInputList(List<UtmDataInputStatus> dataInputList) {
        this.dataInputList = dataInputList;
    }
}

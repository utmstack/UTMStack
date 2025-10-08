package com.park.utmstack.service.dto.network_scan;

import com.park.utmstack.domain.UtmDataInputStatus;
import com.park.utmstack.domain.network_scan.UtmAssetGroup;
import com.park.utmstack.domain.network_scan.UtmAssetTypes;
import com.park.utmstack.domain.network_scan.UtmNetworkScan;
import com.park.utmstack.domain.network_scan.UtmPorts;
import com.park.utmstack.domain.network_scan.enums.AssetRegisteredMode;
import com.park.utmstack.domain.network_scan.enums.AssetStatus;
import com.park.utmstack.repository.UtmDataInputStatusRepository;
import lombok.Getter;
import lombok.Setter;
import org.springframework.util.CollectionUtils;

import java.time.Instant;
import java.util.*;
import java.util.stream.Collectors;
import java.util.stream.Stream;

@Getter
@Setter
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

    public NetworkScanDTO(UtmNetworkScan scan, boolean details, UtmDataInputStatusRepository utmDataInputStatusRepository) {
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
        this.dataInputList = utmDataInputStatusRepository.findByIpOrHostname(scan.getAssetIp(), scan.getAssetName());

        if (!CollectionUtils.isEmpty(scan.getMetrics()))
            scan.getMetrics().forEach(metric -> this.metrics.put(metric.getMetric(), metric.getAmount()));

        if (details) {
            if (!CollectionUtils.isEmpty(scan.getPorts()))
                this.ports = scan.getPorts().stream().map(Port::new).collect(Collectors.toSet());
        }
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

package com.park.utmstack.domain.network_scan;


import com.fasterxml.jackson.annotation.JsonIgnore;
import com.park.utmstack.domain.UtmAssetMetrics;
import com.park.utmstack.domain.UtmDataInputStatus;
import com.park.utmstack.domain.network_scan.enums.AssetRegisteredMode;
import com.park.utmstack.domain.network_scan.enums.AssetStatus;
import com.park.utmstack.domain.network_scan.enums.UpdateLevel;
import com.park.utmstack.service.dto.agent_manager.AgentDTO;
import com.park.utmstack.service.dto.agent_manager.AgentStatusEnum;
import com.park.utmstack.service.dto.network_scan.NetworkScanDTO;
import org.apache.http.conn.util.InetAddressUtils;

import javax.persistence.*;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.time.Instant;
import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.util.ArrayList;
import java.util.List;
import java.util.Objects;

/**
 * A UtmNetworkScan.
 */
@Entity
@Table(name = "utm_network_scan")
public class UtmNetworkScan implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Size(max = 255)
    @Column(name = "asset_ip")
    private String assetIp;

    @Column(name = "asset_addresses")
    private String assetAddresses;

    @Size(max = 255)
    @Column(name = "asset_mac")
    private String assetMac;

    @Size(max = 255)
    @Column(name = "asset_os")
    private String assetOs;

    @Size(max = 100)
    @Column(name = "asset_os_arch", length = 100)
    private String assetOsArch;

    @Size(max = 20)
    @Column(name = "asset_os_major_version", length = 20)
    private String assetOsMajorVersion;

    @Size(max = 20)
    @Column(name = "asset_os_minor_version", length = 20)
    private String assetOsMinorVersion;

    @Size(max = 100)
    @Column(name = "asset_os_platform", length = 100)
    private String assetOsPlatform;

    @Size(max = 100)
    @Column(name = "asset_os_version", length = 100)
    private String assetOsVersion;

    @Size(max = 255)
    @Column(name = "asset_name")
    private String assetName;

    @Size(max = 500)
    @Column(name = "asset_aliases", length = 500)
    private String assetAliases;

    @Column(name = "asset_alias")
    private String assetAlias;

    @Column(name = "asset_alive")
    private Boolean assetAlive;

    @Enumerated(EnumType.STRING)
    @Column(name = "asset_status")
    private AssetStatus assetStatus;

    @Column(name = "asset_severity_metric")
    private Float assetSeverityMetric;

    @JsonIgnore
    @Column(name = "asset_type_id")
    private Long assetTypeId;

    @Column(name = "discovered_at")
    private Instant discoveredAt;

    @Column(name = "modified_at")
    private Instant modifiedAt;

    @Column(name = "asset_notes")
    private String assetNotes;

    @Size(max = 255)
    @Column(name = "server_name")
    private String serverName;

    @Column(name = "group_id")
    private Long groupId;

    @Enumerated(EnumType.STRING)
    @Column(name = "registered_mode")
    private AssetRegisteredMode registeredMode;

    @Column(name = "is_agent")
    private Boolean isAgent;

    @Size(max = 50)
    @Column(name = "register_ip", length = 50)
    private String registerIp;

    @Enumerated(EnumType.STRING)
    @Column(name = "update_level")
    private UpdateLevel updateLevel;

    @OneToOne
    @JoinColumn(name = "asset_type_id", referencedColumnName = "id", insertable = false, updatable = false)
    private UtmAssetTypes assetType;

    @OneToMany(mappedBy = "scanId", fetch = FetchType.LAZY)
    private List<UtmPorts> ports = new ArrayList<>();

    @Transient
    private List<UtmAssetMetrics> metrics;
    @Transient
    private List<UtmDataInputStatus> dataInputList;
    @OneToOne
    @JoinColumn(name = "group_id", referencedColumnName = "id", insertable = false, updatable = false)
    private UtmAssetGroup assetGroup;

    public UtmNetworkScan() {
    }

    public UtmNetworkScan(String hostOrIp, Boolean alive) {
        if (InetAddressUtils.isIPv4Address(hostOrIp) || InetAddressUtils.isIPv6Address(hostOrIp))
            this.assetIp = hostOrIp;
        else
            this.assetName = hostOrIp;

        this.registeredMode = AssetRegisteredMode.DYNAMIC;
        this.updateLevel = UpdateLevel.DATASOURCE;
        this.assetStatus = AssetStatus.NEW;
        this.discoveredAt = LocalDateTime.now().toInstant(ZoneOffset.UTC);
        this.assetSeverityMetric = -1F;
        this.assetAlive = alive;
    }

    public UtmNetworkScan(NetworkScanDTO assetDto) {
        this.id = assetDto.getId();
        this.assetIp = assetDto.getAssetIp();
        this.assetAddresses = assetDto.getAssetAddresses();
        this.assetMac = assetDto.getAssetMac();
        this.assetOs = assetDto.getAssetOs();
        this.assetOsArch = assetDto.getAssetOsArch();
        this.assetOsMajorVersion = assetDto.getAssetOsMajorVersion();
        this.assetOsMinorVersion = assetDto.getAssetOsMinorVersion();
        this.assetOsPlatform = assetDto.getAssetOsPlatform();
        this.assetOsVersion = assetDto.getAssetOsVersion();
        this.assetName = assetDto.getAssetName();
        this.assetAliases = assetDto.getAssetAliases();
        this.assetAlive = assetDto.getAssetAlive();
        this.assetNotes = assetDto.getAssetNotes();
        this.assetAlias = assetDto.getAssetAlias();

        if (assetDto.getId() == null) {
            this.discoveredAt = LocalDateTime.now().toInstant(ZoneOffset.UTC);
            this.assetStatus = AssetStatus.CHECK;
            this.registeredMode = AssetRegisteredMode.CUSTOM;
        } else {
            this.discoveredAt = assetDto.getDiscoveredAt();
            this.assetStatus = assetDto.getAssetStatus();
            this.registeredMode = assetDto.getRegisteredMode();
        }

        if (assetDto.getAssetType() != null)
            this.assetTypeId = assetDto.getAssetType().getId();

        if (assetDto.getGroup() != null)
            this.groupId = assetDto.getGroup().getId();
    }

    public UtmNetworkScan(AgentDTO agent) {
        this.assetIp = agent.getIp();
        this.updateLevel = UpdateLevel.AGENT;
        this.registerIp = agent.getIp();
        this.isAgent = true;

        this.assetOs = agent.getOs();
        this.assetOsPlatform = agent.getPlatform();
        this.assetOsVersion = agent.getVersion();
        this.assetOsMinorVersion = agent.getOsMinorVersion();
        this.assetOsMajorVersion = agent.getOsMajorVersion();

        this.assetName = agent.getHostname();
        this.assetAlive = agent.getStatus() == AgentStatusEnum.ONLINE;
        this.registeredMode = AssetRegisteredMode.DYNAMIC;
        this.assetStatus = AssetStatus.NEW;
        this.discoveredAt = LocalDateTime.now().toInstant(ZoneOffset.UTC);
        this.assetSeverityMetric = -1F;
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

    public UtmNetworkScan assetIp(String assetIp) {
        this.assetIp = assetIp;
        return this;
    }

    public String getAssetAddresses() {
        return assetAddresses;
    }

    public UtmNetworkScan assetAddresses(String assetAddresses) {
        this.assetAddresses = assetAddresses;
        return this;
    }

    public String getAssetMac() {
        return assetMac;
    }

    public UtmNetworkScan assetMac(String assetMac) {
        this.assetMac = assetMac;
        return this;
    }

    public String getAssetOs() {
        return assetOs;
    }

    public UtmNetworkScan assetOs(String assetOs) {
        this.assetOs = assetOs;
        return this;
    }

    public String getAssetOsArch() {
        return assetOsArch;
    }

    public UtmNetworkScan assetOsArch(String assetOsArch) {
        this.assetOsArch = assetOsArch;
        return this;
    }

    public String getAssetOsMajorVersion() {
        return assetOsMajorVersion;
    }

    public UtmNetworkScan assetOsMajorVersion(String assetOsMajorVersion) {
        this.assetOsMajorVersion = assetOsMajorVersion;
        return this;
    }

    public String getAssetOsMinorVersion() {
        return assetOsMinorVersion;
    }

    public UtmNetworkScan assetOsMinorVersion(String assetOsMinorVersion) {
        this.assetOsMinorVersion = assetOsMinorVersion;
        return this;
    }

    public String getAssetOsPlatform() {
        return assetOsPlatform;
    }

    public UtmNetworkScan assetOsPlatform(String assetOsPlatform) {
        this.assetOsPlatform = assetOsPlatform;
        return this;
    }

    public String getAssetOsVersion() {
        return assetOsVersion;
    }

    public UtmNetworkScan assetOsVersion(String assetOsVersion) {
        this.assetOsVersion = assetOsVersion;
        return this;
    }

    public String getAssetAlias() {
        return assetAlias;
    }

    public UtmNetworkScan assetAlias(String assetAlias) {
        this.assetAlias = assetAlias;
        return this;
    }

    public String getAssetName() {
        return assetName;
    }

    public UtmNetworkScan assetName(String assetName) {
        this.assetName = assetName;
        return this;
    }

    public String getAssetAliases() {
        return assetAliases;
    }

    public UtmNetworkScan assetAliases(String assetAliases) {
        this.assetAliases = assetAliases;
        return this;
    }

    public Boolean getAssetAlive() {
        return assetAlive;
    }

    public UtmNetworkScan assetAlive(Boolean assetAlive) {
        this.assetAlive = assetAlive;
        return this;
    }

    public AssetStatus getAssetStatus() {
        return assetStatus;
    }

    public UtmNetworkScan assetStatus(AssetStatus assetStatus) {
        this.assetStatus = assetStatus;
        return this;
    }

    public Float getAssetSeverityMetric() {
        return assetSeverityMetric;
    }

    public void setAssetSeverityMetric(Float assetSeverityMetric) {
        this.assetSeverityMetric = assetSeverityMetric;
    }

    public UtmNetworkScan assetSeverityMetric(Float assetSeverityMetric) {
        this.assetSeverityMetric = assetSeverityMetric;
        return this;
    }

    public Long getAssetTypeId() {
        return assetTypeId;
    }

    public void setAssetTypeId(Long assetTypeId) {
        this.assetTypeId = assetTypeId;
    }

    public Instant getDiscoveredAt() {
        return discoveredAt;
    }

    public UtmNetworkScan discoveredAt(Instant discoveredAt) {
        this.discoveredAt = discoveredAt;
        return this;
    }

    public Instant getModifiedAt() {
        return modifiedAt;
    }

    public UtmNetworkScan modifiedAt(Instant modifiedAt) {
        this.modifiedAt = modifiedAt;
        return this;
    }

    public List<UtmPorts> getPorts() {
        return ports;
    }

    public void setPorts(List<UtmPorts> ports) {
        this.ports = ports;
    }

    public UtmNetworkScan ports(List<UtmPorts> ports) {
        this.ports = ports;
        return this;
    }

    public String getAssetNotes() {
        return assetNotes;
    }

    public void setAssetNotes(String assetNotes) {
        this.assetNotes = assetNotes;
    }

    public UtmAssetTypes getAssetType() {
        return assetType;
    }

    public void setAssetType(UtmAssetTypes assetType) {
        this.assetType = assetType;
    }

    public List<UtmAssetMetrics> getMetrics() {
        return metrics;
    }

    public void setMetrics(List<UtmAssetMetrics> metrics) {
        this.metrics = metrics;
    }

    public String getServerName() {
        return serverName;
    }

    public void setServerName(String serverName) {
        this.serverName = serverName;
    }

    public UtmNetworkScan serverName(String serverName) {
        this.serverName = serverName;
        return this;
    }

    public Long getGroupId() {
        return groupId;
    }

    public void setGroupId(Long groupId) {
        this.groupId = groupId;
    }

    public UtmAssetGroup getAssetGroup() {
        return assetGroup;
    }

    public void setAssetGroup(UtmAssetGroup assetGroup) {
        this.assetGroup = assetGroup;
    }

    public AssetRegisteredMode getRegisteredMode() {
        return registeredMode;
    }

    public void setRegisteredMode(AssetRegisteredMode registeredMode) {
        this.registeredMode = registeredMode;
    }

    public UtmNetworkScan registeredMode(AssetRegisteredMode registeredMode) {
        this.registeredMode = registeredMode;
        return this;
    }

    public Boolean getIsAgent() {
        return isAgent != null && isAgent;
    }

    public void setIsAgent(Boolean isAgent) {
        this.isAgent = isAgent;
    }

    public UtmNetworkScan isAgent(Boolean isAgent) {
        this.isAgent = isAgent;
        return this;
    }

    public String getRegisterIp() {
        return registerIp;
    }

    public void setRegisterIp(String registerIp) {
        this.registerIp = assetAddresses;
    }

    public UtmNetworkScan registerIp(String registerIp) {
        this.registerIp = registerIp;
        return this;
    }

    public UpdateLevel getUpdateLevel() {
        return updateLevel;
    }

    public void setUpdateLevel(UpdateLevel updateLevel) {
        this.updateLevel = updateLevel;
    }

    public UtmNetworkScan updateLevel(UpdateLevel updateLevel) {
        this.updateLevel = updateLevel;
        return this;
    }

    @Override
    public int hashCode() {
        return Objects.hash(assetIp, assetOs, assetName, assetAlive);
    }

    public int getUUID() {
        return Objects.hash(assetIp, assetMac);
    }

    public List<UtmDataInputStatus> getDataInputList() {
        return dataInputList;
    }

    public void setDataInputList(List<UtmDataInputStatus> dataInputList) {
        this.dataInputList = dataInputList;
    }
}



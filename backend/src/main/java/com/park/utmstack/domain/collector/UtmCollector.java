package com.park.utmstack.domain.collector;

import com.park.utmstack.domain.network_scan.UtmAssetGroup;

import javax.persistence.*;
import java.time.LocalDateTime;

@Entity
@Table(name = "utm_collectors")
public class UtmCollector {

    @Id
    private Long id;

    @Column(name = "status", nullable = false)
    private String status;

    @Column(name = "collector_key", nullable = false, length = 255)
    private String collectorKey;

    @Column(name = "ip", nullable = false, length = 45)
    private String ip;

    @Column(name = "hostname", nullable = false, length = 255)
    private String hostname;

    @Column(name = "version", length = 50)
    private String version;

    @Column(name = "module", nullable = false)
    private String module;

    @Column(name = "last_seen")
    private LocalDateTime lastSeen;

    @Column(name = "group_id")
    private Long groupId;

    @OneToOne
    @JoinColumn(name = "group_id", referencedColumnName = "id", insertable = false, updatable = false)
    private UtmAssetGroup assetGroup;

    private boolean active;


    public UtmCollector(Long id, String status, String collectorKey, String ip, String hostname, String version, String module, LocalDateTime lastSeen) {
        this.id = id;
        this.status = status;
        this.collectorKey = collectorKey;
        this.ip = ip;
        this.hostname = hostname;
        this.version = version;
        this.module = module;
        this.lastSeen = lastSeen;
    }

    public UtmCollector() {

    }

    // Getters and Setters
    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }

    public String getCollectorKey() {
        return collectorKey;
    }

    public void setCollectorKey(String collectorKey) {
        this.collectorKey = collectorKey;
    }

    public String getIp() {
        return ip;
    }
    public String getHostname() {
        return hostname;
    }

    public void setHostname(String hostname) {
        this.hostname = hostname;
    }

    public String getVersion() {
        return version;
    }

    public void setVersion(String version) {
        this.version = version;
    }

    public String getModule() {
        return module;
    }

    public void setModule(String module) {
        this.module = module;
    }

    public LocalDateTime getLastSeen() {
        return lastSeen;
    }

    public void setLastSeen(LocalDateTime lastSeen) {
        this.lastSeen = lastSeen;
    }

    public void setIp(String ip) {
        this.ip = ip;
    }

    public UtmAssetGroup getAssetGroup() {
        return assetGroup;
    }

    public void setAssetGroup(UtmAssetGroup assetGroup) {
        this.assetGroup = assetGroup;
    }

    public Long getGroupId() {
        return groupId;
    }

    public void setGroupId(Long groupId) {
        this.groupId = groupId;
    }

    public boolean isActive() {
        return active;
    }

    public void setActive(boolean active) {
        this.active = active;
    }
}


package com.park.utmstack.service.dto.collectors.dto;


import agent.CollectorOuterClass.Collector;
import com.park.utmstack.domain.collector.UtmCollector;
import com.park.utmstack.domain.network_scan.UtmAssetGroup;
import com.park.utmstack.service.dto.collectors.CollectorModuleEnum;
import com.park.utmstack.service.dto.collectors.CollectorStatusEnum;

public class CollectorDTO {
    private int id;
    private CollectorStatusEnum status;
    private String collectorKey;
    private String ip;
    private String hostname;
    private String version;
    private CollectorModuleEnum module;
    private String lastSeen;

    private String groupId;

    private UtmAssetGroup group;

    private boolean active;

    public CollectorDTO(){}


    public CollectorDTO(UtmCollector collector) {
        this.id = collector.getId().intValue();
        this.status = CollectorStatusEnum.valueOf(collector.getStatus());
        this.collectorKey = collector.getCollectorKey();
        this.ip = collector.getIp();
        this.hostname = collector.getHostname();
        this.version = collector.getVersion();
        this.module = CollectorModuleEnum.valueOf(collector.getModule());
        this.lastSeen = collector.getLastSeen().toString();
        this.group = collector.getAssetGroup();
        this.active = collector.isActive();
    }

    public int getId() {
        return id;
    }

    public void setId(int id) {
        this.id = id;
    }

    public CollectorStatusEnum getStatus() {
        return status;
    }

    public void setStatus(CollectorStatusEnum status) {
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

    public void setIp(String ip) {
        this.ip = ip;
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

    public CollectorModuleEnum getModule() {
        return module;
    }

    public void setModule(CollectorModuleEnum module) {
        this.module = module;
    }

    public String getLastSeen() {
        return lastSeen;
    }

    public void setLastSeen(String lastSeen) {
        this.lastSeen = lastSeen;
    }

    public String getGroupId() {
        return groupId;
    }

    public void setGroupId(String groupId) {
        this.groupId = groupId;
    }

    public UtmAssetGroup getGroup() {
        return group;
    }

    public void setGroup(UtmAssetGroup group) {
        this.group = group;
    }

    public boolean isActive() {
        return active;
    }

    public void setActive(boolean active) {
        this.active = active;
    }
}

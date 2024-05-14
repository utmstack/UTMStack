package com.park.utmstack.service.dto.collectors;


import agent.CollectorOuterClass.Collector;

public class CollectorDTO {
    private int id;
    private CollectorStatusEnum status;
    private String collectorKey;
    private String ip;
    private String hostname;
    private String version;
    private CollectorModuleEnum module;
    private String lastSeen;

    public CollectorDTO(){}

    public CollectorDTO(Collector collector) {
        this.id = collector.getId();
        this.status = CollectorStatusEnum.valueOf(collector.getStatus().toString());
        this.collectorKey = collector.getCollectorKey();
        this.ip = collector.getIp();
        this.hostname = collector.getHostname();
        this.version = collector.getVersion();
        this.module = CollectorModuleEnum.valueOf(collector.getModule().toString());
        this.lastSeen = collector.getLastSeen();
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
}

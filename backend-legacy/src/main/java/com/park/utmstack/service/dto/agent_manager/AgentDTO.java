package com.park.utmstack.service.dto.agent_manager;


import com.park.utmstack.service.grpc.Agent;

public class AgentDTO {
    private String ip;
    private String hostname;
    private String os;
    private AgentStatusEnum status;
    private String platform;
    private String version;
    private String agentKey;
    private int id;
    private String lastSeen;

    private String mac;
    private String osMajorVersion;
    private String osMinorVersion;
    private String aliases;
    private String addresses;


    public AgentDTO(Agent agent) {
        this.ip = agent.getIp();
        this.hostname = agent.getHostname();
        this.os = agent.getOs();
        this.status = AgentStatusEnum.valueOf(agent.getStatus().toString());
        this.platform = agent.getPlatform();
        this.version = agent.getVersion();
        this.agentKey = agent.getAgentKey();
        this.id = agent.getId();
        this.lastSeen = agent.getLastSeen();
        this.mac = agent.getMac();
        this.osMajorVersion = agent.getOsMajorVersion();
        this.osMinorVersion = agent.getOsMinorVersion();
        this.aliases = agent.getAliases();
        this.addresses = agent.getAddresses();
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

    public String getOs() {
        return os;
    }

    public void setOs(String os) {
        this.os = os;
    }

    public AgentStatusEnum getStatus() {
        return status;
    }

    public void setStatus(AgentStatusEnum status) {
        this.status = status;
    }

    public String getPlatform() {
        return platform;
    }

    public void setPlatform(String platform) {
        this.platform = platform;
    }

    public String getVersion() {
        return version;
    }

    public void setVersion(String version) {
        this.version = version;
    }

    public String getAgentKey() {
        return agentKey;
    }

    public void setAgentKey(String agentKey) {
        this.agentKey = agentKey;
    }

    public int getId() {
        return id;
    }

    public void setId(int id) {
        this.id = id;
    }

    public String getLastSeen() {
        return lastSeen;
    }

    public void setLastSeen(String lastSeen) {
        this.lastSeen = lastSeen;
    }

    public String getMac() {
        return mac;
    }

    public void setMac(String mac) {
        this.mac = mac;
    }

    public String getOsMajorVersion() {
        return osMajorVersion;
    }

    public void setOsMajorVersion(String osMajorVersion) {
        this.osMajorVersion = osMajorVersion;
    }

    public String getOsMinorVersion() {
        return osMinorVersion;
    }

    public void setOsMinorVersion(String osMinorVersion) {
        this.osMinorVersion = osMinorVersion;
    }

    public String getAliases() {
        return aliases;
    }

    public void setAliases(String aliases) {
        this.aliases = aliases;
    }

    public String getAddresses() {
        return addresses;
    }

    public void setAddresses(String addresses) {
        this.addresses = addresses;
    }
}




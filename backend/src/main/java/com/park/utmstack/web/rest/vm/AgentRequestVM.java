package com.park.utmstack.web.rest.vm;

import com.park.utmstack.service.grpc.AgentRequest;

import javax.validation.constraints.Min;
import javax.validation.constraints.NotEmpty;
import javax.validation.constraints.NotNull;

public class AgentRequestVM {
    @NotEmpty
    private String ip;
    @NotEmpty
    private String hostname;
    private String os;
    private String platform;
    private String version;
    @NotEmpty
    private String mac;
    private String osMajorVersion;
    private String osMinorVersion;
    private String aliases;
    private String addresses;
    @NotEmpty
    private String agentKey;
    @Min(1)
    private int id;


    public AgentRequestVM() {}

    public AgentRequest getAgentRequest() {
        return AgentRequest.newBuilder()
                .setIp(this.ip)
                .setHostname(this.hostname)
                .setOs(this.os)
                .setPlatform(this.platform)
                .setVersion(this.version)
                .setMac(this.mac)
                .setOsMajorVersion(this.osMajorVersion)
                .setOsMinorVersion(this.osMinorVersion)
                .setAliases(this.aliases)
                .setAddresses(this.addresses)
                .build();
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
}

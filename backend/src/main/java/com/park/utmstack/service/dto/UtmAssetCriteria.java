package com.park.utmstack.service.dto;

import tech.jhipster.service.filter.*;

import java.io.Serializable;

/**
 * Criteria class for the UtmAsset entity. This class is used in UtmAssetResource to receive all the possible filtering
 * options from the Http GET request parameters. For example the following could be a valid requests:
 * <code> /utm-assets?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use fix type
 * specific filters.
 */
public class UtmAssetCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;

    private StringFilter hostname;

    private StringFilter mac;

    private StringFilter ip;

    private IntegerFilter filebeat;

    private IntegerFilter winlogbeat;

    private IntegerFilter hids;

    private StringFilter agentVersion;

    private InstantFilter lastSeen;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public StringFilter getHostname() {
        return hostname;
    }

    public void setHostname(StringFilter hostname) {
        this.hostname = hostname;
    }

    public StringFilter getMac() {
        return mac;
    }

    public void setMac(StringFilter mac) {
        this.mac = mac;
    }

    public StringFilter getIp() {
        return ip;
    }

    public void setIp(StringFilter ip) {
        this.ip = ip;
    }

    public IntegerFilter getFilebeat() {
        return filebeat;
    }

    public void setFilebeat(IntegerFilter filebeat) {
        this.filebeat = filebeat;
    }

    public IntegerFilter getWinlogbeat() {
        return winlogbeat;
    }

    public void setWinlogbeat(IntegerFilter winlogbeat) {
        this.winlogbeat = winlogbeat;
    }

    public IntegerFilter getHids() {
        return hids;
    }

    public void setHids(IntegerFilter hids) {
        this.hids = hids;
    }

    public StringFilter getAgentVersion() {
        return agentVersion;
    }

    public void setAgentVersion(StringFilter agentVersion) {
        this.agentVersion = agentVersion;
    }

    public InstantFilter getLastSeen() {
        return lastSeen;
    }

    public void setLastSeen(InstantFilter lastSeen) {
        this.lastSeen = lastSeen;
    }
}

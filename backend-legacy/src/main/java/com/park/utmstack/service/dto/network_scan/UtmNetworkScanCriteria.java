package com.park.utmstack.service.dto.network_scan;

import com.park.utmstack.domain.network_scan.enums.AssetStatus;
import tech.jhipster.service.filter.*;

import java.io.Serializable;

/**
 * Criteria class for the UtmNetworkScan entity. This class is used in UtmNetworkScanResource to
 * receive all the possible filtering options from the Http GET request parameters.
 * For example the following could be a valid requests:
 * <code> /utm-network-scans?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use
 * fix type specific filters.
 */
public class UtmNetworkScanCriteria implements Serializable {
    /**
     * Class for filtering AssetStatus
     */
    public static class AssetStatusFilter extends Filter<AssetStatus> {
    }

    private static final long serialVersionUID = 1L;

    private LongFilter id;

    private StringFilter ip;

    private StringFilter mac;

    private StringFilter os;

    private StringFilter name;

    private BooleanFilter alive;

    private AssetStatusFilter status;

    private InstantFilter discoveredAt;

    private InstantFilter modifiedAt;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public StringFilter getIp() {
        return ip;
    }

    public void setIp(StringFilter ip) {
        this.ip = ip;
    }

    public StringFilter getMac() {
        return mac;
    }

    public void setMac(StringFilter mac) {
        this.mac = mac;
    }

    public StringFilter getOs() {
        return os;
    }

    public void setOs(StringFilter os) {
        this.os = os;
    }

    public StringFilter getName() {
        return name;
    }

    public void setName(StringFilter name) {
        this.name = name;
    }

    public BooleanFilter getAlive() {
        return alive;
    }

    public void setAlive(BooleanFilter alive) {
        this.alive = alive;
    }

    public AssetStatusFilter getStatus() {
        return status;
    }

    public void setStatus(AssetStatusFilter status) {
        this.status = status;
    }

    public InstantFilter getDiscoveredAt() {
        return discoveredAt;
    }

    public void setDiscoveredAt(InstantFilter discoveredAt) {
        this.discoveredAt = discoveredAt;
    }

    public InstantFilter getModifiedAt() {
        return modifiedAt;
    }

    public void setModifiedAt(InstantFilter modifiedAt) {
        this.modifiedAt = modifiedAt;
    }
}

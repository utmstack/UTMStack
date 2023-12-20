package com.park.utmstack.domain.network_scan;

import java.time.Instant;
import java.util.List;

public class AssetGroupFilter {
    private Long id;
    private String groupName;
    private List<String> type;
    private Instant initDate;
    private Instant endDate;
    private List<String> probe;
    private List<String> os;
    private List<String> assetIp;
    private List<String> assetName;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getGroupName() {
        return groupName;
    }

    public void setGroupName(String groupName) {
        this.groupName = groupName;
    }

    public Instant getInitDate() {
        return initDate;
    }

    public void setInitDate(Instant initDate) {
        this.initDate = initDate;
    }

    public Instant getEndDate() {
        return endDate;
    }

    public void setEndDate(Instant endDate) {
        this.endDate = endDate;
    }

    public List<String> getAssetIp() {
        return assetIp;
    }

    public void setAssetIp(List<String> assetIp) {
        this.assetIp = assetIp;
    }

    public List<String> getAssetName() {
        return assetName;
    }

    public void setAssetName(List<String> assetName) {
        this.assetName = assetName;
    }

    public List<String> getType() {
        return type;
    }

    public void setType(List<String> type) {
        this.type = type;
    }

    public List<String> getProbe() {
        return probe;
    }

    public void setProbe(List<String> probe) {
        this.probe = probe;
    }

    public List<String> getOs() {
        return os;
    }

    public void setOs(List<String> os) {
        this.os = os;
    }
}

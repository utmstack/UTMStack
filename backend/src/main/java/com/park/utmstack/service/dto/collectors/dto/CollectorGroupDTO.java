package com.park.utmstack.service.dto.collectors.dto;

import com.park.utmstack.domain.network_scan.UtmAssetGroup;
import org.springframework.util.CollectionUtils;

import java.time.Instant;
import java.util.HashMap;
import java.util.Map;

public class CollectorGroupDTO {
    private Long id;
    private String groupName;
    private String groupDescription;
    private Instant createdDate;
    private int assetsCount = 0;

    public CollectorGroupDTO() {
    }

    public CollectorGroupDTO(UtmAssetGroup group) {
        this.id = group.getId();
        this.groupName = group.getGroupName();
        this.groupDescription = group.getGroupDescription();
        this.createdDate = group.getCreatedDate();
        this.assetsCount = group.getAssets().size();
    }

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

    public String getGroupDescription() {
        return groupDescription;
    }

    public void setGroupDescription(String groupDescription) {
        this.groupDescription = groupDescription;
    }

    public Instant getCreatedDate() {
        return createdDate;
    }

    public void setCreatedDate(Instant createdDate) {
        this.createdDate = createdDate;
    }

    public Integer getAssetsCount() {
        return assetsCount;
    }

    public void setAssetsCount(Integer assetsCount) {
        this.assetsCount = assetsCount;
    }
}

package com.park.utmstack.service.dto.network_scan;

import com.park.utmstack.domain.network_scan.UtmAssetGroup;
import org.springframework.util.CollectionUtils;

import java.time.Instant;
import java.util.HashMap;
import java.util.Map;

public class AssetGroupDTO {
    private Long id;
    private String groupName;
    private String groupDescription;
    private Instant createdDate;
    private int assetsCount = 0;
    private Map<String, Long> metrics;

    public AssetGroupDTO() {
    }

    public AssetGroupDTO(UtmAssetGroup group) {
        this.id = group.getId();
        this.groupName = group.getGroupName();
        this.groupDescription = group.getGroupDescription();
        this.createdDate = group.getCreatedDate();
        this.assetsCount = group.getAssets().size();

        if (!CollectionUtils.isEmpty(group.getMetrics())) {
            this.metrics = new HashMap<>();
            group.getMetrics().forEach(metric -> {
                metrics.putIfAbsent(metric.getMetric(), 0L);
                metrics.put(metric.getMetric(), metrics.get(metric.getMetric()) + metric.getAmount());
            });
        }
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

    public Map<String, Long> getMetrics() {
        return metrics;
    }

    public void setMetrics(Map<String, Long> metrics) {
        this.metrics = metrics;
    }
}

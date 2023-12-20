package com.park.utmstack.domain.network_scan;


import com.park.utmstack.domain.UtmAssetMetrics;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.time.Instant;
import java.util.ArrayList;
import java.util.List;

/**
 * A UtmAssetGroup.
 */
@Entity
@Table(name = "utm_asset_group")
public class UtmAssetGroup implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @NotNull
    @Size(max = 100)
    @Column(name = "group_name", length = 100, nullable = false, unique = true)
    private String groupName;

    @Column(name = "group_description")
    private String groupDescription;

    @Column(name = "created_date")
    private Instant createdDate;

    @Transient
    private List<UtmAssetMetrics> metrics;

    @Transient
    private List<UtmNetworkScan> assets = new ArrayList<>();

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getGroupName() {
        return groupName;
    }

    public UtmAssetGroup groupName(String groupName) {
        this.groupName = groupName;
        return this;
    }

    public void setGroupName(String groupName) {
        this.groupName = groupName;
    }

    public String getGroupDescription() {
        return groupDescription;
    }

    public UtmAssetGroup groupDescription(String groupDescription) {
        this.groupDescription = groupDescription;
        return this;
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

    public List<UtmAssetMetrics> getMetrics() {
        return metrics;
    }

    public void setMetrics(List<UtmAssetMetrics> metrics) {
        this.metrics = metrics;
    }

    public List<UtmNetworkScan> getAssets() {
        return assets;
    }

    public void setAssets(List<UtmNetworkScan> assets) {
        this.assets = assets;
    }
}

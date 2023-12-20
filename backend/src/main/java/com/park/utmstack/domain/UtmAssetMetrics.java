package com.park.utmstack.domain;


import javax.persistence.Column;
import javax.persistence.Entity;
import javax.persistence.Id;
import javax.persistence.Table;
import javax.validation.constraints.NotNull;
import java.io.Serializable;

/**
 * A UtmAssetMetrics.
 */
@Entity
@Table(name = "utm_asset_metrics")
public class UtmAssetMetrics implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    private String id;

    @NotNull
    @Column(name = "asset_name", nullable = false)
    private String assetName;

    @NotNull
    @Column(name = "metric", nullable = false)
    private String metric;

    @NotNull
    @Column(name = "amount", nullable = false)
    private Long amount;

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getAssetName() {
        return assetName;
    }

    public UtmAssetMetrics assetName(String assetName) {
        this.assetName = assetName;
        return this;
    }

    public void setAssetName(String assetName) {
        this.assetName = assetName;
    }

    public String getMetric() {
        return metric;
    }

    public UtmAssetMetrics metric(String metric) {
        this.metric = metric;
        return this;
    }

    public void setMetric(String metric) {
        this.metric = metric;
    }

    public Long getAmount() {
        return amount;
    }

    public UtmAssetMetrics amount(Long amount) {
        this.amount = amount;
        return this;
    }

    public void setAmount(Long amount) {
        this.amount = amount;
    }
}

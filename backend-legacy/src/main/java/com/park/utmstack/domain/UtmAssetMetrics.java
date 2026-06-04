package com.park.utmstack.domain;


import com.fasterxml.jackson.annotation.JsonIgnore;
import com.park.utmstack.domain.network_scan.UtmNetworkScan;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.RequiredArgsConstructor;
import lombok.Setter;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import java.io.Serializable;

/**
 * A UtmAssetMetrics.
 */
@Setter
@Getter
@NoArgsConstructor
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

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "asset_name", referencedColumnName = "asset_name", insertable = false, updatable = false, nullable = false)
    @JsonIgnore
    private UtmNetworkScan asset;
}

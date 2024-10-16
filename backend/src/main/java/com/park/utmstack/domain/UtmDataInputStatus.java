package com.park.utmstack.domain;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.park.utmstack.domain.network_scan.UtmNetworkScan;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

import javax.persistence.*;
import javax.validation.constraints.NotEmpty;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.util.Objects;
import java.util.concurrent.TimeUnit;

/**
 * A UtmDataInputStatus.
 */
@Setter
@Getter
@NoArgsConstructor
@Entity
@Table(name = "utm_data_input_status")
public class UtmDataInputStatus implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    private String id;

    @NotNull
    @Size(max = 256)
    @Column(name = "source", length = 256, nullable = false)
    private String source;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "source", referencedColumnName = "asset_name", insertable = false, updatable = false, nullable = false)
    @JsonIgnore
    private UtmNetworkScan assetName;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "source", referencedColumnName = "asset_ip", insertable = false, updatable = false, nullable = false)
    @JsonIgnore
    private UtmNetworkScan assetIp;

    @NotNull
    @Size(max = 50)
    @Column(name = "data_type", length = 50, nullable = false)
    private String dataType;

    @NotNull
    @Column(name = "timestamp", nullable = false)
    private Long timestamp;

    @Column(name = "median")
    private Long median;

    /**
     * Define if a source is down or up.
     * Null is returned when the calculation could not be done.
     *
     * @return True if this source is down
     */
    public Boolean isDown() {
        if (Objects.isNull(timestamp) || Objects.isNull(median))
            return null;
        long currentTimeInSeconds = TimeUnit.MILLISECONDS.toSeconds(System.currentTimeMillis());
        return (currentTimeInSeconds - timestamp) > (median * 6);
    }
}

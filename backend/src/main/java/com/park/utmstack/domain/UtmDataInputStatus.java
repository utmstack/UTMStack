package com.park.utmstack.domain;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.park.utmstack.domain.network_scan.UtmNetworkScan;
import lombok.*;

import javax.persistence.*;
import javax.validation.constraints.NotEmpty;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.time.Instant;
import java.util.Objects;
import java.util.concurrent.TimeUnit;

/**
 * A UtmDataInputStatus.
 */
@Setter
@Getter
@NoArgsConstructor
@AllArgsConstructor
@Builder
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
            return false;
        long now = Instant.now().getEpochSecond();
        return (now - timestamp) > (median * 1.5);
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof UtmDataInputStatus that)) return false;
        return Objects.equals(id, that.id);
    }

    @Override
    public int hashCode() {
        return Objects.hash(id);
    }
}

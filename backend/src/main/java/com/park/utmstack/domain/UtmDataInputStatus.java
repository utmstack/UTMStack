package com.park.utmstack.domain;

import javax.persistence.Column;
import javax.persistence.Entity;
import javax.persistence.Id;
import javax.persistence.Table;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.util.Objects;
import java.util.concurrent.TimeUnit;

/**
 * A UtmDataInputStatus.
 */
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

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getSource() {
        return source;
    }

    public UtmDataInputStatus source(String source) {
        this.source = source;
        return this;
    }

    public void setSource(String source) {
        this.source = source;
    }

    public String getDataType() {
        return dataType;
    }

    public UtmDataInputStatus dataType(String dataType) {
        this.dataType = dataType;
        return this;
    }

    public void setDataType(String dataType) {
        this.dataType = dataType;
    }

    public Long getTimestamp() {
        return timestamp;
    }

    public UtmDataInputStatus timestamp(Long timestamp) {
        this.timestamp = timestamp;
        return this;
    }

    public void setTimestamp(Long timestamp) {
        this.timestamp = timestamp;
    }

    public Long getMedian() {
        return median;
    }

    public UtmDataInputStatus median(Long median) {
        this.median = median;
        return this;
    }

    public void setMedian(Long median) {
        this.median = median;
    }

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

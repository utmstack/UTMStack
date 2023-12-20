package com.park.utmstack.service.dto;

import tech.jhipster.service.filter.Filter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;
import java.util.Objects;

/**
 * Criteria class for the UtmDataInputStatus entity. This class is used in UtmDataInputStatusResource to
 * receive all the possible filtering options from the Http GET request parameters.
 * For example the following could be a valid requests:
 * <code> /utm-data-input-statuses?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use
 * fix type specific filters.
 */
public class UtmDataInputStatusCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private StringFilter id;

    private StringFilter source;

    private StringFilter dataType;

    private LongFilter timestamp;

    private LongFilter median;

    public StringFilter getId() {
        return id;
    }

    public void setId(StringFilter id) {
        this.id = id;
    }

    public StringFilter getSource() {
        return source;
    }

    public void setSource(StringFilter source) {
        this.source = source;
    }

    public StringFilter getDataType() {
        return dataType;
    }

    public void setDataType(StringFilter dataType) {
        this.dataType = dataType;
    }

    public LongFilter getTimestamp() {
        return timestamp;
    }

    public void setTimestamp(LongFilter timestamp) {
        this.timestamp = timestamp;
    }

    public LongFilter getMedian() {
        return median;
    }

    public void setMedian(LongFilter median) {
        this.median = median;
    }


    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        final UtmDataInputStatusCriteria that = (UtmDataInputStatusCriteria) o;
        return
            Objects.equals(id, that.id) &&
                Objects.equals(source, that.source) &&
                Objects.equals(dataType, that.dataType) &&
                Objects.equals(timestamp, that.timestamp) &&
                Objects.equals(median, that.median);
    }

    @Override
    public int hashCode() {
        return Objects.hash(
            id,
            source,
            dataType,
            timestamp,
            median
        );
    }

    @Override
    public String toString() {
        return "UtmDataInputStatusCriteria{" +
            (id != null ? "id=" + id + ", " : "") +
            (source != null ? "source=" + source + ", " : "") +
            (dataType != null ? "dataType=" + dataType + ", " : "") +
            (timestamp != null ? "timestamp=" + timestamp + ", " : "") +
            (median != null ? "median=" + median + ", " : "") +
            "}";
    }

}

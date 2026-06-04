package com.park.utmstack.domain;


import javax.persistence.*;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.util.Objects;

/**
 * A UtmSchedule.
 */
@Entity
@Table(name = "utm_schedule")
public class UtmSchedule implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @NotNull
    @Size(max = 100)
    @Column(name = "name", length = 100, nullable = false)
    private String name;

    @Size(max = 255)
    @Column(name = "jhi_comment", length = 255)
    private String comment;

    @Column(name = "first_time")
    private Long firstTime;

    @Column(name = "period")
    private Long period;

    @Column(name = "period_month")
    private Long periodMonth;

    @Column(name = "duration")
    private Long duration;

    @Column(name = "timezone")
    private String timezone;

    @Column(name = "initial_offset")
    private Long initialOffset;

    @Column(name = "creation_time")
    private Long creationTime;

    @Column(name = "modification_time")
    private Long modificationTime;

    // jhipster-needle-entity-add-field - JHipster will add fields here, do not remove
    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getName() {
        return name;
    }

    public UtmSchedule name(String name) {
        this.name = name;
        return this;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getComment() {
        return comment;
    }

    public UtmSchedule comment(String comment) {
        this.comment = comment;
        return this;
    }

    public void setComment(String comment) {
        this.comment = comment;
    }

    public Long getFirstTime() {
        return firstTime;
    }

    public UtmSchedule firstTime(Long firstTime) {
        this.firstTime = firstTime;
        return this;
    }

    public void setFirstTime(Long firstTime) {
        this.firstTime = firstTime;
    }

    public Long getPeriod() {
        return period;
    }

    public UtmSchedule period(Long period) {
        this.period = period;
        return this;
    }

    public void setPeriod(Long period) {
        this.period = period;
    }

    public Long getPeriodMonth() {
        return periodMonth;
    }

    public UtmSchedule periodMonth(Long periodMonth) {
        this.periodMonth = periodMonth;
        return this;
    }

    public void setPeriodMonth(Long periodMonth) {
        this.periodMonth = periodMonth;
    }

    public Long getDuration() {
        return duration;
    }

    public UtmSchedule duration(Long duration) {
        this.duration = duration;
        return this;
    }

    public void setDuration(Long duration) {
        this.duration = duration;
    }

    public String getTimezone() {
        return timezone;
    }

    public UtmSchedule timezone(String timezone) {
        this.timezone = timezone;
        return this;
    }

    public void setTimezone(String timezone) {
        this.timezone = timezone;
    }

    public Long getInitialOffset() {
        return initialOffset;
    }

    public UtmSchedule initialOffset(Long initialOffset) {
        this.initialOffset = initialOffset;
        return this;
    }

    public void setInitialOffset(Long initialOffset) {
        this.initialOffset = initialOffset;
    }

    public Long getCreationTime() {
        return creationTime;
    }

    public UtmSchedule creationTime(Long creationTime) {
        this.creationTime = creationTime;
        return this;
    }

    public void setCreationTime(Long creationTime) {
        this.creationTime = creationTime;
    }

    public Long getModificationTime() {
        return modificationTime;
    }

    public UtmSchedule modificationTime(Long modificationTime) {
        this.modificationTime = modificationTime;
        return this;
    }

    public void setModificationTime(Long modificationTime) {
        this.modificationTime = modificationTime;
    }
    // jhipster-needle-entity-add-getters-setters - JHipster will add getters and setters here, do not remove

    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        UtmSchedule utmSchedule = (UtmSchedule) o;
        if (utmSchedule.getId() == null || getId() == null) {
            return false;
        }
        return Objects.equals(getId(), utmSchedule.getId());
    }

    @Override
    public int hashCode() {
        return Objects.hashCode(getId());
    }

    @Override
    public String toString() {
        return "UtmSchedule{" + "id=" + getId() + ", name='" + getName() + "'" + ", comment='" + getComment() + "'" + ", firstTime=" + getFirstTime() + ", period=" + getPeriod() + ", periodMonth=" + getPeriodMonth() + ", duration=" + getDuration() + ", timezone='" + getTimezone() + "'" + ", initialOffset=" + getInitialOffset() + ", creationTime=" + getCreationTime() + ", modificationTime=" + getModificationTime() + "}";
    }
}

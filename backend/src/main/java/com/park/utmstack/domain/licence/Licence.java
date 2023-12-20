package com.park.utmstack.domain.licence;

import com.fasterxml.jackson.annotation.JsonInclude;

import java.time.Instant;
import java.time.LocalDateTime;
import java.time.ZoneOffset;
import java.time.format.DateTimeFormatter;

@JsonInclude(value = JsonInclude.Include.NON_NULL)
public class Licence {
    private String licenseKey;
    private String expiresAt;
    private Integer validFor;
    private Integer status;
    private Integer timesActivated;
    private Integer timesActivatedMax;
    private Integer remainingActivations;
    private String createdAt;
    private String updatedAt;

    public String getLicenseKey() {
        return licenseKey;
    }

    public void setLicenseKey(String licenseKey) {
        this.licenseKey = licenseKey;
    }

    public String getExpiresAt() {
        return expiresAt;
    }

    public void setExpiresAt(String expiresAt) {
        this.expiresAt = expiresAt;
    }

    public Integer getValidFor() {
        return validFor;
    }

    public void setValidFor(Integer validFor) {
        this.validFor = validFor;
    }

    public Integer getStatus() {
        return status;
    }

    public void setStatus(Integer status) {
        this.status = status;
    }

    public Integer getTimesActivated() {
        return timesActivated;
    }

    public void setTimesActivated(Integer timesActivated) {
        this.timesActivated = timesActivated;
    }

    public Integer getTimesActivatedMax() {
        return timesActivatedMax;
    }

    public void setTimesActivatedMax(Integer timesActivatedMax) {
        this.timesActivatedMax = timesActivatedMax;
    }

    public Integer getRemainingActivations() {
        return remainingActivations;
    }

    public void setRemainingActivations(Integer remainingActivations) {
        this.remainingActivations = remainingActivations;
    }

    public String getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(String createdAt) {
        this.createdAt = createdAt;
    }

    public String getUpdatedAt() {
        return updatedAt;
    }

    public void setUpdatedAt(String updatedAt) {
        this.updatedAt = updatedAt;
    }

    public Instant getCreatedAtAsInstant() {
        return LocalDateTime.parse(createdAt, DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss")).toInstant(ZoneOffset.UTC);
    }

    public Instant getExpiresAtAsInstant() {
        return LocalDateTime.parse(expiresAt, DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss")).toInstant(ZoneOffset.UTC);
    }
}

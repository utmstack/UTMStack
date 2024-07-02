package com.park.utmstack.domain.correlation.config;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.fasterxml.jackson.databind.annotation.JsonSerialize;
import com.park.utmstack.util.UtilSerializer;
import com.park.utmstack.util.exceptions.UtmSerializationException;
import org.hibernate.annotations.GenericGenerator;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import javax.persistence.*;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.time.Clock;
import java.time.Instant;
import java.util.List;

/**
 * Tenant config entity template, used by correlation rule engine.
 */
@Entity
@Table(name = "utm_tenant_config")
public class UtmTenantConfig implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GenericGenerator(name = "CustomIdentityGenerator", strategy = "com.park.utmstack.util.CustomIdentityGenerator")
    @GeneratedValue(generator = "CustomIdentityGenerator")
    @Column(name = "id", updatable = false)
    private Long id;

    @Size(max = 250)
    @Column(name = "asset_name", length = 250, nullable = false)
    private String assetName;

    @JsonIgnore
    @Column(name = "asset_hostname_list_def")
    private String assetHostnameListDef;

    @Transient
    @JsonSerialize
    @JsonDeserialize
    private List<String> assetHostnameList;

    @JsonIgnore
    @Column(name = "asset_ip_list_def")
    private String assetIpListDef;

    @Transient
    @JsonSerialize
    @JsonDeserialize
    private List<String> assetIpList;

    @Column(name = "asset_confidentiality", nullable = false)
    private Integer assetConfidentiality;

    @Column(name = "asset_integrity", nullable = false)
    private Integer assetIntegrity;

    @Column(name = "asset_availability", nullable = false)
    private Integer assetAvailability;

    @Column(name = "last_update")
    private Instant lastUpdate;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public List<String> getAssetHostnameList() throws UtmSerializationException {
        if (StringUtils.hasText(assetHostnameListDef))
            assetHostnameList = UtilSerializer.jsonDeserializeList(String.class, assetHostnameListDef);
        return assetHostnameList;
    }

    public void setAssetHostnameList(List<String> assetHostnameList) throws UtmSerializationException {
        if (CollectionUtils.isEmpty(assetHostnameList))
            this.assetHostnameListDef = null;
        else
            this.assetHostnameListDef = UtilSerializer.jsonSerialize(assetHostnameList);
        this.assetHostnameList = assetHostnameList;
    }

    public String getAssetHostnameListDef() {
        return assetHostnameListDef;
    }

    public List<String> getAssetIpList() throws UtmSerializationException {
        if (StringUtils.hasText(assetIpListDef))
            assetIpList = UtilSerializer.jsonDeserializeList(String.class, assetIpListDef);
        return assetIpList;
    }

    public void setAssetIpList(List<String> assetIpList) throws UtmSerializationException {
        if (CollectionUtils.isEmpty(assetIpList))
            this.assetIpListDef = null;
        else
            this.assetIpListDef = UtilSerializer.jsonSerialize(assetIpList);
        this.assetIpList = assetIpList;
    }

    public void setAssetHostnameListDef(String assetHostnameListDef) {
        this.assetHostnameListDef = assetHostnameListDef;
    }

    public String getAssetIpListDef() {
        return assetIpListDef;
    }

    public void setAssetIpListDef(String assetIpListDef) {
        this.assetIpListDef = assetIpListDef;
    }

    public Integer getAssetAvailability() {
        return assetAvailability;
    }

    public void setAssetAvailability(Integer assetAvailability) {
        this.assetAvailability = assetAvailability;
    }

    public Integer getAssetIntegrity() {
        return assetIntegrity;
    }

    public void setAssetIntegrity(Integer assetIntegrity) {
        this.assetIntegrity = assetIntegrity;
    }

    public Integer getAssetConfidentiality() {
        return assetConfidentiality;
    }

    public void setAssetConfidentiality(Integer assetConfidentiality) {
        this.assetConfidentiality = assetConfidentiality;
    }

    public String getAssetName() {
        return assetName;
    }

    public void setAssetName(String assetName) {
        this.assetName = assetName;
    }

    public Instant getLastUpdate() {
        return lastUpdate;
    }

    public void setLastUpdate() {
        this.lastUpdate = Instant.now(Clock.systemUTC());
    }
}

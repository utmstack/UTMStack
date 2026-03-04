package com.park.utmstack.domain.correlation.config;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.fasterxml.jackson.databind.annotation.JsonSerialize;
import com.park.utmstack.util.UtilSerializer;
import com.park.utmstack.util.exceptions.UtmSerializationException;
import lombok.Getter;
import lombok.Setter;
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

    @Setter
    @Getter
    @Id
    @GenericGenerator(name = "CustomIdentityGenerator", strategy = "com.park.utmstack.util.CustomIdentityGenerator")
    @GeneratedValue(generator = "CustomIdentityGenerator")
    @Column(name = "id", updatable = false)
    private Long id;

    @Setter
    @Getter
    @Size(max = 250)
    @Column(name = "asset_name", length = 250, nullable = false)
    private String assetName;

    @Setter
    @Getter
    @JsonIgnore
    @Column(name = "asset_hostname_list_def")
    private String assetHostnameListDef;

    @Transient
    @JsonSerialize
    @JsonDeserialize
    private List<String> assetHostnameList;

    @Setter
    @Getter
    @JsonIgnore
    @Column(name = "asset_ip_list_def")
    private String assetIpListDef;

    @Transient
    @JsonSerialize
    @JsonDeserialize
    private List<String> assetIpList;

    @Setter
    @Getter
    @Column(name = "asset_confidentiality", nullable = false)
    private Integer assetConfidentiality;

    @Setter
    @Getter
    @Column(name = "asset_integrity", nullable = false)
    private Integer assetIntegrity;

    @Setter
    @Getter
    @Column(name = "asset_availability", nullable = false)
    private Integer assetAvailability;

    @Getter
    @Column(name = "last_update")
    private Instant lastUpdate;

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

    public void setLastUpdate() {
        this.lastUpdate = Instant.now(Clock.systemUTC());
    }
}

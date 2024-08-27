package com.park.utmstack.domain.correlation.config;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.park.utmstack.domain.correlation.rules.UtmCorrelationRules;
import org.hibernate.annotations.GenericGenerator;

import javax.persistence.*;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.time.Clock;
import java.time.Instant;
import java.util.HashSet;
import java.util.Set;

/**
 * Datatypes entity template.
 */
@Entity
@Table(name = "utm_data_types")
public class UtmDataTypes implements Serializable {
    private static final long serialVersionUID = 1L;

    @Id
    @GenericGenerator(name = "CustomIdentityGenerator", strategy = "com.park.utmstack.util.CustomIdentityGenerator")
    @GeneratedValue(generator = "CustomIdentityGenerator")
    @Column(name = "id", updatable = false)
    private Long id;

    @Size(max = 250)
    @Column(name = "data_type", length = 250, nullable = false)
    private String dataType;

    @Size(max = 250)
    @Column(name = "data_type_name", length = 250, nullable = false)
    private String dataTypeName;

    @Column(name = "data_type_description")
    private String dataTypeDescription;

    @Column(name = "last_update")
    private Instant lastUpdate;

    @Column(name = "included", nullable = false)
    private Boolean included;

    @Column(name = "system_owner", nullable = false)
    private Boolean systemOwner;

    @ManyToMany(fetch = FetchType.LAZY, cascade = {
                    CascadeType.PERSIST,
                    CascadeType.MERGE
            },
            mappedBy = "dataTypes")
    @JsonIgnore
    private Set<UtmCorrelationRules> utmCorrelationRules = new HashSet<>();

    public UtmDataTypes(String dataType) {
        this.dataType = dataType;
        this.dataTypeName = dataType;
        this.dataTypeDescription = "Synchronized datatype";
        setLastUpdate();
        this.included = true;
        this.systemOwner = false;
    }

    public UtmDataTypes() {
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getDataType() {
        return dataType;
    }

    public void setDataType(String dataType) {
        this.dataType = dataType;
    }

    public String getDataTypeName() {
        return dataTypeName;
    }

    public void setDataTypeName(String dataTypeName) {
        this.dataTypeName = dataTypeName;
    }

    public String getDataTypeDescription() {
        return dataTypeDescription;
    }

    public void setDataTypeDescription(String dataTypeDescription) {
        this.dataTypeDescription = dataTypeDescription;
    }

    public Instant getLastUpdate() {
        return lastUpdate;
    }

    public void setLastUpdate() {
        this.lastUpdate = Instant.now(Clock.systemUTC());
    }

    public Boolean getIncluded() {
        return included;
    }

    public void setIncluded(Boolean included) {
        this.included = included;
    }

    public Boolean getSystemOwner() {
        return systemOwner;
    }

    public void setSystemOwner(Boolean systemOwner) {
        this.systemOwner = systemOwner;
    }
}

package com.park.utmstack.domain;

import com.park.utmstack.service.dto.UtmDataSourceConfigDTO;
import org.hibernate.annotations.GenericGenerator;

import javax.persistence.*;
import java.io.Serializable;
import java.util.UUID;

/**
 * A UtmDataSourceConfig.
 */
@Entity
@Table(name = "utm_data_source_config")
public class UtmDataSourceConfig implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(generator = "uuid-hibernate-generator")
    @GenericGenerator(name = "uuid-hibernate-generator", strategy = "org.hibernate.id.UUIDGenerator")
    @Column(name = "id", updatable = false)
    private UUID id;

    @Column(name = "data_type", updatable = false)
    private String dataType;

    @Column(name = "data_type_name", updatable = false)
    private String dataTypeName;

    @Column(name = "system_owner", updatable = false)
    private Boolean systemOwner;

    @Column(name = "included")
    private Boolean included;

    public UtmDataSourceConfig() {
    }

    public UtmDataSourceConfig(String dataType) {
        this.dataType = dataType;
        this.dataTypeName = dataType;
        this.included = true;
        this.systemOwner = false;
    }

    public UtmDataSourceConfig(UtmDataSourceConfigDTO dto) {
        this.id = UUID.fromString(dto.getId());
        this.dataType = dto.getDataType();
        this.dataTypeName = dto.getDataType();
        this.included = dto.getIncluded();
        this.systemOwner = dto.getSystemOwner();
    }

    public UUID getId() {
        return id;
    }

    public void setId(UUID id) {
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

    public Boolean getSystemOwner() {
        return systemOwner;
    }

    public void setSystemOwner(Boolean systemOwner) {
        this.systemOwner = systemOwner;
    }

    public Boolean getIncluded() {
        return included;
    }

    public void setIncluded(Boolean included) {
        this.included = included;
    }
}

package com.park.utmstack.service.dto;

import com.park.utmstack.domain.UtmDataSourceConfig;

import java.io.Serializable;
import java.util.Objects;

/**
 * A DTO for the {@link UtmDataSourceConfig} entity.
 */
public class UtmDataSourceConfigDTO implements Serializable {

    private String id;

    private String dataType;

    private String dataTypeName;

    private Boolean systemOwner;

    private Boolean included;

    public UtmDataSourceConfigDTO() {
    }

    public UtmDataSourceConfigDTO(UtmDataSourceConfig dataSourceConfig) {
        this.id = dataSourceConfig.getId().toString();
        this.dataType = dataSourceConfig.getDataType();
        this.dataTypeName = dataSourceConfig.getDataTypeName();
        this.systemOwner = dataSourceConfig.getSystemOwner();
        this.included = dataSourceConfig.getIncluded();
    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
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

    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (!(o instanceof UtmDataSourceConfigDTO)) {
            return false;
        }

        UtmDataSourceConfigDTO dataSourceConfigDTO = (UtmDataSourceConfigDTO) o;
        if (this.id == null) {
            return false;
        }
        return Objects.equals(this.id, dataSourceConfigDTO.id);
    }

    @Override
    public int hashCode() {
        return Objects.hash(this.id);
    }

    // prettier-ignore
    @Override
    public String toString() {
        return "UtmDataSourceConfigDTO{" +
            "id=" + getId() +
            ", data_type='" + getDataType() + "'" +
            ", data_type_name='" + getDataTypeName() + "'" +
            ", system_owner='" + getSystemOwner() + "'" +
            ", included='" + getIncluded() + "'" +
            "}";
    }
}

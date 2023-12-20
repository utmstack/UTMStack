package com.park.utmstack.domain.application_modules;


import com.fasterxml.jackson.annotation.JsonIgnore;
import com.park.utmstack.domain.application_modules.types.ModuleConfigurationKey;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import java.io.Serializable;

/**
 * A UtmModuleGroupConfiguration.
 */
@Entity
@Table(name = "utm_module_group_configuration")
public class UtmModuleGroupConfiguration implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @NotNull
    @Column(name = "group_id", nullable = false)
    private Long groupId;

    @NotNull
    @Column(name = "conf_key", nullable = false)
    private String confKey;

    @Column(name = "conf_value")
    private String confValue;

    @Column(name = "conf_name")
    private String confName;

    @Column(name = "conf_description")
    private String confDescription;

    @NotNull
    @Column(name = "conf_data_type", nullable = false)
    private String confDataType;

    @NotNull
    @Column(name = "conf_required", nullable = false)
    private Boolean confRequired;

    @JsonIgnore
    @ManyToOne
    @JoinColumn(name = "group_id", insertable = false, updatable = false)
    private UtmModuleGroup moduleGroup;

    public UtmModuleGroupConfiguration() {
    }

    public UtmModuleGroupConfiguration(ModuleConfigurationKey key) {
        this.groupId = key.getGroupId();
        this.confKey = key.getConfKey();
        this.confName = key.getConfName();
        this.confDescription = key.getConfDescription();
        this.confDataType = key.getConfDataType();
        this.confRequired = key.getConfRequired();
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getGroupId() {
        return groupId;
    }

    public void setGroupId(Long groupId) {
        this.groupId = groupId;
    }

    public String getConfKey() {
        return confKey;
    }

    public void setConfKey(String confKey) {
        this.confKey = confKey;
    }

    public String getConfValue() {
        return confValue;
    }

    public void setConfValue(String confValue) {
        this.confValue = confValue;
    }

    public String getConfName() {
        return confName;
    }

    public void setConfName(String confName) {
        this.confName = confName;
    }

    public String getConfDescription() {
        return confDescription;
    }

    public void setConfDescription(String confDescription) {
        this.confDescription = confDescription;
    }

    public String getConfDataType() {
        return confDataType;
    }

    public void setConfDataType(String confDataType) {
        this.confDataType = confDataType;
    }

    public Boolean getConfRequired() {
        return confRequired;
    }

    public void setConfRequired(Boolean confRequired) {
        this.confRequired = confRequired;
    }

    public UtmModuleGroup getModuleGroup() {
        return moduleGroup;
    }

    public void setModuleGroup(UtmModuleGroup moduleGroup) {
        this.moduleGroup = moduleGroup;
    }
}

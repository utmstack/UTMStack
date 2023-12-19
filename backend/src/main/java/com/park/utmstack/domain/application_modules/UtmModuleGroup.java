package com.park.utmstack.domain.application_modules;


import com.fasterxml.jackson.annotation.JsonIgnore;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import java.io.Serializable;
import java.util.HashSet;
import java.util.Set;

/**
 * A UtmConfigurationGroup.
 */
@Entity
@Table(name = "utm_module_group")
public class UtmModuleGroup implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @NotNull
    @Column(name = "module_id", nullable = false)
    private Long moduleId;

    @NotNull
    @Column(name = "group_name", nullable = false)
    private String groupName;

    @Column(name = "group_description")
    private String groupDescription;

    @JsonIgnore
    @ManyToOne
    @JoinColumn(name = "module_id", insertable = false, updatable = false)
    private UtmModule module;

    @OrderBy
    @OneToMany(mappedBy = "moduleGroup", fetch = FetchType.EAGER)
    private Set<UtmModuleGroupConfiguration> moduleGroupConfigurations = new HashSet<>();

    public UtmModuleGroup() {
    }

    public UtmModuleGroup(String groupName, String groupDescription, Long moduleId) {
        this.groupName = groupName;
        this.groupDescription = groupDescription;
        this.moduleId = moduleId;
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getModuleId() {
        return moduleId;
    }

    public void setModuleId(Long moduleId) {
        this.moduleId = moduleId;
    }

    public String getGroupName() {
        return groupName;
    }

    public void setGroupName(String groupName) {
        this.groupName = groupName;
    }

    public String getGroupDescription() {
        return groupDescription;
    }

    public void setGroupDescription(String groupDescription) {
        this.groupDescription = groupDescription;
    }

    public UtmModule getModule() {
        return module;
    }

    public void setModule(UtmModule module) {
        this.module = module;
    }

    public Set<UtmModuleGroupConfiguration> getModuleGroupConfigurations() {
        return moduleGroupConfigurations;
    }

    public void setModuleGroupConfigurations(Set<UtmModuleGroupConfiguration> moduleGroupConfigurations) {
        this.moduleGroupConfigurations = moduleGroupConfigurations;
    }
}

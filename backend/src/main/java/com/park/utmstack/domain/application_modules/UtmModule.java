package com.park.utmstack.domain.application_modules;


import com.fasterxml.jackson.annotation.JsonIgnore;
import com.park.utmstack.domain.UtmServer;
import com.park.utmstack.domain.application_modules.enums.ModuleName;

import javax.persistence.*;
import java.io.Serializable;
import java.util.HashSet;
import java.util.Set;

/**
 * A UtmModule.
 */
@Entity
@Table(name = "utm_module")
public class UtmModule implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "server_id")
    private Long serverId;

    @Column(name = "pretty_name")
    private String prettyName;

    @Enumerated(EnumType.STRING)
    @Column(name = "module_name")
    private ModuleName moduleName;

    @Column(name = "module_description")
    private String moduleDescription;

    @Column(name = "module_active")
    private Boolean moduleActive;

    @Column(name = "module_icon")
    private String moduleIcon;

    @Column(name = "module_category")
    private String moduleCategory;

    @Column(name = "lite_version")
    private Boolean liteVersion;

    @Column(name = "needs_restart")
    private Boolean needsRestart;

    @Column(name = "is_activatable")
    private Boolean isActivatable;

    @ManyToOne
    @JsonIgnore
    @JoinColumn(name = "server_id", insertable = false, updatable = false)
    private UtmServer server;

    @OneToMany(mappedBy = "module", fetch = FetchType.LAZY)
    private Set<UtmModuleGroup> moduleGroups = new HashSet<>();

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getServerId() {
        return serverId;
    }

    public void setServerId(Long serverId) {
        this.serverId = serverId;
    }

    public String getPrettyName() {
        return prettyName;
    }

    public void setPrettyName(String prettyName) {
        this.prettyName = prettyName;
    }

    public ModuleName getModuleName() {
        return moduleName;
    }

    public void setModuleName(ModuleName moduleName) {
        this.moduleName = moduleName;
    }

    public String getModuleDescription() {
        return moduleDescription;
    }

    public void setModuleDescription(String moduleDescription) {
        this.moduleDescription = moduleDescription;
    }

    public Boolean getModuleActive() {
        return moduleActive;
    }

    public void setModuleActive(Boolean moduleActive) {
        this.moduleActive = moduleActive;
    }

    public String getModuleIcon() {
        return moduleIcon;
    }

    public void setModuleIcon(String moduleIcon) {
        this.moduleIcon = moduleIcon;
    }

    public String getModuleCategory() {
        return moduleCategory;
    }

    public void setModuleCategory(String moduleCategory) {
        this.moduleCategory = moduleCategory;
    }

    public Boolean getLiteVersion() {
        return liteVersion;
    }

    public void setLiteVersion(Boolean liteVersion) {
        this.liteVersion = liteVersion;
    }

    public Boolean getNeedsRestart() {
        return needsRestart;
    }

    public void setNeedsRestart(Boolean needsRestart) {
        this.needsRestart = needsRestart;
    }

    public UtmServer getServer() {
        return server;
    }

    public void setServer(UtmServer server) {
        this.server = server;
    }

    public Set<UtmModuleGroup> getModuleGroups() {
        return moduleGroups;
    }

    public void setModuleGroups(Set<UtmModuleGroup> moduleGroups) {
        this.moduleGroups = moduleGroups;
    }

    public Boolean getActivatable() {
        return isActivatable;
    }

    public void setActivatable(Boolean activatable) {
        isActivatable = activatable;
    }
}

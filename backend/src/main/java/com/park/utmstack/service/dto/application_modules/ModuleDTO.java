package com.park.utmstack.service.dto.application_modules;

import com.park.utmstack.domain.UtmServer;
import com.park.utmstack.domain.application_modules.UtmModuleGroup;
import com.park.utmstack.domain.application_modules.enums.ModuleName;
import com.park.utmstack.domain.correlation.config.UtmDataTypes;
import liquibase.structure.core.DataType;

import java.util.HashSet;
import java.util.Set;

public class ModuleDTO {
    private Long id;

    private Long serverId;

    private String prettyName;

    private ModuleName moduleName;

    private String moduleDescription;

    private Boolean moduleActive;

    private String moduleIcon;

    private String moduleCategory;

    private Boolean liteVersion;

    private Boolean needsRestart;

    private Boolean isActivatable;

    private UtmServer server;
    private Set<UtmModuleGroup> moduleGroups = new HashSet<>();

    private UtmDataTypes dataType;

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

    public Boolean getActivatable() {
        return isActivatable;
    }

    public void setActivatable(Boolean activatable) {
        isActivatable = activatable;
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

    public UtmDataTypes getDataType() {
        return dataType;
    }

    public void setDataType(UtmDataTypes dataType) {
        this.dataType = dataType;
    }

}

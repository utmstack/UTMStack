package com.park.utmstack.service.dto.application_modules;

import com.park.utmstack.domain.UtmServer;
import com.park.utmstack.domain.application_modules.UtmModuleGroup;
import com.park.utmstack.domain.application_modules.enums.ModuleName;
import com.park.utmstack.domain.correlation.config.UtmDataTypes;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.Setter;

import java.util.HashSet;
import java.util.Set;

@Getter
@Setter
@RequiredArgsConstructor
@AllArgsConstructor
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

}

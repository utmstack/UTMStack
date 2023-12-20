package com.park.utmstack.domain.application_modules.factory;

import com.park.utmstack.domain.application_modules.UtmModule;
import com.park.utmstack.domain.application_modules.types.ModuleConfigurationKey;
import com.park.utmstack.domain.application_modules.types.ModuleRequirement;

import java.util.List;

public interface IModule {

    UtmModule getDetails(Long serverId) throws Exception;

    List<ModuleRequirement> checkRequirements(Long serverId) throws Exception;

    List<ModuleConfigurationKey> getConfigurationKeys(Long groupId) throws Exception;
}

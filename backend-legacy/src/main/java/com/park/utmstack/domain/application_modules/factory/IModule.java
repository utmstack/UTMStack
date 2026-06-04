package com.park.utmstack.domain.application_modules.factory;

import com.park.utmstack.domain.application_modules.UtmModule;
import com.park.utmstack.domain.application_modules.UtmModuleGroupConfiguration;
import com.park.utmstack.domain.application_modules.enums.ModuleName;
import com.park.utmstack.domain.application_modules.types.ModuleConfigurationKey;
import com.park.utmstack.domain.application_modules.types.ModuleRequirement;

import java.util.List;

public interface IModule {

    UtmModule getDetails(Long serverId) throws Exception;

    List<ModuleRequirement> checkRequirements(Long serverId) throws Exception;

    List<ModuleConfigurationKey> getConfigurationKeys(Long groupId) throws Exception;

    default boolean validateConfiguration(UtmModule module, List<UtmModuleGroupConfiguration> configuration) throws Exception {
        return true;
    }

    ModuleName getName();
}

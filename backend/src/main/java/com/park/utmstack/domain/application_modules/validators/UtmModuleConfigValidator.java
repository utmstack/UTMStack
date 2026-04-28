package com.park.utmstack.domain.application_modules.validators;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.application_modules.UtmModule;
import com.park.utmstack.domain.application_modules.UtmModuleGroupConfiguration;
import com.park.utmstack.repository.UtmModuleGroupConfigurationRepository;
import com.park.utmstack.service.application_modules.connectors.ModuleConfigurationValidationService;
import com.park.utmstack.service.dto.application_modules.UtmModuleGroupConfDTO;
import com.park.utmstack.service.dto.application_modules.UtmModuleGroupConfWrapperDTO;
import com.park.utmstack.util.CipherUtil;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
@RequiredArgsConstructor
public class UtmModuleConfigValidator {

    private final UtmModuleGroupConfigurationRepository moduleGroupConfigurationRepository;
    private final ModuleConfigurationValidationService utmStackConnectionService;

    public boolean validate(UtmModule module, List<UtmModuleGroupConfiguration> keys) {
        if (keys.isEmpty()) return false;

        List<UtmModuleGroupConfiguration> dbConfigs = moduleGroupConfigurationRepository
                .findAllByGroupId(keys.get(0).getGroupId());

        return validate(module, keys, dbConfigs);
    }

    public boolean validate(UtmModule module, List<UtmModuleGroupConfiguration> keys, List<UtmModuleGroupConfiguration> dbConfigs) {
        if (keys.isEmpty()) return false;

        List<UtmModuleGroupConfDTO> configDTOs = dbConfigs.stream()
                .map(dbConf -> {
                    UtmModuleGroupConfiguration override = findInKeys(keys, dbConf.getConfKey());
                    String value;
                    if (override != null && !Constants.MASKED_VALUE.equals(override.getConfValue())) {
                        // User provided a new value — use it as plaintext
                        value = override.getConfValue();
                    } else {
                        // No override or masked
                        value = dbConf.getConfValue();
                    }
                    return new UtmModuleGroupConfDTO(dbConf.getConfDataType(),dbConf.getConfKey(), value);
                })
                .toList();

        UtmModuleGroupConfWrapperDTO body = new UtmModuleGroupConfWrapperDTO(configDTOs);

        return utmStackConnectionService.validateModuleConfiguration(module.getModuleName().name(), body);
    }

    private UtmModuleGroupConfiguration findInKeys(List<UtmModuleGroupConfiguration> keys, String confKey) {
        return keys.stream()
                .filter(k -> k.getConfKey().equals(confKey))
                .findFirst()
                .orElse(null);
    }

}

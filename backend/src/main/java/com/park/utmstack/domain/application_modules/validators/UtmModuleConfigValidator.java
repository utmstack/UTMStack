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

import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import java.util.stream.Collectors;

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

        List<UtmModuleGroupConfDTO> configDTOs = new ArrayList<>(dbConfigs.stream()
                .map(dbConf -> {
                    UtmModuleGroupConfiguration override = findInKeys(keys, dbConf.getConfKey());
                    String value;
                    if (override != null && !Constants.MASKED_VALUE.equals(override.getConfValue())) {
                        // User provided a new value — encrypt if sensitive so the plugin sees
                        // the same payload shape as the update flow (DB ciphertext).
                        value = encryptIfSensitive(override.getConfDataType(), override.getConfValue());
                    } else {
                        // No override or masked
                        value = dbConf.getConfValue();
                    }
                    return new UtmModuleGroupConfDTO(dbConf.getConfDataType(),dbConf.getConfKey(), value);
                })
                .collect(Collectors.toList()));

        Set<String> dbKeys = dbConfigs.stream()
                .map(UtmModuleGroupConfiguration::getConfKey)
                .collect(Collectors.toCollection(HashSet::new));

        keys.stream()
                .filter(k -> !dbKeys.contains(k.getConfKey()))
                .filter(k -> !Constants.MASKED_VALUE.equals(k.getConfValue()))
                .map(k -> new UtmModuleGroupConfDTO(k.getConfDataType(), k.getConfKey(),
                        encryptIfSensitive(k.getConfDataType(), k.getConfValue())))
                .forEach(configDTOs::add);

        UtmModuleGroupConfWrapperDTO body = new UtmModuleGroupConfWrapperDTO(configDTOs);

        return utmStackConnectionService.validateModuleConfiguration(module.getModuleName().name(), body);
    }

    private String encryptIfSensitive(String dataType, String value) {
        if (value == null || value.isEmpty()) return value;
        if (!Constants.CONF_TYPE_PASSWORD.equals(dataType) && !Constants.CONF_TYPE_FILE.equals(dataType)) {
            return value;
        }
        return CipherUtil.encrypt(value, System.getenv(Constants.ENV_ENCRYPTION_KEY));
    }

    private UtmModuleGroupConfiguration findInKeys(List<UtmModuleGroupConfiguration> keys, String confKey) {
        return keys.stream()
                .filter(k -> k.getConfKey().equals(confKey))
                .findFirst()
                .orElse(null);
    }

}

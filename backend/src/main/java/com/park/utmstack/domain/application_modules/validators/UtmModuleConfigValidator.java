package com.park.utmstack.domain.application_modules.validators;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.application_modules.UtmModule;
import com.park.utmstack.domain.application_modules.UtmModuleGroupConfiguration;
import com.park.utmstack.repository.UtmModuleGroupConfigurationRepository;
import com.park.utmstack.service.application_modules.connectors.UtmStackConnectionService;
import com.park.utmstack.service.dto.application_modules.UtmModuleGroupConfDTO;
import com.park.utmstack.service.dto.application_modules.UtmModuleGroupConfWrapperDTO;
import com.park.utmstack.util.CipherUtil;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
public class UtmModuleConfigValidator {

    private final UtmModuleGroupConfigurationRepository moduleGroupConfigurationRepository;
    private final UtmStackConnectionService utmStackConnectionService;

    public boolean validate(UtmModule module, List<UtmModuleGroupConfiguration> keys) throws Exception {
        if (keys.isEmpty()) return false;

        List<UtmModuleGroupConfiguration> dbConfigs = moduleGroupConfigurationRepository
                .findAllByGroupId(keys.get(0).getGroupId());

        List<UtmModuleGroupConfDTO> configDTOs = dbConfigs.stream()
                .map(dbConf -> {
                    UtmModuleGroupConfiguration override = findInKeys(keys, dbConf.getConfKey());
                    UtmModuleGroupConfiguration source = override != null ? override : dbConf;

                    return new UtmModuleGroupConfDTO(
                            source.getConfKey(),
                            decryptIfNeeded(source.getConfDataType(), source.getConfValue())
                    );
                })
                .toList();

        UtmModuleGroupConfWrapperDTO body = new UtmModuleGroupConfWrapperDTO(configDTOs);

        return utmStackConnectionService.testConnection(module.getModuleName().name(), body);
    }

    private UtmModuleGroupConfiguration findInKeys(List<UtmModuleGroupConfiguration> keys, String confKey) {
        return keys.stream()
                .filter(k -> k.getConfKey().equals(confKey))
                .findFirst()
                .orElse(null);
    }

    private String decryptIfNeeded(String dataType, String value) {
        if (Constants.CONF_TYPE_PASSWORD.equals(dataType) || Constants.CONF_TYPE_FILE.equals(dataType)) {
            return CipherUtil.decrypt(value, System.getenv(Constants.ENV_ENCRYPTION_KEY));
        }
        return value;
    }
}

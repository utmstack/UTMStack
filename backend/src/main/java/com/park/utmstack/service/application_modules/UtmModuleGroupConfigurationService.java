package com.park.utmstack.service.application_modules;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.application_modules.UtmModule;
import com.park.utmstack.domain.application_modules.UtmModuleGroupConfiguration;
import com.park.utmstack.domain.application_modules.enums.ModuleName;
import com.park.utmstack.repository.UtmModuleGroupConfigurationRepository;
import com.park.utmstack.repository.UtmModuleGroupRepository;
import com.park.utmstack.repository.application_modules.UtmModuleRepository;
import com.park.utmstack.event_processor.EventProcessorManagerService;
import com.park.utmstack.util.CipherUtil;
import com.park.utmstack.util.exceptions.ApiException;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

/**
 * Service Implementation for managing UtmModuleGroupConfiguration.
 */
@Service
@Transactional
@RequiredArgsConstructor
@Slf4j
public class UtmModuleGroupConfigurationService {

    private static final String CLASSNAME = "UtmModuleGroupConfigurationService";

    private final UtmModuleGroupConfigurationRepository moduleConfigurationRepository;
    private final UtmModuleRepository moduleRepository;

    public void createConfigurationKeys(List<UtmModuleGroupConfiguration> keys) throws Exception {
        final String ctx = CLASSNAME + ".createConfigurationKeys";
        try {
            if (CollectionUtils.isEmpty(keys))
                return;
            for (UtmModuleGroupConfiguration key : keys) {
                if (isSensitiveType(key.getConfDataType()) && StringUtils.hasText(key.getConfValue())) {
                    key.setConfValue(CipherUtil.encrypt(key.getConfValue(), System.getenv(Constants.ENV_ENCRYPTION_KEY)));
                }
            }
            moduleConfigurationRepository.saveAll(keys);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Update configuration of the application modules.
     * Password/file fields with the masked value are skipped (not changed).
     * Password/file fields with a new value are encrypted before saving.
     */
    public UtmModule updateConfigurationKeys(Long moduleId, List<UtmModuleGroupConfiguration> keys) {
        final String ctx = CLASSNAME + ".updateConfigurationKeys";
        try {
            if (CollectionUtils.isEmpty(keys))
                throw new ApiException("No configuration keys were provided to update", HttpStatus.BAD_REQUEST);

            for (UtmModuleGroupConfiguration key : keys) {
                boolean isSensitive = isSensitiveType(key.getConfDataType());

                // Skip masked values — the user did not change this field
                if (isSensitive && Constants.MASKED_VALUE.equals(key.getConfValue())) {
                    continue;
                }

                if (key.getConfRequired() && !StringUtils.hasText(key.getConfValue()))
                    throw new Exception(String.format("No value was found for required configuration: %1$s (%2$s)", key.getConfName(), key.getConfKey()));

                // Encrypt new sensitive values
                if (isSensitive) {
                    key.setConfValue(CipherUtil.encrypt(key.getConfValue(), System.getenv(Constants.ENV_ENCRYPTION_KEY)));
                }
            }

            // Remove masked entries so they don't overwrite DB values
            List<UtmModuleGroupConfiguration> toSave = keys.stream()
                .filter(k -> !(isSensitiveType(k.getConfDataType()) && Constants.MASKED_VALUE.equals(k.getConfValue())))
                .collect(Collectors.toList());

            if (!toSave.isEmpty()) {
                moduleConfigurationRepository.saveAll(toSave);
            }

            List<ModuleName> needRestartModules = Arrays.asList(ModuleName.AWS_IAM_USER, ModuleName.AZURE,
                    ModuleName.GCP, ModuleName.SOPHOS);

            return moduleRepository.findById(moduleId)
                    .map(module -> {
                        module.setNeedsRestart(needRestartModules.contains(module.getModuleName()));
                        return moduleRepository.save(module);
                    })
                    .orElseThrow(() -> new ApiException(String.format("Module with ID %1$s not found", moduleId), HttpStatus.NOT_FOUND));
        } catch (Exception e) {
            log.error("{}: Error updating configuration keys: {}", ctx, e.getMessage());
            throw new ApiException(String.format("%s: Error updating configuration keys: %s", ctx, e.getMessage()), HttpStatus.INTERNAL_SERVER_ERROR);
        }
    }

    /**
     * Find all configurations of a module group.
     * Sensitive values (password, file) are masked before returning.
     */
    public List<UtmModuleGroupConfiguration> findAllByGroupId(Long groupId) throws Exception {
        final String ctx = CLASSNAME + ".findAllByGroupId";
        try {
            List<UtmModuleGroupConfiguration> configs = moduleConfigurationRepository.findAllByGroupId(groupId);
            maskSensitiveValues(configs);
            return configs;
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Gets all configuration parameter for a group and convert it to a map
     */
    public Map<String, String> getGroupConfigurationAsMap(Long groupId) throws Exception {
        final String ctx = CLASSNAME + ".getGroupConfigurationAsMap";
        try {
            List<UtmModuleGroupConfiguration> configurations = moduleConfigurationRepository.findAllByGroupId(groupId);

            if (CollectionUtils.isEmpty(configurations))
                return Collections.emptyMap();

            return configurations.stream().collect(Collectors.toMap(UtmModuleGroupConfiguration::getConfKey, UtmModuleGroupConfiguration::getConfValue));
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Find a configuration parameter by his group and key
     */
    public UtmModuleGroupConfiguration findByGroupIdAndConfKey(Long groupId, String confKey) throws Exception {
        final String ctx = CLASSNAME + ".findByGroupIdAndConfKey";
        try {
            UtmModuleGroupConfiguration config = moduleConfigurationRepository.findByGroupIdAndConfKey(groupId, confKey);
            if (config != null && isSensitiveType(config.getConfDataType())) {
                config.setConfValue(Constants.MASKED_VALUE);
            }
            return config;
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    private boolean isSensitiveType(String dataType) {
        return Constants.CONF_TYPE_PASSWORD.equals(dataType) || Constants.CONF_TYPE_FILE.equals(dataType);
    }

    private void maskSensitiveValues(List<UtmModuleGroupConfiguration> configs) {
        if (configs == null) return;
        for (UtmModuleGroupConfiguration config : configs) {
            if (isSensitiveType(config.getConfDataType()) && StringUtils.hasText(config.getConfValue())) {
                config.setConfValue(Constants.MASKED_VALUE);
            }
        }
    }
}

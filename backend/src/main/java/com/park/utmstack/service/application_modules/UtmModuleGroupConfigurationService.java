package com.park.utmstack.service.application_modules;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.application_modules.UtmModuleGroupConfiguration;
import com.park.utmstack.domain.application_modules.enums.ModuleName;
import com.park.utmstack.repository.UtmModuleGroupConfigurationRepository;
import com.park.utmstack.repository.application_modules.UtmModuleRepository;
import com.park.utmstack.util.CipherUtil;
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
public class UtmModuleGroupConfigurationService {

    private static final String CLASSNAME = "UtmModuleGroupConfigurationService";

    private final UtmModuleGroupConfigurationRepository moduleConfigurationRepository;
    private final UtmModuleRepository moduleRepository;

    public UtmModuleGroupConfigurationService(UtmModuleGroupConfigurationRepository moduleConfigurationRepository,
                                              UtmModuleRepository moduleRepository) {
        this.moduleConfigurationRepository = moduleConfigurationRepository;
        this.moduleRepository = moduleRepository;
    }

    public void createConfigurationKeys(List<UtmModuleGroupConfiguration> keys) throws Exception {
        final String ctx = CLASSNAME + ".createConfigurationKeys";
        try {
            if (CollectionUtils.isEmpty(keys))
                return;
            moduleConfigurationRepository.saveAll(keys);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Update configuration of the application modules
     *
     * @param keys List of configuration keys to save
     * @throws Exception In case of any error
     */
    public void updateConfigurationKeys(Long moduleId, List<UtmModuleGroupConfiguration> keys) throws Exception {
        final String ctx = CLASSNAME + ".updateConfigurationKeys";
        try {
            if (CollectionUtils.isEmpty(keys))
                return;
            for (UtmModuleGroupConfiguration key : keys) {
                if (key.getConfRequired() && !StringUtils.hasText(key.getConfValue()))
                    throw new Exception(String.format("No value was found for required configuration: %1$s (%2$s)", key.getConfName(), key.getConfKey()));
                if (key.getConfDataType().equals("password"))
                    key.setConfValue(CipherUtil.encrypt(key.getConfValue(), System.getenv(Constants.ENV_ENCRYPTION_KEY)));
            }
            moduleConfigurationRepository.saveAll(keys);

            List<ModuleName> needRestartModules = Arrays.asList(ModuleName.AWS_IAM_USER, ModuleName.AZURE,
                ModuleName.GCP, ModuleName.SOPHOS);

            moduleRepository.findById(moduleId).ifPresent(module -> {
                module.setNeedsRestart(needRestartModules.contains(module.getModuleName()));
                moduleRepository.save(module);
            });
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Find all configurations of a module group
     *
     * @param groupId Identifier of the group to get the configurations
     * @return A list of configuration of a group
     * @throws Exception In case of any error
     */
    public List<UtmModuleGroupConfiguration> findAllByGroupId(Long groupId) throws Exception {
        final String ctx = CLASSNAME + ".findAllByGroupId";
        try {
            return moduleConfigurationRepository.findAllByGroupId(groupId);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Gets all configuration parameter for a group and convert it to a map
     *
     * @param groupId Identifier of a group
     * @return A map with the module group configuration
     * @throws Exception In case of any error
     */
    public Map<String, String> getGroupConfigurationAsMap(Long groupId) throws Exception {
        final String ctx = CLASSNAME + ".getGroupConfigurationAsMap";
        try {
            List<UtmModuleGroupConfiguration> configurations = findAllByGroupId(groupId);

            if (CollectionUtils.isEmpty(configurations))
                return Collections.emptyMap();

            return configurations.stream().collect(Collectors.toMap(UtmModuleGroupConfiguration::getConfKey, UtmModuleGroupConfiguration::getConfValue));
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    /**
     * Find a configuration parameter by his group and key
     *
     * @param groupId Identifier of the group to the param belongs
     * @param confKey Key word of the configuration parameter
     * @return A ${@link UtmModuleGroupConfiguration} object with the configuration parameter information
     * @throws Exception In case of any error
     */
    public UtmModuleGroupConfiguration findByGroupIdAndConfKey(Long groupId, String confKey) throws Exception {
        final String ctx = CLASSNAME + ".findByGroupIdAndConfKey";
        try {
            return moduleConfigurationRepository.findByGroupIdAndConfKey(groupId, confKey);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }
}

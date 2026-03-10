package com.park.utmstack.service.application_modules;

import com.park.utmstack.aop.logging.Loggable;
import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.application_modules.UtmModule;
import com.park.utmstack.domain.application_modules.UtmModuleGroup;
import com.park.utmstack.domain.application_modules.UtmModuleGroupConfiguration;
import com.park.utmstack.event_processor.EventProcessorManagerService;
import com.park.utmstack.repository.UtmModuleGroupConfigurationRepository;
import com.park.utmstack.repository.UtmModuleGroupRepository;
import com.park.utmstack.repository.application_modules.UtmModuleRepository;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.application_modules.ModuleActivationDTO;
import com.park.utmstack.service.dto.application_modules.ModuleDTO;
import com.park.utmstack.service.dto.application_modules.UtmModuleMapper;
import com.park.utmstack.service.dto.collectors.dto.CollectorConfigDTO;
import com.park.utmstack.util.CipherUtil;
import com.park.utmstack.util.exceptions.ApiException;
import lombok.RequiredArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import javax.persistence.EntityNotFoundException;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.stream.Collectors;

/**
 * Service Implementation for managing UtmConfigurationGroup.
 */
@Service
@RequiredArgsConstructor
@Transactional
public class UtmModuleGroupService {

    private static final String CLASSNAME = "UtmModuleGroupService";
    private final Logger log = LoggerFactory.getLogger(UtmModuleGroupService.class);

    private final UtmModuleGroupRepository moduleGroupRepository;
    private final UtmModuleService moduleService;
    private final ApplicationEventService applicationEventService;
    private final UtmModuleRepository moduleRepository;
    private final UtmModuleGroupConfigurationRepository moduleGroupConfigurationRepository;
    private final EventProcessorManagerService eventProcessorManagerService;
    private final UtmModuleMapper moduleMapper;


    /**
     * Save a utmConfigurationGroup.
     *
     * @param utmModuleGroup the entity to save
     * @return the persisted entity
     */
    @Loggable
    public UtmModuleGroup save(UtmModuleGroup utmModuleGroup) {
        log.debug("Request to save UtmConfigurationGroup : {}", utmModuleGroup);
        return moduleGroupRepository.save(utmModuleGroup);
    }

    /**
     * Get all the utmConfigurationGroups.
     *
     * @return the list of entities
     */
    @Transactional(readOnly = true)
    public List<UtmModuleGroup> findAll() {
        log.debug("Request to get all UtmConfigurationGroups");
        return moduleGroupRepository.findAll();
    }


    /**
     * Get one utmConfigurationGroup by id.
     *
     * @param id the id of the entity
     * @return the entity
     */
    @Transactional(readOnly = true)
    public Optional<UtmModuleGroup> findOne(Long id) {
        log.debug("Request to get UtmConfigurationGroup : {}", id);
        return moduleGroupRepository.findById(id);
    }

    /**
     * Delete the utmConfigurationGroup by id.
     *
     * @param id the id of the entity
     */

    public UtmModule delete(Long id) {
        final String ctx = CLASSNAME + ".delete";
        long start = System.currentTimeMillis();

        UtmModuleGroup moduleGroup = this.moduleGroupRepository.findById(id)
                .orElseThrow(() -> new EntityNotFoundException("Configuration group not found with ID: " + id));

        String moduleName = String.valueOf(moduleGroup.getModule().getModuleName());
        Map<String, Object> extra = Map.of(
                "ModuleId", moduleGroup.getModule().getId(),
                "ModuleName", moduleName,
                "GroupId", id
        );

        String attemptMsg = String.format("Initiating deletion of configuration group (ID: %d) for module '%s'", id, moduleName);
        applicationEventService.createEvent(attemptMsg, ApplicationEventType.CONFIG_GROUP_DELETE_ATTEMPT, extra);

        moduleGroupRepository.deleteById(id);

        long duration = System.currentTimeMillis() - start;
        String successMsg = String.format("Configuration group (ID: %d) for module '%s' deleted successfully in %dms", id, moduleName, duration);
        applicationEventService.createEvent(successMsg, ApplicationEventType.CONFIG_GROUP_DELETE_SUCCESS, extra);

        return moduleGroup.getModule();
    }


    public void deleteAllByModuleId(Long id) {
        UtmModule module = this.moduleService.findOne(id)
                .orElseThrow(() -> new EntityNotFoundException("Module not found with id " + id));
        String attemptMsg = String.format("Attempt to delete configuration keys for module '%s' initiated", module.getModuleName());

        Map<String, Object> extra = Map.of(
                "ModuleId", id,
                "ModuleName", module.getModuleName()
        );

        applicationEventService.createEvent(attemptMsg, ApplicationEventType.CONFIG_GROUP_BULK_DELETE_ATTEMPT, extra);
        moduleGroupRepository.deleteAllByModuleId(id);

        String successMsg = String.format("Configuration keys for module '%s' deleted successfully", module.getModuleName());
        applicationEventService.createEvent(successMsg, ApplicationEventType.CONFIG_GROUP_BULK_DELETE_SUCCESS, extra);
    }

    public List<UtmModuleGroup> findAllByModuleId(Long moduleId) throws Exception {
        final String ctx = CLASSNAME + ".findAllByModuleName";
        try {
            return moduleGroupRepository.findAllByModuleId(moduleId);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    public List<UtmModuleGroup> findAllByCollectorId(String collectorId) {
        String ctx = CLASSNAME + ".findAllByCollectorId";
        try {
             return moduleGroupRepository.findAllByCollector(collectorId);
        } catch (Exception e) {
            log.error("{}: Error finding module groups by collector id {}: {}", ctx, collectorId, e.getMessage());
            throw new ApiException(String.format("%s: Error finding module groups by collector id %s", ctx, collectorId), HttpStatus.INTERNAL_SERVER_ERROR);
        }

    }

    public List<UtmModuleGroup> findAllWithCollector() throws Exception {
        final String ctx = CLASSNAME + ".findAllByModuleName";
        try {
            return moduleGroupRepository.findAllByCollectorIsNotNull();
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    @Transactional
    public void deleteCollectorById(Long collectorId) {

        List<UtmModuleGroup> groups = moduleGroupRepository.findAllByCollector(collectorId.toString());

        if (groups.isEmpty()) {
            return;
        }

        UtmModuleGroup group = groups.get(0);

        if (group != null) {
            handleModuleDeactivationIfNeeded(group, collectorId);
        }

        moduleGroupRepository.deleteAllByCollector(collectorId.toString());
    }

    private void handleModuleDeactivationIfNeeded(UtmModuleGroup group, Long collectorId) {

        UtmModule module = moduleRepository.findById(group.getModuleId())
                .orElseThrow(() -> new IllegalStateException("Module not found"));

        if (!module.getModuleActive()) {
            return;
        }

        boolean otherCollectorsExist =
                moduleGroupRepository.findAllByModuleId(module.getId())
                        .stream()
                        .anyMatch(m -> !m.getCollector().equals(collectorId.toString()));

        if (!otherCollectorsExist) {
            moduleService.activateDeactivate(
                    ModuleActivationDTO.builder()
                            .serverId(module.getServerId())
                            .moduleName(module.getModuleName())
                            .activationStatus(false)
                            .build()
            );
        }
    }

    public void updateCollectorConfigurationKeys(CollectorConfigDTO collectorConfig) {
        final String ctx = CLASSNAME + ".updateCollectorConfigurationKeys";
        try {

            List<UtmModuleGroup> dbConfigs = moduleGroupRepository
                    .findAllByModuleIdAndCollector(collectorConfig.getModuleId(),
                            String.valueOf(collectorConfig.getCollector().getId()));

            List<UtmModuleGroupConfiguration> keys = collectorConfig.getKeys();

            if (collectorConfig.getKeys().isEmpty()) {
                moduleGroupRepository.deleteAll(dbConfigs);
            } else {
                for (UtmModuleGroupConfiguration key : keys) {
                    if (key.getConfDataType().equals("password")){
                        key.setConfValue(CipherUtil.encrypt(key.getConfValue(), System.getenv(Constants.ENV_ENCRYPTION_KEY)));
                    }
                }
                List<Long> keyGroupIds = keys.stream()
                        .map(UtmModuleGroupConfiguration::getGroupId)
                        .toList();

                List<UtmModuleGroup> groupsToDelete = dbConfigs.stream()
                        .filter(utmModuleGroup -> !keyGroupIds.contains(utmModuleGroup.getId()))
                        .collect(Collectors.toList());

                moduleGroupRepository.deleteAll(groupsToDelete);
                moduleGroupConfigurationRepository.saveAll(keys);
            }

        } catch (Exception e) {
            log.error("{}: Error updating collector configuration keys for collector id {}: {}", ctx, collectorConfig.getCollector().getId(), e.getMessage());
            throw new ApiException(String.format("%s: Error updating collector configuration keys for collector id %d", ctx, collectorConfig.getCollector().getId()), HttpStatus.INTERNAL_SERVER_ERROR);
        }
    }

}

package com.park.utmstack.service.application_modules;

import com.park.utmstack.aop.logging.Loggable;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.application_modules.UtmModule;
import com.park.utmstack.domain.application_modules.UtmModuleGroup;
import com.park.utmstack.repository.UtmModuleGroupRepository;
import com.park.utmstack.service.application_events.ApplicationEventService;
import lombok.RequiredArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import javax.persistence.EntityNotFoundException;
import java.util.List;
import java.util.Map;
import java.util.Optional;

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

    public void delete(Long id) {
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

    public List<UtmModuleGroup> findAllByCollectorId(String collectorId) throws Exception {
        final String ctx = CLASSNAME + ".findAllByModuleName";
        try {
            return moduleGroupRepository.findAllByCollector(collectorId);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
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
}

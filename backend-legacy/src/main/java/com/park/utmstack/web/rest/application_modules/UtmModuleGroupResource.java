package com.park.utmstack.web.rest.application_modules;

import com.park.utmstack.aop.logging.AuditEvent;
import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.application_modules.UtmModule;
import com.park.utmstack.domain.application_modules.UtmModuleGroup;
import com.park.utmstack.domain.application_modules.UtmModuleGroupConfiguration;
import com.park.utmstack.domain.application_modules.factory.ModuleFactory;
import com.park.utmstack.domain.application_modules.types.ModuleConfigurationKey;
import com.park.utmstack.event_processor.EventProcessorManagerService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.application_modules.UtmModuleGroupConfigurationService;
import com.park.utmstack.service.application_modules.UtmModuleGroupService;
import com.park.utmstack.service.application_modules.UtmModuleService;
import com.park.utmstack.service.dto.application_modules.ModuleDTO;
import com.park.utmstack.service.dto.application_modules.UtmModuleMapper;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.vm.ModuleGroupVM;
import lombok.RequiredArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.dao.DataIntegrityViolationException;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.util.CollectionUtils;
import org.springframework.web.bind.annotation.*;
import tech.jhipster.web.util.ResponseUtil;

import javax.validation.Valid;
import java.net.URISyntaxException;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing UtmConfigurationGroup.
 */
@RestController
@RequiredArgsConstructor
@RequestMapping("/api")
public class UtmModuleGroupResource {

    private static final String CLASSNAME = "UtmModuleGroupResource";
    private final Logger log = LoggerFactory.getLogger(UtmModuleGroupResource.class);

    private static final String ENTITY_NAME = "utmConfigurationGroup";

    private final UtmModuleGroupService moduleGroupService;
    private final ApplicationEventService eventService;
    private final ModuleFactory moduleFactory;
    private final UtmModuleService moduleService;
    private final UtmModuleGroupConfigurationService moduleGroupConfigurationService;
    private final UtmModuleMapper moduleMapper;
    private final EventProcessorManagerService eventProcessorManagerService;

    @PostMapping("/utm-configuration-groups")
    @AuditEvent(
            attemptType = ApplicationEventType.CONFIG_GROUP_CREATE_ATTEMPT,
            attemptMessage = "Attempt to create configuration group initiated",
            successType = ApplicationEventType.CONFIG_GROUP_CREATE_SUCCESS,
            successMessage = "Configuration group created successfully"
    )
    public ResponseEntity<UtmModuleGroup> createConfigurationGroup(@Valid @RequestBody ModuleGroupVM moduleGroupVM) throws URISyntaxException {
        final String ctx = CLASSNAME + ".createConfigurationGroup";
        try {
            UtmModuleGroup group = new UtmModuleGroup(moduleGroupVM.getName(), moduleGroupVM.getDescription(), moduleGroupVM.getModuleId());
            if(moduleGroupVM.getCollector() != null){
                group.setCollector(moduleGroupVM.getCollector());
            }
            UtmModuleGroup result = moduleGroupService.save(group);

            UtmModule module = moduleService.findOne(moduleGroupVM.getModuleId())
                .orElseThrow(() -> new Exception(String.format("Module with ID: %1$s not found", moduleGroupVM.getModuleId())));

            List<ModuleConfigurationKey> defaultConfigurationKeys = moduleFactory.getInstance(module.getModuleName()).getConfigurationKeys(group.getId());

            if (CollectionUtils.isEmpty(defaultConfigurationKeys))
                return ResponseEntity.ok(result);

            List<UtmModuleGroupConfiguration> keys = new ArrayList<>();
            defaultConfigurationKeys.forEach(key -> keys.add(new UtmModuleGroupConfiguration(key)));
            moduleGroupConfigurationService.createConfigurationKeys(keys);

            return ResponseEntity.ok(result);
        } catch (DataIntegrityViolationException e) {
            String msg = ctx + ": " + e.getMostSpecificCause().getMessage().replaceAll("\n", "");
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * PUT  /utm-configuration-groups : Updates an existing utmConfigurationGroup.
     *
     * @param utmModuleGroup the utmConfigurationGroup to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmConfigurationGroup,
     * or with status 400 (Bad Request) if the utmConfigurationGroup is not valid,
     * or with status 500 (Internal Server Error) if the utmConfigurationGroup couldn't be updated
     * @throws URISyntaxException if the Location URI syntax is incorrect
     */
    @PutMapping("/utm-configuration-groups")
    @AuditEvent(
            attemptType = ApplicationEventType.CONFIG_GROUP_UPDATE_ATTEMPT,
            attemptMessage = "Attempt to update configuration group initiated",
            successType = ApplicationEventType.CONFIG_GROUP_UPDATE_SUCCESS,
            successMessage = "Configuration group updated successfully"
    )
    public ResponseEntity<UtmModuleGroup> updateUtmConfigurationGroup(@Valid @RequestBody UtmModuleGroup utmModuleGroup) {
        final String ctx = CLASSNAME + ".updateUtmConfigurationGroup";

        if (utmModuleGroup.getId() == null)
            throw new RuntimeException("Can't update the configuration group because ID is null");
        UtmModuleGroup result = moduleGroupService.save(utmModuleGroup);
        return ResponseEntity.ok(result);
    }

    /**
     * GET  /utm-configuration-groups : get all the utmConfigurationGroups.
     *
     * @return the ResponseEntity with status 200 (OK) and the list of utmConfigurationGroups in body
     */
    @GetMapping("/utm-configuration-groups/module-groups")
    public ResponseEntity<List<UtmModuleGroup>> getModuleGroups(@RequestParam Long moduleId) {
        final String ctx = CLASSNAME + ".getModuleGroups";
        try {
            List<UtmModuleGroup> groups = moduleGroupService.findAllByModuleId(moduleId);
            for (UtmModuleGroup group : groups) {
                if (group.getModuleGroupConfigurations() != null) {
                    for (UtmModuleGroupConfiguration conf : group.getModuleGroupConfigurations()) {
                        if ((Constants.CONF_TYPE_PASSWORD.equals(conf.getConfDataType())
                                || Constants.CONF_TYPE_FILE.equals(conf.getConfDataType()))
                                && conf.getConfValue() != null) {
                            conf.setConfValue(Constants.MASKED_VALUE);
                        }
                    }
                }
            }
            return ResponseEntity.ok(groups);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/utm-configuration-groups/{groupId}")
    public ResponseEntity<UtmModuleGroup> getConfigurationGroup(@PathVariable Long groupId) {
        final String ctx = CLASSNAME + ".getConfigurationGroups";

        Optional<UtmModuleGroup> group = moduleGroupService.findOne(groupId);
        group.ifPresent(g -> {
            if (g.getModuleGroupConfigurations() != null) {
                for (UtmModuleGroupConfiguration conf : g.getModuleGroupConfigurations()) {
                    if ((Constants.CONF_TYPE_PASSWORD.equals(conf.getConfDataType())
                            || Constants.CONF_TYPE_FILE.equals(conf.getConfDataType()))
                            && conf.getConfValue() != null) {
                        conf.setConfValue(Constants.MASKED_VALUE);
                    }
                }
            }
        });
        return ResponseUtil.wrapOrNotFound(group);

    }

    @DeleteMapping("/utm-configuration-groups/delete-single-module-group")
    @AuditEvent(
            attemptType = ApplicationEventType.CONFIG_GROUP_DELETE_ATTEMPT,
            attemptMessage = "Attempt to delete single configuration group initiated",
            successType = ApplicationEventType.CONFIG_GROUP_DELETE_SUCCESS,
            successMessage = "Configuration group deleted successfully"
    )
    public ResponseEntity<Void> deleteSingleModuleGroup(@RequestParam Long groupId) {
        final String ctx = CLASSNAME + ".deleteSingleModuleGroup";

        moduleGroupService.deleteAndFetch(groupId);
        return ResponseEntity.ok().build();

    }

    @DeleteMapping("/utm-configuration-groups/delete-all-module-groups")
    @AuditEvent(
            attemptType = ApplicationEventType.CONFIG_GROUP_BULK_DELETE_ATTEMPT,
            attemptMessage = "Attempt to delete all configuration groups for module initiated",
            successType = ApplicationEventType.CONFIG_GROUP_BULK_DELETE_SUCCESS,
            successMessage = "All configuration groups for module deleted successfully"
    )
    public ResponseEntity<Void> deleteAllModuleGroups(@RequestParam Long moduleId) {
        final String ctx = CLASSNAME + ".deleteAllModuleGroups";

        moduleGroupService.deleteAllByModuleId(moduleId);
        return ResponseEntity.ok().build();

    }
}

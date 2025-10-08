package com.park.utmstack.web.rest.application_modules;

import com.park.utmstack.aop.logging.AuditEvent;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.application_modules.UtmModule;
import com.park.utmstack.domain.application_modules.enums.ModuleName;
import com.park.utmstack.domain.application_modules.enums.ModuleRequirementStatus;
import com.park.utmstack.domain.application_modules.factory.ModuleFactory;
import com.park.utmstack.domain.application_modules.types.ModuleRequirement;
import com.park.utmstack.repository.UtmServerRepository;
import com.park.utmstack.security.internalApiKey.InternalApiKeyFilter;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.application_modules.UtmModuleQueryService;
import com.park.utmstack.service.application_modules.UtmModuleService;
import com.park.utmstack.event_processor.EventProcessorManagerService;
import com.park.utmstack.service.dto.application_modules.CheckRequirementsResponse;
import com.park.utmstack.service.dto.application_modules.ModuleActivationDTO;
import com.park.utmstack.service.dto.application_modules.ModuleDTO;
import com.park.utmstack.service.dto.application_modules.UtmModuleCriteria;
import com.park.utmstack.util.ResponseUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.Setter;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

/**
 * REST controller for managing UtmModule.
 */
@RestController
@RequestMapping("/api")
@RequiredArgsConstructor
public class UtmModuleResource {
    private static final String CLASSNAME = "UtmModuleResource";
    private final Logger log = LoggerFactory.getLogger(UtmModuleResource.class);

    private final UtmModuleService moduleService;
    private final ModuleFactory moduleFactory;
    private final UtmModuleQueryService utmModuleQueryService;
    private final ApplicationEventService eventService;
    private final UtmServerRepository utmServerRepository;
    // List of configurations of type 'file' that needs content decryption
    private final EventProcessorManagerService eventProcessorManagerService;



    @AuditEvent(
            attemptType = ApplicationEventType.MODULE_ACTIVATION_ATTEMPT,
            attemptMessage = "Attempt to activate/deactivate module initiated",
            successType = ApplicationEventType.MODULE_ACTIVATION_SUCCESS,
            successMessage = "Module activated/deactivated successfully"
    )
    @PutMapping("/utm-modules/activateDeactivate")
    public ResponseEntity<UtmModule> activateDeactivate(@RequestParam Long serverId,
                                                        @RequestParam ModuleName nameShort,
                                                        @RequestParam Boolean activationStatus) {
        return ResponseEntity.ok(moduleService.activateDeactivate(ModuleActivationDTO.builder()
                        .serverId(serverId)
                        .moduleName(nameShort)
                        .activationStatus(activationStatus)
                .build()));
    }

    /**
     * GET  /utm-modules : get all the utmModules.
     *
     * @param pageable the pagination information
     * @param criteria the criteria which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the list of utmModules in body
     */
    @GetMapping("/utm-modules")
    public ResponseEntity<List<ModuleDTO>> getAllUtmModules(UtmModuleCriteria criteria, Pageable pageable) {
        final String ctx = CLASSNAME + ".getAllUtmModules";
        try {
            Page<ModuleDTO> page = utmModuleQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-modules");
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseUtil.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @GetMapping("/utm-modules/{id}")
    public ResponseEntity<ModuleDTO> getModuleById(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".getModuleById";
        try {
            return ResponseEntity.ok().body(utmModuleQueryService.findById(id));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseUtil.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @GetMapping("/utm-modules/moduleDetails")
    public ResponseEntity<UtmModule> getModuleDetails(@RequestParam Long serverId,
                                                      @RequestParam ModuleName nameShort) {
        final String ctx = CLASSNAME + ".getModuleDetails";
        try {
            return ResponseEntity.ok(moduleFactory.getInstance(nameShort).getDetails(serverId));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseUtil.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @GetMapping("/utm-modules/module-details-decrypted")
    public ResponseEntity<UtmModule> getModuleDetailsDecrypted(@RequestParam ModuleName nameShort) {
        final String ctx = CLASSNAME + ".getModuleDetailsDecrypted";
        try {
            UtmModule module = moduleFactory.getInstance(nameShort).getDetails(utmServerRepository.getUtmServer());
            if (InternalApiKeyFilter.isApiKeyHeaderInUse()) {
                this.eventProcessorManagerService.decryptModuleConfig(module);
            } else {
                String msg = ctx + ": You must provide the header used to communicate internally with this resource";
                log.error(msg);
                eventService.createEvent(msg, ApplicationEventType.ERROR);
                return ResponseUtil.buildErrorResponse(HttpStatus.BAD_REQUEST, msg);
            }

            return ResponseEntity.ok(module);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseUtil.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @GetMapping("/utm-modules/checkRequirements")
    public ResponseEntity<CheckRequirementsResponse> checkRequirements(@RequestParam Long serverId,
                                                                       @RequestParam ModuleName nameShort) throws Exception {
        final String ctx = CLASSNAME + ".checkRequirements";
        try {
            CheckRequirementsResponse rs = new CheckRequirementsResponse();
            rs.setStatus(ModuleRequirementStatus.OK);

            List<ModuleRequirement> checkResults = moduleFactory.getInstance(nameShort).checkRequirements(serverId);

            rs.setChecks(checkResults);

            for (ModuleRequirement check : checkResults) {
                if (check.getCheckStatus() == ModuleRequirementStatus.FAIL) {
                    rs.setStatus(ModuleRequirementStatus.FAIL);
                    break;
                }
            }
            return ResponseEntity.ok(rs);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseUtil.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @GetMapping("/utm-modules/moduleCategories")
    public ResponseEntity<List<String>> getModuleCategories(@RequestParam(required = false) Long serverId) {
        final String ctx = CLASSNAME + ".getModuleCategories";
        try {
            return ResponseEntity.ok(moduleService.getModuleCategories(serverId));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseUtil.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @GetMapping("/utm-modules/is-active")
    public ResponseEntity<Boolean> isActive(@RequestParam ModuleName moduleName) {
        final String ctx = CLASSNAME + ".isActive";
        try {
            return ResponseEntity.ok(moduleService.isModuleActive(moduleName));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseUtil.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }
}

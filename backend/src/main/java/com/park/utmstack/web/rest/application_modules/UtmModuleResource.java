package com.park.utmstack.web.rest.application_modules;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.application_modules.UtmModule;
import com.park.utmstack.domain.application_modules.UtmModuleGroup;
import com.park.utmstack.domain.application_modules.enums.ModuleName;
import com.park.utmstack.domain.application_modules.enums.ModuleRequirementStatus;
import com.park.utmstack.domain.application_modules.factory.ModuleFactory;
import com.park.utmstack.domain.application_modules.types.ModuleRequirement;
import com.park.utmstack.repository.UtmServerRepository;
import com.park.utmstack.security.internalApiKey.InternalApiKeyFilter;
import com.park.utmstack.service.UtmStackService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.application_modules.UtmModuleQueryService;
import com.park.utmstack.service.application_modules.UtmModuleService;
import com.park.utmstack.service.dto.application_modules.UtmModuleCriteria;
import com.park.utmstack.util.CipherUtil;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.park.utmstack.web.rest.util.PaginationUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.util.StringUtils;
import org.springframework.web.bind.annotation.*;

import java.io.IOException;
import java.util.Formattable;
import java.util.List;
import java.util.Set;
import java.util.logging.FileHandler;
import java.util.logging.Formatter;
import java.util.logging.Level;

/**
 * REST controller for managing UtmModule.
 */
@RestController
@RequestMapping("/api")
public class UtmModuleResource {
    private static final String CLASSNAME = "UtmModuleResource";
    private final Logger log = LoggerFactory.getLogger(UtmModuleResource.class);

    private final UtmModuleService moduleService;
    private final ModuleFactory moduleFactory;
    private final UtmModuleQueryService utmModuleQueryService;
    private final ApplicationEventService eventService;
    private final UtmStackService utmStackService;
    private final UtmServerRepository utmServerRepository;

    public UtmModuleResource(UtmModuleService moduleService,
                             ModuleFactory moduleFactory,
                             UtmModuleQueryService utmModuleQueryService,
                             ApplicationEventService eventService,
                             UtmStackService utmStackService,
                             UtmServerRepository utmServerRepository) {
        this.moduleService = moduleService;
        this.moduleFactory = moduleFactory;
        this.utmModuleQueryService = utmModuleQueryService;
        this.eventService = eventService;
        this.utmStackService = utmStackService;
        this.utmServerRepository = utmServerRepository;
    }

    @PutMapping("/utm-modules/activateDeactivate")
    public ResponseEntity<UtmModule> activateDeactivate(@RequestParam Long serverId,
                                                        @RequestParam ModuleName nameShort,
                                                        @RequestParam Boolean activationStatus) {
        final String ctx = CLASSNAME + ".activateDeactivate";
        try {
            return ResponseEntity.ok(moduleService.activateDeactivate(serverId, nameShort, activationStatus));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * GET  /utm-modules : get all the utmModules.
     *
     * @param pageable the pagination information
     * @param criteria the criteria which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the list of utmModules in body
     */
    @GetMapping("/utm-modules")
    public ResponseEntity<List<UtmModule>> getAllUtmModules(UtmModuleCriteria criteria, Pageable pageable) {
        final String ctx = CLASSNAME + ".getAllUtmModules";
        try {
            Page<UtmModule> page = utmModuleQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-modules");
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
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
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @GetMapping("/utm-modules/module-details-decrypted")
    public ResponseEntity<UtmModule> getModuleDetailsDecrypted(@RequestParam ModuleName nameShort) {
        final String ctx = CLASSNAME + ".getModuleDetailsDecrypted";
        try {
            UtmModule module = moduleFactory.getInstance(nameShort).getDetails(utmServerRepository.getUtmServer());
            if (InternalApiKeyFilter.isApiKeyHeaderInUse()) {
                Set<UtmModuleGroup> groups = module.getModuleGroups();
                groups.forEach((gp) -> {
                    gp.getModuleGroupConfigurations().forEach((gpc) -> {
                        if (gpc.getConfDataType().equals("password") && StringUtils.hasText(gpc.getConfValue())) {
                            gpc.setConfValue(CipherUtil.decrypt(gpc.getConfValue(), System.getenv(Constants.ENV_ENCRYPTION_KEY)));
                        }
                    });
                });
            } else {
                String msg = ctx + ": You must provide the header used to communicate internally with this resource";
                log.error(msg);
                eventService.createEvent(msg, ApplicationEventType.ERROR);
                return UtilResponse.buildErrorResponse(HttpStatus.BAD_REQUEST, msg);
            }

            return ResponseEntity.ok(module);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
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
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
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
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
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
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    public static class CheckRequirementsResponse {
        private ModuleRequirementStatus status;
        private List<ModuleRequirement> checks;

        public ModuleRequirementStatus getStatus() {
            return status;
        }

        public void setStatus(ModuleRequirementStatus status) {
            this.status = status;
        }

        public List<ModuleRequirement> getChecks() {
            return checks;
        }

        public void setChecks(List<ModuleRequirement> checks) {
            this.checks = checks;
        }
    }
}

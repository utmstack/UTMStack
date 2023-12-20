package com.park.utmstack.web.rest.application_modules;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.application_modules.UtmModuleGroupConfiguration;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.application_modules.UtmModuleGroupConfigurationService;
import com.park.utmstack.web.rest.util.HeaderUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import javax.validation.constraints.NotEmpty;
import javax.validation.constraints.NotNull;
import java.util.List;

@RestController
@RequestMapping("/api")
public class UtmModuleGroupConfigurationResource {

    private static final String CLASSNAME = "UtmModuleGroupConfigurationResource";
    private final Logger log = LoggerFactory.getLogger(UtmModuleGroupConfigurationResource.class);

    private final UtmModuleGroupConfigurationService moduleGroupConfigurationService;
    private final ApplicationEventService applicationEventService;

    public UtmModuleGroupConfigurationResource(UtmModuleGroupConfigurationService moduleGroupConfigurationService,
                                               ApplicationEventService applicationEventService) {
        this.moduleGroupConfigurationService = moduleGroupConfigurationService;
        this.applicationEventService = applicationEventService;
    }

    @PutMapping("/module-group-configurations/update")
    public ResponseEntity<Void> updateConfiguration(@Valid @RequestBody UpdateConfigurationKeysBody body) {
        final String ctx = CLASSNAME + ".updateConfiguration";
        try {
            moduleGroupConfigurationService.updateConfigurationKeys(body.getModuleId(), body.getKeys());
            return ResponseEntity.ok().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    @GetMapping("/module-group-configurations/by-group-id")
    public ResponseEntity<List<UtmModuleGroupConfiguration>> getConfigurationByGroupId(@RequestParam Long groupId) {
        final String ctx = CLASSNAME + ".getConfigurationByGroupId";
        try {
            return ResponseEntity.ok(moduleGroupConfigurationService.findAllByGroupId(groupId));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    @GetMapping("/module-group-configurations/by-group-and-key")
    public ResponseEntity<UtmModuleGroupConfiguration> getConfigurationByGroupIdAndConfKey(@RequestParam Long groupId,
                                                                                           @RequestParam String confKey) {
        final String ctx = CLASSNAME + ".getConfigurationByGroupIdAndConfKey";
        try {
            return ResponseEntity.ok(moduleGroupConfigurationService.findByGroupIdAndConfKey(groupId, confKey));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    public static class UpdateConfigurationKeysBody {
        @NotNull
        private Long moduleId;
        @NotEmpty
        private List<UtmModuleGroupConfiguration> keys;

        public Long getModuleId() {
            return moduleId;
        }

        public void setModuleId(Long moduleId) {
            this.moduleId = moduleId;
        }

        public List<UtmModuleGroupConfiguration> getKeys() {
            return keys;
        }

        public void setKeys(List<UtmModuleGroupConfiguration> keys) {
            this.keys = keys;
        }
    }
}

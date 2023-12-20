package com.park.utmstack.web.rest.licence;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.licence.LicenceManagerService;
import com.park.utmstack.util.exceptions.UtmInvalidLicenceException;
import com.park.utmstack.web.rest.util.HeaderUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/licence/v2")
public class LicenceManagerResource {
    private static final String CLASSNAME = "LicenceManagerResource";

    private final Logger log = LoggerFactory.getLogger(LicenceManagerResource.class);

    private final LicenceManagerService licenceService;
    private final ApplicationEventService applicationEventService;

    public LicenceManagerResource(LicenceManagerService licenceService, ApplicationEventService applicationEventService) {
        this.licenceService = licenceService;
        this.applicationEventService = applicationEventService;
    }

    @GetMapping("/activate-licence")
    public ResponseEntity<Void> activateLicence(@RequestParam String licenceKey,
                                                @RequestParam String email) {
        final String ctx = CLASSNAME + ".activateLicence";
        try {
            licenceService.activateLicence(licenceKey, email);
            return ResponseEntity.ok().build();
        } catch (UtmInvalidLicenceException e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.LOCKED).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/check-licence")
    public ResponseEntity<Boolean> checkLicence() {
        final String ctx = CLASSNAME + ".checkLicence";
        try {
            return ResponseEntity.ok(licenceService.checkClientLicence());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }
}

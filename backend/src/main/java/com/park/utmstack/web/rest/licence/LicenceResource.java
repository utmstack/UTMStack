package com.park.utmstack.web.rest.licence;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.licence.LicenceType;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.licence.LicenceService;
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
@RequestMapping("/api/licence")
public class LicenceResource {
    private static final String CLASSNAME = "LicenceResource";

    private final Logger log = LoggerFactory.getLogger(LicenceResource.class);

    private final LicenceService licenceService;
    private final ApplicationEventService applicationEventService;

    public LicenceResource(LicenceService licenceService, ApplicationEventService applicationEventService) {
        this.licenceService = licenceService;
        this.applicationEventService = applicationEventService;
    }

    @GetMapping("/check")
    public ResponseEntity<LicenceType> checkLicence(@RequestParam String licence) {
        final String ctx = CLASSNAME + ".checkLicence";
        try {
            return ResponseEntity.ok(licenceService.checkLicence(licence));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/activate")
    public ResponseEntity<LicenceType.Status> activateLicence(@RequestParam String name,
                                                              @RequestParam String email,
                                                              @RequestParam String licence) {
        final String ctx = CLASSNAME + ".activateLicence";
        try {
            return ResponseEntity.ok(licenceService.activateLicence(name, email, licence));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }
}

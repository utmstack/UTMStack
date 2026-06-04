package com.park.utmstack.web.rest.federation_service;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.federation_service.UtmFederationServiceClientService;
import com.park.utmstack.web.rest.util.HeaderUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import javax.validation.constraints.NotNull;

/**
 * REST controller for managing UtmFederationServiceClient.
 */
@RestController
@RequestMapping("/api")
public class UtmFederationServiceClientResource {

    private static final String CLASSNAME = "UtmFederationServiceClientResource";
    private final Logger log = LoggerFactory.getLogger(UtmFederationServiceClientResource.class);

    private final UtmFederationServiceClientService federationServiceClientService;
    private final ApplicationEventService eventService;

    public UtmFederationServiceClientResource(UtmFederationServiceClientService federationServiceClientService,
                                              ApplicationEventService eventService) {
        this.federationServiceClientService = federationServiceClientService;
        this.eventService = eventService;
    }

    @GetMapping("/federation-service/generateApiToken")
    public ResponseEntity<String> generateApiToken() {
        final String ctx = CLASSNAME + ".activate";
        try {
            return ResponseEntity.ok(federationServiceClientService.generateApiToken());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/federation-service/token")
    public ResponseEntity<String> getFederationServiceManagerToken() {
        final String ctx = CLASSNAME + ".getFederationServiceManagerToken";
        try {
            return ResponseEntity.ok(federationServiceClientService.getFederationServiceManagerToken());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    public static class ActivateBody {
        @NotNull
        private Boolean activate;
        private String name;
        private String domain;

        public Boolean getActivate() {
            return activate;
        }

        public void setActivate(Boolean activate) {
            this.activate = activate;
        }

        public String getName() {
            return name;
        }

        public void setName(String name) {
            this.name = name;
        }

        public String getDomain() {
            return domain;
        }

        public void setDomain(String domain) {
            this.domain = domain;
        }
    }
}

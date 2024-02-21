package com.park.utmstack.web.rest.soc_ai;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.soc_ai.SocAIService;
import com.park.utmstack.web.rest.AccountResource;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/soc-ai")
public class UtmSocAiResource {

    private final Logger log = LoggerFactory.getLogger(AccountResource.class);

    private static final String CLASSNAME = "UtmSocAiResource";

    private final ApplicationEventService applicationEventService;
    private final SocAIService socAIService;
    public UtmSocAiResource(SocAIService socAIService, ApplicationEventService applicationEventService) {
        this.socAIService = socAIService;
        this.applicationEventService = applicationEventService;
    }

    @PostMapping("/send-alerts")
    public ResponseEntity<String> sendData(@RequestBody String[] alertsId) {
        final String ctx = CLASSNAME + ".sendAlertsIds";
        try {
            socAIService.sendData(alertsId);
            return ResponseEntity.ok("Processing successful");
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);

            return ResponseEntity.internalServerError().body(msg);
        }
    }
}

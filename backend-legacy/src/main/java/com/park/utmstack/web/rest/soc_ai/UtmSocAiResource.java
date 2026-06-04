package com.park.utmstack.web.rest.soc_ai;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.shared_types.alert.UtmAlert;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.soc_ai.SocAIService;
import com.park.utmstack.web.rest.AccountResource;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

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

    /**
     * POST /api/soc-ai/analyze : Submit an alert for SOC-AI analysis
     *
     * @param alert the complete alert object to analyze
     * @return status of the submission
     */
    @PostMapping("/analyze")
    public ResponseEntity<Object> analyzeAlert(@RequestBody UtmAlert alert) {
        final String ctx = CLASSNAME + ".analyzeAlert";
        try {
            if (alert == null || alert.getId() == null) {
                return ResponseEntity.badRequest().body(Map.of("status", "error", "message", "Alert ID is required"));
            }

            socAIService.analyzeAlert(alert);
            return ResponseEntity.accepted().body(Map.of(
                "status", "queued",
                "alertId", alert.getId(),
                "message", "Alert queued for SOC-AI analysis"
            ));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);

            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(Map.of(
                "status", "error",
                "message", msg
            ));
        }
    }

}

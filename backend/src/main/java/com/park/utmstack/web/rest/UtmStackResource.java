package com.park.utmstack.web.rest;


import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.UtmConfigurationParameter;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.mail_sender.MailConfig;
import com.park.utmstack.service.UtmConfigurationParameterService;
import com.park.utmstack.service.UtmStackService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.mail_config.MailConfigService;
import com.park.utmstack.util.CipherUtil;
import com.park.utmstack.util.UtilResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.actuate.info.InfoEndpoint;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.util.StringUtils;
import org.springframework.web.bind.annotation.*;

import javax.mail.MessagingException;
import javax.validation.Valid;
import java.util.List;
import java.util.Map;

/**
 * REST controller for managing the current user's account.
 */
@RestController
@RequestMapping("/api")
public class UtmStackResource {
    private final Logger log = LoggerFactory.getLogger(UtmStackResource.class);
    private static final String CLASSNAME = "UtmStackResource";

    private final UtmStackService utmStackService;
    private final ApplicationEventService applicationEventService;
    private final UtmConfigurationParameterService utmConfigurationParameterService;
    private final InfoEndpoint infoEndpoint;



    public UtmStackResource(UtmStackService utmStackService,
                            ApplicationEventService applicationEventService,
                            UtmConfigurationParameterService utmConfigurationParameterService,
                            InfoEndpoint infoEndpoint) {
        this.utmStackService = utmStackService;
        this.applicationEventService = applicationEventService;
        this.utmConfigurationParameterService = utmConfigurationParameterService;
        this.infoEndpoint = infoEndpoint;
    }

    @GetMapping("/ping")
    public ResponseEntity<HttpStatus> ping() {
        return ResponseEntity.ok(HttpStatus.OK);
    }

    @GetMapping("/date-format")
    public ResponseEntity<Map<String, String>> dateFormat() {
        final String ctx = CLASSNAME + ".dateFormat";
        try {
            return ResponseEntity.ok(utmConfigurationParameterService.getValueMapForDateSetting());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @GetMapping("/healthcheck")
    public ResponseEntity<HttpStatus> healthCheck() {
        final String ctx = CLASSNAME + ".healthCheck";
        try {
            utmStackService.executeChecks();
            return ResponseEntity.ok(HttpStatus.OK);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @GetMapping("/isInDevelop")
    public ResponseEntity<Boolean> isInDevelop() {
        final String ctx = CLASSNAME + ".isInDevelop";
        try {
            return ResponseEntity.ok(utmStackService.isInDevelop());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @PostMapping("/encrypt")
    public ResponseEntity<String> encrypt(@RequestBody String str) {
        final String ctx = CLASSNAME + ".encrypt";
        try {
            if (!StringUtils.hasText(str))
                throw new Exception("Content to encrypt is missing");
            return ResponseEntity.ok(CipherUtil.encrypt(str, System.getenv(Constants.ENV_ENCRYPTION_KEY)));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }
}

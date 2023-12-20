package com.park.utmstack.web.rest.alert_response_rule;

import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseRule;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.alert_response_rule.UtmAlertResponseRuleQueryService;
import com.park.utmstack.service.alert_response_rule.UtmAlertResponseRuleService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.UtmAlertResponseRuleCriteria;
import com.park.utmstack.service.dto.UtmAlertResponseRuleDTO;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.web.rest.util.PaginationUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springdoc.api.annotations.ParameterObject;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@RestController
@RequestMapping("/api")
public class UtmAlertResponseRuleResource {

    private static final String CLASSNAME = "UtmAlertResponseRuleResource";
    private final Logger log = LoggerFactory.getLogger(UtmAlertResponseRuleResource.class);

    private final UtmAlertResponseRuleService alertResponseRuleService;
    private final UtmAlertResponseRuleQueryService alertResponseRuleQueryService;
    private final ApplicationEventService eventService;

    public UtmAlertResponseRuleResource(UtmAlertResponseRuleService alertResponseRuleService,
                                        UtmAlertResponseRuleQueryService alertResponseRuleQueryService,
                                        ApplicationEventService eventService) {
        this.alertResponseRuleService = alertResponseRuleService;
        this.alertResponseRuleQueryService = alertResponseRuleQueryService;
        this.eventService = eventService;
    }

    @PostMapping("/utm-alert-response-rules")
    public ResponseEntity<UtmAlertResponseRuleDTO> createAlertResponseRule(@Valid @RequestBody UtmAlertResponseRuleDTO dto) {
        final String ctx = CLASSNAME + ".createAlertResponseRule";
        try {
            if (dto.getId() != null) {
                String msg = ctx + ": A new rule cannot already have an ID";
                log.error(msg);
                eventService.createEvent(msg, ApplicationEventType.ERROR);
                return UtilResponse.buildErrorResponse(HttpStatus.BAD_REQUEST, msg);
            }
            return ResponseEntity.ok(new UtmAlertResponseRuleDTO(alertResponseRuleService.save(new UtmAlertResponseRule(dto))));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @PutMapping("/utm-alert-response-rules")
    public ResponseEntity<UtmAlertResponseRuleDTO> updateAlertResponseRule(@Valid @RequestBody UtmAlertResponseRuleDTO dto) {
        final String ctx = CLASSNAME + ".updateAlertResponseRule";
        try {
            if (dto.getId() == null) {
                String msg = ctx + ": The rule you are trying to update does not have a valid ID";
                log.error(msg);
                eventService.createEvent(msg, ApplicationEventType.ERROR);
                return UtilResponse.buildErrorResponse(HttpStatus.BAD_REQUEST, msg);
            }
            return ResponseEntity.ok(new UtmAlertResponseRuleDTO(alertResponseRuleService.save(new UtmAlertResponseRule(dto))));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @GetMapping("/utm-alert-response-rules")
    public ResponseEntity<List<UtmAlertResponseRuleDTO>> getAllAlertResponseRules(@ParameterObject UtmAlertResponseRuleCriteria criteria,
                                                                                  @ParameterObject Pageable pageable) {
        final String ctx = CLASSNAME + ".getAllAlertResponseRules";
        try {
            Page<UtmAlertResponseRule> page = alertResponseRuleQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-alert-Response-rules");
            return ResponseEntity.ok().headers(headers).body(page.getContent().stream().map(UtmAlertResponseRuleDTO::new)
                .collect(Collectors.toList()));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @GetMapping("/utm-alert-response-rules/{id}")
    public ResponseEntity<UtmAlertResponseRuleDTO> getAlertResponseRule(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".getAlertResponseRule";
        try {
            return alertResponseRuleService.findOne(id).map(r -> ResponseEntity.ok(new UtmAlertResponseRuleDTO(r)))
                .orElse(ResponseEntity.notFound().build());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @GetMapping("/utm-alert-response-rules/resolve-filter-values")
    public ResponseEntity<Map<String, List<String>>> resolveFilterValues() {
        final String ctx = CLASSNAME + ".getAlertResponseRule";
        try {
            return ResponseEntity.ok(alertResponseRuleService.resolveFilterValues());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }
}

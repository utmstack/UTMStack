package com.park.utmstack.web.rest.alert_response_rule;

import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseRuleExecution;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.alert_response_rule.UtmAlertResponseRuleExecutionQueryService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.UtmAlertResponseRuleExecutionCriteria;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.web.rest.util.PaginationUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/api")
public class UtmAlertResponseRuleExecutionResource {

    private static final String CLASSNAME = "UtmAlertResponseRuleExecutionResource";
    private final Logger log = LoggerFactory.getLogger(UtmAlertResponseRuleExecutionResource.class);

    private final UtmAlertResponseRuleExecutionQueryService ruleExecutionQueryService;
    private final ApplicationEventService eventService;

    public UtmAlertResponseRuleExecutionResource(UtmAlertResponseRuleExecutionQueryService ruleExecutionQueryService,
                                                 ApplicationEventService eventService) {
        this.ruleExecutionQueryService = ruleExecutionQueryService;
        this.eventService = eventService;
    }

    @GetMapping("/utm-alert-response-rule-executions")
    public ResponseEntity<List<UtmAlertResponseRuleExecution>> getAllAlertResponseRuleExecutions(UtmAlertResponseRuleExecutionCriteria criteria, Pageable pageable) {
        final String ctx = CLASSNAME + ".getAllAlertResponseRuleExecutions";
        try {
            Page<UtmAlertResponseRuleExecution> page = ruleExecutionQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-alert-response-rule-executions");
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildInternalServerErrorResponse(msg);
        }
    }
}

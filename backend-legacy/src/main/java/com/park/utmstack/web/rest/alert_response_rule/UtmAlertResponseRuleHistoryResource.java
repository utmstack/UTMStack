package com.park.utmstack.web.rest.alert_response_rule;

import com.park.utmstack.domain.alert_response_rule.UtmAlertResponseRuleHistory;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.alert_response_rule.UtmAlertResponseRuleHistoryQueryService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.UtmAlertResponseRuleHistoryCriteria;
import com.park.utmstack.service.dto.UtmAlertResponseRuleHistoryDTO;
import com.park.utmstack.web.rest.util.PaginationUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springdoc.api.annotations.ParameterObject;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;
import java.util.stream.Collectors;

/**
 * REST controller for managing UtmAlertResponseRuleHistory.
 */
@RestController
@RequestMapping("/api")
public class UtmAlertResponseRuleHistoryResource {
    private static final String CLASSNAME = "UtmAlertResponseRuleHistoryResource";
    private final Logger log = LoggerFactory.getLogger(UtmAlertResponseRuleHistoryResource.class);

    private final UtmAlertResponseRuleHistoryQueryService alertResponseRuleHistoryQueryService;
    private final ApplicationEventService eventService;

    public UtmAlertResponseRuleHistoryResource(UtmAlertResponseRuleHistoryQueryService alertResponseRuleHistoryQueryService,
                                               ApplicationEventService eventService) {
        this.alertResponseRuleHistoryQueryService = alertResponseRuleHistoryQueryService;
        this.eventService = eventService;
    }

    @GetMapping("/utm-alert-response-rule-histories")
    public ResponseEntity<List<UtmAlertResponseRuleHistoryDTO>> getAllAlertResponseRuleHistories(@ParameterObject UtmAlertResponseRuleHistoryCriteria criteria,
                                                                                                 @ParameterObject Pageable pageable) {
        final String ctx = CLASSNAME + ".getAllAlertResponseRuleHistories";
        try {
            Page<UtmAlertResponseRuleHistory> page = alertResponseRuleHistoryQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-alert-response-rule-histories");
            return ResponseEntity.ok().headers(headers).body(page.getContent().stream()
                .map(UtmAlertResponseRuleHistoryDTO::new).collect(Collectors.toList()));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            eventService.createEvent(msg, ApplicationEventType.ERROR);
            throw new RuntimeException(msg);
        }
    }
}

package com.park.utmstack.web.rest;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.UtmAlertService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.util.AlertUtil;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.web.rest.util.HeaderUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Pattern;
import java.util.List;

/**
 * REST controller for managing UtmAlert.
 */
@RestController
@RequestMapping("/api")
public class UtmAlertResource {

    private static final String CLASSNAME = "UtmAlertResource";
    private final Logger log = LoggerFactory.getLogger(UtmAlertResource.class);

    private final UtmAlertService utmAlertService;
    private final ApplicationEventService applicationEventService;
    private final AlertUtil alertUtil;

    public UtmAlertResource(UtmAlertService utmAlertService,
                            ApplicationEventService applicationEventService,
                            AlertUtil alertUtil) {
        this.utmAlertService = utmAlertService;
        this.applicationEventService = applicationEventService;
        this.alertUtil = alertUtil;
    }

    @PostMapping("/utm-alerts/status")
    public ResponseEntity<Void> updateAlertStatus(@RequestBody UpdateAlertStatusRequestBody rq) {
        final String ctx = CLASSNAME + ".updateAlertStatus";
        try {
            utmAlertService.updateStatus(rq.getAlertIds(), rq.getStatus(), rq.getStatusObservation());
            return ResponseEntity.ok().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @PostMapping("/utm-alerts/notes")
    public ResponseEntity<Void> updateAlertNotes(@RequestBody(required = false) String notes, @RequestParam String alertId) {
        final String ctx = CLASSNAME + ".updateAlertNotes";
        try {
            utmAlertService.updateNotes(alertId, notes);
            return ResponseEntity.ok().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @PostMapping("/utm-alerts/tags")
    public ResponseEntity<Void> updateAlertTags(@RequestBody @Valid UpdateAlertTagsRequestBody body) {
        final String ctx = CLASSNAME + ".updateAlertTags";
        try {
            utmAlertService.updateTags(body.getAlertIds(), body.getTags(), body.isCreateRule());
            return ResponseEntity.ok().build();
        } catch (Exception ex) {
            String msg = ctx + ": " + ex.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @PostMapping("/utm-alerts/convert-to-incident")
    public ResponseEntity<Void> convertToIncident(@RequestBody @Valid ConvertToIncidentRequestBody body) {
        final String ctx = CLASSNAME + ".convertToIncident";
        try {
            utmAlertService.convertToIncident(body.getEventIds(), body.getIncidentName(),body.getIncidentId(), body.getIncidentSource());
            return ResponseEntity.ok().build();
        } catch (Exception ex) {
            String msg = ctx + ": " + ex.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/utm-alerts/count-open-alerts")
    public ResponseEntity<Long> countOpenAlerts() {
        final String ctx = CLASSNAME + ".countOpenAlerts";
        try {
            return ResponseEntity.ok(alertUtil.countAlertsByStatus(2));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    public static class UpdateAlertTagsRequestBody {
        @NotNull
        private List<String> alertIds;
        private List<String> tags;
        @NotNull
        private Boolean createRule;

        public List<String> getAlertIds() {
            return alertIds;
        }

        public void setAlertIds(List<String> alertIds) {
            this.alertIds = alertIds;
        }

        public List<String> getTags() {
            return tags;
        }

        public void setTags(List<String> tags) {
            this.tags = tags;
        }

        public Boolean isCreateRule() {
            return createRule;
        }

        public void setCreateRule(Boolean createRule) {
            this.createRule = createRule;
        }
    }

    public static class UpdateAlertStatusRequestBody {
        @NotNull
        private List<String> alertIds;
        private String statusObservation;
        @NotNull
        private int status;

        public List<String> getAlertIds() {
            return alertIds;
        }

        public void setAlertIds(List<String> alertIds) {
            this.alertIds = alertIds;
        }

        public String getStatusObservation() {
            return statusObservation;
        }

        public void setStatusObservation(String statusObservation) {
            this.statusObservation = statusObservation;
        }

        public int getStatus() {
            return status;
        }

        public void setStatus(int status) {
            this.status = status;
        }
    }

    public static class UpdateAlertSolutionRequestBody {
        @NotNull
        private String alertName;
        @NotNull
        private String solution;

        public String getAlertName() {
            return alertName;
        }

        public void setAlertName(String alertName) {
            this.alertName = alertName;
        }

        public String getSolution() {
            return solution;
        }

        public void setSolution(String solution) {
            this.solution = solution;
        }
    }

    public static class ConvertToIncidentRequestBody {
        @NotNull
        private List<String> eventIds;
        @NotNull
        @Pattern(regexp = "^[^\"]*$", message = "Double quotes are not allowed")
        private String incidentName;
        @NotNull
        private Integer incidentId;
        @NotNull
        private String incidentSource;

        public List<String> getEventIds() {
            return eventIds;
        }

        public void setEventIds(List<String> eventIds) {
            this.eventIds = eventIds;
        }

        public String getIncidentName() {
            return incidentName;
        }

        public void setIncidentName(String incidentName) {
            this.incidentName = incidentName;
        }

        public Integer getIncidentId() {
            return incidentId;
        }

        public void setIncidentId(Integer incidentId) {
            this.incidentId = incidentId;
        }

        public String getIncidentSource() {
            return incidentSource;
        }

        public void setIncidentSource(String incidentSource) {
            this.incidentSource = incidentSource;
        }
    }
}

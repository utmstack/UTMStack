package com.park.utmstack.web.rest.incident_response;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.incident_response.UtmIncidentActionCommand;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.incident_response.UtmIncidentActionCommandCriteria;
import com.park.utmstack.service.incident_response.UtmIncidentActionCommandQueryService;
import com.park.utmstack.service.incident_response.UtmIncidentActionCommandService;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.net.URISyntaxException;
import java.util.List;

/**
 * REST controller for managing UtmIncidentActionCommand.
 */
@RestController
@RequestMapping("/api")
public class UtmIncidentActionCommandResource {

    private static final String CLASSNAME = "UtmIncidentActionCommandResource";
    private final Logger log = LoggerFactory.getLogger(UtmIncidentActionCommandResource.class);

    private final UtmIncidentActionCommandService incidentActionCommandService;
    private final UtmIncidentActionCommandQueryService incidentActionCommandQueryService;
    private final ApplicationEventService applicationEventService;

    public UtmIncidentActionCommandResource(UtmIncidentActionCommandService incidentActionCommandService,
                                            UtmIncidentActionCommandQueryService incidentActionCommandQueryService,
                                            ApplicationEventService applicationEventService) {
        this.incidentActionCommandService = incidentActionCommandService;
        this.incidentActionCommandQueryService = incidentActionCommandQueryService;
        this.applicationEventService = applicationEventService;
    }

    @PostMapping("/incident-action-commands")
    public ResponseEntity<UtmIncidentActionCommand> createIncidentActionCommand(@Valid @RequestBody UtmIncidentActionCommand incidentActionCommand) {
        final String ctx = CLASSNAME + ".createIncidentActionCommand";
        try {
            if (incidentActionCommand.getId() != null)
                throw new Exception("A new incident action command cannot already have an ID");
            return ResponseEntity.ok(incidentActionCommandService.save(incidentActionCommand));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @PutMapping("/incident-action-commands")
    public ResponseEntity<UtmIncidentActionCommand> updateIncidentActionCommand(@Valid @RequestBody UtmIncidentActionCommand incidentActionCommand) throws URISyntaxException {
        final String ctx = CLASSNAME + ".updateIncidentActionCommand";
        try {
            if (incidentActionCommand.getId() == null)
                throw new Exception("The incident action command identifier is null");
            return ResponseEntity.ok(incidentActionCommandService.save(incidentActionCommand));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/incident-action-commands")
    public ResponseEntity<List<UtmIncidentActionCommand>> getAllIncidentActionCommands(UtmIncidentActionCommandCriteria criteria, Pageable pageable) {
        final String ctx = CLASSNAME + ".getAllIncidentActionCommands";
        try {
            Page<UtmIncidentActionCommand> page = incidentActionCommandQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-incident-action-commands");
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @DeleteMapping("/incident-action-commands/{id}")
    public ResponseEntity<Void> deleteIncidentActionCommand(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".deleteIncidentActionCommand";
        try {
            incidentActionCommandService.delete(id);
            return ResponseEntity.ok().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }
}

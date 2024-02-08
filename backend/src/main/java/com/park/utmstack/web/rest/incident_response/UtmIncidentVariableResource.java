package com.park.utmstack.web.rest.incident_response;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.incident_response.UtmIncidentVariable;
import com.park.utmstack.security.SecurityUtils;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.incident_response.UtmIncidentVariableCriteria;
import com.park.utmstack.service.incident_response.UtmIncidentVariableQueryService;
import com.park.utmstack.service.incident_response.UtmIncidentVariableService;
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

import java.time.Instant;
import java.util.List;

/**
 * REST controller for managing UtmIncidentVariable.
 */
@RestController
@RequestMapping("/api")
public class UtmIncidentVariableResource {

    private final Logger log = LoggerFactory.getLogger(UtmIncidentVariableResource.class);

    private static final String ENTITY_NAME = "utmIncidentVariable";
    private static final String CLASS_NAME = "UtmIncidentVariableResource";

    private final UtmIncidentVariableService utmIncidentVariableService;
    private final UtmIncidentVariableQueryService utmIncidentVariableQueryService;

    private final ApplicationEventService applicationEventService;

    public UtmIncidentVariableResource(UtmIncidentVariableService utmIncidentVariableService,
                                       UtmIncidentVariableQueryService utmIncidentVariableQueryService,
                                       ApplicationEventService applicationEventService) {
        this.utmIncidentVariableService = utmIncidentVariableService;
        this.utmIncidentVariableQueryService = utmIncidentVariableQueryService;
        this.applicationEventService = applicationEventService;
    }

    /**
     * POST  /utm-incident-variables : Create a new Incident variable.
     *
     * @param utmIncidentVariable the utmIncidentVariable to create
     * @return the ResponseEntity with status 201 (Created) and with body the new utmIncidentVariable, or with status 400 (Bad
     * Request) if the utmIncidentVariable has already an ID
     */
    @PostMapping("/utm-incident-variables")
    public ResponseEntity<UtmIncidentVariable> createUtmIncidentVariable(@RequestBody UtmIncidentVariable utmIncidentVariable) {
        final String ctx = CLASS_NAME + ".createUtmIncidentVariable";
        try {
            if (utmIncidentVariable.getId() != null)
                throw new Exception("A new utmIncidentVariable cannot already have an ID");

            utmIncidentVariable.setLastModifiedDate(Instant.now());
            utmIncidentVariable.setCreatedBy(SecurityUtils.getCurrentUserLogin().orElse("system"));

            UtmIncidentVariable result = utmIncidentVariableService.save(utmIncidentVariable);
            return ResponseEntity.ok(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * PUT  /utm-incident-variables : Updates an existing variable.
     *
     * @param utmIncidentVariable the utmIncidentVariable to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmIncidentVariable, or with status 400 (Bad
     * Request) if the utmIncidentVariable is not valid, or with status 500 (Internal Server Error) if the utmIncidentVariable
     * couldn't be updated
     */
    @PutMapping("/utm-incident-variables")
    public ResponseEntity<UtmIncidentVariable> updateUtmIncidentVariable(@RequestBody UtmIncidentVariable utmIncidentVariable) {
        final String ctx = CLASS_NAME + ".updateUtmIncidentVariable";

        try {
            if (utmIncidentVariable.getId() == null)
                throw new Exception("Id can't be null");

            utmIncidentVariable.setLastModifiedDate(Instant.now());
            utmIncidentVariable.setLastModifiedBy(SecurityUtils.getCurrentUserLogin().orElse("system"));

            UtmIncidentVariable result = utmIncidentVariableService.save(utmIncidentVariable);
            return ResponseEntity.ok(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * GET  /utm-incident-variables : get all the variables.
     *
     * @param pageable the pagination information
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the list of utmIncidentVariables in body
     */
    @GetMapping("/utm-incident-variables")
    public ResponseEntity<List<UtmIncidentVariable>> getAllUtmIncidentVariables(UtmIncidentVariableCriteria criteria,
                                                                                Pageable pageable) {
        final String ctx = CLASS_NAME + ".getAllUtmIncidentVariables";
        try {
            Page<UtmIncidentVariable> page = utmIncidentVariableQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-incident-actions");
            List<UtmIncidentVariable> content = page.getContent();

            return ResponseEntity.ok().headers(headers).body(content);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * DELETE  /utm-incident-variables/:id : delete the "id" utmIncidentVariable.
     *
     * @param id the id of the utmIncidentVariable to delete
     * @return the ResponseEntity with status 200 (OK)
     */
    @DeleteMapping("/utm-incident-variables/{id}")
    public ResponseEntity<Void> deleteUtmIncidentVariable(@PathVariable Long id) {
        final String ctx = CLASS_NAME + ".deleteUtmIncidentVariable";
        try {
            utmIncidentVariableService.delete(id);
            return ResponseEntity.ok().headers(HeaderUtil.createEntityDeletionAlert(ENTITY_NAME, id.toString())).build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }
}

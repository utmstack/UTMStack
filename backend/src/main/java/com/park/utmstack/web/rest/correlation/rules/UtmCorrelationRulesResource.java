package com.park.utmstack.web.rest.correlation.rules;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.correlation.rules.UtmCorrelationRulesFilter;
import com.park.utmstack.domain.network_scan.NetworkScanFilter;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.correlation.rules.UtmCorrelationRulesService;
import com.park.utmstack.service.dto.network_scan.NetworkScanDTO;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import com.park.utmstack.web.rest.vm.UtmCorrelationRulesVM;
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


/**
 * REST controller for managing {@link UtmCorrelationRulesResource}.
 */
@RestController
@RequestMapping("/api")
public class UtmCorrelationRulesResource {
    private static final String CLASSNAME = "UtmCorrelationRulesResource";
    private final Logger log = LoggerFactory.getLogger(UtmCorrelationRulesResource.class);

    private final ApplicationEventService applicationEventService;

    private final UtmCorrelationRulesService rulesService;

    public UtmCorrelationRulesResource(ApplicationEventService applicationEventService, UtmCorrelationRulesService rulesService) {
        this.applicationEventService = applicationEventService;
        this.rulesService = rulesService;
    }

    /**
     * {@code POST  /correlation-rule} : Add a new correlation rule definition with its datatypes.
     *
     * @param rulesVM the correlation rule definition to insert.
     * @return the {@link ResponseEntity} with status {@code 204 (No Content)}, with status {@code 400 (Bad Request)}, or with status {@code 500 (Internal)} if errors occurred.
     */
    @PostMapping("/correlation-rule")
    public ResponseEntity<Void> addCorrelationRule(@Valid @RequestBody UtmCorrelationRulesVM rulesVM) {
        final String ctx = CLASSNAME + ".addCorrelationRule";
        try {
            rulesService.addRule(rulesVM);
            return ResponseEntity.noContent().build();
        } catch (BadRequestAlertException e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.BAD_REQUEST, msg);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * {@code POST  /correlation-rule/activate-deactivate} : Activate or deactivate a correlation rule.
     *
     * @param rulesVM the correlation rule definition to activate or deactivate.
     * @return the {@link ResponseEntity} with status {@code 204 (No Content)}, with status {@code 400 (Bad Request)}, or with status {@code 500 (Internal)} if errors occurred.
     */
    @PostMapping("/correlation-rule/activate-deactivate")
    public ResponseEntity<Void> activateOrDeactivateCorrelationRule(@Valid @RequestBody UtmCorrelationRulesVM rulesVM,
                                                                    @RequestParam Boolean active) {
        final String ctx = CLASSNAME + ".activateOrDeactivateCorrelationRule";
        try {
            rulesService.setRuleActivation(rulesVM, active);
            return ResponseEntity.noContent().build();
        } catch (BadRequestAlertException e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.BAD_REQUEST, msg);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * {@code PUT  /correlation-rule} : Update a correlation rule definition with its datatypes.
     *
     * @param rulesVM the correlation rule definition to update.
     * @return the {@link ResponseEntity} with status {@code 204 (No Content)}, with status {@code 400 (Bad Request)}, or with status {@code 500 (Internal)} if errors occurred.
     */
    @PutMapping("/correlation-rule")
    public ResponseEntity<Void> updateCorrelationRule(@Valid @RequestBody UtmCorrelationRulesVM rulesVM) {
        final String ctx = CLASSNAME + ".updateCorrelationRule";
        try {
            rulesService.updateRule(rulesVM);
            return ResponseEntity.noContent().build();
        } catch (BadRequestAlertException e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.BAD_REQUEST, msg);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @GetMapping("/correlation-rule/search-by-filters")
    public ResponseEntity<List<UtmCorrelationRulesVM>> searchByFilters(@ParameterObject UtmCorrelationRulesFilter filters,
                                                                @ParameterObject Pageable pageable) {
        final String ctx = CLASSNAME + ".searchByFilters";
        try {
            Page<UtmCorrelationRulesVM> page = rulesService.searchByFilters(filters, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/correlation-rule/search-by-filters");
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * {@code DELETE  /correlation-rule/:id} : Remove a correlation rule definition with its datatypes.
     *
     * @param id the id of the correlation rule to remove.
     * @return the {@link ResponseEntity} with status {@code 204 (No Content)}, with status {@code 400 (Bad Request)}, or with status {@code 500 (Internal)} if errors occurred.
     */
    @DeleteMapping("/correlation-rule/{id}")
    public ResponseEntity<Void> removeCorrelationRule(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".removeCorrelationRule";
        try {
            rulesService.deleteRule(id);
            return ResponseEntity.noContent().build();
        } catch (BadRequestAlertException e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.BAD_REQUEST, msg);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }
}

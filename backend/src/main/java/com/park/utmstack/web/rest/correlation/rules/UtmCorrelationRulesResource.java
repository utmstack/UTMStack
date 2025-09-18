package com.park.utmstack.web.rest.correlation.rules;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.correlation.rules.UtmCorrelationRules;
import com.park.utmstack.domain.correlation.rules.UtmCorrelationRulesFilter;
import com.park.utmstack.domain.network_scan.Property;
import com.park.utmstack.service.UtmStackService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.correlation.rules.UtmCorrelationRulesService;
import com.park.utmstack.service.dto.correlation.UtmCorrelationRulesDTO;
import com.park.utmstack.service.dto.correlation.UtmCorrelationRulesMapper;
import com.park.utmstack.service.dto.correlation.validators.CorrelationRuleValidator;
import com.park.utmstack.util.ResponseUtil;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import io.undertow.util.BadRequestException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springdoc.api.annotations.ParameterObject;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.WebDataBinder;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.persistence.EntityNotFoundException;
import javax.validation.Valid;
import java.util.List;
import java.util.Optional;


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

    private final UtmCorrelationRulesMapper utmCorrelationRulesMapper;

    private UtmStackService utmStackService;

    private final CorrelationRuleValidator correlationRuleValidator;

    public UtmCorrelationRulesResource(ApplicationEventService applicationEventService,
                                       UtmCorrelationRulesService rulesService,
                                       UtmCorrelationRulesMapper utmCorrelationRulesMapper,
                                       UtmStackService utmStackService,
                                       CorrelationRuleValidator correlationRuleValidator) {
        this.applicationEventService = applicationEventService;
        this.rulesService = rulesService;
        this.utmCorrelationRulesMapper = utmCorrelationRulesMapper;
        this.utmStackService = utmStackService;
        this.correlationRuleValidator = correlationRuleValidator;
    }
    @InitBinder("utmCorrelationRulesDTO")
    protected void initBinder(WebDataBinder binder) {
        binder.addValidators(correlationRuleValidator);
    }

    /**
     * {@code POST  /correlation-rule} : Add a new correlation rule definition with its datatypes.
     *
     * @param utmCorrelationRulesDTO the correlation rule definition to insert.
     * @return the {@link ResponseEntity} with status {@code 204 (No Content)}, with status {@code 400 (Bad Request)}, or with status {@code 500 (Internal)} if errors occurred.
     */
    @PostMapping("/correlation-rule")
    public ResponseEntity<Void> addCorrelationRule(@Valid @RequestBody UtmCorrelationRulesDTO utmCorrelationRulesDTO) {
        final String ctx = CLASSNAME + ".addCorrelationRule";

        try {
            if (utmStackService.isInDevelop()) {
                utmCorrelationRulesDTO.setId(rulesService.getSystemSequenceNextValue());
                utmCorrelationRulesDTO.setSystemOwner(true);
            } else {
                utmCorrelationRulesDTO.setSystemOwner(false);
            }

            rulesService.save(this.utmCorrelationRulesMapper.toEntity(utmCorrelationRulesDTO));
            return ResponseEntity.noContent().build();
        } catch (BadRequestException e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseUtil.buildErrorResponse(HttpStatus.BAD_REQUEST, msg);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseUtil.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * {@code POST  /correlation-rule/activate-deactivate} : Activate or deactivate a correlation rule.
     *
     * @param id the correlation rule definition to activate or deactivate.
     * @return the {@link ResponseEntity} with status {@code 204 (No Content)}, with status {@code 400 (Bad Request)}, or with status {@code 500 (Internal)} if errors occurred.
     */
    @PutMapping("/correlation-rule/activate-deactivate")
    public ResponseEntity<Void> activateOrDeactivateCorrelationRule(@RequestParam Long id,
                                                                    @RequestParam Boolean active) {
        final String ctx = CLASSNAME + ".activateOrDeactivateCorrelationRule";
        try {
            rulesService.setRuleActivation(id, active);
            return ResponseEntity.noContent().build();
        } catch (BadRequestException e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseUtil.buildErrorResponse(HttpStatus.BAD_REQUEST, msg);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseUtil.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * {@code PUT  /correlation-rule} : Update a correlation rule definition with its datatypes.
     *
     * @param correlationRulesDTO the correlation rule definition to update.
     * @return the {@link ResponseEntity} with status {@code 204 (No Content)}, with status {@code 400 (Bad Request)}, or with status {@code 500 (Internal)} if errors occurred.
     */
    @PutMapping("/correlation-rule")
    public ResponseEntity<Void> updateCorrelationRule(@Valid @RequestBody UtmCorrelationRulesDTO correlationRulesDTO) {
        final String ctx = CLASSNAME + ".updateCorrelationRule";
        try {
            if (correlationRulesDTO.getDefinition() == null) {
                throw new BadRequestException(ctx + ": The rule's definition field can't be null.");
            }
            rulesService.updateRule(this.utmCorrelationRulesMapper.toEntity(correlationRulesDTO));
            return ResponseEntity.noContent().build();
        }  catch (BadRequestException e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseUtil.buildErrorResponse(HttpStatus.BAD_REQUEST, msg);
        }  catch (EntityNotFoundException e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseUtil.buildErrorResponse(HttpStatus.NOT_FOUND, msg);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseUtil.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @GetMapping("/correlation-rule/search-by-filters")
    public ResponseEntity<List<UtmCorrelationRulesDTO>> searchByFilters(@ParameterObject UtmCorrelationRulesFilter filters,
                                                                @ParameterObject Pageable pageable) {
        final String ctx = CLASSNAME + ".searchByFilters";
        try {
            Page<UtmCorrelationRulesDTO> page = rulesService.searchByFilters(filters, pageable);
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

    @GetMapping("/correlation-rule/search-property-values")
    public ResponseEntity<List<?>> searchPropertyValues(@RequestParam Property prop,
                                                        @RequestParam(required = false) String value,
                                                        Pageable pageable) {
        final String ctx = CLASSNAME + ".searchPropertyValues";
        try {
            return ResponseEntity.ok(rulesService.searchPropertyValues(prop, value, pageable));
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * GET  /correlation-rule/:id : The id of the datatype.
     *
     * @param id the id of the correlation rule to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the rule with its relations, or with status 404 (Not Found)
     */
    @GetMapping("/correlation-rule/{id}")
    public ResponseEntity<UtmCorrelationRulesDTO> getRule(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".getRule";
        try {
            Optional<UtmCorrelationRules> utmCorrelationRule = rulesService.findOne(id);
            if (utmCorrelationRule.isPresent()) {
                UtmCorrelationRulesDTO dto = utmCorrelationRulesMapper.toDto(utmCorrelationRule.get());
                return tech.jhipster.web.util.ResponseUtil.wrapOrNotFound(Optional.of(dto));
            } else {
                return tech.jhipster.web.util.ResponseUtil.wrapOrNotFound(Optional.empty());
            }
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
            return ResponseUtil.buildErrorResponse(HttpStatus.BAD_REQUEST, msg);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseUtil.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }
}

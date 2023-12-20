package com.park.utmstack.web.rest;

import com.park.utmstack.domain.UtmConfigurationParameter;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.UtmConfigurationParameterQueryService;
import com.park.utmstack.service.UtmConfigurationParameterService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.dto.UtmConfigurationParameterCriteria;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springdoc.api.annotations.ParameterObject;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.util.Assert;
import org.springframework.web.bind.annotation.*;
import tech.jhipster.web.util.ResponseUtil;

import javax.validation.Valid;
import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing UtmConfigurationParameter.
 */
@RestController
@RequestMapping("/api")
public class UtmConfigurationParameterResource {

    private final Logger log = LoggerFactory.getLogger(UtmConfigurationParameterResource.class);

    private static final String ENTITY_NAME = "utmConfigurationParameter";
    private static final String CLASSNAME = "UtmConfigurationParameterResource";

    private final UtmConfigurationParameterService utmConfigurationParameterService;

    private final UtmConfigurationParameterQueryService utmConfigurationParameterQueryService;

    private final ApplicationEventService applicationEventService;

    public UtmConfigurationParameterResource(UtmConfigurationParameterService utmConfigurationParameterService,
                                             UtmConfigurationParameterQueryService utmConfigurationParameterQueryService, ApplicationEventService applicationEventService) {
        this.utmConfigurationParameterService = utmConfigurationParameterService;
        this.utmConfigurationParameterQueryService = utmConfigurationParameterQueryService;
        this.applicationEventService = applicationEventService;
    }

    /**
     * PUT  /utm-configuration-parameters : Updates an existing utmConfigurationParameter.
     *
     * @param parameters the utmConfigurationParameter to update
     * @return the ResponseEntity with status 200 (OK) and with body the updated utmConfigurationParameter,
     * or with status 400 (Bad Request) if the utmConfigurationParameter is not valid,
     * or with status 500 (Internal Server Error) if the utmConfigurationParameter couldn't be updated
     */
    @PutMapping("/utm-configuration-parameters")
    public ResponseEntity<Void> updateConfigurationParameters(@Valid @RequestBody List<UtmConfigurationParameter> parameters) {
        final String ctx = CLASSNAME + ".updateUtmConfigurationParameter";
        try {
            Assert.notEmpty(parameters, "There isn't any parameter to update");
            utmConfigurationParameterService.saveAll(parameters);
            return ResponseEntity.ok().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    /**
     * GET  /utm-configuration-parameters : get all the utmConfigurationParameters.
     *
     * @param pageable the pagination information
     * @param criteria the criterias which the requested entities should match
     * @return the ResponseEntity with status 200 (OK) and the list of utmConfigurationParameters in body
     */
    @GetMapping("/utm-configuration-parameters")
    public ResponseEntity<List<UtmConfigurationParameter>> getAllUtmConfigurationParameters(@ParameterObject UtmConfigurationParameterCriteria criteria,
                                                                                            @ParameterObject Pageable pageable) {
        log.debug("REST request to get UtmConfigurationParameters by criteria: {}", criteria);
        Page<UtmConfigurationParameter> page = utmConfigurationParameterQueryService.findByCriteria(criteria, pageable);
        HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/utm-configuration-parameters");
        return ResponseEntity.ok().headers(headers).body(page.getContent());
    }

    /**
     * GET  /utm-configuration-parameters/:id : get the "id" utmConfigurationParameter.
     *
     * @param id the id of the utmConfigurationParameter to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the utmConfigurationParameter, or with status 404 (Not Found)
     */
    @GetMapping("/utm-configuration-parameters/{id}")
    public ResponseEntity<UtmConfigurationParameter> getUtmConfigurationParameter(@PathVariable Long id) {
        log.debug("REST request to get UtmConfigurationParameter : {}", id);
        Optional<UtmConfigurationParameter> utmConfigurationParameter = utmConfigurationParameterService.findOne(id);
        return ResponseUtil.wrapOrNotFound(utmConfigurationParameter);
    }
}

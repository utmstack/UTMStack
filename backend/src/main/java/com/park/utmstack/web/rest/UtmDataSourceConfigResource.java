package com.park.utmstack.web.rest;

import com.park.utmstack.domain.UtmDataSourceConfig;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.correlation.config.UtmDataTypes;
import com.park.utmstack.service.UtmDataSourceConfigService;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.util.UtilResponse;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springdoc.api.annotations.ParameterObject;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.support.ServletUriComponentsBuilder;
import tech.jhipster.web.util.PaginationUtil;

import javax.validation.Valid;
import java.util.List;

/**
 * REST controller for managing {@link UtmDataSourceConfig}.
 */
@RestController
@RequestMapping("/api")
public class UtmDataSourceConfigResource {
    private static final String CLASSNAME = "UtmDataSourceConfigResource";
    private final Logger log = LoggerFactory.getLogger(UtmDataSourceConfigResource.class);

    private final UtmDataSourceConfigService dataSourceConfigService;
    private final ApplicationEventService applicationEventService;

    public UtmDataSourceConfigResource(UtmDataSourceConfigService dataSourceConfigService,
                                       ApplicationEventService applicationEventService) {
        this.dataSourceConfigService = dataSourceConfigService;
        this.applicationEventService = applicationEventService;
    }

    @PutMapping("/utm-datasource-configs")
    public ResponseEntity<Void> update(@Valid @RequestBody List<UtmDataTypes> configs) {
        final String ctx = CLASSNAME + ".update";
        try {
            dataSourceConfigService.update(configs);
            return ResponseEntity.noContent().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    @GetMapping("/utm-datasource-configs")
    public ResponseEntity<List<UtmDataTypes>> getAllDataSourceConfigs(@ParameterObject Pageable pageable) {
        final String ctx = CLASSNAME + ".getAllDataSourceConfigs";
        Page<UtmDataTypes> page = dataSourceConfigService.findAll(pageable);
        try {
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(ServletUriComponentsBuilder.fromCurrentRequest(), page);
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }
}

package com.park.utmstack.web.rest.correlation.config;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.correlation.config.UtmTenantConfig;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.correlation.config.UtmTenantConfigService;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import io.undertow.util.BadRequestException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import tech.jhipster.web.util.ResponseUtil;

import javax.validation.Valid;
import java.util.List;
import java.util.Optional;

/**
 * REST controller for managing {@link UtmTenantConfigResource}.
 */
@RestController
@RequestMapping("/api")
public class UtmTenantConfigResource {
    private static final String CLASSNAME = "UtmTenantConfigResource";
    private final Logger log = LoggerFactory.getLogger(UtmTenantConfigResource.class);

    private final ApplicationEventService applicationEventService;
    private final UtmTenantConfigService tenantConfigService;

    public UtmTenantConfigResource(ApplicationEventService applicationEventService, UtmTenantConfigService tenantConfigService) {
        this.applicationEventService = applicationEventService;
        this.tenantConfigService = tenantConfigService;
    }
    /**
     * {@code POST  /tenant-config} : Add a new tenant configuration.
     *
     * @param tenantConfig the tenant configuration to insert.
     * @return the {@link ResponseEntity} with status {@code 204 (No Content)}, with status {@code 400 (Bad Request)}, or with status {@code 500 (Internal)} if errors occurred.
     */
    @PostMapping("/tenant-config")
    public ResponseEntity<Void> addTenantConfig(@Valid @RequestBody UtmTenantConfig tenantConfig) {
        final String ctx = CLASSNAME + ".addTenantConfig";
        try {
            tenantConfigService.addTenantConfig(tenantConfig);
            return ResponseEntity.noContent().build();
        } catch (BadRequestException e) {
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
     * {@code PUT  /tenant-config} : Update a tenant configuration.
     *
     * @param tenantConfig the tenant configuration to update.
     * @return the {@link ResponseEntity} with status {@code 204 (No Content)}, with status {@code 400 (Bad Request)}, or with status {@code 500 (Internal)} if errors occurred.
     */
    @PutMapping("/tenant-config")
    public ResponseEntity<Void> updateTenantConfig(@Valid @RequestBody UtmTenantConfig tenantConfig) {
        final String ctx = CLASSNAME + ".updateTenantConfig";
        try {
            tenantConfigService.updateTenantConfig(tenantConfig);
            return ResponseEntity.noContent().build();
        } catch (BadRequestException e) {
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
     * {@code DELETE  /tenant-config/:id} : Remove a tenant configuration.
     *
     * @param id the id of the tenant configuration to remove.
     * @return the {@link ResponseEntity} with status {@code 204 (No Content)}, with status {@code 400 (Bad Request)}, or with status {@code 500 (Internal)} if errors occurred.
     */
    @DeleteMapping("/tenant-config/{id}")
    public ResponseEntity<Void> removeTenantConfig(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".removeTenantConfig";
        try {
            tenantConfigService.delete(id);
            return ResponseEntity.noContent().build();
        } catch (BadRequestException e) {
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
     * GET  /tenant-config : get all the tenant configuration.
     *
     * @param pageable the pagination information
     * @return the ResponseEntity with status 200 (OK) and the list of tenant configuration in body
     */
    @GetMapping("/tenant-config")
    public ResponseEntity<List<UtmTenantConfig>> getAllTenantConfig(@RequestParam(required = false) String search, Pageable pageable) {
        final String ctx = CLASSNAME + ".getAllTenantConfig";
        try {
            Page<UtmTenantConfig> page = tenantConfigService.findAll(search, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/tenant-config");
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
     * GET  /tenant-config/:id : The id of the tenant configuration.
     *
     * @param id the id of the tenant configuration to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the tenant configuration, or with status 404 (Not Found)
     */
    @GetMapping("/tenant-config/{id}")
    public ResponseEntity<UtmTenantConfig> getTenantConfig(@PathVariable Long id) {
        log.debug("REST request to get UtmTenantConfig : {}", id);
        Optional<UtmTenantConfig> tenantConfig = tenantConfigService.findOne(id);
        return ResponseUtil.wrapOrNotFound(tenantConfig);
    }
}

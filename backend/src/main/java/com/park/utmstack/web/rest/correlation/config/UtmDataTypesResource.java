package com.park.utmstack.web.rest.correlation.config;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.correlation.config.UtmDataTypes;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.correlation.config.UtmDataTypesService;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
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
 * REST controller for managing {@link UtmDataTypesResource}.
 */
@RestController
@RequestMapping("/api")
public class UtmDataTypesResource {
    private static final String CLASSNAME = "UtmDataTypesResource";
    private final Logger log = LoggerFactory.getLogger(UtmDataTypesResource.class);

    private final ApplicationEventService applicationEventService;
    private final UtmDataTypesService dataTypesService;

    public UtmDataTypesResource(ApplicationEventService applicationEventService, UtmDataTypesService dataTypesService) {
        this.applicationEventService = applicationEventService;
        this.dataTypesService = dataTypesService;
    }
    /**
     * {@code POST  /data-types} : Add a new datatype.
     *
     * @param dataTypes the datatype to insert.
     * @return the {@link ResponseEntity} with status {@code 204 (No Content)}, with status {@code 400 (Bad Request)}, or with status {@code 500 (Internal)} if errors occurred.
     */
    @PostMapping("/data-types")
    public ResponseEntity<Void> addDataType(@Valid @RequestBody UtmDataTypes dataTypes) {
        final String ctx = CLASSNAME + ".addDataType";
        try {
            dataTypesService.addDataType(dataTypes);
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
     * {@code PUT  /data-types} : Update a datatype.
     *
     * @param dataTypes the datatype to update.
     * @return the {@link ResponseEntity} with status {@code 204 (No Content)}, with status {@code 400 (Bad Request)}, or with status {@code 500 (Internal)} if errors occurred.
     */
    @PutMapping("/data-types")
    public ResponseEntity<Void> updateDataTypes(@Valid @RequestBody UtmDataTypes dataTypes) {
        final String ctx = CLASSNAME + ".updateDataTypes";
        try {
            dataTypesService.updateDataType(dataTypes);
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
     * {@code PUT  /data-types/include-exclude-list} : Used to update only the datatype included field and synchronize the assets
     * */
    @PutMapping("/data-types/include-exclude-list")
    public ResponseEntity<Void> updateDataTypesList(@Valid @RequestBody List<UtmDataTypes> dataTypes) {
        final String ctx = CLASSNAME + ".updateDataTypesList";
        try {
            dataTypesService.updateList(dataTypes);
            return ResponseEntity.noContent().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    /**
     * {@code DELETE  /data-types/:id} : Remove a datatype.
     *
     * @param id the id of the datatype to remove.
     * @return the {@link ResponseEntity} with status {@code 204 (No Content)}, with status {@code 400 (Bad Request)}, or with status {@code 500 (Internal)} if errors occurred.
     */
    @DeleteMapping("/data-types/{id}")
    public ResponseEntity<Void> removeDataTypes(@PathVariable Long id) {
        final String ctx = CLASSNAME + ".removeDataTypes";
        try {
            dataTypesService.delete(id);
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
     * GET  /data-types : get all the datatypes.
     *
     * @param pageable the pagination information
     * @return the ResponseEntity with status 200 (OK) and the list of datatypes in body
     */
    @GetMapping("/data-types")
    public ResponseEntity<List<UtmDataTypes>> getAllDataTypes(@RequestParam(required = false) String search, Pageable pageable) {
        final String ctx = CLASSNAME + ".getAllDataTypes";
        try {
            Page<UtmDataTypes> page = dataTypesService.findAll(search, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/data-types");
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
     * GET  /data-types/:id : The id of the datatype.
     *
     * @param id the id of the datatype to retrieve
     * @return the ResponseEntity with status 200 (OK) and with body the datatype, or with status 404 (Not Found)
     */
    @GetMapping("/data-types/{id}")
    public ResponseEntity<UtmDataTypes> getDataType(@PathVariable Long id) {
        log.debug("REST request to get UtmDataTypes : {}", id);
        Optional<UtmDataTypes> datatype = dataTypesService.findOne(id);
        return ResponseUtil.wrapOrNotFound(datatype);
    }
}

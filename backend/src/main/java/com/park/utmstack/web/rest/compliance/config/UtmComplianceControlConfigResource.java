package com.park.utmstack.web.rest.compliance.config;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.compliance.config.UtmComplianceControlConfigService;
import com.park.utmstack.service.compliance.config.UtmComplianceControlConfigCriteriaService;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigCriteria;
import com.park.utmstack.service.dto.compliance.UtmComplianceControlConfigDto;
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

import java.util.List;

@RestController
@RequestMapping("/api/compliance/control-config")
public class UtmComplianceControlConfigResource {
    private final Logger log = LoggerFactory.getLogger(UtmComplianceControlConfigResource.class);
    private static final String CLASS_NAME = "UtmComplianceControlConfigResource";

    private final UtmComplianceControlConfigService controlService;
    private final UtmComplianceControlConfigCriteriaService criteriaService;
    private final ApplicationEventService applicationEventService;


    public UtmComplianceControlConfigResource(
            UtmComplianceControlConfigService controlService,
            UtmComplianceControlConfigCriteriaService criteriaService,
            ApplicationEventService applicationEventService
    ) {
        this.controlService = controlService;
        this.criteriaService = criteriaService;
        this.applicationEventService = applicationEventService;
    }

    @PostMapping
    public ResponseEntity<UtmComplianceControlConfigDto> createControl(
            @RequestBody UtmComplianceControlConfigDto dto
    ) {
        var created = controlService.create(dto);
        return ResponseEntity.ok(created);
    }

    @GetMapping("/{id}")
    public ResponseEntity<UtmComplianceControlConfigDto> getControl(@PathVariable Long id) {
        var entity = controlService.findById(id);
        if (entity == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(entity);
    }

    @PutMapping("/{id}")
    public ResponseEntity<UtmComplianceControlConfigDto> updateControl(
            @PathVariable Long id,
            @RequestBody UtmComplianceControlConfigDto dto
    ) {
        var updated = controlService.update(id, dto);
        return ResponseEntity.ok(updated);
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteControl(@PathVariable Long id) {
        controlService.delete(id);
        return ResponseEntity.noContent().build();
    }

    @GetMapping
    public ResponseEntity<List<UtmComplianceControlConfigDto>> getAllComplianceControlConfig(
            UtmComplianceControlConfigCriteria criteria,
            Pageable pageable) {

        final String ctx = CLASS_NAME + ".getAllComplianceControlConfig";

        try {
            Page<UtmComplianceControlConfigDto> page = criteriaService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/compliance/control-config");

            return ResponseEntity.ok().headers(headers).body(page.getContent());

        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);

            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .headers(HeaderUtil.createFailureAlert("", "", msg))
                    .body(null);
        }
    }

}

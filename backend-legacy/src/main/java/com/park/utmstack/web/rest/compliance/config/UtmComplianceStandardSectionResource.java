package com.park.utmstack.web.rest.compliance.config;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.compliance.UtmComplianceStandardSection;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.compliance.config.UtmComplianceStandardSectionQueryService;
import com.park.utmstack.service.compliance.config.UtmComplianceStandardSectionService;
import com.park.utmstack.service.dto.compliance.UtmComplianceStandardSectionCriteria;
import com.park.utmstack.web.rest.util.HeaderUtil;
import com.park.utmstack.web.rest.util.PaginationUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.core.env.Environment;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import tech.jhipster.config.JHipsterConstants;
import tech.jhipster.web.util.ResponseUtil;

import javax.validation.Valid;
import java.util.Arrays;
import java.util.List;
import java.util.Optional;

@RestController
@RequestMapping("/api/compliance")
public class UtmComplianceStandardSectionResource {
    private final Logger log = LoggerFactory.getLogger(UtmComplianceStandardSectionResource.class);
    private static final String CLASS_NAME = "UtmComplianceStandardSectionResource";

    private final UtmComplianceStandardSectionService standardSectionService;
    private final UtmComplianceStandardSectionQueryService complianceStandardSectionQueryService;
    private final ApplicationEventService applicationEventService;
    private final Environment env;

    public UtmComplianceStandardSectionResource(UtmComplianceStandardSectionService standardSectionService,
                                                UtmComplianceStandardSectionQueryService complianceStandardSectionQueryService,
                                                ApplicationEventService applicationEventService, Environment env) {
        this.standardSectionService = standardSectionService;
        this.complianceStandardSectionQueryService = complianceStandardSectionQueryService;
        this.applicationEventService = applicationEventService;
        this.env = env;
    }

    @PostMapping("/standard-section")
    public ResponseEntity<UtmComplianceStandardSection> createComplianceStandardSection(
        @Valid @RequestBody UtmComplianceStandardSection complianceStandardSection) {
        final String ctx = CLASS_NAME + ".createComplianceStandardSection";

        try {
            if (complianceStandardSection.getId() != null) {
                String msg = ctx + ": A new complianceStandardSection can't have an id";
                return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
            }

            if (Arrays.asList(env.getActiveProfiles()).contains(JHipsterConstants.SPRING_PROFILE_DEVELOPMENT))
                complianceStandardSection.setId(standardSectionService.getSystemSequenceNextValue(complianceStandardSection.getStandardId()));

            UtmComplianceStandardSection result = standardSectionService.save(complianceStandardSection);
            return ResponseEntity.ok(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @PutMapping("/standard-section")
    public ResponseEntity<UtmComplianceStandardSection> updateComplianceStandardSection(
        @Valid @RequestBody UtmComplianceStandardSection complianceStandardSection) {
        final String ctx = CLASS_NAME + ".updateComplianceStandardSection";

        try {
            if (complianceStandardSection.getId() == null) {
                String msg = ctx + ": Missing complianceStandardSection id";
                return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
            }
            UtmComplianceStandardSection result = standardSectionService.save(complianceStandardSection);
            return ResponseEntity.ok(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/standard-section")
    public ResponseEntity<List<UtmComplianceStandardSection>> getAllComplianceStandardSection(
        UtmComplianceStandardSectionCriteria criteria, Pageable pageable) {
        final String ctx = CLASS_NAME + ".getAllComplianceStandardSection";
        try {
            Page<UtmComplianceStandardSection> page = complianceStandardSectionQueryService.findByCriteria(criteria,
                pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/compliance/standard-section");
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/standard-section/{id}")
    public ResponseEntity<UtmComplianceStandardSection> getComplianceStandardSection(@PathVariable Long id) {
        final String ctx = CLASS_NAME + ".getComplianceStandardSection";
        try {
            Optional<UtmComplianceStandardSection> standard = standardSectionService.findOne(id);
            return ResponseUtil.wrapOrNotFound(standard);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @DeleteMapping("/standard-section/{id}")
    public ResponseEntity<Void> deleteComplianceStandardSection(@PathVariable Long id) {
        final String ctx = CLASS_NAME + ".deleteComplianceStandardSection";
        try {
            standardSectionService.delete(id);
            return ResponseEntity.ok().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/standard-section/sections-with-reports")
    public ResponseEntity<List<UtmComplianceStandardSection>> getSectionsWithReports(@RequestParam Long standardId) {
        final String ctx = CLASS_NAME + ".getSectionsWithReports";
        try {
            List<UtmComplianceStandardSection> results = standardSectionService.getSectionsWithReportsByStandard(
                standardId);
            return ResponseEntity.ok(results);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }
}

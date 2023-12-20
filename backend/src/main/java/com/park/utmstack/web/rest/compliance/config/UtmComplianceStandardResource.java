package com.park.utmstack.web.rest.compliance.config;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.compliance.UtmComplianceStandard;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.compliance.config.UtmComplianceStandardQueryService;
import com.park.utmstack.service.compliance.config.UtmComplianceStandardService;
import com.park.utmstack.service.dto.compliance.UtmComplianceStandardCriteria;
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
public class UtmComplianceStandardResource {
    private final Logger log = LoggerFactory.getLogger(UtmComplianceStandardResource.class);
    private static final String CLASS_NAME = "UtmComplianceStandardResource";

    private final UtmComplianceStandardService complianceStandardService;
    private final UtmComplianceStandardQueryService complianceStandardQueryService;
    private final ApplicationEventService applicationEventService;
    private final Environment env;

    public UtmComplianceStandardResource(UtmComplianceStandardService complianceStandardService,
                                         UtmComplianceStandardQueryService complianceStandardQueryService,
                                         ApplicationEventService applicationEventService,
                                         Environment env) {
        this.complianceStandardService = complianceStandardService;
        this.complianceStandardQueryService = complianceStandardQueryService;
        this.applicationEventService = applicationEventService;
        this.env = env;
    }

    @PostMapping("/standard")
    public ResponseEntity<UtmComplianceStandard> createComplianceStandard(
        @Valid @RequestBody UtmComplianceStandard complianceStandard) {
        final String ctx = CLASS_NAME + ".createComplianceStandard";

        try {
            if (complianceStandard.getId() != null) {
                String msg = ctx + ": A new utmComplianceStandard can't have an id";
                return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
            }

            if (Arrays.asList(env.getActiveProfiles()).contains(JHipsterConstants.SPRING_PROFILE_DEVELOPMENT)) {
                complianceStandard.setId(complianceStandardService.getSystemSequenceNextValue());
                complianceStandard.setSystemOwner(true);
            }

            UtmComplianceStandard result = complianceStandardService.save(complianceStandard);
            return ResponseEntity.ok(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @PutMapping("/standard")
    public ResponseEntity<UtmComplianceStandard> updateComplianceStandard(
        @Valid @RequestBody UtmComplianceStandard complianceStandard) {
        final String ctx = CLASS_NAME + ".updateComplianceStandard";

        try {
            if (complianceStandard.getId() == null) {
                String msg = ctx + ": Missing utmComplianceStandard id";
                return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
            }

            complianceStandard.setSystemOwner(complianceStandard.getSystemOwner() == null ? complianceStandard.getId() < 1000000 : complianceStandard.getSystemOwner());

            UtmComplianceStandard result = complianceStandardService.save(complianceStandard);
            return ResponseEntity.ok(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/standard")
    public ResponseEntity<List<UtmComplianceStandard>> getAllComplianceStandard(UtmComplianceStandardCriteria criteria,
                                                                                Pageable pageable) {
        final String ctx = CLASS_NAME + ".getAllComplianceStandard";
        try {
            Page<UtmComplianceStandard> page = complianceStandardQueryService.findByCriteria(criteria, pageable);
            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/compliance/standard");
            return ResponseEntity.ok().headers(headers).body(page.getContent());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/standard/{id}")
    public ResponseEntity<UtmComplianceStandard> getComplianceStandard(@PathVariable Long id) {
        final String ctx = CLASS_NAME + ".getComplianceStandard";
        try {
            Optional<UtmComplianceStandard> standard = complianceStandardService.findOne(id);
            return ResponseUtil.wrapOrNotFound(standard);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @DeleteMapping("/standard/{id}")
    public ResponseEntity<Void> deleteComplianceStandard(@PathVariable Long id) {
        final String ctx = CLASS_NAME + ".deleteComplianceStandard";
        try {
            complianceStandardService.delete(id);
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

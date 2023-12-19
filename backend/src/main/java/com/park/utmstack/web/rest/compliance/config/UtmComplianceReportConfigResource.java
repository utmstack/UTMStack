package com.park.utmstack.web.rest.compliance.config;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.domain.compliance.UtmComplianceReportConfig;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.chart_builder.UtmDashboardService;
import com.park.utmstack.service.chart_builder.UtmDashboardVisualizationService;
import com.park.utmstack.service.compliance.config.UtmComplianceReportConfigQueryService;
import com.park.utmstack.service.compliance.config.UtmComplianceReportConfigService;
import com.park.utmstack.service.compliance.config.UtmComplianceStandardSectionService;
import com.park.utmstack.service.compliance.config.UtmComplianceStandardService;
import com.park.utmstack.service.dto.compliance.UtmComplianceReportConfigCriteria;
import com.park.utmstack.util.UtilResponse;
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
import tech.jhipster.web.util.ResponseUtil;

import javax.validation.Valid;
import javax.validation.constraints.NotNull;
import java.util.List;
import java.util.Objects;
import java.util.Optional;
import java.util.stream.Collectors;

@RestController
@RequestMapping("/api/compliance")
public class UtmComplianceReportConfigResource {
    private final Logger log = LoggerFactory.getLogger(UtmComplianceReportConfigResource.class);
    private static final String CLASS_NAME = "UtmComplianceReportConfigResource";

    private final UtmComplianceReportConfigService complianceReportConfigService;
    private final UtmComplianceReportConfigQueryService complianceReportConfigQueryService;
    private final UtmComplianceStandardSectionService standardSectionService;
    private final UtmComplianceStandardService standardService;
    private final UtmDashboardVisualizationService dashboardVisualizationService;
    private final UtmDashboardService dashboardService;
    private final ApplicationEventService applicationEventService;

    public UtmComplianceReportConfigResource(UtmComplianceReportConfigService complianceReportConfigService,
                                             UtmComplianceReportConfigQueryService complianceReportConfigQueryService,
                                             UtmComplianceStandardSectionService standardSectionService,
                                             UtmComplianceStandardService standardService,
                                             UtmDashboardVisualizationService dashboardVisualizationService,
                                             UtmDashboardService dashboardService,
                                             ApplicationEventService applicationEventService) {
        this.complianceReportConfigService = complianceReportConfigService;
        this.complianceReportConfigQueryService = complianceReportConfigQueryService;
        this.standardSectionService = standardSectionService;
        this.standardService = standardService;
        this.dashboardVisualizationService = dashboardVisualizationService;
        this.dashboardService = dashboardService;
        this.applicationEventService = applicationEventService;
    }

    @PostMapping("/report-config")
    public ResponseEntity<UtmComplianceReportConfig> createComplianceReportConfig(
        @Valid @RequestBody UtmComplianceReportConfig complianceReportConfig) {
        final String ctx = CLASS_NAME + ".createComplianceReportConfig";

        try {
            if (complianceReportConfig.getId() != null) {
                String msg = ctx + ": A new complianceReportConfig can't have an id";
                return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
            }

            UtmComplianceReportConfig result = complianceReportConfigService.save(complianceReportConfig);
            return ResponseEntity.ok(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @PutMapping("/report-config")
    public ResponseEntity<UtmComplianceReportConfig> updateComplianceReportConfig(
        @Valid @RequestBody UtmComplianceReportConfig complianceReportConfig) {
        final String ctx = CLASS_NAME + ".updateComplianceReportConfig";

        try {
            if (complianceReportConfig.getId() == null) {
                String msg = ctx + ": Missing complianceReportConfig id";
                return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                    HeaderUtil.createFailureAlert("", "", msg)).body(null);
            }

            UtmComplianceReportConfig result = complianceReportConfigService.save(complianceReportConfig);
            return ResponseEntity.ok(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/report-config")
    public ResponseEntity<List<UtmComplianceReportConfig>> getAllComplianceReportConfig(
        UtmComplianceReportConfigCriteria criteria, Pageable pageable) {
        final String ctx = CLASS_NAME + ".getAllComplianceReportConfig";
        try {
            Page<UtmComplianceReportConfig> page = complianceReportConfigQueryService.findByCriteria(criteria, pageable);
            List<UtmComplianceReportConfig> reports = page.getContent();

            List<UtmComplianceReportConfig> result = reports.stream().filter(UtmComplianceReportConfig::getConfigReportEditable).collect(Collectors.toList());

            if (!Objects.isNull(criteria.getExpandDashboard()) && criteria.getExpandDashboard().getEquals()) {
                for (UtmComplianceReportConfig report : result)
                    dashboardVisualizationService.findAllByIdDashboard(report.getDashboardId()).ifPresent(report::setDashboard);
            }

            HttpHeaders headers = PaginationUtil.generatePaginationHttpHeaders(page, "/api/compliance/report-config");
            return ResponseEntity.ok().headers(headers).body(result);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/report-config/{id}")
    public ResponseEntity<UtmComplianceReportConfig> getComplianceReportConfig(@PathVariable Long id,
                                                                               @RequestParam(required = false) Boolean expandDashboard) {
        final String ctx = CLASS_NAME + ".getComplianceReportConfig";
        try {
            Optional<UtmComplianceReportConfig> standard = complianceReportConfigService.findOne(id);
            if (standard.isPresent() && !Objects.isNull(expandDashboard) && expandDashboard) {
                UtmComplianceReportConfig report = standard.get();
                dashboardVisualizationService.findAllByIdDashboard(report.getDashboardId()).ifPresent(report::setDashboard);
                return ResponseEntity.ok(report);
            }
            return ResponseUtil.wrapOrNotFound(standard);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @DeleteMapping("/report-config/{id}")
    public ResponseEntity<Void> deleteComplianceReportConfig(@PathVariable Long id) {
        final String ctx = CLASS_NAME + ".deleteComplianceReportConfig";
        try {
            complianceReportConfigService.delete(id);
            return ResponseEntity.ok().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @GetMapping("/report-config/get-by-filters")
    public ResponseEntity<List<UtmComplianceReportConfig>> getReportsByFilters(@RequestParam(required = false) Long standardId,
                                                                               @RequestParam(required = false) String solution,
                                                                               @RequestParam(required = false) Long sectionId,
                                                                               @RequestParam(required = false) Boolean expandDashboard,
                                                                               Pageable pageable) {
        final String ctx = CLASS_NAME + ".getReportsByFilters";
        try {
            List<UtmComplianceReportConfig> page = complianceReportConfigService.getReportsByFilters(standardId, solution, sectionId, pageable);

            if (!Objects.isNull(expandDashboard) && expandDashboard) {
                for (UtmComplianceReportConfig report : page)
                    dashboardVisualizationService.findAllByIdDashboard(report.getDashboardId()).ifPresent(report::setDashboard);
            }
            return ResponseEntity.ok().body(page);
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert("", "", msg)).body(null);
        }
    }

    @PostMapping("/report-config/import")
    public ResponseEntity<Void> importReports(@Valid @RequestBody ImportReportsBody body) {
        final String ctx = CLASS_NAME + ".importReports";
        try {
            complianceReportConfigService.importReports(body.getReports(), body.getOverride());
            return ResponseEntity.ok().build();
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }

    public static class ImportReportsBody {
        @NotNull
        private List<UtmComplianceReportConfig> reports;

        @NotNull
        private Boolean override;

        public List<UtmComplianceReportConfig> getReports() {
            return reports;
        }

        public void setReports(List<UtmComplianceReportConfig> reports) {
            this.reports = reports;
        }

        public Boolean getOverride() {
            return override;
        }

        public void setOverride(Boolean override) {
            this.override = override;
        }
    }
}

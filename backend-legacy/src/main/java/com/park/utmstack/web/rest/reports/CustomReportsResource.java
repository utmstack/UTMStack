package com.park.utmstack.web.rest.reports;

import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.reports.CustomReportService;
import com.park.utmstack.web.rest.util.HeaderUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.core.io.ByteArrayResource;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.io.ByteArrayOutputStream;
import java.time.Instant;
import java.util.Optional;


/**
 * REST controller for managing UtmReports.
 */
@RestController
@RequestMapping("/api")
public class CustomReportsResource {

    private static final String CLASS_NAME = "CustomReportsResource";
    private final Logger log = LoggerFactory.getLogger(CustomReportsResource.class);

    private static final String ENTITY_NAME = "utmReports";

    private final ApplicationEventService applicationEventService;
    private final CustomReportService customReportService;

    public CustomReportsResource(ApplicationEventService applicationEventService,
                                 CustomReportService customReportService) {
        this.applicationEventService = applicationEventService;
        this.customReportService = customReportService;
    }

    @GetMapping("/custom-reports/buildThreatActivityForAlerts")
    public ResponseEntity<ByteArrayResource> buildThreatActivityForAlerts(@RequestParam Instant from,
                                                                          @RequestParam Instant to,
                                                                          @RequestParam Integer top) {
        final String ctx = CLASS_NAME + ".buildThreatActivity";
        try {
            Optional<ByteArrayOutputStream> baos = customReportService.buildThreatActivityForAlerts(from, to, top);
            return baos.map(byteArrayOutputStream -> ResponseEntity.ok().header(HttpHeaders.CONTENT_DISPOSITION,
                "attachment;filename=report.pdf").contentType(
                MediaType.APPLICATION_PDF).contentLength(byteArrayOutputStream.size())
                .body(new ByteArrayResource(byteArrayOutputStream.toByteArray()))).orElseGet(() -> ResponseEntity.ok().build());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    @GetMapping("/custom-reports/buildThreatActivityForIncidents")
    public ResponseEntity<ByteArrayResource> buildThreatActivityForIncidents(@RequestParam Instant from,
                                                                             @RequestParam Instant to,
                                                                             @RequestParam Integer top) {
        final String ctx = CLASS_NAME + ".buildThreatActivityForIncidents";
        try {
            Optional<ByteArrayOutputStream> baos = customReportService.buildThreatActivityForIncidents(from, to, top);
            return baos.map(byteArrayOutputStream -> ResponseEntity.ok().header(HttpHeaders.CONTENT_DISPOSITION,
                "attachment;filename=report.pdf").contentType(
                MediaType.APPLICATION_PDF).contentLength(byteArrayOutputStream.size())
                .body(new ByteArrayResource(byteArrayOutputStream.toByteArray()))).orElseGet(() -> ResponseEntity.ok().build());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }

    @GetMapping("/custom-reports/buildAssetManagement")
    public ResponseEntity<ByteArrayResource> buildAssetManagement(@RequestParam Instant from,
                                                                  @RequestParam Instant to,
                                                                  @RequestParam Integer top) {
        final String ctx = CLASS_NAME + ".buildAssetManagement";
        try {
            Optional<ByteArrayOutputStream> baos = customReportService.buildAssetManagement(from, to, top);
            return baos.map(byteArrayOutputStream -> ResponseEntity.ok().header(HttpHeaders.CONTENT_DISPOSITION,
                "attachment;filename=report.pdf").contentType(
                MediaType.APPLICATION_PDF).contentLength(byteArrayOutputStream.size())
                .body(new ByteArrayResource(byteArrayOutputStream.toByteArray()))).orElseGet(() -> ResponseEntity.ok().build());
        } catch (Exception e) {
            String msg = ctx + ": " + e.getMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).headers(
                HeaderUtil.createFailureAlert(null, null, msg)).body(null);
        }
    }
}

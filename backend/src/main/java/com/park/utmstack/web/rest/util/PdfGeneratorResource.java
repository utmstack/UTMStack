package com.park.utmstack.web.rest.util;

import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.application_events.enums.ApplicationEventType;
import com.park.utmstack.security.jwt.JWTFilter;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.util.PdfService;
import com.park.utmstack.util.UtilResponse;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import javassist.NotFoundException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

/**
 * REST controller for managing {@link PdfGeneratorResource}.
 */
@RestController
@RequestMapping("/api")
public class PdfGeneratorResource {

    private static final String CLASSNAME = "PdfGeneratorResource";
    private final Logger log = LoggerFactory.getLogger(PdfGeneratorResource.class);

    private final ApplicationEventService applicationEventService;


    private final PdfService pdfService;

    public PdfGeneratorResource(
        ApplicationEventService applicationEventService,
        PdfService pdfService) {
        this.applicationEventService = applicationEventService;
        this.pdfService = pdfService;
    }

    /**
     * {@code GET  /generate-pdf-report} : get pdf report in bytes, from a dashboard or compliance.
     *
     * @return the {@link ResponseEntity} with status {@code 200 (OK)} and the pdf report in bytes in body.
     */
    @RequestMapping(
        value = "/generate-pdf-report",
        produces = MediaType.APPLICATION_PDF_VALUE,
        method = RequestMethod.GET
    )
    public ResponseEntity getPdfReportInBytes(@RequestParam String url,
                                              @RequestParam PdfService.PdfAccessTypes accessType,
                                              @RequestParam String filename,
                                              @RequestHeader(JWTFilter.AUTHORIZATION_HEADER) String accessKey) {
        final String ctx = CLASSNAME + ".getPdfReportInBytes";
        try {
            if (accessType == PdfService.PdfAccessTypes.PDF_TYPE_INTERNAL) {
                throw new BadRequestAlertException("Access type ("+ accessType.get() +") not allowed","pdfreport","accesskeynotvalid");
            }
            log.debug("REST request to get pdf report");

            HttpHeaders headers = new HttpHeaders();
            headers.add(HttpHeaders.CACHE_CONTROL, "no-cache, no-store, must-revalidate");
            headers.add(HttpHeaders.PRAGMA, "no-cache");
            headers.add(HttpHeaders.EXPIRES, "0");
            headers.add(HttpHeaders.CONTENT_DISPOSITION, "attachment;filename=" + filename);
            headers.add(HttpHeaders.CONTENT_TYPE, MediaType.APPLICATION_PDF_VALUE);

            byte [] resultPdf = pdfService.getPdf( url, accessKey.substring(7), accessType.get());
            if (resultPdf != null && resultPdf.length > 0 ) {
                return ResponseEntity.ok().headers(headers).body(resultPdf);
            } else {
                throw new NotFoundException("We couldn't generate the pdf, reason: No data returned from PDF service");
            }
        } catch (Exception e) {
            String msg = ctx + ": " + e.getLocalizedMessage();
            log.error(msg);
            applicationEventService.createEvent(msg, ApplicationEventType.ERROR);
            return UtilResponse.buildErrorResponse(HttpStatus.INTERNAL_SERVER_ERROR, msg);
        }
    }
}

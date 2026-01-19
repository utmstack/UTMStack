package com.park.utmstack.web.rest.util;

import com.park.utmstack.security.jwt.JWTFilter;
import com.park.utmstack.service.dto.web_pdf.PdfServiceResponse;
import com.park.utmstack.service.util.PdfService;
import com.park.utmstack.web.rest.errors.BadRequestAlertException;
import lombok.RequiredArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

/**
 * REST controller for managing {@link PdfGeneratorResource}.
 */
@RestController
@RequiredArgsConstructor
@RequestMapping("/api")
public class PdfGeneratorResource {

    private static final String CLASSNAME = "PdfGeneratorResource";
    private final Logger log = LoggerFactory.getLogger(PdfGeneratorResource.class);

    private final PdfService pdfService;

    @GetMapping("/generate-pdf-report")
    public ResponseEntity<?> getPdfReportInBytes(@RequestParam String url,
                                                 @RequestParam PdfService.PdfAccessTypes accessType,
                                                 @RequestParam String filename,
                                                 @RequestHeader(JWTFilter.AUTHORIZATION_HEADER) String accessKey) {

        final String ctx = CLASSNAME + ".getPdfReportInBytes";

        if (accessType == PdfService.PdfAccessTypes.PDF_TYPE_INTERNAL) {
            throw new BadRequestAlertException(String.format("PDF Service not implemented for %s", accessType), CLASSNAME, null);
        }

        log.debug("REST request to get pdf report");

        PdfServiceResponse pdfResponse =
                pdfService.downloadPdf(url, accessKey.substring(7), accessType.get());

        HttpHeaders headers = new HttpHeaders();
        headers.add(HttpHeaders.CACHE_CONTROL, "no-cache, no-store, must-revalidate");
        headers.add(HttpHeaders.PRAGMA, "no-cache");
        headers.add(HttpHeaders.EXPIRES, "0");
        headers.add(HttpHeaders.CONTENT_DISPOSITION, "attachment;filename=" + filename);
        headers.add(HttpHeaders.CONTENT_TYPE, MediaType.APPLICATION_PDF_VALUE);

        return ResponseEntity.ok()
                .headers(headers)
                .body(pdfResponse.getPdfBytes());
    }
}

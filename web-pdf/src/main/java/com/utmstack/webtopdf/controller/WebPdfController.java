package com.utmstack.webtopdf.controller;

import com.utmstack.webtopdf.config.enums.AccessType;
import com.utmstack.webtopdf.dto.ResponseDto;
import com.utmstack.webtopdf.service.PdfGenerationService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestHeader;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;


@Slf4j
@RequiredArgsConstructor
@RestController
@RequestMapping(path = "/")
public class WebPdfController {

    private final PdfGenerationService pdfGenerationService;

    @GetMapping("/generate-pdf")
    public ResponseEntity<ResponseDto> generatePdf(@RequestParam String baseUrl,
                                                   @RequestParam String url,
                                                   @RequestParam String accessType,
                                                   @RequestParam String accessKey,
                                                   @RequestHeader(value = "X-Utm-Internal-Key", required = false) String internalKey) {

        if (!internalKeyMatches(internalKey, System.getenv("INTERNAL_KEY"))) {
            log.warn("Rejected /generate-pdf request: missing or invalid X-Utm-Internal-Key");
            return ResponseEntity
                    .status(HttpStatus.UNAUTHORIZED)
                    .body(ResponseDto.builder()
                            .error(true)
                            .message("Unauthorized")
                            .build());
        }


        byte[] pdfBytes = pdfGenerationService.generatePdf(baseUrl,
                                                           url,
                                                           accessKey,
                                                           AccessType.valueOf(accessType.toUpperCase()));

        if (pdfBytes == null || pdfBytes.length == 0) {
            log.error("PDF generation returned empty bytes for URL: {}", url);

            return ResponseEntity
                    .status(HttpStatus.BAD_REQUEST)
                    .body(ResponseDto.builder()
                            .error(true)
                            .message("Failed to generate PDF: No content returned")
                            .build());
        }

        return ResponseEntity.ok(
                ResponseDto.builder()
                        .pdfBytes(pdfBytes)
                        .error(false)
                        .message("PDF generated successfully")
                        .build());
    }

    /**
     * Constant-time comparison of the caller's internal key against the
     * deployment's {@code INTERNAL_KEY} environment value.
     *
     * <p>Fails closed: if either side is null or blank (e.g. the env is unset
     * in a dev setup) the check returns {@code false} and the endpoint 401s
     * — there is no dev bypass.
     */
    static boolean internalKeyMatches(String headerKey, String envKey) {
        if (headerKey == null || headerKey.isEmpty()
                || envKey == null || envKey.isEmpty()) {
            return false;
        }
        byte[] a = headerKey.getBytes(StandardCharsets.UTF_8);
        byte[] b = envKey.getBytes(StandardCharsets.UTF_8);
        return MessageDigest.isEqual(a, b);
    }
}

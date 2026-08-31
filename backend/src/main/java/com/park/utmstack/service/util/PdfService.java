package com.park.utmstack.service.util;

import com.park.utmstack.config.Constants;
import com.park.utmstack.service.dto.web_pdf.PdfServiceResponse;
import com.park.utmstack.service.web_clients.rest_template.RestTemplateService;
import com.park.utmstack.util.exceptions.ApiException;
import lombok.RequiredArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.StringUtils;
import org.springframework.web.util.UriComponentsBuilder;

import java.util.Set;

/**
 * Service Implementation for PDF generation.
 */
@Service
@RequiredArgsConstructor
@Transactional
public class PdfService {
    private final Logger log = LoggerFactory.getLogger(PdfService.class);
    private static final String CLASSNAME = "PdfService";
    private final RestTemplateService restTemplateService;


    /**
     * Path prefixes allowed for PDF generation. The frontend only ever requests
     * print/export views of its own application; anything else is rejected so the
     * url parameter cannot be used to make the PDF service render arbitrary
     * pages (SSRF via the headless browser).
     */
    private static final Set<String> ALLOWED_PDF_PATH_PREFIXES = Set.of(
            "/dashboard",
            "/compliance",
            "/data/alert/detail"
    );

    public PdfServiceResponse downloadPdf(String url, String accessKey, String accessType) {
        final String ctx = CLASSNAME + ".getPdf";

        validatePdfUrl(url);

        String urlService = UriComponentsBuilder.fromUriString(Constants.PDF_SERVICE_URL)
                .queryParam("baseUrl", Constants.FRONT_BASE_URL)
                .queryParam("url", url)
                .queryParam("accessKey", accessKey)
                .queryParam("accessType", accessType)
                .build().toUriString();
        try {
            log.info("Requesting PDF creation to URL : {}", urlService);
            ResponseEntity<PdfServiceResponse> rs =
                    restTemplateService.getRaw(urlService, PdfServiceResponse.class);

            if (!rs.getStatusCode().is2xxSuccessful()) {
                PdfServiceResponse errorBody = rs.getBody();

                String message = (errorBody != null && errorBody.getMessage() != null)
                        ? errorBody.getMessage()
                        : "Unknown error returned from PDF service";

                throw new ApiException(message, rs.getStatusCode());
            }

            PdfServiceResponse body = rs.getBody();

            if (body == null || body.getPdfBytes() == null || body.getPdfBytes().length == 0) {
                log.error("{}: No data returned from PDF service", ctx);

                PdfServiceResponse error = new PdfServiceResponse();
                error.setError(true);
                error.setMessage("No data returned from PDF service");
                return error;
            }

            return body;

        } catch (ApiException e) {
            throw e;
        } catch (Exception e){
            log.error("{}: Exception occurred while requesting PDF service: {}", ctx, e.getMessage());
            throw new ApiException(e.getMessage(), HttpStatus.INTERNAL_SERVER_ERROR);
        }
    }

    /**
     * The url must be a relative path into the UTMStack frontend. The frontend
     * always sends its own Angular routes (e.g. {@code /dashboard/overview},
     * {@code /compliance/print-view}); the headless browser in the web-pdf
     * service loads them client-side against the fixed frontend origin.
     *
     * <p>Anything that is not a plain relative route — an absolute URL
     * ({@code http://...}), a protocol-relative target ({@code //host}), or a
     * non-http scheme ({@code javascript:}, {@code file:}, {@code data:}) — is
     * rejected, because it would make the headless browser leave the frontend
     * and fetch an arbitrary (potentially internal) URL (SSRF,
     * CVE-2026-82044). As defense in depth, the route must also fall under one
     * of the known print/export prefixes.
     */
    private void validatePdfUrl(String url) {
        final String ctx = CLASSNAME + ".validatePdfUrl";

        if (!StringUtils.hasText(url)) {
            throw new ApiException("PDF report url is required", HttpStatus.BAD_REQUEST);
        }

        String candidate = url.trim().replace('\\', '/');

        if (candidate.contains("://") || candidate.startsWith("//") || !candidate.startsWith("/")) {
            log.warn("{}: Rejected non-relative PDF url: {}", ctx, candidate);
            throw new ApiException("PDF report url must be a relative path on the UTMStack frontend", HttpStatus.BAD_REQUEST);
        }

        String lowerPath = candidate.toLowerCase();
        if (!ALLOWED_PDF_PATH_PREFIXES.stream().anyMatch(p -> lowerPath.equals(p) || lowerPath.startsWith(p + "/"))) {
            log.warn("{}: Rejected PDF url outside allowed route prefixes: {}", ctx, candidate);
            throw new ApiException("PDF report url is not allowed", HttpStatus.BAD_REQUEST);
        }
    }


    /**
     * Enum used to define type of access used when accessing the PDF microservice
     * */
    public enum PdfAccessTypes {
        PDF_TYPE_INTERNAL("Utm_Internal_Key"),
        PDF_TYPE_TOKEN("Utm_Token");

        private String type;
        PdfAccessTypes (String type) {
            this.type = type;
        }
        public String get() {
            return this.type;
        }
    }
}

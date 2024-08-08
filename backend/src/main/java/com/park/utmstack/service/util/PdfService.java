package com.park.utmstack.service.util;

import com.park.utmstack.config.Constants;
import com.park.utmstack.service.compliance.ComplianceFileResponse;
import com.park.utmstack.service.federation_service.UtmFederationServiceClientService;
import com.park.utmstack.service.web_clients.rest_template.RestTemplateService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.web.util.UriComponentsBuilder;

import java.util.Objects;

/**
 * Service Implementation for PDF generation.
 */
@Service
@Transactional
public class PdfService {
    private final Logger log = LoggerFactory.getLogger(PdfService.class);
    private static final String CLASSNAME = "PdfService";
    private final UtmFederationServiceClientService fsService;
    private final RestTemplateService restTemplateService;

    private final String COMPLIANCE_EXPORT_URL = "/dashboard/export-compliance/";

    public PdfService(UtmFederationServiceClientService fsService,
                      RestTemplateService restTemplateService) {
        this.fsService = fsService;
        this.restTemplateService = restTemplateService;
    }

    /**
     * Get pdf report in bytes array.
     *
     * @param url the url of the compliance report.
     * @return the pdf report in bytes array.
     */
    @Transactional(readOnly = true)
    public ResponseEntity<byte []> getPdfReportByUrlInBytes(String url, String accessKey, String accessType) throws Exception{
        log.debug("Request to get UtmComplianceReportSchedule : {}", url);
        return ResponseEntity.ok().body(getPdf(Constants.FRONT_BASE_URL + url, accessKey, accessType));
    }

    /**
     * Method to get pdf in bytes
     */
    public byte[] getPdf(String url, String accessKey, String accessType) throws Exception {
        final String ctx = CLASSNAME + ".getPdf";

        String urlService = UriComponentsBuilder.fromUriString(Constants.PDF_SERVICE_URL)
            .queryParam("baseUrl", Constants.FRONT_BASE_URL)
            .queryParam("url", url)
            .queryParam("accessKey", accessKey)
            .queryParam("accessType", accessType)
            .build().toUriString();

        ResponseEntity<ComplianceFileResponse> rs = restTemplateService.get(urlService, ComplianceFileResponse.class);
        log.info("Requesting PDF creation to URL : {}", Constants.PDF_SERVICE_URL + "?url=" + url);
        if (!rs.getStatusCode().is2xxSuccessful()) {
            log.error(ctx + ": " + restTemplateService.extractErrorMessage(rs));
        } else {
            byte[] pdfInBytes = Objects.requireNonNull(rs.getBody()).getPdfBytes();
            if (pdfInBytes != null && pdfInBytes.length > 0) {
                return pdfInBytes;
            } else {
                log.error(ctx + ": We couldn't generate the pdf, reason: No data returned from PDF service");
            }
        }
        return null;
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

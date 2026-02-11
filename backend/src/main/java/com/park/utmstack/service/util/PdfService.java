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
import org.springframework.web.util.UriComponentsBuilder;

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


    public PdfServiceResponse downloadPdf(String url, String accessKey, String accessType) {
        final String ctx = CLASSNAME + ".getPdf";

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

        } catch (Exception e){
            log.error("{}: Exception occurred while requesting PDF service: {}", ctx, e.getMessage());
            throw new ApiException(e.getMessage(), HttpStatus.INTERNAL_SERVER_ERROR);
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

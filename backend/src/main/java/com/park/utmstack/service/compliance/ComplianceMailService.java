package com.park.utmstack.service.compliance;

import com.park.utmstack.service.MailService;
import com.park.utmstack.service.dto.web_pdf.PdfServiceResponse;
import com.park.utmstack.service.util.PdfService;
import lombok.RequiredArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.time.Clock;
import java.time.Instant;

@Service
@RequiredArgsConstructor
public class ComplianceMailService {

    private static final Logger log = LoggerFactory.getLogger(ComplianceMailService.class);
    private static final String CLASSNAME = "ComplianceMailService";

    private final MailService mailService;
    private final PdfService pdfService;

    public void sendComplianceByMail(String url, String userEmail) {
        final String ctx = CLASSNAME + ".sendComplianceByMail";

        String accessKey = System.getenv("INTERNAL_KEY");

        if (accessKey == null || accessKey.isBlank()) {
            log.error("{}: INTERNAL_KEY environment variable is missing", ctx);
            return;
        }

        PdfServiceResponse response =
                pdfService.downloadPdf(url, accessKey, PdfService.PdfAccessTypes.PDF_TYPE_INTERNAL.get());

        if (response.getPdfBytes() == null || response.getPdfBytes().length == 0) {
            log.error("{}: PDF service returned empty content for URL {}", ctx, url);
            return;
        }

        String filename = "Compliance_Report_" + Instant.now(Clock.systemUTC()) + ".pdf";

        mailService.sendComplianceReportEmail(
                userEmail,
                "UTMStack Compliance Report Delivery",
                "This is a scheduled email delivery of a Compliance Report, please do not answer this email.",
                filename,
                response.getPdfBytes()
        );

        log.info("{}: Email successfully sent to {}", ctx, userEmail);

    }
}


package com.park.utmstack.service.compliance;

import com.park.utmstack.service.MailService;
import com.park.utmstack.service.util.PdfService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Clock;
import java.time.Instant;

/**
 * Service Implementation for managing Compliance PDF Delivery.
 */
@Service
@Transactional
public class ComplianceMailService {
    private final Logger log = LoggerFactory.getLogger(ComplianceMailService.class);
    private static final String CLASSNAME = "ComplianceMailService";
    private final MailService mailService;
    private final PdfService pdfService;

    public ComplianceMailService(MailService mailService,
                                 PdfService pdfService) {
        this.mailService = mailService;
        this.pdfService = pdfService;
    }

    /**
     * Method to generate dashboard in PDF format and send via email
     */
    public void sendComplianceByMail(String url, String userEmail) throws Exception {
        final String ctx = CLASSNAME + ".sendComplianceByMail";

        String accessKey = System.getenv("INTERNAL_KEY");
        byte[] pdfInBytes = pdfService.getPdf(url, accessKey,PdfService.PdfAccessTypes.PDF_TYPE_INTERNAL.get());
        if (pdfInBytes != null && pdfInBytes.length > 0) {
            mailService.sendComplianceReportEmail(userEmail, "UTMStack Compliance Report Delivery"
                , "This is a scheduled email delivery of a Compliance Report, please do not answer this email. ",
                "Compliance_Report_" + Instant.now(Clock.systemUTC()) + ".pdf", pdfInBytes);
        } else {
            log.error(ctx + ": We couldn't send the email, reason: No data returned from PDF service");
        }
    }
}

package com.park.utmstack.service.compliance;

import com.park.utmstack.domain.Authority;
import com.park.utmstack.domain.User;
import com.park.utmstack.repository.UserRepository;
import com.park.utmstack.security.jwt.TokenProvider;
import com.park.utmstack.service.MailService;
import com.park.utmstack.service.dto.web_pdf.PdfServiceResponse;
import com.park.utmstack.service.util.PdfService;
import lombok.RequiredArgsConstructor;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.GrantedAuthority;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.stereotype.Service;

import java.time.Clock;
import java.time.Instant;
import java.util.Collection;
import java.util.Collections;
import java.util.List;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
public class ComplianceMailService {

    private static final Logger log = LoggerFactory.getLogger(ComplianceMailService.class);
    private static final String CLASSNAME = "ComplianceMailService";

    private final MailService mailService;
    private final PdfService pdfService;
    private final TokenProvider tokenProvider;
    private final UserRepository userRepository;

    public void sendComplianceByMail(String url, String userEmail) {
        final String ctx = CLASSNAME + ".sendComplianceByMail";

        // The scheduled path has no logged-in user, so it mints a short-lived admin
        // JWT and drives the headless browser through the normal ?token= flow
        // (PdfAccessTypes.PDF_TYPE_TOKEN). It must NOT use the Utm-Internal-Key:
        // since the CVE-2026-82042 hardening, the internal key only authenticates
        // on a fixed machine-to-machine allowlist, so the report's data endpoints
        // (/api/account, /api/compliance/report-config, /api/utm-dashboard-visualizations)
        // 401'd and the PDF never rendered (#3144).
        User admin = userRepository.findAdminUsers().stream()
                .findFirst()
                .orElseThrow(() -> new IllegalStateException(ctx + ": no active admin user to render the compliance report"));
        Collection<Authority> userAuthorities =
                admin.getAuthorities() == null ? Collections.emptySet() : admin.getAuthorities();
        List<GrantedAuthority> authorities = userAuthorities.stream()
                .map(a -> (GrantedAuthority) new SimpleGrantedAuthority(a.getName()))
                .collect(Collectors.toList());
        UsernamePasswordAuthenticationToken auth =
                new UsernamePasswordAuthenticationToken(admin.getLogin(), null, authorities);
        String accessToken = tokenProvider.createToken(auth, false, true);

        PdfServiceResponse response =
                pdfService.downloadPdf(url, accessToken, PdfService.PdfAccessTypes.PDF_TYPE_TOKEN.get());

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

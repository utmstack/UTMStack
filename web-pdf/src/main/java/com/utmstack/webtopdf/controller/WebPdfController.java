package com.utmstack.webtopdf.controller;

import com.utmstack.webtopdf.config.enums.AccessType;
import com.utmstack.webtopdf.dto.ResponseDto;
import com.utmstack.webtopdf.service.PdfGenerationService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;


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
                                                   @RequestParam String accessKey) {


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
}

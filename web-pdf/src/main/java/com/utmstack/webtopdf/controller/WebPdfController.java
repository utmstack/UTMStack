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
    public ResponseEntity<ResponseDto> generatePdf(@RequestParam String baseUrl, @RequestParam String url, @RequestParam String accessType, @RequestParam String accessKey) {
        try {
            byte[] pdfBytes = pdfGenerationService.generatePdf(baseUrl, url, accessKey, AccessType.valueOf(accessType.toUpperCase()));

            return ResponseEntity.ok().body(ResponseDto.builder().pdfBytes(pdfBytes).build());

        }  catch (Exception e) {
            log.error("Error generating the PDF for the URL: {}", url, e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body(ResponseDto.builder().message("Error generating the PDF for the URL").build());
        }
    }
}

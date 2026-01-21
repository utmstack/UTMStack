package com.park.utmstack.service.dto.web_pdf;

import lombok.Data;

@Data
public class PdfServiceResponse {
    private boolean error;
    private String message;
    private String details;
    private byte[] pdfBytes;
}

